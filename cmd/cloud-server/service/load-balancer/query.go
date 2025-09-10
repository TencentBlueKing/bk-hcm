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
	"errors"
	"fmt"
	"strings"

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	proto "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
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

// ListBizLoadBalancerWithDelProtect list biz load balancer with delete protect
func (svc *lbSvc) ListBizLoadBalancerWithDelProtect(cts *rest.Contexts) (any, error) {
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
	return svc.listTargetsHealthByTGID(cts, handler.ResOperateAuth)
}

// ListBizTargetsHealthByTGID 查询资源下指定目标组负载均衡下的RS端口健康信息
func (svc *lbSvc) ListBizTargetsHealthByTGID(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetsHealthByTGID(cts, handler.BizOperateAuth)
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
		// 查询对应负载均衡信息
		lbReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", newCloudLbIDs),
				tools.RuleEqual("vendor", tgInfo.Vendor),
				tools.RuleEqual("account_id", tgInfo.AccountID),
			),
			Page: core.NewDefaultBasePage(),
		}
		lbResp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(cts.Kit, lbReq)
		if err != nil {
			logs.Errorf("fail to find load balancer(%v) for target group health, err: %v, rid: %s",
				newCloudLbIDs, err, cts.Kit.Rid)
			return nil, err
		}
		if len(lbResp.Details) != len(newCloudLbIDs) {
			return nil, errors.New("some of given load balancer can not be found")
		}
		req.Region = ""
		req.AccountID = tgInfo.AccountID
		req.CloudLbIDs = newCloudLbIDs
		for _, detail := range lbResp.Details {
			if req.Region == "" {
				req.Region = detail.Region
				continue
			}
			if req.Region != detail.Region {
				return nil, fmt.Errorf("load balancers have different regions: %s,%s", req.Region, detail.Region)
			}
		}
		return svc.client.HCService().TCloud.Clb.ListTargetHealth(cts.Kit, req)
	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", tgID, basicInfo.Vendor)
	}
}

// checkBindGetTargetGroupInfo 检查目标组是否存在、是否已绑定其他监听器，给定云id可能重复，
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
	newCloudLbIDs = slice.Unique(newCloudLbIDs) //去重，避免重复ID
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

// getListenerByIDAndBiz get listener by id and bizID.
func (svc *lbSvc) getListenerByIDAndBiz(kt *kit.Kit, vendor enumor.Vendor, bizID int64, lblID string) (
	*corelb.BaseListener, *types.CloudResourceBasicInfo, error) {

	lblResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt,
		&core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("id", lblID),
				tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("bk_biz_id", bizID)),
			Page: core.NewDefaultBasePage(),
		})
	if err != nil {
		logs.Errorf("fail to list listener(%s), err: %v, rid: %s", lblID, err, kt.Rid)
		return nil, nil, err
	}
	if len(lblResp.Details) == 0 {
		return nil, nil, errf.New(errf.RecordNotFound, "listener not found, id: "+lblID)
	}
	lblInfo := &lblResp.Details[0]
	basicInfo := &types.CloudResourceBasicInfo{
		ResType:   enumor.ListenerCloudResType,
		ID:        lblID,
		Vendor:    vendor,
		AccountID: lblInfo.AccountID,
		BkBizID:   lblInfo.BkBizID,
	}

	return lblInfo, basicInfo, nil
}

// getListenerByID get listener by id.
func (svc *lbSvc) getListenerByID(kt *kit.Kit, lblID string) (*corelb.BaseListener, error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", lblID),
		),
		Page: core.NewDefaultBasePage(),
	}
	lblResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, req)
	if err != nil {
		logs.Errorf("fail to list listener(%s), err: %v, rid: %s", lblID, err, kt.Rid)
		return nil, err
	}
	if len(lblResp.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "listener not found, id: "+lblID)
	}
	lblInfo := &lblResp.Details[0]

	return lblInfo, nil
}

