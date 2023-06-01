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

package routetable

import (
	"hcm/pkg/api/core"
)

// AzureRoute defines azure route info.
type AzureRoute struct {
	ID                string  `json:"id"`
	CloudID           string  `json:"cloud_id"`
	RouteTableID      string  `json:"route_table_id"`
	CloudRouteTableID string  `json:"cloud_route_table_id"`
	Name              string  `json:"name"`
	AddressPrefix     string  `json:"address_prefix"`
	NextHopType       string  `json:"next_hop_type"`
	NextHopIPAddress  *string `json:"next_hop_ip_address,omitempty"`
	ProvisioningState string  `json:"provisioning_state"`
	*core.Revision    `json:",inline"`
}

// GetID ...
func (route AzureRoute) GetID() string {
	return route.ID
}

// GetCloudID ...
func (route AzureRoute) GetCloudID() string {
	return route.CloudID
}
