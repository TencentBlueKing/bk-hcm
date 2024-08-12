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
	"net"

	"hcm/pkg/api/core"
	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/client/data-service/tcloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

const (
	ipDomainTypeIPv4   = "IPv4"
	ipDomainTypeIPv6   = "IPv6"
	ipDomainTypeDomain = "Domain"
)

var (
	supportIPDomainType = map[string]string{
		"IPv4": ipDomainTypeIPv4,
		"IPv6": ipDomainTypeIPv6,
		"域名":   ipDomainTypeDomain,
	}

	// IPDomainTypeMap ip_domain_type Translation map
	IPDomainTypeMap = map[string]string{
		ipDomainTypeIPv4:   "IPv4",
		ipDomainTypeIPv6:   "IPv6",
		ipDomainTypeDomain: "域名",
	}

	tcpProtocolMap = map[enumor.ProtocolType]struct{}{
		enumor.TcpProtocol:    {},
		enumor.HttpsProtocol:  {},
		enumor.HttpProtocol:   {},
		enumor.TcpSslProtocol: {},
	}

	udpProtocolMap = map[enumor.ProtocolType]struct{}{
		enumor.UdpProtocol:  {},
		enumor.QuicProtocol: {},
	}
)

// BindRSRecord 基于 BindRSRawInput 根据VIPs和VPorts进行记录的拆分后的结果,
// 结果表现为 一个listener以及一组RSInfo
type BindRSRecord struct {
	// 本次批量绑定RS的操作类型
	Action enumor.BatchOperationActionType `json:"action"`

	ListenerName string              `json:"name"`
	Protocol     enumor.ProtocolType `json:"protocol"`

	IPDomainType string `json:"-"`
	VIP          string `json:"-"`
	VPorts       []int  `json:"ports"`
	HaveEndPort  bool   `json:"-"` // 是否是端口端

	Domain      string   `json:"domain"`         // 域名
	URLPath     string   `json:"url"`            // URL路径
	ServerCerts []string `json:"cert_cloud_ids"` // ref: pkg/api/core/cloud/load-balancer/tcloud.go:188
	ClientCert  string   `json:"ca_cloud_id"`    // 客户端证书

	InstType enumor.InstType `json:"-"` // 后端类型 CVM、ENI
	RSIPs    []string        `json:"-"`
	RSPorts  []int           `json:"-"`
	Weights  []int           `json:"-"`
	RSInfos  []*RSInfo       `json:"rs_infos"` // 后端实例信息

	Scheduler      string `json:"scheduler"`       // 均衡方式
	SessionExpired int64  `json:"session_expired"` // 会话保持时间，单位秒
	HealthCheck    bool   `json:"health_check"`    // 是否开启健康检查

	// 以及一些依赖外部数据的校验，不属于BindRSRecord.Validate()的职责，比如：
	// 判断RS是否存在，CLB是否存在等等，应该是一个别的行为

	ListenerID    string `json:"-"`
	RuleID        string `json:"-"`
	TargetGroupID string `json:"-"`
}

// CheckWithDataService 依赖DB的校验逻辑
func (l *BindRSRecord) CheckWithDataService(kt *kit.Kit, client *dataservice.Client,
	bkBizID int64) []*cloud.BatchOperationValidateError {

	errList := make([]*cloud.BatchOperationValidateError, 0)
	errList = append(errList, l.checkCert(kt, client, bkBizID)...)
	lb, err := l.GetLoadBalancer(kt, client, bkBizID)
	if err != nil {
		errList = append(errList, &cloud.BatchOperationValidateError{
			Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
		})
	}
	if lb == nil {
		return errList
	}

	if err = l.checkAction(kt, lb, client); err != nil {
		errList = append(errList, &cloud.BatchOperationValidateError{
			Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
		})
	}

	if err = l.checkPortConflict(kt, lb.ID, client); err != nil {
		errList = append(errList, &cloud.BatchOperationValidateError{
			Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
		})
	}

	errList = append(errList, l.checkRSInfos(kt, lb, client)...)

	errList = append(errList, l.checkRSDuplicate(kt, lb.ID, client)...)

	return errList
}

