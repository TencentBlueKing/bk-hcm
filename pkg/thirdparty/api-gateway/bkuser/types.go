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

package bkuser

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
)

// ---------------------------- batch_lookup_virtual_user ----------------------------

// BatchLookupVirtualUserResult bk-user batch lookup virtual user result
type BatchLookupVirtualUserResult struct {
	Data []VirtualUserItem `json:"data"`
}

// VirtualUserItem virtual user item
type VirtualUserItem struct {
	BkUsername  string `json:"bk_username"`
	LoginName   string `json:"login_name"`
	DisplayName string `json:"display_name"`
}

// TenantStatus tenant status
type TenantStatus string

const (
	// TenantStatusEnabled enabled
	TenantStatusEnabled TenantStatus = "enabled"
	// TenantStatusDisabled disabled
	TenantStatusDisabled TenantStatus = "disabled"
)

// Tenant tenant item
type Tenant struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// BKUser tenant status definition is diff from HCM, use GetStatus to get correct HCM tenant Status
	BKUserStatus TenantStatus `json:"status"`
}

// GetStatus get HCM tenant status
func (t Tenant) GetStatus() enumor.TenantStatus {
	status := enumor.TenantDisable
	if t.BKUserStatus == TenantStatusEnabled {
		status = enumor.TenantEnable
	}
	return status
}

// String ...
func (t Tenant) String() string {
	return fmt.Sprintf("{%s:%s:%s}", t.Id, t.Name, t.BKUserStatus)
}

// TenantListResult tenant list result
type TenantListResult struct {
	Data []Tenant `json:"data"`
}
