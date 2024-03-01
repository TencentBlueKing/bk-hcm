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

package clb

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/rest"
)

// BaseClb define base clb.
type BaseClb struct {
	ID        string        `json:"id"`
	CloudID   string        `json:"cloud_id"`
	Name      string        `json:"name"`
	Vendor    enumor.Vendor `json:"vendor"`
	AccountID string        `json:"account_id"`
	BkBizID   int64         `json:"bk_biz_id"`

	Region               string    `json:"region" validate:"omitempty"`
	Zones                []*string `json:"zones"`
	BackupZones          []*string `json:"backup_zones"`
	VpcID                string    `json:"vpc_id" validate:"omitempty"`
	CloudVpcID           string    `json:"cloud_vpc_id" validate:"omitempty"`
	SubnetID             string    `json:"subnet_id" validate:"omitempty"`
	CloudSubnetID        string    `json:"cloud_subnet_id" validate:"omitempty"`
	PrivateIPv4Addresses []*string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []*string `json:"private_ipv6_addresses"`
	PublicIPv4Addresses  []*string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses  []*string `json:"public_ipv6_addresses"`
	Domain               string    `json:"domain"`
	Status               string    `json:"status"`
	CloudCreatedTime     string    `json:"cloud_created_time"`
	CloudStatusTime      string    `json:"cloud_status_time"`
	CloudExpiredTime     string    `json:"cloud_expired_time"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// Clb define clb.
type Clb[Ext Extension] struct {
	BaseClb   `json:",inline"`
	Extension *Ext `json:"extension"`
}

// GetID ...
func (cert Clb[T]) GetID() string {
	return cert.BaseClb.ID
}

// GetCloudID ...
func (cert Clb[T]) GetCloudID() string {
	return cert.BaseClb.CloudID
}

// Extension extension.
type Extension interface {
	TCloudClbExtension
}

// ClbCreateResp ...
type ClbCreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ClbCreateResult `json:"data"`
}

// ClbCreateResult ...
type ClbCreateResult struct {
	ID string `json:"id"`
}
