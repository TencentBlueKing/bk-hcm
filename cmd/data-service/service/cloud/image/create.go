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
	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecloud "hcm/pkg/dal/table/cloud/image"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateImageExt ...
func (svc *imageSvc) BatchCreateImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchCreateImageExt[coreimage.TCloudExtension](cts, svc, vendor)
	case enumor.Aws:
		return batchCreateImageExt[coreimage.AwsExtension](cts, svc, vendor)
	case enumor.Gcp:
		return batchCreateImageExt[coreimage.GcpExtension](cts, svc, vendor)
	case enumor.HuaWei:
		return batchCreateImageExt[coreimage.HuaWeiExtension](cts, svc, vendor)
	case enumor.Azure:
		return batchCreateImageExt[coreimage.AzureExtension](cts, svc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchCreateImageExt[T coreimage.Extension](cts *rest.Contexts, svc *imageSvc, vendor enumor.Vendor) (
	interface{}, error) {

	req := new(dataproto.BatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		images := make([]*tablecloud.ImageModel, len(req.Items))

		for index, one := range req.Items {
			extensionJson, err := json.MarshalToString(one.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
			images[index] = &tablecloud.ImageModel{
				Vendor:       string(vendor),
				CloudID:      one.CloudID,
				Name:         one.Name,
				Architecture: one.Architecture,
				Platform:     one.Platform,
				OsType:       one.OsType,
				State:        one.State,
				Type:         one.Type,
				Extension:    tabletype.JsonField(extensionJson),
				Creator:      cts.Kit.User,
				Reviser:      cts.Kit.User,
			}
		}
		return svc.dao.Image().BatchCreateWithTx(cts.Kit, txn, images)
	})
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: imageIDs.([]string)}, nil
}
