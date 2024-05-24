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
	// CloudSelectionScheme define cloud selection scheme resource type to register iam.
	CloudSelectionScheme client.TypeID = "cloud_selection_scheme"
)

const (
	// AccountSelection is account instance selection id to register iam.
	AccountSelection client.InstanceSelectionID = "account"
	// BizSelection is biz instance selection id to register iam.
	BizSelection client.InstanceSelectionID = "business"
	// CloudSelectionSchemeSelection 云选型方案实例视图
	CloudSelectionSchemeSelection client.InstanceSelectionID = "cloud_selection_scheme"
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

// TODO 名称中New去掉进行替换
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

	// BizCLBResCreate biz clb resource create action id to register iam.
	BizCLBResCreate client.ActionID = "biz_clb_resource_create"
	// BizCLBResOperate biz clb resource operate action id to register iam.
	BizCLBResOperate client.ActionID = "biz_clb_resource_operate"
	// BizCLBResDelete biz clb resource delete action id to register iam.
	BizCLBResDelete client.ActionID = "biz_clb_resource_delete"

	// BizCertResCreate biz cert resource create action id to register iam.
	BizCertResCreate client.ActionID = "biz_cert_resource_create"
	// BizCertResDelete biz cert resource delete action id to register iam.
	BizCertResDelete client.ActionID = "biz_cert_resource_delete"

	// BizArrangeResCreate biz arrange resource create action id to register iam.
	// BizArrangeResCreate client.ActionID = "biz_arrange_resource_create"
	// BizArrangeResOperate biz arrange resource operate action id to register iam.
	// BizArrangeResOperate client.ActionID = "biz_arrange_resource_operate"
	// BizArrangeResDelete biz arrange resource delete action id to register iam.
	// BizArrangeResDelete client.ActionID = "biz_arrange_resource_delete"

	// BizRecycleBinOperate biz recycle bin operate action id to register iam.
	BizRecycleBinOperate client.ActionID = "biz_recycle_bin_operate"
	// BizRecycleBinConfig biz recycle bin config action id to register iam.
	BizRecycleBinConfig client.ActionID = "biz_recycle_bin_config"

	// BizOperationRecordFind biz operation record find action id to register iam.
	BizOperationRecordFind client.ActionID = "biz_operation_record_find"

	// AccountFind account find action id to register iam.
	AccountFind client.ActionID = "account_find"
	// AccountImport account import action id to register iam.
	AccountImport client.ActionID = "account_import"
	// AccountEdit account edit action id to register iam.
	AccountEdit client.ActionID = "account_edit"
	// SubAccountEdit sub account edit action id to register iam.
	SubAccountEdit client.ActionID = "sub_account_edit"
	// AccountDelete account delete action id to register iam.
	AccountDelete client.ActionID = "account_delete"

	// ResourceFind resource find action id to register iam.
	ResourceFind client.ActionID = "resource_find"
	// ResourceAssign resource assign action id to register iam.
	ResourceAssign client.ActionID = "resource_assign"
	// IaaSResCreate iaas resource create action id to register iam.
	IaaSResCreate client.ActionID = "iaas_resource_create"
	// IaaSResOperate iaas resource operate action id to register iam.
	IaaSResOperate client.ActionID = "iaas_resource_operate"
	// IaaSResDelete iaas resource delete action id to register iam.
	IaaSResDelete client.ActionID = "iaas_resource_delete"

	// CLBResCreate clb resource create action id to register iam.
	CLBResCreate client.ActionID = "clb_resource_create"
	// CLBResOperate clb resource operate action id to register iam.
	CLBResOperate client.ActionID = "clb_resource_operate"
	// CLBResDelete clb resource delete action id to register iam.
	CLBResDelete client.ActionID = "clb_resource_delete"

	// CertResCreate cert resource create action id to register iam.
	CertResCreate client.ActionID = "cert_resource_create"
	// CertResDelete cert resource delete action id to register iam.
	CertResDelete client.ActionID = "cert_resource_delete"

	// RecycleBinAccess recycle bin find action id to register iam.
	RecycleBinAccess client.ActionID = "recycle_bin_access"
	// RecycleBinOperate recycle bin operate action id to register iam.
	RecycleBinOperate client.ActionID = "recycle_bin_operate"
	// RecycleBinConfig recycle bin config action id to register iam.
	RecycleBinConfig client.ActionID = "recycle_bin_config"

	// OperationRecordFind operation record find action id to register iam.
	OperationRecordFind client.ActionID = "operation_record_find"

	// CostManage bill manage action id to register iam.
	CostManage client.ActionID = "cost_manage"
	// AccountKeyAccess account secret key access action id to register iam.
	AccountKeyAccess client.ActionID = "account_key_access"

	// GlobalConfiguration global configuration action id to register iam.
	GlobalConfiguration client.ActionID = "global_configuration"

	// CloudSelectionRecommend 选型推荐
	CloudSelectionRecommend client.ActionID = "cloud_selection_recommend"
	// CloudSelectionSchemeFind 方案查看
	CloudSelectionSchemeFind client.ActionID = "cloud_selection_find"
	// CloudSelectionSchemeEdit 方案编辑
	CloudSelectionSchemeEdit client.ActionID = "cloud_selection_edit"
	// CloudSelectionSchemeDelete 方案删除
	CloudSelectionSchemeDelete client.ActionID = "cloud_selection_delete"

	// Skip is an action that no need to auth
	Skip client.ActionID = "skip"
)

