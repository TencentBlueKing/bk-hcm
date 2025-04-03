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

// Package hccvm ...
package hccvm

import (
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// CvmSyncReq cvm sync request
type CvmSyncReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	Region            string   `json:"region" validate:"omitempty"`
	Zone              string   `json:"zone" validate:"omitempty"`
	ResourceGroupName string   `json:"resource_group_name" validate:"omitempty"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate cvm sync request.
func (req *CvmSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// OperateSyncReq cvm oprate sync request.
type OperateSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// ListCvmNetworkInterfaceReq defines list cvm mac address request.
type ListCvmNetworkInterfaceReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CvmIDs    []string `json:"cvm_ids" validate:"required,min=1,max=50"`
}

// Validate validate list cvm mac address request.
func (r *ListCvmNetworkInterfaceReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ListCvmNetworkInterfaceRespItem defines list cvm mac address response item.
type ListCvmNetworkInterfaceRespItem struct {
	MacAddressToPrivateIpAddresses map[string][]string `json:"mac_address_to_private_ip"`
}
