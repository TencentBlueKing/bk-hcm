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

/*
	audit.go store audit related enum values.
*/

// AuditResourceType audit resource type.
type AuditResourceType string

const (
	AccountAuditResType           AuditResourceType = "account"
	SecurityGroupAuditResType     AuditResourceType = "security_group"
	SecurityGroupRuleAuditResType AuditResourceType = "security_group_rule"
	VpcCloudAuditResType          AuditResourceType = "vpc"
	SubnetAuditResType            AuditResourceType = "subnet"
	DiskAuditResType              AuditResourceType = "disk"
	CvmAuditResType               AuditResourceType = "cvm"
	RouteTableAuditResType        AuditResourceType = "route_table"
	RouteAuditResType             AuditResourceType = "route"
	EipAuditResType               AuditResourceType = "eip"
	GcpFirewallRuleAuditResType   AuditResourceType = "gcp_firewall_rule"
	NetworkInterfaceAuditResType  AuditResourceType = "network_interface"
)

// AuditResourceTypeEnums resource type map.
var AuditResourceTypeEnums = map[AuditResourceType]struct{}{
	AccountAuditResType:           {},
	SecurityGroupAuditResType:     {},
	SecurityGroupRuleAuditResType: {},
	VpcCloudAuditResType:          {},
	SubnetAuditResType:            {},
	DiskAuditResType:              {},
	CvmAuditResType:               {},
	RouteTableAuditResType:        {},
	EipAuditResType:               {},
	GcpFirewallRuleAuditResType:   {},
	NetworkInterfaceAuditResType:  {},
}

// Exist judge enum value exist.
func (a AuditResourceType) Exist() bool {
	_, exist := AuditResourceTypeEnums[a]
	return exist
}

// AuditAction audit action type.
type AuditAction string

const (
	// Create 创建
	Create AuditAction = "create"
	// Update 更新
	Update AuditAction = "update"
	// Delete 删除
	Delete AuditAction = "delete"
	// Assign 分配
	Assign AuditAction = "assign"
	// Recycle 回收
	Recycle AuditAction = "recycle"
	// Reboot 重启
	Reboot AuditAction = "reboot"
	// Start 开机
	Start AuditAction = "start"
	// Stop 关机
	Stop AuditAction = "stop"
	// ResetPwd 重置密码
	ResetPwd AuditAction = "reset_pwd"
)

// AuditActionEnums op type map.
var AuditActionEnums = map[AuditAction]struct{}{
	Create:   {},
	Update:   {},
	Delete:   {},
	Assign:   {},
	Recycle:  {},
	Reboot:   {},
	Start:    {},
	Stop:     {},
	ResetPwd: {},
}

// Exist judge enum value exist.
func (a AuditAction) Exist() bool {
	_, exist := AuditActionEnums[a]
	return exist
}

// AuditAssignedResType audit assigned resource type.
type AuditAssignedResType string

const (
	BizAuditAssignedResType       AuditAssignedResType = "biz"
	CloudAreaAuditAssignedResType AuditAssignedResType = "cloud_area"
)

// AuditAssignedResTypeEnums audit assigned resource type map.
var AuditAssignedResTypeEnums = map[AuditAssignedResType]struct{}{
	BizAuditAssignedResType:       {},
	CloudAreaAuditAssignedResType: {},
}

// Exist judge enum value exist.
func (a AuditAssignedResType) Exist() bool {
	_, exist := AuditAssignedResTypeEnums[a]
	return exist
}
