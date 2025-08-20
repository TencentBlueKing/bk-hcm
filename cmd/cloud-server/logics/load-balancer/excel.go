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

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	loadbalancer "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/table"
)

var (
	layer4ListenerHeaders [][]string
	layer7ListenerHeaders [][]string
	ruleHeaders           [][]string
	layer4RsHeaders       [][]string
	layer7RsHeaders       [][]string
)

func init() {
	var err error
	layer4ListenerHeaders, err = Layer4ListenerDetail{}.GetHeaders()
	if err != nil {
		logs.Errorf("get layer4 listener headers failed: %v", err)
	}
	layer7ListenerHeaders, err = Layer7ListenerDetail{}.GetHeaders()
	if err != nil {
		logs.Errorf("get layer7 listener headers failed: %v", err)
	}
	ruleHeaders, err = RuleDetail{}.GetHeaders()
	if err != nil {
		logs.Errorf("get rule headers failed: %v", err)
	}
	layer4RsHeaders, err = Layer4RsDetail{}.GetHeaders()
	if err != nil {
		logs.Errorf("get layer4 rs headers failed: %v", err)
	}
	layer7RsHeaders, err = Layer7RsDetail{}.GetHeaders()
	if err != nil {
		logs.Errorf("get layer7 rs headers failed: %v", err)
	}
}

var _ table.Table = (*Layer4ListenerDetail)(nil)

// Layer4ListenerDetail ...
type Layer4ListenerDetail struct {
	ClbVipDomain    string                        `json:"clb_vip_domain" header:"clb_vip/clb_domain;负载均衡vip/域名"`
	CloudClbID      string                        `json:"cloud_clb_id" header:"clb_id;负载均衡云ID"`
	Protocol        enumor.ProtocolType           `json:"protocol" header:"protocol;监听器协议"`
	ListenerPortStr string                        `json:"listener_port_str" header:"listener_port;监听器端口"`
	Scheduler       enumor.Scheduler              `json:"scheduler" header:"scheduler;均衡方式"`
	Session         int                           `json:"session" header:"session(0=disable);会话保持（0为不开启）"`
	HealthCheckStr  enumor.ListenerHealthCheckStr `json:"health_check" header:"health_check;健康检查"`
	Name            string                        `json:"name" header:"listener_name;监听器名称(可选)"`
	UserRemark      string                        `json:"user_remark" header:"user_remark;用户备注(可选)"`
	ExportInfo      string                        `json:"export_info" header:"export_info;导出备注(可选)"`
}

// GetHeaders 获取表头列
func (l Layer4ListenerDetail) GetHeaders() ([][]string, error) {
	return table.GetHeaders(l)
}

// GetValuesByHeader 获取表头对应的数据
func (l Layer4ListenerDetail) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(l)
}

// Layer7ListenerDetail ...
type Layer7ListenerDetail struct {
	ClbVipDomain    string              `json:"clb_vip_domain" header:"clb_vip/clb_domain;负载均衡vip/域名"`
	CloudClbID      string              `json:"cloud_clb_id" header:"clb_id;负载均衡云ID"`
	Protocol        enumor.ProtocolType `json:"protocol" header:"protocol;监听器协议"`
	ListenerPortStr string              `json:"listener_port_str" header:"listener_port;监听器端口"`
	SSLMode         string              `json:"ssl_mode" header:"ssl_mode;证书认证方式"`
	CertCloudID     string              `json:"cert_cloud_id" header:"cert_id;服务器证书"`
	CACloudID       string              `json:"ca_cloud_id" header:"cert_ca_id;客户端证书"`
	Name            string              `json:"name" header:"listener_name;监听器名称(可选)"`
	UserRemark      string              `json:"user_remark" header:"user_remark;用户备注(可选)"`
	ExportInfo      string              `json:"export_info" header:"export_info;导出备注(可选)"`
}

// GetHeaders 获取表头列
func (l Layer7ListenerDetail) GetHeaders() ([][]string, error) {
	return table.GetHeaders(l)
}

// GetValuesByHeader 获取表头对应的数据
func (l Layer7ListenerDetail) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(l)
}

// RuleDetail ...
type RuleDetail struct {
	ClbVipDomain    string                        `json:"clb_vip_domain" header:"clb_vip/clb_domain;负载均衡vip/域名"`
	CloudClbID      string                        `json:"cloud_clb_id" header:"clb_id;负载均衡云ID"`
	Protocol        enumor.ProtocolType           `json:"protocol" header:"protocol;监听器协议"`
	ListenerPortStr string                        `json:"listener_port_str" header:"listener_port;监听器端口"`
	Domain          string                        `json:"domain" header:"domain;域名"`
	DefaultDomain   bool                          `json:"default_domain" header:"default_server;是/否默认域名"`
	UrlPath         string                        `json:"url_path" header:"url_path;url路径"`
	Scheduler       enumor.Scheduler              `json:"scheduler" header:"scheduler;均衡方式"`
	Session         int                           `json:"session" header:"session(0=disable);会话保持（0为不开启）"`
	HealthCheckStr  enumor.ListenerHealthCheckStr `json:"health_check" header:"health_check;健康检查"`
	UserRemark      string                        `json:"user_remark" header:"user_remark;用户备注(可选)"`
	ExportInfo      string                        `json:"export_info" header:"export_info;导出备注(可选)"`
}

