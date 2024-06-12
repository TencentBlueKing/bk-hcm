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

// GenerateStaticInstanceSelections return need registered static instance selection.
func GenerateStaticInstanceSelections() []client.InstanceSelection {
	return []client.InstanceSelection{
		{
			ID:     AccountSelection,
			Name:   "账号列表",
			NameEn: "Account List",
			ResourceTypeChain: []client.ResourceChain{
				{
					SystemID: SystemIDHCM,
					ID:       Account,
				},
			},
		},
		{
			ID:     CloudSelectionSchemeSelection,
			Name:   "方案列表",
			NameEn: "Scheme List",
			ResourceTypeChain: []client.ResourceChain{
				{
					SystemID: SystemIDHCM,
					ID:       CloudSelectionScheme,
				},
			},
		},
		{
			ID:     MainAccountSelection,
			Name:   "二级账号列表",
			NameEn: "Main Account List",
			ResourceTypeChain: []client.ResourceChain{
				{
					SystemID: SystemIDHCM,
					ID:       MainAccount,
				},
			},
		},
		{
			ID:     RootAccountSelection,
			Name:   "一级账号列表",
			NameEn: "Root Account List",
			ResourceTypeChain: []client.ResourceChain{
				{
					SystemID: SystemIDHCM,
					ID:       RootAccount,
				},
			},
		},
	}
}
