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

// Package table ...
package table

import "fmt"

// Table defines all the database table
// related resources.
type Table interface {
	TableName() Name
}

var validTableNames = make(map[Name]bool)

// Name is database table's name type
type Name string

const (
	// IDGenerator is id generator table's name.
	IDGenerator Name = "id_generator"
	// AuditTable is audit table's name
	AuditTable Name = "audit"
	// RecycleRecordTable is recycle record table name
	RecycleRecordTable Name = "recycle_record"
	// AccountTable is account table's name.
	AccountTable Name = "account"
	// SubAccountTable is sub account table's name.
	SubAccountTable Name = "sub_account"
	// AccountBizRelTable is account and biz relation table's name.
	AccountBizRelTable Name = "account_biz_rel"
	// SecurityGroupTable is security group table's name.
	SecurityGroupTable Name = "security_group"
	// VpcSecurityGroupRelTable is vpc and security group table's name.
	VpcSecurityGroupRelTable Name = "vpc_security_group_rel"
	// SecurityGroupTagTable is security group tag table's name.
	SecurityGroupTagTable Name = "security_group_tag"
	// SecurityGroupSubnetTable is security group subnet table's name.
	SecurityGroupSubnetTable Name = "security_group_subnet_rel"
	// SecurityGroupCvmTable is security group cvm table's name.
	SecurityGroupCvmTable Name = "security_group_cvm_rel"
	// SGSecurityGroupRuleTable is security group and rule rel table's name.
	SGSecurityGroupRuleTable = "security_group_security_group_rule"
	// TCloudSecurityGroupRuleTable is tcloud security group rule table's name.
	TCloudSecurityGroupRuleTable = "tcloud_security_group_rule"
	// AwsSecurityGroupRuleTable is aws security group rule table's name.
	AwsSecurityGroupRuleTable = "aws_security_group_rule"
	// HuaWeiSecurityGroupRuleTable is huawei security group rule table's name.
	HuaWeiSecurityGroupRuleTable = "huawei_security_group_rule"
	// AzureSecurityGroupRuleTable is azure security group rule table's name.
	AzureSecurityGroupRuleTable = "azure_security_group_rule"
	// SGNetworkInterfaceRelTable is security group and network interface rel table's name.
	SGNetworkInterfaceRelTable = "security_group_network_interface_rel"
	// GcpFirewallRuleTable is gcp firewall rule table's name.
	GcpFirewallRuleTable = "gcp_firewall_rule"
	// VpcTable is vpc table's name.
	VpcTable Name = "vpc"
	// SubnetTable is subnet table's name.
	SubnetTable Name = "subnet"
	// HuaWeiRegionTable is huawei region table's name.
	HuaWeiRegionTable Name = "huawei_region"
	// AzureRGTable is azure resource group table's name.
	AzureRGTable Name = "azure_resource_group"
	// AzureRegionTable is azure region table's name.
	AzureRegionTable Name = "azure_region"
	// DiskTable is disk table's name.
	DiskTable Name = "disk"
	// TCloudRegionTable is tcloud region table's name.
	TCloudRegionTable Name = "tcloud_region"
	// AwsRegionTable is aws region table's name.
	AwsRegionTable Name = "aws_region"
	// GcpRegionTable is gcp region table's name.
	GcpRegionTable Name = "gcp_region"
	// EipTable is eip table's name
	EipTable Name = "eip"
	// ImageTable is image table's name
	ImageTable Name = "image"
	// ZoneTable is zone table's name.
	ZoneTable Name = "zone"
	// CvmTable is cvm table's name.
	CvmTable Name = "cvm"
	// RouteTableTable is route table's table name.
	RouteTableTable Name = "route_table"
	// TCloudRouteTable is tcloud route's table name.
	TCloudRouteTable Name = "tcloud_route"
	// AwsRouteTable is aws route's table name.
	AwsRouteTable Name = "aws_route"
	// AzureRouteTable is azure route's table name.
	AzureRouteTable Name = "azure_route"
	// HuaWeiRouteTable is huawei route's table name.
	HuaWeiRouteTable Name = "huawei_route"
	// GcpRouteTable is gcp route's table name.
	GcpRouteTable Name = "gcp_route"
	// DiskCvmRelTableName is disk_cvm_rel's table name.
	DiskCvmRelTableName Name = "disk_cvm_rel"
	// EipCvmRelTableName is eip_cvm_rel's table name.
	EipCvmRelTableName Name = "eip_cvm_rel"

	// AccountSyncDetailTable is account_sync_detail table's name.
	AccountSyncDetailTable Name = "account_sync_detail"

	// ApplicationTable is application table name
	ApplicationTable Name = "application"
	// ApprovalProcessTable is approval process table name
	ApprovalProcessTable Name = "approval_process"
	// NetworkInterfaceTable is network interface table's name.
	NetworkInterfaceTable Name = "network_interface"
	// NetworkInterfaceCvmRelTable is network interface and cvm rel table's name.
	NetworkInterfaceCvmRelTable Name = "network_interface_cvm_rel"
	// AccountBillConfigTable is account bill config table's name.
	AccountBillConfigTable Name = "account_bill_config"

	// RecycleRecordTableTaskID is recycle record table's task id.
	// TODO: 之后考虑非表id的id_generator如何更优雅的使用
	RecycleRecordTableTaskID Name = "recycle_record_task_id"

	// UserCollectionTable 用户收藏表
	UserCollectionTable Name = "user_collection"

	// AsyncFlowTable is async flow table's name.
	AsyncFlowTable Name = "async_flow"
	// AsyncFlowTaskTable is async flow task table's name.
	AsyncFlowTaskTable Name = "async_flow_task"

	// CloudSelectionSchemeTable is cloud selection scheme table's name.
	CloudSelectionSchemeTable Name = "cloud_selection_scheme"
	// CloudSelectionBizTypeTable 云选型业务类型
	CloudSelectionBizTypeTable Name = "cloud_selection_biz_type"
	// CloudSelectionIdcTable 云选型机房信息
	CloudSelectionIdcTable Name = "cloud_selection_idc"

	// ArgumentTemplateTable is argument template table's name.
	ArgumentTemplateTable Name = "argument_template"
)

