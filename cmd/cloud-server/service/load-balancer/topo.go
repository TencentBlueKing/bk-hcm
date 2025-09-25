/*
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

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	daotypeslb "hcm/pkg/dal/dao/types/load-balancer"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
)

// ListTargetByTopo list target by topo
func (svc *lbSvc) ListTargetByTopo(cts *rest.Contexts) (any, error) {
	req := new(cslb.LbTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	return svc.listTargetByTopo(cts.Kit, bizID, vendor, req)
}

// CountTargetByTopo count target by topo
func (svc *lbSvc) CountTargetByTopo(cts *rest.Contexts) (any, error) {
	req := new(cslb.LbTopoCond)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	info, err := svc.getTargetTopoInfoByReq(cts.Kit, bizID, vendor, req)
	if err != nil {
		logs.Errorf("get clb topo info failed, err: %v, bizID: %d, vendor: %s, req: %+v, rid: %s", err, bizID, vendor,
			req, cts.Kit.Rid)
		return nil, err
	}
	if !info.Match {
		return core.ListResult{Count: 0}, nil
	}

	targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", maps.Keys(info.TgMap))}
	targetCond = append(targetCond, req.GetTargetCond()...)
	targetReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: targetCond},
		Page:   core.NewCountPage(),
	}
	resp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(cts.Kit, &targetReq)
	if err != nil {
		logs.Errorf("count target failed, err: %v, req: %+v, rid: %s", err, targetReq, cts.Kit.Rid)
		return nil, err
	}

	return core.ListResult{Count: resp.Count}, nil
}

func (svc *lbSvc) listTargetByTopo(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.LbTopoReq) (any, error) {
	// 查询target关联的clb拓扑信息
	info, err := svc.getTargetTopoInfoByReq(kt, bizID, vendor, &req.LbTopoCond)
	if err != nil {
		logs.Errorf("get clb topo info failed, err: %v, bizID: %d, vendor: %s, req: %+v, rid: %s", err, bizID, vendor,
			req, kt.Rid)
		return nil, err
	}
	if !info.Match {
		return core.ListResultT[cslb.InstWithTargets]{Details: make([]cslb.InstWithTargets, 0)}, nil
	}

	// 根据条件查询对应的instance信息
	targetInstCond := make([]filter.RuleFactory, 0)
	tgIDs := maps.Keys(info.TgMap)
	targetInstCond = append(targetInstCond, tools.RuleIn("target_group_id", tgIDs))
	targetInstCond = append(targetInstCond, req.GetTargetCond()...)
	instInfoReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: targetInstCond},
		Page:   req.Page,
	}
	instInfoResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetInstInfo(kt, &instInfoReq)
	if err != nil {
		logs.Errorf("get instance info failed, err: %v, instInfoReq: %+v, rid: %s", err, instInfoReq, kt.Rid)
		return nil, err
	}
	if req.Page.Count {
		return core.ListResultT[cslb.InstWithTargets]{Count: instInfoResp.Count}, nil
	}
	if len(instInfoResp.Details) == 0 {
		return core.ListResultT[cslb.InstWithTargets]{Details: make([]cslb.InstWithTargets, 0)}, nil
	}

	// 根据条件查询RS信息
	targetCond := make([]filter.RuleFactory, 0)
	targetCond = append(targetCond, tools.RuleIn("target_group_id", tgIDs))
	ips := make([]string, 0)
	for _, cvmInfo := range instInfoResp.Details {
		ips = append(ips, cvmInfo.IP)
	}
	targetCond = append(targetCond, tools.RuleIn("ip", ips))
	if len(req.TargetPorts) != 0 {
		targetCond = append(targetCond, tools.RuleIn("port", req.TargetPorts))
	}
	targets, err := svc.getTargetByCond(kt, targetCond)
	if err != nil {
		logs.Errorf("get target failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
		return nil, err
	}

	// 组装数据进行返回
	details, err := buildInstWithTargetsInfo(kt, info, targets, instInfoResp.Details)
	if err != nil {
		logs.Errorf("build cvm with targets info failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return core.ListResultT[cslb.InstWithTargets]{Details: details}, nil
}

func (svc *lbSvc) getTargetTopoInfoByReq(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.LbTopoCond) (
	*cslb.TargetTopoInfo, error) {

	commonCond := make([]filter.RuleFactory, 0)
	commonCond = append(commonCond, tools.RuleEqual("bk_biz_id", bizID))
	commonCond = append(commonCond, tools.RuleEqual("vendor", vendor))
	commonCond = append(commonCond, tools.RuleEqual("account_id", req.AccountID))
	// 根据条件查询clb信息
	lbCond := make([]filter.RuleFactory, 0)
	lbCond = append(lbCond, commonCond...)
	lbCond = append(lbCond, req.GetLbCond()...)
	lbMap, err := svc.getLbByCond(kt, lbCond)
	if err != nil {
		logs.Errorf("get lb by cond failed, err: %v, lbCond: %v, rid: %s", err, lbCond, kt.Rid)
		return nil, err
	}
	if len(lbMap) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}

	// 根据条件查询监听器信息
	lbIDs := maps.Keys(lbMap)
	lblCond := make([]filter.RuleFactory, 0)
	lblCond = append(lblCond, tools.RuleIn("lb_id", lbIDs))
	lblCond = append(lblCond, req.GetLblCond()...)
	lblMap, err := svc.getLblByCond(kt, vendor, lblCond)
	if err != nil {
		logs.Errorf("get lbl by cond failed, err: %v, lblCond: %v, rid: %s", err, lblCond, kt.Rid)
		return nil, err
	}
	if len(lblMap) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}

	// 根据条件查询规则
	lblIDs := maps.Keys(lblMap)
	ruleCond := make([]filter.RuleFactory, 0)
	ruleCond = append(ruleCond, tools.RuleIn("lbl_id", lblIDs))
	ruleCond = append(ruleCond, req.GetRuleCond()...)
	ruleMap, err := svc.getRuleByCond(kt, vendor, ruleCond)
	if err != nil {
		logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
		return nil, err
	}
	if len(ruleMap) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}

	// 根据条件查询clb和目标组关系, 注：tgLbRelCond中的vendor条件不能去掉，不同vendor的规则在不同表里，自增id不共用，不加的话可能串数据
	ruleIDs := maps.Keys(ruleMap)
	tgLbRelCond := []filter.RuleFactory{tools.RuleIn("listener_rule_id", ruleIDs), tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
	tgLbRels, err := svc.getTgLbRelByCond(kt, tgLbRelCond)
	if err != nil {
		logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
		return nil, err
	}
	if len(tgLbRels) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}
	tgIDMap := make(map[string]struct{})
	for _, tgLbRel := range tgLbRels {
		tgIDMap[tgLbRel.TargetGroupID] = struct{}{}
	}

	// 根据条件查询目标组
	tgIDs := maps.Keys(tgIDMap)
	tgCond := []filter.RuleFactory{tools.RuleIn("id", tgIDs)}
	tgMap, err := svc.getTgByCond(kt, tgCond)
	if err != nil {
		logs.Errorf("get tg by cond failed, err: %v, tgCond: %v, rid: %s", err, tgCond, kt.Rid)
		return nil, err
	}
	if len(tgMap) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}

	return &cslb.TargetTopoInfo{
		Match: true, LbMap: lbMap, LblMap: lblMap, RuleMap: ruleMap, TgLbRels: tgLbRels, TgMap: tgMap,
	}, nil
}

func buildInstWithTargetsInfo(kt *kit.Kit, clbTopoInfo *cslb.TargetTopoInfo, targets []corelb.BaseTarget,
	instInfos []daotypeslb.ListInstInfo) ([]cslb.InstWithTargets, error) {

	if clbTopoInfo == nil || len(targets) == 0 || len(instInfos) == 0 {
		return make([]cslb.InstWithTargets, 0), nil
	}
	tgIDRelMap := make(map[string]corelb.BaseTargetListenerRuleRel)
	for _, rel := range clbTopoInfo.TgLbRels {
		tgIDRelMap[rel.TargetGroupID] = rel
	}
	ipTargetsMap := make(map[string][]cslb.TargetWithTopo)
	for _, target := range targets {
		tgID := target.TargetGroupID
		targetGroup, ok := clbTopoInfo.TgMap[tgID]
		if !ok {
			logs.Errorf("target group not found, tgID: %s, rid: %s", tgID, kt.Rid)
			return nil, fmt.Errorf("target group not found, id: %s", tgID)
		}
		rel, ok := tgIDRelMap[tgID]
		if !ok {
			logs.Errorf("tg lb rel not found, tgID: %s, rid: %s", tgID, kt.Rid)
			return nil, fmt.Errorf("target group loadBalancer relation not found, target group id: %s", tgID)
		}
		lb, ok := clbTopoInfo.LbMap[rel.LbID]
		if !ok {
			logs.Errorf("lb not found, lbID: %s, rid: %s", rel.LbID, kt.Rid)
			return nil, fmt.Errorf("loadBalancer not found, id: %s", rel.LbID)
		}
		lbl, ok := clbTopoInfo.LblMap[rel.LblID]
		if !ok {
			logs.Errorf("lbl not found, lblID: %s, rid: %s", rel.LblID, kt.Rid)
			return nil, fmt.Errorf("listener not found, id: %s", rel.LblID)
		}
		var lblEndPort *int64
		if lbl.Extension != nil {
			lblEndPort = lbl.Extension.EndPort
		}
		rule, ok := clbTopoInfo.RuleMap[rel.ListenerRuleID]
		if !ok {
			logs.Errorf("rule not found, ruleID: %s, rid: %s", rel.ListenerRuleID, kt.Rid)
			return nil, fmt.Errorf("rule not found, id: %s", rel.ListenerRuleID)
		}
		targetWithTopo := cslb.TargetWithTopo{
			BaseTarget:      target,
			TargetGroupName: targetGroup.Name,
			LbID:            lb.ID,
			CloudLbID:       lb.CloudID,
			LbVips:          getLbVips(lb),
			LbDomain:        lb.Domain,
			LbRegion:        lb.Region,
			LbNetworkType:   typeslb.TCloudLoadBalancerType(lb.LoadBalancerType),
			LblID:           lbl.ID,
			LblPort:         lbl.Port,
			LblEndPort:      lblEndPort,
			LblName:         lbl.Name,
			LblProtocol:     lbl.Protocol,
			RuleID:          rule.ID,
			RuleUrl:         rule.URL,
			RuleDomain:      rule.Domain,
		}
		ipTargetsMap[target.IP] = append(ipTargetsMap[target.IP], targetWithTopo)
	}

	details := make([]cslb.InstWithTargets, 0)
	for _, instInfo := range instInfos {
		instWithTargets := cslb.InstWithTargets{
			InstID:      instInfo.InstID,
			InstType:    instInfo.InstType,
			InstName:    instInfo.InstName,
			IP:          instInfo.IP,
			Zone:        instInfo.Zone,
			CloudVpcIDs: instInfo.CloudVpcIDs,
		}
		targetWithTopos, ok := ipTargetsMap[instInfo.IP]
		if ok {
			instWithTargets.Targets = targetWithTopos
		}
		details = append(details, instWithTargets)
	}
	return details, nil
}

func getLbVips(lb corelb.BaseLoadBalancer) []string {
	lbVips := make([]string, 0)
	lbVips = append(lbVips, lb.PrivateIPv4Addresses...)
	lbVips = append(lbVips, lb.PublicIPv4Addresses...)
	lbVips = append(lbVips, lb.PrivateIPv6Addresses...)
	lbVips = append(lbVips, lb.PublicIPv6Addresses...)

	return lbVips
}

// getLbByCond get lb by condition
func (svc *lbSvc) getLbByCond(kt *kit.Kit, lbCond []filter.RuleFactory) (map[string]corelb.BaseLoadBalancer, error) {
	if len(lbCond) == 0 {
		return nil, errors.New("no lb condition")
	}

	lbReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: lbCond},
		Page:   core.NewDefaultBasePage(),
	}
	lbMap := make(map[string]corelb.BaseLoadBalancer)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, &lbReq)
		if err != nil {
			logs.Errorf("get lb failed, err: %v, req: %+v, rid: %s", err, lbReq, kt.Rid)
			return nil, err
		}

		for _, lb := range resp.Details {
			lbMap[lb.ID] = lb
		}
		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		lbReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return lbMap, nil
}

// getLblByCond get lbl by condition
func (svc *lbSvc) getLblByCond(kt *kit.Kit, vendor enumor.Vendor, lblCond []filter.RuleFactory) (
	map[string]corelb.TCloudListener, error) {

	if len(lblCond) == 0 {
		return nil, errors.New("no lbl condition")
	}

	lblReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: lblCond},
		Page:   core.NewDefaultBasePage(),
	}
	lblMap := make(map[string]corelb.TCloudListener)
	for {
		resp := &cloud.TCloudListenerListResult{}
		var err error
		switch vendor {
		case enumor.TCloud:
			resp, err = svc.client.DataService().TCloud.LoadBalancer.ListListener(kt, &lblReq)
			if err != nil {
				logs.Errorf("get lbl failed, err: %v, req: %+v, rid: %s", err, lblReq, kt.Rid)
				return nil, err
			}
		default:
			return nil, fmt.Errorf("vendor: %s not support", vendor)
		}

		for _, lbl := range resp.Details {
			lblMap[lbl.ID] = lbl
		}
		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		lblReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return lblMap, nil
}

// getRuleByCond get rule by condition
func (svc *lbSvc) getRuleByCond(kt *kit.Kit, vendor enumor.Vendor, ruleCond []filter.RuleFactory) (
	map[string]corelb.TCloudLbUrlRule, error) {

	if len(ruleCond) == 0 {
		return nil, errors.New("no rule condition")
	}

	ruleReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: ruleCond},
		Page:   core.NewDefaultBasePage(),
	}
	ruleMap := make(map[string]corelb.TCloudLbUrlRule)
	for {
		resp := &cloud.TCloudURLRuleListResult{}
		var err error
		switch vendor {
		case enumor.TCloud:
			resp, err = svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &ruleReq)
			if err != nil {
				logs.Errorf("get rule failed, err: %v, req: %+v, rid: %s", err, ruleReq, kt.Rid)
				return nil, err
			}
		default:
			return nil, fmt.Errorf("vendor: %s not support", vendor)
		}

		for _, rule := range resp.Details {
			ruleMap[rule.ID] = rule
		}
		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		ruleReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return ruleMap, nil
}

// getTgLbRelByCond get target group and clb relation by condition
func (svc *lbSvc) getTgLbRelByCond(kt *kit.Kit, tgLbRelCond []filter.RuleFactory) ([]corelb.BaseTargetListenerRuleRel,
	error) {

	if len(tgLbRelCond) == 0 {
		return nil, errors.New("no tg lb rel condition")
	}

	tgLbRelReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: tgLbRelCond},
		Page:   core.NewDefaultBasePage(),
	}
	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &tgLbRelReq)
		if err != nil {
			logs.Errorf("get tg lb rel failed, err: %v, req: %+v, rid: %s", err, tgLbRelReq, kt.Rid)
			return nil, err
		}

		tgLbRels = append(tgLbRels, resp.Details...)
		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		tgLbRelReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return tgLbRels, nil
}

// getTgByCond get target group by condition
func (svc *lbSvc) getTgByCond(kt *kit.Kit, tgCond []filter.RuleFactory) (map[string]corelb.BaseTargetGroup, error) {
	if len(tgCond) == 0 {
		return nil, errors.New("no tg condition")
	}

	tgReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: tgCond},
		Page:   core.NewDefaultBasePage(),
	}
	tgMap := make(map[string]corelb.BaseTargetGroup)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, &tgReq)
		if err != nil {
			logs.Errorf("get tg failed, err: %v, req: %+v, rid: %s", err, tgReq, kt.Rid)
			return nil, err
		}

		for _, detail := range resp.Details {
			tgMap[detail.ID] = detail
		}
		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		tgReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return tgMap, nil
}

// getTgLbRelByCond get target by condition
func (svc *lbSvc) getTargetByCond(kt *kit.Kit, targetCond []filter.RuleFactory) ([]corelb.BaseTarget, error) {
	if len(targetCond) == 0 {
		return nil, errors.New("no target condition")
	}

	targetReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: targetCond},
		Page:   core.NewDefaultBasePage(),
	}
	targets := make([]corelb.BaseTarget, 0)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, &targetReq)
		if err != nil {
			logs.Errorf("get target failed, err: %v, req: %+v, rid: %s", err, targetReq, kt.Rid)
			return nil, err
		}

		targets = append(targets, resp.Details...)
		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		targetReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return targets, nil
}

// ListListenerByTopo list listener by topo
func (svc *lbSvc) ListListenerByTopo(cts *rest.Contexts) (any, error) {
	req := new(cslb.LbTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	return svc.listListenerByTopo(cts.Kit, bizID, vendor, req)
}

func (svc *lbSvc) listListenerByTopo(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.LbTopoReq) (any, error) {
	info, err := svc.getLblTopoInfoByReq(kt, bizID, vendor, req)
	if err != nil {
		logs.Errorf("list listener topo info by req failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if !info.Match {
		return core.ListResultT[cslb.ListenerWithTopo]{Details: make([]cslb.ListenerWithTopo, 0)}, nil
	}

	lblCond := make([]filter.RuleFactory, 0)
	lblCond = append(lblCond, info.LblCond...)
	lblCond = append(lblCond, req.GetLblCond()...)
	lblReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: lblCond},
		Page:   req.Page,
	}
	resp := &cloud.TCloudListenerListResult{}
	switch vendor {
	case enumor.TCloud:
		resp, err = svc.client.DataService().TCloud.LoadBalancer.ListListener(kt, &lblReq)
		if err != nil {
			logs.Errorf("get lbl failed, err: %v, req: %+v, rid: %s", err, lblReq, kt.Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}
	if req.Page.Count {
		return core.ListResultT[cslb.ListenerWithTopo]{Count: resp.Count}, nil
	}
	if len(resp.Details) == 0 {
		return core.ListResultT[cslb.ListenerWithTopo]{Details: make([]cslb.ListenerWithTopo, 0)}, nil
	}

	details, err := svc.buildListenerWithTopoInfo(kt, vendor, info, resp.Details)
	if err != nil {
		logs.Errorf("build listener with topo info failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return core.ListResultT[cslb.ListenerWithTopo]{Details: details}, nil
}

func (svc *lbSvc) buildListenerWithTopoInfo(kt *kit.Kit, vendor enumor.Vendor, info *cslb.LblTopoInfo,
	listeners []corelb.TCloudListener) ([]cslb.ListenerWithTopo, error) {

	lblIDRulesMap, lblTargetCountMap, lblNonZeroWeightTargetCountMap, lblIDTgIDMap, err := svc.getListenerRelInfo(kt,
		vendor, listeners)
	if err != nil {
		logs.Errorf("get listener rule and target count failed, err: %v, listeners: %+v, rid: %s", err, listeners,
			kt.Rid)
		return nil, err
	}

	details := make([]cslb.ListenerWithTopo, 0)
	for _, lbl := range listeners {
		lb, ok := info.LbMap[lbl.LbID]
		if !ok {
			logs.Errorf("lb not found, lbID: %s, rid: %s", lbl.LbID, kt.Rid)
			return nil, fmt.Errorf("lb not found, lbID: %s", lbl.LbID)
		}
		var endPort *int64
		if lbl.Extension != nil {
			endPort = lbl.Extension.EndPort
		}

		ruleDomainCount, urlCount := 0, 0
		var scheduler string
		rules := lblIDRulesMap[lbl.ID]
		var targetGroupID string
		if lbl.Protocol.IsLayer7Protocol() && len(rules) != 0 {
			domains := converter.SliceToMap(rules,
				func(r corelb.TCloudLbUrlRule) (string, struct{}) { return r.Domain, struct{}{} })
			ruleDomainCount = len(domains)
			urlCount = len(rules)
		}
		if lbl.Protocol.IsLayer4Protocol() && len(rules) != 0 {
			scheduler = rules[0].Scheduler
			targetGroupID = lblIDTgIDMap[lbl.ID]
		}

		detail := cslb.ListenerWithTopo{
			BaseListener:             converter.PtrToVal(lbl.BaseListener),
			EndPort:                  endPort,
			Scheduler:                scheduler,
			LbVips:                   getLbVips(lb),
			LbDomain:                 lb.Domain,
			LbRegion:                 lb.Region,
			LbNetworkType:            typeslb.TCloudLoadBalancerType(lb.LoadBalancerType),
			RuleDomainCount:          ruleDomainCount,
			UrlCount:                 urlCount,
			TargetCount:              lblTargetCountMap[lbl.ID],
			NonZeroWeightTargetCount: lblNonZeroWeightTargetCountMap[lbl.ID],
			TargetGroupID:            targetGroupID,
		}
		details = append(details, detail)
	}

	return details, nil
}

func (svc *lbSvc) getListenerRelInfo(kt *kit.Kit, vendor enumor.Vendor, listeners []corelb.TCloudListener) (
	map[string][]corelb.TCloudLbUrlRule, map[string]int, map[string]int, map[string]string, error) {

	if len(listeners) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("listeners is empty")
	}

	// 查询监听器规则
	lblIDs := make([]string, 0)
	for _, lbl := range listeners {
		lblIDs = append(lblIDs, lbl.ID)
	}
	lblIDRulesMap := make(map[string][]corelb.TCloudLbUrlRule)
	ruleCond := []filter.RuleFactory{tools.RuleIn("lbl_id", lblIDs)}
	ruleMap, err := svc.getRuleByCond(kt, vendor, ruleCond)
	if err != nil {
		logs.Errorf("get rule by cond failed, err: %v, req: %+v, rid: %s", err, ruleCond, kt.Rid)
		return nil, nil, nil, nil, err
	}
	for _, rule := range maps.Values(ruleMap) {
		if _, ok := lblIDRulesMap[rule.LblID]; !ok {
			lblIDRulesMap[rule.LblID] = make([]corelb.TCloudLbUrlRule, 0)
		}
		lblIDRulesMap[rule.LblID] = append(lblIDRulesMap[rule.LblID], rule)
	}

	// 获取监听器关联的target数量和权重不为0的target数量, 监听器关联的目标组
	tgLbRels, err := svc.getTgLbRelByCond(kt, []filter.RuleFactory{tools.RuleIn("lbl_id", lblIDs)})
	if err != nil {
		logs.Errorf("get tg lb rel by cond failed, err: %v, lblIDs: %+v, rid: %s", err, lblIDs, kt.Rid)
		return nil, nil, nil, nil, err
	}
	tgIDLblIDMap := make(map[string]string)
	lblIDTgIDMap := make(map[string]string)
	tgIDs := make([]string, 0)
	tgIDBindStatusMap := make(map[string]enumor.BindingStatus)
	for _, tgLbRel := range tgLbRels {
		tgIDLblIDMap[tgLbRel.TargetGroupID] = tgLbRel.LblID
		lblIDTgIDMap[tgLbRel.LblID] = tgLbRel.TargetGroupID
		tgIDs = append(tgIDs, tgLbRel.TargetGroupID)
		tgIDBindStatusMap[tgLbRel.TargetGroupID] = tgLbRel.BindingStatus
	}
	targets, err := svc.getTargetByCond(kt, []filter.RuleFactory{tools.RuleIn("target_group_id", tgIDs)})
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, tgIDs: %+v, rid: %s", err, tgIDs, kt.Rid)
		return nil, nil, nil, nil, err
	}
	lblTargetCountMap := make(map[string]int)
	lblNonZeroWeightTargetCountMap := make(map[string]int)
	for _, target := range targets {
		lblID, ok := tgIDLblIDMap[target.TargetGroupID]
		if !ok {
			return nil, nil, nil, nil, fmt.Errorf("target group not found, tg id: %s, target id: %s",
				target.TargetGroupID, target.ID)
		}
		if tgIDBindStatusMap[target.TargetGroupID] != enumor.SuccessBindingStatus {
			continue
		}

		lblTargetCountMap[lblID]++
		if converter.PtrToVal(target.Weight) != 0 {
			lblNonZeroWeightTargetCountMap[lblID]++
		}
	}

	return lblIDRulesMap, lblTargetCountMap, lblNonZeroWeightTargetCountMap, lblIDTgIDMap, nil
}

func (svc *lbSvc) getLblTopoInfoByReq(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.LbTopoReq) (
	*cslb.LblTopoInfo, error) {

	commonCond := make([]filter.RuleFactory, 0)
	commonCond = append(commonCond, tools.RuleEqual("bk_biz_id", bizID))
	commonCond = append(commonCond, tools.RuleEqual("vendor", vendor))
	commonCond = append(commonCond, tools.RuleEqual("account_id", req.AccountID))

	// 根据条件查询clb信息
	lbCond := make([]filter.RuleFactory, 0)
	lbCond = append(lbCond, commonCond...)
	lbCond = append(lbCond, req.GetLbCond()...)
	lbMap, err := svc.getLbByCond(kt, lbCond)
	if err != nil {
		logs.Errorf("get lb by cond failed, err: %v, lbCond: %v, rid: %s", err, lbCond, kt.Rid)
		return nil, err
	}
	if len(lbMap) == 0 {
		return &cslb.LblTopoInfo{Match: false}, nil
	}
	lbIDs := maps.Keys(lbMap)

	reqRuleCond := req.GetRuleCond()
	reqTargetCond := req.GetTargetCond()
	// 如果请求没有规则和RS条件，那么可以直接返回CLB匹配的监听器条件
	if len(reqRuleCond) == 0 && len(reqTargetCond) == 0 {
		lblCond := []filter.RuleFactory{tools.RuleIn("lb_id", lbIDs)}
		return &cslb.LblTopoInfo{Match: true, LbMap: lbMap, LblCond: lblCond}, nil
	}

	tgLbRelCond := []filter.RuleFactory{tools.RuleIn("lb_id", lbIDs),
		tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}

	// 如果请求中存在规则条件，那么需要根据条件查询规则，进一步得到匹配的监听器条件
	if len(reqRuleCond) != 0 {
		ruleCond := make([]filter.RuleFactory, 0)
		ruleCond = append(ruleCond, tools.RuleIn("lb_id", lbIDs))
		ruleCond = append(ruleCond, reqRuleCond...)
		ruleMap, err := svc.getRuleByCond(kt, vendor, ruleCond)
		if err != nil {
			logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
			return nil, err
		}
		if len(ruleMap) == 0 {
			return &cslb.LblTopoInfo{Match: false}, nil
		}

		// 如果请求中不含RS的条件，那么可以直接返回监听器条件
		if len(reqTargetCond) == 0 {
			lblIDMap := make(map[string]struct{})
			for _, rule := range ruleMap {
				lblIDMap[rule.LblID] = struct{}{}
			}
			lblIDs := maps.Keys(lblIDMap)
			lblCond := []filter.RuleFactory{tools.RuleIn("id", lblIDs)}
			return &cslb.LblTopoInfo{Match: true, LbMap: lbMap, LblCond: lblCond}, nil
		}
		// 注：tgLbRelCond中的vendor条件不能去掉，不同vendor的规则在不同表里，自增id不共用，不加的话可能串数据
		tgLbRelCond = []filter.RuleFactory{tools.RuleIn("listener_rule_id", maps.Keys(ruleMap)),
			tools.RuleEqual("vendor", vendor), tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
	}

	// 根据RS条件查询，得到监听器条件
	lblCond, err := svc.getLblCondByTargetCond(kt, tgLbRelCond, reqTargetCond)
	if err != nil {
		logs.Errorf("get lbl cond by target cond failed, err: %v, tgLbRelCond: %v, reqTargetCond: %v, rid: %s", err,
			tgLbRelCond, reqTargetCond, kt.Rid)
		return nil, err
	}
	if len(lblCond) == 0 {
		return &cslb.LblTopoInfo{Match: false}, nil
	}

	return &cslb.LblTopoInfo{Match: true, LbMap: lbMap, LblCond: lblCond}, nil
}

func (svc *lbSvc) getLblCondByTargetCond(kt *kit.Kit, tgLbRelCond []filter.RuleFactory,
	reqTargetCond []filter.RuleFactory) ([]filter.RuleFactory, error) {

	// 根据条件查询clb和目标组关系
	tgLbRels, err := svc.getTgLbRelByCond(kt, tgLbRelCond)
	if err != nil {
		logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
		return nil, err
	}
	if len(tgLbRels) == 0 {
		return make([]filter.RuleFactory, 0), nil
	}

	tgIDLblIDMap := make(map[string]string)
	for _, tgLbRel := range tgLbRels {
		tgIDLblIDMap[tgLbRel.TargetGroupID] = tgLbRel.LblID
	}

	// 根据条件查询RS
	targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", maps.Keys(tgIDLblIDMap))}
	targetCond = append(targetCond, reqTargetCond...)
	targets, err := svc.getTargetByCond(kt, targetCond)
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
		return nil, err
	}
	if len(targets) == 0 {
		return make([]filter.RuleFactory, 0), nil
	}

	// 根据RS反向推出匹配的监听器条件
	lblIDMap := make(map[string]struct{})
	for _, target := range targets {
		lblID, ok := tgIDLblIDMap[target.TargetGroupID]
		if !ok {
			logs.Errorf("use target group id not found listener, tgID: %s, rid: %s", target.TargetGroupID, kt.Rid)
			return nil, fmt.Errorf("use target group not found listener, tgID: %s", target.TargetGroupID)
		}
		lblIDMap[lblID] = struct{}{}
	}
	lblIDs := maps.Keys(lblIDMap)
	if len(lblIDs) == 0 {
		return make([]filter.RuleFactory, 0), nil
	}

	return []filter.RuleFactory{tools.RuleIn("id", lblIDs)}, nil
}

// ListUrlRulesByTopo list url rules by topo
func (svc *lbSvc) ListUrlRulesByTopo(cts *rest.Contexts) (any, error) {
	req := new(cslb.LbTopoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	return svc.listUrlRulesByTopo(cts.Kit, bizID, vendor, req)
}

func (svc *lbSvc) listUrlRulesByTopo(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.LbTopoReq) (any, error) {

	info, err := svc.getUrlRuleTopoInfoByReq(kt, bizID, vendor, req)
	if err != nil {
		logs.Errorf("list url rule topo info by req failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if !info.Match {
		return core.ListResultT[cslb.UrlRuleWithTopo]{Details: make([]cslb.UrlRuleWithTopo, 0)}, nil
	}

	ruleCond := make([]filter.RuleFactory, 0)
	ruleCond = append(ruleCond, info.RuleCond...)
	ruleCond = append(ruleCond, req.GetRuleCond()...)
	ruleCond = append(ruleCond, tools.RuleEqual("rule_type", enumor.Layer7RuleType))

	ruleReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: ruleCond},
		Page:   req.Page,
	}

	resp := &cloud.TCloudURLRuleListResult{}
	switch vendor {
	case enumor.TCloud:
		resp, err = svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, &ruleReq)
		if err != nil {
			logs.Errorf("get url rule failed, err: %v, req: %+v, rid: %s", err, ruleReq, kt.Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}

	if req.Page.Count {
		return core.ListResultT[cslb.UrlRuleWithTopo]{Count: resp.Count}, nil
	}
	if len(resp.Details) == 0 {
		return core.ListResultT[cslb.UrlRuleWithTopo]{Details: make([]cslb.UrlRuleWithTopo, 0)}, nil
	}

	details, err := svc.buildUrlRuleWithTopoInfo(kt, vendor, info, resp.Details)
	if err != nil {
		logs.Errorf("build url rule with topo info failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return core.ListResultT[cslb.UrlRuleWithTopo]{Details: details}, nil
}

func (svc *lbSvc) getUrlRuleTopoInfoByReq(kt *kit.Kit, bizID int64, vendor enumor.Vendor, req *cslb.LbTopoReq) (
	*cslb.UrlRuleTopoInfo, error) {

	commonCond := make([]filter.RuleFactory, 0)
	commonCond = append(commonCond, tools.RuleEqual("bk_biz_id", bizID))
	commonCond = append(commonCond, tools.RuleEqual("vendor", vendor))
	commonCond = append(commonCond, tools.RuleEqual("account_id", req.AccountID))

	lbCond := make([]filter.RuleFactory, 0)
	lbCond = append(lbCond, commonCond...)
	lbCond = append(lbCond, req.GetLbCond()...)
	lbMap, err := svc.getLbByCond(kt, lbCond)
	if err != nil {
		logs.Errorf("get lb by cond failed, err: %v, lbCond: %v, rid: %s", err, lbCond, kt.Rid)
		return nil, err
	}
	if len(lbMap) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}

	lbIDs := maps.Keys(lbMap)
	reqLblCond := req.GetLblCond()
	reqTargetCond := req.GetTargetCond()

	// 如果请求没有监听器和RS条件，那么可以直接返回CLB匹配的规则条件
	if len(reqLblCond) == 0 && len(reqTargetCond) == 0 {
		ruleCond := []filter.RuleFactory{tools.RuleIn("lb_id", lbIDs)}
		return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, RuleCond: ruleCond}, nil
	}

	tgLbRelCond := []filter.RuleFactory{tools.RuleIn("lb_id", lbIDs),
		tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}

	// 如果请求中存在监听器条件，那么需要根据条件查询监听器，进一步得到匹配的规则条件
	if len(reqLblCond) != 0 {
		lblCond := make([]filter.RuleFactory, 0)
		lblCond = append(lblCond, tools.RuleIn("lb_id", lbIDs))
		lblCond = append(lblCond, reqLblCond...)
		lblMap, err := svc.getLblByCond(kt, vendor, lblCond)

		if err != nil {
			logs.Errorf("get lbl by cond failed, err: %v, lblCond: %v, rid: %s", err, lblCond, kt.Rid)
			return nil, err
		}
		if len(lblMap) == 0 {
			return &cslb.UrlRuleTopoInfo{Match: false}, nil
		}

		if len(reqTargetCond) == 0 {
			lblIDs := maps.Keys(lblMap)
			ruleCond := []filter.RuleFactory{tools.RuleIn("lbl_id", lblIDs)}
			return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, RuleCond: ruleCond}, nil
		}

		// 注：tgLbRelCond中的vendor条件不能去掉，不同vendor的规则在不同表里，自增id不共用，不加的话可能串数据
		tgLbRelCond = []filter.RuleFactory{tools.RuleIn("lbl_id", maps.Keys(lblMap)),
			tools.RuleEqual("vendor", vendor), tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
	}

	// 根据RS条件查询，得到规则条件
	ruleCond, err := svc.getRuleCondByTargetCond(kt, tgLbRelCond, reqTargetCond)
	if err != nil {
		logs.Errorf("get rule cond by target cond failed, err: %v, tgLbRelCond: %v, reqTargetCond: %v, rid: %s", err,
			tgLbRelCond, reqTargetCond, kt.Rid)
		return nil, err
	}
	if len(ruleCond) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}

	return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, RuleCond: ruleCond}, nil
}

func (svc *lbSvc) getRuleCondByTargetCond(kt *kit.Kit, tgLbRelCond []filter.RuleFactory,
	reqTargetCond []filter.RuleFactory) ([]filter.RuleFactory, error) {

	// 根据条件查询clb和目标组关系
	tgLbRels, err := svc.getTgLbRelByCond(kt, tgLbRelCond)
	if err != nil {
		logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
		return nil, err
	}
	if len(tgLbRels) == 0 {
		return make([]filter.RuleFactory, 0), nil
	}

	tgIDRuleIDMap := make(map[string]string)
	for _, tgLbRel := range tgLbRels {
		tgIDRuleIDMap[tgLbRel.TargetGroupID] = tgLbRel.ListenerRuleID
	}

	// 根据条件查询RS
	targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", maps.Keys(tgIDRuleIDMap))}
	targetCond = append(targetCond, reqTargetCond...)
	targets, err := svc.getTargetByCond(kt, targetCond)
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
		return nil, err
	}
	if len(targets) == 0 {
		return make([]filter.RuleFactory, 0), nil
	}

	// 根据RS反向推出匹配的规则条件
	ruleIDMap := make(map[string]struct{})
	for _, target := range targets {
		ruleID, ok := tgIDRuleIDMap[target.TargetGroupID]
		if !ok {
			logs.Errorf("use target group id not found rule, tgID: %s, rid: %s", target.TargetGroupID, kt.Rid)
			return nil, fmt.Errorf("use target group not found rule, tgID: %s", target.TargetGroupID)
		}
		ruleIDMap[ruleID] = struct{}{}
	}
	ruleIDs := maps.Keys(ruleIDMap)
	if len(ruleIDs) == 0 {
		return make([]filter.RuleFactory, 0), nil
	}

	return []filter.RuleFactory{tools.RuleIn("id", ruleIDs)}, nil
}

func (svc *lbSvc) buildUrlRuleWithTopoInfo(kt *kit.Kit, vendor enumor.Vendor, info *cslb.UrlRuleTopoInfo,
	urlRules []corelb.TCloudLbUrlRule) ([]cslb.UrlRuleWithTopo, error) {

	ruleIDTargetCountMap, err := svc.getUrlRuleTargetCount(kt, vendor, urlRules)
	if err != nil {
		logs.Errorf("get url rule target count failed, err: %v, urlRules: %+v, rid: %s", err, urlRules, kt.Rid)
		return nil, err
	}

	// 获取监听器信息
	lblIDMap := make(map[string]struct{})
	for _, rule := range urlRules {
		lblIDMap[rule.LblID] = struct{}{}
	}
	lblIDs := maps.Keys(lblIDMap)
	lblMap, err := svc.getLblByCond(kt, vendor, []filter.RuleFactory{tools.RuleIn("id", lblIDs)})
	if err != nil {
		logs.Errorf("get lbl by cond failed, err: %v, lblIDs: %+v, rid: %s", err, lblIDs, kt.Rid)
		return nil, err
	}

	details := make([]cslb.UrlRuleWithTopo, 0)
	for _, rule := range urlRules {
		lb, ok := info.LbMap[rule.LbID]
		if !ok {
			logs.Errorf("lb not found, lbID: %s, rid: %s", rule.LbID, kt.Rid)
			return nil, fmt.Errorf("lb not found, lbID: %s", rule.LbID)
		}

		lbl, ok := lblMap[rule.LblID]
		if !ok {
			logs.Errorf("lbl not found, lblID: %s, rid: %s", rule.LblID, kt.Rid)
			return nil, fmt.Errorf("lbl not found, lblID: %s", rule.LblID)
		}

		// 获取CLB的VIP地址
		lbVips := getLbVips(lb)

		detail := cslb.UrlRuleWithTopo{
			ID:          rule.ID,
			LbVips:      lbVips,
			LblProtocol: string(lbl.Protocol),
			LblPort:     int(lbl.Port),
			RuleUrl:     rule.URL,
			RuleDomain:  rule.Domain,
			TargetCount: ruleIDTargetCountMap[rule.ID],
			LbID:        lb.ID,
			CloudLblID:  lbl.CloudID,
		}
		details = append(details, detail)
	}

	return details, nil
}

// getUrlRuleTargetCount 获取规则的RS数量
func (svc *lbSvc) getUrlRuleTargetCount(kt *kit.Kit, vendor enumor.Vendor, rules []corelb.TCloudLbUrlRule) (map[string]int, error) {
	if len(rules) == 0 {
		return make(map[string]int), nil
	}

	ruleIDs := make([]string, 0)
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	tgLbRelCond := []filter.RuleFactory{tools.RuleIn("listener_rule_id", ruleIDs), tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
	tgLbRels, err := svc.getTgLbRelByCond(kt, tgLbRelCond)
	if err != nil {
		logs.Errorf("get tg lb rel by cond failed, err: %v, ruleIDs: %+v, rid: %s", err, ruleIDs, kt.Rid)
		return nil, err
	}
	tgIDRuleIDMap := make(map[string]string)
	tgIDs := make([]string, 0)
	for _, tgLbRel := range tgLbRels {
		tgIDRuleIDMap[tgLbRel.TargetGroupID] = tgLbRel.ListenerRuleID
		tgIDs = append(tgIDs, tgLbRel.TargetGroupID)
	}
	targets, err := svc.getTargetByCond(kt, []filter.RuleFactory{tools.RuleIn("target_group_id", tgIDs)})
	if err != nil {
		logs.Errorf("get target by cond failed, err: %v, tgIDs: %+v, rid: %s", err, tgIDs, kt.Rid)
		return nil, err
	}
	ruleTargetCountMap := make(map[string]int)
	for _, target := range targets {
		ruleID, ok := tgIDRuleIDMap[target.TargetGroupID]
		if !ok {
			return nil, fmt.Errorf("target group not found, tg id: %s, target id: %s",
				target.TargetGroupID, target.ID)
		}
		ruleTargetCountMap[ruleID]++
	}

	return ruleTargetCountMap, nil
}
