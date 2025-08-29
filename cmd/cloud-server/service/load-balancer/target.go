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
	"strings"

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/cidr"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// ListBizTargetByCond list biz target by cond.
func (svc *lbSvc) ListBizTargetByCond(cts *rest.Contexts) (any, error) {
	return svc.listTargetByCond(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listTargetByCond(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (any, error) {
	req := new(cslb.ListTargetByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		logs.Errorf("list listener by cond auth failed, noPermFlag: %v, err: %v, rid: %s",
			noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		logs.Errorf("list listener no perm auth, noPermFlag: %v, req: %+v, rid: %s", noPermFlag, req, cts.Kit.Rid)
		return nil, errf.New(errf.PermissionDenied, "permission denied for get listener by cond")
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, fmt.Errorf("get account basic info failed, err: %v", err)
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resList := &cslb.ListTargetByCondResp{Details: make([]*cslb.ListTargetByCondResult, 0)}
	for _, item := range req.RuleQueryList {
		result, err := svc.listTargetByCondWithQueryLine(cts.Kit, accountInfo.Vendor, bkBizID, item, req.AccountID)
		if err != nil {
			return nil, err
		}
		resList.Details = append(resList.Details, result...)
	}
	return resList, nil

}

func (svc *lbSvc) listTargetByCondWithQueryLine(kt *kit.Kit, vendor enumor.Vendor, bkBizID int64,
	req cslb.TargetQueryLine, accountID string) ([]*cslb.ListTargetByCondResult, error) {

	cloudLBIDs, cloudIDToLB, err := svc.listLoadBalancerAndValidateIP(kt, vendor, bkBizID, req)
	if err != nil {
		return nil, err
	}

	listeners, err := svc.getListenerWithCond(kt, req, cloudLBIDs, vendor, bkBizID, accountID)
	if err != nil {
		return nil, err
	}
	if len(listeners) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no listener found"))
	}
	lblIDs := slice.Map(listeners, func(item loadbalancer.BaseListener) string {
		return item.ID
	})
	lblMap := cvt.SliceToMap(listeners, func(item loadbalancer.BaseListener) (string, loadbalancer.BaseListener) {
		return item.ID, item
	})

	// 查询URL rule
	ruleIDs, ruleMap, err := svc.listUrlRuleByCond(kt, vendor, req, lblIDs)
	if err != nil {
		return nil, err
	}
	// 根据URL rule 查询对应的目标组，最后根据目标组 加上rs的条件，查询rs
	tgIDs, tgToRuleID, err := svc.listTargetGroupByRuleIDs(kt, ruleIDs)
	if err != nil {
		return nil, err
	}
	targets, err := svc.listTargetByTgIDsAndCond(kt, tgIDs, req)
	if err != nil {
		return nil, err
	}

	result := make([]*cslb.ListTargetByCondResult, 0)
	for _, target := range targets {
		item := &cslb.ListTargetByCondResult{
			BkBizId: bkBizID,
			Region:  req.Region,
			Vendor:  vendor,

			InstType: string(target.InstType),
			RsIp:     target.IP,
			RsPort:   target.Port,
			RsWeight: cvt.PtrToVal(target.Weight),
		}
		ruleID := tgToRuleID[target.TargetGroupID]
		ruleInfo := ruleMap[ruleID]
		item.Domain = ruleInfo.domain
		item.Url = ruleInfo.url

		clb := cloudIDToLB[ruleInfo.cloudLBID]
		item.ClbId = clb.ID
		item.CloudLbId = clb.CloudID
		vipDomain, err := getClbVipDomain(clb)
		if err != nil {
			return nil, err
		}
		item.ClbVipDomain = strings.Join(vipDomain, ",")
		item.LblId = ruleInfo.lblID
		item.CloudLblId = ruleInfo.cloudLblID
		lbl := lblMap[ruleInfo.lblID]
		item.Protocol = lbl.Protocol
		item.Port = lbl.Port
		result = append(result, item)
	}
	return result, nil
}

func (svc *lbSvc) listTargetByTgIDsAndCond(kt *kit.Kit, tgIDs []string, req cslb.TargetQueryLine) (
	[]loadbalancer.BaseTarget, error) {

	result := make([]loadbalancer.BaseTarget, 0)
	cond := &filter.Expression{
		Op:    filter.And,
		Rules: make([]filter.RuleFactory, 0),
	}
	if len(req.RsIps) > 0 {
		cond.Rules = append(cond.Rules, tools.RuleIn("ip", req.RsIps))
	}
	if len(req.RsPorts) > 0 {
		cond.Rules = append(cond.Rules, tools.RuleIn("port", req.RsPorts))
	}
	for _, batch := range slice.Split(tgIDs, int(core.DefaultMaxPageLimit)) {
		finalCond, err := tools.And(
			tools.RuleIn("target_group_id", batch),
			cond,
		)
		if err != nil {
			return nil, err
		}

		listReq := &core.ListReq{
			Filter: finalCond,
			Page:   core.NewDefaultBasePage(),
		}
		for {
			resp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listReq)
			if err != nil {
				logs.Errorf("list target by tgIDs and cond failed, err: %v, req: %+v, rid: %s", err, listReq, kt.Rid)
				return nil, err
			}
			result = append(result, resp.Details...)
			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	return result, nil
}

func (svc *lbSvc) listTargetGroupByRuleIDs(kt *kit.Kit, ruleIDs []string) ([]string, map[string]string, error) {

	tgIDs := make([]string, 0)
	tgToRuleIDs := make(map[string]string)
	for _, batch := range slice.Split(ruleIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Fields: []string{"target_group_id", "listener_rule_id"},
			Filter: tools.ExpressionAnd(
				tools.RuleIn("listener_rule_id", batch),
			),
			Page: core.NewDefaultBasePage(),
		}
		for {
			resp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, listReq)
			if err != nil {
				return nil, nil, err
			}
			for _, rel := range resp.Details {
				tgIDs = append(tgIDs, rel.TargetGroupID)
				tgToRuleIDs[rel.TargetGroupID] = rel.ListenerRuleID
			}
			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}
	return tgIDs, tgToRuleIDs, nil
}

func (svc *lbSvc) listUrlRuleByCond(kt *kit.Kit, vendor enumor.Vendor, req cslb.TargetQueryLine, lblIDs []string) (
	[]string, map[string]urlRuleInfo, error) {

	switch vendor {
	case enumor.TCloud:
		return svc.listUrlRuleByCondForTCloud(kt, req, lblIDs)
	default:
		return nil, nil, errf.NewFromErr(errf.InvalidParameter,
			fmt.Errorf("unsupported vendor: %s for listUrlRuleByCond", vendor))
	}
}

type urlRuleInfo struct {
	domain     string
	url        string
	lblID      string
	cloudLblID string
	cloudLBID  string
}

func (svc *lbSvc) listUrlRuleByCondForTCloud(kt *kit.Kit, req cslb.TargetQueryLine, lblIDs []string) ([]string,
	map[string]urlRuleInfo, error) {
	ruleType := enumor.Layer4RuleType
	if req.Protocol.IsLayer7Protocol() {
		ruleType = enumor.Layer7RuleType
	}
	queryRules := []*filter.AtomRule{
		tools.RuleEqual("rule_type", ruleType),
	}
	if len(req.Domains) > 0 {
		queryRules = append(queryRules, tools.RuleIn("domain", req.Domains))
	}
	if len(req.Urls) > 0 {
		queryRules = append(queryRules, tools.RuleIn("url", req.Urls))
	}
	result := make([]string, 0)
	ruleMap := make(map[string]urlRuleInfo)
	for _, batch := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				append(queryRules,
					tools.RuleIn("lbl_id", batch),
				)...,
			),
			Page: core.NewDefaultBasePage(),
		}
		for {
			resp, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, listReq)
			if err != nil {
				return nil, nil, err
			}
			for _, detail := range resp.Details {
				result = append(result, detail.ID)
				ruleMap[detail.ID] = urlRuleInfo{
					domain:     detail.Domain,
					url:        detail.URL,
					lblID:      detail.LblID,
					cloudLblID: detail.CloudLBLID,
					cloudLBID:  detail.CloudLbID,
				}
			}
			if len(resp.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	return result, ruleMap, nil
}

func (svc *lbSvc) listLoadBalancerAndValidateIP(kt *kit.Kit, vendor enumor.Vendor, bkBizID int64,
	req cslb.TargetQueryLine) ([]string, map[string]loadbalancer.BaseLoadBalancer, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleIn("cloud_id", req.CloudLbIds),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, listReq)
	if err != nil {
		logs.Errorf("list load balancer failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, nil, err
	}
	cloudLBIDs, _, lbMap, err := checkClbVipAndDomain(resp.Details, req.CloudLbIds, req.ClbVipDomains)
	if err != nil {
		logs.Errorf("check clb vip and domain failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, nil, err
	}
	return cloudLBIDs, lbMap, nil
}

func (svc *lbSvc) getListenerWithCond(kt *kit.Kit, req cslb.TargetQueryLine, cloudLBIDs []string,
	vendor enumor.Vendor, bkBizID int64, accountID string) ([]loadbalancer.BaseListener, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("protocol", req.Protocol),
			tools.RuleEqual("region", req.Region),
			tools.RuleIn("cloud_lb_id", cloudLBIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	if len(req.ListenerPorts) > 0 {
		listReq.Filter.Rules = append(listReq.Filter.Rules, tools.RuleIn("port", req.ListenerPorts))
	}
	result := make([]loadbalancer.BaseListener, 0)
	for {
		lblResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, listReq)
		if err != nil {
			return nil, err
		}
		result = append(result, lblResp.Details...)
		if uint(len(lblResp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return result, nil
}

func checkClbVipAndDomain(list []loadbalancer.BaseLoadBalancer, paramClbIDs, clbVipDomains []string) (
	[]string, []string, map[string]loadbalancer.BaseLoadBalancer, error) {

	lbMap := cvt.SliceToMap(list, func(item loadbalancer.BaseLoadBalancer) (string, loadbalancer.BaseLoadBalancer) {
		return item.CloudID, item
	})

	cloudClbIDs := make([]string, 0)
	clbIDs := make([]string, 0)
	for idx, cloudID := range paramClbIDs {
		lbInfo, ok := lbMap[cloudID]
		if !ok {
			return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] is not found", cloudID)
		}

		// 检查对应的负载均衡VIP/域名是否匹配
		vipDomain := clbVipDomains[idx]
		if cidr.IsDomainName(vipDomain) && lbInfo.Domain != vipDomain {
			return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] domain is not match, "+
				"paramDomain: %s, clbDomain: %s", cloudID, vipDomain, lbInfo.Domain)
		}

		switch lbInfo.LoadBalancerType {
		case string(typeslb.InternalLoadBalancerType): // 内网
			if cidr.IsIPv4(vipDomain) && !slice.IsItemInSlice(lbInfo.PrivateIPv4Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] privateIPv4 is not match, "+
					"paramIPv4: %s, clbPrivateIPv4: %v", cloudID, vipDomain, lbInfo.PrivateIPv4Addresses)
			}
			if cidr.IsIPv6(vipDomain) && !slice.IsItemInSlice(lbInfo.PrivateIPv6Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] privateIPv6 is not match, "+
					"paramIPv6: %s, clbPrivateIPv6: %v", cloudID, vipDomain, lbInfo.PrivateIPv6Addresses)
			}
		case string(typeslb.OpenLoadBalancerType): // 公网
			if cidr.IsIPv4(vipDomain) && !slice.IsItemInSlice(lbInfo.PublicIPv4Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] publicIPv4 is not match, "+
					"paramIPv4: %s, clbPublicIPv4: %v", cloudID, vipDomain, lbInfo.PublicIPv4Addresses)
			}
			if cidr.IsIPv6(vipDomain) && !slice.IsItemInSlice(lbInfo.PublicIPv6Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] publicIPv6 is not match, "+
					"paramIPv6: %s, clbPublicIPv6: %v", cloudID, vipDomain, lbInfo.PublicIPv6Addresses)
			}
		default:
			return nil, nil, nil, errf.Newf(errf.InvalidParameter, "unsupported hcm lb type: %s", lbInfo.LoadBalancerType)
		}
		cloudClbIDs = append(cloudClbIDs, cloudID)
		clbIDs = append(clbIDs, lbInfo.ID)
	}

	return slice.Unique(cloudClbIDs), slice.Unique(clbIDs), lbMap, nil
}

func getClbVipDomain(lbInfo loadbalancer.BaseLoadBalancer) ([]string, error) {
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