// checkRSInfos 校验RS，检查RS是否存在
func (l *BindRSRecord) checkRSInfos(kt *kit.Kit, lb *loadbalancer.BaseLoadBalancer,
	client *dataservice.Client) []*cloud.BatchOperationValidateError {

	errList := make([]*cloud.BatchOperationValidateError, 0)
	for _, rs := range l.RSInfos {
		err := rs.CheckTarget(kt, lb.Vendor, lb.BkBizID, client)
		if err != nil {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
			})
		}

		if l.Action == enumor.AppendRS {
			// 只有追加RS的时候，才会有target_group的实体，需要校验是否有已经存在的target
			err := l.LoadDataFromDB(kt, client, lb)
			if err != nil {
				errList = append(errList, &cloud.BatchOperationValidateError{
					Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
				})
				continue
			}

			err = rs.checkTargetAlreadyExist(kt, client, l.TargetGroupID)
			if err != nil {
				errList = append(errList, &cloud.BatchOperationValidateError{
					Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
				})
			}
		}
	}

	return errList
}

// LoadDataFromDB 仅用于追加RS的场景，从DB中加载listener_id, rule_id, target_group_id, 其余action调用都会报错.
func (l *BindRSRecord) LoadDataFromDB(kt *kit.Kit, client *dataservice.Client, lb *loadbalancer.BaseLoadBalancer) error {
	var err error
	l.ListenerID, err = l.GetListenerID(kt, client.Global.LoadBalancer, lb.ID)
	if err != nil {
		return err
	}

	l.RuleID, err = l.getListenerRuleID(kt, client, lb.ID, l.ListenerID, lb.Vendor)
	if err != nil {
		return err

	}
	l.TargetGroupID, err = l.getTargetGroupID(kt, client, lb.ID, l.RuleID)
	if err != nil {
		return err
	}
	return nil
}

// checkCert 校验证书是否存在
func (l *BindRSRecord) checkCert(kt *kit.Kit, client *dataservice.Client,
	bkBizID int64) []*cloud.BatchOperationValidateError {

	result := make([]*cloud.BatchOperationValidateError, 0)
	certList := make([]string, 0)
	if len(l.ServerCerts) > 0 {
		certList = append(certList, l.ServerCerts...)
	}
	if len(l.ClientCert) > 0 {
		certList = append(certList, l.ClientCert)
	}
	if len(certList) == 0 {
		return nil
	}

	expr, err := tools.And(
		tools.ContainersExpression("cloud_id", certList),
		tools.RuleEqual("bk_biz_id", bkBizID),
	)
	listReq := &core.ListReq{
		Page:   core.NewDefaultBasePage(),
		Filter: expr,
	}
	certs, err := client.Global.ListCert(kt, listReq)
	if err != nil {
		result = append(result, &cloud.BatchOperationValidateError{
			Reason: fmt.Sprintf("%s %v", l.GetKey(), err),
		})
		return result
	}

	certMap := make(map[string]struct{})
	for _, detail := range certs.Details {
		certMap[detail.CloudID] = struct{}{}
	}

	for _, cert := range certList {
		if _, ok := certMap[cert]; !ok {
			result = append(result, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s cert:%s not found", l.GetKey(), cert),
			})
		}
	}

	return result
}

