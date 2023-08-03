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

package subaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablesubaccount "hcm/pkg/dal/table/cloud/sub-account"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// GetSubAccount get sub account with extension.
func (svc *service) GetSubAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID := cts.PathParameter("id").String()

	// 查询账号信息
	dbAccount, err := svc.dao.SubAccount().Get(cts.Kit, accountID)
	if err != nil {
		logs.Errorf("get sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	var account interface{}
	switch dbAccount.Vendor {
	case enumor.TCloud:
		account, err = convCoreSubAccount[coresubaccount.TCloudExtension](dbAccount)
	case enumor.Aws:
		account, err = convCoreSubAccount[coresubaccount.AwsExtension](dbAccount)
	case enumor.HuaWei:
		account, err = convCoreSubAccount[coresubaccount.HuaWeiExtension](dbAccount)
	case enumor.Gcp:
		account, err = convCoreSubAccount[coresubaccount.GcpExtension](dbAccount)
	case enumor.Azure:
		account, err = convCoreSubAccount[coresubaccount.AzureExtension](dbAccount)
	}

	if err != nil {
		logs.Errorf("conv core sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("conv core sub account failed, err: %v", err)
	}

	return account, nil
}

func convCoreSubAccount[T coresubaccount.Extension](db *tablesubaccount.Table) (*coresubaccount.SubAccount[T], error) {
	extension := new(T)
	if len(db.Extension) != 0 {
		err := json.UnmarshalFromString(string(db.Extension), extension)
		if err != nil {
			return nil, fmt.Errorf("unmarshal sub account extension failed, err: %v", err)
		}
	}

	return &coresubaccount.SubAccount[T]{
		BaseSubAccount: convCoreBaseSubAccount(*db),
		Extension:      extension,
	}, nil
}

// ListSubAccount list sub account.
func (svc *service) ListSubAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	daoAccountResp, err := svc.dao.SubAccount().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list sub account failed, err: %v", err)
	}
	if req.Page.Count {
		return &dssubaccount.ListResult{Count: daoAccountResp.Count}, nil
	}

	details := make([]coresubaccount.BaseSubAccount, 0, len(daoAccountResp.Details))
	for _, one := range daoAccountResp.Details {
		details = append(details, convCoreBaseSubAccount(one))
	}

	return &dssubaccount.ListResult{Details: details}, nil
}

func convCoreBaseSubAccount(one tablesubaccount.Table) coresubaccount.BaseSubAccount {
	return coresubaccount.BaseSubAccount{
		ID:        one.ID,
		CloudID:   one.CloudID,
		Name:      one.Name,
		Vendor:    one.Vendor,
		Site:      one.Site,
		AccountID: one.AccountID,
		Managers:  one.Managers,
		BkBizIDs:  one.BkBizIDs,
		Memo:      one.Memo,
		Revision: core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// ListSubAccountExt list account with extension.
func (svc *service) ListSubAccountExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.SubAccount().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list sub account failed, err: %v", err)
	}

	if req.Page.Count {
		return &dssubaccount.ListExtResult[coresubaccount.TCloudExtension]{Count: result.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		return convListExtResult[coresubaccount.TCloudExtension](result.Details)
	case enumor.Aws:
		return convListExtResult[coresubaccount.AwsExtension](result.Details)
	case enumor.HuaWei:
		return convListExtResult[coresubaccount.HuaWeiExtension](result.Details)
	case enumor.Azure:
		return convListExtResult[coresubaccount.AzureExtension](result.Details)
	case enumor.Gcp:
		return convListExtResult[coresubaccount.GcpExtension](result.Details)

	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convListExtResult[T coresubaccount.Extension](models []tablesubaccount.Table) (
	*dssubaccount.ListExtResult[T], error) {

	details := make([]coresubaccount.SubAccount[T], 0, len(models))
	for _, one := range models {
		account, err := convCoreSubAccount[T](&one)
		if err != nil {
			return nil, err
		}

		details = append(details, *account)
	}

	return &dssubaccount.ListExtResult[T]{
		Details: details,
	}, nil
}
