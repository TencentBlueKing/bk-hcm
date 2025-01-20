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
	"fmt"

	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	coresync "hcm/pkg/api/core/cloud/sync"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// ListAccount ...
func (a *accountSvc) ListAccount(cts *rest.Contexts) (interface{}, error) {
	return a.listAccount(cts, meta.Account)
}

// ResourceList ...
func (a *accountSvc) ResourceList(cts *rest.Contexts) (interface{}, error) {
	return a.listResource(cts, meta.CloudResource)
}

func (a *accountSvc) listResource(cts *rest.Contexts, typ meta.ResourceType) (interface{}, error) {
	req := new(proto.AccountListResourceReq)
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

	listReq := &dataproto.AccountListReq{
		Filter: reqFilter,
		Page:   req.Page,
	}
	accounts, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	respIDs := make([]string, 0, len(accounts.Details))
	for _, one := range accounts.Details {
		respIDs = append(respIDs, one.ID)
	}
	accountDetailsMap, err := a.getAccountsSyncDetail(cts, respIDs...)
	if err != nil {
		logs.Errorf("get account sync detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	for _, one := range accounts.Details {
		for _, detail := range accountDetailsMap[one.ID] {
			one.SyncStatus = detail.ResStatus
			if detail.ResStatus == string(enumor.SyncFailed) {
				one.SyncFailedReason = string(detail.ResFailedReason)
				break
			}
		}
		one.RecycleReserveTime = convertRecycleReverseTime(one.RecycleReserveTime)
	}

	return accounts, nil
}

func (a *accountSvc) listAccount(cts *rest.Contexts, typ meta.ResourceType) (*dataproto.AccountListResult, error) {
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

	if isAny {
		accounts, err := a.listAccountByFilter(cts.Kit, req.Filter)
		if err != nil {
			logs.Errorf("list account by filter failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		err = a.fillAccountSyncDetail(cts, accounts)
		if err != nil {
			logs.Errorf("fill account sync detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return &dataproto.AccountListResult{
			Details: accounts,
			Count:   uint64(len(accounts)),
		}, nil
	}

	bizAccounts, err := a.listAccountByBiz(cts)
	if err != nil {
		logs.Errorf("list account by biz failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	accountIDs = append(accountIDs, bizAccounts...)
	accountIDs = slice.Unique(accountIDs)

	// 构造权限过滤条件
	accounts := make([]*cloud.BaseAccount, 0)
	for _, ids := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		innerFilter := tools.ExpressionOr(
			tools.RuleJSONContains("managers", cts.Kit.User),
			tools.RuleIn("id", ids),
		)
		// 加上请求里过滤条件
		var reqFilter *filter.Expression
		if req.Filter != nil && !req.Filter.IsEmpty() {
			reqFilter, err = tools.And(innerFilter, req.Filter)
			if err != nil {
				logs.Errorf("merge filter failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		} else {
			reqFilter = innerFilter
		}
		accountList, err := a.listAccountByFilter(cts.Kit, reqFilter)
		if err != nil {
			logs.Errorf("list account by filter failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		accounts = append(accounts, accountList...)
	}

	err = a.fillAccountSyncDetail(cts, accounts)
	if err != nil {
		logs.Errorf("fill account sync detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &dataproto.AccountListResult{
		Details: accounts,
		Count:   uint64(len(accounts)),
	}, nil
}

// fillAccountSyncDetail 补全同步状态信息
func (a *accountSvc) fillAccountSyncDetail(cts *rest.Contexts, accounts []*cloud.BaseAccount) error {
	syncAccountMap := make(map[string]*cloud.BaseAccount)
	for _, one := range accounts {
		if one.Type == enumor.RegistrationAccount {
			continue
		}
		syncAccountMap[one.ID] = one
	}
	if len(syncAccountMap) == 0 {
		return nil
	}
	syncDetails, err := a.getAccountsSyncDetail(cts, converter.MapKeyToSlice(syncAccountMap)...)
	if err != nil {
		logs.Errorf("get account sync detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	for _, one := range syncAccountMap {
		for _, detail := range syncDetails[one.ID] {
			one.SyncStatus = detail.ResStatus
			if detail.ResStatus == string(enumor.SyncFailed) {
				one.SyncFailedReason = string(detail.ResFailedReason)
				break
			}
		}
		one.RecycleReserveTime = convertRecycleReverseTime(one.RecycleReserveTime)
	}
	return nil
}

func (a *accountSvc) listAccountByFilter(kt *kit.Kit, reqFilter *filter.Expression) ([]*cloud.BaseAccount, error) {
	page := &core.BasePage{
		Count: false,
		Start: 0,
		Limit: core.DefaultMaxPageLimit,
		Sort:  "id",
	}
	accounts := make([]*cloud.BaseAccount, 0)
	for {
		listReq := &dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   page,
		}
		resp, err := a.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, req: %v, rid: %s", err, listReq, kt.Rid)
			return nil, err
		}
		if len(resp.Details) == 0 {
			break
		}
		accounts = append(accounts, resp.Details...)
		page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return accounts, nil
}

// listAccountByBiz 根据账号可见业务查询账号列表
func (a *accountSvc) listAccountByBiz(cts *rest.Contexts) ([]string, error) {
	bizIDs, _, err := a.listAuthorized(cts, meta.Access, meta.Biz)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]struct{})
	for _, ids := range slice.Split(bizIDs, int(core.DefaultMaxPageLimit)) {

		intIDs := converter.StringSliceToInt64Slice(ids)
		offset := uint32(0)
		for {
			listReq := &core.ListReq{
				Filter: tools.ExpressionAnd(
					tools.RuleIn("bk_biz_id", intIDs),
				),
				Page: &core.BasePage{
					Count: false,
					Start: offset,
					Limit: core.DefaultMaxPageLimit,
				},
			}
			resp, err := a.client.DataService().Global.Account.ListAccountBizRel(cts.Kit.Ctx, cts.Kit.Header(), listReq)
			if err != nil {
				logs.Errorf("list account biz relation failed, err: %v, req: %v, rid: %s", err, listReq, cts.Kit.Rid)
				return nil, err
			}
			if len(resp.Details) == 0 {
				break
			}

			for _, detail := range resp.Details {
				resultMap[detail.AccountID] = struct{}{}
			}
			offset += uint32(core.DefaultMaxPageLimit)
		}
	}

	return converter.MapKeyToSlice(resultMap), nil
}

func getSliceByPage[T any](data []T, page *core.BasePage) []T {
	length := len(data)
	if length == 0 {
		return []T{}
	}
	// safe slice
	begin := min(int(page.Start), length)
	end := min(length, int(page.Start)+int(page.Limit))
	return data[begin:end]
}

func (a *accountSvc) getAccountsSyncDetail(cts *rest.Contexts, accountIDs ...string) (
	map[string][]coresync.AccountSyncDetailTable, error) {

	if len(accountIDs) == 0 {
		return nil, fmt.Errorf("accountIDs is empty")
	}

	result := make(map[string][]coresync.AccountSyncDetailTable)
	for _, ids := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("account_id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		accountSyncDetail, err := a.client.DataService().Global.AccountSyncDetail.List(cts.Kit, listReq)
		if err != nil {
			logs.Errorf("list account sync detail failed, err: %v, req: %v, rid: %s", err, listReq, cts.Kit.Rid)
			return nil, err
		}
		for _, detail := range accountSyncDetail.Details {
			result[detail.AccountID] = append(result[detail.AccountID], detail)
		}
	}

	return result, nil
}
