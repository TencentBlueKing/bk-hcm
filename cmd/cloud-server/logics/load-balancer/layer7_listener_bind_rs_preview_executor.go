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

	cloudCvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
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
func (l *Layer7ListenerBindRSPreviewExecutor) Execute(kt *kit.Kit, rawData [][]string, headers []string) (interface{}, error) {
	err := l.convertDataToPreview(rawData, headers)
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

// layer7listenerBindRSExcelTableHeaderLen 表头长度
const layer7listenerBindRSExcelTableHeaderLen = 12

func (l *Layer7ListenerBindRSPreviewExecutor) convertDataToPreview(rawData [][]string, headers []string) error {
	if len(headers) < layer7listenerBindRSExcelTableHeaderLen {
		return fmt.Errorf("headers length less than %d, got: %d, headers: %v",
			layer7listenerBindRSExcelTableHeaderLen, len(headers), headers)
	}
	for i, data := range rawData {
		data = trimSpaceForSlice(data)
		if len(data) < layer7listenerBindRSExcelTableLen {
			return fmt.Errorf("line[%d] data length less than %d, got: %d, data: %v",
				i+excelTableLineNumberOffset, layer7listenerBindRSExcelTableLen, len(data), data)
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
		weight, err := strconv.ParseInt(strings.TrimSpace(data[9]), 10, 64)
		if err != nil {
			return err
		}
		detail.Weight = converter.ValToPtr(weight)
		if len(data) > layer7listenerBindRSExcelTableLen {
			detail.UserRemark = data[10]
		}

		l.details = append(l.details, detail)
	}
	return nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validate(kt *kit.Kit) error {
	if len(l.details) == 0 {
		return fmt.Errorf("there are no details to be executed")
	}
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

	concurrentErr := concurrence.BaseExec(cc.CloudServer().ConcurrentConfig.CLBImportCount, l.details,
		func(detail *Layer7ListenerBindRSDetail) error {

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
				return nil
			}
			detail.RegionID = lb.Region

			lblCloudID, err := l.validateListener(kt, detail)
			if err != nil {
				logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			err = l.validateURLRule(kt, lb.CloudID, lblCloudID, detail)
			if err != nil {
				logs.Errorf("validate url rule failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			err = l.validateRS(kt, detail, lb)
			if err != nil {
				logs.Errorf("validate rs failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			return nil
		})
	if concurrentErr != nil {
		logs.Errorf("validate concurrent failed, err: %v, rid: %s", concurrentErr, kt.Rid)
		return err
	}
	if err = l.validateDetailsTarget(kt); err != nil {
		logs.Errorf("validate details target failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validateDetailsTarget(kt *kit.Kit) error {
	ruleCloudIDs := slice.Map(l.details, func(detail *Layer7ListenerBindRSDetail) string {
		return detail.urlRuleCloudID
	})
	ruleCloudIDsToTGIDMap, err := getTargetGroupByRuleCloudIDs(kt, l.dataServiceCli, ruleCloudIDs)
	if err != nil {
		logs.Errorf("get target group by rule cloud ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	concurrentErr := concurrence.BaseExec(cc.CloudServer().ConcurrentConfig.CLBImportCount, l.details,
		func(detail *Layer7ListenerBindRSDetail) error {
			if err = l.validateTarget(kt, detail, ruleCloudIDsToTGIDMap); err != nil {
				logs.Errorf("validate target failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			return nil
		})
	if concurrentErr != nil {
		logs.Errorf("validate details target failed, err: %v, rid: %s", concurrentErr, kt.Rid)
		return err
	}
	return nil
}

// validateTarget 校验RS是否已经绑定到对应的监听器中, 如果已经绑定则校验权重是否一致. 没有绑定则直接返回.
func (l *Layer7ListenerBindRSPreviewExecutor) validateTarget(kt *kit.Kit,
	detail *Layer7ListenerBindRSDetail, ruleCloudIDsToTGIDMap map[string]string) error {

	if detail.urlRuleCloudID == "" || detail.cvm == nil {
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult, "url rule not found or rs not found")
		return nil
	}
	tgID, ok := ruleCloudIDsToTGIDMap[detail.urlRuleCloudID]
	if !ok {
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult,
			fmt.Sprintf("target group not found for url rule cloud id: %s", detail.urlRuleCloudID))
		return nil
	}
	detail.targetGroupID = tgID
	target, err := getTarget(kt, l.dataServiceCli, tgID, detail.cvm.CloudID, detail.RsPort[0])
	if err != nil {
		return err
	}
	if target == nil {
		return nil
	}

	if converter.PtrToVal(target.Weight) != converter.PtrToVal(detail.Weight) {
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult,
			fmt.Sprintf("RS is already bound, and the weights are inconsistent."))
		return nil
	}

	detail.Status.SetExisting()
	detail.ValidateResult = append(detail.ValidateResult, fmt.Sprintf("RS is already bound"))
	return nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validateRS(kt *kit.Kit, curDetail *Layer7ListenerBindRSDetail,
	lb corelb.LoadBalancerRaw) error {

	isCrossRegionV1, isCrossRegionV2, targetCloudVpcID, lbTargetRegion, err := parseSnapInfoTCloudLBExtension(kt,
		lb.Extension)
	if err != nil {
		logs.Errorf("parse snap info for tcloud lb extension failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvm, err := validateCvmExist(kt, l.dataServiceCli, curDetail.RsIp, lb,
		isCrossRegionV1, isCrossRegionV2, targetCloudVpcID)
	if err != nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, err.Error())
		return nil
	}
	curDetail.cvm = newCvmInfo(cvm)

	// 支持跨域2.0 不校验
	if isCrossRegionV2 {
		return nil
	}

	targetRegion := lb.Region
	if isCrossRegionV1 {
		// 跨域1.0 校验 extension中的 target region
		targetRegion = lbTargetRegion
	}
	// 支持跨域1.0 校验 target region
	if cvm.Region != targetRegion {
		// 非跨域情况下才校验region
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("rs(%s) region not match, rs.region: %s, lb.region: %v",
				curDetail.RsIp, cvm.Region, lb.Region))
		return nil
	}
	return nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) fillRSValidateCvmNotFoundError(
	kt *kit.Kit, curDetail *Layer7ListenerBindRSDetail, lbCloudVpcID string) error {

	// 找不到对应的CVM, 根据IP查询CVM完善报错
	cvmList, err := getCvmWithoutVpc(kt, l.dataServiceCli, curDetail.RsIp, l.vendor, l.bkBizID, l.accountID)
	if err != nil {
		logs.Errorf("get cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if len(cvmList) == 0 {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, fmt.Sprintf("rs ip(%s) not found",
			curDetail.RsIp))
		return nil
	}

	cvmCloudIDs := slice.Map(cvmList, cloudCvm.BaseCvm.GetCloudID)
	curDetail.Status.SetNotExecutable()
	curDetail.ValidateResult = append(curDetail.ValidateResult,
		fmt.Sprintf("VPC of %s is different from loadbalancer's VPC (%s).",
			strings.Join(cvmCloudIDs, ","), lbCloudVpcID))
	return nil
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

	curDetail.listenerCloudID = listener.CloudID
	return listener.CloudID, nil
}

func (l *Layer7ListenerBindRSPreviewExecutor) validateURLRule(kt *kit.Kit, lbCloudID, lblCloudID string,
	detail *Layer7ListenerBindRSDetail) error {

	rule, err := getURLRule(kt, l.dataServiceCli, l.vendor, lbCloudID, lblCloudID, detail.Domain, detail.URLPath)
	if err != nil {
		logs.Errorf("get url rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if rule == nil {
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult, "url rule not found")
		return nil
	}
	detail.urlRuleCloudID = rule.CloudID
	return nil
}

// Layer7ListenerBindRSDetail ...
type Layer7ListenerBindRSDetail struct {
	Layer7RsDetail `json:",inline"`
	ListenerPort   []int `json:"listener_port"`
	RsPort         []int `json:"rs_port"`

	Status         ImportStatus `json:"status"`
	ValidateResult []string     `json:"validate_result"`

	RegionID string `json:"region_id"`

	// listenerCloudID 在 validateListener 阶段填充, 后续submit阶段会重复使用到,
	// 如果为空, 那就意味着当前detail的条件无法匹配到对应的listener, 可以认为listener not found
	listenerCloudID string
	// urlRuleCloudID 在 validateURLRule 阶段填充, 后续submit阶段会重复使用到,
	// 如果为空, 那就意味着当前detail的条件无法匹配到对应的URL rule, 可以认为url rule not found
	urlRuleCloudID string
	// targetGroupID 在 validateTarget 阶段填充, 后续submit阶段会重复使用到,
	// 如果为空, 那就意味着当前detail的条件无法匹配到对应的targetGroup, 可以认为targetGroup not found
	targetGroupID string
	// cvm 在 validateRS 阶段填充, 在validateTarget和submit阶段会使用,
	// 如果为空, 代表了rs not found
	cvm *cvmInfo
}

type cvmInfo struct {
	CloudID              string
	Name                 string
	PrivateIPv4Addresses []string
	PublicIPv4Addresses  []string
	Zone                 string
}

func newCvmInfo(one *cloudCvm.BaseCvm) *cvmInfo {
	return &cvmInfo{
		CloudID:              one.CloudID,
		Name:                 one.Name,
		PrivateIPv4Addresses: one.PrivateIPv4Addresses,
		PublicIPv4Addresses:  one.PublicIPv4Addresses,
		Zone:                 one.Zone,
	}
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
	if len(c.RsPort) == 2 && converter.PtrToVal(c.Weight) == 0 {
		err = errors.New("the RS weight of the port segment must be greater than 0")
		return
	}
	if len(c.Domain) == 0 || len(c.URLPath) == 0 {
		err = errors.New("domain and url are required")
		return
	}
}
