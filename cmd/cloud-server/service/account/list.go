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

package account

import (
	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListAccount ...
func (a *accountSvc) ListAccount(cts *rest.Contexts) (interface{}, error) {
	return a.list(cts, meta.Account)
}

// ResourceList ...
func (a *accountSvc) ResourceList(cts *rest.Contexts) (interface{}, error) {
	return a.list(cts, meta.CloudResource)
}

func (a *accountSvc) list(cts *rest.Contexts, typ meta.ResourceType) (interface{}, error) {
	req := new(proto.AccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户是否有查看权限，有权限的ID列表
	accountIDs, isAny, err := a.listAuthorized(cts, meta.Find, typ)
	if err != nil {
		return nil, err
	}
	// 无任何账号权限
	if len(accountIDs) == 0 && !isAny {
		return []map[string]interface{}{}, nil
	}

	// 构造权限过滤条件
	var reqFilter *filter.Expression
	if isAny {
		reqFilter = req.Filter
	} else {
		reqFilter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: accountIDs},
			},
		}
		// 加上请求里过滤条件
		if req.Filter != nil && !req.Filter.IsEmpty() {
			reqFilter.Rules = append(reqFilter.Rules, req.Filter)
		}
	}

	accounts, err := a.client.DataService().Global.Account.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   req.Page,
		},
	)
	if err != nil {
		return nil, err
	}

	for _, one := range accounts.Details {
		status, failedReason, err := a.getAccountSyncDetail(cts, one.ID, string(one.Vendor))
		if err != nil {
			return nil, err
		}
		one.SyncStatus = status
		one.SyncFailedReason = failedReason
		one.RecycleReserveTime = convertRecycleReverseTime(one.RecycleReserveTime)
	}

	return accounts, nil
}

func (a *accountSvc) getAccountSyncDetail(cts *rest.Contexts, accountID string,
	vendor string) (string, string, error) {

	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: accountID,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	accountSyncDetail, err := a.client.DataService().Global.AccountSyncDetail.List(cts.Kit, listReq)
	if err != nil {
		return "", "", err
	}

	status := ""
	failedReason := ""
	for _, one := range accountSyncDetail.Details {
		status = one.ResStatus
		if one.ResStatus == string(enumor.SyncFailed) {
			failedReason = string(one.ResFailedReason)
			break
		}
	}

	return status, failedReason, nil
}
