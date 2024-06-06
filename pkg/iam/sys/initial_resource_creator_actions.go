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

// GenerateResourceCreatorActions generate all the resource creator actions that need to be registered to IAM.
func GenerateResourceCreatorActions() client.ResourceCreatorActions {
	return client.ResourceCreatorActions{
		Config: []client.ResourceCreatorAction{
			{
				ResourceID: Account,
				Actions: []client.CreatorRelatedAction{
					{
						ID:         AccountFind,
						IsRequired: false,
					},
					{
						ID:         AccountEdit,
						IsRequired: false,
					},
				},
				SubResourceTypes: nil,
			},
			{
				ResourceID: CloudSelectionScheme,
				Actions: []client.CreatorRelatedAction{
					{
						ID:         CloudSelectionSchemeFind,
						IsRequired: false,
					},
					{
						ID:         CloudSelectionSchemeEdit,
						IsRequired: false,
					},
					{
						ID:         CloudSelectionSchemeDelete,
						IsRequired: false,
					},
				},
				SubResourceTypes: nil,
			},
			{
				ResourceID: MainAccount,
				Actions: []client.CreatorRelatedAction{
					{
						ID:         MainAccountFind,
						IsRequired: false,
					},
					{
						ID:         MainAccountEdit,
						IsRequired: false,
					},
				},
				SubResourceTypes: nil,
			},
			{
				ResourceID: RootAccount,
				Actions: []client.CreatorRelatedAction{
					{
						ID:         RootAccountFind,
						IsRequired: false,
					},
					{
						ID:         RootAccountEdit,
						IsRequired: false,
					},
				},
				SubResourceTypes: nil,
			},
		},
	}
}
