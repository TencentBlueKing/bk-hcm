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
	hsaccount "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListTCloudAuthPolicies 查询账号授权策略(代理接口)
func (a *accountSvc) ListTCloudAuthPolicies(cts *rest.Contexts) (interface{}, error) {
	req := new(hsaccount.ListTCloudAuthPolicyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Import}}
	if err := a.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("list tcloud auth policies auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return a.client.HCService().TCloud.Account.ListAuthPolicy(cts.Kit, req)
}
