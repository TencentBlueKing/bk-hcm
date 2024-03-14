package loadbalancer

import (
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListLbUrlRule list lb url rule.
func (svc *lbSvc) ListLbUrlRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listLbUrlRule(cts, handler.ListResourceAuthRes, constant.UnassignedBiz)
}

// ListBizLbUrlRule list biz lb url rule.
func (svc *lbSvc) ListBizLbUrlRule(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.listLbUrlRule(cts, handler.ListBizAuthRes, bkBizID)
}

func (svc *lbSvc) listLbUrlRule(cts *rest.Contexts, authHandler handler.ListAuthResHandler, bkBizID int64) (
	interface{}, error) {

	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tgList, err := svc.getTargetGroupByID(cts.Kit, tgID, bkBizID)
	if err != nil {
		return nil, err
	}
	if len(tgList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group %s is not found", tgID)
	}

	// list authorized instances
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list lb url rule auth failed, targetGroupID: %s, noPermFlag: %v, err: %v, rid: %s",
			tgID, noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		logs.Errorf("list lb url rule no perm auth, targetGroupID: %s, noPermFlag: %v, rid: %s",
			tgID, noPermFlag, cts.Kit.Rid)
		return &cslb.ListClbUrlRuleResult{Count: 0, Details: make([]cslb.ListClbUrlRuleBase, 0)}, nil
	}

	//urlRuleReq := core.ListReq{Filter: expr, Page: req.Page}
	urlRuleList, err := svc.getTCloudUrlRule(cts.Kit, tgID, req)

	logs.Errorf("list lb url rule no perm DEBUG:61, targetGroupID: %s, req: %+v, urlRuleList: %+v, rid: %s",
		tgID, req, urlRuleList, cts.Kit.Rid)
	if err != nil {
		return nil, err
	}

	if len(urlRuleList.Details) == 0 {
		return &cslb.ListClbUrlRuleResult{Count: 0, Details: make([]cslb.ListClbUrlRuleBase, 0)}, nil
	}

	resList, err := svc.listTCloudClbUrlRule(cts.Kit, urlRuleList)
	if err != nil {
		return nil, err
	}

	return resList, nil
}

func (svc *lbSvc) listTCloudClbUrlRule(kt *kit.Kit, urlRuleList *dataproto.TCloudURLRuleListResult) (
	interface{}, error) {

	lbIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	targetIDs := make([]string, 0)
	resList := &cslb.ListClbUrlRuleResult{Count: urlRuleList.Count, Details: make([]cslb.ListClbUrlRuleBase, 0)}
	for _, ruleItem := range urlRuleList.Details {
		lbIDs = append(lbIDs, ruleItem.LbID)
		lblIDs = append(lblIDs, ruleItem.LblID)
		targetIDs = append(targetIDs, ruleItem.TargetGroupID)
		resList.Details = append(resList.Details, cslb.ListClbUrlRuleBase{
			BaseTCloudClbURLRule: ruleItem,
		})
	}

	// 批量获取clb信息
	clbMap, err := svc.listLoadBalancerMap(kt, lbIDs)
	if err != nil {
		return nil, err
	}

	// 批量获取listener信息
	listenerMap, err := svc.listListenerMap(kt, lblIDs)
	if err != nil {
		return nil, err
	}

	// 批量获取vpc信息
	vpcIDs := make([]string, 0)
	for _, item := range clbMap {
		vpcIDs = append(vpcIDs, item.VpcID)
	}
	vpcMap, err := svc.listVpcMap(kt, vpcIDs)
	if err != nil {
		return nil, err
	}

	// 批量获取target信息
	targetMap, err := svc.listClbTargetMap(kt, targetIDs)
	if err != nil {
		return nil, err
	}

	for idx, ruleItem := range resList.Details {
		resList.Details[idx].LbName = clbMap[ruleItem.LbID].Name
		tmpVpcID := clbMap[ruleItem.LbID].VpcID
		resList.Details[idx].VpcID = tmpVpcID
		resList.Details[idx].CloudVpcID = clbMap[ruleItem.LbID].CloudVpcID
		resList.Details[idx].PrivateIPv4Addresses = clbMap[ruleItem.LbID].PrivateIPv4Addresses
		resList.Details[idx].PrivateIPv6Addresses = clbMap[ruleItem.LbID].PrivateIPv6Addresses
		resList.Details[idx].PublicIPv4Addresses = clbMap[ruleItem.LbID].PublicIPv4Addresses
		resList.Details[idx].PublicIPv6Addresses = clbMap[ruleItem.LbID].PublicIPv6Addresses

		resList.Details[idx].VpcName = vpcMap[tmpVpcID].Name

		resList.Details[idx].LblName = listenerMap[ruleItem.LblID].Name
		resList.Details[idx].Protocol = listenerMap[ruleItem.LblID].Protocol
		resList.Details[idx].Port = listenerMap[ruleItem.LblID].Port

		if len(targetMap[ruleItem.TargetGroupID]) > 0 {
			resList.Details[idx].InstType = targetMap[ruleItem.TargetGroupID][0].InstType
		}
	}

	return resList, nil
}

func (svc *lbSvc) getTCloudUrlRule(kt *kit.Kit, tgID string, listReq *core.ListReq) (
	*dataproto.TCloudURLRuleListResult, error) {

	tcloudUrlRuleReq := &dataproto.ListTCloudURLRuleReq{
		TargetGroupID: tgID,
		ListReq:       listReq,
	}
	urlRuleList, err := svc.client.DataService().Global.LoadBalancer.ListUrlRule(kt, tcloudUrlRuleReq)
	if err != nil {
		logs.Errorf("list tcloud url rule failed, targetGroupID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
		return nil, err
	}

	return urlRuleList, nil
}

func (svc *lbSvc) listVpcMap(kt *kit.Kit, vpcIDs []string) (map[string]cloud.BaseVpc, error) {
	if len(vpcIDs) == 0 {
		return nil, nil
	}

	vpcReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", vpcIDs),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.client.DataService().Global.Vpc.List(kt.Ctx, kt.Header(), vpcReq)
	if err != nil {
		logs.Errorf("[clb] list vpc failed, vpcIDs: %v, err: %v, rid: %s", vpcIDs, err, kt.Rid)
		return nil, err
	}

	vpcMap := make(map[string]cloud.BaseVpc, len(list.Details))
	for _, item := range list.Details {
		vpcMap[item.ID] = item
	}

	return vpcMap, nil
}

func (svc *lbSvc) listClbTargetMap(kt *kit.Kit, targetIDs []string) (map[string][]corelb.BaseClbTarget, error) {
	if len(targetIDs) == 0 {
		return nil, nil
	}

	targetReq := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", targetIDs),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
	if err != nil {
		logs.Errorf("[clb] list target failed, targetIDs: %v, err: %v, rid: %s", targetIDs, err, kt.Rid)
		return nil, err
	}

	targetMap := make(map[string][]corelb.BaseClbTarget, len(list.Details))
	for _, item := range list.Details {
		if _, ok := targetMap[item.TargetGroupID]; !ok {
			targetMap[item.TargetGroupID] = make([]corelb.BaseClbTarget, 0)
		}
		targetMap[item.TargetGroupID] = append(targetMap[item.TargetGroupID], item)
	}

	return targetMap, nil
}
