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
	"hcm/pkg/criteria/constant"
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
	"hcm/pkg/tools/slice"
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

	var count uint64
	for _, batch := range slice.Split(maps.Keys(info.TgMap), int(core.DefaultMaxPageLimit)) {
		targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", batch)}
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
		count += resp.Count
		if count > constant.CLBDataByTopoCondReturnLimit {
			return nil, errf.Newf(errf.ClbTopoResExceedErr, "target count exceed limit, limit: %d",
				constant.CLBDataByTopoCondReturnLimit)
		}
	}

	return core.ListResult{Count: count}, nil
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

	// 根据条件查询对应的instance、RS信息
	tgIDs := maps.Keys(info.TgMap)
	instInfoMap := make(map[string]daotypeslb.ListInstInfo)
	targets := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		cond := make([]filter.RuleFactory, 0)
		cond = append(cond, tools.RuleIn("target_group_id", batch))
		cond = append(cond, req.GetTargetCond()...)
		instResult, err := svc.getTargetInstInfoByCond(kt, cond)
		if err != nil {
			logs.Errorf("get target inst failed, err: %v, cond: %v, rid: %s", err, cond, kt.Rid)
			return nil, err
		}
		for _, inst := range instResult {
			instInfoMap[inst.Key()] = inst
		}
		if len(instInfoMap) > constant.CLBDataByTopoCondReturnLimit {
			return nil, errf.Newf(errf.ClbTopoResExceedErr, "instance count exceed limit, limit: %d",
				constant.CLBDataByTopoCondReturnLimit)
		}

		targetResult, err := svc.getTargetByCond(kt, cond, make([]string, 0))
		if err != nil {
			logs.Errorf("get target failed, err: %v, targetCond: %v, rid: %s", err, cond, kt.Rid)
			return nil, err
		}
		targets = append(targets, targetResult...)
		if len(targets) > constant.CLBDataByTopoCondReturnLimit {
			return nil, errf.Newf(errf.ClbTopoResExceedErr, "target count exceed limit, limit: %d",
				constant.CLBDataByTopoCondReturnLimit)
		}
	}
	instInfos := maps.Values(instInfoMap)
	if len(instInfos) == 0 || len(targets) == 0 {
		return core.ListResultT[cslb.InstWithTargets]{Details: make([]cslb.InstWithTargets, 0)}, nil
	}

	// 组装数据进行返回
	details, err := buildInstWithTargetsInfo(kt, info, targets, instInfos)
	if err != nil {
		logs.Errorf("build cvm with targets info failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return core.ListResultT[cslb.InstWithTargets]{Details: details, Count: uint64(len(details))}, nil
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
	lblMap := make(map[string]corelb.TCloudListener)
	for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		lblCond := make([]filter.RuleFactory, 0)
		lblCond = append(lblCond, tools.RuleIn("lb_id", batch))
		lblCond = append(lblCond, req.GetLblCond()...)
		batchResult, err := svc.getLblByCond(kt, vendor, lblCond)
		if err != nil {
			logs.Errorf("get lbl by cond failed, err: %v, lblCond: %v, rid: %s", err, lblCond, kt.Rid)
			return nil, err
		}
		lblMap = maps.MapMerge(lblMap, batchResult)
	}
	if len(lblMap) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}

	// 根据条件查询规则
	lblIDs := maps.Keys(lblMap)
	ruleMap := make(map[string]corelb.TCloudLbUrlRule)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		ruleCond := make([]filter.RuleFactory, 0)
		ruleCond = append(ruleCond, tools.RuleIn("lbl_id", batch))
		ruleCond = append(ruleCond, req.GetRuleCond()...)
		batchResult, err := svc.getRuleByCond(kt, vendor, ruleCond, []string{"id", "url", "domain"})
		if err != nil {
			logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
			return nil, err
		}
		ruleMap = maps.MapMerge(ruleMap, batchResult)
	}

	if len(ruleMap) == 0 {
		return &cslb.TargetTopoInfo{Match: false}, nil
	}

	// 根据条件查询clb和目标组关系, 注：tgLbRelCond中的vendor条件不能去掉，不同vendor的规则在不同表里，自增id不共用，不加的话可能串数据
	ruleIDs := maps.Keys(ruleMap)
	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, batch := range slice.Split(ruleIDs, int(core.DefaultMaxPageLimit)) {
		tgLbRelCond := []filter.RuleFactory{tools.RuleIn("listener_rule_id", batch), tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
		batchResult, err := svc.getTgLbRelByCond(kt, tgLbRelCond, make([]string, 0))
		if err != nil {
			logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
			return nil, err
		}
		tgLbRels = append(tgLbRels, batchResult...)
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
	tgMap := make(map[string]corelb.BaseTargetGroup)
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		tgCond := []filter.RuleFactory{tools.RuleIn("id", batch)}
		batchResult, err := svc.getTgByCond(kt, tgCond, []string{"id", "name"})
		if err != nil {
			logs.Errorf("get tg by cond failed, err: %v, tgCond: %v, rid: %s", err, tgCond, kt.Rid)
			return nil, err
		}
		tgMap = maps.MapMerge(tgMap, batchResult)
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
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.CLBTopoFindPageLimit,
		},
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
		if uint(len(resp.Details)) < lblReq.Page.Limit {
			break
		}
		lblReq.Page.Start += uint32(lblReq.Page.Limit)
	}

	return lblMap, nil
}

