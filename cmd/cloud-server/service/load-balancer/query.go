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
	proto "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
)

// ListLoadBalancer list load balancer.
func (svc *lbSvc) ListLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.listLoadBalancer(cts, handler.ListResourceAuthRes)
}

// ListBizLoadBalancer list biz load balancer.
func (svc *lbSvc) ListBizLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.listLoadBalancer(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listLoadBalancer(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Find,
		Filter:     req.Filter,
	})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(cts.Kit, listReq)
}

// ListLoadBalancerWithDeleteProtect list load balancer with delete protect
func (svc *lbSvc) ListLoadBalancerWithDeleteProtect(cts *rest.Contexts) (any, error) {
	return svc.listLoadBalancerWithDeleteProtect(cts, handler.ListResourceAuthRes)
}

// ListBizLoadBalancerWithDeleteProtect list biz load balancer with delete protect
func (svc *lbSvc) ListBizLoadBalancerWithDeleteProtect(cts *rest.Contexts) (any, error) {
	return svc.listLoadBalancerWithDeleteProtect(cts, handler.ListBizAuthRes)
}

// list load balancer with delete protect
func (svc *lbSvc) listLoadBalancerWithDeleteProtect(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	any, error) {

	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Find,
		Filter:     req.Filter,
	})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]any, 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	dataResp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancerRaw(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to list load balancer with extension for delete protection, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}
	lbResult := core.ListResultT[*corelb.LoadBalancerWithDeleteProtect]{
		Count: dataResp.Count,
	}
	for _, detail := range dataResp.Details {
		lb := &corelb.LoadBalancerWithDeleteProtect{BaseLoadBalancer: detail.BaseLoadBalancer}

		// 目前仅支持tcloud 的删除保护
		if detail.Vendor == enumor.TCloud {
			extension := corelb.TCloudClbExtension{}
			err := json.Unmarshal(detail.Extension, &extension)
			if err != nil {
				logs.Errorf("fail parse lb extension for delete protection, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
			lb.DeleteProtect = cvt.PtrToVal(extension.DeleteProtect)
		}
		lbResult.Details = append(lbResult.Details, lb)

	}
	return lbResult, nil
}

// GetLoadBalancer getLoadBalancer clb.
func (svc *lbSvc) GetLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.getLoadBalancer(cts, handler.ListResourceAuthRes)
}

// GetBizLoadBalancer getLoadBalancer biz clb.
func (svc *lbSvc) GetBizLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return svc.getLoadBalancer(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) getLoadBalancer(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.LoadBalancerCloudResType,
		id)
	if err != nil {
		logs.Errorf("fail to get clb basic info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get clb")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.LoadBalancer.Get(cts.Kit, id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

// ListTargetsByTGID ...
func (svc *lbSvc) ListTargetsByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsByTGID(cts, handler.ResOperateAuth)
}

// ListBizTargetsByTGID ...
func (svc *lbSvc) ListBizTargetsByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsByTGID(cts, handler.BizOperateAuth)
}

// listTargetsByTGID 目标组下RS列表
func (svc *lbSvc) listTargetsByTGID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {
	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TargetGroup,
		Action:     meta.Find,
		BasicInfo:  basicInfo,
	})
	if err != nil {
		return nil, err
	}
	filter, err := tools.And(req.Filter, tools.RuleEqual("target_group_id", tgID))
	if err != nil {
		logs.Errorf("merge filter failed, err: %v, target_group_id: %s, rid: %s", err, tgID, cts.Kit.Rid)
		return nil, err
	}
	listReq := &core.ListReq{
		Filter: filter,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.LoadBalancer.ListTarget(cts.Kit, listReq)
}

// ListTargetsHealthByTGID 查询业务下指定目标组绑定的负载均衡下的RS端口健康信息
func (svc *lbSvc) ListTargetsHealthByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsHealthByTGID(cts, handler.BizOperateAuth)
}

// ListBizTargetsHealthByTGID 查询资源下指定目标组负载均衡下的RS端口健康信息
func (svc *lbSvc) ListBizTargetsHealthByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsHealthByTGID(cts, handler.ResOperateAuth)
}

// listTargetsHealthByTGID 目标组绑定的负载均衡下的RS端口健康信息
func (svc *lbSvc) listTargetsHealthByTGID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(hcproto.TCloudTargetHealthReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TargetGroup,
		Action:     meta.Find,
		BasicInfo:  basicInfo,
	})
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		tgInfo, newCloudLbIDs, err := svc.checkBindGetTargetGroupInfo(cts.Kit, tgID, req.CloudLbIDs)
		if err != nil {
			return nil, err
		}

		req.AccountID = tgInfo.AccountID
		req.Region = tgInfo.Region
		req.CloudLbIDs = newCloudLbIDs
		return svc.client.HCService().TCloud.Clb.ListTargetHealth(cts.Kit, req)
	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", tgID, basicInfo.Vendor)
	}
}

// checkBindGetTargetGroupInfo 检查目标组是否存在、是否已绑定其他监听器
func (svc *lbSvc) checkBindGetTargetGroupInfo(kt *kit.Kit, tgID string, cloudLbIDs []string) (
	*corelb.BaseTargetGroup, []string, error) {

	// 查询目标组的基本信息
	tgInfo, err := svc.getTargetGroupByID(kt, tgID)
	if err != nil {
		return nil, nil, err
	}

	if tgInfo == nil {
		return nil, nil, errf.Newf(errf.RecordNotFound, "target group: %s is not found", tgID)
	}

	// 查询该目标组绑定的负载均衡、监听器数据
	ruleRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("target_group_id", tgID),
			tools.RuleIn("cloud_lb_id", cloudLbIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
	if err != nil {
		logs.Errorf("list tcloud listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
		return nil, nil, err
	}

	if len(ruleRelList.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotUpdate, "target group: %s has not bound listener", tgID)
	}

	// 以当前目标组绑定的负载均衡ID为准
	newCloudLbIDs := slice.Map(ruleRelList.Details, func(one corelb.BaseTargetListenerRuleRel) string {
		return one.CloudLbID
	})
	return tgInfo, newCloudLbIDs, nil
}

// GetLoadBalancerLockStatus get load balancer status.
func (svc *lbSvc) GetLoadBalancerLockStatus(cts *rest.Contexts) (interface{}, error) {
	return svc.getLoadBalancerLockStatus(cts, handler.ListResourceAuthRes)
}

// GetBizLoadBalancerLockStatus get biz load balancer status.
func (svc *lbSvc) GetBizLoadBalancerLockStatus(cts *rest.Contexts) (interface{}, error) {
	return svc.getLoadBalancerLockStatus(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) getLoadBalancerLockStatus(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.LoadBalancerCloudResType, id)
	if err != nil {
		logs.Errorf("fail to get load balancer basic info, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get load balancer")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		// 预检测-是否有执行中的负载均衡
		flowRelResp, err := svc.checkResFlowRel(cts.Kit, id, enumor.LoadBalancerCloudResType)
		if err != nil {
			logs.Errorf("load balancer %s is executing flow, err: %v, rid: %s", id, err, cts.Kit.Rid)
			flowStatus := &cslb.ResourceFlowStatusResp{Status: enumor.ExecutingResFlowStatus}
			if flowRelResp != nil {
				flowStatus.ResID = flowRelResp.ResID
				flowStatus.ResType = flowRelResp.ResType
				flowStatus.FlowID = flowRelResp.Owner
			}
			return flowStatus, nil
		}

		return &cslb.ResourceFlowStatusResp{Status: enumor.SuccessResFlowStatus}, nil
	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}
