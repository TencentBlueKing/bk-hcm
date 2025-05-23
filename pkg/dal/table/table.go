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

	// SslCertTable is ssl cert table's name.
	SslCertTable Name = "ssl_cert"

	// LoadBalancerTable is load_balancer table's name.
	LoadBalancerTable Name = "load_balancer"
	// SecurityGroupCommonRelTable is security group common rel table's name.
	SecurityGroupCommonRelTable Name = "security_group_common_rel"
	// LoadBalancerListenerTable is load_balancer_listener table's name.
	LoadBalancerListenerTable Name = "load_balancer_listener"
	// TCloudLbUrlRuleTable is tcloud_lb_url_rule table's name.
	TCloudLbUrlRuleTable Name = "tcloud_lb_url_rule"
	// LoadBalancerTargetTable is load_balancer_target table's name.
	LoadBalancerTargetTable Name = "load_balancer_target"
	// LoadBalancerTargetGroupTable is load_balancer_target_group table's name.
	LoadBalancerTargetGroupTable Name = "load_balancer_target_group"
	// TargetGroupListenerRuleRelTable is target_group_listener_rule_rel table's name.
	TargetGroupListenerRuleRelTable Name = "target_group_listener_rule_rel"
	// ResourceFlowRelTable is resource_flow_rel table's name.
	ResourceFlowRelTable Name = "resource_flow_rel"
	// ResourceFlowLockTable is resource_flow_lock table's name.
	ResourceFlowLockTable Name = "resource_flow_lock"

	// MainAccountTable is main_account table's name
	MainAccountTable Name = "main_account"
	// RootAccountTable is main_account table's name
	RootAccountTable Name = "root_account"

	// AccountBillSummaryVersionTable 月度汇总账单版本
	AccountBillSummaryVersionTable = "account_bill_summary_version"
	// AccountBillSummaryDailyTable 每天汇总账单版本
	AccountBillSummaryDailyTable = "account_bill_summary_daily"
	// AccountBillItemTable 分账后的账单明细
	AccountBillItemTable = "account_bill_item"
	// AccountBillAdjustmentItemTable 手动调账表
	AccountBillAdjustmentItemTable = "account_bill_adjustment_item"
	// AccountBillMonthTaskTable 月度任务表
	AccountBillMonthTaskTable = "account_bill_month_task"
	// AccountBillDailyPullTaskTable 日账单拉取任务表
	AccountBillDailyPullTaskTable = "account_bill_daily_pull_task"
	// AccountBillSummaryRootTable 一级账号账单汇总信息
	AccountBillSummaryRootTable = "account_bill_summary_root"
	// AccountBillSummaryMainTable 月度汇总账单
	AccountBillSummaryMainTable = "account_bill_summary_main"
	// RootAccountBillConfigTable 一级账号账单配置表
	RootAccountBillConfigTable = "root_account_bill_config"
	// AccountBillExchangeRateTable 账单汇率换算表
	AccountBillExchangeRateTable = "account_bill_exchange_rate"
	// AccountBillSyncRecordTable 账单同步记录
	AccountBillSyncRecordTable = "account_bill_sync_record"
	// TaskDetailTable 任务详情表
	TaskDetailTable = "task_detail"
	// TenantTable 租户表
	TenantTable = "tenant"
	// TaskManagementTable 任务管理表
	TaskManagementTable = "task_management"
	//	GlobalConfigTable 全局配置表
	GlobalConfigTable = "global_config"

	// ResUsageBizRelTable 资源-使用业务关联表
	ResUsageBizRelTable = "res_usage_biz_rel"
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

// TableConfig 存储表名和表配置的信息
type TableConfig struct {
	EnableTenant bool // 是否需要支持多租户
}

