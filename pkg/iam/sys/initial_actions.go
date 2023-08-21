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

	bizResource = []client.RelateResourceType{
		{
			SystemID: SystemIDCMDB,
			ID:       Biz,
			InstanceSelections: []client.RelatedInstanceSelection{
				{
					SystemID: SystemIDCMDB,
					ID:       BizSelection,
				},
			},
		},
	}
)

// GenerateStaticActions return need to register action.
func GenerateStaticActions() []client.ResourceAction {
	resourceActionList := make([]client.ResourceAction, 0)

	resourceActionList = append(resourceActionList, genBizActions()...)
	resourceActionList = append(resourceActionList, genAccountActions()...)
	resourceActionList = append(resourceActionList, genResourceActions()...)
	resourceActionList = append(resourceActionList, genRecycleBinActions()...)
	resourceActionList = append(resourceActionList, genAuditActions()...)
	resourceActionList = append(resourceActionList, genSystemManageActions()...)

	return resourceActionList
}

func genBizActions() []client.ResourceAction {
	return []client.ResourceAction{{
		ID:                   BizAccess,
		Name:                 ActionIDNameMap[BizAccess],
		NameEn:               "Access Biz",
		Type:                 View,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []client.ActionID{ResourceAssign},
		Version:              1,
	}, {
		ID:                   BizIaaSResCreate,
		Name:                 ActionIDNameMap[BizIaaSResCreate],
		NameEn:               "Create Biz IaaS Resource",
		Type:                 Create,
		RelatedResourceTypes: bizResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   BizIaaSResOperate,
		Name:                 ActionIDNameMap[BizIaaSResOperate],
		NameEn:               "Operate Biz IaaS Resource",
		Type:                 Edit,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []client.ActionID{BizAccess},
		Version:              1,
	}, {
		ID:                   BizIaaSResDelete,
		Name:                 ActionIDNameMap[BizIaaSResDelete],
		NameEn:               "Delete Biz IaaS Resource",
		Type:                 Delete,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []client.ActionID{BizAccess},
		Version:              1,
	}}
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
		ID:                   AccountImport,
		Name:                 ActionIDNameMap[AccountImport],
		NameEn:               "Import Account",
		Type:                 Create,
		RelatedResourceTypes: nil,
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
		RelatedActions:       []client.ActionID{RecycleBinFind},
		Version:              1,
	}, {
		ID:                   ResourceAssign,
		Name:                 ActionIDNameMap[ResourceAssign],
		NameEn:               "Assign Resource To Business",
		Type:                 Edit,
		RelatedResourceTypes: append(accountResource, bizResource...),
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   IaaSResourceCreate,
		Name:                 ActionIDNameMap[IaaSResourceCreate],
		NameEn:               "Create IaaS Resource",
		Type:                 Create,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   IaaSResourceOperate,
		Name:                 ActionIDNameMap[IaaSResourceOperate],
		NameEn:               "Operate IaaS Resource",
		Type:                 Edit,
		RelatedResourceTypes: accountResource,
		RelatedActions:       []client.ActionID{ResourceFind},
		Version:              1,
	}, {
		ID:                   IaaSResourceDelete,
		Name:                 ActionIDNameMap[IaaSResourceDelete],
		NameEn:               "Delete IaaS Resource",
		Type:                 Delete,
		RelatedResourceTypes: accountResource,
		RelatedActions:       []client.ActionID{ResourceFind},
		Version:              1,
	}}
}

func genRecycleBinActions() []client.ResourceAction {
	return []client.ResourceAction{{
		ID:                   RecycleBinFind,
		Name:                 ActionIDNameMap[RecycleBinFind],
		NameEn:               "Find Resource In Recycle Bin",
		Type:                 View,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   RecycleBinManage,
		Name:                 ActionIDNameMap[RecycleBinManage],
		NameEn:               "Manage Resource In Recycle Bin",
		Type:                 Edit,
		RelatedResourceTypes: nil,
		RelatedActions:       []client.ActionID{RecycleBinFind},
		Version:              1,
	}}
}

func genAuditActions() []client.ResourceAction {
	return []client.ResourceAction{{
		ID:                   BizAuditFind,
		Name:                 ActionIDNameMap[BizAuditFind],
		NameEn:               "Find Biz audit Log",
		Type:                 View,
		RelatedResourceTypes: bizResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   ResourceAuditFind,
		Name:                 ActionIDNameMap[ResourceAuditFind],
		NameEn:               "Find Resource audit Log",
		Type:                 View,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}}
}

func genSystemManageActions() []client.ResourceAction {
	return []client.ResourceAction{{
		ID:                   CostManage,
		Name:                 ActionIDNameMap[CostManage],
		NameEn:               "Cost Manage",
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
	}}
}
