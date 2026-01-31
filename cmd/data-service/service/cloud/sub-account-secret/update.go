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
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablesass "hcm/pkg/dal/table/cloud/sub-account-secret"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// BatchUpdateSubAccountSecret batch update sub account secret.
func (svc *subAccountSecretSvc) BatchUpdateSubAccountSecret(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateForTCloud(svc, cts)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func batchUpdateForTCloud(svc *subAccountSecretSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SubAccountSecretBatchUpdateReq[coresass.TCloudSubAccountSecretExtension])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	models := make([]tablesass.Table, 0, len(req.SubAccountSecrets))
	for _, one := range req.SubAccountSecrets {
		model := tablesass.Table{
			ID:      one.ID,
			Reviser: cts.Kit.User,
		}

		if one.Status != nil {
			model.Status = cvt.PtrToVal(one.Status)
		}
		if one.CloudCreatedAt != nil {
			model.CloudCreatedAt = tabletype.Time(cvt.PtrToVal(one.CloudCreatedAt))
		}
		if one.DisabledTime != nil {
			model.DisabledTime = tabletype.Time(cvt.PtrToVal(one.DisabledTime))
		}
		if one.LastUsedTime != nil {
			model.LastUsedTime = tabletype.Time(cvt.PtrToVal(one.LastUsedTime))
		}

		if one.Extension != nil {
			// 查询原有数据以合并 extension
			dbModel, err := getSubAccountSecretFromTable(cts.Kit, one.ID, svc)
			if err != nil {
				return nil, err
			}

			// 合并覆盖 dbExtension
			updatedExtension, err := json.UpdateMerge(cvt.PtrToVal(one.Extension), string(dbModel.Extension))
			if err != nil {
				return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
			}

			model.Extension = tabletype.JsonField(updatedExtension)
		}

		models = append(models, model)
	}

	if err := svc.dao.SubAccountSecret().BatchUpdate(cts.Kit, models); err != nil {
		logs.Errorf("batch update sub account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// getSubAccountSecretFromTable 从数据库查询子账号密钥
func getSubAccountSecretFromTable(kt *kit.Kit, id string, svc *subAccountSecretSvc) (*tablesass.Table, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	listResult, err := svc.dao.SubAccountSecret().List(kt, opt)
	if err != nil {
		logs.Errorf("list sub account secret failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list sub account secret failed, err: %v", err)
	}

	details := listResult.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list sub account secret failed, sub_account_secret(id=%s) don't exist", id)
	}

	return &details[0], nil
}
