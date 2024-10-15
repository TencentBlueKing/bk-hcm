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
func (c *CreateLayer4ListenerPreviewExecutor) Execute(kt *kit.Kit, rawData [][]string) (interface{}, error) {
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

const createLayer4ListenerExcelTableLen = 7

func (c *CreateLayer4ListenerPreviewExecutor) convertDataToPreview(rawData [][]string) error {
	for _, data := range rawData {
		data = trimSpaceForSlice(data)
		if len(data) < createLayer4ListenerExcelTableLen {
			return fmt.Errorf("invalid data")
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
		if len(data) > createLayer4ListenerExcelTableLen {
			detail.UserRemark = data[7]
		}
		c.details = append(c.details, detail)
	}
	return nil
}

func (c *CreateLayer4ListenerPreviewExecutor) validate(kt *kit.Kit) error {
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
			detail.ValidateResult = append(detail.ValidateResult, fmt.Sprintf("clb vip(%s)not match", detail.ClbVipDomain))
		}
		detail.RegionID = lb.Region

		if err = c.validateListener(kt, detail); err != nil {
			logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
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

	rule, err := c.getURLRule(kt, curDetail.CloudClbID, listener.CloudID)
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

func (c *CreateLayer4ListenerPreviewExecutor) getURLRule(kt *kit.Kit, lbCloudID, listenerCloudID string) (
	*corelb.TCloudLbUrlRule, error) {

	switch c.vendor {
	case enumor.TCloud:
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("cloud_lb_id", lbCloudID),
				tools.RuleEqual("cloud_lbl_id", listenerCloudID),
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
	ClbVipDomain string `json:"clb_vip_domain"`
	CloudClbID   string `json:"cloud_clb_id"`

	Protocol       enumor.ProtocolType `json:"protocol"`
	ListenerPorts  []int               `json:"listener_port"`
	Scheduler      enumor.Scheduler    `json:"scheduler"`
	Session        int                 `json:"session"`
	HealthCheck    bool                `json:"health_check"`
	UserRemark     string              `json:"user_remark"`
	Status         ImportStatus        `json:"status"`
	ValidateResult []string            `json:"validate_result"`

	RegionID string `json:"region_id"`
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
