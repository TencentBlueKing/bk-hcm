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

import (
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// GenerateStaticActionGroups generate all the static resource action groups.
func GenerateStaticActionGroups() []iam.ActionGroup {
	ActionGroups := make([]iam.ActionGroup, 0)

	// generate business Management action groups, contains business related actions
	ActionGroups = append(ActionGroups, genResManagementActionGroups()...)

	return ActionGroups
}

// TODO 开启clb和编排相关功能后放开注释
func genResManagementActionGroups() []iam.ActionGroup {
	actionGroups := []iam.ActionGroup{
		{
			Name:   "资源管理",
			NameEn: "Res Management",
			Actions: []iam.ActionWithID{
				{ID: BizAccess},
			},
			SubGroups: []iam.ActionGroup{
				{
					Name:   "IaaS资源",
					NameEn: "Biz IaaS Resource Management",
					Actions: []iam.ActionWithID{
						{ID: BizIaaSResCreate},
						{ID: BizIaaSResOperate},
						{ID: BizIaaSResDelete},
					},
				},
				{
					Name:   "负载均衡",
					NameEn: "Biz CLB Resource Management",
					Actions: []iam.ActionWithID{
						{ID: BizCLBResCreate},
						{ID: BizCLBResOperate},
						{ID: BizCLBResDelete},
					},
				}, {
					Name:   "证书管理",
					NameEn: "Biz Cert Resource Management",
					Actions: []iam.ActionWithID{
						{ID: BizCertResCreate},
						{ID: BizCertResDelete},
					},
				},
				/*{
					Name:   "资源编排",
					NameEn: "Biz Arrange Resource Management",
					Actions: []iam.ActionWithID{
						{ID: BizArrangeResCreate},
						{ID: BizArrangeResOperate},
						{ID: BizArrangeResDelete},
					},
				},*/
				{
					Name:   "回收站",
					NameEn: "Biz Recycle Bin",
					Actions: []iam.ActionWithID{
						{ID: BizRecycleBinOperate},
						{ID: BizRecycleBinConfig},
					},
				},
				{
					Name:   "操作记录",
					NameEn: "Biz Operation Record",
					Actions: []iam.ActionWithID{
						{ID: BizOperationRecordFind},
					},
				},
				{
					Name:   "任务管理",
					NameEn: "Biz Task Management",
					Actions: []iam.ActionWithID{
						{ID: BizTaskManagementOperate},
					},
				},
			},
		},
	}

	actionGroups = append(actionGroups, genResourceAccessActionGroups())
	actionGroups = append(actionGroups, genCloudSelectionActionGroups())
	actionGroups = append(actionGroups, genPlatformManageActionGroups())
	actionGroups = append(actionGroups, genCloudAccountActionGroups())

	return actionGroups
}

func genCloudSelectionActionGroups() iam.ActionGroup {
	return iam.ActionGroup{
		Name:   "资源选型",
		NameEn: "Resource Selection",
		SubGroups: []iam.ActionGroup{
			{
				Name:   "资源选型",
				NameEn: "Resource Selection",
				Actions: []iam.ActionWithID{
					{ID: CloudSelectionRecommend},
				},
			},
			{
				Name:   "部署方案",
				NameEn: "Deployment Scheme",
				Actions: []iam.ActionWithID{
					{ID: CloudSelectionSchemeFind},
					{ID: CloudSelectionSchemeEdit},
					{ID: CloudSelectionSchemeDelete},
				},
			},
		},
	}
}

func genResourceAccessActionGroups() iam.ActionGroup {
	return iam.ActionGroup{
		Name:   "资源接入",
		NameEn: "Resource Access",
		SubGroups: []iam.ActionGroup{
			{
				Name:   "云账号",
				NameEn: "Cloud account",
				Actions: []iam.ActionWithID{
					{ID: AccountFind},
					{ID: AccountImport},
					{ID: AccountEdit},
					{ID: SubAccountEdit},
					{ID: AccountDelete},
				},
			},
			{
				Name:   "IaaS资源",
				NameEn: "IaaS Resource Management",
				Actions: []iam.ActionWithID{
					{ID: ResourceFind},
					{ID: ResourceAssign},
					{ID: IaaSResCreate},
					{ID: IaaSResOperate},
					{ID: IaaSResDelete},
				},
			},
			{
				Name:   "负载均衡",
				NameEn: "CLB Resource Management",
				Actions: []iam.ActionWithID{
					{ID: CLBResCreate},
					{ID: CLBResOperate},
					{ID: CLBResDelete},
				},
			}, {
				Name:   "证书管理",
				NameEn: "Cert Resource Management",
				Actions: []iam.ActionWithID{
					{ID: CertResCreate},
					{ID: CertResDelete},
				},
			},
			{
				Name:   "回收站",
				NameEn: "Recycle Bin",
				Actions: []iam.ActionWithID{
					{ID: RecycleBinAccess},
					{ID: RecycleBinOperate},
					{ID: RecycleBinConfig},
				},
			},
			{
				Name:   "操作记录",
				NameEn: "Operation Record",
				Actions: []iam.ActionWithID{
					{ID: OperationRecordFind},
				},
			},
			{
				Name:   "COS资源",
				NameEn: "COS Resource",
				Actions: []iam.ActionWithID{
					{ID: CosBucketCreate},
					{ID: CosBucketFind},
					{ID: CosBucketDelete},
				},
			},
		},
	}
}

func genPlatformManageActionGroups() iam.ActionGroup {
	return iam.ActionGroup{
		Name:   "平台管理",
		NameEn: "Platform Management",
		SubGroups: []iam.ActionGroup{
			{
				Name:   "平台权限",
				NameEn: "Platform Permissions",
				Actions: []iam.ActionWithID{
					{ID: CostManage},
					{ID: AccountKeyAccess},
				},
			},
			{
				Name:   "配置管理",
				NameEn: "Configuration Management",
				Actions: []iam.ActionWithID{
					{ID: GlobalConfiguration},
				},
			},
			{
				Name:   "云账号管理",
				NameEn: "Root Account Management",
				Actions: []iam.ActionWithID{
					{ID: RootAccountManage},
				},
			},
			{
				Name:   "云账单管理",
				NameEn: "Account Bill Management",
				Actions: []iam.ActionWithID{
					{ID: AccountBillPull},
					{ID: AccountBillManage},
				},
			},
			{
				Name:   "服务请求",
				NameEn: "Service Request",
				Actions: []iam.ActionWithID{
					{ID: ApplicationManage},
				},
			},
		},
	}
}

func genCloudAccountActionGroups() iam.ActionGroup {
	return iam.ActionGroup{
		Name:   "云账号管理",
		NameEn: "Cloud Account Management",
		SubGroups: []iam.ActionGroup{
			{
				Name:   "二级账号",
				NameEn: "Main Account",
				Actions: []iam.ActionWithID{
					{ID: MainAccountFind},
					{ID: MainAccountCreate},
					{ID: MainAccountEdit},
				},
			},
		},
	}
}
