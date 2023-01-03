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

package enumor

/*
	audit.go store audit related enum values.
*/

// AuditResourceType audit resource type.
type AuditResourceType string

const (
	// Account 策略资源
	Account AuditResourceType = "account"
	// Vpc 策略资源
	Vpc AuditResourceType = "vpc"
)

// AuditResourceTypeEnums resource type map.
var AuditResourceTypeEnums = map[AuditResourceType]bool{
	Account: true,
	Vpc:     true,
}

// Exist judge enum value exist.
func (a AuditResourceType) Exist() bool {
	_, exist := AuditResourceTypeEnums[a]
	return exist
}

// AuditAction audit action type.
type AuditAction string

const (
	// Create 创建
	Create AuditAction = "create"
	// Update 更新
	Update AuditAction = "update"
	// Delete 删除
	Delete AuditAction = "delete"
)

// AuditActionEnums op type map.
var AuditActionEnums = map[AuditAction]bool{
	Create: true,
	Update: true,
	Delete: true,
}

// Exist judge enum value exist.
func (a AuditAction) Exist() bool {
	_, exist := AuditActionEnums[a]
	return exist
}