// GetLoadBalancer 获取CLB信息
func (l *BindRSRecord) GetLoadBalancer(kt *kit.Kit, client *dataservice.Client, bkBizID int64) (
	*loadbalancer.BaseLoadBalancer, error) {

	var expression *filter.Expression
	switch l.IPDomainType {
	case ipDomainTypeIPv4:
		expression = tools.ExpressionOr(
			tools.RuleJSONContains("private_ipv4_addresses", l.VIP),
			tools.RuleJSONContains("public_ipv4_addresses", l.VIP),
		)
	case ipDomainTypeIPv6:
		expression = tools.ExpressionOr(
			tools.RuleJSONContains("private_ipv6_addresses", l.VIP),
			tools.RuleJSONContains("public_ipv6_addresses", l.VIP),
		)
	case ipDomainTypeDomain:
		expression = tools.ExpressionAnd(
			tools.RuleEqual("domain", l.VIP),
		)
	default:
		return nil, fmt.Errorf("unsupported ip_domain_type: %s", l.IPDomainType)
	}

	expr, err := tools.And(expression, tools.RuleEqual("bk_biz_id", bkBizID))
	if err != nil {
		return nil, err
	}
	listReq := &core.ListReq{
		Page:   core.NewDefaultBasePage(),
		Filter: expr,
	}
	balancers, err := client.Global.LoadBalancer.ListLoadBalancer(kt, listReq)
	if err != nil {
		return nil, err
	}
	if len(balancers.Details) == 0 {
		return nil, fmt.Errorf("no load balancer found for VIP: %s", l.VIP)
	}
	if len(balancers.Details) > 1 {
		return nil, fmt.Errorf("more than one load balancer found for VIP: %s", l.VIP)
	}

	return &balancers.Details[0], nil
}

func (l *BindRSRecord) getTargetGroupID(kt *kit.Kit, client *dataservice.Client,
	lbID, ruleID string) (string, error) {

	listReq := &core.ListReq{
		Fields: []string{"target_group_id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("listener_rule_id", ruleID),
		),
	}
	rel, err := client.Global.LoadBalancer.ListTargetGroupListenerRel(kt, listReq)
	if err != nil {
		return "", err
	}

	if len(rel.Details) == 0 {
		return "", fmt.Errorf("target group not found")
	}
	return rel.Details[0].TargetGroupID, nil
}

