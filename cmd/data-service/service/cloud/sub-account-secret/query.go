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
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablesass "hcm/pkg/dal/table/cloud/sub-account-secret"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// ListSubAccountSecret list sub account secret
func (svc *subAccountSecretSvc) ListSubAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SubAccountSecretListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}

	result, err := svc.dao.SubAccountSecret().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list sub account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.SubAccountSecretListResult{Count: result.Count}, nil
	}

	details := make([]coresass.BaseSubAccountSecret, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, convTableToBaseSubAccountSecret(one))
	}

	return &protocloud.SubAccountSecretListResult{Details: details}, nil
}

func convTableToBaseSubAccountSecret(one tablesass.Table) coresass.BaseSubAccountSecret {
	return coresass.BaseSubAccountSecret{
		ID:             one.ID,
		Vendor:         one.Vendor,
		Status:         one.Status,
		AccountID:      one.AccountID,
		SubAccountID:   one.SubAccountID,
		CloudCreatedAt: one.CloudCreatedAt,
		DisabledTime:   one.DisabledTime,
		LastUsedTime:   one.LastUsedTime,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// ListSubAccountSecretWithExtension list sub account secret with extension
func (svc *subAccountSecretSvc) ListSubAccountSecretWithExtension(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.SubAccountSecretExtListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}

	result, err := svc.dao.SubAccountSecret().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list sub account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.SubAccountSecretExtListResult[coresass.TCloudSubAccountSecretExtension]{
			Count: result.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		return convSubAccountSecretListResult[coresass.TCloudSubAccountSecretExtension](result.Details)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func convSubAccountSecretListResult[T coresass.Extension](tables []tablesass.Table) (
	*protocloud.SubAccountSecretExtListResult[T], error) {

	details := make([]coresass.SubAccountSecret[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		if len(one.Extension) != 0 {
			if err := json.UnmarshalFromString(string(one.Extension), &extension); err != nil {
				return nil, fmt.Errorf("unmarshal extension failed, err: %v", err)
			}
		}

		details = append(details, coresass.SubAccountSecret[T]{
			BaseSubAccountSecret: convTableToBaseSubAccountSecret(one),
			Extension:            extension,
		})
	}

	return &protocloud.SubAccountSecretExtListResult[T]{
		Details: details,
	}, nil
}

// ListSubAccountSecretJoinExt lists sub account secrets with sub_account and account joined.
// Biz filter: sub_account.bk_biz_ids contains bk_biz_id OR account.bk_biz_id equals bk_biz_id (AND with other filters).
func (svc *subAccountSecretSvc) ListSubAccountSecretJoinExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.SubAccountSecretJoinExtListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.listSecretJoinExtTCloud(cts, req)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter,
			fmt.Errorf("join sub account secret list is not supported for vendor: %s", vendor))
	}
}

// listSecretJoinExtTCloud handles join list for Tencent Cloud with extension.
func (svc *subAccountSecretSvc) listSecretJoinExtTCloud(cts *rest.Contexts,
	req *protocloud.SubAccountSecretJoinExtListReq) (interface{}, error) {

	daoOpt := &types.ListSecretJoinAccountOption{
		Vendor:             enumor.TCloud,
		BkBizID:            req.BkBizID,
		IDs:                req.IDs,
		Status:             req.Status,
		AccountIDs:         req.AccountIDs,
		SubAccountIDs:      req.SubAccountIDs,
		AccountManagers:    req.AccountManagers,
		SubAccountManagers: req.SubAccountManagers,
		Page:               req.Page,
		Extension:          req.Extension,
	}
	result, err := svc.dao.SubAccountSecret().ListJoinAccountAndSubAccount(cts.Kit, daoOpt)
	if err != nil {
		logs.Errorf("list join sub account secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.SubAccountSecretJoinExtListResult{Count: result.Count}, nil
	}

	details := make([]protocloud.SubAccountSecretJoinExtDetail, 0, len(result.Details))
	for i := range result.Details {
		d, err := convJoinRowToListDetailTCloud(result.Details[i])
		if err != nil {
			return nil, err
		}
		details = append(details, *d)
	}

	return &protocloud.SubAccountSecretJoinExtListResult{Details: details}, nil
}

func convJoinRowToListDetailTCloud(row types.SubAccountSecretBizJoinRow) (
	*protocloud.SubAccountSecretJoinExtDetail, error) {

	ext := new(coresass.TCloudSubAccountSecretJoinExtension)
	if len(row.Extension) != 0 {
		if err := json.UnmarshalFromString(string(row.Extension), ext); err != nil {
			return nil, fmt.Errorf("unmarshal secret extension failed, err: %w", err)
		}
	}

	detail := &protocloud.SubAccountSecretJoinExtDetail{
		BaseSubAccountSecret: convTableToBaseSubAccountSecret(row.Table),
		Extension:            ext,
		AccountManagers:      []string(row.AccountManagers),
		AccountName:          row.AccountName,
		SubAccountManagers:   []string(row.SubAccountManagers),
		SubAccountName:       row.SubAccountName,
	}

	// sub_account.extension JSON -> TCloudExtension (includes console_login for tcloud).
	if len(row.SubAccountExtensionJSON) != 0 {
		var saExt coresubaccount.TCloudExtension
		if err := json.UnmarshalFromString(string(row.SubAccountExtensionJSON), &saExt); err != nil {
			return nil, fmt.Errorf("unmarshal sub account extension failed, err: %w", err)
		}
		detail.Extension.ConsoleLogin = saExt.ConsoleLogin
	}

	return detail, nil
}