// getRuleByCond get rule by condition
func (svc *lbSvc) getRuleByCond(kt *kit.Kit, vendor enumor.Vendor, ruleCond []filter.RuleFactory, fields []string) (
	map[string]corelb.TCloudLbUrlRule, error) {

	if len(ruleCond) == 0 {
		return nil, errors.New("no rule condition")
	}

	ruleReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: ruleCond},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.CLBTopoFindPageLimit,
		},
	}
	if len(fields) != 0 {
		ruleReq.Fields = fields
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
		if uint(len(resp.Details)) < ruleReq.Page.Limit {
			break
		}
		ruleReq.Page.Start += uint32(ruleReq.Page.Limit)
	}

	return ruleMap, nil
}

// getTgLbRelByCond get target group and clb relation by condition
func (svc *lbSvc) getTgLbRelByCond(kt *kit.Kit, tgLbRelCond []filter.RuleFactory,
	fields []string) ([]corelb.BaseTargetListenerRuleRel, error) {

	if len(tgLbRelCond) == 0 {
		return nil, errors.New("no tg lb rel condition")
	}

	tgLbRelReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: tgLbRelCond},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.CLBTopoFindPageLimit,
		},
	}
	if len(fields) != 0 {
		tgLbRelReq.Fields = fields
	}

	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &tgLbRelReq)
		if err != nil {
			logs.Errorf("get tg lb rel failed, err: %v, req: %+v, rid: %s", err, tgLbRelReq, kt.Rid)
			return nil, err
		}

		tgLbRels = append(tgLbRels, resp.Details...)
		if uint(len(resp.Details)) < tgLbRelReq.Page.Limit {
			break
		}
		tgLbRelReq.Page.Start += uint32(tgLbRelReq.Page.Limit)
	}

	return tgLbRels, nil
}

// getTgByCond get target group by condition
func (svc *lbSvc) getTgByCond(kt *kit.Kit, tgCond []filter.RuleFactory, fields []string) (
	map[string]corelb.BaseTargetGroup, error) {

	if len(tgCond) == 0 {
		return nil, errors.New("no tg condition")
	}

	tgReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: tgCond},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.CLBTopoFindPageLimit,
		},
	}
	if len(fields) != 0 {
		tgReq.Fields = fields
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
		if uint(len(resp.Details)) < tgReq.Page.Limit {
			break
		}
		tgReq.Page.Start += uint32(tgReq.Page.Limit)
	}

	return tgMap, nil
}

// getTgLbRelByCond get target by condition
func (svc *lbSvc) getTargetByCond(kt *kit.Kit, targetCond []filter.RuleFactory, fields []string) (
	[]corelb.BaseTarget, error) {

	if len(targetCond) == 0 {
		return nil, errors.New("no target condition")
	}

	targetReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: targetCond},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.CLBTopoFindPageLimit,
		},
	}

	if len(fields) != 0 {
		targetReq.Fields = fields
	}

	targets := make([]corelb.BaseTarget, 0)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, &targetReq)
		if err != nil {
			logs.Errorf("get target failed, err: %v, req: %+v, rid: %s", err, targetReq, kt.Rid)
			return nil, err
		}

		targets = append(targets, resp.Details...)
		if uint(len(resp.Details)) < targetReq.Page.Limit {
			break
		}
		targetReq.Page.Start += uint32(targetReq.Page.Limit)
	}

	return targets, nil
}

