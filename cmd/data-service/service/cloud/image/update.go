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
	coreimage "hcm/pkg/api/core/cloud/image"
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
		return batchUpdateImageExt[coreimage.TCloudExtension](cts, svc)
	case enumor.Aws:
		return batchUpdateImageExt[coreimage.AwsExtension](cts, svc)
	case enumor.Gcp:
		return batchUpdateImageExt[coreimage.GcpExtension](cts, svc)
	case enumor.HuaWei:
		return batchUpdateImageExt[coreimage.HuaWeiExtension](cts, svc)
	case enumor.Azure:
		return batchUpdateImageExt[coreimage.AzureExtension](cts, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchUpdateImageExt[T coreimage.Extension](cts *rest.Contexts, svc *imageSvc) (interface{}, error) {

	req := new(dataproto.BatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	queryIDs := make([]string, len(req.Items))
	for index, one := range req.Items {
		queryIDs[index] = one.ID
	}
	rawExtensions, err := svc.rawExtensions(cts, tools.ContainersExpression("id", queryIDs))
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Items {
			updateData := &tablecloud.ImageModel{
				State:  one.State,
				OsType: one.OsType,
			}

			if one.Extension != nil {
				rawExtension, exist := rawExtensions[one.ID]
				if !exist {
					return nil, fmt.Errorf("image id (%s) not exit", one.ID)
				}
				mergedExtension, err := json.UpdateMerge(one.Extension, string(rawExtension))
				if err != nil {
					return nil, fmt.Errorf("image id (%s) merge extension failed, err: %v", one.ID, err)
				}
				updateData.Extension = tabletype.JsonField(mergedExtension)
			}

			if err := svc.dao.Image().UpdateByIDWithTx(cts.Kit, txn, one.ID, updateData); err != nil {
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
func (svc *imageSvc) rawExtensions(cts *rest.Contexts, filterExp *filter.Expression) (
	map[string]tabletype.JsonField, error) {

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
