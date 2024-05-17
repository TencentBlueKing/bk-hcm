package loadbalancer

import (
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
	"hcm/pkg/runtime/filter"
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
		logs.Errorf("fail to merge loadbalancer id rule into request filter, err: %v, req.Filter: %+v, rid: %s",
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

	resList := &cslb.ListListenerResult{Count: 0, Details: make([]cslb.ListListenerBase, 0)}
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

	switch basicInfo.Vendor {
	case enumor.TCloud:
		urlRuleMap, targetWeightMap, lblTargetGroupMap, err := svc.getTCloudUrlRuleAndTargetGroupMap(
			cts.Kit, expr, req, lbID, resList)
		if err != nil {
			return nil, err
		}

		for idx, lblItem := range resList.Details {
			tmpTargetGroupID := lblTargetGroupMap[lblItem.ID].TargetGroupID
			resList.Details[idx].TargetGroupID = tmpTargetGroupID
			resList.Details[idx].Scheduler = urlRuleMap[lblItem.ID].Scheduler
			if lblItem.Protocol.IsLayer7Protocol() {
				resList.Details[idx].DomainNum = urlRuleMap[lblItem.ID].DomainNum
				resList.Details[idx].UrlNum = urlRuleMap[lblItem.ID].UrlNum
			}
			if len(tmpTargetGroupID) > 0 {
				resList.Details[idx].RsWeightNonZeroNum = targetWeightMap[tmpTargetGroupID].RsWeightNonZeroNum
				resList.Details[idx].RsWeightZeroNum = targetWeightMap[tmpTargetGroupID].RsWeightZeroNum
			}
			resList.Details[idx].BindingStatus = lblTargetGroupMap[lblItem.ID].BindingStatus
		}

		return resList, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "lbID: %s vendor: %s not support", lbID, basicInfo.Vendor)
	}
}

