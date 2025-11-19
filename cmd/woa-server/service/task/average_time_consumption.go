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

// GetAverageTimeConsumptionOverview get average time consumption overview
func (s *service) GetAverageTimeConsumptionOverview(cts *rest.Contexts) (any, error) {
	input := new(types.AverageTimeConsumptionReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode average time consumption request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid average time consumption request, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to apply biz hosts statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetAverageTimeConsumptionOverview(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get average time consumption overview, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return types.AverageTimeConsumptionOverviewResp{Details: rst}, nil
}

// GetAverageTimeConsumptionCompare get average time consumption compare
func (s *service) GetAverageTimeConsumptionCompare(cts *rest.Contexts) (any, error) {
	input := new(types.AverageTimeConsumptionCompareReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode average time consumption compare request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid average time consumption compare request, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to apply biz hosts statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetAverageTimeConsumptionCompare(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get average time consumption compare, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return rst, nil
}

// GetPercentileTimeConsumptionOverview get percentile time consumption overview
func (s *service) GetPercentileTimeConsumptionOverview(cts *rest.Contexts) (any, error) {
	input := new(types.PercentileTimeConsumptionReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode percentile time consumption request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid percentile time consumption request, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to apply biz hosts statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetPercentileTimeConsumptionOverview(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get percentile time consumption overview, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return types.PercentileTimeConsumptionOverviewResp{Details: rst}, nil
}

// GetPercentileTimeConsumptionCompare get percentile time consumption compare
func (s *service) GetPercentileTimeConsumptionCompare(cts *rest.Contexts) (any, error) {
	input := new(types.PercentileTimeConsumptionCompareReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode percentile time consumption compare request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid percentile time consumption compare request, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to apply biz hosts statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetPercentileTimeConsumptionCompare(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get percentile time consumption compare, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return rst, nil
}

// GetDeliveryRateStatistics get delivery rate statistics
func (s *service) GetDeliveryRateStatistics(cts *rest.Contexts) (any, error) {
	input := new(types.DeliveryRateStatisticsReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode delivery rate statistics request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid delivery rate statistics request, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to get delivery rate statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetDeliveryRateStatistics(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get delivery rate statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return types.DeliveryRateStatisticsResp{Details: rst}, nil
}

// GetDeliveryRateDetail get delivery rate detail
func (s *service) GetDeliveryRateDetail(cts *rest.Contexts) (any, error) {
	input := new(types.DeliveryRateDetailReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode delivery rate detail request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if errKey, err := input.Validate(); err != nil {
		logs.Errorf("invalid delivery rate detail request, err: %v, errKey: %s, rid: %s", err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanResDeliverAnalyze, Action: meta.Find},
	}); err != nil {
		logs.Errorf("no permission to get delivery rate detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Operation().GetDeliveryRateDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get delivery rate detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return rst, nil
}
