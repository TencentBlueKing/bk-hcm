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

	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

var _ ImportPreviewExecutor = (*Layer7ListenerBindRSPreviewExecutor)(nil)

func newLayer7ListenerBindRSPreviewExecutor(cli *dataservice.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) *Layer7ListenerBindRSPreviewExecutor {

	return &Layer7ListenerBindRSPreviewExecutor{
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// Layer7ListenerBindRSPreviewExecutor excel导入——创建四层监听器执行器
type Layer7ListenerBindRSPreviewExecutor struct {
	*basePreviewExecutor

	details []*Layer7ListenerBindRSDetail
}

// Execute ...
func (l *Layer7ListenerBindRSPreviewExecutor) Execute(kt *kit.Kit, rawData [][]string) (interface{}, error) {
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

const layer7listenerBindRSExcelTableLen = 10

func (l *Layer7ListenerBindRSPreviewExecutor) convertDataToPreview(rawData [][]string) error {
	for _, data := range rawData {
		data = trimSpaceForSlice(data)
		if len(data) < layer7listenerBindRSExcelTableLen {
			return errors.New("invalid data")
		}
		detail := &Layer7ListenerBindRSDetail{
			ValidateResult: make([]string, 0),
		}
		detail.ClbVipDomain = data[0]
		detail.CloudClbID = data[1]
		detail.Protocol = enumor.ProtocolType(strings.ToUpper(data[2]))
		listenerPorts, err := parsePort(data[3])
		if err != nil {
			return err
		}
		detail.ListenerPort = listenerPorts
		detail.Domain = data[4]
		detail.URLPath = data[5]
		detail.InstType = enumor.InstType(strings.ToUpper(data[6]))
		detail.RsIp = data[7]
		rsPort, err := parsePort(data[8])
		if err != nil {
			return err
		}
		detail.RsPort = rsPort
		weight, err := strconv.Atoi(strings.TrimSpace(data[9]))
		if err != nil {
			return err
		}
		detail.Weight = weight
		if len(data) > layer7listenerBindRSExcelTableLen {
			detail.UserRemark = data[10]
		}

		l.details = append(l.details, detail)
	}
	return nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validate(kt *kit.Kit) error {
	recordMap := make(map[string]int)
	clbIDMap := make(map[string]struct{})
	for cur, detail := range l.details {
		detail.validate()
		// 检查记录是否重复
		key := fmt.Sprintf("%s-%s-%v-%s-%s-%s-%v",
			detail.CloudClbID, detail.Protocol, detail.ListenerPort,
			detail.Domain, detail.URLPath, detail.RsIp, detail.RsPort)
		if i, ok := recordMap[key]; ok {
			l.details[i].Status.SetNotExecutable()
			l.details[i].ValidateResult = append(l.details[i].ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d;", i+1))

			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(detail.ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d;", cur+1))
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

func (l *Layer7ListenerBindRSPreviewExecutor) validateWithDB(kt *kit.Kit, cloudIDs []string) error {
	lbMap, err := getLoadBalancersMapByCloudID(kt, l.dataServiceCli, l.vendor, l.accountID, l.bkBizID, cloudIDs)
	if err != nil {
		return err
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
			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(detail.ValidateResult,
				fmt.Sprintf("clb vip(%s) not match", detail.ClbVipDomain))
			continue
		}
		detail.RegionID = lb.Region

		lblCloudID, err := l.validateListener(kt, detail)
		if err != nil {
			logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		ruleCloudID, err := l.validateURLRule(kt, lb.CloudID, lblCloudID, detail)
		if err != nil {
			logs.Errorf("validate url rule failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		instID, err := l.validateRS(kt, detail, lb.ID)
		if err != nil {
			logs.Errorf("validate rs failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if err = l.validateTarget(kt, detail, ruleCloudID, instID, detail.RsPort[0]); err != nil {
			logs.Errorf("validate target failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

// validateTarget 校验RS是否已经绑定到对应的监听器中, 如果已经绑定则校验权重是否一致. 没有绑定则直接返回.
func (l *Layer7ListenerBindRSPreviewExecutor) validateTarget(kt *kit.Kit, detail *Layer7ListenerBindRSDetail,
	ruleCloudID, instID string, port int) error {

	if ruleCloudID == "" || instID == "" {
		return nil
	}
	tgID, err := getTargetGroupID(kt, l.dataServiceCli, ruleCloudID)
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
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult,
			fmt.Sprintf("RS is already bound, and the weights are inconsistent."))
		return nil
	}

	detail.Status.SetExisting()
	detail.ValidateResult = append(detail.ValidateResult, fmt.Sprintf("RS is already bound"))
	return nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validateRS(kt *kit.Kit,
	curDetail *Layer7ListenerBindRSDetail, lbID string) (string, error) {

	if curDetail.InstType == enumor.EniInstType {
		// ENI 不做校验
		return "", nil
	}

	lb, err := getTCloudLoadBalancer(kt, l.dataServiceCli, lbID)
	if err != nil {
		return "", err
	}
	cloudVpcIDs := []string{lb.CloudVpcID}
	isSnap := converter.PtrToVal(lb.Extension.Snat)
	isSnapPro := converter.PtrToVal(lb.Extension.SnatPro)
	if isSnap {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(lb.Extension.TargetCloudVpcID))
	}

	cvm, err := getCvm(kt, l.dataServiceCli, curDetail.RsIp, l.vendor, l.bkBizID, l.accountID, cloudVpcIDs)
	if err != nil {
		return "", err
	}
	if cvm == nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("rs(%s) not exist", curDetail.RsIp))
		return "", nil
	}
	if !(isSnap || isSnapPro) && cvm.Region != lb.Region {
		// 非跨域情况下才校验region
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("rs(%s) region not match, rs.region: %s, lb.region: %v",
				curDetail.RsIp, cvm.Region, lb.Region))
		return cvm.CloudID, nil
	}

	return cvm.CloudID, nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validateListener(kt *kit.Kit,
	curDetail *Layer7ListenerBindRSDetail) (string, error) {

	listener, err := getListener(kt, l.dataServiceCli, l.accountID, curDetail.CloudClbID, curDetail.Protocol,
		curDetail.ListenerPort[0], l.bkBizID, l.vendor)
	if err != nil {
		return "", err
	}
	if listener == nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, "listener not found")
		return "", nil
	}
	if listener.Protocol != curDetail.Protocol {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("Listener protocol does not match. input(%s) actual(%s)",
				curDetail.Protocol, listener.Protocol))
		return "", nil
	}

	return listener.CloudID, nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validateURLRule(kt *kit.Kit, lbCloudID, lblCloudID string,
	detail *Layer7ListenerBindRSDetail) (string, error) {

	rule, err := getURLRule(kt, l.dataServiceCli, l.vendor, lbCloudID, lblCloudID, detail.Domain, detail.URLPath)
	if err != nil {
		logs.Errorf("get url rule failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	if rule == nil {
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult, "url rule not found")
		return "", nil
	}
	return rule.CloudID, nil
}

// Layer7ListenerBindRSDetail ...
type Layer7ListenerBindRSDetail struct {
	ClbVipDomain string              `json:"clb_vip_domain"`
	CloudClbID   string              `json:"cloud_clb_id"`
	Protocol     enumor.ProtocolType `json:"protocol"`
	ListenerPort []int               `json:"listener_port"`
	Domain       string              `json:"domain"`
	URLPath      string              `json:"url_path"`

	InstType       enumor.InstType `json:"inst_type"`
	RsIp           string          `json:"rs_ip"`
	RsPort         []int           `json:"rs_port"`
	Weight         int             `json:"weight"`
	UserRemark     string          `json:"user_remark"`
	Status         ImportStatus    `json:"status"`
	ValidateResult []string        `json:"validate_result"`

	RegionID string `json:"region_id"`
}

func (c *Layer7ListenerBindRSDetail) validate() {
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
		err = errors.New("the RS weight of the port segment must be greater than 0")
		return
	}
	if len(c.Domain) == 0 || len(c.URLPath) == 0 {
		err = errors.New("domain and url are required")
		return
	}
}
