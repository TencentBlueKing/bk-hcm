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
	// Biz defines hcm biz resource type to register iam.
	Biz client.TypeID = "biz"
)

// ActionID action id to register iam.
const (
	// BizAccess biz resource access action id to register iam.
	BizAccess client.ActionID = "biz_access"
	// BizIaaSResCreate biz iaas resource create action id to register iam.
	BizIaaSResCreate client.ActionID = "biz_iaas_resource_create"
	// BizIaaSResOperate biz iaas resource operate action id to register iam.
	BizIaaSResOperate client.ActionID = "biz_iaas_resource_operate"
	// BizIaaSResDelete biz iaas resource delete action id to register iam.
	BizIaaSResDelete client.ActionID = "biz_iaas_resource_delete"

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
	// IaaSResourceCreate iaas resource create action id to register iam.
	IaaSResourceCreate client.ActionID = "iaas_resource_create"
	// IaaSResourceOperate iaas resource operate action id to register iam.
	IaaSResourceOperate client.ActionID = "iaas_resource_operate"
	// IaaSResourceDelete iaas resource delete action id to register iam.
	IaaSResourceDelete client.ActionID = "iaas_resource_delete"

	// RecycleBinFind recycle bin find action id to register iam.
	RecycleBinFind client.ActionID = "recycle_bin_find"
	// RecycleBinManage recycle bin manage action id to register iam.
	RecycleBinManage client.ActionID = "recycle_bin_manage"

	// BizAuditFind biz audit find action id to register iam.
	BizAuditFind client.ActionID = "biz_audit_find"
	// ResourceAuditFind account audit find action id to register iam.
	ResourceAuditFind client.ActionID = "resource_audit_find"

	// Skip is an action that no need to auth
	Skip client.ActionID = "skip"
)

// ActionIDNameMap is action id type map.
var ActionIDNameMap = map[client.ActionID]string{
	BizAccess:           "业务访问",
	BizIaaSResCreate:    "业务-IaaS资源创建",
	BizIaaSResOperate:   "业务-IaaS资源操作",
	BizIaaSResDelete:    "业务-IaaS资源删除",
	AccountFind:         "账号查看",
	AccountImport:       "账号录入",
	AccountEdit:         "账号编辑",
	AccountDelete:       "账号删除",
	AccountKeyAccess:    "账号密钥访问",
	ResourceFind:        "资源查看",
	ResourceAssign:      "资源分配",
	IaaSResourceCreate:  "IaaS资源创建",
	IaaSResourceOperate: "IaaS资源操作",
	IaaSResourceDelete:  "IaaS资源删除",
	RecycleBinFind:      "回收站查看",
	RecycleBinManage:    "回收站管理",
	BizAuditFind:        "业务审计查看",
	ResourceAuditFind:   "资源审计查看",
}

const (
	// AccountSelection is account instance selection id to register iam.
	AccountSelection client.InstanceSelectionID = "account"
	// BizSelection is biz instance selection id to register iam.
	BizSelection client.InstanceSelectionID = "business"
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
