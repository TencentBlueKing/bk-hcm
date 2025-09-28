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
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
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
				fmt.Sprintf("clb vip(%s)not match", detail.ClbVipDomain))
			continue
		}
		detail.RegionID = lb.Region

		lblCloudID, err := l.validateListener(kt, detail)
		if err != nil {
			logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		instID, err := l.validateRS(kt, detail, lb.ID)
		if err != nil {
			logs.Errorf("validate rs failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if err = l.validateTarget(kt, lb.ID, detail, lblCloudID, instID, detail.RsPort[0]); err != nil {
			logs.Errorf("validate target failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

// validateTarget 校验RS是否已经绑定到对应的监听器中, 如果已经绑定则校验权重是否一致. 没有绑定则直接返回.
func (l *Layer4ListenerBindRSPreviewExecutor) validateTarget(kt *kit.Kit, lbID string,
	detail *Layer4ListenerBindRSDetail, lblCloudID, instID string, port int) error {

	if lblCloudID == "" || instID == "" {
		return nil
	}
	tgID, err := getTargetGroupID(kt, l.dataServiceCli, lbID, lblCloudID)
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

func (l *Layer4ListenerBindRSPreviewExecutor) validateRS(kt *kit.Kit,
	curDetail *Layer4ListenerBindRSDetail, lbID string) (string, error) {

	if curDetail.InstType == enumor.EniInstType {
		// ENI 不做校验
		return "", nil
	}

	var lb *corelb.LoadBalancer[corelb.TCloudClbExtension]
	var err error
	switch l.vendor {
	case enumor.TCloud:
		lb, err = getTCloudLoadBalancer(kt, l.dataServiceCli, lbID)
	default:
		return "", fmt.Errorf("layer4 listener bind rs preview validate, unsupported vendor: %s", l.vendor)
	}
	if err != nil {
		return "", err
	}
	cloudVpcIDs := []string{lb.CloudVpcID}
	isCrossRegionV2 := converter.PtrToVal(lb.Extension.SnatPro)
	if isCrossRegionV2 {
		// 跨域2.0 本地无法校验，因此此处不进行校验，由云上接口判断
		return "", nil
	}
	isCrossRegionV1 := lb.Extension.SupportCrossRegionV1()
	if isCrossRegionV1 {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(lb.Extension.TargetCloudVpcID))
	}

	cvm, err := getCvm(kt, l.dataServiceCli, curDetail.RsIp, l.vendor, l.bkBizID, l.accountID, cloudVpcIDs)
	if err != nil {
		return "", err
	}
	if cvm == nil {
		// 找不到对应的CVM, 根据IP查询CVM完善报错
		return "", l.fillRSValidateCvmNotFoundError(kt, curDetail, lb.CloudVpcID)
	}
	targetRegion := lb.Region
	if isCrossRegionV1 {
		// 跨域1.0 校验 target region
		targetRegion = converter.PtrToVal(lb.Extension.TargetRegion)
	}
	if cvm.Region != targetRegion {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("rs(%s) region not match, rs.region: %s, targetRegion: %v",
				curDetail.RsIp, cvm.Region, targetRegion))
		return cvm.CloudID, nil
	}

	return cvm.CloudID, nil
}

func (l *Layer4ListenerBindRSPreviewExecutor) fillRSValidateCvmNotFoundError(
	kt *kit.Kit, curDetail *Layer4ListenerBindRSDetail, lbCloudVpcID string) error {

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

func (l *Layer4ListenerBindRSPreviewExecutor) validateListener(kt *kit.Kit,
	curDetail *Layer4ListenerBindRSDetail) (string, error) {

	listener, err := getListener(kt, l.dataServiceCli, l.accountID,
		curDetail.CloudClbID, curDetail.Protocol, curDetail.ListenerPort[0], l.bkBizID, l.vendor)
	if err != nil {
		return "", err
	}
	if listener == nil {
		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult, "listener not found")
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
	Weight         *int            `json:"weight"`
	UserRemark     string          `json:"user_remark"`
	Status         ImportStatus    `json:"status"`
	ValidateResult []string        `json:"validate_result"`

	RegionID string `json:"region_id"`
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
