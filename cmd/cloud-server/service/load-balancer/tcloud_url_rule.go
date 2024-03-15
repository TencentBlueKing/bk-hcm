package loadbalancer

import (
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
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

// ListTCloudRuleByTG ...
func (svc *lbSvc) ListTCloudRuleByTG(cts *rest.Contexts) (interface{}, error) {
	return svc.listTCloudLbUrlRuleByTG(cts, handler.ResOperateAuth)
}

// ListBizTCloudRuleByTG ...
func (svc *lbSvc) ListBizTCloudRuleByTG(cts *rest.Contexts) (interface{}, error) {
	return svc.listTCloudLbUrlRuleByTG(cts, handler.BizOperateAuth)
}

// listTCloudLbUrlRuleByTG 返回目标组绑定的四层监听器或者七层规则（都能绑定目标组或者rs）
func (svc *lbSvc) listTCloudLbUrlRuleByTG(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (any, error) {

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

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.TargetGroupCloudResType, tgID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	urlRuleList, err := svc.listRuleByTargetGroup(cts.Kit, tgID, req)
	if err != nil {
		return nil, err
	}

	if len(urlRuleList.Details) == 0 {
		return &cslb.ListLbUrlRuleResult{Count: urlRuleList.Count, Details: make([]cslb.ListLbUrlRuleBase, 0)}, nil
	}

	resList, err := svc.fillRuleRelatedRes(cts.Kit, urlRuleList)
	if err != nil {
		return nil, err
	}

	return resList, nil
}

// fillRuleRelatedRes 填充 监听器、vpc相关信息
func (svc *lbSvc) fillRuleRelatedRes(kt *kit.Kit, urlRuleList *dataproto.TCloudURLRuleListResult) (
	interface{}, error) {

	lbIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	targetIDs := make([]string, 0)
	resList := &cslb.ListLbUrlRuleResult{Count: urlRuleList.Count, Details: make([]cslb.ListLbUrlRuleBase, 0)}
	for _, ruleItem := range urlRuleList.Details {
		lbIDs = append(lbIDs, ruleItem.LbID)
		lblIDs = append(lblIDs, ruleItem.LblID)
		targetIDs = append(targetIDs, ruleItem.TargetGroupID)
		resList.Details = append(resList.Details, cslb.ListLbUrlRuleBase{BaseTCloudLbUrlRule: ruleItem})
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

	}

	return resList, nil
}

func (svc *lbSvc) listRuleByTargetGroup(kt *kit.Kit, tgID string, listReq *core.ListReq) (
	*dataproto.TCloudURLRuleListResult, error) {

	combinedFilter, err := tools.And(listReq.Filter,
		filter.AtomRule{Field: "target_group_id", Op: filter.Equal.Factory(), Value: tgID})
	if err != nil {
		logs.Errorf("fail to merge list request, err: %v, listReq: %+v, rid: %s", err, listReq, kt.Rid)
		return nil, err
	}

	req := &core.ListReq{
		Filter: combinedFilter,
		Page:   listReq.Page,
		Fields: listReq.Fields,
	}
	urlRuleList, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, req)
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
