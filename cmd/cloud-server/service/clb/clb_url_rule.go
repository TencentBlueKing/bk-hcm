package clb

import (
	csclb "hcm/pkg/api/cloud-server/clb"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	coreclb "hcm/pkg/api/core/cloud/clb"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListClbUrlRule list clb url rule.
func (svc *clbSvc) ListClbUrlRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listClbUrlRule(cts, handler.ListResourceAuthRes)
}

// ListBizClbUrlRule list biz clb url rule.
func (svc *clbSvc) ListBizClbUrlRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listClbUrlRule(cts, handler.ListBizAuthRes)
}

func (svc *clbSvc) listClbUrlRule(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
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

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Clb, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list clb url rule auth failed, targetGroupID: %s, noPermFlag: %v, err: %v, rid: %s",
			tgID, noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		logs.Errorf("list clb url rule no perm auth, targetGroupID: %s, noPermFlag: %v, rid: %s",
			tgID, noPermFlag, cts.Kit.Rid)
		return &csclb.ListClbUrlRuleResult{Count: 0, Details: make([]csclb.ListClbUrlRuleBase, 0)}, nil
	}

	urlRuleReq := core.ListReq{Filter: expr, Page: req.Page}
	urlRuleList, err := svc.getTCloudUrlRule(cts.Kit, tgID, urlRuleReq)
	if err != nil {
		return nil, err
	}

	if len(urlRuleList.Details) == 0 {
		return &csclb.ListClbUrlRuleResult{Count: 0, Details: make([]csclb.ListClbUrlRuleBase, 0)}, nil
	}

	resList, err := svc.listTCloudClbUrlRule(cts.Kit, urlRuleList)
	if err != nil {
		return nil, err
	}

	return resList, nil
}

func (svc *clbSvc) listTCloudClbUrlRule(kt *kit.Kit, urlRuleList *dataproto.TCloudURLRuleListResult) (
	interface{}, error) {

	lbIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	targetIDs := make([]string, 0)
	resList := &csclb.ListClbUrlRuleResult{Count: urlRuleList.Count, Details: make([]csclb.ListClbUrlRuleBase, 0)}
	for _, ruleItem := range urlRuleList.Details {
		lbIDs = append(lbIDs, ruleItem.LbID)
		lblIDs = append(lblIDs, ruleItem.LblID)
		targetIDs = append(targetIDs, ruleItem.TargetGroupID)
		resList.Details = append(resList.Details, csclb.ListClbUrlRuleBase{
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

func (svc *clbSvc) getTCloudUrlRule(kt *kit.Kit, tgID string, listReq core.ListReq) (
	*dataproto.TCloudURLRuleListResult, error) {

	tcloudUrlRuleReq := &dataproto.ListTCloudURLRuleReq{
		TargetGroupID: tgID,
		ListReq:       listReq,
	}
	urlRuleList, err := svc.client.DataService().Global.LoadBalancer.ListClbWithUrlRule(kt, tcloudUrlRuleReq)
	if err != nil {
		logs.Errorf("[clb] list tcloud url rule failed, targetGroupID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
		return nil, err
	}

	return urlRuleList, nil
}

func (svc *clbSvc) listVpcMap(kt *kit.Kit, vpcIDs []string) (map[string]cloud.BaseVpc, error) {
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

func (svc *clbSvc) listClbTargetMap(kt *kit.Kit, targetIDs []string) (map[string][]coreclb.BaseClbTarget, error) {
	if len(targetIDs) == 0 {
		return nil, nil
	}

	targetReq := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", targetIDs),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.client.DataService().Global.LoadBalancer.ListClbTarget(kt, targetReq)
	if err != nil {
		logs.Errorf("[clb] list clb target failed, targetIDs: %v, err: %v, rid: %s", targetIDs, err, kt.Rid)
		return nil, err
	}

	targetMap := make(map[string][]coreclb.BaseClbTarget, len(list.Details))
	for _, item := range list.Details {
		if _, ok := targetMap[item.TargetGroupID]; !ok {
			targetMap[item.TargetGroupID] = make([]coreclb.BaseClbTarget, 0)
		}
		targetMap[item.TargetGroupID] = append(targetMap[item.TargetGroupID], item)
	}

	return targetMap, nil
}
