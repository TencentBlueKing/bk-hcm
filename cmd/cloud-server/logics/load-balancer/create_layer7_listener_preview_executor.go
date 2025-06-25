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
	"fmt"
	"reflect"
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

var _ ImportPreviewExecutor = (*CreateLayer7ListenerPreviewExecutor)(nil)

func newCreateLayer7ListenerPreviewExecutor(cli *dataservice.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) *CreateLayer7ListenerPreviewExecutor {

	return &CreateLayer7ListenerPreviewExecutor{
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateLayer7ListenerPreviewExecutor 创建七层监听器预览执行器
type CreateLayer7ListenerPreviewExecutor struct {
	*basePreviewExecutor

	details []*CreateLayer7ListenerDetail
}

// Execute 导入执行器的唯一入口
func (c *CreateLayer7ListenerPreviewExecutor) Execute(kt *kit.Kit, rawData [][]string) (interface{}, error) {
	err := c.convertDataToPreview(rawData)
	if err != nil {
		logs.Errorf("convert create listener layer7 data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = c.validate(kt)
	if err != nil {
		logs.Errorf("validate create listener layer7 data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return c.details, nil
}

var createLayer7ListenerDetailRefType = reflect.TypeOf(CreateLayer7ListenerDetail{})

// createLayer7ListenerExcelTableLen excel表格最小应填的列数, clbIP, clbCloudID, protocol, listenerPorts
const createLayer7ListenerExcelTableLen = 4

func (c *CreateLayer7ListenerPreviewExecutor) convertDataToPreview(rawData [][]string) error {
	for i, data := range rawData {
		data = trimSpaceForSlice(data)
		if len(data) < createLayer7ListenerExcelTableLen {
			return fmt.Errorf("line[%d] data length less than %d, got: %d, data: %v",
				i+excelTableLineNumberOffset, createLayer7ListenerExcelTableLen, len(data), data)
		}
		detail := &CreateLayer7ListenerDetail{
			ValidateResult: make([]string, 0),
		}
		for i, value := range data {
			field := createLayer7ListenerDetailRefType.Field(i)
			fieldValue := reflect.ValueOf(detail).Elem().FieldByName(field.Name)
			if fieldValue.Type().Kind() == reflect.String {
				fieldValue.SetString(value)
				continue
			}
			switch field.Name {
			case "ListenerPorts":
				ports, err := parsePort(value)
				if err != nil {
					return err
				}
				detail.ListenerPorts = ports
			case "CertCloudIDs":
				certStr := strings.TrimRight(strings.TrimLeft(value, "["), "]")
				for _, s := range strings.Split(certStr, ",") {
					cur := strings.TrimSpace(s)
					if len(cur) == 0 {
						continue
					}
					detail.CertCloudIDs = append(detail.CertCloudIDs, cur)
				}
			}
		}

		detail.Protocol = enumor.ProtocolType(strings.ToUpper(string(detail.Protocol)))
		c.details = append(c.details, detail)
	}

	return nil
}

func (c *CreateLayer7ListenerPreviewExecutor) validate(kt *kit.Kit) error {

	if len(c.details) == 0 {
		return fmt.Errorf("there are no details to be executed")
	}
	recordMap := make(map[string]int)
	clbIDMap := make(map[string]struct{})
	for cur, detail := range c.details {
		detail.validate()
		// 检查记录是否重复, 重复则设置状态为不可执行
		key := fmt.Sprintf("%s-%s-%v", detail.CloudClbID, detail.Protocol, detail.ListenerPorts)
		if i, ok := recordMap[key]; ok {
			c.details[i].Status.SetNotExecutable()
			c.details[i].ValidateResult = append(c.details[i].ValidateResult,
				fmt.Sprintf("Duplicate records exist, line: %d;", i+1))

			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(detail.ValidateResult,
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

func (c *CreateLayer7ListenerPreviewExecutor) validateWithDB(kt *kit.Kit, cloudIDs []string) error {
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
				fmt.Sprintf("clb.vip(%s) not match", detail.ClbVipDomain))
		}
		detail.RegionID = lb.Region

		if err = c.validateListener(kt, detail); err != nil {
			logs.Errorf("validate listener failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

func (c *CreateLayer7ListenerPreviewExecutor) validateListener(kt *kit.Kit, detail *CreateLayer7ListenerDetail) error {

	switch c.vendor {
	case enumor.TCloud:
		listeners, err := c.getTCloudListenersByPort(kt, detail.CloudClbID, detail.ListenerPorts[0])
		if err != nil {
			return err
		}
		if len(listeners) == 0 {
			return nil
		}
		return c.validateTCloudListener(kt, detail, listeners)
	default:
		return fmt.Errorf("vendor(%s) not support", c.vendor)
	}
}

func (c *CreateLayer7ListenerPreviewExecutor) validateTCloudListener(kt *kit.Kit, detail *CreateLayer7ListenerDetail,
	listeners []corelb.Listener[corelb.TCloudListenerExtension]) error {

	for _, listener := range listeners {
		if listener.Protocol == enumor.UdpProtocol {
			// udp协议可以和 http/https共用端口
			continue
		}

		detail.Status.SetExisting()

		if listener.SniSwitch == enumor.SniTypeOpen {
			// sni开启时, 将不做重复检查, 直接返回不可执行
			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(detail.ValidateResult,
				fmt.Sprintf("clb(%s) listener(%d) already exist, and SNI is enable;", listener.CloudLbID, detail.ListenerPorts[0]))
			return nil
		}

		if listener.Protocol != detail.Protocol {
			detail.Status.SetNotExecutable()
			detail.ValidateResult = append(detail.ValidateResult,
				fmt.Sprintf("clb(%s) listener(%d) already exist, and the protocol does not match",
					listener.CloudLbID, detail.ListenerPorts[0]))
			return nil
		}

		if detail.Protocol == enumor.HttpProtocol {
			return nil
		}

		err := c.validateTCloudCert(kt, listener.Extension.Certificate, detail, listener.CloudLbID, listener.ID)
		if err != nil {
			return err
		}

		err = c.validateCert(kt, detail)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CreateLayer7ListenerPreviewExecutor) validateTCloudCert(kt *kit.Kit,
	listenerCert *corelb.TCloudCertificateInfo, curDetail *CreateLayer7ListenerDetail,
	cloudLBID, lblID string) error {

	if listenerCert == nil {
		logs.Errorf("listener(%s) cert is nil, rid: %s", lblID, kt.Rid)
		return fmt.Errorf("listner(%s) cert is nil", lblID)
	}
	if converter.PtrToVal(listenerCert.SSLMode) != curDetail.SSLMode ||
		converter.PtrToVal(listenerCert.CaCloudID) != curDetail.CACloudID ||
		len(listenerCert.CertCloudIDs) != len(curDetail.CertCloudIDs) {

		curDetail.Status.SetNotExecutable()
		curDetail.ValidateResult = append(curDetail.ValidateResult,
			fmt.Sprintf("clb(%s) listener(%d) already exist, and the cert info dose not match",
				cloudLBID, curDetail.ListenerPorts[0]))
		return nil
	}

	for _, id := range curDetail.CertCloudIDs {
		if !slice.IsItemInSlice(listenerCert.CertCloudIDs, id) {
			curDetail.Status.SetNotExecutable()
			curDetail.ValidateResult = append(curDetail.ValidateResult,
				fmt.Sprintf("clb(%s) listener(%d) already exist, and the cert info dose not match",
					cloudLBID, curDetail.ListenerPorts[0]))
			return nil
		}
	}
	return nil
}

func (c *CreateLayer7ListenerPreviewExecutor) validateCert(kt *kit.Kit, curDetail *CreateLayer7ListenerDetail) error {

	cloudIDs := curDetail.CertCloudIDs
	if len(curDetail.CACloudID) > 0 {
		cloudIDs = append(cloudIDs, curDetail.CACloudID)
	}
	if len(cloudIDs) == 0 {
		return nil
	}
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", c.accountID),
			tools.RuleEqual("bk_biz_id", c.bkBizID),
			tools.RuleEqual("vendor", c.vendor),
			tools.RuleIn("cloud_id", cloudIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	certs, err := c.dataServiceCli.Global.ListCert(kt, listReq)
	if err != nil {
		logs.Errorf("list cert failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	certCloudIDs := make(map[string]struct{})
	for _, detail := range certs.Details {
		certCloudIDs[detail.CloudID] = struct{}{}
	}
	for _, cloudID := range cloudIDs {
		if _, ok := certCloudIDs[cloudID]; !ok {
			curDetail.Status.SetNotExecutable()
			curDetail.ValidateResult = append(curDetail.ValidateResult, fmt.Sprintf("cert(%s) not found", cloudID))
		}
	}
	return nil
}

func (c *CreateLayer7ListenerPreviewExecutor) getTCloudListenersByPort(kt *kit.Kit, lbCloudID string, port int) (
	[]corelb.Listener[corelb.TCloudListenerExtension], error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", c.accountID),
			tools.RuleEqual("bk_biz_id", c.bkBizID),
			tools.RuleEqual("cloud_lb_id", lbCloudID),
			tools.RuleEqual("port", port),
			tools.RuleEqual("vendor", c.vendor),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := c.dataServiceCli.TCloud.LoadBalancer.ListListener(kt, req)
	if err != nil {
		logs.Errorf("list listener failed, port: %d, cloudLBID: %s, err: %v, rid: %s",
			port, lbCloudID, err, kt.Rid)
		return nil, err
	}
	if len(resp.Details) > 0 {
		return resp.Details, nil
	}
	return nil, nil
}

// CreateLayer7ListenerDetail 创建七层监听器预览记录
type CreateLayer7ListenerDetail struct {
	ClbVipDomain string `json:"clb_vip_domain"`
	CloudClbID   string `json:"cloud_clb_id"`

	Name          string              `json:"name"`
	Protocol      enumor.ProtocolType `json:"protocol"`
	ListenerPorts []int               `json:"listener_port"`
	SSLMode       string              `json:"ssl_mode"`
	CertCloudIDs  []string            `json:"cert_cloud_ids"`
	CACloudID     string              `json:"ca_cloud_id"`
	UserRemark    string              `json:"user_remark"`

	Status         ImportStatus `json:"status"`
	ValidateResult []string     `json:"validate_result"`

	RegionID string `json:"region_id"`
}

func (c *CreateLayer7ListenerDetail) validate() {
	// 验证监听器名称，校验规则同"CLB名称"
	if len(c.Name) > 60 {
		c.Status.SetNotExecutable()
		c.ValidateResult = append(c.ValidateResult, "The length of the listener name should not exceed 60")
	}

	err := validatePort(c.ListenerPorts)
	if err != nil {
		c.Status.SetNotExecutable()
		c.ValidateResult = append(c.ValidateResult, err.Error())
		return
	}

	switch c.Protocol {
	case enumor.HttpProtocol:
		if len(c.SSLMode) > 0 || len(c.CertCloudIDs) > 0 || len(c.CACloudID) > 0 {
			c.Status.SetNotExecutable()
			c.ValidateResult = append(c.ValidateResult,
				"The HTTP protocol does not support filling in certificate information")
			return
		}
	case enumor.HttpsProtocol:
		if len(c.SSLMode) == 0 || len(c.CertCloudIDs) == 0 {
			c.Status.SetNotExecutable()
			c.ValidateResult = append(c.ValidateResult,
				"The HTTPS protocol must have certificate information filled in")
			return
		}
		if c.SSLMode != "MUTUAL" && c.SSLMode != "UNIDIRECTIONAL" {
			c.Status.SetNotExecutable()
			c.ValidateResult = append(c.ValidateResult, "Certificate authentication method error")
			return
		}
	default:
		c.Status.SetNotExecutable()
		c.ValidateResult = append(c.ValidateResult, "Protocol type error")
		return
	}

	c.Status.SetExecutable()
}
