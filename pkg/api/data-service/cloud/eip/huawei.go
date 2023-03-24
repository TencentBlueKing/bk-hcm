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

package eip

// HuaWeiEipExtensionCreateReq ...
type HuaWeiEipExtensionCreateReq struct {
	PortID              *string `json:"port_id"`
	BandwidthId         *string `json:"bandwidth_id"`
	BandwidthName       *string `json:"bandwidth_name"`
	BandwidthSize       *int32  `json:"bandwidth_size"`
	EnterpriseProjectId *string `json:"enterprise_project_id"`
	Type                *string `json:"type"`
	BandwidthShareType  string  `json:"bandwidth_share_type"`
	ChargeMode          string  `json:"charge_mode"`
}

// HuaWeiEipExtensionResult ...
type HuaWeiEipExtensionResult struct {
	PortID              *string `json:"port_id"`
	BandwidthId         *string `json:"bandwidth_id"`
	BandwidthName       *string `json:"bandwidth_name"`
	BandwidthSize       *int32  `json:"bandwidth_size"`
	EnterpriseProjectId *string `json:"enterprise_project_id"`
	Type                *string `json:"type"`
	BandwidthShareType  string  `json:"bandwidth_share_type"`
	ChargeMode          string  `json:"charge_mode"`
}

// HuaWeiEipExtensionUpdateReq ...
type HuaWeiEipExtensionUpdateReq struct {
	PortID              *string `json:"port_id"`
	BandwidthId         *string `json:"bandwidth_id"`
	BandwidthName       *string `json:"bandwidth_name"`
	BandwidthSize       *int32  `json:"bandwidth_size"`
	EnterpriseProjectId *string `json:"enterprise_project_id"`
	Type                *string `json:"type"`
	BandwidthShareType  string  `json:"bandwidth_share_type"`
	ChargeMode          string  `json:"charge_mode"`
}
