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
	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

// ModifyWeightRecord 基于 ModifyWeightRawInput 根据VIPs和VPorts进行记录的拆分后的结果,
// 结果表现为 一个listener以及一组RSInfo
type ModifyWeightRecord struct {
	// 本次批量绑定RS的操作类型
	Action enumor.BatchOperationActionType `json:"action"`

	ListenerName string              `json:"name"`
	Protocol     enumor.ProtocolType `json:"protocol"`

	IPDomainType string `json:"-"`
	VIP          string `json:"-"`
	VPorts       []int  `json:"ports"`
	HaveEndPort  bool   `json:"-"` // 是否是端口端

	Domain  string `json:"domain"` // 域名
	URLPath string `json:"url"`    // URL路径

	RSIPs     []string        `json:"-"`
	RSPorts   []int           `json:"-"`
	OldWeight []int           `json:"-"`
	Weights   []int           `json:"-"`
	RSInfos   []*RSUpdateInfo `json:"rs_infos"` // 后端实例信息

	// listener、rule、targetGroup 信息
	ListenerID    string `json:"-"`
	RuleID        string `json:"-"`
	TargetGroupID string `json:"-"`
}

// CheckWithDataService 依赖DB的校验逻辑
func (r *ModifyWeightRecord) CheckWithDataService(kt *kit.Kit, client *dataservice.Client,
	bkBizID int64) []*dataproto.BatchOperationValidateError {

	lb, err := r.GetLoadBalancer(kt, client, bkBizID)
	if err != nil {
		return []*dataproto.BatchOperationValidateError{{Reason: fmt.Sprintf("%s %v", r.GetKey(), err)}}
	}

	if err = r.loadDataFromDB(kt, client, lb); err != nil {
		return []*dataproto.BatchOperationValidateError{{Reason: fmt.Sprintf("%s %v", r.GetKey(), err)}}
	}

	errList := make([]*dataproto.BatchOperationValidateError, 0)
	for _, info := range r.RSInfos {
		err := info.ValidateOldWeight(kt, client.Global.LoadBalancer, r.TargetGroupID)
		if err != nil {
			errList = append(errList, &dataproto.BatchOperationValidateError{Reason: fmt.Sprintf("%s %v", r.GetKey(), err)})
		}
	}
	return errList
}

func (r *ModifyWeightRecord) loadDataFromDB(kt *kit.Kit, client *dataservice.Client, lb *loadbalancer.BaseLoadBalancer) error {
	var err error
	r.ListenerID, err = r.getListenerID(kt, client.Global.LoadBalancer, lb.ID)
	if err != nil {
		return err
	}

	r.RuleID, err = r.getListenerRuleID(kt, client, lb.ID, r.ListenerID, lb.Vendor)
	if err != nil {
		return err
	}

	r.TargetGroupID, err = r.getTargetGroupID(kt, client, lb.ID, r.RuleID)
	if err != nil {
		return err
	}
	return nil
}

// GetTargets 获取将要修改的Target信息
func (r *ModifyWeightRecord) GetTargets(kt *kit.Kit, client *dataservice.Client,
	lb *loadbalancer.BaseLoadBalancer) ([]*dataproto.TargetBaseReq, error) {

	if err := r.loadDataFromDB(kt, client, lb); err != nil {
		return nil, err
	}

	rsList := make([]*dataproto.TargetBaseReq, 0)
	for _, info := range r.RSInfos {
		target, err := info.GetTarget(kt, client.Global.LoadBalancer, r.TargetGroupID, lb.AccountID)
		if err != nil {
			return nil, err
		}
		rsList = append(rsList, target)
	}

	return rsList, nil
}

func (r *ModifyWeightRecord) getTargetGroupID(kt *kit.Kit, client *dataservice.Client,
	lbID, ruleID string) (string, error) {

	rel, err := client.Global.LoadBalancer.ListTargetGroupListenerRel(kt, &core.ListReq{
		Fields: []string{"target_group_id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("listener_rule_id", ruleID),
		),
	})
	if err != nil {
		return "", err
	}

	if len(rel.Details) == 0 {
		return "", fmt.Errorf("target group not found")
	}
	return rel.Details[0].TargetGroupID, nil
}

