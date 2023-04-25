/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package image

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/image"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateImageExt ...
func (svc *imageSvc) BatchUpdateImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchUpdateImageExt[dataproto.TCloudImageExtensionUpdateReq](cts, svc)
	case enumor.Aws:
		return batchUpdateImageExt[dataproto.AwsImageExtensionUpdateReq](cts, svc)
	case enumor.Gcp:
		return batchUpdateImageExt[dataproto.GcpImageExtensionUpdateReq](cts, svc)
	case enumor.HuaWei:
		return batchUpdateImageExt[dataproto.HuaWeiImageExtensionUpdateReq](cts, svc)
	case enumor.Azure:
		return batchUpdateImageExt[dataproto.AzureImageExtensionUpdateReq](cts, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchUpdateImageExt[T dataproto.ImageExtensionUpdateReq](
	cts *rest.Contexts,
	svc *imageSvc,
) (interface{}, error) {
	req := new(dataproto.ImageExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	queryIDs := make([]string, len(*req))
	for indx, imageReq := range *req {
		queryIDs[indx] = imageReq.ID
	}
	rawExtensions, err := svc.rawExtensions(cts, tools.ContainersExpression("id", queryIDs))
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, imageReq := range *req {
			updateData := &tablecloud.ImageModel{
				State: imageReq.State,
			}

			if imageReq.Extension != nil {
				rawExtension, exist := rawExtensions[imageReq.ID]
				if !exist {
					return nil, fmt.Errorf("image id (%s) not exit", imageReq.ID)
				}
				mergedExtension, err := json.UpdateMerge(imageReq.Extension, string(rawExtension))
				if err != nil {
					return nil, fmt.Errorf("image id (%s) merge extension failed, err: %v", imageReq.ID, err)
				}
				updateData.Extension = tabletype.JsonField(mergedExtension)
			}

			if err := svc.dao.Image().UpdateByIDWithTx(cts.Kit, txn, imageReq.ID, updateData); err != nil {
				return nil, fmt.Errorf("update image failed, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// rawExtensions 根据条件查询原始的 extension 字段, 返回字典结构 {"镜像 ID": "原始的 extension 字段"}
// TODO 不同资源可以复用 rawExtensions 逻辑
func (svc *imageSvc) rawExtensions(
	cts *rest.Contexts,
	filterExp *filter.Expression,
) (map[string]tabletype.JsonField, error) {
	opt := &types.ListOption{
		Filter: filterExp,
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
		Fields: []string{"id", "extension"},
	}
	data, err := svc.dao.Image().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	extensions := make(map[string]tabletype.JsonField)
	for _, d := range data.Details {
		extensions[d.ID] = d.Extension
	}

	return extensions, nil
}