// getTargetInstInfoByCond get target instance info by condition
func (svc *lbSvc) getTargetInstInfoByCond(kt *kit.Kit, targetInstCond []filter.RuleFactory) (
	[]daotypeslb.ListInstInfo, error) {

	if len(targetInstCond) == 0 {
		return nil, errors.New("no target condition")
	}

	instInfoReq := core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: targetInstCond},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.CLBTopoFindPageLimit,
		},
	}
	instInfos := make([]daotypeslb.ListInstInfo, 0)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetInstInfo(kt, &instInfoReq)
		if err != nil {
			logs.Errorf("get instance info failed, err: %v, instInfoReq: %+v, rid: %s", err, instInfoReq, kt.Rid)
			return nil, err
		}

		instInfos = append(instInfos, resp.Details...)
		if uint(len(resp.Details)) < instInfoReq.Page.Limit {
			break
		}
		instInfoReq.Page.Start += uint32(instInfoReq.Page.Limit)
	}

	return instInfos, nil
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

	listeners := make([]corelb.Listener[corelb.TCloudListenerExtension], 0)
	for _, subLblCond := range info.LblConds {
		lblCond := make([]filter.RuleFactory, 0)
		lblCond = append(lblCond, subLblCond...)
		lblCond = append(lblCond, req.GetLblCond()...)
		subResult, err := svc.getLblByCond(kt, vendor, lblCond)
		if err != nil {
			logs.Errorf("list listener by req failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		listeners = append(listeners, maps.Values(subResult)...)
		if len(listeners) > constant.CLBDataByTopoCondReturnLimit {
			return nil, errf.Newf(errf.ClbTopoResExceedErr, "listener count exceed limit, limit: %d",
				constant.CLBDataByTopoCondReturnLimit)
		}
	}

	if len(listeners) == 0 {
		return core.ListResultT[cslb.ListenerWithTopo]{Details: make([]cslb.ListenerWithTopo, 0)}, nil
	}

	details, err := svc.buildListenerWithTopoInfo(kt, vendor, info, listeners)
	if err != nil {
		logs.Errorf("build listener with topo info failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return core.ListResultT[cslb.ListenerWithTopo]{Details: details, Count: uint64(len(details))}, nil
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
	lblIDs = slice.Unique(lblIDs)
	ruleMap := make(map[string]corelb.TCloudLbUrlRule)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		ruleCond := []filter.RuleFactory{tools.RuleIn("lbl_id", batch)}
		batchResult, err := svc.getRuleByCond(kt, vendor, ruleCond, make([]string, 0))
		if err != nil {
			logs.Errorf("get rule by cond failed, err: %v, req: %+v, rid: %s", err, ruleCond, kt.Rid)
			return nil, nil, nil, nil, err
		}
		ruleMap = maps.MapMerge(ruleMap, batchResult)
	}
	lblIDRulesMap := make(map[string][]corelb.TCloudLbUrlRule)
	for _, rule := range maps.Values(ruleMap) {
		if _, ok := lblIDRulesMap[rule.LblID]; !ok {
			lblIDRulesMap[rule.LblID] = make([]corelb.TCloudLbUrlRule, 0)
		}
		lblIDRulesMap[rule.LblID] = append(lblIDRulesMap[rule.LblID], rule)
	}

	// 获取监听器关联的target数量和权重不为0的target数量, 监听器关联的目标组
	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		batchResult, err := svc.getTgLbRelByCond(kt, []filter.RuleFactory{tools.RuleIn("lbl_id", batch)},
			make([]string, 0))
		if err != nil {
			logs.Errorf("get tg lb rel by cond failed, err: %v, lblIDs: %+v, rid: %s", err, batch, kt.Rid)
			return nil, nil, nil, nil, err
		}
		tgLbRels = append(tgLbRels, batchResult...)
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
	tgIDs = slice.Unique(tgIDs)
	targets := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		batchResult, err := svc.getTargetByCond(kt, []filter.RuleFactory{tools.RuleIn("target_group_id", batch)},
			[]string{"id", "target_group_id", "weight"})
		if err != nil {
			logs.Errorf("get target by cond failed, err: %v, tgIDs: %+v, rid: %s", err, batch, kt.Rid)
			return nil, nil, nil, nil, err
		}
		targets = append(targets, batchResult...)
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
		lblConds := make([][]filter.RuleFactory, 0)
		for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
			lblCond := []filter.RuleFactory{tools.RuleIn("lb_id", batch)}
			lblConds = append(lblConds, lblCond)
		}
		return &cslb.LblTopoInfo{Match: true, LbMap: lbMap, LblConds: lblConds}, nil
	}

	tgLbRelConds := make([][]filter.RuleFactory, 0)
	for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		tgLbRelCond := []filter.RuleFactory{tools.RuleIn("lb_id", batch),
			tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
		tgLbRelConds = append(tgLbRelConds, tgLbRelCond)
	}

	// 如果请求中存在规则条件，那么需要根据条件查询规则，进一步得到匹配的监听器条件
	if len(reqRuleCond) != 0 {
		ruleMap := make(map[string]corelb.TCloudLbUrlRule)
		for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
			ruleCond := make([]filter.RuleFactory, 0)
			ruleCond = append(ruleCond, tools.RuleIn("lb_id", batch))
			ruleCond = append(ruleCond, reqRuleCond...)
			batchResult, err := svc.getRuleByCond(kt, vendor, ruleCond, []string{"id", "lbl_id"})
			if err != nil {
				logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
				return nil, err
			}
			ruleMap = maps.MapMerge(ruleMap, batchResult)
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
			lblConds := make([][]filter.RuleFactory, 0)
			for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
				lblCond := []filter.RuleFactory{tools.RuleIn("id", batch)}
				lblConds = append(lblConds, lblCond)
			}
			return &cslb.LblTopoInfo{Match: true, LbMap: lbMap, LblConds: lblConds}, nil
		}
		// 注：tgLbRelCond中的vendor条件不能去掉，不同vendor的规则在不同表里，自增id不共用，不加的话可能串数据
		tgLbRelConds = make([][]filter.RuleFactory, 0)
		for _, batch := range slice.Split(maps.Keys(ruleMap), int(core.DefaultMaxPageLimit)) {
			tgLbRelCond := []filter.RuleFactory{tools.RuleIn("listener_rule_id", batch),
				tools.RuleEqual("vendor", vendor), tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
			tgLbRelConds = append(tgLbRelConds, tgLbRelCond)
		}
	}

	// 根据RS条件查询，得到监听器条件
	lblConds, err := svc.getLblCondByTargetCond(kt, tgLbRelConds, reqTargetCond)
	if err != nil {
		logs.Errorf("get lbl cond by target cond failed, err: %v, tgLbRelCond: %v, reqTargetCond: %v, rid: %s", err,
			tgLbRelConds, reqTargetCond, kt.Rid)
		return nil, err
	}
	if len(lblConds) == 0 {
		return &cslb.LblTopoInfo{Match: false}, nil
	}

	return &cslb.LblTopoInfo{Match: true, LbMap: lbMap, LblConds: lblConds}, nil
}