func (r *ModifyWeightRecord) getListenerRuleID(kt *kit.Kit, client *dataservice.Client,
	lbID, lblID string, vendor enumor.Vendor) (string, error) {

	rules := []*filter.AtomRule{
		{
			Field: "lb_id",
			Op:    filter.Equal.Factory(),
			Value: lbID,
		},
		{
			Field: "lbl_id",
			Op:    filter.Equal.Factory(),
			Value: lblID,
		},
	}

	if r.Protocol.IsLayer7Protocol() {
		rules = append(rules,
			tools.RuleEqual("url", r.URLPath),
			tools.RuleEqual("domain", r.Domain),
		)
	}

	switch vendor {
	case enumor.TCloud:
		rule, err := client.TCloud.LoadBalancer.ListUrlRule(kt,
			&core.ListReq{
				Page:   core.NewDefaultBasePage(),
				Filter: tools.ExpressionAnd(rules...),
			})
		if err != nil {
			return "", err
		}
		if len(rule.Details) == 0 {
			return "", fmt.Errorf("listener rule not found")
		}
		return rule.Details[0].ID, nil
	default:
		return "", fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (r *ModifyWeightRecord) getListenerID(kt *kit.Kit, client *global.LoadBalancerClient,
	lbID string) (string, error) {

	listeners, err := client.ListListener(kt, &core.ListReq{
		Fields: []string{"id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("protocol", r.Protocol),
			tools.RuleEqual("port", r.VPorts[0]),
		),
	})
	if err != nil {
		return "", err
	}

	if listeners == nil || len(listeners.Details) == 0 {
		return "", nil
	}

	return listeners.Details[0].ID, nil
}

// GetLoadBalancer 获取CLB信息, 依赖ModifyWeightRecord.VIP等信息
func (r *ModifyWeightRecord) GetLoadBalancer(kt *kit.Kit, client *dataservice.Client, bkBizID int64) (
	*loadbalancer.BaseLoadBalancer, error) {

	var expression *filter.Expression
	switch r.IPDomainType {
	case ipDomainTypeIPv4:
		expression = tools.ExpressionOr(
			tools.RuleJSONContains("private_ipv4_addresses", r.VIP),
			tools.RuleJSONContains("public_ipv4_addresses", r.VIP),
		)
	case ipDomainTypeIPv6:
		expression = tools.ExpressionOr(
			tools.RuleJSONContains("private_ipv6_addresses", r.VIP),
			tools.RuleJSONContains("public_ipv6_addresses", r.VIP),
		)
	case ipDomainTypeDomain:
		expression = tools.ExpressionAnd(
			tools.RuleEqual("domain", r.VIP),
		)
	}

	expr, err := tools.And(expression, tools.RuleEqual("bk_biz_id", bkBizID))
	if err != nil {
		return nil, err
	}
	balancers, err := client.Global.LoadBalancer.ListLoadBalancer(kt, &core.ListReq{
		Page:   core.NewDefaultBasePage(),
		Filter: expr,
	})
	if err != nil {
		return nil, err
	}
	if len(balancers.Details) == 0 {
		return nil, fmt.Errorf("no load balancer found for VIP: %s", r.VIP)
	}
	if len(balancers.Details) > 1 {
		return nil, fmt.Errorf("more than one load balancer found for VIP: %s", r.VIP)
	}

	return &balancers.Details[0], nil
}

// GetKey 获取当前记录的标签
func (r *ModifyWeightRecord) GetKey() string {
	var port string
	if r.HaveEndPort {
		port = fmt.Sprintf("[%d,%d]", r.VPorts[0], r.VPorts[len(r.VPorts)-1])
	} else {
		port = fmt.Sprintf("%d", r.VPorts[0])
	}
	if len(r.Domain) > 0 {
		return fmt.Sprintf("%s %s:%s %s %s", r.VIP, r.Protocol, port, r.Domain, r.URLPath)
	}
	return fmt.Sprintf("%s %s:%s", r.VIP, r.Protocol, port)
}

// Validate 无需外部依赖的校验逻辑
func (r *ModifyWeightRecord) Validate() error {
	err := r.validateListenerName()
	if err != nil {
		return err
	}

	if err = r.validateIpDomainType(); err != nil {
		return err
	}

	if err = r.validateProtocol(); err != nil {
		return err
	}

	if err = r.validateVIPsAndVPorts(); err != nil {
		return err
	}

	if err = r.validateRS(); err != nil {
		return err
	}

	if err = r.validateRSInfoDuplicate(); err != nil {
		return err
	}

	return nil
}

func (r *ModifyWeightRecord) validateListenerName() error {
	if len(r.ListenerName) == 0 || r.ListenerName == "" {
		return fmt.Errorf("listener name should not be empty")
	}
	if len(r.ListenerName) > 255 {
		return fmt.Errorf("listener name is too long")
	}
	return nil
}

func (r *ModifyWeightRecord) validateIpDomainType() error {
	if r.IPDomainType == ipDomainTypeDomain {
		return nil
	}
	version, err := getIPVersion(r.VIP)
	if err != nil {
		return err
	}

	if version == enumor.Ipv4 && r.IPDomainType != ipDomainTypeIPv4 {
		return fmt.Errorf("ip version is ipv4, but ip domain type got %s", r.IPDomainType)
	}

	if version == enumor.Ipv6 && r.IPDomainType != ipDomainTypeIPv6 {
		return fmt.Errorf("ip version is ipv6, but ip domain type got %s", r.IPDomainType)
	}

	return nil
}

func (r *ModifyWeightRecord) validateProtocol() error {
	if _, ok := supportedProtocols[r.Protocol]; !ok {
		return fmt.Errorf("unsupported protocol: %s", r.Protocol)
	}

	return nil
}

func (r *ModifyWeightRecord) validateRS() error {
	err := r.validateWeight()
	if err != nil {
		return err
	}

	if len(r.RSIPs) == 0 || len(r.RSPorts) == 0 || len(r.Weights) == 0 {
		return fmt.Errorf("RSIPs, RSPorts and NewWeight cannot be empty")
	}

	if r.HaveEndPort {
		if len(r.RSPorts) != 2 {
			return fmt.Errorf("port range should have two ports")
		}

		if r.VPorts[1]-r.VPorts[0] != r.RSPorts[1]-r.RSPorts[0] {
			return fmt.Errorf("port range should have the same length")
		}

		if len(r.RSIPs) != 1 || len(r.Weights) != 1 {
			return fmt.Errorf("RSIPs and NewWeight should have only one element")
		}

		for _, port := range r.RSPorts {
			if !validPort(port) {
				return fmt.Errorf("invalid RSPort: %d", port)
			}
		}

		r.RSInfos = append(r.RSInfos, &RSUpdateInfo{
			IP:        r.RSIPs[0],
			Port:      r.RSPorts[0],
			EndPort:   r.RSPorts[1],
			NewWeight: r.Weights[0],
			OldWeight: r.OldWeight[0],
		})
		return nil
	}

	if len(r.RSPorts) > 1 && len(r.RSPorts) != len(r.RSIPs) {
		return fmt.Errorf("the number of RSPorts and RSIPs should be equal or 1")
	}

	if len(r.Weights) > 1 && len(r.Weights) != len(r.RSIPs) {
		return fmt.Errorf("the number of NewWeight and RSIPs should be equal or 1")
	}

	// validate port
	for _, port := range r.RSPorts {
		if !validPort(port) {
			return fmt.Errorf("invalid RSPort: %d", port)
		}
	}

	/** 数据补全
	input: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80] NewWeight: [1 1]
	output: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80 80] NewWeight: [1 1]

	input: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80 80] NewWeight: [1]
	output: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80 80] NewWeight: [1 1]
	*/
	for len(r.RSPorts) < len(r.RSIPs) {
		r.RSPorts = append(r.RSPorts, r.RSPorts[0])
	}

	for len(r.Weights) < len(r.RSIPs) {
		r.Weights = append(r.Weights, r.Weights[0])
	}

	for len(r.OldWeight) < len(r.RSIPs) {
		r.OldWeight = append(r.OldWeight, r.OldWeight[0])
	}

	for i := 0; i < len(r.RSIPs); i++ {
		r.RSInfos = append(r.RSInfos, &RSUpdateInfo{
			IP:        r.RSIPs[i],
			Port:      r.RSPorts[i],
			NewWeight: r.Weights[i],
			OldWeight: r.OldWeight[i],
		})
	}
	return nil
}

func (r *ModifyWeightRecord) validateRSInfoDuplicate() error {
	m := map[string]struct{}{}
	for _, info := range r.RSInfos {
		flag := fmt.Sprintf("%s:%d", info.IP, info.Port)
		if _, ok := m[flag]; ok {
			return fmt.Errorf("duplicate RS: %s", flag)
		}
		m[flag] = struct{}{}
	}
	return nil
}

func (r *ModifyWeightRecord) validateWeight() error {
	if len(r.Weights) != len(r.OldWeight) {
		return fmt.Errorf("the number of NewWeight and OldWeight should be equal")
	}

	for _, weight := range r.Weights {
		if weight < 0 || weight > 100 {
			return fmt.Errorf("invalid weight value: %d", weight)
		}
	}

	for _, weight := range r.OldWeight {
		if weight < 0 || weight > 100 {
			return fmt.Errorf("invalid weight value: %d", weight)
		}
	}

	return nil
}

func (r *ModifyWeightRecord) validateVIPsAndVPorts() error {
	if len(r.VIP) == 0 || len(r.VPorts) == 0 {
		return fmt.Errorf("VIP and VPorts cannot be empty")
	}

	if r.HaveEndPort && len(r.VPorts) != 2 {
		return fmt.Errorf("port range should have two ports")
	}

	if !r.HaveEndPort && len(r.VPorts) != 1 {
		return fmt.Errorf("parse VPorts error, should be 1 or 2")
	}

	for _, port := range r.VPorts {
		if !validPort(port) {
			return fmt.Errorf("invalid VPort: %d", port)
		}
	}

	return nil
}