// ActionIDNameMap is action id type map.
var ActionIDNameMap = map[client.ActionID]string{
	BizAccess:         "业务访问",
	BizIaaSResCreate:  "业务-IaaS资源创建",
	BizIaaSResOperate: "业务-IaaS资源操作",
	BizIaaSResDelete:  "业务-IaaS资源删除",

	BizCLBResCreate:  "业务-负载均衡创建",
	BizCLBResOperate: "业务-负载均衡操作",
	BizCLBResDelete:  "业务-负载均衡删除",
	BizCertResCreate: "业务-证书创建",
	BizCertResDelete: "业务-证书删除",

	// BizArrangeResCreate:    "业务-资源编排创建",
	// BizArrangeResOperate:   "业务-资源编排操作",
	// BizArrangeResDelete:    "业务-资源编排删除",
	BizRecycleBinOperate:   "业务-回收站操作",
	BizRecycleBinConfig:    "业务-回收站配置",
	BizOperationRecordFind: "业务-操作记录查看",

	AccountFind:    "资源-账号查看",
	AccountImport:  "资源-账号录入",
	AccountEdit:    "资源-账号编辑",
	AccountDelete:  "资源-账号删除",
	SubAccountEdit: "资源-子账号编辑",
	ResourceFind:   "资源-资源查看",
	ResourceAssign: "资源-资源分配",
	IaaSResCreate:  "资源-IaaS资源创建",
	IaaSResOperate: "资源-IaaS资源操作",
	IaaSResDelete:  "资源-IaaS资源删除",

	CLBResCreate:  "负载均衡创建",
	CLBResOperate: "负载均衡操作",
	CLBResDelete:  "负载均衡删除",

	CertResCreate:       "资源-证书创建",
	CertResDelete:       "资源-证书删除",
	RecycleBinAccess:    "资源-回收站查看",
	RecycleBinOperate:   "资源-回收站操作",
	RecycleBinConfig:    "资源-回收站配置",
	OperationRecordFind: "资源-操作记录查看",

	CloudSelectionRecommend:    "选型推荐",
	CloudSelectionSchemeFind:   "方案查看",
	CloudSelectionSchemeEdit:   "方案编辑",
	CloudSelectionSchemeDelete: "方案删除",

	CostManage:          "平台-云成本管理",
	AccountKeyAccess:    "平台-账号密钥访问",
	GlobalConfiguration: "平台-全局配置",
}