func (svc *lbSvc) getTCloudUrlRuleAndTargetGroupMap(kt *kit.Kit, expr *filter.Expression, req *core.ListReq,
	lbID string, resList *cslb.ListListenerResult) (map[string]cslb.ListListenerBase,
	map[string]cslb.ListListenerBase, map[string]cslb.ListListenerBase, error) {

	listenerReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	listenerList, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, listenerReq)
	if err != nil {
		logs.Errorf("list listener failed, lbID: %s, err: %v, rid: %s", lbID, err, kt.Rid)
		return nil, nil, nil, err
	}
	if req.Page.Count || len(listenerList.Details) == 0 {
		resList.Count = listenerList.Count
		return nil, nil, nil, nil
	}

	resList.Count = listenerList.Count
	lblIDs := make([]string, 0)
	for _, listenerItem := range listenerList.Details {
		lblIDs = append(lblIDs, listenerItem.ID)
		resList.Details = append(resList.Details, cslb.ListListenerBase{
			BaseListener: listenerItem,
		})
	}

	urlRuleMap, err := svc.listTCloudLbUrlRuleMap(kt, lbID, lblIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	// 根据lbID、lblID获取绑定的目标组ID列表
	ruleRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleIn("lbl_id", lblIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	ruleRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
	if err != nil {
		logs.Errorf("list target group listener rule rel failed, lbID: %s, lblIDs: %v, err: %v, rid: %s",
			lbID, lblIDs, err, kt.Rid)
		return nil, nil, nil, err
	}
	// 没有对应的目标组、监听器关联关系记录
	if len(ruleRelList.Details) == 0 {
		return nil, nil, nil, nil
	}

	targetGroupIDs := make([]string, 0)
	lblTargetGroupMap := make(map[string]cslb.ListListenerBase, 0)
	for _, item := range ruleRelList.Details {
		targetGroupIDs = append(targetGroupIDs, item.TargetGroupID)
		lblTargetGroupMap[item.LblID] = cslb.ListListenerBase{
			TargetGroupID: item.TargetGroupID,
			BindingStatus: item.BindingStatus,
		}
	}

	// TODO 后面拆成独立接口，让前端异步调用
	targetWeightMap, err := svc.listTargetWeightNumMap(kt, targetGroupIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return urlRuleMap, targetWeightMap, lblTargetGroupMap, nil
}

func (svc *lbSvc) listTCloudLbUrlRuleMap(kt *kit.Kit, lbID string, lblIDs []string) (
	map[string]cslb.ListListenerBase, error) {

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

	listenerRuleMap := make(map[string]cslb.ListListenerBase, 0)
	domainsExist := make(map[string]struct{}, 0)
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := listenerRuleMap[ruleItem.LblID]; !ok {
			listenerRuleMap[ruleItem.LblID] = cslb.ListListenerBase{
				TargetGroupID: ruleItem.TargetGroupID,
				Scheduler:     ruleItem.Scheduler,
				SessionType:   ruleItem.SessionType,
				SessionExpire: ruleItem.SessionExpire,
				HealthCheck:   ruleItem.HealthCheck,
				Certificate:   ruleItem.Certificate,
			}
		}

		tmpListener := listenerRuleMap[ruleItem.LblID]
		// 计算监听器下的域名数量
		calcDomainNumByListener(&tmpListener, ruleItem, domainsExist)
		if len(ruleItem.URL) > 0 {
			tmpListener.UrlNum++
		}

		listenerRuleMap[ruleItem.LblID] = tmpListener
	}

	return listenerRuleMap, nil
}

// 计算监听器下的域名数量
func calcDomainNumByListener(tmpListener *cslb.ListListenerBase, ruleItem corelb.TCloudLbUrlRule,
	domainsExist map[string]struct{}) {

	domainUnique := fmt.Sprintf("%s-%s", ruleItem.LblID, ruleItem.Domain)
	if _, ok := domainsExist[domainUnique]; !ok && len(ruleItem.Domain) > 0 {
		tmpListener.DomainNum++
		domainsExist[domainUnique] = struct{}{}
	}
	return
}

func (svc *lbSvc) listListenerMap(kt *kit.Kit, lblIDs []string) (map[string]corelb.BaseListener, error) {
	if len(lblIDs) == 0 {
		return nil, nil
	}

	lblReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", lblIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lblList, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		logs.Errorf("[clb] list clb listener failed, lblIDs: %v, err: %v, rid: %s", lblIDs, err, kt.Rid)
		return nil, err
	}

	lblMap := make(map[string]corelb.BaseListener, len(lblList.Details))
	for _, clbItem := range lblList.Details {
		lblMap[clbItem.ID] = clbItem
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

	urlRuleMap, err := svc.listTCloudLbUrlRuleMap(kt, listenerInfo.LbID, []string{lblID})
	if err != nil {
		return nil, err
	}

	targetGroupID := urlRuleMap[listenerInfo.ID].TargetGroupID
	result := &cslb.GetTCloudListenerDetail{
		TCloudListener: *listenerInfo,
		LblID:          listenerInfo.ID,
		LblName:        listenerInfo.Name,
		CloudLblID:     listenerInfo.CloudID,
		TargetGroupID:  targetGroupID,
		Scheduler:      urlRuleMap[listenerInfo.ID].Scheduler,
		SessionType:    urlRuleMap[listenerInfo.ID].SessionType,
		SessionExpire:  urlRuleMap[listenerInfo.ID].SessionExpire,
		HealthCheck:    urlRuleMap[listenerInfo.ID].HealthCheck,
	}
	if listenerInfo.Protocol.IsLayer7Protocol() {
		result.DomainNum = urlRuleMap[listenerInfo.ID].DomainNum
		result.UrlNum = urlRuleMap[listenerInfo.ID].UrlNum
		// 只有SNI开启时，证书才会出现在域名上面，才需要返回Certificate字段
		if listenerInfo.SniSwitch == enumor.SniTypeOpen {
			result.Certificate = urlRuleMap[listenerInfo.ID].Certificate
			result.Extension.Certificate = nil
		}
	}

	// 只有4层监听器才显示目标组信息
	if !listenerInfo.Protocol.IsLayer7Protocol() {
		tg, err := svc.getTargetGroupByID(kt, targetGroupID)
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

func (svc *lbSvc) listTargetWeightNumMap(kt *kit.Kit, targetGroupIDs []string) (
	map[string]cslb.ListListenerBase, error) {

	targetList, err := svc.getTargetByTGIDs(kt, targetGroupIDs)
	if err != nil {
		return nil, err
	}

	targetWeightMap := make(map[string]cslb.ListListenerBase, 0)
	for _, item := range targetList {
		tmpTarget := targetWeightMap[item.TargetGroupID]
		if cvt.PtrToVal(item.Weight) == 0 {
			tmpTarget.RsWeightZeroNum++
		} else {
			tmpTarget.RsWeightNonZeroNum++
		}
		targetWeightMap[item.TargetGroupID] = tmpTarget
	}

	return targetWeightMap, nil
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
