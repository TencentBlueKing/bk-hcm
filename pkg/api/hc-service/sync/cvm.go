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

package sync

import (
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// SyncHuaWeiCvmReq define sync huawei cvm req.
type SyncHuaWeiCvmReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncHuaWeiCvmReq
func (req SyncHuaWeiCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncGcpCvmReq ...
type SyncGcpCvmReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
}

// Validate SyncGcpCvmReq
func (req SyncGcpCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAzureCvmReq ...
type SyncAzureCvmReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate SyncAzureCvmReq
func (req SyncAzureCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SyncAwsCvmReq define sync aws cvm option.
type SyncAwsCvmReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate SyncAwsCvmReq
func (req SyncAwsCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// OtherSyncHostReq other vendor sync host request.
type OtherSyncHostReq struct {
	AccountID  string `json:"account_id" validate:"required"`
	BizID      int64  `json:"bk_biz_id" validate:"required"`
	Concurrent uint   `json:"concurrent,omitempty"`
}

// Validate ...
func (req *OtherSyncHostReq) Validate() error {
	return validator.Validate.Struct(req)
}

// OtherSyncHostByCondReq other vendor sync host by cond request.
type OtherSyncHostByCondReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	BizID     int64   `json:"bk_biz_id" validate:"required"`
	HostIDs   []int64 `json:"bk_host_ids" validate:"required"`
}

// Validate ...
func (req *OtherSyncHostByCondReq) Validate() error {
	if len(req.HostIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("host ids should <= %d", constant.CloudResourceSyncMaxLimit)
	}

	if req.BizID == 0 {
		return fmt.Errorf("bk_biz_id is invalid")
	}

	if req.AccountID == "" {
		return fmt.Errorf("account_id is invalid")
	}

	return nil
}

// OtherDelHostByCondReq other vendor delete host by condition request.
type OtherDelHostByCondReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	HostIDs   []int64 `json:"bk_host_ids" validate:"required"`
}

// Validate ...
func (req *OtherDelHostByCondReq) Validate() error {
	if len(req.HostIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("host ids should <= %d", constant.CloudResourceSyncMaxLimit)
	}

	if req.AccountID == "" {
		return fmt.Errorf("account_id is invalid")
	}

	return nil
}
