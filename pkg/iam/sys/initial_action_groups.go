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
	ActionGroups = append(ActionGroups, genBusinessManagementActionGroups()...)

	return ActionGroups
}

func genBusinessManagementActionGroups() []client.ActionGroup {
	return []client.ActionGroup{
		{
			Name:   "业务",
			NameEn: "Biz Management",
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
			},
		},
		{
			Name:   "账号",
			NameEn: "Account Management",
			Actions: []client.ActionWithID{
				{ID: AccountFind},
				{ID: AccountImport},
				{ID: AccountEdit},
				{ID: AccountDelete},
				{ID: AccountKeyAccess},
			},
		},
		{
			Name:   "资源",
			NameEn: "Resource Management",
			SubGroups: []client.ActionGroup{
				{
					Name:   "公共",
					NameEn: "Public Resource Management",
					Actions: []client.ActionWithID{
						{ID: ResourceFind},
						{ID: ResourceAssign},
					},
				},
				{
					Name:   "IaaS资源",
					NameEn: "IaaS Resource Management",
					Actions: []client.ActionWithID{
						{ID: IaaSResourceCreate},
						{ID: IaaSResourceOperate},
						{ID: IaaSResourceDelete},
					},
				},
				{
					Name:   "回收站",
					NameEn: "Recycle Bin",
					Actions: []client.ActionWithID{
						{ID: RecycleBinFind},
						{ID: RecycleBinManage},
					},
				},
			},
		},
		{
			Name:   "工作台",
			NameEn: "Workspace",
			Actions: []client.ActionWithID{
				{ID: BizAuditFind},
				{ID: ResourceAuditFind},
			},
		},
	}
}
