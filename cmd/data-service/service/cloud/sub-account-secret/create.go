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

package subaccountsecret

import (
	"fmt"

	"hcm/pkg/api/core"
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablesass "hcm/pkg/dal/table/cloud/sub-account-secret"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateSubAccountSecret batch create sub account secret.
func (svc *subAccountSecretSvc) BatchCreateSubAccountSecret(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateForTCloud(vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func batchCreateForTCloud(vendor enumor.Vendor, svc *subAccountSecretSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SubAccountSecretBatchCreateReq[coresass.TCloudSubAccountSecretExtension])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablesass.Table, 0, len(req.SubAccountSecrets))
		for _, one := range req.SubAccountSecrets {
			extensionJson, err := json.MarshalToString(cvt.PtrToVal(one.Extension))
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
			model := tablesass.Table{
				Vendor:       vendor,
				Status:       one.Status,
				Extension:    tabletype.JsonField(extensionJson),
				AccountID:    one.AccountID,
				SubAccountID: one.SubAccountID,
				Creator:      cts.Kit.User,
				Reviser:      cts.Kit.User,
			}
			if one.CloudCreatedAt != "" {
				model.CloudCreatedAt = tabletype.Time(one.CloudCreatedAt)
			}
			if one.DisabledTime != "" {
				model.DisabledTime = tabletype.Time(one.DisabledTime)
			}
			if one.LastUsedTime != "" {
				model.LastUsedTime = tabletype.Time(one.LastUsedTime)
			}

			models = append(models, model)
		}

		ids, err := svc.dao.SubAccountSecret().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("batch create sub account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return &core.BatchCreateResult{IDs: ids}, nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
