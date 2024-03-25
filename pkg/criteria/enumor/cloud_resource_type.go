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

package enumor

import (
	"fmt"

	"hcm/pkg/dal/table"
)

// CloudResourceType defines the cloud resource type.
type CloudResourceType string

var typeMapping = map[CloudResourceType]table.Name{
	AccountCloudResType:          table.AccountTable,
	SubAccountCloudResType:       table.SubAccountTable,
	SecurityGroupCloudResType:    table.SecurityGroupTable,
	GcpFirewallRuleCloudResType:  table.GcpFirewallRuleTable,
	VpcCloudResType:              table.VpcTable,
	SubnetCloudResType:           table.SubnetTable,
	EipCloudResType:              table.EipTable,
	DiskCloudResType:             table.DiskTable,
	CvmCloudResType:              table.CvmTable,
	RouteTableCloudResType:       table.RouteTableTable,
	NetworkInterfaceCloudResType: table.NetworkInterfaceTable,
	ZoneCloudResType:             table.ZoneTable,
	AzureResourceGroup:           table.AzureRGTable,
	ArgumentTemplateResType:      table.ArgumentTemplateTable,
	CertCloudResType:             table.SslCertTable,
	LoadBalancerCloudResType:     table.LoadBalancerTable,
	ListenerCloudResType:         table.LoadBalancerListenerTable,
	TargetGroupCloudResType:      table.LoadBalancerTargetGroupTable,
	TCLoudUrlRuleCloudResType:    table.TCloudLbUrlRuleTable,
}

// ConvTableName conv CloudResourceType to table.Name.
func (rt CloudResourceType) ConvTableName() (table.Name, error) {
	name := table.Name(rt)
	if err := name.Validate(); err == nil {
		return name, nil
	}

	name, exists := typeMapping[rt]
	if !exists {
		return "", fmt.Errorf("%s does not have a corresponding table name", rt)
	}
	return name, nil
}

// CloudResourceType define all cloud resource type.
const (
	AccountCloudResType          CloudResourceType = "account"
	SubAccountCloudResType       CloudResourceType = "sub_account"
	SecurityGroupCloudResType    CloudResourceType = "security_group"
	GcpFirewallRuleCloudResType  CloudResourceType = "gcp_firewall_rule"
	VpcCloudResType              CloudResourceType = "vpc"
	SubnetCloudResType           CloudResourceType = "subnet"
	EipCloudResType              CloudResourceType = "eip"
	CvmCloudResType              CloudResourceType = "cvm"
	DiskCloudResType             CloudResourceType = "disk"
	RouteTableCloudResType       CloudResourceType = "route_table"
	RouteCloudResType            CloudResourceType = "route"
	NetworkInterfaceCloudResType CloudResourceType = "network_interface"
	RegionCloudResType           CloudResourceType = "region"
	ImageCloudResType            CloudResourceType = "image"
	ZoneCloudResType             CloudResourceType = "zone"
	AzureResourceGroup           CloudResourceType = "azure_resource_group"
	ArgumentTemplateResType      CloudResourceType = "argument_template"
	CertCloudResType             CloudResourceType = "cert"
	LoadBalancerCloudResType     CloudResourceType = "load_balancer"
	ListenerCloudResType         CloudResourceType = "listener"
	TargetGroupCloudResType      CloudResourceType = "target_group"
	TCLoudUrlRuleCloudResType    CloudResourceType = "tcloud_url_rule"
)
