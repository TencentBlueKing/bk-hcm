package loadbalancer

import (
	"errors"
	"fmt"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	apits "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
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
		resList.Details = append(resList.Details, cslb.ListLbUrlRuleBase{TCloudLbUrlRule: ruleItem})
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
	return svc.listRuleWithCondition(cts.Kit, req,
		tools.RuleEqual("lbl_id", lblID),
		tools.RuleEqual("rule_type", enumor.Layer7RuleType))
}

// ListBizListenerDomains 指定监听器下的域名列表
func (svc *lbSvc) ListBizListenerDomains(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(core.ListReq)
	req.Filter = tools.ExpressionAnd(
		tools.RuleEqual("lbl_id", lblID),
		tools.RuleEqual("rule_type", enumor.Layer7RuleType))
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

	if !lbl.Protocol.IsLayer7Protocol() {
		return nil, errf.Newf(errf.InvalidParameter, "unsupported listener protocol type: %s", lbl.Protocol)
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
			tools.RuleEqual("rule_type", enumor.Layer7RuleType),
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

// CreateBizTCloudUrlRule 业务下新建腾讯云url规则 TODO: 改成一次只创建一个规则
func (svc *lbSvc) CreateBizTCloudUrlRule(cts *rest.Contexts) (any, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID < 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id id is required")
	}

	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	// 限制一次只能创建一条规则
	req := new(cslb.TCloudRuleCreate)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, lblBasicInfo, err := svc.getTCloudListenerByID(cts, bizID, lblID)
	if err != nil {
		return nil, err
	}

	// if SNI Switch is off, certificates can only be set in listener not its rule
	if lblInfo.SniSwitch == enumor.SniTypeClose && req.Certificate != nil {
		return nil, errf.New(errf.InvalidParameter, "can not set certificate on rule of sni_switch off listener")
	}

	// 业务校验、鉴权
	valOpt := &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Create,
		BasicInfo:  lblBasicInfo,
	}
	if err = handler.BizOperateAuth(cts, valOpt); err != nil {
		return nil, err
	}

	tg, err := svc.targetGroupBindCheck(cts.Kit, bizID, req.TargetGroupID)
	if err != nil {
		return nil, err
	}

	// 预检测-是否有执行中的负载均衡
	_, err = svc.checkResFlowRel(cts.Kit, lblInfo.LbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		return nil, err
	}

	hcReq := &hcproto.TCloudRuleBatchCreateReq{Rules: []hcproto.TCloudRuleCreate{convRuleCreate(req, tg)}}
	createResp, err := svc.client.HCService().TCloud.Clb.BatchCreateUrlRule(cts.Kit, lblID, hcReq)
	if err != nil {
		logs.Errorf("fail to create tcloud url rule, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(createResp.SuccessCloudIDs) == 0 {
		logs.Errorf("no rule have been created, lblID: %s, req: %+v, rid: %s", lblID, hcReq, cts.Kit.Rid)
		return nil, errors.New("create failed, reason: unknown")
	}
	err = svc.applyTargetToRule(cts.Kit, tg.ID, createResp.SuccessCloudIDs[0], lblInfo)
	if err != nil {
		logs.Errorf("fail to create target register flow, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return createResp, nil
}

// 构建异步任务将目标组中的RS绑定到对应规则上
func (svc *lbSvc) applyTargetToRule(kt *kit.Kit, tgID, ruleCloudID string, lblInfo *corelb.BaseListener) error {

	// 查找目标组中的rs
	listRsReq := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgID),
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: constant.BatchAddRSCloudMaxLimit,
		},
	}
	// Build Task
	tasks := make([]apits.CustomFlowTask, 0)
	getNextID := counter.NewNumStringCounter(1, 10)
	// 判断规则类型
	var ruleType enumor.RuleType
	if lblInfo.Protocol.IsLayer7Protocol() {
		ruleType = enumor.Layer7RuleType
	} else {
		ruleType = enumor.Layer4RuleType
	}
	// 按目标组数量拆分任务批次
	for {
		rsResp, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, listRsReq)
		if err != nil {
			logs.Errorf("fail to list target, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		if len(rsResp.Details) == 0 {
			break
		}

		rsReq := &hcproto.BatchRegisterTCloudTargetReq{
			CloudListenerID: lblInfo.CloudID,
			CloudRuleID:     ruleCloudID,
			TargetGroupID:   tgID,
			RuleType:        ruleType,
			Targets:         make([]*hcproto.RegisterTarget, 0, len(rsResp.Details)),
		}
		for _, target := range rsResp.Details {
			rsReq.Targets = append(rsReq.Targets, &hcproto.RegisterTarget{
				CloudInstID:      target.CloudInstID,
				InstType:         string(target.InstType),
				Port:             target.Port,
				Weight:           converter.PtrToVal(target.Weight),
				Zone:             target.Zone,
				InstName:         target.InstName,
				PrivateIPAddress: target.PrivateIPAddress,
				PublicIPAddress:  target.PublicIPAddress,
			})
		}
		tasks = append(tasks, apits.CustomFlowTask{
			ActionID:   action.ActIDType(getNextID()),
			ActionName: enumor.ActionListenerRuleAddTarget,
			Params: actionlb.ListenerRuleAddTargetOption{
				LoadBalancerID:               lblInfo.LbID,
				BatchRegisterTCloudTargetReq: rsReq,
			},
			DependOn: nil,
			Retry:    tableasync.NewRetryWithPolicy(10, 100, 500),
		})

		if len(rsResp.Details) < constant.BatchAddRSCloudMaxLimit {
			break
		}
		listRsReq.Page.Start += constant.BatchAddRSCloudMaxLimit
	}

	if len(tasks) == 0 {
		return nil
	}
	return svc.createApplyTGFlow(kt, tgID, lblInfo, tasks)
}

