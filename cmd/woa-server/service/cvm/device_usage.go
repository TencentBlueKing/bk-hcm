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

package cvm

import (
	"errors"

	types "hcm/cmd/woa-server/types/cvm"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/tools/util"
)

const (
	// deviceLoadAchievedThreshold 设备负载利用率达标阈值，兜底默认值
	deviceLoadAchievedThreshold = 30
)

// GetDeviceLoadUsage 查询设备负载利用率信息
func (s *service) GetDeviceLoadUsage(cts *rest.Contexts) (interface{}, error) {
	// 获取路径参数 bk_biz_id
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		logs.Errorf("failed to get bk_biz_id from path, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}

	// 解析请求体
	input := new(types.GetDeviceLoadUsageReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode request body, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验请求参数
	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate get device load usage request, err: %v, input: %+v, rid: %s",
			err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 业务访问鉴权
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("failed to authorize biz access, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用finops接口
	params := &finops.GetDeviceLoadComplianceParam{
		BizID: bkBizID,
		Date:  input.Date,
	}
	result, err := s.finOpsCli.GetDeviceLoadCompliance(cts.Kit, params)
	if err != nil {
		logs.Errorf("failed to get device load compliance, bizID: %d, date: %s, err: %v, rid: %s",
			bkBizID, input.Date, err, cts.Kit.Rid)
		return nil, err
	}

	// 从 global_config 表获取 threshold
	threshold, err := s.getDeviceLoadThreshold(cts.Kit)
	if err != nil {
		logs.Warnf("failed to get device load threshold from global_config, err: %v, rid: %s", err, cts.Kit.Rid)
		// 如果查询失败，使用默认值
		threshold = deviceLoadAchievedThreshold
	}

	// 根据 threshold 和 cpu_usage 判断 achieved_kpi
	achievedKPI := result.CPUUsage >= float64(threshold)
	// 构造返回结果
	rst := &types.GetDeviceLoadUsageResp{
		Threshold:        threshold,
		CPUUsage:         result.CPUUsage,
		AchievedKPI:      achievedKPI,
		EmptyLoadCPUCore: result.EmptyLoadCPUCore,
		EmptyLoadOS:      result.EmptyLoadOS,
		LowLoadCPUCore:   result.LowLoadCPUCore,
		LowLoadOS:        result.LowLoadOS,
	}

	return rst, nil
}

// getDeviceLoadThreshold 从 global_config 表获取设备负载阈值
func (s *service) getDeviceLoadThreshold(kt *kit.Kit) (int64, error) {
	req := core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeDeviceLoad),
			tools.RuleEqual("config_key", enumor.GlobalConfigDeviceLoadThreshold),
		),
		Page: core.NewDefaultBasePage(),
	}

	cfgResp, err := s.client.DataService().Global.GlobalConfig.List(kt, &req)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return 0, err
	}

	if len(cfgResp.Details) == 0 {
		return 0, errors.New("device_load_threshold config not found")
	}

	// 解析 config_value，期望是int64
	threshold, err := util.GetInt64ByInterface(cfgResp.Details[0].ConfigValue)
	if err != nil {
		logs.Errorf("failed to convert device load threshold, err: %v, value: %v, rid: %s", err,
			cfgResp.Details[0].ConfigValue, kt.Rid)
		return 0, err
	}

	if threshold < 0 {
		return 0, errors.New("device_load_threshold must be greater than or equal to 0")
	}

	return threshold, nil
}

// ListDeviceCPUUsageTrend 查询设备CPU利用率趋势信息
func (s *service) ListDeviceCPUUsageTrend(cts *rest.Contexts) (interface{}, error) {
	// 获取路径参数 bk_biz_id
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		logs.Errorf("failed to get bk_biz_id from path, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}

	// 解析请求体
	input := new(types.ListDeviceCPUUsageTrendReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode request body, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验请求参数
	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate list device cpu usage trend request, err: %v, input: %+v, rid: %s",
			err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 业务访问鉴权
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("failed to authorize biz access, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用finops接口
	params := &finops.GetDeviceCPUUsageTrendParam{
		BizID:           bkBizID,
		TimeGranularity: input.TimeGranularity,
		DateRange: &finops.DateRange{
			Start: input.DateRange.Start,
			End:   input.DateRange.End,
		},
	}
	// 设置固定参数
	params.Envs = []enumor.LoadUsageDeviceENV{
		enumor.DeviceENVIDC,
	}
	params.DevTypes = []enumor.FinOpsDeviceType{
		enumor.FinOpsDeviceTypeCVM,
	}
	params.IncludeExamine = []bool{
		true,
	}
	params.IncludeReport = []bool{
		true,
	}

	result, err := s.finOpsCli.GetDeviceCPUUsageTrend(cts.Kit, params)
	if err != nil {
		logs.Errorf("failed to get device cpu usage trend, bizID: %d, err: %v, rid: %s", bkBizID, err,
			cts.Kit.Rid)
		return nil, err
	}

	// 转换返回结果格式，将 Trend 数组包装成结构体返回
	trendData := make([]types.CPUUsageTrendData, 0, len(result.Trend))
	for _, item := range result.Trend {
		trendData = append(trendData, types.CPUUsageTrendData{
			Date:     item.Date,
			CPUUsage: item.CPUUsage,
		})
	}

	return trendData, nil
}
