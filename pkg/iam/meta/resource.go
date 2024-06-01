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

package meta

// ResourceType 表示 hcm 这一侧的资源类型， 对应的有 client.TypeID 表示 iam 一侧的资源类型
// 两者之间有映射关系，详情见 AdaptAuthOptions
type ResourceType string

// String convert ResourceType to string.
func (r ResourceType) String() string {
	return string(r)
}

const (
	// Account defines cloud account resource's hcm auth resource type
	Account ResourceType = "account"
	// SubAccount defines cloud sub account resource's hcm auth resource type
	SubAccount ResourceType = "sub_account"
	// Vpc defines vpc hcm auth resource type
	Vpc ResourceType = "vpc"
	// Subnet defines subnet hcm auth resource type
	Subnet ResourceType = "subnet"
	// Disk defines disk hcm auth resource type
	Disk ResourceType = "disk"
	// SecurityGroup defines security group's hcm auth resource type
	SecurityGroup ResourceType = "security_group"
	// SecurityGroupRule defines security group rule's hcm auth resource type
	SecurityGroupRule ResourceType = "security_group_rule"
	// GcpFirewallRule defines gcp firewall rule's hcm auth resource type
	GcpFirewallRule ResourceType = "gcp_firewall_rule"
	// Eip defines eip hcm auth resource type
	Eip ResourceType = "eip"
	// Cvm defines cvm hcm auth resource type
	Cvm ResourceType = "cvm"
	// RouteTable defines route table's hcm auth resource type
	RouteTable ResourceType = "route_table"
	// Route defines route's hcm auth resource type
	Route ResourceType = "route"
	// RecycleBin defines recycle bin's hcm auth resource type
	RecycleBin ResourceType = "recycle_bin"
	// NetworkInterface defines eip hcm network_interface resource type
	NetworkInterface ResourceType = "network_interface"
	// Audit defines audit log's hcm auth resource type
	Audit ResourceType = "biz_audit"
	// Biz defines biz's hcm auth resource type
	Biz ResourceType = "biz"
	// CloudResource is a special resource type that contains all cloud resource.
	CloudResource ResourceType = "cloud_resource"
	// Quota 配额
	Quota ResourceType = "quota"
	// InstanceType 机型
	InstanceType ResourceType = "instance_type"
	// CostManage defines cost manage's hcm auth resource type
	CostManage ResourceType = "cost_manage"
	// BizCollection 业务收藏
	BizCollection ResourceType = "biz_collection"
	// CloudSelectionScheme 云选型方案
	CloudSelectionScheme ResourceType = "cloud_selection_scheme"
	// CloudSelectionIdc 云选型机房
	CloudSelectionIdc ResourceType = "cloud_selection_idc"
	// CloudSelectionBizType 云选型业务类型
	CloudSelectionBizType ResourceType = "cloud_selection_biz_type"
	// CloudSelectionDataSource 云选型数据
	CloudSelectionDataSource ResourceType = "cloud_selection_data_source"
	// ArgumentTemplate 参数模版
	ArgumentTemplate ResourceType = "argument_template"
	// Cert defines cert hcm auth resource type
	Cert ResourceType = "cert"
	// LoadBalancer defines clb hcm auth resource type
	LoadBalancer ResourceType = "load_balancer"
	// Listener defines listener hcm auth resource type
	Listener ResourceType = "listener"
	// TargetGroup defines target group hcm auth resource type
	TargetGroup ResourceType = "target_group"
	// UrlRuleAuditResType url规则
	UrlRuleAuditResType ResourceType = "url_rule"
	// MainAccount defines main cloud account resource's hcm auth resource type
	MainAccount ResourceType = "main_account"
	// RootAccount defines main cloud account resource's hcm auth resource type
	RootAccount ResourceType = "root_account"
)