func (l *BindRSRecord) getListenerRuleID(kt *kit.Kit, client *dataservice.Client,
	lbID, lblID string, vendor enumor.Vendor) (string, error) {

	rules := []*filter.AtomRule{
		tools.RuleEqual("lb_id", lbID),
		tools.RuleEqual("lbl_id", lblID),
	}

	if l.Protocol.IsLayer7Protocol() {
		rules = append(rules,
			tools.RuleEqual("domain", l.Domain),
			tools.RuleEqual("url", l.URLPath),
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
			return "", nil
		}
		return rule.Details[0].ID, nil
	default:
		return "", fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// GetListenerID ...
func (l *BindRSRecord) GetListenerID(kt *kit.Kit, client *global.LoadBalancerClient,
	lbID string) (string, error) {

	listReq := &core.ListReq{
		Fields: []string{"id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("protocol", l.Protocol),
			tools.RuleEqual("port", l.VPorts[0]),
		),
	}
	listeners, err := client.ListListener(kt, listReq)
	if err != nil {
		logs.Errorf("list listener failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if listeners == nil || len(listeners.Details) == 0 {
		return "", nil
	}

	return listeners.Details[0].ID, nil
}

// checkAction 和数据库的数据进行校验，确认当前记录对应的操作类型
func (l *BindRSRecord) checkAction(kt *kit.Kit, lb *loadbalancer.BaseLoadBalancer, client *dataservice.Client) error {

	// listener
	listenerID, err := l.GetListenerID(kt, client.Global.LoadBalancer, lb.ID)
	if err != nil {
		logs.Errorf("get listener id failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	var expectAction enumor.BatchOperationActionType

	if l.Protocol.IsLayer7Protocol() {
		if listenerID == "" {
			expectAction = enumor.CreateListenerWithURLAndAppendRS
		} else {
			ruleID, err := l.getListenerRuleID(kt, client, lb.ID, listenerID, lb.Vendor)
			if err != nil {
				logs.Errorf("get listener rule id failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			if len(ruleID) > 0 {
				expectAction = enumor.AppendRS
			} else {
				expectAction = enumor.CreateURLAndAppendRS
			}
		}
	} else {
		if listenerID == "" {
			expectAction = enumor.CreateListenerAndAppendRS
		} else {
			expectAction = enumor.AppendRS
		}
	}

	if l.Action != expectAction {
		return fmt.Errorf("action not match, actual action: %s, input action: %s", expectAction, l.Action)
	}

	return nil
}

func (l *BindRSRecord) isExistTcloudUrlRule(kt *kit.Kit, client *tcloud.LoadBalancerClient,
	lbID, lblID string) (bool, error) {

	listReq := &core.ListReq{
		Page: core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("lbl_id", lblID),
			tools.RuleEqual("domain", l.Domain),
			tools.RuleEqual("url", l.URLPath),
		),
	}
	rules, err := client.ListUrlRule(kt, listReq)
	if err != nil {
		logs.Errorf("list url rule failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}
	if rules == nil || len(rules.Details) == 0 {
		return false, nil
	}

	return true, nil
}

// GetKey 获取当前记录的标签
func (l *BindRSRecord) GetKey() string {
	var port string
	if l.HaveEndPort {
		port = fmt.Sprintf("[%d,%d]", l.VPorts[0], l.VPorts[len(l.VPorts)-1])
	} else {
		port = fmt.Sprintf("%d", l.VPorts[0])
	}
	if len(l.Domain) > 0 {
		return fmt.Sprintf("%s %s:%s %s %s", l.VIP, l.Protocol, port, l.Domain, l.URLPath)
	}
	return fmt.Sprintf("%s %s:%s", l.VIP, l.Protocol, port)
}

// GetRSKeys 获取RS列表的标签
func (l *BindRSRecord) GetRSKeys() []string {
	result := make([]string, 0, len(l.RSInfos))
	for _, info := range l.RSInfos {
		result = append(result, info.GetKey())
	}
	return result
}

// checkRSDuplicate 检查RS是否重复(四层转发规则不允许重复)
func (l *BindRSRecord) checkRSDuplicate(kt *kit.Kit, lbID string, client *dataservice.Client) []*cloud.BatchOperationValidateError {
	if l.Protocol.IsLayer7Protocol() {
		return nil
	}

	// 查找该负载均衡下的4层监听器，绑定的所有目标组ID
	tgIDs, err := getBindTargetGroupIDsByLBID(kt, client, lbID, l.Protocol)
	if err != nil {
		logs.Errorf("get bind target group ids by lb id failed, err: %v, rid: %s", err, kt.Rid)
		return []*cloud.BatchOperationValidateError{{Reason: fmt.Sprintf("%s: err: %v", l.GetKey(), err)}}
	}

	// 查找关联表中所有目标组的rs
	listRelRsReq := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", tgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	relRsResp, err := client.Global.LoadBalancer.ListTarget(kt, listRelRsReq)
	if err != nil {
		return []*cloud.BatchOperationValidateError{{Reason: fmt.Sprintf("%s: err: %v", l.GetKey(), err)}}
	}
	existRsMap := make(map[string]struct{})
	for _, tgItem := range relRsResp.Details {
		uniqueKey := fmt.Sprintf("%s:%d", tgItem.IP, tgItem.Port)
		existRsMap[uniqueKey] = struct{}{}
	}

	result := make([]*cloud.BatchOperationValidateError, 0)
	for _, key := range l.GetRSKeys() {
		if _, ok := existRsMap[key]; ok {
			result = append(result, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s: rs: %s already exist,"+
					" (vip+protocol+rsip+rsport) should be globally unique"+
					" for fourth layer listeners", l.GetKey(), key),
			})
		}
	}

	return result
}

func getBindTargetGroupIDsByLBID(kt *kit.Kit, client *dataservice.Client, lbID string, protocol enumor.ProtocolType) ([]string, error) {
	listTGReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("binding_status", enumor.SuccessBindingStatus),
			tools.RuleEqual("listener_rule_type", enumor.Layer4RuleType),
		),
		Page: core.NewDefaultBasePage(),
	}
	tgResp, err := client.Global.LoadBalancer.ListTargetGroupListenerRel(kt, listTGReq)
	if err != nil {
		return nil, err
	}
	if len(tgResp.Details) == 0 {
		return nil, nil
	}

	lblIDs := make([]string, len(tgResp.Details))
	lblTGMap := make(map[string][]string)
	for _, item := range tgResp.Details {
		lblIDs = append(lblIDs, item.LblID)
		lblTGMap[item.LblID] = append(lblTGMap[item.LblID], item.TargetGroupID)
	}
	// 查找对应Protocol的监听器列表
	lblReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", lblIDs), tools.RuleEqual("protocol", protocol)),
		Page:   core.NewDefaultBasePage(),
	}
	lblResp, err := client.Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		logs.Errorf("fail to list listener by lblids, err: %v, lblIDs: %v, rid: %s", err, lblIDs, kt.Rid)
		return nil, err
	}
	if len(lblResp.Details) == 0 {
		return nil, nil
	}
	targetGroupIDs := make([]string, 0)
	for _, item := range lblResp.Details {
		tmpTGIDs, ok := lblTGMap[item.ID]
		if !ok {
			continue
		}
		targetGroupIDs = append(targetGroupIDs, tmpTGIDs...)
	}
	return targetGroupIDs, nil
}

func (l *BindRSRecord) checkPortConflict(kt *kit.Kit, lbID string, client *dataservice.Client) error {
	// 只有新增监听器才需要校验
	if l.Action != enumor.CreateListenerAndAppendRS && l.Action != enumor.CreateListenerWithURLAndAppendRS {
		return nil
	}
	lblReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lb_id", lbID),
			tools.RuleEqual("port", l.VPorts[0])),
		Page: core.NewDefaultBasePage(),
	}
	lblResp, err := client.Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		return err
	}
	if len(lblResp.Details) == 0 {
		return nil
	}

	_, tcpOK := tcpProtocolMap[l.Protocol]
	_, udpOK := udpProtocolMap[l.Protocol]
	for _, lbl := range lblResp.Details {
		if _, ok := tcpProtocolMap[lbl.Protocol]; tcpOK && ok {
			return fmt.Errorf("port[%d] has been occupied by listener: %s(%s)",
				l.VPorts[0], lbl.ID, fmt.Sprintf("%s:%d", lbl.Protocol, lbl.Port))
		}
		if _, ok := udpProtocolMap[lbl.Protocol]; udpOK && ok {
			return fmt.Errorf("port[%d] has been occupied by listener: %s(%s)",
				l.VPorts[0], lbl.ID, fmt.Sprintf("%s:%d", lbl.Protocol, lbl.Port))
		}
	}

	return nil
}