// listVpcMap 根据vpcIDs查询vpc信息
func (svc *lbSvc) listVpcMap(kt *kit.Kit, vpcIDs []string) (map[string]cloud.BaseVpc, error) {
	if len(vpcIDs) == 0 {
		return nil, nil
	}

	vpcMap := make(map[string]cloud.BaseVpc, len(vpcIDs))
	for _, parts := range slice.Split(vpcIDs, int(core.DefaultMaxPageLimit)) {
		vpcReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", parts),
			Page:   core.NewDefaultBasePage(),
		}
		list, err := svc.client.DataService().Global.Vpc.List(kt.Ctx, kt.Header(), vpcReq)
		if err != nil {
			logs.Errorf("[clb] list vpc failed, vpcIDs: %v, err: %v, rid: %s", vpcIDs, err, kt.Rid)
			return nil, err
		}
		for _, item := range list.Details {
			vpcMap[item.ID] = item
		}
	}
	return vpcMap, nil
}
func (svc *lbSvc) getLoadBalancerByID(kt *kit.Kit, lbID string) (*corelb.BaseLoadBalancer, error) {
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", lbID),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, req)
	if err != nil {
		logs.Errorf("list load balancer failed, req: %v, error: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	if len(resp.Details) == 0 {
		err = fmt.Errorf("load balancer not found, id: %s", lbID)
		logs.Errorf("load balancer not found, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &resp.Details[0], nil
}

func (svc *lbSvc) listTargetsByIDs(kt *kit.Kit, targetIDs []string) ([]corelb.BaseTarget, error) {
	if len(targetIDs) == 0 {
		return nil, nil
	}

	result := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(targetIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		rsList, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, req)
		if err != nil {
			logs.Errorf("list target failed, targetIDs: %v, err: %v, rid: %s", targetIDs, err, kt.Rid)
			return nil, err
		}
		result = append(result, rsList.Details...)
	}

	return result, nil
}

func (svc *lbSvc) listTGListenerRuleRelMapByTGIDs(kt *kit.Kit, tgIDs []string) (
	map[string]corelb.BaseTargetListenerRuleRel, error) {

	if len(tgIDs) == 0 {
		return nil, nil
	}

	result := make(map[string]corelb.BaseTargetListenerRuleRel, 0)
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("target_group_id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		list, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, req)
		if err != nil {
			logs.Errorf("list target group listener rel failed, tgIDs: %v, err: %v, rid: %s", tgIDs, err, kt.Rid)
			return nil, err
		}
		for _, detail := range list.Details {
			result[detail.TargetGroupID] = detail
		}
	}
	return result, nil
}

func (svc *lbSvc) listListenersByIDs(kt *kit.Kit, lblIDs []string) ([]corelb.BaseListener, error) {
	if len(lblIDs) == 0 {
		return nil, nil
	}

	result := make([]corelb.BaseListener, 0, len(lblIDs))

	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		lblReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		listLblResult, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
		if err != nil {
			logs.Errorf("[clb] list clb listener failed, lblIDs: %v, err: %v, rid: %s", lblIDs, err, kt.Rid)
			return nil, err
		}
		result = append(result, listLblResult.Details...)
	}

	return result, nil
}

func (svc *lbSvc) listLoadBalancerMapByIDs(kt *kit.Kit, lbIDs []string) (map[string]corelb.BaseLoadBalancer, error) {
	result := make(map[string]corelb.BaseLoadBalancer, len(lbIDs))
	for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, req)
		if err != nil {
			logs.Errorf("list load balancer failed, req: %v, error: %v, rid: %s", req, err, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			result[detail.ID] = detail
		}
	}
	return result, nil
}

type urlRuleInfo struct {
	domain     string
	url        string
	lblID      string
	cloudLblID string
	cloudLBID  string
}

// listUrlRuleMapByIDs 根据url rule id获取url rule信息
func (svc *lbSvc) listUrlRuleMapByIDs(kt *kit.Kit, vendor enumor.Vendor, ids []string) (
	map[string]urlRuleInfo, error) {

	switch vendor {
	case enumor.TCloud:
		return svc.listUrlRuleMapByIDsForTCloud(kt, ids)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s for listUrlRuleMapByIDs", vendor)
	}
}

func (svc *lbSvc) listUrlRuleMapByIDsForTCloud(kt *kit.Kit, ids []string) (map[string]urlRuleInfo, error) {
	result := make(map[string]urlRuleInfo, 0)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, listReq)
		if err != nil {
			return nil, err
		}
		for _, detail := range resp.Details {
			result[detail.ID] = urlRuleInfo{
				domain:     detail.Domain,
				url:        detail.URL,
				lblID:      detail.LblID,
				cloudLblID: detail.CloudLBLID,
				cloudLBID:  detail.CloudLbID,
			}
		}
	}
	return result, nil
}

