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

import dataproto "hcm/pkg/api/data-service/cloud/eip"

// EipResult ...
type EipResult struct {
	*dataproto.EipResult `json:",inline"`
	CvmID                string `json:"cvm_id,omitempty"`
}

// EipListResult ...
type EipListResult struct {
	Count   *uint64      `json:"count,omitempty"`
	Details []*EipResult `json:"details"`
}

// TCloudEipExtResult ...
type TCloudEipExtResult struct {
	*dataproto.EipExtResult[dataproto.TCloudEipExtensionResult] `json:",inline"`
	CvmID                                                       string `json:"cvm_id,omitempty"`
}

// AwsEipExtResult ...
type AwsEipExtResult struct {
	*dataproto.EipExtResult[dataproto.AwsEipExtensionResult] `json:",inline"`
	CvmID                                                    string `json:"cvm_id,omitempty"`
}

// AzureEipExtResult ...
type AzureEipExtResult struct {
	*dataproto.EipExtResult[dataproto.AzureEipExtensionResult] `json:",inline"`
	CvmID                                                      string `json:"cvm_id,omitempty"`
}

// HuaWeiEipExtResult ...
type HuaWeiEipExtResult struct {
	*dataproto.EipExtResult[dataproto.HuaWeiEipExtensionResult] `json:",inline"`
	CvmID                                                       string `json:"cvm_id,omitempty"`
}

// GcpEipExtResult ...
type GcpEipExtResult struct {
	*dataproto.EipExtResult[dataproto.GcpEipExtensionResult] `json:",inline"`
	CvmID                                                    string `json:"cvm_id,omitempty"`
}
