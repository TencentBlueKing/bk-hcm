/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package task

import (
	configtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetApplyStatistics get apply operation statistics
func (s *service) GetApplyStatistics(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyStatReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get resource apply operation statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Operation().GetApplyStatistics(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateApplyOrderStatisticsConfig 创建申请单统计配置
func (s *service) CreateApplyOrderStatisticsConfig(cts *rest.Contexts) (any, error) {
	input := new(configtypes.CreateApplyOrderStatisticsConfigParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode create apply order statistics config request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("invalid create apply order statistics config request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.configLogics.ApplyOrderStatistics().CreateConfig(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to create apply order statistics config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateApplyOrderStatisticsConfig 更新申请单统计配置
func (s *service) UpdateApplyOrderStatisticsConfig(cts *rest.Contexts) (any, error) {
	input := new(configtypes.UpdateApplyOrderStatisticsConfigParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode update apply order statistics config request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("invalid update apply order statistics config request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	err := s.configLogics.ApplyOrderStatistics().UpdateConfig(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to update apply order statistics config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListApplyOrderStatisticsConfig 查询指定月份的配置列表
func (s *service) ListApplyOrderStatisticsConfig(cts *rest.Contexts) (any, error) {
	input := new(configtypes.ListApplyOrderStatisticsConfigParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode list apply order statistics config request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("invalid list apply order statistics config request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.configLogics.ApplyOrderStatistics().ListConfig(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to list apply order statistics config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// ListApplyOrderStatisticsYearMonths 查询配置表中的月份列表
func (s *service) ListApplyOrderStatisticsYearMonths(cts *rest.Contexts) (any, error) {
	rst, err := s.configLogics.ApplyOrderStatistics().ListYearMonths(cts.Kit)
	if err != nil {
		logs.Errorf("failed to list apply order statistics year months, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
