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

package loadbalancer

import (
	"fmt"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// ListUrlRulesByTopo 查询URL规则信息
func (svc *lbSvc) ListUrlRulesByTopo(cts *rest.Contexts) (any, error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cslb.ListUrlRulesByTopologyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	_, noPermFlag, err := handler.ListBizAuthRes(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		logs.Errorf("list url rules by topo failed, noPermFlag: %v, err: %v, rid: %s", noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		logs.Errorf("list url rules by topo no auth, req: %+v, rid: %s", noPermFlag, req, cts.Kit.Rid)
		return &cslb.ListUrlRulesByTopologyResp{Count: 0, Details: make([]cslb.UrlRuleDetail, 0)}, nil
	}
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	filters, err := svc.buildUrlRuleQueryFilter(cts.Kit, bizID, vendor, req)
	if err != nil {
		return nil, err
	}

	urlRuleList, err := svc.queryUrlRulesByFilter(cts.Kit, vendor, filters, req.Page)
	if err != nil {
		return nil, err
	}

	result, err := svc.buildUrlRuleResponse(cts.Kit, urlRuleList)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// buildUrlRuleQueryFilter 构建URL规则查询条件
func (svc *lbSvc) buildUrlRuleQueryFilter(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) (*filter.Expression, error) {

	lbIDs, err := svc.queryLoadBalancerIDsByUserConditions(kt, bizID, vendor, req)
	if err != nil {
		return nil, fmt.Errorf("query load balancer ids by user conditions failed, err: %v", err)
	}

	if len(lbIDs) == 0 && req.HasLbConditions() {
		return &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				tools.RuleEqual("id", "never_match_condition"),
			},
		}, nil
	}

	var listenerIDs []string
	if req.HasListenerConditions() {
		listenerIDs, err = svc.queryListenerIDsByLbIDs(kt, vendor, req, lbIDs)
		if err != nil {
			return nil, fmt.Errorf("query listener ids by lb ids failed, err: %v", err)
		}

		if len(listenerIDs) == 0 {
			return &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					tools.RuleEqual("id", "never_match_condition"),
				},
			}, nil
		}
	}

	conditions := []*filter.AtomRule{
		tools.RuleIn("rule_type", []string{string(enumor.Layer4RuleType), string(enumor.Layer7RuleType)}),
	}

	if len(lbIDs) > 0 {
		conditions = append(conditions, tools.RuleIn("lb_id", lbIDs))
	}

	if len(listenerIDs) > 0 {
		conditions = append(conditions, tools.RuleIn("cloud_lbl_id", listenerIDs))
	}

	conditions = svc.addRuleConditions(req, conditions)

	conditions, err = svc.addTargetConditions(kt, req, conditions)
	if err != nil {
		return nil, err
	}
	finalFilter := tools.ExpressionAnd(conditions...)
	return finalFilter, nil
}

// addLoadBalancerConditions 添加负载均衡器相关条件
func (svc *lbSvc) addLoadBalancerConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq, conditions []*filter.AtomRule) ([]*filter.AtomRule, error) {

	hasLbConditions := len(req.LbRegions) > 0 || len(req.LbNetworkTypes) > 0 || len(req.LbIpVersions) > 0 ||
		len(req.CloudLbIds) > 0 || len(req.LbVips) > 0 || len(req.LbDomains) > 0

	if hasLbConditions {

		lbIDs, err := svc.queryLoadBalancerIDsByConditions(kt, bizID, vendor, req)
		if err != nil {
			return nil, fmt.Errorf("query load balancer ids failed, err: %v", err)
		}

		if len(lbIDs) == 0 {
			return conditions, nil
		}

		conditions = append(conditions, tools.RuleIn("lb_id", lbIDs))
	}

	return conditions, nil
}