// TGRelatedInfo tg关联信息，包括lb, listener, url rule
type TGRelatedInfo struct {
	CloudLBID    string `json:"cloud_lb_id"`
	ClbVipDomain string `json:"clb_vip_domain"`

	Protocol enumor.ProtocolType `json:"protocol"`
	Port     int64               `json:"listener_port"`

	Domain string `json:"domain"`
	URL    string `json:"url"`
}

// listTGRelatedInfoByRels 根据tg rel获取tg关联信息, 返回值 map[TGID]TGRelatedInfo
func (svc *lbSvc) listTGRelatedInfoByRels(kt *kit.Kit, vendor enumor.Vendor, rels []corelb.BaseTargetListenerRuleRel) (map[string]TGRelatedInfo, error) {

	lbMap, err := svc.listLoadBalancerMapByIDs(kt, slice.Map(rels, corelb.BaseTargetListenerRuleRel.GetLbID))
	if err != nil {
		return nil, err
	}

	lbls, err := svc.listListenersByIDs(kt, slice.Map(rels, corelb.BaseTargetListenerRuleRel.GetLblID))
	if err != nil {
		return nil, err
	}
	lblMap := cvt.SliceToMap(lbls, func(item corelb.BaseListener) (string, corelb.BaseListener) {
		return item.ID, item
	})

	ruleMap, err := svc.listUrlRuleMapByIDs(kt, vendor, slice.Map(rels, corelb.BaseTargetListenerRuleRel.GetListenerRuleID))
	if err != nil {
		return nil, err
	}

	result := make(map[string]TGRelatedInfo, len(rels))
	for _, rel := range rels {
		lb, ok := lbMap[rel.LbID]
		if !ok {
			logs.Errorf("lb not found: %s, rel: %+v, rid: %s", rel.LbID, rel, kt.Rid)
			return nil, fmt.Errorf("lb not found: %s", rel.LbID)
		}
		vipDomain, err := getClbVipDomain(lb)
		if err != nil {
			return nil, err
		}

		lbl, ok := lblMap[rel.LblID]
		if !ok {
			logs.Errorf("listener not found: %s, rel: %+v, rid: %s", rel.LblID, rel, kt.Rid)
			return nil, fmt.Errorf("listener not found: %s", rel.LblID)
		}

		rule, ok := ruleMap[rel.ListenerRuleID]
		if !ok {
			logs.Errorf("url rule not found: %s, rel: %+v, rid: %s", rel.ListenerRuleID, rel, kt.Rid)
			return nil, fmt.Errorf("url rule not found: %s", rel.ListenerRuleID)
		}

		item := TGRelatedInfo{
			CloudLBID:    lb.CloudID,
			ClbVipDomain: strings.Join(vipDomain, ","),
			Protocol:     lbl.Protocol,
			Port:         lbl.Port,
			Domain:       rule.domain,
			URL:          rule.url,
		}
		result[rel.TargetGroupID] = item
	}
	return result, nil
}

func getClbVipDomain(lbInfo corelb.BaseLoadBalancer) ([]string, error) {
	vipDomains := make([]string, 0)
	switch lbInfo.LoadBalancerType {
	case string(typeslb.InternalLoadBalancerType):
		if lbInfo.IPVersion == enumor.Ipv4 {
			vipDomains = append(vipDomains, lbInfo.PrivateIPv4Addresses...)
		} else {
			vipDomains = append(vipDomains, lbInfo.PrivateIPv6Addresses...)
		}
	case string(typeslb.OpenLoadBalancerType):
		if lbInfo.IPVersion == enumor.Ipv4 {
			vipDomains = append(vipDomains, lbInfo.PublicIPv4Addresses...)
		} else {
			vipDomains = append(vipDomains, lbInfo.PublicIPv6Addresses...)
		}
	default:
		return nil, fmt.Errorf("unsupported lb_type: %s(%s)", lbInfo.LoadBalancerType, lbInfo.CloudID)
	}

	// 如果IP为空则获取负载均衡域名
	if len(vipDomains) == 0 && len(lbInfo.Domain) > 0 {
		vipDomains = append(vipDomains, lbInfo.Domain)
	}

	return vipDomains, nil
}
