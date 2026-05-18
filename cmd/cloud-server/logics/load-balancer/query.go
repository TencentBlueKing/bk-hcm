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
	"encoding/json"
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"
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

func getURLRule(kt *kit.Kit, cli *dataservice.Client, vendor enumor.Vendor,
	lbCloudID, listenerCloudID, domain, url string) (*corelb.TCloudLbUrlRule, error) {

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
	accountID string, bkBizID int64, cloudIDs []string) (map[string]corelb.LoadBalancerRaw, error) {

	result := make(map[string]corelb.LoadBalancerRaw, len(cloudIDs))
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
		resp, err := cli.Global.LoadBalancer.ListLoadBalancerRaw(kt, req)
		if err != nil {
			logs.Errorf("list load balancer failed, req: %v, err: %v, rid: %s", req, err, kt.Rid)
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

func getTargetGroupID(kt *kit.Kit, cli *dataservice.Client, lbID string, ruleCloudID string) (string, error) {
	listReq := &core.ListReq{
		Fields: []string{"target_group_id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
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

func getTargetGroupByRuleCloudIDs(kt *kit.Kit, cli *dataservice.Client, ruleCloudIDs []string) (
	map[string]string, error) {

	result := make(map[string]string, len(ruleCloudIDs))
	for _, batch := range slice.Split(ruleCloudIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Fields: []string{"target_group_id", "cloud_listener_rule_id"},
			Page:   core.NewDefaultBasePage(),
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_listener_rule_id", batch),
			),
		}
		rel, err := cli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, listReq)
		if err != nil {
			logs.Errorf("list target group listener rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, detail := range rel.Details {
			result[detail.CloudListenerRuleID] = detail.TargetGroupID
		}
	}
	return result, nil
}

func getCvm(kt *kit.Kit, cli *dataservice.Client, ip string,
	vendor enumor.Vendor, bkBizID int64, accountID string, cloudVPCs []string) (*corecvm.BaseCvm, error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
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
		},
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

// getCvmWithoutVpc 不指定VPC查询主机
func getCvmWithoutVpc(kt *kit.Kit, cli *dataservice.Client, ip string, vendor enumor.Vendor, bkBizID int64,
	accountID string) ([]corecvm.BaseCvm, error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.ExpressionOr(
				tools.RuleJSONContains("private_ipv4_addresses", ip),
				tools.RuleJSONContains("private_ipv6_addresses", ip),
				tools.RuleJSONContains("public_ipv4_addresses", ip),
				tools.RuleJSONContains("public_ipv6_addresses", ip),
			),
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("account_id", accountID),
		},
	}
	listReq := &core.ListReq{
		Filter: expr,
		Page:   core.NewDefaultBasePage(),
	}
	cvms, err := cli.Global.Cvm.ListCvm(kt, listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return cvms.Details, nil
}

// cvmIPBatchSize 单次按 IP 批量查询 CVM 的批次大小。
// JSON_OVERLAPS 走不了索引，单次评估的元素数（批次大小 × OR 字段数）越大越容易触发 data-service 超时，
// 故此处取小于 core.DefaultMaxPageLimit 的值，平衡查询次数与单次执行耗时。
const cvmIPBatchSize = 100

// buildBatchGetCvmWithoutVpcExpr 按指定的 IP 字段集合构建查询条件。
// ipFields 仅包含与 partIPs 版本匹配的字段（v4 或 v6），避免无意义的 JSON_OVERLAPS。
func buildBatchGetCvmWithoutVpcExpr(partIPs []string, ipFields []string, vendor enumor.Vendor, bkBizID int64,
	accountID string) *filter.Expression {

	ipRules := make([]*filter.AtomRule, 0, len(ipFields))
	for _, field := range ipFields {
		ipRules = append(ipRules, tools.RuleJsonOverlaps(field, partIPs))
	}

	return &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.ExpressionOr(ipRules...),
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("account_id", accountID),
		},
	}
}

// splitIPsByVersion 按 IP 版本拆分，非法 IP 直接丢弃（上层对未匹配 IP 已有兜底报错逻辑）
func splitIPsByVersion(kt *kit.Kit, ips []string) (v4IPs, v6IPs []string) {
	v4IPs = make([]string, 0, len(ips))
	v6IPs = make([]string, 0, len(ips))
	for _, ip := range ips {
		switch {
		case cidr.IsIPv4(ip):
			v4IPs = append(v4IPs, ip)
		case cidr.IsIPv6(ip):
			v6IPs = append(v6IPs, ip)
		default:
			logs.Warnf("skip invalid rs ip: %s, rid: %s", ip, kt.Rid)
		}
	}
	return v4IPs, v6IPs
}

