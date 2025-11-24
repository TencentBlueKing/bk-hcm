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

package task

import (
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetOrderTimeCostOverview get order time cost overview
func (s *service) GetOrderTimeCostOverview(cts *rest.Contexts) (any, error) {
	input := new(types.OrderTimeCostReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode order time cost request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid order time cost request, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to apply biz hosts statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetOrderTimeCostOverview(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get order time cost overview, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return types.OrderTimeCostOverviewResp{Details: rst}, nil
}

// GetOrderTimeCostCompare get order time cost compare
func (s *service) GetOrderTimeCostCompare(cts *rest.Contexts) (any, error) {
	input := new(types.OrderTimeCostCompareReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode order time cost compare request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid order time cost compare request, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to apply biz hosts statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetOrderTimeCostCompare(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get order time cost compare, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return rst, nil
}
