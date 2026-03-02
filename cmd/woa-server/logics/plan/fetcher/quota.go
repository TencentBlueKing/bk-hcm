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

package fetcher

import (
	"errors"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tablegconf "hcm/pkg/dal/table/global-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"
)

// GetPlanTransferQuotaConfigs 获取预测转移额度配置
func (f *ResPlanFetcher) GetPlanTransferQuotaConfigs(kt *kit.Kit) (ptypes.TransferQuotaConfig, error) {
	dbConfigs, err := f.GetConfigsFromData(kt, constant.ResourcePlanTransferKey)
	if err != nil {
		logs.Errorf("failed to get plan transfer quota configs from db, err: %v, rid: %s", err, kt.Rid)
		return ptypes.TransferQuotaConfig{}, err
	}

	configMap := cvt.SliceToMap(dbConfigs, func(t tablegconf.GlobalConfigTable) (string, interface{}) {
		return t.ConfigKey, t.ConfigValue
	})

	config := ptypes.TransferQuotaConfig{
		Quota:      f.resPlanCfg.RefreshTransferQuota.Quota,      // 预测转移额度
		AuditQuota: f.resPlanCfg.RefreshTransferQuota.AuditQuota, // 预测转移审批额度
	}
	if v, ok := configMap[constant.TransferQuotaKey]; ok {
		config.Quota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert biz quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return ptypes.TransferQuotaConfig{}, err
		}
	}

	if v, ok := configMap[constant.TransferAuditQuotaKey]; ok {
		config.AuditQuota, err = util.GetInt64ByInterface(v)
		if err != nil {
			logs.Warnf("failed to convert ieg quota, err: %v, rid: %s, value: %v", err, kt.Rid, v)
			return ptypes.TransferQuotaConfig{}, err
		}
	}

	return config, nil
}

// GetConfigsFromData 从数据层global_config获取配置
func (f *ResPlanFetcher) GetConfigsFromData(kt *kit.Kit, configType string) ([]tablegconf.GlobalConfigTable, error) {
	filter := tools.ExpressionAnd(tools.RuleEqual("config_type", configType))

	dataReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
	}
	dataResp, err := f.client.DataService().Global.GlobalConfig.List(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, *dataReq, kt.Rid)
		return nil, err
	}

	return dataResp.Details, nil
}

// ListRemainTransferQuota list remain transfer quota.
func (f *ResPlanFetcher) ListRemainTransferQuota(kt *kit.Kit, req *ptypes.ListResPlanTransferQuotaSummaryReq) (
	*ptypes.ResPlanTransferQuotaSummaryResp, error) {

	// 不带业务ID的查询已使用额度
	appliedRules := []*filter.AtomRule{tools.RuleEqual("year", req.Year)}
	if len(req.SubTicketID) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("sub_ticket_id", req.SubTicketID))
	}
	if len(req.TechnicalClass) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("technical_class", req.TechnicalClass))
	}
	if len(req.ObsProject) > 0 {
		appliedRules = append(appliedRules, tools.RuleIn("obs_project", req.ObsProject))
	}

	// 带业务ID的查询已使用额度
	bizRules := make([]*filter.AtomRule, len(appliedRules))
	copy(bizRules, appliedRules)
	if len(req.BkBizIDs) > 0 {
		bizRules = append(bizRules, tools.RuleIn("bk_biz_id", req.BkBizIDs))
	}

	// 查询当前业务-已使用的额度
	bizTarReq := &rpproto.TransferAppliedRecordListReq{ListReq: core.ListReq{
		Filter: tools.ExpressionAnd(bizRules...),
		Page:   core.NewDefaultBasePage(),
	}}
	bizAppliedQuota, err := f.client.DataService().Global.ResourcePlan.SumResPlanTransferAppliedRecord(kt, bizTarReq)
	if err != nil {
		logs.Errorf("failed to list res plan transfer applied record quota, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	// 查询CRP侧的预测额度
	demands, err := f.QueryCRPTransferPoolDemands(kt, req.ObsProject, req.TechnicalClass)
	if err != nil {
		logs.Errorf("failed to query ieg demands, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var remainQuota int64
	for _, demand := range demands {
		remainQuota += demand.CoreAmount
	}

	return &ptypes.ResPlanTransferQuotaSummaryResp{
		UsedQuota:   bizAppliedQuota.SumAppliedCore,
		RemainQuota: remainQuota,
	}, nil
}

// QueryCRPTransferPoolDemands 查询预测中转池中剩余CRP预测需求数
func (f *ResPlanFetcher) QueryCRPTransferPoolDemands(kt *kit.Kit, obsProjects []enumor.ObsProject,
	technicalClasses []string) ([]*cvmapi.CvmCbsPlanQueryItem, error) {

	obsProjectStr := slice.Map(obsProjects, func(o enumor.ObsProject) string {
		return string(o)
	})

	// init request parameter.
	queryReq := &cvmapi.CvmCbsPlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanQueryMethod,
		},
		Params: &cvmapi.CvmCbsPlanQueryParam{
			Page: &cvmapi.Page{
				Start: 0,
				Size:  int(core.DefaultMaxPageLimit),
			},
			BgName:         []string{cvmapi.CvmCbsPlanQueryBgName},
			ProductName:    []string{cvmapi.TransferOpProductName},
			ProjectName:    obsProjectStr,
			TechnicalClass: technicalClasses,
		},
	}

	// query all demands.
	result := make([]*cvmapi.CvmCbsPlanQueryItem, 0)
	for start := 0; ; start += int(core.DefaultMaxPageLimit) {
		queryReq.Params.Page.Start = start
		rst, err := f.crpCli.QueryCvmCbsPlans(kt.Ctx, kt.Header(), queryReq)
		if err != nil {
			logs.Errorf("query crp demands failed, err: %v, params: %+v, rid: %s", err, queryReq.Params, kt.Rid)
			return nil, err
		}

		if rst.Error.Code != 0 {
			logs.Errorf("failed to query cvm cbs plans, err: %s, crp_trace: %s, rid: %s", rst.Error.Message,
				rst.TraceId, kt.Rid)
			return nil, errors.New(rst.Error.Message)
		}

		result = append(result, rst.Result.Data...)

		if len(rst.Result.Data) < int(core.DefaultMaxPageLimit) {
			break
		}
	}

	return result, nil
}
