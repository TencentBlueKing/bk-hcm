/*
 *
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
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// TCloudCreateUrlRule 创建url规则
func (svc *clbSvc) TCloudCreateUrlRule(cts *rest.Contexts) (any, error) {

	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	req := new(protolb.TCloudRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lb, listener, err := svc.getListenerWithLb(cts.Kit, lblID)
	if err != nil {
		return nil, err
	}
	if !listener.Protocol.IsLayer7Protocol() {
		return nil, errf.New(errf.InvalidParameter,
			"rule creation is only supports by layer 7 listener, not "+string(listener.Protocol))
	}

	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, listener.AccountID)
	if err != nil {
		return nil, err
	}

	ruleOption := typelb.TCloudCreateRuleOption{
		Region:         lb.Region,
		LoadBalancerId: lb.CloudID,
		ListenerId:     listener.CloudID,
	}
	ruleOption.Rules = slice.Map(req.Rules, convRuleCreate)
	creatResult, err := tcloudAdpt.CreateRule(cts.Kit, &ruleOption)
	if err != nil {
		logs.Errorf("create tcloud url rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := &cloud.TCloudUrlRuleBatchCreateReq{}

	for i, cloudID := range creatResult.SuccessCloudIDs {
		createReq.UrlRules = append(createReq.UrlRules, cloud.TCloudUrlRuleCreate{
			LbID:               lb.ID,
			CloudLbID:          lb.CloudID,
			LblID:              listener.ID,
			CloudLBLID:         listener.CloudID,
			CloudID:            cloudID,
			Name:               req.Rules[i].Url,
			RuleType:           enumor.Layer7RuleType,
			TargetGroupID:      "",
			CloudTargetGroupID: "",
			Domain:             req.Rules[i].Domains[0],
			URL:                req.Rules[i].Url,
			Scheduler:          cvt.PtrToVal(req.Rules[i].Scheduler),
			SessionType:        "",
			SessionExpire:      0,
			HealthCheck:        nil,
			Certificate:        nil,
			Memo:               nil,
		})
	}
	_, err = svc.dataCli.TCloud.LoadBalancer.BatchCreateTCloudUrlRule(cts.Kit, createReq)
	if err != nil {
		return nil, err
	}
	// TODO 同步对应监听器

	// if err := svc.lblSync(cts.Kit, tcloudAdpt, req.AccountID, req.Region, result.SuccessCloudIDs); err != nil {
	// 	return nil, err
	// }

	return creatResult, nil
}
func convRuleCreate(r protolb.TCloudRuleCreate) *typelb.RuleInfo {
	cloud := &typelb.RuleInfo{
		Url:               cvt.ValToPtr(r.Url),
		SessionExpireTime: r.SessionExpireTime,
		HealthCheck:       r.HealthCheck,
		Certificate:       r.Certificates,
		Scheduler:         r.Scheduler,
		ForwardType:       r.ForwardType,
		DefaultServer:     r.DefaultServer,
		Http2:             r.Http2,
		TargetType:        r.TargetType,
		TrpcCallee:        r.TrpcCallee,
		TrpcFunc:          r.TrpcFunc,
		Quic:              r.Quic,
	}
	if len(r.Domains) == 1 {
		cloud.Domain = cvt.ValToPtr(r.Domains[0])
	}
	if len(r.Domains) > 1 {
		cloud.Domains = cvt.SliceToPtr(r.Domains)
	}

	return cloud
}

// TCloudUpdateUrlRule 修改监听器规则
func (svc *clbSvc) TCloudUpdateUrlRule(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	ruleID := cts.PathParameter("rule_id").String()
	if len(ruleID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "rule id is required")
	}
	req := new(protolb.TCloudRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lb, rules, err := svc.getL7RulesWithLb(cts.Kit, lblID, []string{ruleID})
	if err != nil {
		return nil, err
	}

	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, lb.AccountID)
	if err != nil {
		return nil, err
	}

	ruleOption := typelb.TCloudUpdateRuleOption{
		Region:            lb.Region,
		LoadBalancerId:    lb.CloudID,
		ListenerId:        rules[0].CloudLBLID,
		LocationId:        rules[0].CloudID,
		Url:               req.Url,
		HealthCheck:       req.HealthCheck,
		Scheduler:         req.Scheduler,
		SessionExpireTime: req.SessionExpireTime,
		ForwardType:       req.ForwardType,
		TrpcCallee:        req.TrpcCallee,
		TrpcFunc:          req.TrpcFunc,
	}

	if err = tcloudAdpt.UpdateRule(cts.Kit, &ruleOption); err != nil {
		logs.Errorf("fail to update rule, err: %v, id: %s, rid: %s", err, ruleID, cts.Kit.Rid)
		return nil, err
	}
	// TODO 同步
	return nil, nil
}

// getL7RuleWithLb 查询同一个监听器下的规则
func (svc *clbSvc) getL7RulesWithLb(kt *kit.Kit, lblID string, ruleIDs []string) (*corelb.BaseLoadBalancer,
	[]corelb.BaseTCloudLbUrlRule, error) {

	// 只能查到7层规则
	ruleResp, err := svc.dataCli.TCloud.LoadBalancer.ListUrlRule(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", ruleIDs),
			tools.RuleEqual("lbl_id", lblID),
			tools.RuleEqual("rule_type", enumor.Layer7RuleType),
		),
		Page: core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list tcloud url rule, err: %v, ids: %s, rid: %s", err, ruleIDs, kt.Rid)
		return nil, nil, err
	}
	if len(ruleResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "rule not found")
	}
	rule := ruleResp.Details[0]

	// 查询负载均衡
	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(kt, &core.ListReq{
		Filter: tools.EqualExpression("id", rule.LbID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	})
	if err != nil {
		logs.Errorf("fail to tcloud list load balancer, err: %v, id: %s, rid: %s", err, rule.LbID, kt.Rid)
		return nil, nil, err
	}
	if len(lbResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "lb not found")
	}
	lb := lbResp.Details[0]
	return &lb, ruleResp.Details, nil
}

// TCloudBatchDeleteUrlRule 批量删除规则
func (svc *clbSvc) TCloudBatchDeleteUrlRule(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	req := new(protolb.TCloudBatchDeleteRuleReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var lb *corelb.BaseLoadBalancer
	var err error

	ruleOption := typelb.TCloudDeleteRuleOption{}
	dbReq := &cloud.LoadBalancerBatchDeleteReq{}
	if len(req.RuleIDs) > 0 {
		// 指定规则id删除
		var rules []corelb.BaseTCloudLbUrlRule
		lb, rules, err = svc.getL7RulesWithLb(cts.Kit, lblID, req.RuleIDs)
		if err != nil {
			logs.Errorf("fail to get lb info for rule deletion by rule ids(%v), err: %v, rid: %s",
				req.RuleIDs, err, cts.Kit.Rid)
			return nil, err
		}
		ruleOption.ListenerId = rules[0].CloudLBLID
		ruleOption.CloudIDs = slice.Map(rules, func(r corelb.BaseTCloudLbUrlRule) string { return r.CloudID })
		dbReq.Filter = tools.ContainersExpression("id", req.RuleIDs)

	} else {
		// 按域名删除模式
		var listener *corelb.BaseListener
		lb, listener, err = svc.getListenerWithLb(cts.Kit, lblID)
		if err != nil {
			logs.Errorf("fail to get lb info for rule deletion by domain(%s), err: %v, rid: %s",
				req.Domain, err, cts.Kit.Rid)
			return nil, err
		}
		ruleOption.ListenerId = listener.CloudID
		ruleOption.Domain = req.Domain
		ruleOption.NewDefaultServerDomain = req.NewDefaultDomain
		dbReq.Filter = tools.ExpressionAnd(
			tools.RuleEqual("domain", req.Domain),
			tools.RuleEqual("lbl_id", lblID))

	}

	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, lb.AccountID)
	if err != nil {
		return nil, err
	}

	ruleOption.Region = lb.Region
	ruleOption.LoadBalancerId = lb.CloudID
	if err = tcloudAdpt.DeleteRule(cts.Kit, &ruleOption); err != nil {
		logs.Errorf("fail to delete rule, err: %v, ids: %v, rid: %s", err, req.RuleIDs, cts.Kit.Rid)
		return nil, err
	}

	// TODO 同步对应监听器
	return svc.dataCli.TCloud.LoadBalancer.BatchDeleteTCloudUrlRule(cts.Kit, dbReq)
}