// Validate 基于自身拥有的信息进行合法性校验
// 对于需要依赖外部数据的校验，不在 Validate 的职责中
// 比如：
// 判断RS是否存在，CLB是否存在等等，应该由别的行为进行处理
func (l *BindRSRecord) Validate() error {
	err := l.validateRS()
	if err != nil {
		return err
	}

	if err = l.validateProtocol(); err != nil {
		return err
	}

	if err = l.validateSessionExpired(); err != nil {
		return err
	}

	if err = l.validateCertAndURL(); err != nil {
		return err
	}

	if err = l.validateVIPsAndVPorts(); err != nil {
		return err
	}

	if err = l.validateInstType(); err != nil {
		return err
	}

	if err = l.validateWeight(); err != nil {
		return err
	}

	if err = l.validateScheduler(); err != nil {
		return err
	}

	if err = l.validateListenerName(); err != nil {
		return err
	}

	if err = l.validateIpDomainType(); err != nil {
		return err
	}

	if err = l.validateRSInfoDuplicate(); err != nil {
		return err
	}

	return nil
}

func (l *BindRSRecord) validateSessionExpired() error {
	if l.SessionExpired == 0 || (l.SessionExpired >= 30 && l.SessionExpired <= 3600) {
		return nil
	}
	return fmt.Errorf("session expired should be 0 or between 30 and 3600")
}

