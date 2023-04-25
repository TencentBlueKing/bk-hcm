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

package eip

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecloud "hcm/pkg/dal/table/cloud/eip"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateEipExt ...
func (svc *eipSvc) BatchCreateEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchCreateEipExt[dataproto.TCloudEipExtensionCreateReq](cts, svc, vendor)
	case enumor.Aws:
		return batchCreateEipExt[dataproto.AwsEipExtensionCreateReq](cts, svc, vendor)
	case enumor.Gcp:
		return batchCreateEipExt[dataproto.GcpEipExtensionCreateReq](cts, svc, vendor)
	case enumor.HuaWei:
		return batchCreateEipExt[dataproto.HuaWeiEipExtensionCreateReq](cts, svc, vendor)
	case enumor.Azure:
		return batchCreateEipExt[dataproto.AzureEipExtensionCreateReq](cts, svc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchCreateEipExt[T dataproto.EipExtensionCreateReq](
	cts *rest.Contexts,
	svc *eipSvc,
	vendor enumor.Vendor,
) (interface{}, error) {
	req := new(dataproto.EipExtBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	eipIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		eips := make([]*tablecloud.EipModel, len(*req))
		for idx, eipReq := range *req {
			extensionJson, err := json.MarshalToString(eipReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			bkBizID := int64(constant.UnassignedBiz)
			if eipReq.BkBizID > 0 {
				bkBizID = eipReq.BkBizID
			}
			eips[idx] = &tablecloud.EipModel{
				Vendor:    string(vendor),
				AccountID: eipReq.AccountID,
				CloudID:   eipReq.CloudID,
				BkBizID:   bkBizID,
				Name:      eipReq.Name,
				Region:    eipReq.Region,
				Status:    eipReq.Status,
				PublicIp:  eipReq.PublicIp,
				PrivateIp: eipReq.PrivateIp,
				Extension: tabletype.JsonField(extensionJson),
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			}
		}
		return svc.dao.Eip().BatchCreateWithTx(cts.Kit, txn, eips)
	})
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: eipIDs.([]string)}, nil
}
