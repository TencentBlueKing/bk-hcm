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

package mainaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// List list main account with options
func (s *service) List(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListWithoutFieldReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户是否有查看权限，有权限的ID列表
	accountIDs, isAny, err := s.listAuthorized(cts, meta.Find, meta.MainAccount)
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
		reqFilter = tools.ExpressionAnd(tools.RuleIn("id", accountIDs))

		// 加上请求里过滤条件
		if req.Filter != nil && !req.Filter.IsEmpty() {
			reqFilter.Rules = append(reqFilter.Rules, req.Filter)
		}
	}

	accounts, err := s.client.DataService().Global.MainAccount.List(
		cts.Kit,
		&core.ListReq{
			Filter: reqFilter,
			Page:   req.Page,
		},
	)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (s *service) listAuthorized(cts *rest.Contexts, action meta.Action,
	typ meta.ResourceType) ([]string, bool, error) {

	resources, err := s.authorizer.ListAuthorizedInstances(cts.Kit, &meta.ListAuthResInput{Type: typ,
		Action: action})
	if err != nil {
		return []string{}, false, errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("list account of %s permission failed, err: %v", action, err),
		)
	}

	return resources.IDs, resources.IsAny, err

}
