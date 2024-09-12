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

	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchBizRuleOffline batch biz rule offline
func (svc *lbSvc) BatchBizRuleOffline(cts *rest.Contexts) (any, error) {
	return svc.batchRuleOffline(cts, handler.BizOperateAuth)
}

// BatchResRuleOffline batch rule offline
func (svc *lbSvc) BatchResRuleOffline(cts *rest.Contexts) (any, error) {
	return svc.batchRuleOffline(cts, handler.ResOperateAuth)
}

func (svc *lbSvc) batchRuleOffline(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch sops rule offline request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.UrlRuleAuditResType,
		Action: meta.Delete, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch sops rule offline auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get sops account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("get sops account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.parseAndBuildDeleteTCloudRule(cts.Kit, req.Data, accountInfo.AccountID, enumor.TCloud)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) parseAndBuildDeleteTCloudRule(kt *kit.Kit, body json.RawMessage,
	accountID string, vendor enumor.Vendor) (any, error) {
	req := new(cslb.TCloudSopsRuleBatchDeleteReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 得到 四层listenerID列表 和 七层urlRuleID列表
	listenerIDs, urlRuleIDs, err := svc.parseSOpsTargetParamsForRuleOffline(kt, accountID, vendor, req.RuleQueryList)
	if err != nil {
		logs.Errorf("parse sops target params for rule offline failed, err: %v", err)
		return nil, err
	}
	tcloudBatchDeleteRuleReq := &cslb.TcloudBatchDeleteRuleReq{
		ListenerIDs: listenerIDs,
		URLRuleIDs:  urlRuleIDs,
	}

	reqJson, _ := json.Marshal(tcloudBatchDeleteRuleReq)
	return svc.buildDeleteTCloudRule(kt, reqJson, vendor)
}

// parseSOpsTargetParamsForRuleOffline 解析标准运维参数-规则下线专属
func (svc *lbSvc) parseSOpsTargetParamsForRuleOffline(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	ruleQueryList []cslb.RuleQueryItemForRuleOffline) ([]string, []string, error) {

	listenerIDs := make([]string, 0)
	urlRuleIDs := make([]string, 0)
	index := 1
	for _, item := range ruleQueryList {
		// 1.VIP,VPort,Protocol确定监听器
		vipVportProtocolLblIDs, err := svc.parseSOpsVipAndVportAndProtocolToListenerIDs(kt, item.Region, accountID, vendor,
			item.Vip, item.VPort, item.Protocol)
		if err != nil {
			logs.Errorf("parse vip and vport and protocol to listener failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, nil, err
		}

		// 2.RsIP,RsType确定监听器
		rsIPRsTypeLblIDs, err := svc.parseSOpsRsIpAndRsTypeToListenerIDs(kt, accountID,
			item.RsIP, item.RsType)
		if err != nil {
			logs.Errorf("parse rs ip and rs type to listener failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, nil, err
		}

		// 3.取交集
		lblIDs := vipVportProtocolLblIDs
		if len(rsIPRsTypeLblIDs) != 0 {
			lblIDs = slice.Intersection(vipVportProtocolLblIDs, rsIPRsTypeLblIDs)
		}
		lblIDs = slice.Unique(lblIDs)
		if len(lblIDs) == 0 {
			// 当前行条件没有能匹配到的listener
			return nil, nil, fmt.Errorf("no matching listener were found for line: %d, param: %+v", index, item)
		}

		if item.Protocol[0].IsLayer4Protocol() {
			// 四层，确定Listener
			listenerIDs = append(listenerIDs, lblIDs...)
			index++
		} else if item.Protocol[0].IsLayer7Protocol() {
			// 七层，确定UrlRule
			// domain,url确定urlRule，lbl属于上面的监听器列表
			urIDs, err := svc.parseSOpsProtocolAndDomainAndUrlToUrlRuleIDs(kt, true,
				lblIDs, item.Domain, item.Url)
			if err != nil {
				logs.Errorf("parse domain and url to url rule failed, accountID: %s, item: %+v, err: %v, rid: %s",
					accountID, item, err, kt.Rid)
				return nil, nil, err
			}
			if len(urIDs) == 0 {
				// 当前行条件没有能匹配到的urlRule
				return nil, nil, fmt.Errorf("no matching url rule were found for line: %d, param: %+v", index, item)
			}
			urlRuleIDs = append(urlRuleIDs, urIDs...)
			index++
		} else {
			return nil, nil, fmt.Errorf("unsppurt protocol: %s", item.Protocol)
		}
	}

	return listenerIDs, urlRuleIDs, nil
}

// parseSOpsVipAndVportAndProtocolToListenerIDs 根据Vip、Vport、Protocol查询到对应负载均衡下的对应监听器
func (svc *lbSvc) parseSOpsVipAndVportAndProtocolToListenerIDs(kt *kit.Kit, region, accountID string, vendor enumor.Vendor,
	vip []string, vport []int, protocol []enumor.ProtocolType) ([]string, error) {

	if len(vip) == 0 && len(vport) == 0 && len(protocol) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}

	lblIDs := make([]string, 0)
	lbIDs := make([]string, 0)
	if len(vip) != 0 {
		// 若有vip筛选条件，则查询符合的负载均衡列表
		for _, vipItem := range vip {
			lbReq := &core.ListReq{
				Filter: tools.ExpressionAnd(
					tools.RuleEqual("vendor", vendor),
					tools.RuleEqual("account_id", accountID),
					tools.RuleEqual("region", region),
					tools.RuleJSONContains("public_ipv4_addresses", vipItem),
				),
				Page: core.NewDefaultBasePage(),
			}
			lbList, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
			if err != nil {
				logs.Errorf("list load balancer failed, req: %+v, err: %v", lbReq, err)
				return nil, err
			}

			for _, lbItem := range lbList.Details {
				lbIDs = append(lbIDs, lbItem.ID)
			}
		}
		// 若没有对应vip的负载均衡，则直接返回
		if len(lbIDs) == 0 {
			return lblIDs, nil
		}
	}

	// 查询符合的监听器列表
	lblFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.RuleEqual("vendor", vendor),
		},
	}
	if len(lbIDs) != 0 {
		lblFilter.Rules = append(lblFilter.Rules, tools.RuleIn("lb_id", lbIDs))
	}
	if len(vport) != 0 {
		lblFilter.Rules = append(lblFilter.Rules, tools.RuleIn("port", vport))
	}
	if len(protocol) != 0 {
		lblFilter.Rules = append(lblFilter.Rules, tools.RuleIn("protocol", protocol))
	}
	lblReq := &core.ListReq{
		Filter: lblFilter,
		Page:   core.NewDefaultBasePage(),
	}
	for {
		lblListResult, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
		if err != nil {
			logs.Errorf("list listener failed, err: %v", err)
			return nil, err
		}
		for _, detail := range lblListResult.Details {
			lblIDs = append(lblIDs, detail.ID)
		}

		if uint(len(lblListResult.Details)) < core.DefaultMaxPageLimit {
			break
		}

		lblReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return lblIDs, nil
}

// parseSOpsRsIpAndRsTypeToListenerIDs 根据RsIp和RsType查询Target，获取对应的目标组的Listener
func (svc *lbSvc) parseSOpsRsIpAndRsTypeToListenerIDs(kt *kit.Kit, accountID string,
	rsIPList []string, rsType string) ([]string, error) {

	if len(rsIPList) == 0 || len(rsType) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}

	tgIDs := make([]string, 0)
	for _, rsIP := range rsIPList {
		// 查询出对应的目标
		filter := tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("inst_type", rsType),
			tools.RuleJSONContains("private_ip_address", rsIP))
		targetReq := &core.ListReq{
			Fields: []string{"target_group_id"},
			Filter: filter,
			Page:   core.NewDefaultBasePage(),
		}
		targetResult, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
		if err != nil {
			return nil, err
		}

		// 记录目标对应的目标组ID
		for _, target := range targetResult.Details {
			if len(target.TargetGroupID) == 0 {
				continue
			}
			tgIDs = append(tgIDs, target.TargetGroupID)
		}
	}

	tgLblRelReq := &core.ListReq{
		Fields: []string{"lbl_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleIn("target_group_id", tgIDs)),
		Page: core.NewDefaultBasePage(),
	}
	tgLblRelResult, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, tgLblRelReq)
	if err != nil {
		return nil, err
	}
	lblIDs := make([]string, 0)
	for _, tgLblRel := range tgLblRelResult.Details {
		if len(tgLblRel.LblID) == 0 {
			continue
		}
		lblIDs = append(lblIDs, tgLblRel.LblID)
	}

	return slice.Unique(lblIDs), nil
}