func (l *BindRSRecord) validateProtocol() error {
	if _, ok := supportedProtocols[l.Protocol]; !ok {
		return fmt.Errorf("unsupported protocol: %s", l.Protocol)
	}

	return nil
}

func (l *BindRSRecord) validateInstType() error {
	if _, ok := supportInstTypes[l.InstType]; !ok {
		return fmt.Errorf("unsupported instance type: %s", l.InstType)
	}
	return nil
}

func (l *BindRSRecord) validateWeight() error {
	for _, weight := range l.Weights {
		if weight < 0 || weight > 100 {
			return fmt.Errorf("invalid weight value: %d", weight)
		}
	}
	return nil
}

func (l *BindRSRecord) validateScheduler() error {
	if l.Scheduler != "WRR" && l.Scheduler != "LEAST_CONN" {
		return fmt.Errorf("unsupported scheduler: %s", l.Scheduler)
	}
	return nil
}

func (l *BindRSRecord) validateRS() error {
	err := l.validateWeight()
	if err != nil {
		return err
	}

	if len(l.RSIPs) == 0 || len(l.RSPorts) == 0 || len(l.Weights) == 0 {
		return fmt.Errorf("RSIPs, RSPorts and NewWeight cannot be empty")
	}

	if l.HaveEndPort {
		if len(l.RSPorts) != 2 {
			return fmt.Errorf("port range should have two ports")
		}

		if l.VPorts[1]-l.VPorts[0] != l.RSPorts[1]-l.RSPorts[0] {
			return fmt.Errorf("port range should have the same length")
		}

		if len(l.RSIPs) != 1 || len(l.Weights) != 1 {
			return fmt.Errorf("RSIPs and NewWeight should have only one element")
		}

		for _, port := range l.RSPorts {
			if !validPort(port) {
				return fmt.Errorf("invalid RSPort: %d", port)
			}
		}

		l.RSInfos = append(l.RSInfos, &RSInfo{
			InstType: l.InstType,
			IP:       l.RSIPs[0],
			Port:     l.RSPorts[0],
			EndPort:  l.RSPorts[1],
			Weight:   l.Weights[0],
		})
		return nil
	}

	if len(l.RSPorts) > 1 && len(l.RSPorts) != len(l.RSIPs) {
		return fmt.Errorf("the number of RSPorts and RSIPs should be equal or 1")
	}

	if len(l.Weights) > 1 && len(l.Weights) != len(l.RSIPs) {
		return fmt.Errorf("the number of NewWeight and RSIPs should be equal or 1")
	}

	/** 数据补全
	input: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80] NewWeight: [1 1]
	output: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80 80] NewWeight: [1 1]

	input: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80 80] NewWeight: [1]
	output: RSIPs: [1.1.1.1 2.2.2.2] RSPorts: [80 80] NewWeight: [1 1]
	*/
	for len(l.RSPorts) < len(l.RSIPs) {
		l.RSPorts = append(l.RSPorts, l.RSPorts[0])
	}

	for len(l.Weights) < len(l.RSIPs) {
		l.Weights = append(l.Weights, l.Weights[0])
	}

	for i := 0; i < len(l.RSIPs); i++ {
		l.RSInfos = append(l.RSInfos, &RSInfo{
			InstType: l.InstType,
			IP:       l.RSIPs[i],
			Port:     l.RSPorts[i],
			Weight:   l.Weights[i],
		})

	}
	return nil
}

