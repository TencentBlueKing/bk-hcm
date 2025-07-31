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
	"hcm/pkg/tools/classifier"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListListener list listener.
func (svc *lbSvc) ListListener(cts *rest.Contexts) (interface{}, error) {
	return svc.listListener(cts, handler.ListResourceAuthRes)
}

// ListBizListener list biz listener.
func (svc *lbSvc) ListBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.listListener(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listListener(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lb_id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	filterWithLb, err := tools.And(tools.RuleEqual("lb_id", lbID), req.Filter)
	if err != nil {
		logs.Errorf("fail to merge load balancer id rule into request filter, err: %v, req.Filter: %+v, rid: %s",
			err, req.Filter, cts.Kit.Rid)
		return nil, err
	}
	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find, Filter: filterWithLb})
	if err != nil {
		logs.Errorf("list listener auth failed, lbID: %s, noPermFlag: %v, err: %v, rid: %s",
			lbID, noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}
	resList := &cslb.ListListenerResult{Count: 0, Details: make([]*cslb.ListenerListInfo, 0)}
	if noPermFlag {
		logs.Errorf("list listener no perm auth, lbID: %s, noPermFlag: %v, rid: %s", lbID, noPermFlag, cts.Kit.Rid)
		return resList, nil
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.LoadBalancerCloudResType, lbID)
	if err != nil {
		logs.Errorf("fail to get load balancer basic info, lbID: %s, err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}
	req.Filter = expr
	if req.Page.Count {
		return svc.client.DataService().Global.LoadBalancer.ListListener(cts.Kit, req)
	}
	switch basicInfo.Vendor {
	case enumor.TCloud:
		lblInfoList, err := svc.getTCloudUrlRuleAndTargetGroupMap(cts.Kit, lbID, req)
		return &cslb.ListListenerResult{Details: lblInfoList}, err
	default:
		return nil, errf.Newf(errf.InvalidParameter, "lbID: %s vendor: %s not support", lbID, basicInfo.Vendor)
	}
}

// 返回监听器信息， 域名数量和url数量，绑定目标组同步状态
func (svc *lbSvc) getTCloudUrlRuleAndTargetGroupMap(kt *kit.Kit, lbID string,
	req *core.ListReq) ([]*cslb.ListenerListInfo, error) {

	listenerList, err := svc.client.DataService().TCloud.LoadBalancer.ListListener(kt, req)
	if err != nil {
		logs.Errorf("list listener failed, lbID: %s, err: %v, rid: %s", lbID, err, kt.Rid)
		return nil, err
	}

	if len(listenerList.Details) == 0 {
		return nil, nil
	}

	baseLblList := listenerList.Details
	lblInfoList := make([]*cslb.ListenerListInfo, 0, len(baseLblList))
	lblIDs := make([]string, 0)
	for _, lbl := range baseLblList {
		lblIDs = append(lblIDs, lbl.ID)
		lblInfoList = append(lblInfoList, &cslb.ListenerListInfo{
			BaseListener: *lbl.BaseListener,
			EndPort:      lbl.Extension.EndPort,
		})
	}

	// 2. 拼接规则信息到监听器表中，如果是4层监听器，拼接均衡方式和同步状态，7层监听器拼接域名数量和url数量
	lblRuleMap, err := svc.listTCloudRuleMap(kt, lbID, lblIDs)
	if err != nil {
		logs.Errorf("fail to list tcloud rule map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 3. 根据lbID、lblID获取绑定的目标组ID列表
	relMap, err := svc.listTgLblRelMap(kt, lbID, lblIDs)
	if err != nil {
		logs.Errorf("fail to list target group  listener rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, lblInfo := range lblInfoList {
		lblID := lblInfo.BaseListener.ID
		rules := lblRuleMap[lblID]
		if len(rules) == 0 {
			continue
		}
		if lblInfo.Protocol.IsLayer7Protocol() {
			// 7层监听器获取url规则数量和域名数量
			domains := cvt.SliceToMap(rules,
				func(r corelb.TCloudLbUrlRule) (string, struct{}) { return r.Domain, struct{}{} })
			lblInfo.DomainNum = int64(len(domains))
			lblInfo.UrlNum = int64(len(rules))
		} else {
			// 4层监听器
			lblInfo.Scheduler = rules[0].Scheduler
			lblInfo.SessionType = rules[0].SessionType
			lblInfo.SessionExpire = rules[0].SessionExpire
			lblInfo.HealthCheck = rules[0].HealthCheck
			lblInfo.Certificate = rules[0].Certificate
			// 获取同步状态和目标组id
			lblInfo.TargetGroupID = relMap[lblID].TargetGroupID
			lblInfo.BindingStatus = relMap[lblID].BindingStatus
		}
	}

	return lblInfoList, nil
}

func (svc *lbSvc) listTgLblRelMap(kt *kit.Kit, lbID string, lblIDs []string) (
	map[string]corelb.BaseTargetListenerRuleRel, error) {

	ruleRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleIn("lbl_id", lblIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	ruleRelResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
	if err != nil {
		logs.Errorf("list target group listener rule rel failed, lbID: %s, lblIDs: %v, err: %v, rid: %s",
			lbID, lblIDs, err, kt.Rid)
		return nil, err
	}
	relMap := make(map[string]corelb.BaseTargetListenerRuleRel)
	for _, rel := range ruleRelResp.Details {
		relMap[rel.LblID] = rel
	}
	return relMap, nil
}

func (svc *lbSvc) listTCloudRuleMap(kt *kit.Kit, lbID string, lblIDs []string) (
	map[string][]corelb.TCloudLbUrlRule, error) {

	urlRuleReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleIn("lbl_id", lblIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud url rule failed, lbID: %s, lblIDs: %v, err: %v, rid: %s", lbID, lblIDs, err, kt.Rid)
		return nil, err
	}
	lblRuleMap := classifier.ClassifySlice(urlRuleList.Details,
		func(r corelb.TCloudLbUrlRule) string { return r.LblID })
	return lblRuleMap, nil
}

func (svc *lbSvc) listListenerMap(kt *kit.Kit, lblIDs []string) (map[string]corelb.BaseListener, error) {
	if len(lblIDs) == 0 {
		return nil, nil
	}

	lblMap := make(map[string]corelb.BaseListener, 0)
	lblReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", lblIDs),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		listLblResult, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
		if err != nil {
			logs.Errorf("[clb] list clb listener failed, lblIDs: %v, err: %v, rid: %s", lblIDs, err, kt.Rid)
			return nil, err
		}
		for _, listenerItem := range listLblResult.Details {
			lblMap[listenerItem.ID] = listenerItem
		}

		if uint(len(listLblResult.Details)) < core.DefaultMaxPageLimit {
			break
		}

		lblReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return lblMap, nil
}

// GetListener get clb listener.
func (svc *lbSvc) GetListener(cts *rest.Contexts) (interface{}, error) {
	return svc.getListener(cts, handler.ListResourceAuthRes)
}

// GetBizListener get biz clb listener.
func (svc *lbSvc) GetBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.getListener(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) getListener(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, id)
	if err != nil {
		logs.Errorf("fail to get listener basic info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.Listener, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get listener")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.getTCloudListener(cts.Kit, id)

	default:
		return nil, errf.Newf(errf.InvalidParameter, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

func (svc *lbSvc) getTCloudListener(kt *kit.Kit, lblID string) (*cslb.GetTCloudListenerDetail, error) {
	listenerInfo, err := svc.client.DataService().TCloud.LoadBalancer.GetListener(kt, lblID)
	if err != nil {
		logs.Errorf("get tcloud listener detail failed, lblID: %s, err: %v, rid: %s", lblID, err, kt.Rid)
		return nil, err
	}

	urlRuleMap, err := svc.listTCloudRuleMap(kt, listenerInfo.LbID, []string{lblID})
	if err != nil {
		return nil, err
	}
	rules := urlRuleMap[listenerInfo.ID]
	var rule *corelb.TCloudLbUrlRule
	if len(rules) > 0 {
		rule = &rules[0]
	}

	result := &cslb.GetTCloudListenerDetail{
		TCloudListener: *listenerInfo,
		LblID:          listenerInfo.ID,
		LblName:        listenerInfo.Name,
		CloudLblID:     listenerInfo.CloudID,
		EndPort:        listenerInfo.Extension.EndPort,
	}
	if listenerInfo.Protocol.IsLayer7Protocol() {
		domains := cvt.SliceToMap(rules,
			func(r corelb.TCloudLbUrlRule) (string, struct{}) { return r.Domain, struct{}{} })
		result.DomainNum = int64(len(domains))
		result.UrlNum = int64(len(rules))
		// SNI关闭，证书在在监听器上，需要返回监听器自己的证书信息
		// SNI打开，证书在域名（规则）上
		if listenerInfo.SniSwitch == enumor.SniTypeClose {
			result.Certificate = result.Extension.Certificate
		}
	}

	// 只有4层监听器才显示目标组信息
	if !listenerInfo.Protocol.IsLayer7Protocol() {
		if rule == nil {
			logs.Errorf("fail to find related rule fo lbl(%s),rid: %s", lblID, kt.Rid)
			return nil, errors.New("related rule not found")
		}
		result.TargetGroupID = rule.TargetGroupID
		result.Scheduler = rule.Scheduler
		result.SessionType = rule.SessionType
		result.SessionExpire = rule.SessionExpire
		result.HealthCheck = rule.HealthCheck

		tg, err := svc.getTargetGroupByID(kt, rule.TargetGroupID)
		if err != nil {
			return nil, err
		}
		if tg != nil {
			result.TargetGroupName = tg.Name
			result.CloudTargetGroupID = tg.CloudID
		}
	}

	return result, nil
}

// ListListenerCountByLbIDs list listener count by lbIDs.
func (svc *lbSvc) ListListenerCountByLbIDs(cts *rest.Contexts) (interface{}, error) {
	return svc.listListenerCountByLbIDs(cts, handler.ListResourceAuthRes)
}

// ListBizListenerCountByLbIDs list biz listener count by lbIDs.
func (svc *lbSvc) ListBizListenerCountByLbIDs(cts *rest.Contexts) (interface{}, error) {
	return svc.listListenerCountByLbIDs(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listListenerCountByLbIDs(cts *rest.Contexts,
	authHandler handler.ListAuthResHandler) (interface{}, error) {

	req := new(dataproto.ListListenerCountByLbIDsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	filterLb, err := tools.And(tools.RuleIn("lb_id", req.LbIDs))
	if err != nil {
		logs.Errorf("fail to merge load balancer id into request filter, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	// list authorized instances
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find, Filter: filterLb})
	if err != nil {
		logs.Errorf("list listener by lbIDs auth failed, lbIDs: %v, noPermFlag: %v, err: %v, rid: %s",
			req.LbIDs, noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	resList := &dataproto.ListListenerCountResp{Details: make([]*dataproto.ListListenerCountResult, 0)}
	if noPermFlag {
		logs.Errorf("list listener no perm auth, lbIDs: %v, noPermFlag: %v, rid: %s",
			req.LbIDs, noPermFlag, cts.Kit.Rid)
		return resList, nil
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.LoadBalancerCloudResType, req.LbIDs[0])
	if err != nil {
		logs.Errorf("fail to get load balancer basic info, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		resList, err = svc.client.DataService().Global.LoadBalancer.CountLoadBalancerListener(cts.Kit, req)
		if err != nil {
			logs.Errorf("tcloud count load balancer listener failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}
		return resList, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "lbIDs: %v vendor: %s not support", req.LbIDs, basicInfo.Vendor)
	}
}

// ListBizListenerWithTargets list biz listener with targets.
func (svc *lbSvc) ListBizListenerWithTargets(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id: %d is invalid", bkBizID)
	}

	return svc.listListenerWithTarget(cts, handler.ListBizAuthRes, bkBizID)
}

func (svc *lbSvc) listListenerWithTarget(cts *rest.Contexts, authHandler handler.ListAuthResHandler,
	bkBizID int64) (any, error) {

	req := new(dataproto.ListListenerWithTargetsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		logs.Errorf("list listener with targets auth failed, noPermFlag: %v, err: %v, rid: %s",
			noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	resList := &dataproto.ListListenerWithTargetsResp{Details: make([]*dataproto.ListBatchListenerResult, 0)}
	if noPermFlag {
		logs.Errorf("list listener no perm auth, noPermFlag: %v, req: %+v, rid: %s", noPermFlag, req, cts.Kit.Rid)
		return resList, nil
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, fmt.Errorf("get account basic info failed, err: %v", err)
	}

	req.BkBizID = bkBizID
	switch accountInfo.Vendor {
	case enumor.TCloud:
		resList, err = svc.client.DataService().Global.LoadBalancer.ListLoadBalancerListenerWithTargets(cts.Kit, req)
		if err != nil {
			logs.Errorf("tcloud list listener with targets failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}
		return resList, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "list listener with targets failed, vendor: %s not support, "+
			"listTargetReq: %+v", accountInfo.Vendor, req)
	}
}

// ListBizListenerByCond list biz listener by cond.
func (svc *lbSvc) ListBizListenerByCond(cts *rest.Contexts) (any, error) {
	return svc.listListenerByCond(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listListenerByCond(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (any, error) {
	req := new(dataproto.ListListenerByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
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

	resList := &dataproto.ListListenerByCondResp{Details: make([]*dataproto.ListBatchListenerResult, 0)}
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
		return nil, err
	}

	req.BkBizID = bkBizID
	switch accountInfo.Vendor {
	case enumor.TCloud:
		resList, err = svc.client.DataService().Global.LoadBalancer.ListListenerByCond(cts.Kit, req)
		if err != nil {
			logs.Errorf("tcloud list listener by cond failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}
		return resList, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "list listener by cond failed, vendor: %s not support, req: %+v",
			accountInfo.Vendor, req)
	}
}
