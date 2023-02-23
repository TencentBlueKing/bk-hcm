/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by accountlicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) accountlicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package sys

import "hcm/pkg/iam/client"

const (
	// SystemIDHCM is hcm system id in iam.
	SystemIDHCM = "bk-hcm"
	// SystemNameHCMEn is hcm system english name in iam.
	SystemNameHCMEn = "hcmv2"
	// SystemNameHCM is hcm system name in iam.
	SystemNameHCM = "海垒v2"

	// SystemIDCMDB is cmdb system id in iam.
	SystemIDCMDB = "bk_cmdb"
	// SystemNameCMDB is cmdb system name in iam.
	SystemNameCMDB = "配置平台"
)

// SystemIDNameMap is system id to name map.
var SystemIDNameMap = map[string]string{
	SystemIDHCM:  SystemNameHCM,
	SystemIDCMDB: SystemNameCMDB,
}

const (
	// Account defines cloud account resource type to register iam.
	Account client.TypeID = "account"
)

// ActionID action id to register iam.
const (
	// AccountFind account find action id to register iam.
	AccountFind client.ActionID = "account_find"
	// AccountKeyAccess account secret key access action id to register iam.
	AccountKeyAccess client.ActionID = "account_key_access"
	// AccountImport account import action id to register iam.
	AccountImport client.ActionID = "account_import"
	// AccountEdit account edit action id to register iam.
	AccountEdit client.ActionID = "account_edit"
	// AccountDelete account delete action id to register iam.
	AccountDelete client.ActionID = "account_delete"

	// ResourceFind resource find action id to register iam.
	ResourceFind client.ActionID = "resource_find"
	// ResourceAssign resource assign action id to register iam.
	ResourceAssign client.ActionID = "resource_assign"
	// ResourceManage resource manage action id to register iam.
	ResourceManage client.ActionID = "resource_manage"

	// RecycleBinFind recycle bin find action id to register iam.
	RecycleBinFind client.ActionID = "recycle_bin_find"
	// RecycleBinManage recycle bin manage action id to register iam.
	RecycleBinManage client.ActionID = "recycle_bin_manage"

	// AuditFind audit find action id to register iam.
	AuditFind client.ActionID = "audit_find"

	// Skip is an action that no need to auth
	Skip client.ActionID = "skip"
)

// ActionIDNameMap is action id type map.
var ActionIDNameMap = map[client.ActionID]string{
	AccountFind:      "账号查看",
	AccountKeyAccess: "账号密钥访问",
	AccountImport:    "账号录入",
	AccountEdit:      "账号编辑",
	AccountDelete:    "账号删除",
	ResourceFind:     "资源查看",
	ResourceAssign:   "资源分配",
	ResourceManage:   "资源管理",
	RecycleBinFind:   "回收站查看",
	RecycleBinManage: "回收站管理",
	AuditFind:        "审计查看",
}

const (
	// AccountSelection is account instance selection id to register iam.
	AccountSelection client.InstanceSelectionID = "account"
)

// ActionType action type to register iam.
const (
	Create client.ActionType = "create"
	Delete client.ActionType = "delete"
	View   client.ActionType = "view"
	Edit   client.ActionType = "edit"
	List   client.ActionType = "list"
)

const (
	// UserSubjectType is user's iam authorized subject type.
	UserSubjectType client.SubjectType = "user"
)
