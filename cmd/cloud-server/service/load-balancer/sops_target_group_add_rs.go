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
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchBizAddTargetGroupRS batch biz add target group rs.
func (svc *lbSvc) BatchBizAddTargetGroupRS(cts *rest.Contexts) (any, error) {
	return svc.batchAddTargetGroupRS(cts, handler.BizOperateAuth)
}

// BatchAddTargetGroupRS batch add target group rs.
func (svc *lbSvc) BatchAddTargetGroupRS(cts *rest.Contexts) (any, error) {
	return svc.batchAddTargetGroupRS(cts, handler.ResOperateAuth)
}

func (svc *lbSvc) batchAddTargetGroupRS(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch sops add target group rs request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch sops add target auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get sops account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.buildCreateTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildCreateTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string) (any, error) {
	req := new(cslb.TCloudSopsTargetBatchCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询规则列表，查出符合条件的目标组
	tgIDsMap, err := svc.parseSOpsTargetParams(kt, accountID, req.RuleQueryList)
	if err != nil {
		return nil, err
	}
	if len(tgIDsMap) == 0 {
		return nil, errf.New(errf.RecordNotFound, "no matching target groups were found")
	}

	// 按照clb分组targetGroup
	lbTgsMap := make(map[string][]string)
	for _, tgIDs := range tgIDsMap {
		for _, tgID := range tgIDs {
			// 根据目标组ID，获取目标组绑定的监听器、规则列表
			ruleRelReq := &core.ListReq{
				Filter: tools.EqualExpression("target_group_id", tgID),
				Page:   core.NewDefaultBasePage(),
			}
			ruleRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
			if err != nil {
				logs.Errorf("list tcloud listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
				return nil, err
			}

			// 未绑定监听器及规则，分组到lbID=-1的默认组中
			lbID := "-1"
			if len(ruleRelList.Details) != 0 {
				// 已经绑定了监听器及规则，归属某一clb
				lbID = ruleRelList.Details[0].LbID
			}
			if _, exists := lbTgsMap[lbID]; !exists {
				lbTgsMap[lbID] = make([]string, 0)
			}
			lbTgsMap[lbID] = append(lbTgsMap[lbID], tgID)
		}
	}

	// 根据RS IP获取CVM的云端ID
	instCloudIDMap := make(map[string]string)
	switch req.RsType {
	case enumor.CvmInstType:
		instCloudIDMap, err = svc.parseTCloudRsIPForCvmInstIDMap(kt, accountID, req)
		if err != nil {
			return nil, err
		}
	case enumor.EniInstType:
		// ENI也同样的去CVM表中查询，查不到则报错（表示没有找到ENI绑定的CVM）
		instCloudIDMap, err = svc.parseTCloudRsIPForCvmInstIDMap(kt, accountID, req)
		if err != nil {
			return nil, err
		}
	}

	// 获取到targets
	targets := make([]*dataproto.TargetBaseReq, 0)
	for idx, tmpIP := range req.RsIP {
		tmpCloudInstID, ok := instCloudIDMap[tmpIP]
		if !ok {
			continue
		}
		if err != nil {
			return nil, err
		}
		targets = append(targets, &dataproto.TargetBaseReq{
			InstType:    req.RsType,
			IP:          tmpIP,
			CloudInstID: tmpCloudInstID,
			Port:        int64(req.RsPort[idx]),
			Weight:      cvt.ValToPtr(req.RsWeight),
		})
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("target list is empty")
	}

	flowStateResults := make([]*core.FlowStateResult, 0)
	for lbID, lbTgIDs := range lbTgsMap {
		params := &cslb.TCloudTargetBatchCreateReq{
			TargetGroups: []*cslb.TCloudBatchAddTargetReq{},
		}
		for _, tgID := range lbTgIDs {
			tmpTargetReq := &cslb.TCloudBatchAddTargetReq{
				TargetGroupID: tgID,
				Targets:       targets,
			}

			params.TargetGroups = append(params.TargetGroups, tmpTargetReq)
		}

		if len(params.TargetGroups) == 0 {
			logs.Errorf("build sops tcloud add target params parse failed, err: %v, accountID: %s, lbTgIDs: %v, rid: %s",
				err, accountID, lbTgIDs, kt.Rid)
			return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("build add target param parse empty"))
		}

		addTargetJSON, err := json.Marshal(params)
		if err != nil {
			logs.Errorf("build sops tcloud add target params marshal failed, err: %v, params: %+v, rid: %s",
				err, params, kt.Rid)
			return nil, err
		}

		// 记录标准运维参数转换后的数据，方便排查问题
		logs.Infof("build sops tcloud add target params jsonmarshal success, lbID: %s, lbTgIDs: %v, addTargetJSON: %s, rid: %s",
			lbID, lbTgIDs, addTargetJSON, kt.Rid)

		result, err := svc.buildAddTCloudTarget(kt, addTargetJSON, accountID)
		if err != nil {
			return nil, err
		}
		resultValue, ok := result.(*core.FlowStateResult)
		if !ok {
			return nil, fmt.Errorf("buildAddTCloudTarget failed, result: %v", resultValue)
		}
		flowStateResults = append(flowStateResults, resultValue)
	}

	return flowStateResults, nil
}

// parseTCloudRsIPForCvmInstIDMap 解析标准运维参数-根据RS IP获取CVM的云端ID
func (svc *lbSvc) parseTCloudRsIPForCvmInstIDMap(kt *kit.Kit, accountID string,
	req *cslb.TCloudSopsTargetBatchCreateReq) (map[string]string, error) {

	instCloudIDMap := make(map[string]string)
	for _, tmpRsIP := range req.RsIP {
		// 已有相同的ip映射记录
		if len(instCloudIDMap[tmpRsIP]) != 0 {
			continue
		}
		cvmReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", enumor.TCloud),
				tools.RuleEqual("account_id", accountID),
				tools.RuleJSONContains("private_ipv4_addresses", tmpRsIP),
			),
			Page: core.NewDefaultBasePage(),
		}
		cvmList, err := svc.client.DataService().Global.Cvm.ListCvm(kt, cvmReq)
		if err != nil {
			logs.Errorf("list cvm by tcloud rs ip failed, accountID: %s, privateRsIP: %s, err: %v, rid: %s",
				accountID, tmpRsIP, err, kt.Rid)
			return nil, err
		}
		if len(cvmList.Details) != 1 {
			// 一个rs_ip应对应唯一的一个CVM
			logs.Errorf("one rs_ip: %s should correspond to one CVM, but %d were found", tmpRsIP, len(cvmList.Details))
			return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("one rs_ip: %s should correspond to one CVM, but %d were found", tmpRsIP, len(cvmList.Details)))
		}

		// 记录当前temRsIP与CVMCloudID的映射关系
		instCloudIDMap[tmpRsIP] = cvmList.Details[0].CloudID
	}
	return instCloudIDMap, nil
}

