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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// BaseSecurityGroup define base security group.
type BaseSecurityGroup struct {
	ID               string        `json:"id"`
	Vendor           enumor.Vendor `json:"vendor"`
	CloudID          string        `json:"cloud_id"`
	Region           string        `json:"region"`
	Name             string        `json:"name"`
	Memo             *string       `json:"memo"`
	CloudCreatedTime string        `json:"cloud_created_time"`
	CloudUpdateTime  string        `json:"cloud_update_time"`
	Tags             core.TagMap   `json:"tags"`
	AccountID        string        `json:"account_id"`
	BkBizID          int64         `json:"bk_biz_id"`
	Creator          string        `json:"creator"`
	Reviser          string        `json:"reviser"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
}

// SecurityGroup define security group
type SecurityGroup[Extension SecurityGroupExtension] struct {
	BaseSecurityGroup `json:",inline"`
	Extension         *Extension `json:"extension"`
}

// GetID ...
func (sg SecurityGroup[T]) GetID() string {
	return sg.BaseSecurityGroup.ID
}

// GetCloudID ...
func (sg SecurityGroup[T]) GetCloudID() string {
	return sg.BaseSecurityGroup.CloudID
}

// SecurityGroupExtension define security group extension.
type SecurityGroupExtension interface {
	TCloudSecurityGroupExtension | AwsSecurityGroupExtension | HuaWeiSecurityGroupExtension |
		AzureSecurityGroupExtension
}

// TCloudSecurityGroupExtension define tcloud security group extension.
type TCloudSecurityGroupExtension struct {
	// CloudProjectID 项目id，默认0。
	CloudProjectID *string `json:"cloud_project_id"`
}

// AwsSecurityGroupExtension define aws security group extension.
type AwsSecurityGroupExtension struct {
	VpcID string `json:"vpc_id"`
	// CloudVpcID vpc云主键ID。
	CloudVpcID *string `json:"cloud_vpc_id"`
	// CloudOwnerID 拥有该安全组的Amazon账号ID。
	CloudOwnerID *string `json:"cloud_owner_id"`
}

// HuaWeiSecurityGroupExtension define huawei security group extension.
type HuaWeiSecurityGroupExtension struct {
	// CloudProjectID 安全组所属的项目ID。
	CloudProjectID string `json:"cloud_project_id"`
	// CloudEnterpriseProjectID 安全组所属的企业项目ID。取值范围：最大长度36字节，
	// 带“-”连字符的UUID格式，或者是字符串“0”。“0”表示默认企业项目。
	CloudEnterpriseProjectID string `json:"cloud_enterprise_project_id"`
}

// AzureSecurityGroupExtension define azure security group extension.
type AzureSecurityGroupExtension struct {
	// ResourceGroupName 资源组名称。
	ResourceGroupName string `json:"resource_group_name"`
	// Etag 唯一只读字符串，每当资源更改都会更新。
	Etag *string `json:"etag"`
	// FlushConnection 启用后，在更新规则时，将重新评估从网络安全组连接创建的流。初始启用将触发重新评估。
	FlushConnection *bool `json:"flush_connection"`
	// ResourceGUID 网络安全组资源的资源GUID。
	ResourceGUID *string `json:"resource_guid"`
}
