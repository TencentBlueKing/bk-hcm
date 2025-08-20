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

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// BizGetAccountUsageBizs 获取账号使用业务列表
func (a *accountSvc) BizGetAccountUsageBizs(cts *rest.Contexts) (any, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("account_id is empty"))
	}
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		logs.Errorf("get bk_biz_id failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bkBizID,
	}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if !authorized {
		return nil, fmt.Errorf("no permission for access account %s", accountID)
	}

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	accounts, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, req: %v, rid: %s", err, listReq, cts.Kit.Rid)
		return nil, err
	}
	if len(accounts.Details) == 0 {
		return nil, fmt.Errorf("account not found: %s", accountID)
	}
	account := accounts.Details[0]
	if !slice.IsItemInSlice(account.UsageBizIDs, bkBizID) &&
		!slice.IsItemInSlice(account.UsageBizIDs, constant.AttachedAllBiz) {
		// 当前业务不在账号的使用业务内，且账号的使用业务非全业务
		return nil, fmt.Errorf("biz %d is not in account %s usage biz list", bkBizID, accountID)
	}
	return account.UsageBizIDs, nil
}

// GetAccountUsageBizs 获取账号使用业务列表
func (a *accountSvc) GetAccountUsageBizs(cts *rest.Contexts) (any, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("account_id is empty"))
	}

	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.Account,
			Action:     meta.Find,
			ResourceID: accountID,
		},
	}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if !authorized {
		return nil, fmt.Errorf("no permission for access account %s", accountID)
	}

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	accounts, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, req: %v, rid: %s", err, listReq, cts.Kit.Rid)
		return nil, err
	}
	if len(accounts.Details) == 0 {
		return nil, fmt.Errorf("account not found: %s", accountID)
	}
	return accounts.Details[0].UsageBizIDs, nil
}