func (svc *lbSvc) createApplyTGFlow(kt *kit.Kit, tgID string, lblInfo *corelb.BaseListener,
	tasks []apits.CustomFlowTask) error {

	mainFlowResult, err := svc.client.TaskServer().CreateCustomFlow(kt, &apits.AddCustomFlowReq{
		Name:        enumor.FlowApplyTargetGroupToListenerRule,
		IsInitState: true,
		Tasks:       tasks,
	})
	if err != nil {
		logs.Errorf("fail to create target register flow, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	flowID := mainFlowResult.ID
	// 创建从任务并加锁
	flowWatchReq := &apits.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []apits.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowID,
				ResID:      lblInfo.LbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  []string{tgID},
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.ApplyTargetGroupType,
			},
		}},
	}
	_, err = svc.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call task server to create res flow status watch flow failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return err
	}

	// 锁定负载均衡跟Flow的状态
	err = svc.lockResFlowStatus(kt, lblInfo.LbID, enumor.LoadBalancerCloudResType, flowID, enumor.ApplyTargetGroupType)
	if err != nil {
		logs.Errorf("fail to lock load balancer(%s) for flow(%s), err: %v, rid: %s",
			lblInfo.LbID, flowID, err, kt.Rid)
		return err
	}
	return nil
}

func (svc *lbSvc) getTCloudListenerByID(cts *rest.Contexts, bizID int64, lblID string) (*corelb.BaseListener,
	*types.CloudResourceBasicInfo, error) {

	lblResp, err := svc.client.DataService().Global.LoadBalancer.ListListener(cts.Kit,
		&core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("id", lblID),
				tools.RuleEqual("vendor", enumor.TCloud),
				tools.RuleEqual("bk_biz_id", bizID)),
			Page: core.NewDefaultBasePage(),
		})
	if err != nil {
		logs.Errorf("fail to list listener(%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, nil, err
	}
	if len(lblResp.Details) == 0 {
		return nil, nil, errf.New(errf.RecordNotFound, "listener not found, id: "+lblID)
	}
	lblInfo := &lblResp.Details[0]
	basicInfo := &types.CloudResourceBasicInfo{
		ResType:   enumor.ListenerCloudResType,
		ID:        lblID,
		Vendor:    enumor.TCloud,
		AccountID: lblInfo.AccountID,
		BkBizID:   lblInfo.BkBizID,
	}

	return lblInfo, basicInfo, nil
}

