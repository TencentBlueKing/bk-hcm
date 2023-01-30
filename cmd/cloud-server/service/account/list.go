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
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// List ...
func (a *accountSvc) List(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户是否有查看权限，有权限的ID列表
	accountIDs, isAny, err := a.listAuthorized(cts, meta.Find)
	if err != nil {
		return nil, err
	}
	// 无任何账号权限
	if len(accountIDs) == 0 && !isAny {
		return []map[string]interface{}{}, nil
	}

	// FIXME: 由于底层filter暂时不支持嵌套，所以这里暂时忽略来着前端请求的req.Filter，后续再支持
	var reqFilter *filter.Expression
	if isAny {
		reqFilter = &filter.Expression{
			Op:    filter.Or,
			Rules: []filter.RuleFactory{},
		}
	} else {
		reqFilter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: accountIDs},
			},
		}
	}

	return a.client.DataService().Global.Account.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   req.Page,
		},
	)
}
