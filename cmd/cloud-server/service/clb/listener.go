package clb

import (
	csclb "hcm/pkg/api/cloud-server/clb"
	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
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

// ListListener list listener.
func (svc *clbSvc) ListListener(cts *rest.Contexts) (interface{}, error) {
	return svc.listClbListener(cts, handler.ListResourceAuthRes)
}

// ListBizListener list biz listener.
func (svc *clbSvc) ListBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.listClbListener(cts, handler.ListBizAuthRes)
}

func (svc *clbSvc) listClbListener(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	clbID := cts.PathParameter("clb_id").String()
	if len(clbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "clb_id is required")
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
		ResType: meta.Clb, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list listener auth failed, clbID: %s, noPermFlag: %v, err: %v, rid: %s",
			clbID, noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	resList := &csclb.ListListenerResult{Count: 0, Details: make([]csclb.ListListenerBase, 0)}
	if noPermFlag {
		logs.Errorf("list listener no perm auth, clbID: %s, noPermFlag: %v, rid: %s", clbID, noPermFlag, cts.Kit.Rid)
		return resList, nil
	}

	listenerReq := &dataproto.ListListenerReq{
		ClbID: clbID,
		ListReq: core.ListReq{
			Filter: expr,
			Page:   req.Page,
		},
	}
	listenerList, err := svc.client.DataService().Global.LoadBalancer.ListListener(cts.Kit, listenerReq)
	if err != nil {
		logs.Errorf("[clb] list listener failed, clbID: %s, err: %v, rid: %s", clbID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(listenerList.Details) == 0 {
		return resList, nil
	}

	resList.Count = listenerList.Count
	lblIDs := make([]string, 0)
	for _, listenerItem := range listenerList.Details {
		lblIDs = append(lblIDs, listenerItem.ID)
		resList.Details = append(resList.Details, csclb.ListListenerBase{
			BaseListener: listenerItem,
		})
	}

	urlRuleMap, err := svc.listTCloudClbUrlRuleMap(cts.Kit, clbID, lblIDs)
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

func (svc *clbSvc) listTCloudClbUrlRuleMap(kt *kit.Kit, clbID string, lblIDs []string) (
	map[string]csclb.ListListenerBase, error) {

	urlRuleReq := &dataproto.ListTCloudURLRuleReq{
		ListReq: core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "lb_id",
						Op:    filter.Equal.Factory(),
						Value: clbID,
					},
					filter.AtomRule{
						Field: "lbl_id",
						Op:    filter.In.Factory(),
						Value: lblIDs,
					},
				},
			},
			Page: core.NewDefaultBasePage(),
		},
	}
	urlRuleList, err := svc.client.DataService().Global.LoadBalancer.ListUrlRule(kt, urlRuleReq)
	if err != nil {
		logs.Errorf("[clb] list tcloud url rule failed, clbID: %s, lblIDs: %v, err: %v, rid: %s",
			clbID, lblIDs, err, kt.Rid)
		return nil, err
	}

	listenerRuleMap := make(map[string]csclb.ListListenerBase, 0)
	for _, ruleItem := range urlRuleList.Details {
		if _, ok := listenerRuleMap[ruleItem.LblID]; !ok {
			listenerRuleMap[ruleItem.LblID] = csclb.ListListenerBase{
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

func (svc *clbSvc) listListenerMap(kt *kit.Kit, lblIDs []string) (map[string]coreclb.BaseListener, error) {
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

	lblMap := make(map[string]coreclb.BaseListener, len(lblList.Details))
	for _, clbItem := range lblList.Details {
		lblMap[clbItem.ID] = clbItem
	}

	return lblMap, nil
}

// GetListener get clb listener.
func (svc *clbSvc) GetListener(cts *rest.Contexts) (interface{}, error) {
	return svc.getListener(cts, handler.ListResourceAuthRes)
}

// GetBizListener get biz clb listener.
func (svc *clbSvc) GetBizListener(cts *rest.Contexts) (interface{}, error) {
	return svc.getListener(cts, handler.ListBizAuthRes)
}

func (svc *clbSvc) getListener(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
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
		//return nil, errf.New(errf.PermissionDenied, "permission denied for get listener")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.getTCloudListener(cts.Kit, id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

func (svc *clbSvc) getTCloudListener(kt *kit.Kit, lblID string) (*csclb.GetListenerDetail, error) {
	listenerInfo, err := svc.client.DataService().TCloud.LoadBalancer.GetListener(kt, lblID)
	if err != nil {
		logs.Errorf("[clb] get tcloud listener detail failed, lblID: %s, err: %v, rid: %s", lblID, err, kt.Rid)
	}

	urlRuleMap, err := svc.listTCloudClbUrlRuleMap(kt, listenerInfo.LbID, []string{lblID})
	if err != nil {
		return nil, err
	}

	targetGroupID := urlRuleMap[listenerInfo.ID].TargetGroupID
	result := &csclb.GetListenerDetail{
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
		targetGroupList, err := svc.getTargetGroupByID(kt, targetGroupID)
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

func (svc *clbSvc) getTargetGroupByID(kt *kit.Kit, targetGroupID string) ([]coreclb.BaseClbTargetGroup, error) {
	tgReq := &core.ListReq{
		Filter: tools.EqualExpression("id", targetGroupID),
		Page:   core.NewDefaultBasePage(),
	}
	targetGroupInfo, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, tgReq)
	if err != nil {
		logs.Errorf("[clb] list target group failed, tgID: %s, err: %v, rid: %s", targetGroupID, err, kt.Rid)
		return nil, err
	}

	return targetGroupInfo.Details, nil
}
