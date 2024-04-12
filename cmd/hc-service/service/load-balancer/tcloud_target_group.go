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

	req := new(protolb.TCloudBatchOperateTargetReq)
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
	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
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
	return svc.batchAddTargetsToGroup(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchAddTargetsToGroup(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) (
	*protolb.BatchCreateResult, error) {

	tcloudAdpt, err := svc.ad.TCloud(kt, tgInfo.AccountID)
	if err != nil {
		return nil, err
	}

	cloudLBExists := make(map[string]struct{}, 0)
	rsOpt := &typelb.TCloudRegisterTargetsOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := cloudLBExists[ruleItem.CloudLbID]; !ok {
			rsOpt.LoadBalancerId = ruleItem.CloudLbID
			cloudLBExists[ruleItem.CloudLbID] = struct{}{}
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

	rsIDs, err := svc.batchCreateTargetDb(kt, req, tgInfo.AccountID, tgInfo.ID)
	if err != nil {
		return nil, err
	}
	return &protolb.BatchCreateResult{SuccessCloudIDs: rsIDs.IDs}, nil
}

func (svc *clbSvc) batchCreateTargetDb(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
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

// BatchRemoveTCloudTargets 批量移除RS
func (svc *clbSvc) BatchRemoveTCloudTargets(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
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

	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
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

	// 调用云端批量解绑四七层后端服务接口
	return nil, svc.batchUnRegisterTargetCloud(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchUnRegisterTargetCloud(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) error {

	tcloudAdpt, err := svc.ad.TCloud(kt, tgInfo.AccountID)
	if err != nil {
		return err
	}

	cloudLBExists := make(map[string]struct{}, 0)
	rsOpt := &typelb.TCloudRegisterTargetsOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := cloudLBExists[ruleItem.CloudLbID]; !ok {
			rsOpt.LoadBalancerId = ruleItem.CloudLbID
			cloudLBExists[ruleItem.CloudLbID] = struct{}{}
		}
		for _, rsItem := range req.RsList {
			tmpRs := &typelb.BatchTarget{
				ListenerId: cvt.ValToPtr(ruleItem.CloudLBLID),
				InstanceId: cvt.ValToPtr(rsItem.CloudInstID),
				Port:       cvt.ValToPtr(rsItem.Port),
			}
			if ruleItem.RuleType == enumor.Layer7RuleType {
				tmpRs.LocationId = cvt.ValToPtr(ruleItem.CloudID)
			}
			rsOpt.Targets = append(rsOpt.Targets, tmpRs)
		}
		failIDs, err := tcloudAdpt.DeRegisterTargets(kt, rsOpt)
		if err != nil {
			logs.Errorf("unregister tcloud target api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt, kt.Rid)
			return err
		}
		if len(failIDs) > 0 {
			logs.Errorf("unregister tcloud target api partially failed, failLblIDs: %v, req: %+v, rsOpt: %+v, rid: %s",
				failIDs, req, rsOpt, kt.Rid)
			return errf.Newf(errf.PartialFailed, "unregister tcloud target failed, failListenerIDs: %v", failIDs)
		}
	}

	err = svc.batchDeleteTargetDb(kt, req, tgInfo.AccountID, tgInfo.ID)
	if err != nil {
		return err
	}
	return nil
}

func (svc *clbSvc) batchDeleteTargetDb(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	accountID, tgID string) error {

	// 检查RS是否已绑定该目标组
	rsID := make([]string, 0)
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
			return err
		}
		if len(tmpRsList.Details) > 0 {
			rsID = append(rsID, tmpRsList.Details[0].ID)
		}
	}
	if len(rsID) == 0 {
		return nil
	}

	delReq := &dataproto.LoadBalancerBatchDeleteReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", rsID),
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("target_group_id", tgID),
		),
	}
	return svc.dataCli.Global.LoadBalancer.BatchDeleteTarget(kt, delReq)
}

// BatchModifyTCloudTargetsPort 批量修改RS端口
func (svc *clbSvc) BatchModifyTCloudTargetsPort(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
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

	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
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

	// 调用云端批量解绑四七层后端服务接口
	return nil, svc.batchModifyTargetPortCloud(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchModifyTargetPortCloud(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) error {

	tcloudAdpt, err := svc.ad.TCloud(kt, tgInfo.AccountID)
	if err != nil {
		return err
	}

	rsOpt := &typelb.TCloudTargetPortUpdateOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		rsOpt.LoadBalancerId = ruleItem.CloudLbID
		rsOpt.ListenerId = ruleItem.CloudLBLID
		if ruleItem.RuleType == enumor.Layer7RuleType {
			rsOpt.LocationId = cvt.ValToPtr(ruleItem.CloudID)
		}
		for _, rsItem := range req.RsList {
			rsOpt.Targets = append(rsOpt.Targets, &typelb.BatchTarget{
				Type:       cvt.ValToPtr(string(rsItem.InstType)),
				InstanceId: cvt.ValToPtr(rsItem.CloudInstID),
				Port:       cvt.ValToPtr(rsItem.Port),
			})
		}
		rsOpt.NewPort = cvt.PtrToVal(req.RsList[0].NewPort)
		err = tcloudAdpt.ModifyTargetPort(kt, rsOpt)
		if err != nil {
			logs.Errorf("batch modify tcloud target port api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt, kt.Rid)
			return errf.Newf(errf.PartialFailed, "batch modify tcloud target port api failed, err: %v", err)
		}
	}

	err = svc.batchUpdateTargetPortWeightDb(kt, req)
	if err != nil {
		return err
	}
	return nil
}

func (svc *clbSvc) batchUpdateTargetPortWeightDb(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq) error {
	// 检查RS是否已绑定该目标组
	updateReq := &dataproto.TargetBatchUpdateReq{}
	for _, item := range req.RsList {
		tgReq := &core.ListReq{
			Filter: tools.EqualExpression("id", item.ID),
			Page:   core.NewDefaultBasePage(),
		}
		tmpRsList, err := svc.dataCli.Global.LoadBalancer.ListTarget(kt, tgReq)
		if err != nil {
			return err
		}
		if len(tmpRsList.Details) > 0 {
			updateReq.Targets = append(updateReq.Targets, &dataproto.TargetUpdate{
				ID:     item.ID,
				Port:   cvt.PtrToVal(item.NewPort),
				Weight: item.NewWeight,
			})
		}
	}
	if len(updateReq.Targets) == 0 {
		return nil
	}

	return svc.dataCli.Global.LoadBalancer.BatchUpdateTarget(kt, updateReq)
}

// BatchModifyTCloudTargetsWeight 批量修改RS权重
func (svc *clbSvc) BatchModifyTCloudTargetsWeight(cts *rest.Contexts) (any, error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(protolb.TCloudBatchOperateTargetReq)
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

	// 该目标组尚未绑定监听器及规则，不需要云端操作
	if len(ruleRelList.Details) == 0 {
		return &protolb.BatchCreateResult{}, nil
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

	// 批量修改监听器绑定的后端机器的转发权重
	return nil, svc.batchModifyTargetWeightCloud(cts.Kit, req, tgList[0], urlRuleList)
}

func (svc *clbSvc) batchModifyTargetWeightCloud(kt *kit.Kit, req *protolb.TCloudBatchOperateTargetReq,
	tgInfo corelb.BaseTargetGroup, urlRuleList *dataproto.TCloudURLRuleListResult) error {

	tcloudAdpt, err := svc.ad.TCloud(kt, tgInfo.AccountID)
	if err != nil {
		return err
	}

	rsOpt := &typelb.TCloudTargetWeightUpdateOption{
		Region: tgInfo.Region,
	}
	for _, ruleItem := range urlRuleList.Details {
		rsOpt.LoadBalancerId = ruleItem.CloudLbID
		tmpRs := &typelb.TargetWeightRule{
			ListenerId: cvt.ValToPtr(ruleItem.CloudLBLID),
		}
		if ruleItem.RuleType == enumor.Layer7RuleType {
			tmpRs.LocationId = cvt.ValToPtr(ruleItem.CloudID)
		}
		for _, rsItem := range req.RsList {
			tmpRs.Targets = append(tmpRs.Targets, &typelb.BatchTarget{
				Type:       cvt.ValToPtr(string(rsItem.InstType)),
				InstanceId: cvt.ValToPtr(rsItem.CloudInstID),
				Port:       cvt.ValToPtr(rsItem.Port),
				Weight:     rsItem.NewWeight,
			})
			rsOpt.ModifyList = append(rsOpt.ModifyList, tmpRs)
		}
		err = tcloudAdpt.ModifyTargetWeight(kt, rsOpt)
		if err != nil {
			logs.Errorf("batch modify tcloud target port api failed, err: %v, rsOpt: %+v, rid: %s", err, rsOpt, kt.Rid)
			return errf.Newf(errf.PartialFailed, "batch modify tcloud target port api failed, err: %v", err)
		}
	}

	err = svc.batchUpdateTargetPortWeightDb(kt, req)
	if err != nil {
		return err
	}
	return nil
}

// ListTCloudTargetsHealth 查询目标组所在负载均衡的端口健康数据
func (svc *clbSvc) ListTCloudTargetsHealth(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudTargetHealthReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}
	if len(req.Region) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "region is required")
	}

	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudListTargetHealthOption{
		Region:          req.Region,
		LoadBalancerIDs: req.CloudLbIDs,
	}
	healthList, err := tcloudAdpt.ListTargetHealth(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud target health api failed, err: %v, cloudLbIDs: %v, rid: %s",
			err, req.CloudLbIDs, cts.Kit.Rid)
		return nil, err
	}

	healths := &protolb.TCloudTargetHealthResp{}
	for _, item := range healthList {
		tmpHealthInfo := protolb.TCloudTargetHealthResult{CloudLbID: cvt.PtrToVal(item.LoadBalancerId)}
		for _, lblItem := range item.Listeners {
			tmpListener := &protolb.TCloudTargetHealthLblResult{
				CloudLblID:   cvt.PtrToVal(lblItem.ListenerId),
				Protocol:     enumor.ProtocolType(cvt.PtrToVal(lblItem.Protocol)),
				ListenerName: cvt.PtrToVal(lblItem.ListenerName),
			}
			for _, ruleItem := range lblItem.Rules {
				var healthNum, unHealthNum int64
				for _, targetItem := range ruleItem.Targets {
					// 当前健康状态，true：健康，false：不健康（包括尚未开始探测、探测中、状态异常等几种状态）。
					if cvt.PtrToVal(targetItem.HealthStatus) {
						healthNum++
					} else {
						unHealthNum++
					}
				}

				if !tmpListener.Protocol.IsLayer7Protocol() {
					tmpListener.HealthCheck = &corelb.TCloudHealthCheckInfo{
						HealthNum:   cvt.ValToPtr(healthNum),
						UnHealthNum: cvt.ValToPtr(unHealthNum),
					}
					break
				} else {
					tmpListener.Rules = append(tmpListener.Rules, &protolb.TCloudTargetHealthRuleResult{
						CloudRuleID: cvt.PtrToVal(ruleItem.LocationId),
						HealthCheck: &corelb.TCloudHealthCheckInfo{
							HealthNum:   cvt.ValToPtr(healthNum),
							UnHealthNum: cvt.ValToPtr(unHealthNum),
						},
					})
				}
			}
			tmpHealthInfo.Listeners = append(tmpHealthInfo.Listeners, tmpListener)
		}
		healths.Details = append(healths.Details, tmpHealthInfo)
	}

	return healths, nil
}