// parseSOpsProtocolAndDomainAndUrlToUrlRuleIDs 根据RuleType、Domain、URL查询UrlRule
func (svc *lbSvc) parseSOpsProtocolAndDomainAndUrlToUrlRuleIDs(kt *kit.Kit,
	isLayer7RuleType bool, lblIDs, domain, url []string) ([]string, error) {
	// 筛选查询urlRule
	urlRuleFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.RuleIn("lbl_id", lblIDs),
		},
	}
	if isLayer7RuleType {
		urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleEqual("rule_type", enumor.Layer7RuleType))
		if len(domain) != 0 && !equalsAll(domain[0]) {
			urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleIn("domain", domain))
		}
		if len(url) != 0 && !equalsAll(url[0]) {
			urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleIn("url", url))
		}
	} else {
		urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleEqual("rule_type", enumor.Layer4RuleType))
	}

	urlRuleIDs := make([]string, 0)
	urlRuleReq := &core.ListReq{
		Filter: urlRuleFilter,
		Page:   core.NewDefaultBasePage(),
	}
	for {
		urlRuleResult, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleReq)
		if err != nil {
			logs.Errorf("list url rule failed, req: %+v, err: %v, rid: %s", urlRuleReq, err, kt.Rid)
			return nil, err
		}
		// 记录urlRuleID
		for _, ruleItem := range urlRuleResult.Details {
			urlRuleIDs = append(urlRuleIDs, ruleItem.ID)
		}

		if uint(len(urlRuleResult.Details)) < core.DefaultMaxPageLimit {
			break
		}

		urlRuleReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return slice.Unique(urlRuleIDs), nil
}

func equalsAll(str string) bool {
	return str == enumor.ParameterWildcard
}
