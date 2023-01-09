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

	// TODO: 校验用户是否有List权限，有权限的ID列表

	// FIXME: 由于data-service 不允许空的过滤条件，所以这里构造了一个id>0的永久有效条件，待添加权限ID列表过滤后则可以去除
	reqFilter := req.Filter
	if req.Filter == nil {
		reqFilter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "sync_status", Op: filter.Equal.Factory(), Value: "not_start"},
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