// addRuleConditions 添加规则相关条件
func (svc *lbSvc) addRuleConditions(req *cslb.ListUrlRulesByTopologyReq,
	conditions []*filter.AtomRule) []*filter.AtomRule {
	if req.HasRuleConditions() {
		if len(req.RuleDomains) > 0 {
			conditions = append(conditions, tools.RuleIn("domain", req.RuleDomains))
		}

		if len(req.RuleUrls) > 0 {
			conditions = append(conditions, tools.RuleIn("url", req.RuleUrls))
		}
	}

	return conditions
}

// addTargetConditions 添加目标相关条件，需要查询目标表
func (svc *lbSvc) addTargetConditions(kt *kit.Kit, req *cslb.ListUrlRulesByTopologyReq,
	conditions []*filter.AtomRule) ([]*filter.AtomRule, error) {
	if !req.HasTargetConditions() {
		return conditions, nil
	}

	targetGroupIDs, err := svc.queryTargetGroupIDsByTargetConditions(kt, req)
	if err != nil {
		return nil, fmt.Errorf("query target group ids failed, err: %v", err)
	}

	if len(targetGroupIDs) > 0 {
		conditions = append(conditions, tools.RuleIn("target_group_id", targetGroupIDs))
	}

	return conditions, nil
}

// queryLoadBalancerIDsByUserConditions 根据用户输入的负载均衡器条件查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByUserConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	lbFilter, err := svc.buildLoadBalancerFilter(bizID, vendor, req)
	if err != nil {
		return nil, err
	}
	return svc.queryLoadBalancerIDsByFilter(kt, lbFilter)
}

// queryLoadBalancerIDsByConditions 根据基础条件查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByConditions(kt *kit.Kit, bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	lbConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", req.AccountID),
		tools.RuleEqual("bk_biz_id", bizID),
	}

	lbFilter := tools.ExpressionAnd(lbConditions...)
	return svc.queryLoadBalancerIDsByFilter(kt, lbFilter)
}

// queryListenerIDsByLbIDs 根据CLB ID列表和监听器条件查询监听器ID
func (svc *lbSvc) queryListenerIDsByLbIDs(kt *kit.Kit, vendor enumor.Vendor, req *cslb.ListUrlRulesByTopologyReq, lbIDs []string) ([]string, error) {
	listenerConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", req.AccountID),
	}

	if len(lbIDs) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("lb_id", lbIDs))
	}

	if len(req.LblProtocols) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("protocol", req.LblProtocols))
	}

	if len(req.LblPorts) > 0 {
		listenerConditions = append(listenerConditions, tools.RuleIn("port", req.LblPorts))
	}

	listenerFilter := tools.ExpressionAnd(listenerConditions...)

	listenerReq := &core.ListReq{
		Filter: listenerFilter,
		Page:   core.NewDefaultBasePage(),
	}

	listenerIDs := make([]string, 0)
	for {
		listenerResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, listenerReq)
		if err != nil {
			return nil, err
		}

		for _, listener := range listenerResp.Details {
			listenerIDs = append(listenerIDs, listener.CloudID)
		}

		if uint(len(listenerResp.Details)) < core.DefaultMaxPageLimit {
			break
		}

		listenerReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return listenerIDs, nil
}