// batchGetCvmWithoutVpc 不指定VPC批量查询主机
func batchGetCvmWithoutVpc(kt *kit.Kit, cli *dataservice.Client, ips []string, vendor enumor.Vendor, bkBizID int64,
	accountID string) ([]corecvm.BaseCvm, error) {

	// 同一个 RS IP 可能因多端口/多监听器绑定而重复，先去重避免无意义的重复查询
	ips = slice.Unique(ips)

	// 按 IP 版本拆分，使每次查询的 OR 条件从 4 个减为 2 个，降低 data-service 单次查询负载
	v4IPs, v6IPs := splitIPsByVersion(kt, ips)

	cvmList := make([]corecvm.BaseCvm, 0)
	v4Cvms, err := batchListCvmByIPs(kt, cli, v4IPs,
		[]string{"private_ipv4_addresses", "public_ipv4_addresses"}, vendor, bkBizID, accountID)
	if err != nil {
		logs.Errorf("batch list cvm by ipv4 failed, bkBizID: %d, err: %v, rid: %s", bkBizID, err, kt.Rid)
		return nil, err
	}
	cvmList = append(cvmList, v4Cvms...)

	v6Cvms, err := batchListCvmByIPs(kt, cli, v6IPs,
		[]string{"private_ipv6_addresses", "public_ipv6_addresses"}, vendor, bkBizID, accountID)
	if err != nil {
		logs.Errorf("batch list cvm by ipv6 failed, bkBizID: %d, err: %v, rid: %s", bkBizID, err, kt.Rid)
		return nil, err
	}
	cvmList = append(cvmList, v6Cvms...)

	return cvmList, nil
}

// batchListCvmByIPs 按 cvmIPBatchSize 分批查询匹配指定 IP 列表的 CVM。
func batchListCvmByIPs(kt *kit.Kit, cli *dataservice.Client, ips []string, ipFields []string,
	vendor enumor.Vendor, bkBizID int64, accountID string) ([]corecvm.BaseCvm, error) {

	if len(ips) == 0 {
		return nil, nil
	}

	cvmList := make([]corecvm.BaseCvm, 0)
	for _, partIPs := range slice.Split(ips, cvmIPBatchSize) {
		expr := buildBatchGetCvmWithoutVpcExpr(partIPs, ipFields, vendor, bkBizID, accountID)
		listReq := &core.ListReq{Filter: expr, Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit}}
		for {
			cvms, err := cli.Global.Cvm.ListCvm(kt, listReq)
			if err != nil {
				logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			cvmList = append(cvmList, cvms.Details...)
			if uint(len(cvms.Details)) < core.DefaultMaxPageLimit {
				break
			}
			listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}
	return cvmList, nil
}

// validateCvmExist 导入新RS前, 校验云主机是否存在
// 跨域1.0如果没找到相同的vpc下的主机，会进行报错
func validateCvmExist(kt *kit.Kit, dataServiceCli *dataservice.Client, rsIP string, lb corelb.LoadBalancerRaw,
	isCrossRegionV1, isCrossRegionV2 bool, targetCloudVpcID string, cvmList []corecvm.BaseCvm) (*corecvm.BaseCvm, error) {

	var cvm *corecvm.BaseCvm

	// 如果上层没有传入cvmList，则自己查询
	if len(cvmList) == 0 {
		var err error
		cvmList, err = getCvmWithoutVpc(kt, dataServiceCli, rsIP, lb.Vendor, lb.BkBizID, lb.AccountID)
		if err != nil {
			logs.Errorf("get cvm without vpc failed, ip: %s, err: %v, rid: %s", rsIP, err, kt.Rid)
			return nil, err
		}
	}

	// 从cvmList中筛选出包含rsIP的CVM
	matchedCvmList := make([]corecvm.BaseCvm, 0)
	for _, cvmItem := range cvmList {
		// 检查rsIP是否在CVM的任意IP地址中
		if slice.IsItemInSlice(cvmItem.PrivateIPv4Addresses, rsIP) ||
			slice.IsItemInSlice(cvmItem.PrivateIPv6Addresses, rsIP) ||
			slice.IsItemInSlice(cvmItem.PublicIPv4Addresses, rsIP) ||
			slice.IsItemInSlice(cvmItem.PublicIPv6Addresses, rsIP) {
			matchedCvmList = append(matchedCvmList, cvmItem)
		}
	}

	if len(matchedCvmList) == 0 {
		return nil, fmt.Errorf("rs(%s) not found", rsIP)
	}

	if isCrossRegionV2 {
		cvm = &matchedCvmList[0]
		return cvm, nil
	}

	cloudVpcIDs := []string{lb.CloudVpcID}
	if isCrossRegionV1 {
		cloudVpcIDs = append(cloudVpcIDs, targetCloudVpcID)
	}
	for _, one := range matchedCvmList {
		if len(slice.Intersection(cloudVpcIDs, one.CloudVpcIDs)) > 0 {
			return &one, nil
		}
	}

	cvmCloudIDs := slice.Map(matchedCvmList, corecvm.BaseCvm.GetCloudID)
	return nil, fmt.Errorf("VPC of %s is different from loadbalancer's VPC (%s)",
		strings.Join(cvmCloudIDs, ","), strings.Join(cloudVpcIDs, ","))
}

func parseSnapInfoTCloudLBExtension(kt *kit.Kit, raw json.RawMessage) (
	isCrossRegionV1, isCrossRegionV2 bool, targetCloudVpcID, lbTargetRegion string, err error) {

	extension := corelb.TCloudClbExtension{}
	err = json.Unmarshal(raw, &extension)
	if err != nil {
		logs.Errorf("fail parse lb extension for delete protection, err: %v, rid: %s", err, kt.Rid)
		return
	}

	isCrossRegionV1 = extension.SupportCrossRegionV1()
	isCrossRegionV2 = converter.PtrToVal(extension.SnatPro)
	targetCloudVpcID = converter.PtrToVal(extension.TargetCloudVpcID)
	lbTargetRegion = converter.PtrToVal(extension.TargetRegion)
	return
}
