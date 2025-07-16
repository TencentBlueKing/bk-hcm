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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListByUsageBizID ...
func (a *accountSvc) ListByUsageBizID(cts *rest.Contexts) (interface{}, error) {
	usageBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	accountType := cts.Request.QueryParameter("account_type")

	// validate biz and authorize
	opt := &handler.ValidWithAuthOption{
		Authorizer: a.authorizer, ResType: meta.Biz,
		Action: meta.Access,
		BasicInfo: &types.CloudResourceBasicInfo{
			ResType: enumor.AccountCloudResType,
			BkBizID: usageBizID,
		}}
	err = handler.BizOperateAuth(cts, opt)
	if err != nil {
		return nil, err
	}

	usageBizIDs := []int64{usageBizID}
	// 关联了所有使用业务的账号也应该被查出来
	if usageBizID != constant.AttachedAllBiz {
		usageBizIDs = append(usageBizIDs, constant.AttachedAllBiz)
	}
	listReq := &protocloud.AccountBizRelWithAccountListReq{
		UsageBizIDs: usageBizIDs,
		AccountType: accountType,
	}

	accounts, err := a.client.DataService().Global.Account.ListAccountBizRelWithAccount(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}
	res := make([]*protocloud.AccountBizRelWithAccount, 0)

	for _, one := range accounts {
		// 仅能查询业务下可用于申请资源的帐号
		if one.Vendor == enumor.Other {
			continue
		}
		// 兼容旧接口的返回值
		one.BkBizIDs = one.UsageBizIDs
		res = append(res, one)
	}
	return res, nil
}
