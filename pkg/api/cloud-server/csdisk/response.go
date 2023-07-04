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

package csdisk

import dataproto "hcm/pkg/api/data-service/cloud/disk"

// DiskResult ...
type DiskResult struct {
	*dataproto.DiskResult `json:",inline"`
	InstanceType          string `json:"instance_type,omitempty"`
	InstanceID            string `json:"instance_id,omitempty"`
}

// DiskListResult ...
type DiskListResult struct {
	Count   *uint64       `json:"count,omitempty"`
	Details []*DiskResult `json:"details"`
}

// TCloudDiskExtResult ...
type TCloudDiskExtResult struct {
	*dataproto.DiskExtResult[dataproto.TCloudDiskExtensionResult] `json:",inline"`
	InstanceType                                                  string `json:"instance_type,omitempty"`
	InstanceID                                                    string `json:"instance_id,omitempty"`
	InstanceName                                                  string `json:"instance_name,omitempty"`
}

// AwsDiskExtResult ...
type AwsDiskExtResult struct {
	*dataproto.DiskExtResult[dataproto.AwsDiskExtensionResult] `json:",inline"`
	InstanceType                                               string `json:"instance_type,omitempty"`
	InstanceID                                                 string `json:"instance_id,omitempty"`
	InstanceName                                               string `json:"instance_name,omitempty"`
}

// AzureDiskExtResult ...
type AzureDiskExtResult struct {
	*dataproto.DiskExtResult[dataproto.AzureDiskExtensionResult] `json:",inline"`
	InstanceType                                                 string `json:"instance_type,omitempty"`
	InstanceID                                                   string `json:"instance_id,omitempty"`
	InstanceName                                                 string `json:"instance_name,omitempty"`
}

// HuaWeiDiskExtResult ...
type HuaWeiDiskExtResult struct {
	*dataproto.DiskExtResult[dataproto.HuaWeiDiskExtensionResult] `json:",inline"`
	InstanceType                                                  string `json:"instance_type,omitempty"`
	InstanceID                                                    string `json:"instance_id,omitempty"`
	InstanceName                                                  string `json:"instance_name,omitempty"`
}

// GcpDiskExtResult ...
type GcpDiskExtResult struct {
	*dataproto.DiskExtResult[dataproto.GcpDiskExtensionResult] `json:",inline"`
	InstanceType                                               string `json:"instance_type,omitempty"`
	InstanceID                                                 string `json:"instance_id,omitempty"`
	InstanceName                                               string `json:"instance_name,omitempty"`
}
