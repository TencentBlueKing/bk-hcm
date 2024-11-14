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
	"strconv"
	"strings"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

var _ ImportPreviewExecutor = (*CreateUrlRulePreviewExecutor)(nil)

func newCreateUrlRulePreviewExecutor(cli *dataservice.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) *CreateUrlRulePreviewExecutor {

	return &CreateUrlRulePreviewExecutor{
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateUrlRulePreviewExecutor 创建七层监听器预览执行器
type CreateUrlRulePreviewExecutor struct {
	*basePreviewExecutor

	details []*CreateUrlRuleDetail
}

// Execute ...
func (c *CreateUrlRulePreviewExecutor) Execute(kt *kit.Kit, rawData [][]string) (interface{}, error) {
	err := c.convertDataToPreview(rawData)
	if err != nil {
		return nil, err
	}

	err = c.validate(kt)
	if err != nil {
		logs.Errorf("validate data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return c.details, nil
}

const createURLRuleExcelTableLen = 10

func (c *CreateUrlRulePreviewExecutor) convertDataToPreview(rawData [][]string) error {
	for i, data := range rawData {
		if len(data) < createURLRuleExcelTableLen {
			return fmt.Errorf("line[%d] data length less than %d, got: %d, data: %v",
				i+excelTableLineNumberOffset, createURLRuleExcelTableLen, len(data), data)
		}

		data = trimSpaceForSlice(data)
		detail := &CreateUrlRuleDetail{
			ValidateResult: make([]string, 0),
		}
		detail.ClbVipDomain = data[0]
		detail.CloudClbID = data[1]
		detail.Protocol = enumor.ProtocolType(strings.ToUpper(data[2]))
		ports, err := parsePort(data[3])
		if err != nil {
			return err
		}
		detail.ListenerPort = ports
		detail.Domain = data[4]

		switch data[5] {
		case "TRUE":
			detail.DefaultDomain = true
		case "FALSE":
			detail.DefaultDomain = false
		default:
			return fmt.Errorf("DefaultDomain: invalid input: %s", data[5])
		}

		detail.UrlPath = data[6]
		detail.Scheduler = enumor.Scheduler(data[7])
		session, err := strconv.Atoi(data[8])
		if err != nil {
			return err
		}
		detail.Session = session

		switch data[9] {
		case "enable":
			detail.HealthCheck = true
		case "disable":
			detail.HealthCheck = false
		default:
			return fmt.Errorf("HealthCheck: invalid input: %s", data[9])
		}

		if len(data) > createURLRuleExcelTableLen {
			detail.UserRemark = data[10]
		}
		c.details = append(c.details, detail)
	}
	return nil
}

func (c *CreateUrlRulePreviewExecutor) validate(kt *kit.Kit) error {
	recordMap := make(map[string]int)
	clbIDMap := make(map[string]struct{})
	for cur, detail := range c.details {
		detail.validate()
		// 检查记录是否重复
		key := fmt.Sprintf("%s-%s-%v-%s-%s", detail.CloudClbID,
			detail.Protocol, detail.ListenerPort, detail.Domain, detail.UrlPath)
		if i, ok := recordMap[key]; ok {
			c.details[i].Status.SetNotExecutable()
			c.details[i].ValidateResult = append(c.details[i].ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d", i+1))

			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(c.details[i].ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d", cur+1))
		}
		recordMap[key] = cur
		clbIDMap[detail.CloudClbID] = struct{}{}
	}
	err := c.validateWithDB(kt, converter.MapKeyToSlice(clbIDMap))
	if err != nil {
		logs.Errorf("validate with db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (c *CreateUrlRulePreviewExecutor) validateWithDB(kt *kit.Kit, cloudIDs []string) error {
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli, c.vendor, c.accountID, c.bkBizID, cloudIDs)
	if err != nil {
		return err
	}

	for _, detail := range c.details {
		lb, ok := lbMap[detail.CloudClbID]
		if !ok {
			return fmt.Errorf("clb(%s) not exist", detail.CloudClbID)
		}
		if _, ok = c.regionIDMap[lb.Region]; !ok {
			return fmt.Errorf("clb region not match, clb.region: %s, input: %v", lb.Region, c.regionIDMap)
		}

		ipSet := append(lb.PrivateIPv4Addresses, lb.PrivateIPv6Addresses...)
		ipSet = append(ipSet, lb.PublicIPv4Addresses...)
		ipSet = append(ipSet, lb.PublicIPv6Addresses...)
		if detail.ClbVipDomain != lb.Domain && !slice.IsItemInSlice(ipSet, detail.ClbVipDomain) {
			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(detail.ValidateResult,
				fmt.Sprintf("clb vip(%s) not match", detail.ClbVipDomain))
		}
		detail.RegionID = lb.Region

		if err = c.validateListener(kt, detail); err != nil {
			logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func (c *CreateUrlRulePreviewExecutor) validateListener(kt *kit.Kit, curDetail *CreateUrlRuleDetail) error {

	listener, err := getListener(kt, c.dataServiceCli, c.accountID, curDetail.CloudClbID, curDetail.Protocol,
		curDetail.ListenerPort[0], c.bkBizID, c.vendor)
	if err != nil {
		return err
	}
	if listener == nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, fmt.Sprintf("lb(%s) listenerPort(%d) does not exist",
			curDetail.CloudClbID, curDetail.ListenerPort[0]))
		return nil
	}

	if curDetail.DefaultDomain && curDetail.Domain != listener.DefaultDomain {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("listener(%s) default domain name(%s) is inconsistent with input record(%s)",
				listener.ID, listener.DefaultDomain, curDetail.Domain))
		return nil
	}

	rule, err := getURLRule(kt, c.dataServiceCli, c.vendor,
		curDetail.CloudClbID, listener.CloudID, curDetail.Domain, curDetail.UrlPath)
	if err != nil {
		return err
	}
	if rule == nil {
		return nil
	}

	var ruleHealthCheck bool
	if rule.HealthCheck != nil && rule.HealthCheck.HealthSwitch != nil {
		ruleHealthCheck = *rule.HealthCheck.HealthSwitch == 1
	}

	if listener.Protocol != curDetail.Protocol ||
		enumor.Scheduler(rule.Scheduler) != curDetail.Scheduler ||
		rule.SessionExpire != int64(curDetail.Session) ||
		ruleHealthCheck != curDetail.HealthCheck {

		// 已存在urlRule且配置与当前导入的记录不一致时, 设置当前记录为不可执行状态
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("URLRule(%s) already exist, and the configuration does not match, port: %d, protocol: %s,"+
				" domain: %s, url: %s, scheduler: %s, session: %d, healthCheck: %v", rule.ID, listener.Port,
				listener.Protocol, rule.Domain, rule.URL, rule.Scheduler, rule.SessionExpire, ruleHealthCheck))
		return nil
	}

	curDetail.Status.SetExisting()
	curDetail.ValidateResult = append(curDetail.ValidateResult,
		fmt.Sprintf("already exist listener(%s), port: %d, protocol: %s", curDetail.CloudClbID, listener.Port,
			listener.Protocol))

	return nil
}

func (c *CreateUrlRulePreviewExecutor) getURLRule(kt *kit.Kit, lbCloudID, listenerCloudID, domain, url string) (
	*corelb.TCloudLbUrlRule, error) {

	switch c.vendor {
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
		rule, err := c.dataServiceCli.TCloud.LoadBalancer.ListUrlRule(kt, req)
		if err != nil {
			logs.Errorf("list url rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(rule.Details) > 0 {
			return &rule.Details[0], nil
		}
	default:
		return nil, fmt.Errorf("vendor(%s) not support", c.vendor)
	}
	return nil, nil
}

// CreateUrlRuleDetail ...
type CreateUrlRuleDetail struct {
	ClbVipDomain string              `json:"clb_vip_domain"`
	CloudClbID   string              `json:"cloud_clb_id"`
	Protocol     enumor.ProtocolType `json:"protocol"`
	ListenerPort []int               `json:"listener_port"`

	Domain        string           `json:"domain"`
	DefaultDomain bool             `json:"default_domain"`
	UrlPath       string           `json:"url_path"`
	Scheduler     enumor.Scheduler `json:"scheduler"`
	Session       int              `json:"session"`
	HealthCheck   bool             `json:"health_check"`

	UserRemark     string       `json:"user_remark"`
	Status         ImportStatus `json:"status"`
	ValidateResult []string     `json:"validate_result"`

	RegionID string `json:"region_id"`
}

func (c *CreateUrlRuleDetail) validate() {
	var err error
	defer func() {
		if err != nil {
			c.Status.SetNotExecutable()
			c.ValidateResult = append(c.ValidateResult, err.Error())
			return
		}
		c.Status.SetExecutable()
	}()

	if c.Protocol != enumor.HttpProtocol && c.Protocol != enumor.HttpsProtocol {
		err = errors.New("protocol type error")
		return
	}
	if len(c.Domain) == 0 || len(c.UrlPath) == 0 {
		err = errors.New("domain and url are required")
		return
	}
	err = validateSession(c.Session)
	if err != nil {
		return
	}
	err = validatePort(c.ListenerPort)
	if err != nil {
		return
	}
	err = validateScheduler(c.Scheduler)
	if err != nil {
		return
	}
}
