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

// Package mainaccount ...
package mainaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListMainAccount list main account
func (svc *service) ListMainAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListWithoutFieldReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}

	daoAccountResp, err := svc.dao.MainAccount().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list main account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list main account failed, err: %v", err)
	}
	if req.Page.Count {
		return &dataproto.MainAccountListResult{Count: daoAccountResp.Count}, nil
	}

	details := make([]*protocore.BaseMainAccount, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		details = append(details, &protocore.BaseMainAccount{
			ID:                account.ID,
			Vendor:            enumor.Vendor(account.Vendor),
			CloudID:           account.CloudID,
			Email:             account.Email,
			Managers:          account.Managers,
			BakManagers:       account.BakManagers,
			Site:              enumor.MainAccountSiteType(account.Site),
			BusinessType:      enumor.MainAccountBusinessType(account.BusinessType),
			Status:            enumor.MainAccountStatus(account.Status),
			ParentAccountName: account.ParentAccountName,
			ParentAccountID:   account.ParentAccountID,
			DeptID:            account.DeptID,
			BkBizID:           account.BkBizID,
			OpProductID:       account.OpProductID,
			Memo:              account.Memo,
			Revision: core.Revision{
				Creator:   account.Creator,
				Reviser:   account.Reviser,
				CreatedAt: account.CreatedAt.String(),
				UpdatedAt: account.UpdatedAt.String(),
			},
		})
	}

	return &dataproto.MainAccountListResult{Details: details}, nil
}
