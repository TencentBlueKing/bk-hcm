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

var (
	accountResource = []client.RelateResourceType{
		{
			SystemID: SystemIDHCM,
			ID:       Account,
			InstanceSelections: []client.RelatedInstanceSelection{
				{
					SystemID: SystemIDHCM,
					ID:       AccountSelection,
				},
			},
		},
	}
)

// GenerateStaticActions return need to register action.
func GenerateStaticActions() []client.ResourceAction {
	resourceActionList := make([]client.ResourceAction, 0)

	resourceActionList = append(resourceActionList, genAccountActions()...)
	resourceActionList = append(resourceActionList, genResourceActions()...)

	return resourceActionList
}

func genAccountActions() []client.ResourceAction {
	return []client.ResourceAction{{
		ID:                   AccountFind,
		Name:                 ActionIDNameMap[AccountFind],
		NameEn:               "Find Account",
		Type:                 View,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   AccountKeyAccess,
		Name:                 ActionIDNameMap[AccountKeyAccess],
		NameEn:               "Access Account Key",
		Type:                 View,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   AccountCreate,
		Name:                 ActionIDNameMap[AccountCreate],
		NameEn:               "Create Account",
		Type:                 Create,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   AccountEdit,
		Name:                 ActionIDNameMap[AccountEdit],
		NameEn:               "Edit Account",
		Type:                 Edit,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   AccountDelete,
		Name:                 ActionIDNameMap[AccountDelete],
		NameEn:               "Delete Account",
		Type:                 Delete,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}}
}

func genResourceActions() []client.ResourceAction {
	return []client.ResourceAction{{
		ID:                   ResourceFind,
		Name:                 ActionIDNameMap[ResourceFind],
		NameEn:               "Find Resource",
		Type:                 View,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   ResourceAssign,
		Name:                 ActionIDNameMap[ResourceAssign],
		NameEn:               "Assign Resource To Business",
		Type:                 Edit,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   ResourceManage,
		Name:                 ActionIDNameMap[ResourceManage],
		NameEn:               "Manage Resource",
		Type:                 Edit,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   ResourceRecycle,
		Name:                 ActionIDNameMap[ResourceRecycle],
		NameEn:               "Recycle Resource From Business",
		Type:                 Edit,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}}
}