// GetHeaders 获取表头列
func (r RuleDetail) GetHeaders() ([][]string, error) {
	return table.GetHeaders(r)
}

// GetValuesByHeader 获取表头对应的数据
func (r RuleDetail) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(r)
}

// Layer4RsDetail ...
type Layer4RsDetail struct {
	ClbVipDomain    string              `json:"clb_vip_domain" header:"clb_vip/clb_domain;负载均衡vip/域名"`
	CloudClbID      string              `json:"cloud_clb_id" header:"clb_id;负载均衡云ID"`
	Protocol        enumor.ProtocolType `json:"protocol" header:"protocol;监听器协议"`
	ListenerPortStr string              `json:"listener_port_str" header:"listener_port;监听器端口"`
	InstType        enumor.InstType     `json:"inst_type" header:"target_type;后端类型"`
	RsIp            string              `json:"rs_ip" header:"rs_ip;rs_ip"`
	RsPortStr       string              `json:"rs_port_str" header:"rs_port;rs_port"`
	Weight          *int64              `json:"weight" header:"weight(0-100);权重(0-100)"`
	UserRemark      string              `json:"user_remark" header:"user_remark;用户备注(可选)"`
	ExportInfo      string              `json:"export_info" header:"export_info;导出备注(可选)"`
}

// GetHeaders 获取表头列
func (l Layer4RsDetail) GetHeaders() ([][]string, error) {
	return table.GetHeaders(l)
}

// GetValuesByHeader 获取表头对应的数据
func (l Layer4RsDetail) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(l)
}

// Layer7RsDetail ...
type Layer7RsDetail struct {
	ClbVipDomain    string              `json:"clb_vip_domain" header:"clb_vip/clb_domain;负载均衡vip/域名"`
	CloudClbID      string              `json:"cloud_clb_id" header:"clb_id;负载均衡云ID"`
	Protocol        enumor.ProtocolType `json:"protocol" header:"protocol;监听器协议"`
	ListenerPortStr string              `json:"listener_port_str" header:"listener_port;监听器端口"`
	Domain          string              `json:"domain" header:"domain;域名"`
	URLPath         string              `json:"url_path" header:"url_path;url路径"`
	InstType        enumor.InstType     `json:"inst_type" header:"target_type;后端类型"`
	RsIp            string              `json:"rs_ip" header:"rs_ip;rs_ip"`
	RsPortStr       string              `json:"rs_port_str" header:"rs_port;rs_port"`
	Weight          *int64              `json:"weight" header:"weight(0-100);权重(0-100)"`
	UserRemark      string              `json:"user_remark" header:"user_remark;用户备注(可选)"`
	ExportInfo      string              `json:"export_info" header:"export_info;导出备注(可选)"`
}

// GetHeaders 获取表头列
func (l Layer7RsDetail) GetHeaders() ([][]string, error) {
	return table.GetHeaders(l)
}

// GetValuesByHeader 获取表头对应的数据
func (l Layer7RsDetail) GetValuesByHeader() ([]string, error) {
	return table.GetValuesByHeader(l)
}

// getFirstRow get first row of excel.
func getFirstRow(vendor enumor.Vendor) ([]string, error) {
	rowData := make([]string, 0)
	rowData = append(rowData, constant.CLBExcelHeaderVendor)
	switch vendor {
	case enumor.TCloud:
		rowData = append(rowData, constant.CLBExcelHeaderTCloud)
	default:
		return nil, fmt.Errorf("unsupported vendor: %v", vendor)
	}

	return rowData, nil
}

// getLbVipOrDomain 当域名存在时，则返回域名，否则返回vip
func getLbVipOrDomain(lb loadbalancer.BaseLoadBalancer) (string, error) {
	if lb.Domain != "" {
		return lb.Domain, nil
	}

	switch typeslb.TCloudLoadBalancerType(lb.LoadBalancerType) {
	case typeslb.InternalLoadBalancerType:
		if lb.IPVersion == enumor.Ipv4 && len(lb.PrivateIPv4Addresses) != 0 {
			return lb.PrivateIPv4Addresses[0], nil
		}
		if len(lb.PrivateIPv6Addresses) != 0 {
			return lb.PrivateIPv6Addresses[0], nil
		}
	case typeslb.OpenLoadBalancerType:
		if lb.IPVersion == enumor.Ipv4 && len(lb.PublicIPv4Addresses) != 0 {
			return lb.PublicIPv4Addresses[0], nil
		}
		if len(lb.PublicIPv6Addresses) != 0 {
			return lb.PublicIPv6Addresses[0], nil
		}
	}

	return "", fmt.Errorf("unsupported lb, cloud id: %s, type: %s, ip version: %s", lb.CloudID, lb.LoadBalancerType,
		lb.IPVersion)
}
