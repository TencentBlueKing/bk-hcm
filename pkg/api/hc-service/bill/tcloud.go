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

package bill

import (
	"hcm/pkg/rest"
)

// -------------------------- List --------------------------

// TCloudBillListResult define tcloud bill list result.
type TCloudBillListResult struct {
	Count   *uint64     `json:"count"`
	Details interface{} `json:"details"`
	// 本次请求的上下文信息，可用于下一次请求的请求参数中，加快查询速度
	// 注意：此字段可能返回 null，表示取不到有效值。
	Context *string `json:"Context,omitempty"`
	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty"`
}

// TCloudBillListResp define tcloud bill list resp.
type TCloudBillListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *TCloudBillListResult `json:"data"`
}
