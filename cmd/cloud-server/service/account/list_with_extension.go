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
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

func canListAccountExtension(appCode string) error {
	// TODO: 校验App来源, 这里只校验了非Web来请求，需要改造从配置文件读取，允许访问该接口的AppCode白名单（目前暂时可以借助APIGateway的应用认证白名单）
	if appCode == "hcm-web-server" {
		return fmt.Errorf("app[%s] has no permission to list account with extension", appCode)
	}

	return nil
}

// ListWithExtension 该接口返回了Extension，包括了SecretKey信息，只提供给安全使用
func (a *accountSvc) ListWithExtension(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验应用权限
	if err := canListAccountExtension(cts.Kit.AppCode); err != nil {
		return nil, err
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

	return a.client.DataService().Global.Account.ListWithExtension(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   req.Page,
		},
	)
}
