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

package hcservice

import "hcm/pkg/criteria/validator"

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq struct {
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate VpcUpdateReq.
func (u *VpcUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Sync --------------------------

// ResourceSyncReq defines sync resource request.
type ResourceSyncReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	Region            string `json:"region" validate:"omitempty"`
	ResourceGroupName string `json:"resource_group_name" validate:"omitempty"`
	VpcName           string `json:"vpc_name" validate:"omitempty"`
}

// ResourceSyncResult defines sync vpc result.
type ResourceSyncResult struct {
	TaskID string `json:"task_id"`
}

// Validate validate sync vpc request.
func (r *ResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}
