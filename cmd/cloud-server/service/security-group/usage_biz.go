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

package securitygroup

import (
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BizListSGMaintainerInfos list security group maintainer information.
func (svc *securityGroupSvc) BizListSGMaintainerInfos(cts *rest.Contexts) (interface{}, error) {

	req := new(cloudserver.ListSGMaintainerInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authFilter, noPerm, err := handler.ListBizAuthRes(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "no permission for list sg usage biz maintainers")
	}

	securityGroups, err := svc.listSecurityGroupByIDsAndFilter(cts.Kit, req.SecurityGroupIDs, authFilter)
	if err != nil {
		logs.Errorf("list security group by ids failed, err: %v, ids: %v, rid: %s", err, req.SecurityGroupIDs, cts.Kit.Rid)
		return nil, err
	}
	sgMap := make(map[string]cloud.BaseSecurityGroup, len(securityGroups))
	usageBizIDs := make([]int64, 0, len(securityGroups))
	accountIDs := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.BkBizID != constant.UnassignedBiz {
			logs.Errorf("security group %s has been assigned to biz %d, rid: %s", sg.ID, sg.BkBizID, cts.Kit.Rid)
			return nil, fmt.Errorf("security group %s has been assigned to biz %d", sg.ID, sg.BkBizID)
		}
		usageBizIDs = append(usageBizIDs, sg.UsageBizIDs...)
		accountIDs = append(accountIDs, sg.AccountID)
		sgMap[sg.ID] = sg
	}

	accountMap, err := svc.listAccountMapByIDs(cts.Kit, accountIDs)
	if err != nil {
		logs.Errorf("list account by ids failed, err: %v, ids: %v, rid: %s", err, accountIDs, cts.Kit.Rid)
		return nil, err
	}

	bizMap, err := svc.searchBusinessByBizIDs(cts.Kit, usageBizIDs)
	if err != nil {
		logs.Errorf("search business by biz ids failed, err: %v, ids: %v, rid: %s", err, usageBizIDs, cts.Kit.Rid)
		return nil, err
	}

	return buildSGMaintainerInfoResult(cts.Kit, req.SecurityGroupIDs, sgMap, bizMap, accountMap)
}

func (svc *securityGroupSvc) listAccountMapByIDs(kt *kit.Kit, accountIDs []string) (
	map[string]*cloud.BaseAccount, error) {

	accountIDs = slice.Unique(accountIDs)
	accountMap := make(map[string]*cloud.BaseAccount, len(accountIDs))
	for _, ids := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &proto.AccountListReq{
			Filter: tools.ContainersExpression("id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}
		for _, account := range resp.Details {
			accountMap[account.ID] = account
		}
	}
	for _, id := range accountIDs {
		if _, ok := accountMap[id]; !ok {
			logs.Errorf("account %s not found, rid: %s", id, kt.Rid)
			return nil, fmt.Errorf("account %s not found", id)
		}
	}

	return accountMap, nil
}

// buildSGMaintainerInfoResult 构建安全组使用业务运维列表结果
func buildSGMaintainerInfoResult(kt *kit.Kit, sgIDs []string, sgMap map[string]cloud.BaseSecurityGroup,
	bizMap map[int64]cmdb.Biz, accountMap map[string]*cloud.BaseAccount) (
	[]*cloudserver.ListSGMaintainerInfoResult, error) {

	result := make([]*cloudserver.ListSGMaintainerInfoResult, 0, len(sgIDs))
	for _, sgID := range sgIDs {
		sg, ok := sgMap[sgID]
		if !ok {
			logs.Errorf("security group %s not found, rid: %s", sgID, kt.Rid)
			return nil, fmt.Errorf("security group %s not found", sgID)
		}
		account, ok := accountMap[sg.AccountID]
		if !ok {
			logs.Errorf("account %s not found, rid: %s", sg.AccountID, kt.Rid)
			return nil, fmt.Errorf("account %s not found", sg.AccountID)
		}
		one := &cloudserver.ListSGMaintainerInfoResult{
			ID:       sgID,
			Managers: account.Managers,
		}
		for _, usageBizID := range sg.UsageBizIDs {
			biz, ok := bizMap[usageBizID]
			if !ok {
				logs.Errorf("business %d not found, rid: %s", usageBizID, kt.Rid)
				return nil, fmt.Errorf("business %d not found", usageBizID)
			}
			one.UsageBizInfos = append(one.UsageBizInfos, biz)
		}
		result = append(result, one)
	}
	return result, nil
}

// searchBusinessByBizIDs searches business information by business IDs.
func (svc *securityGroupSvc) searchBusinessByBizIDs(kt *kit.Kit, bizIDs []int64) (map[int64]cmdb.Biz, error) {
	bizMap := make(map[int64]cmdb.Biz)
	for _, ids := range slice.Split(slice.Unique(bizIDs), int(core.DefaultMaxPageLimit)) {
		param := &cmdb.SearchBizParams{
			Fields: []string{"bk_biz_id", "bk_biz_name", "bk_biz_maintainer"},
			BizPropertyFilter: &cmdb.QueryFilter{
				Rule: &cmdb.CombinedRule{
					Condition: cmdb.ConditionAnd,
					Rules: []cmdb.Rule{
						&cmdb.AtomRule{
							Field:    "bk_biz_id",
							Operator: cmdb.OperatorIn,
							Value:    ids,
						},
					},
				},
			},
		}

		business, err := svc.cmdbClient.SearchBusiness(kt, param)
		if err != nil {
			logs.Errorf("search business failed, err: %v, req: %v, rid: %s", err, param, kt.Rid)
			return nil, err
		}
		for _, biz := range business.Info {
			bizMap[biz.BizID] = biz
		}
	}

	return bizMap, nil
}
