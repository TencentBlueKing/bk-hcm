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
	"encoding/json"
	"fmt"

	actionflow "hcm/cmd/task-server/logics/action/flow"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/producer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchCreateLB 批量创建负载均衡
func (svc *lbSvc) BatchCreateLB(cts *rest.Contexts) (any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create clb request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{
		Type:       meta.LoadBalancer,
		Action:     meta.Create,
		ResourceID: req.AccountID,
	}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create load balancer auth failed, err: %v, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.
		GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudLB(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}

}

func (svc *lbSvc) batchCreateTCloudLB(kt *kit.Kit, rawReq json.RawMessage) (any, error) {
	req := new(hcproto.TCloudBatchCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.client.HCService().TCloud.Clb.BatchCreate(kt, req)
}

// CreateBizTargetGroup create biz target group.
func (svc *lbSvc) CreateBizTargetGroup(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.createBizTargetGroup(cts, handler.BizOperateAuth, bkBizID)
}

func (svc *lbSvc) createBizTargetGroup(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler,
	bkBizID int64) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create target group request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Create, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("create target group auth failed, err: %v, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudTargetGroup(cts.Kit, req.Data, bkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) batchCreateTCloudTargetGroup(kt *kit.Kit, rawReq json.RawMessage, bkBizID int64) (any, error) {
	req := new(dataproto.TargetGroupCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &dataproto.TCloudTargetGroupCreateReq{
		TargetGroups: []dataproto.TargetGroupBatchCreate[corelb.TCloudTargetGroupExtension]{
			{
				Name:            req.Name,
				Vendor:          enumor.TCloud,
				AccountID:       req.AccountID,
				BkBizID:         bkBizID,
				Region:          req.Region,
				Protocol:        req.Protocol,
				Port:            req.Port,
				CloudVpcID:      req.CloudVpcID,
				TargetGroupType: enumor.LocalTargetGroupType,
				RsList:          req.RsList,
			},
		},
	}
	return svc.client.DataService().TCloud.LoadBalancer.BatchCreateTCloudTargetGroup(kt, opt)
}

// CreateBizListener create biz listener.
func (svc *lbSvc) CreateBizListener(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.createListener(cts, handler.BizOperateAuth, bkBizID)
}

func (svc *lbSvc) createListener(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler,
	bkBizID int64) (any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create listener request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if len(req.AccountID) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account_id is required")
	}

	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lb_id is required")
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Listener,
		Action: meta.Create, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("create listener auth failed, err: %v, account id: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudListener(cts.Kit, req.Data, bkBizID, lbID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) batchCreateTCloudListener(kt *kit.Kit, rawReq json.RawMessage, bkBizID int64,
	lbID string) (any, error) {

	req := new(hcproto.ListenerWithRuleCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req.BkBizID = bkBizID
	req.LbID = lbID
	return svc.client.HCService().TCloud.Clb.CreateListener(kt, req)
}

// BatchAddBizTargets create add biz targets.
func (svc *lbSvc) BatchAddBizTargets(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.batchAddBizTarget(cts, handler.BizOperateAuth, bkBizID)
}

func (svc *lbSvc) batchAddBizTarget(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler,
	bkBizID int64) (any, error) {

	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch add rs request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		logs.Errorf("get target group resource info failed, id: %s, err: %s, rid: %s", tgID, err, cts.Kit.Rid)
		return nil, err
	}

	// authorized instances
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: baseInfo})
	if err != nil {
		logs.Errorf("batch add rs auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.buildAddTCloudTargetTasks(cts.Kit, req.Data, tgID, baseInfo.AccountID, bkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
}

func (svc *lbSvc) buildAddTCloudTargetTasks(kt *kit.Kit, body json.RawMessage, tgID, accountID string, bkBizID int64) (
	interface{}, error) {

	req := new(cslb.TCloudTargetBatchCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 预检测
	err := svc.checkResFlowRel(kt, tgID, enumor.TargetGroupCloudResType, bkBizID)
	if err != nil {
		return nil, err
	}

	// 创建Flow跟Task的初始化数据
	tasks := make([]ts.CustomFlowTask, 0)
	elems := slice.Split(req.RsList, constant.BatchAddRSCloudMaxLimit)
	getActionID := counter.NewNumStringCounter(1, 10)
	for _, parts := range elems {
		addRsParams, err := svc.convTCloudAddTargetReq(kt, parts, tgID, accountID)
		if err != nil {
			logs.Errorf("add rs build tcloud request failed, err: %v, tgID: %s, parts: %+v rid: %s",
				err, tgID, parts, kt.Rid)
			return nil, err
		}
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(getActionID()),
			ActionName: enumor.ActionAddRS,
			Params: &actionlb.OperateRsOption{
				Vendor:                      enumor.TCloud,
				TCloudBatchOperateTargetReq: *addRsParams,
			},
			Retry: &tableasync.Retry{
				Enable: true,
				Policy: &tableasync.RetryPolicy{
					Count:        constant.FlowRetryMaxLimit,
					SleepRangeMS: [2]uint{100, 200},
				},
			},
		})
	}
	addReq := &ts.AddCustomFlowReq{
		Name:        enumor.FlowAddRS,
		Tasks:       tasks,
		IsInitState: true,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch add rs custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	flowID := result.ID
	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.FlowWatchOption{
				FlowID:   flowID,
				ResID:    tgID,
				ResType:  enumor.TargetGroupCloudResType,
				TaskType: enumor.AddRSTaskType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return nil, err
	}

	// 锁定资源跟Flow的状态
	err = svc.lockResFlowStatus(kt, tgID, enumor.TargetGroupCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		return nil, err
	}

	return &core.FlowStateResult{FlowID: flowID}, nil
}

// convTCloudAddTargetReq conv tcloud add target req.
func (svc *lbSvc) convTCloudAddTargetReq(kt *kit.Kit, targets []*dataproto.TargetBaseReq, targetGroupID,
	accountID string) (*hcproto.TCloudBatchOperateTargetReq, error) {

	instMap, err := svc.getInstWithTargetMap(kt, targets)
	if err != nil {
		return nil, err
	}

	rsReq := &hcproto.TCloudBatchOperateTargetReq{TargetGroupID: targetGroupID}
	for _, item := range targets {
		item.TargetGroupID = targetGroupID
		item.AccountID = accountID
		item.InstType = item.InstType
		item.InstName = instMap[item.CloudInstID].Name
		item.PrivateIPAddress = instMap[item.CloudInstID].PrivateIPv4Addresses
		item.PublicIPAddress = instMap[item.CloudInstID].PublicIPv4Addresses
		item.CloudVpcIDs = instMap[item.CloudInstID].CloudVpcIDs
		item.Zone = instMap[item.CloudInstID].Zone
		rsReq.RsList = append(rsReq.RsList, item)
	}
	return rsReq, nil
}

func (svc *lbSvc) getInstWithTargetMap(kt *kit.Kit, targets []*dataproto.TargetBaseReq) (
	map[string]corecvm.BaseCvm, error) {

	cloudCvmIDs := make([]string, 0)
	for _, item := range targets {
		if item.InstType == enumor.CvmInstType {
			cloudCvmIDs = append(cloudCvmIDs, item.CloudInstID)
		}
	}

	// 查询Cvm信息
	cvmMap := make(map[string]corecvm.BaseCvm)
	if len(cloudCvmIDs) > 0 {
		cvmReq := &core.ListReq{
			Filter: tools.ContainersExpression("cloud_id", cloudCvmIDs),
			Page:   core.NewDefaultBasePage(),
		}
		cvmList, err := svc.client.DataService().Global.Cvm.ListCvm(kt, cvmReq)
		if err != nil {
			logs.Errorf("failed to list cvm by cloudIDs, cloudIDs: %v, err: %v, rid: %s", cloudCvmIDs, err, kt.Rid)
			return nil, err
		}

		for _, item := range cvmList.Details {
			cvmMap[item.CloudID] = item
		}
	}

	return cvmMap, nil
}

func (svc *lbSvc) lockResFlowStatus(kt *kit.Kit, resID string, resType enumor.CloudResourceType, flowID string,
	taskType enumor.TaskType) error {

	// 锁定资源跟Flow的状态
	opt := &dataproto.ResFlowLockReq{
		ResID:    resID,
		ResType:  resType,
		FlowID:   flowID,
		Status:   enumor.ExecutingResFlowStatus,
		TaskType: taskType,
	}
	err := svc.client.DataService().Global.LoadBalancer.ResFlowLock(kt, opt)
	if err != nil {
		logs.Errorf("call dataservice to lock res and flow failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return err
	}

	// 更新Flow状态为pending
	flowStateReq := &producer.UpdateCustomFlowStateOption{
		FlowInfos: []backend.UpdateFlowInfo{{
			ID:     flowID,
			Source: enumor.FlowInit,
			Target: enumor.FlowPending,
		}},
	}
	err = svc.client.TaskServer().UpdateCustomFlowState(kt, flowStateReq)
	if err != nil {
		logs.Errorf("call taskserver to update flow state failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return err
	}

	return nil
}

func (svc *lbSvc) checkResFlowRel(kt *kit.Kit, resID string, resType enumor.CloudResourceType, bkBizID int64) error {
	// 检查目标组是否存在
	if resType == enumor.TargetGroupCloudResType {
		targetGroupList, err := svc.getTargetGroupByID(kt, resID, bkBizID)
		if err != nil {
			logs.Errorf("list target group by id failed, tgID: %s, err: %v, rid: %s", resID, err, kt.Rid)
			return err
		}
		if len(targetGroupList) == 0 {
			return errf.Newf(errf.RecordNotFound, "target group: %s not found", resID)
		}
	}

	// 预检测-当前资源是否有锁定中的数据
	lockReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("res_type", resType),
		),
		Page: core.NewDefaultBasePage(),
	}
	lockRet, err := svc.client.DataService().Global.LoadBalancer.ListResFlowLock(kt, lockReq)
	if err != nil {
		logs.Errorf("list res flow lock failed, err: %v, resID: %s, resType: %s, rid: %s", err, resID, resType, kt.Rid)
		return err
	}
	if len(lockRet.Details) > 0 {
		return errf.Newf(errf.TooManyRequest, "resID: %s is processing", resID)
	}

	// 预检测-当前资源是否有未终态的状态
	flowRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("status", enumor.ExecutingResFlowStatus),
		),
		Page: core.NewDefaultBasePage(),
	}
	flowRelRet, err := svc.client.DataService().Global.LoadBalancer.ListResFlowRel(kt, flowRelReq)
	if err != nil {
		logs.Errorf("list res flow rel failed, err: %v, resID: %s, resType: %s, rid: %s", err, resID, resType, kt.Rid)
		return err
	}
	if len(flowRelRet.Details) > 0 {
		return errf.Newf(errf.TooManyRequest, "%s of resID: %s is processing", resType, resID)
	}

	return nil
}