// parseSOpsTargetParams 解析标准运维参数
func (svc *lbSvc) parseSOpsTargetParams(kt *kit.Kit, accountID string,
	ruleQueryList []cslb.TargetGroupRuleQueryItem) (map[int][]string, error) {

	tgIDsMap := make(map[int][]string)
	index := 1
	for _, item := range ruleQueryList {
		var protoDomainTgIDs, rsIpTypeTgIDs, vIpPortTgIDs, tgIDsItem []string
		var err error

		// 根据Protocol和Domain查询UrlRule，获取对应的目标组ID
		protoDomainTgIDs, err = svc.parseSOpsProtocolAndDomainForTgIDs(kt, accountID, item)
		if err != nil {
			logs.Errorf("parse protocol and domain for target group failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, err
		}
		if protoDomainTgIDs != nil {
			tgIDsItem = protoDomainTgIDs
		}

		//  根据RsIp和RsType查询Target，获取对应的目标组ID
		rsIpTypeTgIDs, err = svc.parseSOpsRsIpAndRsTypeForTgIDs(kt, accountID, item)
		if err != nil {
			logs.Errorf("parse rsip and rstype for target group failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, err
		}
		if rsIpTypeTgIDs != nil {
			if tgIDsItem != nil {
				// 取交集
				tgIDsItem = slice.Intersection(tgIDsItem, rsIpTypeTgIDs)
			} else {
				tgIDsItem = rsIpTypeTgIDs
			}
		}

		// 根据Vip和Vport查询到对应负载均衡下的对应监听器下的UrlRule，获取对应的目标组ID
		vIpPortTgIDs, err = svc.parseSOpsVipAndVportForTgIDs(kt, accountID, item)
		if err != nil {
			logs.Errorf("parse vip and vport for target group failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, err
		}
		if vIpPortTgIDs != nil {
			if tgIDsItem != nil {
				// 取交集
				tgIDsItem = slice.Intersection(tgIDsItem, vIpPortTgIDs)
			} else {
				tgIDsItem = vIpPortTgIDs
			}
		}

		tgIDsItem = slice.Unique(tgIDsItem)
		if len(tgIDsItem) == 0 {
			return nil, fmt.Errorf("no matching target groups were found for line %d", index)
		}

		// 分别记录每一行条件查询出的目标组ID列表
		tgIDsMap[index] = tgIDsItem
		index++
	}

	return tgIDsMap, nil
}

// parseSOpsProtocolAndDomainForTgIDs 根据Protocol和Domain查询UrlRule，获取对应的目标组ID
func (svc *lbSvc) parseSOpsProtocolAndDomainForTgIDs(kt *kit.Kit, accountID string,
	item cslb.TargetGroupRuleQueryItem) ([]string, error) {

	if len(item.Protocol) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}

	// 筛选查询urlRule
	var urlRuleFilter *filter.Expression
	var err error
	if item.Protocol.IsLayer7Protocol() {
		urlRuleFilter = tools.ExpressionAnd(
			tools.RuleEqual("rule_type", enumor.Layer7RuleType),
		)
		if len(item.Domain) != 0 {
			urlRuleFilter, err = tools.And(urlRuleFilter, tools.RuleEqual("domain", item.Domain))
			if err != nil {
				return nil, err
			}
		}
	} else if item.Protocol.IsLayer4Protocol() {
		urlRuleFilter = tools.ExpressionAnd(
			tools.RuleEqual("rule_type", enumor.Layer4RuleType),
		)
	} else {
		return nil, fmt.Errorf("protocol: %s not support", item.Protocol)
	}

	tgRuleReq := &core.ListReq{
		Filter: urlRuleFilter,
		Page:   core.NewDefaultBasePage(),
	}
	urlRuleResult, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, tgRuleReq)
	if err != nil {
		return nil, err
	}

	// 记录urlRule对应的目标组ID
	tgIDs := make([]string, 0)
	for _, ruleItem := range urlRuleResult.Details {
		if len(ruleItem.TargetGroupID) == 0 {
			continue
		}
		tgIDs = append(tgIDs, ruleItem.TargetGroupID)
	}

	return slice.Unique(tgIDs), nil
}

// parseSOpsRsIpAndRsTypeForTgIDs 根据RsIp和RsType查询Target，获取对应的目标组ID
func (svc *lbSvc) parseSOpsRsIpAndRsTypeForTgIDs(kt *kit.Kit, accountID string,
	item cslb.TargetGroupRuleQueryItem) ([]string, error) {

	if len(item.RsIP) == 0 || len(item.RsType) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}

	tgIDs := make([]string, 0)
	for _, rsIP := range item.RsIP {
		// 查询出对应的目标
		filter := tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("inst_type", item.RsType),
			tools.RuleJSONContains("private_ip_address", rsIP))
		targetReq := &core.ListReq{
			Fields: []string{"target_group_id"},
			Filter: filter,
			Page:   core.NewDefaultBasePage(),
		}
		targetResult, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, targetReq)
		if err != nil {
			return nil, err
		}

		// 记录目标对应的目标组ID
		for _, target := range targetResult.Details {
			if len(target.TargetGroupID) == 0 {
				continue
			}
			tgIDs = append(tgIDs, target.TargetGroupID)
		}
	}

	return slice.Unique(tgIDs), nil
}

