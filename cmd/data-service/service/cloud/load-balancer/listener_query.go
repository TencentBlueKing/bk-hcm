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
	"fmt"
	"strings"

	loadbalancer "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
)

// GetListener ...
func (svc *lbSvc) GetListener(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, lblID: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, fmt.Errorf("get listener failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "listener is not found")
	}

	lblInfo := result.Details[0]
	switch lblInfo.Vendor {
	case enumor.TCloud:
		newLblInfo, err := convTableToListener[corelb.TCloudListenerExtension](&lblInfo)
		if err != nil {
			logs.Errorf("fail to conv listener with extension, lblID: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
			return nil, err
		}
		return newLblInfo, nil
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
}

// ListListener list listener.
func (svc *lbSvc) ListListener(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list listener failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ListenerListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseListener, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseListener(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.ListenerListResult{Details: details}, nil
}

// ListListenerExt list listener with extension.
func (svc *lbSvc) ListListenerExt(cts *rest.Contexts) (any, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list listener failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ListenerListResult{Count: result.Count}, nil
	}

	details := make([]corelb.Listener[corelb.TCloudListenerExtension], 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convTableToListener[corelb.TCloudListenerExtension](&one)
		if err != nil {
			logs.Errorf("fail to conv listener with extension, err: %v, rid: %s", err, cts.Kit.Rid)
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.TCloudListenerListResult{Details: details}, nil
}

func convTableToBaseListener(one *tablelb.LoadBalancerListenerTable) *corelb.BaseListener {
	return &corelb.BaseListener{
		ID:            one.ID,
		CloudID:       one.CloudID,
		Name:          one.Name,
		Vendor:        one.Vendor,
		AccountID:     one.AccountID,
		BkBizID:       one.BkBizID,
		LbID:          one.LBID,
		CloudLbID:     one.CloudLBID,
		Protocol:      one.Protocol,
		Port:          one.Port,
		DefaultDomain: one.DefaultDomain,
		Region:        one.Region,
		Zones:         one.Zones,
		Memo:          one.Memo,
		SniSwitch:     one.SniSwitch,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

func convTableToListener[T corelb.ListenerExtension](table *tablelb.LoadBalancerListenerTable) (
	*corelb.Listener[T], error) {
	base := convTableToBaseListener(table)
	extension := new(T)
	if table.Extension != "" {
		if err := json.UnmarshalFromString(string(table.Extension), extension); err != nil {
			return nil, fmt.Errorf("fail unmarshal listener extension, err: %v", err)
		}
	}
	return &corelb.Listener[T]{
		BaseListener: base,
		Extension:    extension,
	}, nil
}

// ListListenerWithTargets list listener with target.
func (svc *lbSvc) ListListenerWithTargets(cts *rest.Contexts) (any, error) {
	req := new(protocloud.ListListenerWithTargetsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listenerList := &protocloud.ListListenerWithTargetsResp{}
	for _, item := range req.ListenerQueryList {
		queryReq := protocloud.ListListenerQueryReq{
			BkBizID:           req.BkBizID,
			Vendor:            req.Vendor,
			AccountID:         req.AccountID,
			ListenerQueryItem: item,
		}
		lblRsIPList, err := svc.queryListenerWithTargets(cts.Kit, queryReq)
		if err != nil {
			return nil, err
		}
		listenerList.Details = append(listenerList.Details, lblRsIPList...)
	}
	return listenerList, nil
}

func (svc *lbSvc) queryListenerWithTargets(kt *kit.Kit, lblReq protocloud.ListListenerQueryReq) (
	[]*protocloud.ListBatchListenerResult, error) {

	// 查询符合条件的负载均衡列表
	cloudClbIDs, clbIDs, lbMap, err := svc.listLoadBalancerListCheckVip(kt, lblReq)
	if err != nil {
		return nil, err
	}
	// 未查询到符合条件的负载均衡列表
	if len(cloudClbIDs) == 0 {
		logs.Errorf("check list load balancer with targets empty, req: %+v, rid: %s", lblReq, kt.Rid)
		return nil, nil
	}

	// 查询符合条件的监听器列表
	lblMap, cloudLblIDs, _, err := svc.listBizListenerByLbIDs(kt, lblReq, cloudClbIDs)
	if err != nil {
		return nil, err
	}
	// 未查询到符合的监听器列表
	if len(cloudLblIDs) == 0 {
		logs.Errorf("list biz listener with targets empty, req: %+v, rid: %s", lblReq, kt.Rid)
		return nil, nil
	}

	// 获取监听器绑定的目标组ID列表
	cloudTargetGroupIDs, err := svc.listTargetGroupIDsByRelCond(kt, lblReq, cloudLblIDs, clbIDs)
	if err != nil {
		return nil, err
	}

	// 根据RSIP获取绑定的目标组ID列表
	targetGroupRsList, targetGroupIDs, err := svc.listListenerWithTarget(kt, lblReq, cloudTargetGroupIDs)
	if err != nil {
		return nil, err
	}
	// 未查询到符合的目标组列表
	if len(targetGroupIDs) == 0 {
		logs.Errorf("list load balancer target with targets empty, req: %+v, rid: %s", lblReq, kt.Rid)
		return nil, nil
	}

	// 根据负载均衡ID、监听器ID、目标组ID，获取监听器与目标组的绑定关系列表
	lblUrlRuleList := make([]protocloud.LoadBalancerUrlRuleResult, 0)
	switch lblReq.Vendor {
	case enumor.TCloud:
		lblUrlRuleList, err = svc.listTCloudLBUrlRuleByTgIDs(kt, lblReq.ListenerQueryItem, cloudClbIDs,
			cloudLblIDs, targetGroupIDs)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "batch query listener with targets failed, invalid vendor: %s",
			lblReq.Vendor)
	}
	if err != nil {
		return nil, err
	}
	// 未查询到符合的监听器与目标组绑定关系的列表
	if len(lblUrlRuleList) == 0 {
		logs.Errorf("[%s]list load balancer url rule empty, req: %+v, cloudClbIDs: %v, cloudLblIDs: %v, "+
			"targetGroupIDs: %v, rid: %s", lblReq.Vendor, lblReq, cloudClbIDs,
			cloudLblIDs, targetGroupIDs, kt.Rid)
		return nil, nil
	}

	return svc.convertListListenerWithTargets(lbMap, lblUrlRuleList, lblMap, targetGroupRsList)
}

func (svc *lbSvc) convertListListenerWithTargets(lbMap map[string]tablelb.LoadBalancerTable,
	lblUrlRuleList []protocloud.LoadBalancerUrlRuleResult, lblMap map[string]tablelb.LoadBalancerListenerTable,
	targetGroupRsList map[string][]protocloud.LoadBalancerTargetRsList) (
	[]*protocloud.ListBatchListenerResult, error) {

	lblResult := make([]*protocloud.ListBatchListenerResult, 0)
	lblRsMap := make(map[string]*protocloud.ListBatchListenerResult)
	lblExist := make(map[string]struct{})
	for _, item := range lblUrlRuleList {
		// 遍历UrlRule列表，如果有多个监听器需要根据目标组ID，汇总RS列表
		if _, ok := lblExist[item.CloudLblID]; ok {
			lblRsMap = svc.getRsListByTargetGroupIDs(item, targetGroupRsList, lblRsMap)
			continue
		}
		lblExist[item.CloudLblID] = struct{}{}
		// 检查监听器是否存在
		lblInfo, ok := lblMap[item.CloudLblID]
		if !ok {
			continue
		}
		// 检查负载均衡是否存在
		lbInfo, ok := lbMap[item.CloudClbID]
		if !ok {
			continue
		}
		// 获取VIP/域名
		vipDomain, err := svc.getClbVipDomain(lbInfo)
		if err != nil {
			return nil, err
		}
		lblRsMap[item.CloudLblID] = &protocloud.ListBatchListenerResult{
			ClbID:        lbInfo.ID,
			CloudClbID:   item.CloudClbID,
			ClbVipDomain: strings.Join(vipDomain, ","),
			BkBizID:      lblInfo.BkBizID,
			Region:       lbInfo.Region,
			Vendor:       lbInfo.Vendor,
			LblID:        lblInfo.ID,
			CloudLblID:   item.CloudLblID,
			Protocol:     lblInfo.Protocol,
			Port:         lblInfo.Port,
			RsList:       make([]*protocloud.LoadBalancerTargetRsList, 0),
		}
		lblRsMap = svc.getRsListByTargetGroupIDs(item, targetGroupRsList, lblRsMap)
	}

	for _, item := range lblRsMap {
		lblResult = append(lblResult, &protocloud.ListBatchListenerResult{
			ClbID:        item.ClbID,
			CloudClbID:   item.CloudClbID,
			ClbVipDomain: item.ClbVipDomain,
			BkBizID:      item.BkBizID,
			Region:       item.Region,
			Vendor:       item.Vendor,
			LblID:        item.LblID,
			CloudLblID:   item.CloudLblID,
			Protocol:     item.Protocol,
			Port:         item.Port,
			RsList:       item.RsList,
		})
	}

	return lblResult, nil
}

func (svc *lbSvc) getRsListByTargetGroupIDs(item protocloud.LoadBalancerUrlRuleResult,
	targetGroupRsList map[string][]protocloud.LoadBalancerTargetRsList,
	lblRsMap map[string]*protocloud.ListBatchListenerResult) map[string]*protocloud.ListBatchListenerResult {

	if len(item.TargetGroupIDs) == 0 {
		return nil
	}

	for _, targetGroupID := range item.TargetGroupIDs {
		for _, targetGroupItem := range targetGroupRsList[targetGroupID] {
			lblRsMap[item.CloudLblID].RsList = append(lblRsMap[item.CloudLblID].RsList,
				&protocloud.LoadBalancerTargetRsList{
					BaseTarget:  targetGroupItem.BaseTarget,
					RuleID:      item.TargetGrouRuleMap[targetGroupID].RuleID,
					CloudRuleID: item.TargetGrouRuleMap[targetGroupID].CloudRuleID,
					RuleType:    item.TargetGrouRuleMap[targetGroupID].RuleType,
					Domain:      item.TargetGrouRuleMap[targetGroupID].Domain,
					Url:         item.TargetGrouRuleMap[targetGroupID].Url,
				})
		}
	}
	return lblRsMap
}

// getClbVipDomain 获取负载均衡的VIP或域名
func (svc *lbSvc) getClbVipDomain(lbInfo tablelb.LoadBalancerTable) ([]string, error) {
	vipDomains := make([]string, 0)
	switch loadbalancer.TCloudLoadBalancerType(lbInfo.LBType) {
	case loadbalancer.InternalLoadBalancerType:
		if lbInfo.IPVersion == string(enumor.Ipv4) {
			vipDomains = append(vipDomains, lbInfo.PrivateIPv4Addresses...)
		} else {
			vipDomains = append(vipDomains, lbInfo.PrivateIPv6Addresses...)
		}
	case loadbalancer.OpenLoadBalancerType:
		if lbInfo.IPVersion == string(enumor.Ipv4) {
			vipDomains = append(vipDomains, lbInfo.PublicIPv4Addresses...)
		} else {
			vipDomains = append(vipDomains, lbInfo.PublicIPv6Addresses...)
		}
	default:
		return nil, fmt.Errorf("unsupported lb_type: %s(%s)", lbInfo.LBType, lbInfo.CloudID)
	}

	// 如果IP为空则获取负载均衡域名
	if len(vipDomains) == 0 && len(lbInfo.Domain) > 0 {
		vipDomains = append(vipDomains, lbInfo.Domain)
	}

	return vipDomains, nil
}

// ListBatchListeners list batch listener.
func (svc *lbSvc) ListBatchListeners(cts *rest.Contexts) (any, error) {
	req := new(protocloud.BatchDeleteListenerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listenerList := &protocloud.BatchListListenerResp{}
	for _, item := range req.ListenerQueryList {
		lblList, err := svc.batchQueryListeners(cts.Kit, req, item)
		if err != nil {
			return nil, err
		}
		listenerList.Details = append(listenerList.Details, lblList...)
	}
	return listenerList, nil
}

func (svc *lbSvc) batchQueryListeners(kt *kit.Kit, req *protocloud.BatchDeleteListenerReq,
	lblReq *protocloud.ListenerDeleteReq) ([]*corelb.BaseListener, error) {

	// 查询符合条件的负载均衡列表
	queryReq := protocloud.ListListenerQueryReq{
		BkBizID:   req.BkBizID,
		Vendor:    req.Vendor,
		AccountID: req.AccountID,
		ListenerQueryItem: protocloud.ListenerQueryItem{
			Region:        lblReq.Region,
			ClbVipDomains: lblReq.ClbVipDomains,
			CloudLbIDs:    lblReq.CloudLbIDs,
			Protocol:      lblReq.Protocol,
			Ports:         lblReq.Ports,
		},
	}
	cloudClbIDs, _, _, err := svc.listLoadBalancerListCheckVip(kt, queryReq)
	if err != nil {
		return nil, err
	}

	// 未查询到符合条件的负载均衡列表
	if len(cloudClbIDs) == 0 {
		logs.Errorf("check list load balancer empty, req: %+v, lblReq: %+v, rid: %s", cvt.PtrToVal(req), lblReq, kt.Rid)
		return nil, nil
	}

	// 查询符合条件的监听器列表
	_, _, lblList, err := svc.listBizListenerByLbIDs(kt, queryReq, cloudClbIDs)
	if err != nil {
		return nil, err
	}

	// 未查询到符合的监听器列表
	if len(lblList) == 0 {
		logs.Errorf("list biz listener empty, req: %+v, lblReq: %+v, rid: %s", cvt.PtrToVal(req), lblReq, kt.Rid)
		return nil, nil
	}

	return svc.convertBatchListListener(lblList)
}

func (svc *lbSvc) convertBatchListListener(lblList []tablelb.LoadBalancerListenerTable) (
	[]*corelb.BaseListener, error) {

	lblResult := make([]*corelb.BaseListener, 0)
	for _, item := range lblList {
		lblResult = append(lblResult, &corelb.BaseListener{
			ID:            item.ID,
			CloudID:       item.CloudID,
			Name:          item.Name,
			Vendor:        item.Vendor,
			AccountID:     item.AccountID,
			BkBizID:       item.BkBizID,
			LbID:          item.LBID,
			CloudLbID:     item.CloudLBID,
			Protocol:      item.Protocol,
			Port:          item.Port,
			DefaultDomain: item.DefaultDomain,
			Region:        item.Region,
			Zones:         item.Zones,
			SniSwitch:     item.SniSwitch,
		})
	}
	return lblResult, nil
}

func (svc *lbSvc) listTCloudLBUrlRuleByTgIDs(kt *kit.Kit,
	lblReq protocloud.ListenerQueryItem, cloudClbIDs, cloudLblIDs, targetGroupIDs []string) (
	[]protocloud.LoadBalancerUrlRuleResult, error) {

	lblTargetFilter := make([]*filter.AtomRule, 0)
	lblTargetFilter = append(lblTargetFilter, tools.RuleIn("cloud_lb_id", cloudClbIDs))
	lblTargetFilter = append(lblTargetFilter, tools.RuleIn("cloud_lbl_id", cloudLblIDs))
	if len(targetGroupIDs) > 0 {
		lblTargetFilter = append(lblTargetFilter, tools.RuleIn("target_group_id", targetGroupIDs))
	}
	if len(lblReq.RuleType) > 0 {
		lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("rule_type", lblReq.RuleType))
		if lblReq.RuleType == enumor.Layer7RuleType {
			if len(lblReq.Domain) > 0 {
				lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("domain", lblReq.Domain))
			}
			if len(lblReq.Url) > 0 {
				lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("url", lblReq.Url))
			}
		}
	}
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(lblTargetFilter...),
		Page:   core.NewDefaultBasePage(),
	}
	lblTargetList := make([]protocloud.LoadBalancerUrlRuleResult, 0)
	for {
		loopLblTargetList, err := svc.dao.LoadBalancerTCloudUrlRule().List(kt, opt)
		if err != nil {
			logs.Errorf("list load balancer tcloud url rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list load balancer tcloud url rule failed, err: %v", err)
		}

		for _, item := range loopLblTargetList.Details {
			urlRuleResult := protocloud.LoadBalancerUrlRuleResult{
				LbID:              item.LbID,
				CloudClbID:        item.CloudLbID,
				LblID:             item.LblID,
				CloudLblID:        item.CloudLBLID,
				TargetGrouRuleMap: make(map[string]protocloud.DomainUrlRuleInfo),
			}
			urlRuleResult.TargetGroupIDs = append(urlRuleResult.TargetGroupIDs, item.TargetGroupID)
			urlRuleResult.TargetGrouRuleMap[item.TargetGroupID] = protocloud.DomainUrlRuleInfo{
				RuleID:      item.ID,
				CloudRuleID: item.CloudID,
				RuleType:    item.RuleType,
				Domain:      item.Domain,
				Url:         item.URL,
			}
			lblTargetList = append(lblTargetList, urlRuleResult)
		}
		if uint(len(loopLblTargetList.Details)) < core.DefaultMaxPageLimit {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return lblTargetList, nil
}

// listListenerWithTarget 根据账号ID、RsIP查询绑定的目标组列表
func (svc *lbSvc) listListenerWithTarget(kt *kit.Kit, lblReq protocloud.ListListenerQueryReq,
	cloudTargetGroupIDs []string) (map[string][]protocloud.LoadBalancerTargetRsList, []string, error) {

	targetList, err := svc.listTargetByCond(kt, lblReq, cloudTargetGroupIDs)
	if err != nil {
		return nil, nil, err
	}

	// 如果传入了RSPORT，则进行校验
	var targetIPPortMap = make(map[string]struct{}, len(targetList))
	if len(lblReq.ListenerQueryItem.RsPorts) > 0 {
		for idx, ip := range lblReq.ListenerQueryItem.RsIPs {
			targetIPPortMap[fmt.Sprintf("%s_%s_%d", lblReq.ListenerQueryItem.InstType, ip,
				lblReq.ListenerQueryItem.RsPorts[idx])] = struct{}{}
		}
	}

	// 统计每个目标组有多少RS
	targetGroupRsList := make(map[string][]protocloud.LoadBalancerTargetRsList)
	targetGroupIDs := make([]string, 0)
	for _, item := range targetList {
		// 不符合的数据需要过滤掉
		if _, ok := targetIPPortMap[fmt.Sprintf("%s_%s_%d", item.InstType, item.IP, item.Port)]; !ok &&
			len(lblReq.ListenerQueryItem.RsIPs) > 0 && len(lblReq.ListenerQueryItem.RsPorts) > 0 {
			logs.Warnf("list load balancer target rsip[%s] port[%d] is not found, rid: %s", item.IP, item.Port, kt.Rid)
			continue
		}

		if _, ok := targetGroupRsList[item.TargetGroupID]; !ok {
			targetGroupRsList[item.TargetGroupID] = make([]protocloud.LoadBalancerTargetRsList, 0)
		}
		targetGroupIDs = append(targetGroupIDs, item.TargetGroupID)
		targetGroupRsList[item.TargetGroupID] = append(targetGroupRsList[item.TargetGroupID],
			protocloud.LoadBalancerTargetRsList{
				BaseTarget: item,
			})
	}
	return targetGroupRsList, slice.Unique(targetGroupIDs), nil
}

// listBizListenerByLbIDs 获取业务下指定账号、负载均衡ID列表下的监听器列表
func (svc *lbSvc) listBizListenerByLbIDs(kt *kit.Kit, lblReq protocloud.ListListenerQueryReq, cloudClbIDs []string) (
	map[string]tablelb.LoadBalancerListenerTable, []string, []tablelb.LoadBalancerListenerTable, error) {

	lblFilter := make([]*filter.AtomRule, 0)
	lblFilter = append(lblFilter, tools.RuleEqual("vendor", lblReq.Vendor))
	lblFilter = append(lblFilter, tools.RuleEqual("bk_biz_id", lblReq.BkBizID))
	lblFilter = append(lblFilter, tools.RuleEqual("account_id", lblReq.AccountID))
	lblFilter = append(lblFilter, tools.RuleIn("cloud_lb_id", cloudClbIDs))
	lblFilter = append(lblFilter, tools.RuleEqual("protocol", lblReq.ListenerQueryItem.Protocol))
	if len(lblReq.ListenerQueryItem.Ports) > 0 {
		lblFilter = append(lblFilter, tools.RuleIn("port", lblReq.ListenerQueryItem.Ports))
	}

	lblList := make([]tablelb.LoadBalancerListenerTable, 0)
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(lblFilter...),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		loopLblList, err := svc.dao.LoadBalancerListener().List(kt, opt)
		if err != nil {
			logs.Errorf("list biz listener by clbIDs failed, err: %v, req: %+v, rid: %s",
				err, lblReq, kt.Rid)
			return nil, nil, nil, fmt.Errorf("list biz listener by clbIDs failed, err: %v", err)
		}

		lblList = append(lblList, loopLblList.Details...)
		if uint(len(loopLblList.Details)) < core.DefaultMaxPageLimit {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	lblProtocolPortMap := make(map[string]tablelb.LoadBalancerListenerTable, len(lblList))
	lblMap := make(map[string]tablelb.LoadBalancerListenerTable, len(lblList))
	cloudLblIDs := make([]string, 0)
	for _, item := range lblList {
		cloudLblIDs = append(cloudLblIDs, item.CloudID)
		lblMap[item.CloudID] = item
		lblProtocolPortMap[fmt.Sprintf("%s_%d", item.Protocol, item.Port)] = item
	}

	// 如果传入了监听器端口，则需要进行校验
	if len(lblReq.ListenerQueryItem.Ports) > 0 {
		for _, port := range lblReq.ListenerQueryItem.Ports {
			if _, ok := lblProtocolPortMap[fmt.Sprintf("%s_%d", lblReq.ListenerQueryItem.Protocol, port)]; !ok {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "listener protocol[%s] port[%d] is not found",
					lblReq.ListenerQueryItem.Protocol, port)
			}
		}
	}

	return lblMap, cloudLblIDs, lblList, nil
}

// ListListenerByCond list listener by cond.
func (svc *lbSvc) ListListenerByCond(cts *rest.Contexts) (any, error) {
	req := new(protocloud.ListListenerByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var err error
	listenerList := &protocloud.ListListenerByCondResp{}
	for _, item := range req.ListenerQueryList {
		// 负载均衡类型
		ruleType := enumor.Layer4RuleType
		if item.Protocol.IsLayer7Protocol() {
			ruleType = enumor.Layer7RuleType
		}

		queryReq := protocloud.ListListenerQueryReq{
			BkBizID:   req.BkBizID,
			Vendor:    req.Vendor,
			AccountID: req.AccountID,
			ListenerQueryItem: protocloud.ListenerQueryItem{
				Protocol:      item.Protocol,
				Region:        item.Region,
				CloudLbIDs:    item.CloudLbIDs,
				ClbVipDomains: item.ClbVipDomains,
				RuleType:      ruleType,
				RsIPs:         item.RsIPs,
				RsPorts:       item.RsPorts,
			},
		}

		var lblCondList []*protocloud.ListBatchListenerResult
		// 如果传入了RSIP、RSPort，需要查询监听器对应的目标组、目标组里的RS是否匹配
		if len(item.RsIPs) > 0 || len(item.RsPorts) > 0 {
			lblCondList, err = svc.queryListenerWithTargets(cts.Kit, queryReq)
		} else {
			lblCondList, err = svc.queryListenerWithoutTargets(cts.Kit, queryReq)
		}
		if err != nil {
			return nil, err
		}
		listenerList.Details = append(listenerList.Details, lblCondList...)
	}
	return listenerList, nil
}

func (svc *lbSvc) queryListenerWithoutTargets(kt *kit.Kit, lblQueryReq protocloud.ListListenerQueryReq) (
	[]*protocloud.ListBatchListenerResult, error) {

	// 查询符合条件的负载均衡列表
	cloudClbIDs, _, lbMap, err := svc.listLoadBalancerListCheckVip(kt, lblQueryReq)
	if err != nil {
		return nil, err
	}

	// 未查询到符合条件的负载均衡列表
	if len(cloudClbIDs) == 0 {
		logs.Errorf("check list load balancer by cond empty, req: %+v, rid: %s", lblQueryReq, kt.Rid)
		return nil, nil
	}

	// 查询符合条件的监听器列表
	_, cloudLblIDs, lblList, err := svc.listBizListenerByLbIDs(kt, lblQueryReq, cloudClbIDs)
	if err != nil {
		return nil, err
	}

	// 未查询到符合的监听器列表
	if len(cloudLblIDs) == 0 {
		logs.Errorf("list biz listener by cond empty, req: %+v, rid: %s", lblQueryReq, kt.Rid)
		return nil, nil
	}

	lblResult := make([]*protocloud.ListBatchListenerResult, 0)
	for _, item := range lblList {
		// 检查负载均衡是否存在
		lbInfo, ok := lbMap[item.CloudLBID]
		if !ok {
			continue
		}

		// 获取VIP/域名
		vipDomain, err := svc.getClbVipDomain(lbInfo)
		if err != nil {
			return nil, err
		}

		lblResult = append(lblResult, &protocloud.ListBatchListenerResult{
			ClbID:        lbInfo.ID,
			CloudClbID:   lbInfo.CloudID,
			ClbVipDomain: strings.Join(vipDomain, ","),
			BkBizID:      item.BkBizID,
			Region:       item.Region,
			Vendor:       item.Vendor,
			LblID:        item.ID,
			CloudLblID:   item.CloudID,
			Protocol:     item.Protocol,
			Port:         item.Port,
		})
	}

	return lblResult, nil
}
