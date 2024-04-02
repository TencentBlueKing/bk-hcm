/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package loadbalancer

import (
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// BatchCreateTCloudTargets 批量添加RS
func (svc *clbSvc) BatchCreateTCloudTargets(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchCreateTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tgList, err := svc.getTargetGroupByID(cts.Kit, tgID)
	if err != nil {
		return nil, err
	}

	if len(tgList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", tgID)
	}

	// 根据目标组ID，获取目标组绑定的监听器、规则列表
	ruleRelReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.dataCli.Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}
	accountID := tgList[0].AccountID
	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		rsIDs, err := svc.batchCreateTargetDb(cts.Kit, req, accountID, tgID)
		if err != nil {
			return nil, err
		}
		return &protolb.BatchCreateResult{SuccessCloudIDs: rsIDs.IDs}, nil
	}

	// 查询Url规则列表
	ruleIDs := slice.Map(ruleRelList.Details, func(one corelb.BaseTargetListenerRuleRel) string {
		return one.ListenerRuleID
	})
	urlRuleReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ruleIDs),
		Page:   core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.dataCli.TCloud.LoadBalancer.ListUrlRule(cts.Kit, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用云端批量绑定虚拟主机接口
	return svc.batchRegisterTargetCloud(cts.Kit, req, accountID, tgID, tgList[0].Region, urlRuleList)
}

func (svc *clbSvc) batchRegisterTargetCloud(kt *kit.Kit, req *protolb.TCloudBatchCreateTargetReq,
	accountID, tgID, region string, urlRuleList *dataproto.TCloudURLRuleListResult) (
	*protolb.BatchCreateResult, error) {

	tcloudAdpt, err := svc.ad.TCloud(kt, accountID)
	if err != nil {
		return nil, err
	}

	cloudLBExists := make(map[string]struct{}, 0)
	rsOpt := &typelb.TCloudRegisterTargetsOption{
		Region: region,
	}
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := cloudLBExists[ruleItem.CloudLbID]; !ok {
			rsOpt.LoadBalancerId = ruleItem.CloudLbID
		}
		for _, rsItem := range req.RsList {
			tmpRs := &typelb.BatchTarget{
				ListenerId: cvt.ValToPtr(ruleItem.CloudLBLID),
				InstanceId: cvt.ValToPtr(rsItem.CloudInstID),
				Port:       cvt.ValToPtr(rsItem.Port),
				Weight:     rsItem.Weight,
			}
			if ruleItem.RuleType == enumor.Layer7RuleType {
				tmpRs.LocationId = cvt.ValToPtr(ruleItem.CloudID)
			}
			rsOpt.Targets = append(rsOpt.Targets, tmpRs)
		}
		failIDs, err := tcloudAdpt.RegisterTargets(kt, rsOpt)
		if err != nil {
			logs.Errorf("register tcloud target api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt, kt.Rid)
			return nil, err
		}
		if len(failIDs) > 0 {
			logs.Errorf("register tcloud target api partially failed, failLblIDs: %v, req: %+v, rsOpt: %+v, rid: %s",
				failIDs, req, rsOpt, kt.Rid)
			return nil, errf.Newf(errf.PartialFailed, "register tcloud target failed, failListenerIDs: %v", failIDs)
		}
	}

	rsIDs, err := svc.batchCreateTargetDb(kt, req, accountID, tgID)
	if err != nil {
		return nil, err
	}
	return &protolb.BatchCreateResult{SuccessCloudIDs: rsIDs.IDs}, nil
}

func (svc *clbSvc) batchCreateTargetDb(kt *kit.Kit, req *protolb.TCloudBatchCreateTargetReq,
	accountID, tgID string) (*core.BatchCreateResult, error) {

	// 检查RS是否已绑定该目标组
	rsList := make([]*dataproto.TargetBaseReq, 0)
	for _, item := range req.RsList {
		tgReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("account_id", accountID),
				tools.RuleEqual("target_group_id", tgID),
				tools.RuleEqual("cloud_inst_id", item.CloudInstID),
				tools.RuleEqual("port", item.Port),
			),
			Page: core.NewDefaultBasePage(),
		}
		tmpRsList, err := svc.dataCli.Global.LoadBalancer.ListTarget(kt, tgReq)
		if err != nil {
			return nil, err
		}
		if len(tmpRsList.Details) == 0 {
			rsList = append(rsList, item)
		}
	}
	if len(rsList) == 0 {
		return &core.BatchCreateResult{}, nil
	}

	rsReq := &dataproto.TargetBatchCreateReq{}
	for _, item := range rsList {
		rsReq.Targets = append(rsReq.Targets, &dataproto.TargetBaseReq{
			AccountID:     accountID,
			TargetGroupID: tgID,
			InstType:      item.InstType,
			CloudInstID:   item.CloudInstID,
			Port:          item.Port,
			Weight:        item.Weight,
		})
	}
	return svc.dataCli.Global.LoadBalancer.BatchCreateTCloudTarget(kt, rsReq)
}
