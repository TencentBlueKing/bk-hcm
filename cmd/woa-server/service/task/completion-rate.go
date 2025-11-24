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

// GetCompletionRateStatistics get completion rate statistics
func (s *service) GetCompletionRateStatistics(cts *rest.Contexts) (any, error) {
	input := new(types.GetCompletionRateStatReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get completion rate statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get completion rate statistics, err: %v,rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	})
	if err != nil {
		logs.Errorf("no permission to get completion rate statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetCompletionRateStatistics(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get completion rate statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetCompletionRateDetail 获取结单率详情统计
func (s *service) GetCompletionRateDetail(cts *rest.Contexts) (any, error) {
	input := new(types.GetCompletionRateDetailReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get completion rate detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get completion rate detail, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	})
	if err != nil {
		logs.Errorf("no permission to get completion rate detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetCompletionRateDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get completion rate detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
