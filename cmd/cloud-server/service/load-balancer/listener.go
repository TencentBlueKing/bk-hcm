package loadbalancer

import (
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

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find, Filter: req.Filter})
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

	listenerReq := &dataproto.ListListenerReq{
		LbID: lbID,
		ListReq: core.ListReq{
			Filter: expr,
			Page:   req.Page,
		},
	}
	listenerList, err := svc.client.DataService().Global.LoadBalancer.ListListener(cts.Kit, listenerReq)
	if err != nil {
		logs.Errorf("list listener failed, lbID: %s, err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(listenerList.Details) == 0 {
		return resList, nil
	}

	resList.Count = listenerList.Count
	lblIDs := make([]string, 0)
	for _, listenerItem := range listenerList.Details {
		lblIDs = append(lblIDs, listenerItem.ID)
		resList.Details = append(resList.Details, cslb.ListListenerBase{
			BaseListener: listenerItem,
		})
	}

	urlRuleMap, err := svc.listTCloudLbUrlRuleMap(cts.Kit, lbID, lblIDs)
	if err != nil {
		return nil, err
	}

	for idx, lblItem := range resList.Details {
		resList.Details[idx].Scheduler = urlRuleMap[lblItem.ID].Scheduler
		resList.Details[idx].DomainNum = urlRuleMap[lblItem.ID].DomainNum
		resList.Details[idx].UrlNum = urlRuleMap[lblItem.ID].UrlNum
	}

	return resList, nil
}

func (svc *lbSvc) listTCloudLbUrlRuleMap(kt *kit.Kit, lbID string, lblIDs []string) (
	map[string]cslb.ListListenerBase, error) {

	urlRuleReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleIn("lbl_id", lbID),
		),
		Page: core.NewDefaultBasePage(),
	}
	urlRuleList, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud url rule failed, lbID: %s, lblIDs: %v, err: %v, rid: %s", lbID, lblIDs, err, kt.Rid)
		return nil, err
	}

	listenerRuleMap := make(map[string]cslb.ListListenerBase, 0)
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := listenerRuleMap[ruleItem.LblID]; !ok {
			listenerRuleMap[ruleItem.LblID] = cslb.ListListenerBase{
				TargetGroupID: ruleItem.TargetGroupID,
				HealthCheck:   ruleItem.HealthCheck,
				Certificate:   ruleItem.Certificate,
			}
		}

		tmpListener := listenerRuleMap[ruleItem.LblID]
		// 只有4层规则，才在监听器列表显示负载均衡方式（7层在规则列表显示）
		if ruleItem.RuleType == enumor.LayerFourRuleType {
			tmpListener.Scheduler = ruleItem.Scheduler
		}
		if len(ruleItem.Domain) > 0 {
			tmpListener.DomainNum++
		}
		if len(ruleItem.URL) > 0 {
			tmpListener.UrlNum++
		}

		listenerRuleMap[ruleItem.LblID] = tmpListener
	}

	return listenerRuleMap, nil
}

func (svc *lbSvc) listListenerMap(kt *kit.Kit, lblIDs []string) (map[string]corelb.BaseListener, error) {
	if len(lblIDs) == 0 {
		return nil, nil
	}

	lblReq := &dataproto.ListListenerReq{
		ListReq: core.ListReq{
			Filter: tools.ContainersExpression("id", lblIDs),
			Page:   core.NewDefaultBasePage(),
		},
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

func (svc *lbSvc) getTCloudListener(kt *kit.Kit, lblID string) (*cslb.GetListenerDetail, error) {
	listenerInfo, err := svc.client.DataService().TCloud.LoadBalancer.GetListener(kt, lblID)
	if err != nil {
		logs.Errorf("[clb] get tcloud listener detail failed, lblID: %s, err: %v, rid: %s", lblID, err, kt.Rid)
	}

	urlRuleMap, err := svc.listTCloudLbUrlRuleMap(kt, listenerInfo.LbID, []string{lblID})
	if err != nil {
		return nil, err
	}

	targetGroupID := urlRuleMap[listenerInfo.ID].TargetGroupID
	result := &cslb.GetListenerDetail{
		BaseListener:  *listenerInfo,
		LblID:         listenerInfo.ID,
		LblName:       listenerInfo.Name,
		CloudLblID:    listenerInfo.CloudID,
		TargetGroupID: targetGroupID,
		DomainNum:     urlRuleMap[listenerInfo.ID].DomainNum,
		UrlNum:        urlRuleMap[listenerInfo.ID].UrlNum,
		HealthCheck:   urlRuleMap[listenerInfo.ID].HealthCheck,
		Certificate:   urlRuleMap[listenerInfo.ID].Certificate,
	}

	// 只有4层监听器才显示目标组信息
	if len(listenerInfo.DefaultDomain) == 0 {
		targetGroupList, err := svc.getTargetGroupByID(kt, targetGroupID, 0)
		if err != nil {
			return nil, err
		}
		if len(targetGroupList) > 0 {
			result.TargetGroupName = targetGroupList[0].Name
			result.CloudTargetGroupID = targetGroupList[0].CloudID
		}
	}

	return result, nil
}
