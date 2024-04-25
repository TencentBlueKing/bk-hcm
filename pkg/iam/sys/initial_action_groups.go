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

// GenerateStaticActionGroups generate all the static resource action groups.
func GenerateStaticActionGroups() []client.ActionGroup {
	ActionGroups := make([]client.ActionGroup, 0)

	// generate business Management action groups, contains business related actions
	ActionGroups = append(ActionGroups, genResManagementActionGroups()...)

	return ActionGroups
}

// TODO 开启clb和编排相关功能后放开注释
func genResManagementActionGroups() []client.ActionGroup {
	actionGroups := []client.ActionGroup{
		{
			Name:   "资源管理",
			NameEn: "Res Management",
			Actions: []client.ActionWithID{
				{ID: BizAccess},
			},
			SubGroups: []client.ActionGroup{
				{
					Name:   "IaaS资源",
					NameEn: "Biz IaaS Resource Management",
					Actions: []client.ActionWithID{
						{ID: BizIaaSResCreate},
						{ID: BizIaaSResOperate},
						{ID: BizIaaSResDelete},
					},
				},
				{
					Name:   "负载均衡",
					NameEn: "Biz CLB Resource Management",
					Actions: []client.ActionWithID{
						{ID: BizCLBResCreate},
						{ID: BizCLBResOperate},
						{ID: BizCLBResDelete},
					},
				}, {
					Name:   "证书管理",
					NameEn: "Biz Cert Resource Management",
					Actions: []client.ActionWithID{
						{ID: BizCertResCreate},
						{ID: BizCertResDelete},
					},
				},
				/*{
					Name:   "资源编排",
					NameEn: "Biz Arrange Resource Management",
					Actions: []client.ActionWithID{
						{ID: BizArrangeResCreate},
						{ID: BizArrangeResOperate},
						{ID: BizArrangeResDelete},
					},
				},*/
				{
					Name:   "回收站",
					NameEn: "Biz Recycle Bin",
					Actions: []client.ActionWithID{
						{ID: BizRecycleBinOperate},
						{ID: BizRecycleBinConfig},
					},
				},
				{
					Name:   "操作记录",
					NameEn: "Biz Operation Record",
					Actions: []client.ActionWithID{
						{ID: BizOperationRecordFind},
					},
				},
			},
		},
	}

	actionGroups = append(actionGroups, genResourceAccessActionGroups())
	actionGroups = append(actionGroups, genCloudSelectionActionGroups())
	actionGroups = append(actionGroups, genPlatformManageActionGroups())

	return actionGroups
}

func genCloudSelectionActionGroups() client.ActionGroup {
	return client.ActionGroup{
		Name:   "资源选型",
		NameEn: "Resource Selection",
		SubGroups: []client.ActionGroup{
			{
				Name:   "资源选型",
				NameEn: "Resource Selection",
				Actions: []client.ActionWithID{
					{ID: CloudSelectionRecommend},
				},
			},
			{
				Name:   "部署方案",
				NameEn: "Deployment Scheme",
				Actions: []client.ActionWithID{
					{ID: CloudSelectionSchemeFind},
					{ID: CloudSelectionSchemeEdit},
					{ID: CloudSelectionSchemeDelete},
				},
			},
		},
	}
}

func genResourceAccessActionGroups() client.ActionGroup {
	return client.ActionGroup{
		Name:   "资源接入",
		NameEn: "Resource Access",
		SubGroups: []client.ActionGroup{
			{
				Name:   "云账号",
				NameEn: "Cloud account",
				Actions: []client.ActionWithID{
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
				Actions: []client.ActionWithID{
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
				Actions: []client.ActionWithID{
					{ID: CLBResCreate},
					{ID: CLBResOperate},
					{ID: CLBResDelete},
				},
			}, {
				Name:   "证书管理",
				NameEn: "Cert Resource Management",
				Actions: []client.ActionWithID{
					{ID: CertResCreate},
					{ID: CertResDelete},
				},
			},
			{
				Name:   "回收站",
				NameEn: "Recycle Bin",
				Actions: []client.ActionWithID{
					{ID: RecycleBinAccess},
					{ID: RecycleBinOperate},
					{ID: RecycleBinConfig},
				},
			},
			{
				Name:   "操作记录",
				NameEn: "Operation Record",
				Actions: []client.ActionWithID{
					{ID: OperationRecordFind},
				},
			},
		},
	}
}

func genPlatformManageActionGroups() client.ActionGroup {
	return client.ActionGroup{
		Name:   "平台管理",
		NameEn: "Platform Management",
		SubGroups: []client.ActionGroup{
			{
				Name:   "平台权限",
				NameEn: "Platform Permissions",
				Actions: []client.ActionWithID{
					{ID: CostManage},
					{ID: AccountKeyAccess},
				},
			},
			{
				Name:   "配置管理",
				NameEn: "Configuration Management",
				Actions: []client.ActionWithID{
					{ID: GlobalConfiguration},
				},
			},
		},
	}
}