// parseSOpsVipAndVportForTgIDs 根据Vip和Vport查询到对应负载均衡下的对应监听器下的UrlRule，获取对应的目标组ID
func (svc *lbSvc) parseSOpsVipAndVportForTgIDs(kt *kit.Kit, accountID string,
	item cslb.TargetGroupRuleQueryItem) ([]string, error) {

	if len(item.Vip) == 0 && len(item.VPort) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}

	lbIDs := make([]string, 0)
	tgIDs := make([]string, 0)
	if len(item.Vip) != 0 {
		// 若有vip筛选条件，则查询符合的负载均衡列表
		for _, vip := range item.Vip {
			lbReq := &core.ListReq{
				Filter: tools.ExpressionAnd(
					tools.RuleEqual("vendor", enumor.TCloud),
					tools.RuleEqual("account_id", accountID),
					tools.RuleEqual("region", item.Region),
					tools.RuleJSONContains("public_ipv4_addresses", vip),
				),
				Page: core.NewDefaultBasePage(),
			}
			lbList, err := svc.client.DataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
			if err != nil {
				return nil, err
			}

			for _, lbItem := range lbList.Details {
				lbIDs = append(lbIDs, lbItem.ID)
			}
		}
		// 若没有对应vip的负载均衡，则直接返回
		if len(lbIDs) == 0 {
			return tgIDs, nil
		}
	}

	// 查询符合的监听器列表
	lblFilter := tools.ExpressionAnd(
		tools.RuleEqual("vendor", enumor.TCloud),
	)
	var err error
	if len(lbIDs) != 0 {
		lblFilter, err = tools.And(lblFilter, tools.RuleIn("lb_id", lbIDs))
		if err != nil {
			return nil, err
		}
	}
	if len(item.VPort) != 0 {
		lblFilter, err = tools.And(lblFilter, tools.RuleIn("port", item.VPort))
		if err != nil {
			return nil, err
		}
	}
	lblReq := &core.ListReq{
		Filter: lblFilter,
		Page:   core.NewDefaultBasePage(),
	}
	lblList, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		return nil, err
	}
	lblIDs := slice.Map(lblList.Details, func(lbl corelb.BaseListener) string {
		return lbl.ID
	})
	if len(lblIDs) == 0 {
		return tgIDs, nil
	}

	// 查询符合的监听器与目标组绑定关系的列表
	lblRuleReq := &core.ListReq{
		Fields: []string{"target_group_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleIn("lbl_id", lblIDs),
			tools.RuleEqual("binding_status", enumor.SuccessBindingStatus),
		),
		Page: core.NewDefaultBasePage(),
	}
	lblRuleList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(
		kt, lblRuleReq)
	if err != nil {
		return nil, err
	}

	for _, ruleRelItem := range lblRuleList.Details {
		if len(ruleRelItem.TargetGroupID) == 0 {
			continue
		}
		tgIDs = append(tgIDs, ruleRelItem.TargetGroupID)
	}

	return slice.Unique(tgIDs), nil
}
