/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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
	"errors"
	"fmt"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"strconv"
	"strings"
)

var _ ImportPreviewExecutor = (*Layer4ListenerBindRSPreviewExecutor)(nil)

func newLayer4ListenerBindRSPreviewExecutor(cli *dataservice.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) *Layer4ListenerBindRSPreviewExecutor {

	return &Layer4ListenerBindRSPreviewExecutor{
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// Layer4ListenerBindRSPreviewExecutor excel导入——创建四层监听器执行器
type Layer4ListenerBindRSPreviewExecutor struct {
	*basePreviewExecutor

	details []*Layer4ListenerBindRSDetail
}

// Execute ...
func (l *Layer4ListenerBindRSPreviewExecutor) Execute(kt *kit.Kit, rawData [][]string) (interface{}, error) {
	err := l.convertDataToPreview(rawData)
	if err != nil {
		return nil, err
	}

	err = l.validate(kt)
	if err != nil {
		logs.Errorf("validate data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return l.details, nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) convertDataToPreview(rawData [][]string) error {
	for _, data := range rawData {
		data = trimSpaceForSlice(data)

		detail := &Layer4ListenerBindRSDetail{}
		detail.ClbVipDomain = data[0]
		detail.CloudClbID = data[1]
		detail.Protocol = enumor.ProtocolType(strings.ToUpper(data[2]))
		listenerPorts, err := parsePort(data[3])
		if err != nil {
			return err
		}
		detail.ListenerPort = listenerPorts
		detail.InstType = enumor.InstType(strings.ToUpper(data[4]))
		detail.RsIp = data[5]
		rsPort, err := parsePort(data[6])
		if err != nil {
			return err
		}
		detail.RsPort = rsPort
		weight, err := strconv.Atoi(strings.TrimSpace(data[7]))
		if err != nil {
			return err
		}
		detail.Weight = weight
		if len(data) > 8 {
			detail.UserRemark = data[8]
		}

		l.details = append(l.details, detail)
	}
	return nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) validate(kt *kit.Kit) error {
	recordMap := make(map[string]int)
	clbIDMap := make(map[string]struct{})
	for cur, detail := range l.details {
		detail.validate()
		// 检查记录是否重复
		key := fmt.Sprintf("%s-%s-%v-%s-%v",
			detail.CloudClbID, detail.Protocol, detail.ListenerPort, detail.RsIp, detail.RsPort)
		if i, ok := recordMap[key]; ok {
			l.details[i].Status = NotExecutable
			l.details[i].ValidateResult += fmt.Sprintf("存在重复记录, line: %d;", i+1)

			detail.Status = NotExecutable
			detail.ValidateResult += fmt.Sprintf("存在重复记录, line: %d;", cur+1)
		}
		recordMap[key] = cur
		clbIDMap[detail.CloudClbID] = struct{}{}
	}
	err := l.validateWithDB(kt, converter.MapKeyToSlice(clbIDMap))
	if err != nil {
		logs.Errorf("validate with db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) validateWithDB(kt *kit.Kit, cloudIDs []string) error {
	loadBalancers, err := getLoadBalancers(kt, l.dataServiceCli, l.accountID, l.bkBizID, cloudIDs)
	if err != nil {
		return err
	}
	lbMap := make(map[string]corelb.BaseLoadBalancer, len(loadBalancers))
	for _, balancer := range loadBalancers {
		lbMap[balancer.CloudID] = balancer
	}

	for _, detail := range l.details {
		lb, ok := lbMap[detail.CloudClbID]
		if !ok {
			return fmt.Errorf("clb(%s) not exist", detail.CloudClbID)
		}
		if _, ok = l.regionIDMap[lb.Region]; !ok {
			return fmt.Errorf("clb region not match, clb.region: %s, input: %v", lb.Region, l.regionIDMap)
		}

		ipSet := append(lb.PrivateIPv4Addresses, lb.PrivateIPv6Addresses...)
		ipSet = append(ipSet, lb.PublicIPv4Addresses...)
		ipSet = append(ipSet, lb.PublicIPv6Addresses...)
		if detail.ClbVipDomain != lb.Domain && !slice.IsItemInSlice(ipSet, detail.ClbVipDomain) {
			detail.Status = NotExecutable
			detail.ValidateResult += fmt.Sprintf("clb的vip(%s)不匹配", detail.ClbVipDomain)
			continue
		}
		lblCloudID, err := l.validateListener(kt, detail)
		if err != nil {
			logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		instID, err := l.validateRS(kt, detail, lb.Region, lb.CloudVpcID)
		if err != nil {
			logs.Errorf("validate rs failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if err = l.validateTarget(kt, detail, lblCloudID, instID, detail.RsPort[0]); err != nil {
			logs.Errorf("validate target failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

// validateTarget 校验RS是否已经绑定到对应的监听器中, 如果已经绑定则校验权重是否一致. 没有绑定则直接返回.
func (l *Layer4ListenerBindRSPreviewExecutor) validateTarget(kt *kit.Kit, detail *Layer4ListenerBindRSDetail,
	lblCloudID, instID string, port int) error {

	if lblCloudID == "" || instID == "" {
		return nil
	}
	tgID, err := getTargetGroupID(kt, l.dataServiceCli, lblCloudID)
	if err != nil {
		return err
	}
	target, err := getTarget(kt, l.dataServiceCli, tgID, instID, port)
	if err != nil {
		return err
	}
	if target == nil {
		return nil
	}

	if int(converter.PtrToVal(target.Weight)) != detail.Weight {
		detail.Status = NotExecutable
		detail.ValidateResult += fmt.Sprintf("RS已绑定，且权重不一致")
		return nil
	}

	if detail.Status != NotExecutable {
		detail.Status = Existing
		detail.ValidateResult += fmt.Sprintf("RS已绑定")
	}

	return nil
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

func getTargetGroupID(kt *kit.Kit, cli *dataservice.Client, ruleID string) (string, error) {
	listReq := &core.ListReq{
		Fields: []string{"target_group_id"},
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cloud_listener_rule_id", ruleID),
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

func (l *Layer4ListenerBindRSPreviewExecutor) validateRS(kt *kit.Kit,
	curDetail *Layer4ListenerBindRSDetail, region, vpc string) (string, error) {

	if curDetail.InstType == enumor.EniInstType {
		// ENI 不做校验
		return "", nil
	}

	cvm, err := getCvm(kt, l.dataServiceCli, curDetail.RsIp, l.vendor, l.bkBizID, l.accountID)
	if err != nil {
		return "", err
	}
	if cvm == nil {
		curDetail.Status = NotExecutable
		curDetail.ValidateResult += fmt.Sprintf("rs(%s) not exist", curDetail.RsIp)
		return "", nil
	}
	if cvm.Region != region {
		curDetail.Status = NotExecutable
		curDetail.ValidateResult += fmt.Sprintf("rs(%s) region not match, rs.region: %s, lb.region: %v",
			curDetail.RsIp, cvm.Region, region)
		return cvm.CloudID, nil
	}
	if !slice.IsItemInSlice(cvm.CloudVpcIDs, vpc) {
		curDetail.Status = NotExecutable
		curDetail.ValidateResult += fmt.Sprintf("rs(%s) vpc not match, rs.vpc: %v, lb.vpc: %s",
			curDetail.RsIp, cvm.CloudVpcIDs, vpc)
		return cvm.CloudID, nil
	}

	return cvm.CloudID, nil
}

func getCvm(kt *kit.Kit, cli *dataservice.Client, ip string,
	vendor enumor.Vendor, bkBizID int64, accountID string) (*corecvm.BaseCvm, error) {

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

func (l *Layer4ListenerBindRSPreviewExecutor) validateListener(kt *kit.Kit,
	curDetail *Layer4ListenerBindRSDetail) (string, error) {

	listener, err := getListener(kt, l.dataServiceCli, l.accountID,
		curDetail.CloudClbID, curDetail.ListenerPort[0], l.bkBizID)
	if err != nil {
		return "", err
	}
	if listener == nil {
		curDetail.Status = NotExecutable
		curDetail.ValidateResult += "监听器不存在"
		return "", nil
	}
	if listener.Protocol != curDetail.Protocol {
		curDetail.Status = NotExecutable
		curDetail.ValidateResult += fmt.Sprintf("监听器协议不匹配, input(%s) actual(%s)",
			curDetail.Protocol, listener.Protocol)
		return "", nil
	}

	return listener.CloudID, nil
}

// Layer4ListenerBindRSDetail ...
type Layer4ListenerBindRSDetail struct {
	ClbVipDomain string              `json:"clb_vip_domain"`
	CloudClbID   string              `json:"cloud_clb_id"`
	Protocol     enumor.ProtocolType `json:"protocol"`
	ListenerPort []int               `json:"listener_port"`

	InstType       enumor.InstType `json:"inst_type"`
	RsIp           string          `json:"rs_ip"`
	RsPort         []int           `json:"rs_port"`
	Weight         int             `json:"weight"`
	UserRemark     string          `json:"user_remark"`
	Status         ImportStatus    `json:"status"`
	ValidateResult string          `json:"validate_result"`
}

func (c *Layer4ListenerBindRSDetail) validate() {
	var err error
	defer func() {
		if err != nil {
			c.Status = NotExecutable
			c.ValidateResult = err.Error()
			return
		}
		c.Status = Executable
	}()

	if c.Protocol != enumor.UdpProtocol && c.Protocol != enumor.TcpProtocol {
		err = errors.New("协议类型错误")
		return
	}
	err = validatePort(c.ListenerPort)
	if err != nil {
		return
	}
	err = validatePort(c.RsPort)
	if err != nil {
		return
	}
	err = validateInstType(c.InstType)
	if err != nil {
		return
	}
	err = validateWeight(c.Weight)
	if err != nil {
		return
	}
	err = validateEndPort(c.ListenerPort, c.RsPort)
	if err != nil {
		return
	}
	if len(c.RsPort) == 2 && c.Weight == 0 {
		err = errors.New("端口段的RS权重必须大于0")
		return
	}
}
