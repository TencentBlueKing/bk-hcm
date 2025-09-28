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
		return svc.buildCreateTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID, enumor.TCloud)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildCreateTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string, vendor enumor.Vendor) (any, error) {
	req := new(cslb.TCloudSopsTargetBatchCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询规则列表，查出符合条件的目标组
	tgIDsMap, err := svc.parseSOpsTargetParamsForRsOnline(kt, accountID, vendor, req.RuleQueryList)
	if err != nil {
		return nil, err
	}
	if len(tgIDsMap) == 0 {
		return nil, errf.New(errf.RecordNotFound, "no matching target groups were found")
	}

	// 按照clb分组targetGroup
	lbTgsMap, err := svc.iterateTargetGroupGroupByCLB(kt, tgIDsMap)
	if err != nil {
		logs.Errorf("iterate target group group by clb failed, tgIDsMap: %v, err: %v, rid: %s",
			tgIDsMap, err, kt.Rid)
		return nil, err
	}
	// 根据RS IP获取CVM的云端ID
	instCloudIDMap, err := svc.parseInstIDMap(kt, accountID, vendor, req)
	if err != nil {
		logs.Errorf("parse inst id map failed, accountID: %s, vendor: %s, err: %v, req: %+v, rid: %s",
			accountID, vendor, err, req, kt.Rid)
		return nil, err
	}

	// 获取到targets
	targets := make([]*dataproto.TargetBaseReq, 0)
	for idx, tempIP := range req.RsIP {
		tmpCloudInstID, ok := instCloudIDMap[tempIP]
		if !ok {
			logs.Infof("not find inst by ip: %s", tempIP)
			continue
		}
		targets = append(targets, &dataproto.TargetBaseReq{
			InstType:    req.RsType,
			IP:          tempIP,
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
		addTargetJSON, err := buildTCloudTargetBatchCreateReq(lbTgIDs, targets)
		if err != nil {
			logs.Errorf("build sops tcloud add target params failed, "+
				"err: %v, lbID: %s, lbTgIDs: %v, targets: %v, rid: %s", err, lbID, lbTgIDs, targets, kt.Rid)
			return nil, err
		}
		// 记录标准运维参数转换后的数据，方便排查问题
		logs.Infof("build sops tcloud add target params jsonmarshal success,"+
			" lbID: %s, lbTgIDs: %v, addTargetJSON: %s, rid: %s",
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

func buildTCloudTargetBatchCreateReq(lbTgIDs []string, targets []*dataproto.TargetBaseReq) ([]byte, error) {
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
		return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("build add target param parse empty"))
	}
	addTargetJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return addTargetJSON, nil
}

func (svc *lbSvc) parseInstIDMap(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	req *cslb.TCloudSopsTargetBatchCreateReq) (map[string]string, error) {

	switch req.RsType {
	case enumor.CvmInstType:
		instCloudIDMap, err := svc.parseTCloudRsIPForCvmInstIDMap(kt, accountID, vendor, req)
		if err != nil {
			logs.Errorf("parse tcloud rs ip for cvm inst id map failed, err: %v, req: %+v, rid : %s", err, req, kt.Rid)
			return nil, err
		}
		return instCloudIDMap, nil
	case enumor.EniInstType:
		// ENI也同样的去CVM表中查询，查不到则报错（表示没有找到ENI绑定的CVM）
		instCloudIDMap, err := svc.parseTCloudRsIPForCvmInstIDMap(kt, accountID, vendor, req)
		if err != nil {
			logs.Errorf("parse tcloud rs ip for cvm inst id map failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		return instCloudIDMap, nil
	default:
		return nil, fmt.Errorf("unsupport rs type: %s", req.RsType)
	}
}

// parseTCloudRsIPForCvmInstIDMap 解析标准运维参数-根据RS IP获取CVM的云端ID
func (svc *lbSvc) parseTCloudRsIPForCvmInstIDMap(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	req *cslb.TCloudSopsTargetBatchCreateReq) (map[string]string, error) {

	instCloudIDMap := make(map[string]string)
	for _, tmpRsIP := range req.RsIP {
		// 已有相同的ip映射记录
		if len(instCloudIDMap[tmpRsIP]) != 0 {
			continue
		}
		cvmReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
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
			return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf(
				"one rs_ip: %s should correspond to one CVM, but %d were found", tmpRsIP, len(cvmList.Details)))
		}

		// 记录当前temRsIP与CVMCloudID的映射关系
		instCloudIDMap[tmpRsIP] = cvmList.Details[0].CloudID
	}
	return instCloudIDMap, nil
}

// parseSOpsTargetParams 解析标准运维参数
func (svc *lbSvc) parseSOpsTargetParams(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	ruleQueryList []cslb.TargetGroupRuleQueryItem) (map[int][]string, error) {

	tgIDsMap := make(map[int][]string)
	for index, item := range ruleQueryList {
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
		rsIpTypeTgIDs, err = svc.parseSOpsRsIpAndRsTypeForTgIDs(kt, accountID, item.RsIP, item.RsType)
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
		vIpPortTgIDs, err = svc.parseSOpsVipAndVportForTgIDs(kt, accountID, vendor, item.Vip, item.VPort, item.Region)
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

		// 当前行条件没有能匹配到的目标组
		tgIDsItem = slice.Unique(tgIDsItem)
		if len(tgIDsItem) == 0 {
			return nil, fmt.Errorf("no matching target groups were found for line %d", index+1)
		}

		// 分别记录每一行条件查询出的目标组ID列表
		tgIDsMap[index] = tgIDsItem
		index++
	}
	logs.Infof("parse sops target params success, ruleQueryList: %+v, tgIDsMap: %+v, rid: %s",
		ruleQueryList, tgIDsMap, kt.Rid)

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

	tgIDs := make([]string, 0)
	urlRuleReq := &core.ListReq{
		Filter: urlRuleFilter,
		Page:   core.NewDefaultBasePage(),
	}
	for {
		urlRuleResult, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleReq)
		if err != nil {
			logs.Errorf("list url rule failed, req: %+v, err: %v, rid: %s", urlRuleReq, err, kt.Rid)
			return nil, err
		}
		// 记录urlRule对应的目标组ID
		for _, ruleItem := range urlRuleResult.Details {
			if len(ruleItem.TargetGroupID) == 0 {
				continue
			}
			tgIDs = append(tgIDs, ruleItem.TargetGroupID)
		}

		if uint(len(urlRuleResult.Details)) < core.DefaultMaxPageLimit {
			break
		}
		urlRuleReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return slice.Unique(tgIDs), nil
}

// parseSOpsRsIpAndRsTypeForTgIDs 根据RsIp和RsType查询Target，获取对应的目标组ID
func (svc *lbSvc) parseSOpsRsIpAndRsTypeForTgIDs(kt *kit.Kit, accountID string,
	rsIPList []string, rsType string) ([]string, error) {

	if len(rsIPList) == 0 || len(rsType) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}

	tgIDs := make([]string, 0)
	for _, rsIP := range rsIPList {
		// 查询出对应的目标
		filter := tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("inst_type", rsType),
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
func (svc *lbSvc) parseSOpsVipAndVportForTgIDs(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	vip []string, vport []int, region string) ([]string, error) {

	if len(vip) == 0 && len(vport) == 0 {
		// 没有对应的筛选条件，表现为不筛选
		return nil, nil
	}
	lbIDs := make([]string, 0)
	tgIDs := make([]string, 0)
	if len(vip) != 0 {
		// 若有vip筛选条件，则查询符合的负载均衡列表
		for _, vip := range vip {
			lbReq := &core.ListReq{
				Filter: tools.ExpressionAnd(
					tools.RuleEqual("vendor", vendor),
					tools.RuleEqual("account_id", accountID),
					tools.RuleEqual("region", region),
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
		tools.RuleEqual("vendor", vendor),
	)
	if len(lbIDs) != 0 {
		lblFilter.Rules = append(lblFilter.Rules, tools.RuleIn("lb_id", lbIDs))
	}
	if len(vport) != 0 {
		lblFilter.Rules = append(lblFilter.Rules, tools.RuleIn("port", vport))
	}

	lblIDs := make([]string, 0)
	lblReq := &core.ListReq{
		Filter: lblFilter,
		Page:   core.NewDefaultBasePage(),
	}
	for {
		listLblResult, err := svc.client.DataService().Global.LoadBalancer.ListListener(kt, lblReq)
		if err != nil {
			return nil, err
		}
		lblIDs = append(lblIDs, slice.Map(listLblResult.Details, func(lbl corelb.BaseListener) string {
			return lbl.ID
		})...)

		if uint(len(listLblResult.Details)) < core.DefaultMaxPageLimit {
			break
		}
		lblReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	if len(lblIDs) == 0 {
		return tgIDs, nil
	}

	// 查询符合的监听器与目标组绑定关系的列表
	ids, err := svc.listTargetGroupListenerRel(kt, lblIDs)
	if err != nil {
		logs.Errorf("list target group listener rel failed, req: %+v, err: %v, rid: %s", lblIDs, err, kt.Rid)
		return nil, err
	}
	tgIDs = append(tgIDs, ids...)
	return slice.Unique(tgIDs), nil
}

// 查询符合的监听器与目标组绑定关系的列表
func (svc *lbSvc) listTargetGroupListenerRel(kt *kit.Kit, lblIDs []string) ([]string, error) {
	lblRuleReq := &core.ListReq{
		Fields: []string{"target_group_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleIn("lbl_id", lblIDs),
			tools.RuleEqual("binding_status", enumor.SuccessBindingStatus),
		),
		Page: core.NewDefaultBasePage(),
	}
	tgIDs := make([]string, 0)
	for {
		lblRuleListResult, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(
			kt, lblRuleReq)
		if err != nil {
			logs.Errorf("list target group listener rel failed, req: %+v, err: %v, rid: %s",
				lblRuleReq, err, kt.Rid)
			return nil, err
		}
		for _, ruleRelItem := range lblRuleListResult.Details {
			if len(ruleRelItem.TargetGroupID) == 0 {
				continue
			}
			tgIDs = append(tgIDs, ruleRelItem.TargetGroupID)
		}

		if uint(len(lblRuleListResult.Details)) < core.DefaultMaxPageLimit {
			break
		}
		lblRuleReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return tgIDs, nil
}

// parseSOpsTargetParamsForRsOnline 解析标准运维参数-RS上线专属
func (svc *lbSvc) parseSOpsTargetParamsForRsOnline(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	tgQueryList []cslb.TargetGroupQueryItemForRsOnline) (map[int][]string, error) {

	tgIDsMap := make(map[int][]string)
	for index, item := range tgQueryList {
		// 原1.0逻辑
		// 1.根据VIP和RSIP获取到clb列表

		// 2.获取每个clb的listener列表（按照protocol和vport筛选）

		// 3.如果是七层，则按照筛选条件（rsip，rstype，domain，url）筛选 当前监听器下 的 所有规则

		// 4.如果是四层，则按照筛选条件（rsip，rstype）筛选 当前监听器

		// 5.转换为对应的TargetGroup

		// 优化逻辑
		// 1.Protocol、Domain、URL筛选出一批TargetGroup
		protoDomainUrlTgIDs, err := svc.cvtSOpsInfoForTgIDs(
			kt, accountID, vendor, item.Protocol, item.Domain, item.Url)
		if err != nil {
			logs.Errorf("parse protocol and domain and url for target group failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, err
		}

		// 2.VIP、VPort筛选出一批TargetGroup
		vIpPortTgIDs, err := svc.parseSOpsVipAndVportForTgIDs(kt, accountID, vendor, item.Vip, item.VPort, item.Region)
		if err != nil {
			logs.Errorf("parse vip and vport for target group failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, err
		}

		// 3.RSIP、RSTYPE直接查询到一批TargetGroup
		rsIpTypeTgIDs, err := svc.parseSOpsRsIpAndRsTypeForTgIDs(kt, accountID, item.RsIP, item.RsType)
		if err != nil {
			logs.Errorf("parse rsip and rstype for target group failed, accountID: %s, item: %+v, err: %v, rid: %s",
				accountID, item, err, kt.Rid)
			return nil, err
		}

		// 按照情况取交集
		tgIDsItem := protoDomainUrlTgIDs
		if vIpPortTgIDs != nil {
			tgIDsItem = slice.Intersection(tgIDsItem, vIpPortTgIDs)
		}
		if rsIpTypeTgIDs != nil {
			tgIDsItem = slice.Intersection(tgIDsItem, rsIpTypeTgIDs)
		}

		// 当前行条件没有能匹配到的目标组
		tgIDsItem = slice.Unique(tgIDsItem)
		if len(tgIDsItem) == 0 {
			return nil, fmt.Errorf("no matching target groups were found for line %d", index+1)
		}

		// 分别记录每一行条件查询出的目标组ID列表
		tgIDsMap[index] = tgIDsItem
		index++
	}
	logs.Infof("parse sops target params for rs online success, tgQueryList: %+v, tgIDsMap: %+v", tgQueryList, tgIDsMap)

	return tgIDsMap, nil
}

// cvtSOpsInfoForTgIDs 根据Protocol、Domain、URL查询UrlRule，获取对应的目标组ID
func (svc *lbSvc) cvtSOpsInfoForTgIDs(kt *kit.Kit, accountID string, vendor enumor.Vendor,
	protocol enumor.ProtocolType, domain, url []string) ([]string, error) {
	// 查询账号对应的Listener
	lbIDs := make([]string, 0)
	listenerReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID)),
		Page: core.NewDefaultBasePage(),
	}
	switch vendor {
	case enumor.TCloud:
		for {
			listenerResult, err := svc.client.DataService().TCloud.LoadBalancer.ListListener(kt, listenerReq)
			if err != nil {
				logs.Errorf("list url rule failed, req: %+v, err: %v", listenerReq, err)
				return nil, err
			}
			for _, listener := range listenerResult.Details {
				lbIDs = append(lbIDs, listener.ID)
			}

			if uint(len(listenerResult.Details)) < core.DefaultMaxPageLimit {
				break
			}
			listenerReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	// 筛选查询urlRule
	urlRuleFilter := tools.ExpressionAnd(
		tools.RuleIn("lb_id", lbIDs),
	)
	if protocol.IsLayer7Protocol() {
		urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleEqual("rule_type", enumor.Layer7RuleType))
		if len(domain) != 0 && !equalsAll(domain[0]) {
			urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleIn("domain", domain))
		}
		if len(url) != 0 && !equalsAll(url[0]) {
			urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleIn("url", url))
		}
	} else if protocol.IsLayer4Protocol() {
		urlRuleFilter.Rules = append(urlRuleFilter.Rules, tools.RuleEqual("rule_type", enumor.Layer4RuleType))
	} else {
		return nil, fmt.Errorf("protocol: %s not support", protocol)
	}

	var urlRuleSlice []corelb.TCloudLbUrlRule
	var err error
	urlRuleReq := &core.ListReq{
		Filter: urlRuleFilter,
		Page:   core.NewDefaultBasePage(),
	}
	switch vendor {
	case enumor.TCloud:
		urlRuleSlice, err = svc.listURLRule(kt, urlRuleReq)
		if err != nil {
			logs.Errorf("list url rule failed, req: %+v, err: %v, rid: %s", urlRuleReq, err, kt.Rid)
			return nil, err
		}
	}

	// 记录urlRule对应的目标组ID
	var tgIDs []string
	switch vendor {
	case enumor.TCloud:
		tgIDs, err = svc.parseTargetGroupIDs(kt, protocol, urlRuleSlice)
		if err != nil {
			logs.Errorf("parse target group ids failed, urlRuleSlice: %+v, err: %v, rid: %s",
				urlRuleSlice, err, kt.Rid)
			return nil, err
		}
	}
	return slice.Unique(tgIDs), nil
}

func (svc *lbSvc) listURLRule(kt *kit.Kit, urlRuleReq *core.ListReq) ([]corelb.TCloudLbUrlRule, error) {
	urlRuleSlice := make([]corelb.TCloudLbUrlRule, 0)
	for {
		urlRuleResult, err := svc.client.DataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleReq)
		if err != nil {
			logs.Errorf("list url rule failed, req: %+v, err: %v, rid: %s", urlRuleReq, err, kt.Rid)
			return nil, err
		}
		for _, urlRule := range urlRuleResult.Details {
			urlRuleSlice = append(urlRuleSlice, urlRule)
		}

		if uint(len(urlRuleResult.Details)) < core.DefaultMaxPageLimit {
			break
		}
		urlRuleReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return urlRuleSlice, nil
}

func (svc *lbSvc) parseTargetGroupIDs(kt *kit.Kit, protocol enumor.ProtocolType,
	urlRuleSlice []corelb.TCloudLbUrlRule) ([]string, error) {

	tgIDs := make([]string, 0)
	for _, ruleItem := range urlRuleSlice {
		if len(ruleItem.LblID) == 0 || len(ruleItem.TargetGroupID) == 0 {
			continue
		}
		lblResult, err := svc.client.DataService().TCloud.LoadBalancer.GetListener(kt, ruleItem.LblID)
		if err != nil {
			logs.Errorf("get listener failed, listenerID: %s, err: %v, rid: %s", ruleItem.LblID, err, kt.Rid)
			return nil, err
		}
		if lblResult == nil {
			continue
		}
		targetGroupResult, err := svc.client.DataService().TCloud.LoadBalancer.GetTargetGroup(kt, ruleItem.TargetGroupID)
		if err != nil {
			logs.Errorf("get target group failed, targetGroupID: %s, err: %v, rid: %s",
				ruleItem.TargetGroupID, err, kt.Rid)
			return nil, err
		}
		if targetGroupResult == nil {
			continue
		}

		if lblResult.Protocol != targetGroupResult.Protocol {
			return nil, fmt.Errorf("listener and target group protocols are different，listenrID: %s, targetGroupID: %s",
				ruleItem.LblID, ruleItem.TargetGroupID)
		}
		if protocol != targetGroupResult.Protocol {
			continue
		}
		tgIDs = append(tgIDs, ruleItem.TargetGroupID)
	}
	return slice.Unique(tgIDs), nil

}
