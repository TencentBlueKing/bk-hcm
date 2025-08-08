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
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

var _ ImportPreviewExecutor = (*CreateLayer4ListenerPreviewExecutor)(nil)

func newCreateLayer4ListenerPreviewExecutor(cli *dataservice.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) *CreateLayer4ListenerPreviewExecutor {

	return &CreateLayer4ListenerPreviewExecutor{
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateLayer4ListenerPreviewExecutor excel导入——创建四层监听器执行器
type CreateLayer4ListenerPreviewExecutor struct {
	*basePreviewExecutor

	details []*CreateLayer4ListenerDetail
}

// Execute 执行
func (c *CreateLayer4ListenerPreviewExecutor) Execute(kt *kit.Kit, rawData [][]string, headers []string) (interface{},
	error) {
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

const (
	createLayer4ListenerExcelTableLen = 7
	// excel表格表头长度
	createLayer4ListenerExcelTableHeaderLen = 10
	// excel表格行号偏移量, clb数据从第四行开始
	excelTableLineNumberOffset = 4
)

func (c *CreateLayer4ListenerPreviewExecutor) convertDataToPreview(rawData [][]string, headers []string) error {
	if len(headers) < createLayer4ListenerExcelTableHeaderLen {
		return fmt.Errorf("table headers length less than %d, got: %d, headers: %v",
			createLayer4ListenerExcelTableHeaderLen, len(headers), headers)
	}

	for i, data := range rawData {
		data = trimSpaceForSlice(data)
		if len(data) < createLayer4ListenerExcelTableLen {
			return fmt.Errorf("line[%d] data length less than %d, got: %d, data: %v",
				i+excelTableLineNumberOffset, createLayer4ListenerExcelTableLen, len(data), data)
		}

		detail := &CreateLayer4ListenerDetail{
			ValidateResult: make([]string, 0),
		}
		detail.ClbVipDomain = data[0]
		detail.CloudClbID = data[1]
		detail.Protocol = enumor.ProtocolType(strings.ToUpper(data[2]))
		ports, err := parsePort(data[3])
		if err != nil {
			return err
		}
		detail.ListenerPorts = ports
		detail.Scheduler = enumor.Scheduler(data[4])
		session, err := strconv.Atoi(data[5])
		if err != nil {
			return err
		}
		detail.Session = session
		switch data[6] {
		case "enable":
			detail.HealthCheck = true
		case "disable":
			detail.HealthCheck = false
		default:
			return fmt.Errorf("HealthCheck: invalid input: %s", data[6])
		}
		// 监听器名称和用户备注是可选的
		if len(data) > createLayer4ListenerExcelTableLen {
			detail.Name = data[7]
			if len(data) > createLayer4ListenerExcelTableLen+1 {
				detail.UserRemark = data[8]
			}
		}
		c.details = append(c.details, detail)
	}
	return nil
}

func (c *CreateLayer4ListenerPreviewExecutor) validate(kt *kit.Kit) error {
	if len(c.details) == 0 {
		return fmt.Errorf("there are no details to be executed")
	}
	//key: clbID+protocol+port value record index
	recordMap := make(map[string]int)
	clbIDMap := make(map[string]struct{})
	for cur, detail := range c.details {
		detail.validate()
		// 检查记录是否重复
		key := fmt.Sprintf("%s-%s-%v", detail.CloudClbID, detail.Protocol, detail.ListenerPorts)
		if i, ok := recordMap[key]; ok {
			c.details[i].Status.SetNotExecutable()
			c.details[i].ValidateResult = append(c.details[i].ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d;", i+1))

			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(c.details[i].ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d;", cur+1))
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

func (c *CreateLayer4ListenerPreviewExecutor) validateWithDB(kt *kit.Kit, cloudIDs []string) error {
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli, c.vendor, c.accountID, c.bkBizID, cloudIDs)
	if err != nil {
		return err
	}

	concurrentErr := concurrence.BaseExec(cc.CloudServer().ConcurrentConfig.CLBImportCount, c.details,
		func(detail *CreateLayer4ListenerDetail) error {

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
					fmt.Sprintf("clb vip(%s)not match", detail.ClbVipDomain))
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
	return nil
}

func (c *CreateLayer4ListenerPreviewExecutor) validateListener(kt *kit.Kit,
	curDetail *CreateLayer4ListenerDetail) error {

	listener, err := getListener(kt, c.dataServiceCli, c.accountID, curDetail.CloudClbID, curDetail.Protocol,
		curDetail.ListenerPorts[0], c.bkBizID, c.vendor)
	if err != nil {
		return err
	}
	if listener == nil {
		return nil
	}

	rule, err := c.getURLRule(kt, listener.LbID, listener.ID)
	if err != nil {
		return err
	}
	if rule == nil {
		logs.Errorf("listener exist, but url rule not found, listener: %v, rid: %s", listener, kt.Rid)
		return errors.New("listener exist, but url rule not found")
	}
	var ruleHealthCheck bool
	if rule.HealthCheck != nil && rule.HealthCheck.HealthSwitch != nil {
		ruleHealthCheck = *rule.HealthCheck.HealthSwitch == 1
	}

	if enumor.Scheduler(rule.Scheduler) != curDetail.Scheduler || rule.SessionExpire != int64(curDetail.Session) ||
		ruleHealthCheck != curDetail.HealthCheck {

		// 已存在监听器且配置与当前导入的记录不一致时, 设置当前记录为不可执行状态
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("already exist listener(%s), and the configuration mismatch, port: %d, protocol: %s,"+
				" scheduler: %s, session: %d, healthCheck: %v", curDetail.CloudClbID, listener.Port, listener.Protocol,
				rule.Scheduler, rule.SessionExpire, ruleHealthCheck))
		return nil
	}

	curDetail.Status.SetExisting()
	curDetail.ValidateResult = append(curDetail.ValidateResult,
		fmt.Sprintf("already exist listener(%s), port: %d, protocol: %s", curDetail.CloudClbID, listener.Port,
			listener.Protocol))

	return nil
}

func (c *CreateLayer4ListenerPreviewExecutor) getURLRule(kt *kit.Kit, lbID, listenerID string) (
	*corelb.TCloudLbUrlRule, error) {

	switch c.vendor {
	case enumor.TCloud:
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("lb_id", lbID),
				tools.RuleEqual("lbl_id", listenerID),
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

// CreateLayer4ListenerDetail 创建四层监听器预览记录
type CreateLayer4ListenerDetail struct {
	Layer4ListenerDetail `json:",inline"`
	ListenerPorts        []int        `json:"listener_port"`
	HealthCheck          bool         `json:"health_check"`
	Status               ImportStatus `json:"status"`
	ValidateResult       []string     `json:"validate_result"`
	RegionID             string       `json:"region_id"`
}

func (c *CreateLayer4ListenerDetail) validate() {
	var err error
	defer func() {
		if err != nil {
			c.Status.SetNotExecutable()
			c.ValidateResult = append(c.ValidateResult, err.Error())
			return
		}
		c.Status.SetExecutable()
	}()

	// 验证监听器名称，校验规则同"CLB名称"
	if len(c.Name) > 60 {
		err = fmt.Errorf("the length of the listener name should not exceed 60")
		return
	}
	if c.Protocol != enumor.UdpProtocol && c.Protocol != enumor.TcpProtocol {
		err = fmt.Errorf("unsupport listener protocol type: %s", c.Protocol)
		return
	}
	err = validateScheduler(c.Scheduler)
	if err != nil {
		return
	}
	err = validateSession(c.Session)
	if err != nil {
		return
	}
	err = validatePort(c.ListenerPorts)
	if err != nil {
		return
	}
}
