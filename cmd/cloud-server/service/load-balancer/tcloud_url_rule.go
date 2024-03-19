package loadbalancer

import (
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
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

	urlRuleList, err := svc.listRuleWithCondition(cts.Kit, req, tools.RuleEqual("target_group_id", tgID))
	if err != nil {
		return nil, err
	}

	resList, err := svc.fillRuleRelatedRes(cts.Kit, urlRuleList)
	if err != nil {
		return nil, err
	}

	return resList, nil
}

// fillRuleRelatedRes 填充 监听器的协议、端口信息、所在vpc信息
func (svc *lbSvc) fillRuleRelatedRes(kt *kit.Kit, urlRuleList *dataproto.TCloudURLRuleListResult) (
	any, error) {

	if len(urlRuleList.Details) == 0 {
		return &cslb.ListLbUrlRuleResult{Count: urlRuleList.Count, Details: make([]cslb.ListLbUrlRuleBase, 0)}, nil
	}

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

	// 批量获取lb信息
	lbMap, err := lblogic.ListLoadBalancerMap(kt, svc.client.DataService(), lbIDs)
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
	for _, item := range lbMap {
		vpcIDs = append(vpcIDs, item.VpcID)
	}

	vpcMap, err := svc.listVpcMap(kt, vpcIDs)
	if err != nil {
		return nil, err
	}

	for idx, ruleItem := range resList.Details {
		resList.Details[idx].LbName = lbMap[ruleItem.LbID].Name
		tmpVpcID := lbMap[ruleItem.LbID].VpcID
		resList.Details[idx].VpcID = tmpVpcID
		resList.Details[idx].CloudVpcID = lbMap[ruleItem.LbID].CloudVpcID
		resList.Details[idx].PrivateIPv4Addresses = lbMap[ruleItem.LbID].PrivateIPv4Addresses
		resList.Details[idx].PrivateIPv6Addresses = lbMap[ruleItem.LbID].PrivateIPv6Addresses
		resList.Details[idx].PublicIPv4Addresses = lbMap[ruleItem.LbID].PublicIPv4Addresses
		resList.Details[idx].PublicIPv6Addresses = lbMap[ruleItem.LbID].PublicIPv6Addresses

		resList.Details[idx].VpcName = vpcMap[tmpVpcID].Name

		resList.Details[idx].LblName = listenerMap[ruleItem.LblID].Name
		resList.Details[idx].Protocol = listenerMap[ruleItem.LblID].Protocol
		resList.Details[idx].Port = listenerMap[ruleItem.LblID].Port

	}

	return resList, nil
}

// listRuleWithCondition list rule with additional rules
func (svc *lbSvc) listRuleWithCondition(kt *kit.Kit, listReq *core.ListReq, conditions ...filter.RuleFactory) (
	*dataproto.TCloudURLRuleListResult, error) {

	req := &core.ListReq{
		Filter: listReq.Filter,
		Page:   listReq.Page,
		Fields: listReq.Fields,
	}
	if len(conditions) > 0 {
		conditions = append(conditions, listReq.Filter)
		combinedFilter, err := tools.And(conditions...)
		if err != nil {
			logs.Errorf("fail to merge list request, err: %v, listReq: %+v, rid: %s", err, listReq, kt.Rid)
			return nil, err
		}
		req.Filter = combinedFilter
	}

	urlRuleList, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, req)
	if err != nil {
		logs.Errorf("list tcloud url with rule failed,  err: %v, req: %+v, conditions: %+v, rid: %s",
			err, listReq, conditions, kt.Rid)
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

// ListBizUrlRulesByListener 指定监听器下的url规则
func (svc *lbSvc) ListBizUrlRulesByListener(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.Listener,
		Action:     meta.Find,
		BasicInfo:  basicInfo,
	})
	if err != nil {
		return nil, err
	}

	// 查询规则列表
	return svc.listRuleWithCondition(cts.Kit, req, tools.RuleEqual("lbl_id", lblID))
}

// ListBizListenerDomains 指定监听器下的域名列表
func (svc *lbSvc) ListBizListenerDomains(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(core.ListReq)
	req.Filter = tools.EqualExpression("lbl_id", lblID)
	req.Page = core.NewDefaultBasePage()

	lbl, err := svc.client.DataService().TCloud.LoadBalancer.GetListener(cts.Kit, lblID)
	if err != nil {
		logs.Errorf("fail to get listener, err: %v, id: %s, rid: %s", err, lblID, cts.Kit.Rid)
		return nil, err
	}

	// 业务校验、鉴权
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.Listener,
			Action: meta.Find,
		},
		BizID: lbl.BkBizID,
	})
	if err != nil {
		return nil, err
	}

	// 查询规则列表
	ruleList, err := svc.listRuleWithCondition(cts.Kit, req)
	if err != nil {
		logs.Errorf("fail list rule under listener(id=%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}

	// 统计url数量
	domainList := make([]cslb.DomainInfo, 0)
	domainIndex := make(map[string]int)
	for _, detail := range ruleList.Details {
		if _, exists := domainIndex[detail.Domain]; !exists {
			domainIndex[detail.Domain] = len(domainList)
			domainList = append(domainList, cslb.DomainInfo{Domain: detail.Domain})
		}
		domainList[domainIndex[detail.Domain]].UrlCount += 1
	}

	return cslb.GetListenerDomainResult{
		DefaultDomain: lbl.DefaultDomain,
		DomainList:    domainList,
	}, nil
}

// GetBizTCloudUrlRule 业务下腾讯云url规则
func (svc *lbSvc) GetBizTCloudUrlRule(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	ruleID := cts.PathParameter("rule_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "rule is required")
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts,
		&handler.ValidWithAuthOption{
			Authorizer: svc.authorizer,
			ResType:    meta.TargetGroup,
			Action:     meta.Find,
			BasicInfo:  lblInfo,
		})
	if err != nil {
		return nil, err
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", ruleID),
			tools.RuleEqual("lbl_id", lblID),
		),
		Page: core.NewDefaultBasePage(),
	}

	urlRuleList, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(cts.Kit, req)
	if err != nil {
		logs.Errorf("list tcloud url failed, err: %v, lblID: %s, ruleID: %s, rid: %s",
			err, lblID, ruleID, cts.Kit.Rid)
		return nil, err
	}
	if len(urlRuleList.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "rule not found, id: "+ruleID)
	}

	return urlRuleList.Details[0], nil
}
