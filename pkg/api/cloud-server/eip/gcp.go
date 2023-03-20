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

import (
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/validator"
)

// GcpEipAssociateReq ...
type GcpEipAssociateReq struct {
	EipID              string `json:"eip_id" validate:"required"`
	CvmID              string `json:"cvm_id" validate:"required"`
	NetworkInterfaceID string `json:"network_interface_id" validate:"required"`
}

// Validate ...
func (req *GcpEipAssociateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpEipDisassociateReq ...
type GcpEipDisassociateReq struct {
	EipID              string `json:"eip_id" validate:"required"`
	NetworkInterfaceID string `json:"network_interface_id" validate:"required"`
}

// Validate ...
func (req *GcpEipDisassociateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpEipCreateReq ...
type GcpEipCreateReq struct {
	AccountID string `json:"account_id" validate:"required"`
	*eip.GcpEipCreateOption
}