func (svc *lbSvc) getLblCondByTargetCond(kt *kit.Kit, tgLbRelConds [][]filter.RuleFactory,
	reqTargetCond []filter.RuleFactory) ([][]filter.RuleFactory, error) {

	// 根据条件查询clb和目标组关系
	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, tgLbRelCond := range tgLbRelConds {
		subResult, err := svc.getTgLbRelByCond(kt, tgLbRelCond, []string{"target_group_id", "lbl_id"})
		if err != nil {
			logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
			return nil, err
		}
		tgLbRels = append(tgLbRels, subResult...)
	}
	if len(tgLbRels) == 0 {
		return make([][]filter.RuleFactory, 0), nil
	}

	tgIDLblIDMap := make(map[string]string)
	for _, tgLbRel := range tgLbRels {
		tgIDLblIDMap[tgLbRel.TargetGroupID] = tgLbRel.LblID
	}

	// 根据条件查询RS
	targets := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(maps.Keys(tgIDLblIDMap), int(core.DefaultMaxPageLimit)) {
		targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", batch)}
		targetCond = append(targetCond, reqTargetCond...)
		subResult, err := svc.getTargetByCond(kt, targetCond, []string{"target_group_id"})
		if err != nil {
			logs.Errorf("get target by cond failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
			return nil, err
		}
		targets = append(targets, subResult...)
	}
	if len(targets) == 0 {
		return make([][]filter.RuleFactory, 0), nil
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
		return make([][]filter.RuleFactory, 0), nil
	}

	result := make([][]filter.RuleFactory, 0)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		result = append(result, []filter.RuleFactory{tools.RuleIn("id", batch)})
	}
	return result, nil
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

	urlRules := make([]corelb.TCloudLbUrlRule, 0)
	for _, subRuleCond := range info.RuleConds {
		ruleCond := make([]filter.RuleFactory, 0)
		ruleCond = append(ruleCond, subRuleCond...)
		ruleCond = append(ruleCond, req.GetRuleCond()...)
		ruleCond = append(ruleCond, tools.RuleEqual("rule_type", enumor.Layer7RuleType))
		batchResult, err := svc.getRuleByCond(kt, vendor, ruleCond, make([]string, 0))
		if err != nil {
			logs.Errorf("get rule by cond failed, err: %v, ruleCond: %v, rid: %s", err, ruleCond, kt.Rid)
			return nil, err
		}
		urlRules = append(urlRules, maps.Values(batchResult)...)

		if len(urlRules) > constant.CLBDataByTopoCondReturnLimit {
			return nil, errf.Newf(errf.ClbTopoResExceedErr, "url rules count exceed limit, limit: %d",
				constant.CLBDataByTopoCondReturnLimit)
		}
	}

	if len(urlRules) == 0 {
		return core.ListResultT[cslb.UrlRuleWithTopo]{Details: make([]cslb.UrlRuleWithTopo, 0)}, nil
	}

	details, err := svc.buildUrlRuleWithTopoInfo(kt, vendor, info, urlRules)
	if err != nil {
		logs.Errorf("build url rule with topo info failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return core.ListResultT[cslb.UrlRuleWithTopo]{Details: details, Count: uint64(len(details))}, nil
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
		ruleConds := make([][]filter.RuleFactory, 0)
		for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
			ruleConds = append(ruleConds, []filter.RuleFactory{tools.RuleIn("lb_id", batch)})
		}
		return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, RuleConds: ruleConds}, nil
	}

	tgLbRelConds := make([][]filter.RuleFactory, 0)
	for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
		tgLbRelConds = append(tgLbRelConds, []filter.RuleFactory{tools.RuleIn("lb_id", batch)})
	}
	// 如果请求中存在监听器条件，那么需要根据条件查询监听器，进一步得到匹配的规则条件
	if len(reqLblCond) != 0 {
		lblMap := make(map[string]corelb.TCloudListener)
		for _, batch := range slice.Split(lbIDs, int(core.DefaultMaxPageLimit)) {
			lblCond := make([]filter.RuleFactory, 0)
			lblCond = append(lblCond, tools.RuleIn("lb_id", batch))
			lblCond = append(lblCond, reqLblCond...)
			batchResult, err := svc.getLblByCond(kt, vendor, lblCond)
			if err != nil {
				logs.Errorf("get lbl by cond failed, err: %v, lblCond: %v, rid: %s", err, lblCond, kt.Rid)
				return nil, err
			}
			lblMap = maps.MapMerge(lblMap, batchResult)
		}
		if len(lblMap) == 0 {
			return &cslb.UrlRuleTopoInfo{Match: false}, nil
		}
		if len(reqTargetCond) == 0 {
			ruleConds := make([][]filter.RuleFactory, 0)
			for _, batch := range slice.Split(maps.Keys(lblMap), int(core.DefaultMaxPageLimit)) {
				ruleConds = append(ruleConds, []filter.RuleFactory{tools.RuleIn("lbl_id", batch)})
			}
			return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, RuleConds: ruleConds}, nil
		}

		// 注：tgLbRelCond中的vendor条件不能去掉，不同vendor的规则在不同表里，自增id不共用，不加的话可能串数据
		tgLbRelConds = make([][]filter.RuleFactory, 0)
		for _, batch := range slice.Split(maps.Keys(lblMap), int(core.DefaultMaxPageLimit)) {
			tgLbRelConds = append(tgLbRelConds, []filter.RuleFactory{tools.RuleIn("lbl_id", batch),
				tools.RuleEqual("vendor", vendor), tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)})
		}
	}

	// 根据RS条件查询，得到规则条件
	ruleConds, err := svc.getRuleCondByTargetCond(kt, tgLbRelConds, reqTargetCond)
	if err != nil {
		logs.Errorf("get rule cond by target cond failed, err: %v, tgLbRelConds: %v, reqTargetCond: %v, rid: %s", err,
			tgLbRelConds, reqTargetCond, kt.Rid)
		return nil, err
	}
	if len(ruleConds) == 0 {
		return &cslb.UrlRuleTopoInfo{Match: false}, nil
	}

	return &cslb.UrlRuleTopoInfo{Match: true, LbMap: lbMap, RuleConds: ruleConds}, nil
}