// Validate whether the table name is valid or not.
func (n Name) Validate() error {
	valid := validTableNames[n]
	if valid {
		return nil
	}

	if _, ok := TableMap[n]; !ok {
		return fmt.Errorf("unknown table name: %s", n)
	}

	return nil
}

// TableMap table map config
var TableMap = map[Name]struct{}{
	AuditTable:                   {},
	AccountTable:                 {},
	SubAccountTable:              {},
	AccountBizRelTable:           {},
	VpcTable:                     {},
	SubnetTable:                  {},
	IDGenerator:                  {},
	SecurityGroupTable:           {},
	VpcSecurityGroupRelTable:     {},
	SecurityGroupTagTable:        {},
	SecurityGroupSubnetTable:     {},
	SGSecurityGroupRuleTable:     {},
	TCloudSecurityGroupRuleTable: {},
	AwsSecurityGroupRuleTable:    {},
	HuaWeiSecurityGroupRuleTable: {},
	AzureSecurityGroupRuleTable:  {},
	SGNetworkInterfaceRelTable:   {},
	GcpFirewallRuleTable:         {},
	HuaWeiRegionTable:            {},
	AzureRGTable:                 {},
	AzureRegionTable:             {},
	GcpRegionTable:               {},
	AwsRegionTable:               {},
	TCloudRegionTable:            {},
	RouteTableTable:              {},
	TCloudRouteTable:             {},
	AwsRouteTable:                {},
	AzureRouteTable:              {},
	HuaWeiRouteTable:             {},
	GcpRouteTable:                {},
	ZoneTable:                    {},
	CvmTable:                     {},
	ApplicationTable:             {},
	ApprovalProcessTable:         {},
	NetworkInterfaceTable:        {},
	NetworkInterfaceCvmRelTable:  {},
	RecycleRecordTable:           {},
	EipTable:                     {},
	DiskTable:                    {},
	ImageTable:                   {},
	DiskCvmRelTableName:          {},
	EipCvmRelTableName:           {},
	AccountBillConfigTable:       {},
	UserCollectionTable:          {},
	AccountSyncDetailTable:       {},
	CloudSelectionSchemeTable:    {},
	CloudSelectionBizTypeTable:   {},
	CloudSelectionIdcTable:       {},

	// TODO: 临时方案
	RecycleRecordTableTaskID: {},

	AsyncFlowTable:     {},
	AsyncFlowTaskTable: {},

	ArgumentTemplateTable: {},
}

// Register 注册表名
func (n Name) Register() {
	validTableNames[n] = true
}