// buildLoadBalancerFilter 构建负载均衡器查询过滤器
func (svc *lbSvc) buildLoadBalancerFilter(bizID int64, vendor enumor.Vendor,
	req *cslb.ListUrlRulesByTopologyReq) (*filter.Expression, error) {

	lbConditions := []*filter.AtomRule{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", req.AccountID),
		tools.RuleEqual("bk_biz_id", bizID),
	}

	if req.HasLbConditions() {
		if len(req.LbRegions) > 0 {
			lbConditions = append(lbConditions, tools.RuleIn("region", req.LbRegions))
		}
	}
	var networkTypeFilter *filter.Expression
	if len(req.LbNetworkTypes) > 0 {
		var networkTypeConditions []*filter.AtomRule
		for _, networkType := range req.LbNetworkTypes {
			switch networkType {
			case "OPEN":
				networkTypeConditions = append(networkTypeConditions,
					tools.RuleGreaterThan("extension.internet_max_bandwidth_out", 0))
			case "INTERNAL":
				networkTypeConditions = append(networkTypeConditions,
					tools.RuleEqual("extension.internet_max_bandwidth_out", 0))
			}
		}
		if len(networkTypeConditions) > 0 {
			networkTypeFilter = tools.ExpressionOr(networkTypeConditions...)
		}
	}
	if len(req.LbIpVersions) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("ip_version", req.LbIpVersions))
	}
	if len(req.CloudLbIds) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("cloud_id", req.CloudLbIds))
	}
	if len(req.LbDomains) > 0 {
		lbConditions = append(lbConditions, tools.RuleIn("domain", req.LbDomains))
	}

	baseFilter := tools.ExpressionAnd(lbConditions...)

	filters := []*filter.Expression{baseFilter}

	if networkTypeFilter != nil {
		filters = append(filters, networkTypeFilter)
	}

	if len(req.LbVips) > 0 {
		vipConditions := []*filter.AtomRule{
			tools.RuleJsonOverlaps("private_ipv4_addresses", req.LbVips),
			tools.RuleJsonOverlaps("private_ipv6_addresses", req.LbVips),
			tools.RuleJsonOverlaps("public_ipv4_addresses", req.LbVips),
			tools.RuleJsonOverlaps("public_ipv6_addresses", req.LbVips),
		}
		vipOrFilter := tools.ExpressionOr(vipConditions...)
		filters = append(filters, vipOrFilter)
	}
	if len(filters) > 1 {
		combinedFilter := &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		}
		for _, f := range filters {
			combinedFilter.Rules = append(combinedFilter.Rules, f)
		}
		return combinedFilter, nil
	}
	return baseFilter, nil
}