func (svc *lbSvc) getRuleCondByTargetCond(kt *kit.Kit, tgLbRelConds [][]filter.RuleFactory,
	reqTargetCond []filter.RuleFactory) ([][]filter.RuleFactory, error) {

	// 根据条件查询clb和目标组关系
	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, tgLbRelCond := range tgLbRelConds {
		batchResult, err := svc.getTgLbRelByCond(kt, tgLbRelCond, []string{"target_group_id", "listener_rule_id"})
		if err != nil {
			logs.Errorf("get tg lb rel failed, err: %v, tgLbRelCond: %v, rid: %s", err, tgLbRelCond, kt.Rid)
			return nil, err
		}
		tgLbRels = append(tgLbRels, batchResult...)
	}
	if len(tgLbRels) == 0 {
		return make([][]filter.RuleFactory, 0), nil
	}

	tgIDRuleIDMap := make(map[string]string)
	for _, tgLbRel := range tgLbRels {
		tgIDRuleIDMap[tgLbRel.TargetGroupID] = tgLbRel.ListenerRuleID
	}

	// 根据条件查询RS
	targets := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(maps.Keys(tgIDRuleIDMap), int(core.DefaultMaxPageLimit)) {
		targetCond := []filter.RuleFactory{tools.RuleIn("target_group_id", batch)}
		targetCond = append(targetCond, reqTargetCond...)
		batchResult, err := svc.getTargetByCond(kt, targetCond, []string{"target_group_id"})
		if err != nil {
			logs.Errorf("get target by cond failed, err: %v, targetCond: %v, rid: %s", err, targetCond, kt.Rid)
			return nil, err
		}
		targets = append(targets, batchResult...)
	}
	if len(targets) == 0 {
		return make([][]filter.RuleFactory, 0), nil
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
		return make([][]filter.RuleFactory, 0), nil
	}

	result := make([][]filter.RuleFactory, 0)
	for _, batch := range slice.Split(ruleIDs, int(core.DefaultMaxPageLimit)) {
		result = append(result, []filter.RuleFactory{tools.RuleIn("id", batch)})
	}
	return result, nil
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
	lblMap := make(map[string]corelb.TCloudListener)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		batchResult, err := svc.getLblByCond(kt, vendor, []filter.RuleFactory{tools.RuleIn("id", batch)})
		if err != nil {
			logs.Errorf("get lbl by cond failed, err: %v, lblIDs: %+v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}
		lblMap = maps.MapMerge(lblMap, batchResult)
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
func (svc *lbSvc) getUrlRuleTargetCount(kt *kit.Kit, vendor enumor.Vendor,
	rules []corelb.TCloudLbUrlRule) (map[string]int, error) {
	if len(rules) == 0 {
		return make(map[string]int), nil
	}

	ruleIDs := make([]string, 0)
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	ruleIDs = slice.Unique(ruleIDs)
	tgLbRels := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, batch := range slice.Split(ruleIDs, int(core.DefaultMaxPageLimit)) {
		tgLbRelCond := []filter.RuleFactory{tools.RuleIn("listener_rule_id", batch),
			tools.RuleEqual("vendor", vendor), tools.RuleEqual("binding_status", enumor.SuccessBindingStatus)}
		batchResult, err := svc.getTgLbRelByCond(kt, tgLbRelCond, []string{"target_group_id", "listener_rule_id"})
		if err != nil {
			logs.Errorf("get tg lb rel by cond failed, err: %v, ruleIDs: %+v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}
		tgLbRels = append(tgLbRels, batchResult...)
	}

	tgIDRuleIDMap := make(map[string]string)
	tgIDs := make([]string, 0)
	for _, tgLbRel := range tgLbRels {
		tgIDRuleIDMap[tgLbRel.TargetGroupID] = tgLbRel.ListenerRuleID
		tgIDs = append(tgIDs, tgLbRel.TargetGroupID)
	}
	tgIDs = slice.Unique(tgIDs)
	targets := make([]corelb.BaseTarget, 0)
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		batchResult, err := svc.getTargetByCond(kt, []filter.RuleFactory{tools.RuleIn("target_group_id", batch)},
			[]string{"id", "target_group_id"})
		if err != nil {
			logs.Errorf("get target by cond failed, err: %v, tgIDs: %+v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}
		targets = append(targets, batchResult...)
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