func (l *BindRSRecord) validateRSInfoDuplicate() error {
	m := map[string]struct{}{}
	for _, info := range l.RSInfos {
		flag := fmt.Sprintf("%s:%d", info.IP, info.Port)
		if _, ok := m[flag]; ok {
			return fmt.Errorf("duplicate RS: %s", flag)
		}
		m[flag] = struct{}{}
	}
	return nil
}

func (l *BindRSRecord) validateCertAndURL() error {
	if err := l.validateProtocol(); err != nil {
		return err
	}

	// Protocol横向扩展的可能性不高，不太需要上策略模式
	switch l.Protocol {
	case "HTTPS":
		if len(l.Domain) == 0 || len(l.URLPath) == 0 {
			return fmt.Errorf("domain and url path cannot be empty")
		}
		if len(l.ServerCerts) == 0 {
			return fmt.Errorf("server-cert cannot be empty for HTTPS protocol")
		}
		for _, cert := range l.ServerCerts {
			if len(cert) == 0 {
				return fmt.Errorf("server-cert cannot be empty for HTTPS protocol")
			}
		}
		if len(l.ServerCerts) > 2 {
			return fmt.Errorf("server-cert cannot be more than 2 for HTTPS protocol")
		}

	case "HTTP":
		if len(l.Domain) == 0 || len(l.URLPath) == 0 {
			return fmt.Errorf("domain and url path cannot be empty")
		}
		if len(l.ServerCerts) > 0 || len(l.ClientCert) > 0 {
			return fmt.Errorf("server-cert and client-cert cannot be set for HTTP protocol")
		}
	case "TCP", "UDP":
		if len(l.ServerCerts) > 0 || len(l.ClientCert) > 0 || len(l.Domain) > 0 || len(l.URLPath) > 0 {
			return fmt.Errorf("server-cert and client-cert and domain and url cannot be set for TCP、UDP protocol")
		}
	}

	return nil
}

func (l *BindRSRecord) validateVIPsAndVPorts() error {
	if len(l.VIP) == 0 || len(l.VPorts) == 0 {
		return fmt.Errorf("VIP and VPorts cannot be empty")
	}

	if l.HaveEndPort && len(l.VPorts) != 2 {
		return fmt.Errorf("port range should have two ports")
	}

	if !l.HaveEndPort && len(l.VPorts) != 1 {
		return fmt.Errorf("parse VPorts error, should be 1 or 2")
	}

	for _, port := range l.VPorts {
		if !validPort(port) {
			return fmt.Errorf("invalid VPort: %d", port)
		}
	}

	return nil
}

func (l *BindRSRecord) validateListenerName() error {
	if len(l.ListenerName) == 0 || l.ListenerName == "" {
		return fmt.Errorf("listener name should not be empty")
	}
	if len(l.ListenerName) > 255 {
		return fmt.Errorf("listener name is too long")
	}
	return nil
}

func (l *BindRSRecord) validateIpDomainType() error {
	if l.IPDomainType == ipDomainTypeDomain {
		return nil
	}
	version, err := getIPVersion(l.VIP)
	if err != nil {
		return err
	}

	if version == enumor.Ipv4 && l.IPDomainType != ipDomainTypeIPv4 {
		return fmt.Errorf("ip version is ipv4, but ip domain type got %s", l.IPDomainType)
	}

	if version == enumor.Ipv6 && l.IPDomainType != ipDomainTypeIPv6 {
		return fmt.Errorf("ip version is ipv6, but ip domain type got %s", l.IPDomainType)
	}

	return nil
}

func validPort(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	return true
}

func getIPVersion(ipStr string) (enumor.IPAddressType, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}
	if ip.To4() != nil {
		return enumor.Ipv4, nil
	}
	if ip.To16() != nil {
		return enumor.Ipv6, nil
	}
	return "", nil
}
