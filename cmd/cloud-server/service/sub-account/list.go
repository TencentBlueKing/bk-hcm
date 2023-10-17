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

package subaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListSubAccount list sub account.
func (svc *service) ListSubAccount(cts *rest.Contexts) (interface{}, error) {
	return svc.listSubAccount(cts, handler.ListResourceAuthRes)
}

func (svc *service) listSubAccount(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SubAccount, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	req.Filter = expr

	result, err := svc.client.DataService().Global.SubAccount.List(cts.Kit, req)
	if err != nil {
		logs.Errorf("request ds to list sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// ListSubAccountExt list sub account.
func (svc *service) ListSubAccountExt(cts *rest.Contexts) (interface{}, error) {
	return svc.listSubAccountExt(cts, handler.ListResourceAuthRes)
}

func (svc *service) listSubAccountExt(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SubAccount, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	expr, err = tools.And(expr, tools.EqualExpression("vendor", vendor))
	if err != nil {
		logs.Errorf("expression append vendor rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req.Filter = expr

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.SubAccount.ListExt(cts.Kit, req)
	case enumor.Aws:
		return svc.client.DataService().Aws.SubAccount.ListExt(cts.Kit, req)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.SubAccount.ListExt(cts.Kit, req)
	case enumor.Azure:
		return svc.client.DataService().Azure.SubAccount.ListExt(cts.Kit, req)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.SubAccount.ListExt(cts.Kit, req)
	default:
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}
}
