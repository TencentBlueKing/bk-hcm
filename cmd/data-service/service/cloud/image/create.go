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
		return batchCreateImageExt[dataproto.TCloudImageExtensionCreateReq](cts, svc, vendor)
	case enumor.Aws:
		return batchCreateImageExt[dataproto.AwsImageExtensionCreateReq](cts, svc, vendor)
	case enumor.Gcp:
		return batchCreateImageExt[dataproto.GcpImageExtensionCreateReq](cts, svc, vendor)
	case enumor.HuaWei:
		return batchCreateImageExt[dataproto.HuaWeiImageExtensionCreateReq](cts, svc, vendor)
	case enumor.Azure:
		return batchCreateImageExt[dataproto.AzureImageExtensionCreateReq](cts, svc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchCreateImageExt[T dataproto.ImageExtensionCreateReq](
	cts *rest.Contexts,
	svc *imageSvc,
	vendor enumor.Vendor,
) (interface{}, error) {
	req := new(dataproto.ImageExtBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		images := make([]*tablecloud.ImageModel, len(*req))

		for indx, imageReq := range *req {
			extensionJson, err := json.MarshalToString(imageReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
			images[indx] = &tablecloud.ImageModel{
				Vendor:       string(vendor),
				CloudID:      imageReq.CloudID,
				Name:         imageReq.Name,
				Architecture: imageReq.Architecture,
				Platform:     imageReq.Platform,
				State:        imageReq.State,
				Type:         imageReq.Type,
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