// queryLoadBalancerIDsByFilter 根据过滤器查询负载均衡器ID
func (svc *lbSvc) queryLoadBalancerIDsByFilter(kt *kit.Kit, filter *filter.Expression) ([]string, error) {
	lbReq := &core.ListReq{Filter: filter, Page: core.NewDefaultBasePage()}
	lbIDs := make([]string, 0)

	for {
		lbResp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
		if err != nil {
			return nil, err
		}

		for _, lb := range lbResp.Details {
			lbIDs = append(lbIDs, lb.ID)
		}

		if uint(len(lbResp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		lbReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return lbIDs, nil
}

// queryTargetGroupIDsByTargetConditions 根据目标条件查询目标组ID
func (svc *lbSvc) queryTargetGroupIDsByTargetConditions(kt *kit.Kit,
	req *cslb.ListUrlRulesByTopologyReq) ([]string, error) {

	targetConditions := []*filter.AtomRule{
		tools.RuleEqual("account_id", req.AccountID),
	}

	if len(req.TargetIps) > 0 {
		targetConditions = append(targetConditions, tools.RuleIn("ip", req.TargetIps))
	}

	if len(req.TargetPorts) > 0 {
		targetConditions = append(targetConditions, tools.RuleIn("port", req.TargetPorts))
	}

	targetFilter := tools.ExpressionAnd(targetConditions...)
	targetReq := &core.ListReq{
		Filter: targetFilter,
		Page:   core.NewDefaultBasePage(),
	}

	targetGroupIDMap := make(map[string]struct{})
	for {
		targetResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
		if err != nil {
			return nil, err
		}

		for _, target := range targetResp.Details {
			if target.TargetGroupID != "" {
				targetGroupIDMap[target.TargetGroupID] = struct{}{}
			}
		}

		if uint(len(targetResp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		targetReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	targetGroupIDs := make([]string, 0, len(targetGroupIDMap))
	for targetGroupID := range targetGroupIDMap {
		targetGroupIDs = append(targetGroupIDs, targetGroupID)
	}

	return targetGroupIDs, nil
}

// queryUrlRulesByFilter 根据条件查询URL规则
func (svc *lbSvc) queryUrlRulesByFilter(kt *kit.Kit, vendor enumor.Vendor,
	filter *filter.Expression, page *core.BasePage) (*dataproto.TCloudURLRuleListResult, error) {

	req := &core.ListReq{
		Filter: filter,
		Page:   page,
	}

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, req)
	default:
		return nil, fmt.Errorf("vendor: %s not support", vendor)
	}
}
func (svc *lbSvc) buildUrlRuleResponse(kt *kit.Kit,
	urlRuleList *dataproto.TCloudURLRuleListResult) (*cslb.ListUrlRulesByTopologyResp, error) {

	result := &cslb.ListUrlRulesByTopologyResp{
		Count:   int(urlRuleList.Count),
		Details: make([]cslb.UrlRuleDetail, 0, len(urlRuleList.Details)),
	}

	if len(urlRuleList.Details) == 0 {
		return result, nil
	}

	lbIDs := make([]string, 0, len(urlRuleList.Details))
	listenerIDs := make([]string, 0, len(urlRuleList.Details))
	targetGroupIDs := make([]string, 0, len(urlRuleList.Details))

	for _, rule := range urlRuleList.Details {
		lbIDs = append(lbIDs, rule.LbID)
		listenerIDs = append(listenerIDs, rule.LblID)
		if rule.TargetGroupID != "" {
			targetGroupIDs = append(targetGroupIDs, rule.TargetGroupID)
		}
	}

	lbMap, err := svc.batchGetLoadBalancerInfo(kt, lbIDs)
	if err != nil {
		logs.Errorf("batch get load balancer info failed, err: %v, rid: %s", err, kt.Rid)
	}

	listenerMap, err := svc.batchGetListenerInfo(kt, listenerIDs)
	if err != nil {
		logs.Errorf("batch get listener info failed, err: %v, rid: %s", err, kt.Rid)
	}

	targetGroupMap, err := svc.batchGetTargetGroupInfo(kt, targetGroupIDs)
	if err != nil {
		logs.Errorf("batch get target group info failed, err: %v, rid: %s", err, kt.Rid)
	}

	targetCountMap, err := svc.batchGetTargetCountByTargetGroupIDs(kt, targetGroupIDs)
	if err != nil {
		logs.Errorf("batch get target count failed, err: %v, rid: %s", err, kt.Rid)
	}

	for _, rule := range urlRuleList.Details {
		detail := svc.buildUrlRuleDetail(rule, lbMap, listenerMap, targetGroupMap, targetCountMap)
		result.Details = append(result.Details, detail)
	}

	return result, nil
}

// buildUrlRuleDetail URL规则详情
func (svc *lbSvc) buildUrlRuleDetail(rule corelb.TCloudLbUrlRule,
	lbMap map[string]*corelb.BaseLoadBalancer,
	listenerMap map[string]*corelb.BaseListener,
	targetGroupMap map[string]*corelb.BaseTargetGroup,
	targetCountMap map[string]int) cslb.UrlRuleDetail {

	detail := cslb.UrlRuleDetail{
		ID: rule.ID,
	}

	if lb, exists := lbMap[rule.LbID]; exists {
		detail.Ip = svc.getLoadBalancerVip(lb)
		detail.LbID = lb.ID
	}

	if listener, exists := listenerMap[rule.LblID]; exists {
		detail.LblProtocols = string(listener.Protocol)
		detail.LblPort = int(listener.Port)
		detail.CloudLblID = listener.CloudID
	}

	if _, exists := targetGroupMap[rule.TargetGroupID]; exists {

		if targetCount, exists := targetCountMap[rule.TargetGroupID]; exists {
			detail.TargetCount = targetCount
		} else {
			detail.TargetCount = 0
		}
	}

	if rule.RuleType == enumor.Layer7RuleType {
		detail.RuleUrl = rule.URL
		detail.RuleDomain = rule.Domain
	} else {
		detail.RuleUrl = ""
		detail.RuleDomain = ""
	}

	return detail
}

// batchGetLoadBalancerInfo 获取负载均衡器信息
func (svc *lbSvc) batchGetLoadBalancerInfo(kt *kit.Kit, lbIDs []string) (map[string]*corelb.BaseLoadBalancer, error) {
	if len(lbIDs) == 0 {
		return make(map[string]*corelb.BaseLoadBalancer), nil
	}

	lbMap, err := lblogic.ListLoadBalancerMap(kt, svc.client.DataService(), lbIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*corelb.BaseLoadBalancer)
	for id, lb := range lbMap {
		result[id] = &lb
	}

	return result, nil
}

// batchGetListenerInfo 获取监听器信息
func (svc *lbSvc) batchGetListenerInfo(kt *kit.Kit, listenerIDs []string) (map[string]*corelb.BaseListener, error) {
	if len(listenerIDs) == 0 {
		return make(map[string]*corelb.BaseListener), nil
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", listenerIDs)),
		Page:   core.NewDefaultBasePage(),
	}

	listenerMap := make(map[string]*corelb.BaseListener)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, req)
		if err != nil {
			return nil, err
		}

		for _, listener := range resp.Details {
			listenerMap[listener.ID] = &listener
		}

		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return listenerMap, nil
}

// batchGetTargetGroupInfo 获取目标组信息
func (svc *lbSvc) batchGetTargetGroupInfo(kt *kit.Kit,
	targetGroupIDs []string) (map[string]*corelb.BaseTargetGroup, error) {
	if len(targetGroupIDs) == 0 {
		return make(map[string]*corelb.BaseTargetGroup), nil
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", targetGroupIDs)),
		Page:   core.NewDefaultBasePage(),
	}

	targetGroupMap := make(map[string]*corelb.BaseTargetGroup)
	for {
		resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, req)
		if err != nil {
			return nil, err
		}

		for _, targetGroup := range resp.Details {
			targetGroupMap[targetGroup.ID] = &targetGroup
		}

		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return targetGroupMap, nil
}

// getLoadBalancerVip 获取负载均衡器VIP
func (svc *lbSvc) getLoadBalancerVip(lb *corelb.BaseLoadBalancer) string {
	if len(lb.PublicIPv4Addresses) > 0 {
		return lb.PublicIPv4Addresses[0]
	}

	if len(lb.PublicIPv6Addresses) > 0 {
		return lb.PublicIPv6Addresses[0]
	}

	if len(lb.PrivateIPv4Addresses) > 0 {
		return lb.PrivateIPv4Addresses[0]
	}

	if len(lb.PrivateIPv6Addresses) > 0 {
		return lb.PrivateIPv6Addresses[0]
	}

	if lb.Domain != "" {
		return lb.Domain
	}

	return ""
}

// batchGetTargetCountByTargetGroupIDs 获取目标组中的RS数量
func (svc *lbSvc) batchGetTargetCountByTargetGroupIDs(kt *kit.Kit, targetGroupIDs []string) (map[string]int, error) {
	if len(targetGroupIDs) == 0 {
		return make(map[string]int), nil
	}

	uniqueTargetGroupIDs := removeDuplicates(targetGroupIDs)
	targets, err := svc.getTargetByTGIDs(kt, uniqueTargetGroupIDs)
	if err != nil {
		return nil, fmt.Errorf("batch query targets failed, err: %v", err)
	}

	targetCountMap := make(map[string]int)
	for _, target := range targets {
		if target.TargetGroupID != "" {
			targetCountMap[target.TargetGroupID]++
		}
	}

	return targetCountMap, nil
}

// removeDuplicates 去除字符串切片中的重复元素
func removeDuplicates(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
