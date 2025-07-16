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

const layer4listenerBindRSExcelTableLen = 8

func (l *Layer4ListenerBindRSPreviewExecutor) convertDataToPreview(rawData [][]string) error {
	for i, data := range rawData {
		data = trimSpaceForSlice(data)

		if len(data) < layer4listenerBindRSExcelTableLen {
			return fmt.Errorf("line[%d] data length less than %d, got: %d, data: %v",
				i+excelTableLineNumberOffset, layer4listenerBindRSExcelTableLen, len(data), data)
		}

		detail := &Layer4ListenerBindRSDetail{
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
		detail.Weight = converter.ValToPtr(weight)
		if len(data) > layer4listenerBindRSExcelTableLen {
			detail.UserRemark = data[8]
		}

		l.details = append(l.details, detail)
	}
	return nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) validate(kt *kit.Kit) error {
	if len(l.details) == 0 {
		return fmt.Errorf("there are no details to be executed")
	}
	recordMap := make(map[string]int)
	clbIDMap := make(map[string]struct{})
	for cur, detail := range l.details {
		detail.validate()
		// 检查记录是否重复
		key := fmt.Sprintf("%s-%s-%v-%s-%v",
			detail.CloudClbID, detail.Protocol, detail.ListenerPort, detail.RsIp, detail.RsPort)
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

func (l *Layer4ListenerBindRSPreviewExecutor) validateWithDB(kt *kit.Kit, cloudIDs []string) error {
	lbMap, err := getLoadBalancersMapByCloudID(kt, l.dataServiceCli, l.vendor, l.accountID, l.bkBizID, cloudIDs)
	if err != nil {
		return err
	}

	concurrentErr := concurrence.BaseExec(cc.CloudServer().ConcurrentConfig.CLBImportCount, l.details,
		func(detail *Layer4ListenerBindRSDetail) error {

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
					fmt.Sprintf("clb vip(%s)not match", detail.ClbVipDomain))
				return nil
			}
			detail.RegionID = lb.Region

			err := l.validateListener(kt, detail)
			if err != nil {
				logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
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
		logs.Errorf("validate with db failed, err: %v, rid: %s", concurrentErr, kt.Rid)
		return err
	}

	if err = l.validateDetailsTarget(kt); err != nil {
		logs.Errorf("validate details target failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) validateDetailsTarget(kt *kit.Kit) error {

	lblCloudIDs := slice.Map(l.details, func(detail *Layer4ListenerBindRSDetail) string {
		return detail.listenerCloudID
	})
	// 在四层监听器中, ruleCloudID等于 listenerCloudID
	ruleCloudIDsToTGIDMap, err := getTargetGroupByRuleCloudIDs(kt, l.dataServiceCli, lblCloudIDs)
	if err != nil {
		logs.Errorf("get target group by rule cloud ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	concurrentErr := concurrence.BaseExec(cc.CloudServer().ConcurrentConfig.CLBImportCount, l.details,
		func(detail *Layer4ListenerBindRSDetail) error {
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
func (l *Layer4ListenerBindRSPreviewExecutor) validateTarget(kt *kit.Kit,
	detail *Layer4ListenerBindRSDetail, ruleCloudIDsToTGIDMap map[string]string) error {

	if detail.listenerCloudID == "" || detail.cvm == nil {
		return nil
	}
	tgID, ok := ruleCloudIDsToTGIDMap[detail.listenerCloudID]
	if !ok {
		return fmt.Errorf("target group not found for listener cloud id: %s", detail.listenerCloudID)
	}
	detail.targetGroupID = tgID
	target, err := getTarget(kt, l.dataServiceCli, tgID, detail.cvm.CloudID, detail.RsPort[0])
	if err != nil {
		return err
	}
	if target == nil {
		return nil
	}

	if int(converter.PtrToVal(target.Weight)) != converter.PtrToVal(detail.Weight) {
		detail.Status.SetNotExecutable()
		detail.ValidateResult = append(detail.ValidateResult,
			fmt.Sprintf("RS is already bound, and the weights are inconsistent."))
		return nil
	}

	detail.Status.SetExisting()
	detail.ValidateResult = append(detail.ValidateResult, fmt.Sprintf("RS is already bound"))

	return nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) validateRS(kt *kit.Kit, curDetail *Layer4ListenerBindRSDetail,
	lb corelb.LoadBalancerRaw) error {

	if curDetail.InstType == enumor.EniInstType {
		// ENI 不做校验
		return nil
	}

	isCrossRegionV1, isCrossRegionV2, _, lbTargetRegion, err := parseSnapInfoTCloudLBExtension(kt,
		lb.Extension)
	if err != nil {
		logs.Errorf("parse snap info for tcloud lb extension failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	cvm, err := validateCvmExist(kt, l.dataServiceCli, curDetail.RsIp, lb,
		isCrossRegionV1, isCrossRegionV2, lbTargetRegion)
	if err != nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, err.Error())
		return nil
	}
	curDetail.cvm = newCvmInfo(cvm)

	targetRegion := lb.Region
	if isCrossRegionV1 {
		// 跨域1.0 校验 extension中的 target region
		targetRegion = lbTargetRegion
	}
	// 支持跨域1.0 校验 target region
	// 支持跨域2.0 不校验
	if !isCrossRegionV2 && cvm.Region != targetRegion {
		// 非跨域情况下才校验region
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("rs(%s) region not match, rs.region: %s, lb.region: %v",
				curDetail.RsIp, cvm.Region, lb.Region))
		return nil
	}

	return nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) validateListener(kt *kit.Kit,
	curDetail *Layer4ListenerBindRSDetail) error {

	listener, err := getListener(kt, l.dataServiceCli, l.accountID,
		curDetail.CloudClbID, curDetail.Protocol, curDetail.ListenerPort[0], l.bkBizID, l.vendor)
	if err != nil {
		return err
	}
	if listener == nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, "listener not found")
		return nil
	}
	curDetail.listenerCloudID = listener.CloudID
	return nil
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
	Weight         *int            `json:"weight"`
	UserRemark     string          `json:"user_remark"`
	Status         ImportStatus    `json:"status"`
	ValidateResult []string        `json:"validate_result"`

	RegionID string `json:"region_id"`

	// targetGroupID 在 validateTarget 阶段填充, 后续submit阶段会重复使用到,
	// 如果为空, 那就意味着当前detail的条件无法匹配到对应的targetGroup, 可以认为targetGroup not found
	targetGroupID string

	// listenerCloudID 在 validateListener 阶段填充, 后续submit阶段会重复使用到,
	// 如果为空, 那就意味着当前detail的条件无法匹配到对应的listener, 可以认为listener not found
	listenerCloudID string

	// cvm 在 validateRS 阶段填充, 在validateTarget和submit阶段会使用,
	// 会有cvm为空的情况, 例如RSType为ENI, 除此之外的情况都应该有cvm, 否则代表了rs not found
	cvm *cvmInfo
}

func (c *Layer4ListenerBindRSDetail) validate() {
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
}
