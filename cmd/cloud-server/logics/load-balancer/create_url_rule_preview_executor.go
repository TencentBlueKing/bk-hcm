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

	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/concurrence"
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
func (c *CreateUrlRulePreviewExecutor) Execute(kt *kit.Kit, rawData [][]string, headers []string) (interface{}, error) {
	err := c.convertDataToPreview(rawData, headers)
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

// createURLRuleExcelTableHeaderLen 表头长度
const createURLRuleExcelTableHeaderLen = 12

func (c *CreateUrlRulePreviewExecutor) convertDataToPreview(rawData [][]string, headers []string) error {
	if len(headers) < createURLRuleExcelTableHeaderLen {
		return fmt.Errorf("headers length less than %d, got: %d, headers: %v",
			createURLRuleExcelTableHeaderLen, len(headers), headers)
	}
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

		isDefaultDomain := strings.ToUpper(data[5])
		switch isDefaultDomain {
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
	if len(c.details) == 0 {
		return fmt.Errorf("there are no details to be executed")
	}
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

	concurrentErr := concurrence.BaseExec(cc.CloudServer().ConcurrentConfig.CLBImportCount, c.details,
		func(detail *CreateUrlRuleDetail) error {

			lb, ok := lbMap[detail.CloudClbID]
			if !ok {
				return fmt.Errorf("clb(%s) not exist", detail.CloudClbID)
			}
			if _, ok = c.regionIDMap[lb.Region]; !ok {
				return fmt.Errorf("clb region not match, clb.region: %s, input: %+v", lb.Region, c.regionIDMap)
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
			return nil
		})
	if concurrentErr != nil {
		logs.Errorf("validate with db failed, err: %v, rid: %s", concurrentErr, kt.Rid)
		return concurrentErr
	}
	if err = c.validateDefaultDomain(kt); err != nil {
		logs.Errorf("validate default domain failed, err: %v, rid: %s", err, kt.Rid)
		return err
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
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("lb(%s) listenerPort(%d) does not exist",
				curDetail.CloudClbID, curDetail.ListenerPort[0]))
		return nil
	}
	curDetail.listenerID = listener.ID
	curDetail.listenerDefaultDomain = listener.DefaultDomain

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

const (
	listenerDefaultSep = "/"
)

func (c *CreateUrlRulePreviewExecutor) validateDefaultDomain(kt *kit.Kit) error {

	// group by listener
	classifySlice := classifier.ClassifySlice(c.details, classifyFunc)
	for _, details := range classifySlice {
		listenerDefaultDomain := ""
		listenerID := ""
		newDefaultDomain := ""
		for _, detail := range details {
			if len(detail.listenerID) == 0 {
				// listener not found, skip this detail
				continue
			}
			listenerID = detail.listenerID
			if detail.listenerDefaultDomain != "" {
				listenerDefaultDomain = detail.listenerDefaultDomain
				if detail.DefaultDomain && detail.Domain != detail.listenerDefaultDomain {
					detail.Status.SetNotExecutable()
					detail.ValidateResult = append(detail.ValidateResult,
						fmt.Sprintf("listener(%s) default domain: %s, but got: %s",
							detail.listenerID, detail.listenerDefaultDomain, detail.Domain))
				}
				continue
			}

			// 监听器没有配置default domain的情况
			if detail.DefaultDomain {
				if newDefaultDomain == "" {
					newDefaultDomain = detail.Domain
					continue
				}
				if detail.Domain != newDefaultDomain {
					detail.Status.SetNotExecutable()
					detail.ValidateResult = append(detail.ValidateResult,
						fmt.Sprintf("listener(%s) multiple records have been set as default domain, current: %s, previous: %s",
							detail.listenerID, detail.Domain, newDefaultDomain),
					)
				}
			}
		}
		if listenerID == "" {
			// listener not found, skip these details
			continue
		}
		if listenerDefaultDomain == "" && newDefaultDomain == "" {
			for _, detail := range details {
				detail.Status.SetNotExecutable()
				detail.ValidateResult = append(detail.ValidateResult,
					fmt.Sprintf("listener(%s) does not have a default domain, and no default domain is set",
						listenerID),
				)
			}
		}
	}
	return nil
}

func classifyFunc(detail *CreateUrlRuleDetail) string {
	port := ""
	if len(detail.ListenerPort) > 0 {
		port = strconv.Itoa(detail.ListenerPort[0])
	}
	args := []string{detail.CloudClbID, string(detail.Protocol), port}
	return strings.Join(args, listenerDefaultSep)
}

func decodeClassifyKey(key string) (string, enumor.ProtocolType, int, error) {
	arr := strings.Split(key, listenerDefaultSep)
	if len(arr) != 3 {
		return "", "", 0, fmt.Errorf("invalid key: %s", key)
	}
	cloudClbID := arr[0]
	protocol := enumor.ProtocolType(arr[1])
	listenerPort, err := strconv.Atoi(arr[2])
	if err != nil {
		return "", "", 0, err
	}
	return cloudClbID, protocol, listenerPort, nil
}

// CreateUrlRuleDetail ...
type CreateUrlRuleDetail struct {
	RuleDetail   `json:",inline"`
	ListenerPort []int `json:"listener_port"`
	HealthCheck  bool  `json:"health_check"`

	UserRemark     string       `json:"user_remark"`
	Status         ImportStatus `json:"status"`
	ValidateResult []string     `json:"validate_result"`

	RegionID string `json:"region_id"`

	// listenerID 在 validateListener 阶段填充, 后续submit阶段会重复使用到,
	// 如果为空, 那就意味着当前detail的监听器条件无法匹配到对应的listener, 可以认为listener not found
	listenerID string
	// listenerDefaultDomain 的填充逻辑和 listenerID 一样, 但 listener.DefaultDomain 会有为空的可能
	listenerDefaultDomain string
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