// TableMap table map config
// Key是表名，Value是该表的配置信息
var TableMap = map[Name]TableConfig{
	AuditTable:                   {EnableTenant: true},
	AccountTable:                 {EnableTenant: true},
	SubAccountTable:              {},
	AccountBizRelTable:           {},
	VpcTable:                     {EnableTenant: true},
	SubnetTable:                  {EnableTenant: true},
	IDGenerator:                  {},
	SecurityGroupTable:           {EnableTenant: true},
	VpcSecurityGroupRelTable:     {},
	SecurityGroupTagTable:        {},
	SecurityGroupSubnetTable:     {},
	SGSecurityGroupRuleTable:     {},
	TCloudSecurityGroupRuleTable: {},
	AwsSecurityGroupRuleTable:    {},
	HuaWeiSecurityGroupRuleTable: {EnableTenant: true},
	AzureSecurityGroupRuleTable:  {},
	SGNetworkInterfaceRelTable:   {},
	GcpFirewallRuleTable:         {EnableTenant: true},
	HuaWeiRegionTable:            {EnableTenant: true},
	AzureRGTable:                 {EnableTenant: true},
	AzureRegionTable:             {EnableTenant: true},
	GcpRegionTable:               {EnableTenant: true},
	AwsRegionTable:               {},
	TCloudRegionTable:            {EnableTenant: true},
	RouteTableTable:              {EnableTenant: true},
	TCloudRouteTable:             {EnableTenant: true},
	AwsRouteTable:                {EnableTenant: true},
	AzureRouteTable:              {EnableTenant: true},
	HuaWeiRouteTable:             {EnableTenant: true},
	GcpRouteTable:                {EnableTenant: true},
	ZoneTable:                    {EnableTenant: true},
	CvmTable:                     {EnableTenant: true},
	ApplicationTable:             {EnableTenant: true},
	ApprovalProcessTable:         {EnableTenant: true},
	NetworkInterfaceTable:        {EnableTenant: true},
	NetworkInterfaceCvmRelTable:  {},
	RecycleRecordTable:           {EnableTenant: true},
	EipTable:                     {EnableTenant: true},
	DiskTable:                    {EnableTenant: true},
	ImageTable:                   {EnableTenant: true},
	DiskCvmRelTableName:          {},
	EipCvmRelTableName:           {},
	AccountBillConfigTable:       {},
	UserCollectionTable:          {EnableTenant: true},
	AccountSyncDetailTable:       {},
	CloudSelectionSchemeTable:    {EnableTenant: true},
	CloudSelectionBizTypeTable:   {EnableTenant: true},
	CloudSelectionIdcTable:       {EnableTenant: true},
	SslCertTable:                 {EnableTenant: true},

	// TODO: 临时方案
	RecycleRecordTableTaskID: {},

	AsyncFlowTable:     {},
	AsyncFlowTaskTable: {},

	ArgumentTemplateTable: {EnableTenant: true},

	AccountBillMonthTaskTable:       {},
	AccountBillDailyPullTaskTable:   {},
	AccountBillSummaryMainTable:     {},
	AccountBillSummaryVersionTable:  {},
	AccountBillSummaryDailyTable:    {},
	AccountBillItemTable:            {},
	AccountBillAdjustmentItemTable:  {},
	AccountBillSummaryRootTable:     {},
	RootAccountBillConfigTable:      {EnableTenant: true},
	AccountBillExchangeRateTable:    {EnableTenant: true},
	AccountBillSyncRecordTable:      {EnableTenant: true},
	LoadBalancerTable:               {EnableTenant: true},
	SecurityGroupCommonRelTable:     {},
	LoadBalancerListenerTable:       {},
	TCloudLbUrlRuleTable:            {},
	LoadBalancerTargetTable:         {},
	LoadBalancerTargetGroupTable:    {},
	TargetGroupListenerRuleRelTable: {},
	ResourceFlowRelTable:            {},
	ResourceFlowLockTable:           {},

	MainAccountTable: {EnableTenant: true},
	RootAccountTable: {EnableTenant: true},

	TaskManagementTable: {},
	TaskDetailTable:     {},
	TenantTable:         {},

	GlobalConfigTable: {},

	ResUsageBizRelTable: {},
}

// Register 注册表名
func (n Name) Register() {
	validTableNames[n] = true
}
