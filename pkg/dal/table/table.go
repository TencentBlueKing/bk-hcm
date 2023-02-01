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
	// AccountTable is account table's name.
	AccountTable Name = "account"
	// AccountBizRelTable is account and biz relation table's name.
	AccountBizRelTable Name = "account_biz_rel"
	// SecurityGroupTable is security group table's name.
	SecurityGroupTable Name = "security_group"
	// VpcSecurityGroupRelTable is vpc and security group table's name.
	VpcSecurityGroupRelTable Name = "vpc_security_group_rel"
	// SecurityGroupTagTable is security group tag table's name.
	SecurityGroupTagTable Name = "security_group_tag"
	// SecurityGroupSubnetTable is security group subnet table's name.
	SecurityGroupSubnetTable Name = "security_group_subnet"
	// SecurityGroupBizRelTable is security group and biz rel table's name.
	SecurityGroupBizRelTable Name = "security_group_biz_rel"
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
)

// Validate whether the table name is valid or not.
func (n Name) Validate() error {
	valid := validTableNames[n]
	if valid {
		return nil
	}

	switch n {
	case AuditTable:
	case AccountTable:
	case AccountBizRelTable:
	case VpcTable:
	case SubnetTable:
	case IDGenerator:
	case SecurityGroupTable:
	case VpcSecurityGroupRelTable:
	case SecurityGroupTagTable:
	case SecurityGroupSubnetTable:
	case SGSecurityGroupRuleTable:
	case TCloudSecurityGroupRuleTable:
	case AwsSecurityGroupRuleTable:
	case HuaWeiSecurityGroupRuleTable:
	case AzureSecurityGroupRuleTable:
	case SGNetworkInterfaceRelTable:
	case GcpFirewallRuleTable:
	default:
		return fmt.Errorf("unknown table name: %s", n)
	}

	return nil
}

// Register 注册表名
func (n Name) Register() {
	validTableNames[n] = true
}