func convRuleCreate(rule *cslb.TCloudRuleCreate, tg *corelb.BaseTargetGroup) hcproto.TCloudRuleCreate {
	return hcproto.TCloudRuleCreate{
		Url:                rule.Url,
		TargetGroupID:      rule.TargetGroupID,
		CloudTargetGroupID: tg.CloudID,
		Domains:            rule.Domains,
		SessionExpireTime:  rule.SessionExpireTime,
		Scheduler:          rule.Scheduler,
		ForwardType:        rule.ForwardType,
		DefaultServer:      rule.DefaultServer,
		Http2:              rule.Http2,
		TargetType:         rule.TargetType,
		Quic:               rule.Quic,
		TrpcFunc:           rule.TrpcFunc,
		TrpcCallee:         rule.TrpcCallee,
		HealthCheck:        tg.HealthCheck,
		Certificates:       rule.Certificate,
		Memo:               rule.Memo,
	}
}

// 目标组绑定检查，检查成功返回目标组id为索引的map
func (svc *lbSvc) targetGroupBindCheck(kt *kit.Kit, bizID int64, tgId string) (*corelb.BaseTargetGroup, error) {

	// 检查目标组是否存在
	tgResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bizID),
			tools.RuleEqual("id", tgId),
		),
		Page: core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to query target group(id:%s) info, err: %v, rid: %s", tgId, err, kt.Rid)
		return nil, err
	}

	if len(tgResp.Details) == 0 {
		logs.Errorf("target group can not be found, id: %s, rid: %s", tgId, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "target group(%s) can not be found", tgId)
	}
	tg := &tgResp.Details[0]
	// 检查对应的目标组是否被绑定
	relResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", tgId),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		return nil, err
	}
	if len(relResp.Details) > 0 {
		rel := relResp.Details[0]
		return nil, fmt.Errorf("target group(%s) already been bound to rule or listener(%s)",
			rel.TargetGroupID, rel.CloudListenerRuleID)
	}
	return tg, nil
}

// UpdateBizTCloudUrlRule 更新规则
func (svc *lbSvc) UpdateBizTCloudUrlRule(cts *rest.Contexts) (any, error) {

	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	ruleID := cts.PathParameter("rule_id").String()
	if len(ruleID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "rule id is required")
	}

	req := new(hcproto.TCloudRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts,
		&handler.ValidWithAuthOption{
			Authorizer: svc.authorizer,
			ResType:    meta.UrlRuleAuditResType,
			Action:     meta.Update,
			BasicInfo:  lblInfo,
		})
	if err != nil {
		return nil, err
	}

	// 更新审计
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert rule update request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ChildResUpdateAudit(cts.Kit, enumor.UrlRuleAuditResType, lblInfo.ID, ruleID, updateFields)
	if err != nil {
		logs.Errorf("create update rule audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, svc.client.HCService().TCloud.Clb.UpdateUrlRule(cts.Kit, lblID, ruleID, req)
}

// BatchDeleteBizTCloudUrlRule 批量删除规则
func (svc *lbSvc) BatchDeleteBizTCloudUrlRule(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(hcproto.TCloudRuleDeleteByIDReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Delete,
		BasicInfo:  lblInfo,
	})
	if err != nil {
		return nil, err
	}

	// 按规则删除审计
	err = svc.audit.ChildResDeleteAudit(cts.Kit, enumor.UrlRuleAuditResType, lblID, req.RuleIDs)
	if err != nil {
		logs.Errorf("create url rule delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, svc.client.HCService().TCloud.Clb.BatchDeleteUrlRule(cts.Kit, lblID, req)

}

// BatchDeleteBizTCloudUrlRuleByDomain 批量按域名删除规则
func (svc *lbSvc) BatchDeleteBizTCloudUrlRuleByDomain(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}

	req := new(hcproto.TCloudRuleDeleteByDomainReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lblInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.ListenerCloudResType, lblID)
	if err != nil {
		return nil, err
	}

	// 业务校验、鉴权
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.UrlRuleAuditResType,
		Action:     meta.Delete,
		BasicInfo:  lblInfo,
	})
	if err != nil {
		return nil, err
	}

	// 按域名删除审计
	err = svc.audit.ChildResDeleteAudit(cts.Kit, enumor.UrlRuleDomainAuditResType, lblID, req.Domains)
	if err != nil {
		logs.Errorf("create url rule delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, svc.client.HCService().TCloud.Clb.BatchDeleteUrlRuleByDomain(cts.Kit, lblID, req)

}
