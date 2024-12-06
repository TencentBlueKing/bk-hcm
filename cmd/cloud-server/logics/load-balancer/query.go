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

package lblogic

import (
	"fmt"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// ListLoadBalancerMap 批量获取负载均衡列表信息
func ListLoadBalancerMap(kt *kit.Kit, cli *dataservice.Client, lbIDs []string) (
	map[string]corelb.BaseLoadBalancer, error) {
	if len(lbIDs) == 0 {
		return nil, nil
	}

	clbReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", lbIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lbList, err := cli.Global.LoadBalancer.ListLoadBalancer(kt, clbReq)
	if err != nil {
		logs.Errorf("list load balancer failed, lbIDs: %v, err: %v, rid: %s", lbIDs, err, kt.Rid)
		return nil, err
	}

	lbMap := make(map[string]corelb.BaseLoadBalancer, len(lbList.Details))
	for _, lbItem := range lbList.Details {
		lbMap[lbItem.ID] = lbItem
	}

	return lbMap, nil
}

// GetListenerByID 根据监听器ID、业务ID获取监听器信息
func GetListenerByID(kt *kit.Kit, cli *dataservice.Client, lblID string) (corelb.BaseListener, error) {
	listenerInfo := corelb.BaseListener{}
	lblReq := &core.ListReq{
		Filter: tools.EqualExpression("id", lblID),
		Page:   core.NewDefaultBasePage(),
	}
	lblList, err := cli.Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		logs.Errorf("list listener by id failed, lblID: %s, err: %v, rid: %s", lblID, err, kt.Rid)
		return listenerInfo, err
	}
	if len(lblList.Details) == 0 {
		return listenerInfo, errf.Newf(errf.RecordNotFound, "listener_id: %s not found", lblID)
	}

	return lblList.Details[0], nil
}

func getListener(kt *kit.Kit, cli *dataservice.Client, accountID, lbCloudID string, protocol enumor.ProtocolType,
	port int, bkBizID int64, vendor enumor.Vendor) (*corelb.BaseListener, error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cloud_lb_id", lbCloudID),
			tools.RuleEqual("port", port),
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("protocol", protocol),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := cli.Global.LoadBalancer.ListListener(kt, req)
	if err != nil {
		logs.Errorf("list listener failed, port: %d, cloudLBID: %s, err: %v, rid: %s",
			port, lbCloudID, err, kt.Rid)
		return nil, err
	}
	if len(resp.Details) > 0 {
		return &resp.Details[0], nil
	}
	return nil, nil
}

func getURLRule(kt *kit.Kit, cli *dataservice.Client, vendor enumor.Vendor, lbCloudID, listenerCloudID, domain, url string) (
	*corelb.TCloudLbUrlRule, error) {

	switch vendor {
	case enumor.TCloud:
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("cloud_lb_id", lbCloudID),
				tools.RuleEqual("cloud_lbl_id", listenerCloudID),
				tools.RuleEqual("domain", domain),
				tools.RuleEqual("url", url),
			),
			Page: core.NewDefaultBasePage(),
		}
		rule, err := cli.TCloud.LoadBalancer.ListUrlRule(kt, req)
		if err != nil {
			logs.Errorf("list url rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(rule.Details) > 0 {
			return &rule.Details[0], nil
		}
	default:
		return nil, fmt.Errorf("vendor(%s) not support", vendor)
	}
	return nil, nil
}

func getLoadBalancersMapByCloudID(kt *kit.Kit, cli *dataservice.Client, vendor enumor.Vendor,
	accountID string, bkBizID int64, cloudIDs []string) (map[string]corelb.BaseLoadBalancer, error) {

	result := make(map[string]corelb.BaseLoadBalancer, len(cloudIDs))
	for _, ids := range slice.Split(cloudIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("account_id", accountID),
				tools.RuleEqual("bk_biz_id", bkBizID),
				tools.RuleIn("cloud_id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := cli.Global.LoadBalancer.ListLoadBalancer(kt, req)
		if err != nil {
			logs.Errorf("list load balancer failed, req: %v, error: %v, rid: %s", req, err, kt.Rid)
			return nil, err
		}
		for _, lb := range resp.Details {
			result[lb.CloudID] = lb
		}
	}
	return result, nil
}

func getTarget(kt *kit.Kit, cli *dataservice.Client, tgID, instID string, port int) (*corelb.BaseTarget, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("target_group_id", tgID),
			tools.RuleEqual("cloud_inst_id", instID),
			tools.RuleEqual("port", port),
		),
		Page: core.NewDefaultBasePage(),
	}
	targets, err := cli.Global.LoadBalancer.ListTarget(kt, listReq)
	if err != nil {
		logs.Errorf("list target failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(targets.Details) > 0 {
		return &targets.Details[0], nil
	}

	return nil, nil
}

func getTargetGroupID(kt *kit.Kit, cli *dataservice.Client, ruleCloudID string) (string, error) {
	listReq := &core.ListReq{
		Fields: []string{"target_group_id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cloud_listener_rule_id", ruleCloudID),
		),
	}
	rel, err := cli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, listReq)
	if err != nil {
		logs.Errorf("list target group listener rel failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(rel.Details) == 0 {
		return "", fmt.Errorf("target group not found")
	}
	return rel.Details[0].TargetGroupID, nil
}

func getCvm(kt *kit.Kit, cli *dataservice.Client, ip string,
	vendor enumor.Vendor, bkBizID int64, accountID string, cloudVPCs []string) (*corecvm.BaseCvm, error) {

	expr, err := tools.And(
		tools.ExpressionOr(
			tools.RuleJSONContains("private_ipv4_addresses", ip),
			tools.RuleJSONContains("private_ipv6_addresses", ip),
			tools.RuleJSONContains("public_ipv4_addresses", ip),
			tools.RuleJSONContains("public_ipv6_addresses", ip),
		),
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("bk_biz_id", bkBizID),
		tools.RuleEqual("account_id", accountID),
		tools.RuleJsonOverlaps("cloud_vpc_ids", cloudVPCs),
	)
	if err != nil {
		logs.Errorf("failed to create expression, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	listReq := &core.ListReq{
		Filter: expr,
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	cvms, err := cli.Global.Cvm.ListCvm(kt, listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(cvms.Details) > 0 {
		return &cvms.Details[0], nil
	}
	return nil, nil
}

func getTCloudLoadBalancer(kt *kit.Kit, cli *dataservice.Client, lbID string) (
	*corelb.LoadBalancer[corelb.TCloudClbExtension], error) {

	lb, err := cli.TCloud.LoadBalancer.Get(kt, lbID)
	if err != nil {
		logs.Errorf("get tcloud load balancer failed, lb(%s), err: %v, rid: %s", lbID, err, kt.Rid)
		return nil, err
	}
	return lb, nil
}
