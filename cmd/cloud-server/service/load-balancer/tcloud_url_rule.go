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

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
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

// ListRuleByTG ...
func (svc *lbSvc) ListRuleByTG(cts *rest.Contexts) (interface{}, error) {
	return svc.listLbUrlRuleByTG(cts, handler.ResOperateAuth)
}

// ListBizRuleByTG ...
func (svc *lbSvc) ListBizRuleByTG(cts *rest.Contexts) (interface{}, error) {
	return svc.listLbUrlRuleByTG(cts, handler.BizOperateAuth)
}

// listLbUrlRuleByTG 返回目标组绑定的四层监听器或者七层规则（都能绑定目标组或者rs）
func (svc *lbSvc) listLbUrlRuleByTG(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (any, error) {

	tgID := cts.PathParameter("target_group_id").String()
	if len(tgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target_group_id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
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

	var urlRuleList *dataproto.TCloudURLRuleListResult
	switch vendor {
	case enumor.TCloud:
		urlRuleList, err = svc.listRuleWithCondition(cts.Kit, req, tools.RuleEqual("target_group_id", tgID))
		if err != nil {
			return nil, err
		}
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
	targetIDs := make([]string, 0)
	resList := &cslb.ListLbUrlRuleResult{Count: urlRuleList.Count, Details: make([]cslb.ListLbUrlRuleBase, 0)}
	for _, ruleItem := range urlRuleList.Details {
		lbIDs = append(lbIDs, ruleItem.LbID)
		targetIDs = append(targetIDs, ruleItem.TargetGroupID)
		resList.Details = append(resList.Details, cslb.ListLbUrlRuleBase{TCloudLbUrlRule: ruleItem})
	}

	// 批量获取lb信息
	lbMap, err := lblogic.ListLoadBalancerMap(kt, svc.client.DataService(), lbIDs)
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

// ListBizUrlRulesByListener 指定监听器下的url规则
func (svc *lbSvc) ListBizUrlRulesByListener(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
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
	switch vendor {
	case enumor.TCloud:
		result, err := svc.listRuleWithCondition(cts.Kit, req,
			tools.RuleEqual("lbl_id", lblID),
			tools.RuleEqual("rule_type", enumor.Layer7RuleType))
		if err != nil {
			logs.Errorf("fail to list rule under listener(id=%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
			return nil, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupport vendor for list rule: %s", vendor)
	}
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

// GetBizUrlRule 业务下url规则
func (svc *lbSvc) GetBizUrlRule(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
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

	var urlRuleList *dataproto.TCloudURLRuleListResult
	switch vendor {
	case enumor.TCloud:
		urlRuleList, err = svc.listRuleWithCondition(cts.Kit, req)
		if err != nil {
			logs.Errorf("fail to list rule, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupport vendor for list rule: %s", vendor)
	}
	if len(urlRuleList.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "rule not found, id: "+ruleID)
	}

	return urlRuleList.Details[0], nil
}

// ListBizRuleBindingStatus 获取规则绑定目标组状态
func (svc *lbSvc) ListBizRuleBindingStatus(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener is required")
	}
	req := new(cslb.RuleBindingStatusListReq)
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

	ruleRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("lbl_id", lblID),
			tools.RuleIn("listener_rule_id", req.RuleIDs),
		),
		Fields: []string{"listener_rule_id", "binding_status"},
		Page:   core.NewDefaultBasePage(),
	}
	ruleRelResp, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, ruleRelReq)
	if err != nil {
		logs.Errorf("list target group listener rule rel failed, err: %v, lblID: %s, ruleIDs: %v, rid: %s", err,
			lblID, req.RuleIDs, cts.Kit.Rid)
		return nil, err
	}
	ruleBindingStatusMap := make(map[string]enumor.BindingStatus)
	for _, rel := range ruleRelResp.Details {
		ruleBindingStatusMap[rel.ListenerRuleID] = rel.BindingStatus
	}

	resp := new(cslb.RuleBindingStatusListResp)
	for _, ruleID := range req.RuleIDs {
		bindStatus, ok := ruleBindingStatusMap[ruleID]
		if !ok {
			// 没有 rule和 target group的关联关系, 会走到这个逻辑, 默认返回未绑定状态
			bindStatus = enumor.UnBindingStatus
		}

		resp.Details = append(resp.Details, cslb.RuleBindingStatus{
			RuleID:     ruleID,
			BindStatus: bindStatus,
		})
	}

	return resp, nil
}
