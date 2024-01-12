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

package cloud

import (
	"hcm/pkg/criteria/enumor"
)

// SecurityGroupRule define security group rule.
type SecurityGroupRule interface {
	TCloudSecurityGroupRule | AwsSecurityGroupRule | HuaWeiSecurityGroupRule | AzureSecurityGroupRule
}

// TCloudSecurityGroupRule define tcloud security group rule.
type TCloudSecurityGroupRule struct {
	ID                         string                       `json:"id"`
	CloudPolicyIndex           int64                        `json:"cloud_policy_index"`
	Version                    string                       `json:"version"`
	Protocol                   *string                      `json:"protocol"`
	Port                       *string                      `json:"port"`
	ServiceID                  *string                      `json:"service_id"`
	CloudServiceID             *string                      `json:"cloud_service_id"`
	ServiceGroupID             *string                      `json:"service_group_id"`
	CloudServiceGroupID        *string                      `json:"cloud_service_group_id"`
	IPv4Cidr                   *string                      `json:"ipv4_cidr"`
	IPv6Cidr                   *string                      `json:"ipv6_cidr"`
	CloudTargetSecurityGroupID *string                      `json:"cloud_target_security_group_id"`
	AddressID                  *string                      `json:"address_id"`
	CloudAddressID             *string                      `json:"cloud_address_id"`
	AddressGroupID             *string                      `json:"address_group_id"`
	CloudAddressGroupID        *string                      `json:"cloud_address_group_id"`
	Action                     string                       `json:"action"`
	Memo                       *string                      `json:"memo"`
	Type                       enumor.SecurityGroupRuleType `json:"type"`
	CloudSecurityGroupID       string                       `json:"cloud_security_group_id"`
	SecurityGroupID            string                       `json:"security_group_id"`
	Region                     string                       `json:"region"`
	AccountID                  string                       `json:"account_id"`
	Creator                    string                       `json:"creator"`
	Reviser                    string                       `json:"reviser"`
	CreatedAt                  string                       `json:"created_at"`
	UpdatedAt                  string                       `json:"updated_at"`
}

// AwsSecurityGroupRule define aws security group rule.
type AwsSecurityGroupRule struct {
	ID                         string                       `json:"id"`
	CloudID                    string                       `json:"cloud_id"`
	IPv4Cidr                   *string                      `json:"ipv4_cidr"`
	IPv6Cidr                   *string                      `json:"ipv6_cidr"`
	Memo                       *string                      `json:"memo"`
	FromPort                   *int64                       `json:"from_port"`
	ToPort                     *int64                       `json:"to_port"`
	Type                       enumor.SecurityGroupRuleType `json:"type"`
	Protocol                   *string                      `json:"protocol"`
	CloudPrefixListID          *string                      `json:"cloud_prefix_list_id"`
	CloudTargetSecurityGroupID *string                      `json:"cloud_target_security_group_id"`
	CloudSecurityGroupID       string                       `json:"cloud_security_group_id"`
	CloudGroupOwnerID          string                       `json:"cloud_group_owner_id"`
	AccountID                  string                       `json:"account_id"`
	Region                     string                       `json:"region"`
	SecurityGroupID            string                       `json:"security_group_id"`
	Creator                    string                       `json:"creator"`
	Reviser                    string                       `json:"reviser"`
	CreatedAt                  string                       `json:"created_at"`
	UpdatedAt                  string                       `json:"updated_at"`
}

// GetID ...
func (sgr AwsSecurityGroupRule) GetID() string {
	return sgr.ID
}

// GetCloudID ...
func (sgr AwsSecurityGroupRule) GetCloudID() string {
	return sgr.CloudID
}

// HuaWeiSecurityGroupRule define huawei security group rule.
type HuaWeiSecurityGroupRule struct {
	ID                        string                       `json:"id"`
	CloudID                   string                       `json:"cloud_id"`
	Memo                      *string                      `json:"memo"`
	Protocol                  string                       `json:"protocol"`
	Ethertype                 string                       `json:"ethertype"`
	CloudRemoteGroupID        string                       `json:"cloud_remote_group_id"`
	RemoteIPPrefix            string                       `json:"remote_ip_prefix"`
	CloudRemoteAddressGroupID string                       `json:"cloud_remote_address_group_id"`
	Port                      string                       `json:"port"`
	Priority                  int64                        `json:"priority"`
	Action                    string                       `json:"action"`
	Type                      enumor.SecurityGroupRuleType `json:"type"`
	CloudSecurityGroupID      string                       `json:"cloud_security_group_id"`
	CloudProjectID            string                       `json:"cloud_project_id"`
	AccountID                 string                       `json:"account_id"`
	Region                    string                       `json:"region"`
	SecurityGroupID           string                       `json:"security_group_id"`
	Creator                   string                       `json:"creator"`
	Reviser                   string                       `json:"reviser"`
	CreatedAt                 string                       `json:"created_at"`
	UpdatedAt                 string                       `json:"updated_at"`
}

// GetID ...
func (sgr HuaWeiSecurityGroupRule) GetID() string {
	return sgr.ID
}

// GetCloudID ...
func (sgr HuaWeiSecurityGroupRule) GetCloudID() string {
	return sgr.CloudID
}

// AzureSecurityGroupRule define azure security group rule.
type AzureSecurityGroupRule struct {
	ID                                  string                       `json:"id"`
	CloudID                             string                       `json:"cloud_id"`
	Etag                                *string                      `json:"etag"`
	Name                                string                       `json:"name"`
	Memo                                *string                      `json:"memo"`
	DestinationAddressPrefix            *string                      `json:"destination_address_prefix"`
	DestinationAddressPrefixes          []*string                    `json:"destination_address_prefixes"`
	CloudDestinationAppSecurityGroupIDs []*string                    `json:"cloud_destination_app_security_group_ids"`
	DestinationPortRange                *string                      `json:"destination_port_range"`
	DestinationPortRanges               []*string                    `json:"destination_port_ranges"`
	Protocol                            string                       `json:"protocol"`
	ProvisioningState                   string                       `json:"provisioning_state"`
	SourceAddressPrefix                 *string                      `json:"source_address_prefix"`
	SourceAddressPrefixes               []*string                    `json:"source_address_prefixes"`
	CloudSourceAppSecurityGroupIDs      []*string                    `json:"cloud_source_app_security_group_ids"`
	SourcePortRange                     *string                      `json:"source_port_range"`
	SourcePortRanges                    []*string                    `json:"source_port_ranges"`
	Priority                            int32                        `json:"priority"`
	Type                                enumor.SecurityGroupRuleType `json:"type"`
	Access                              string                       `json:"access"`
	CloudSecurityGroupID                string                       `json:"cloud_security_group_id"`
	AccountID                           string                       `json:"account_id"`
	Region                              string                       `json:"region"`
	SecurityGroupID                     string                       `json:"security_group_id"`
	Creator                             string                       `json:"creator"`
	Reviser                             string                       `json:"reviser"`
	CreatedAt                           string                       `json:"created_at"`
	UpdatedAt                           string                       `json:"updated_at"`
}

// GetID ...
func (sgr AzureSecurityGroupRule) GetID() string {
	return sgr.ID
}

// GetCloudID ...
func (sgr AzureSecurityGroupRule) GetCloudID() string {
	return sgr.CloudID
}
