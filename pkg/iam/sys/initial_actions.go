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

var (
	accountResource = []iam.RelateResourceType{
		{
			SystemID: SystemIDHCM,
			ID:       Account,
			InstanceSelections: []iam.RelatedInstanceSelection{
				{
					SystemID: SystemIDHCM,
					ID:       AccountSelection,
				},
			},
		},
	}

	bizResource = []iam.RelateResourceType{
		{
			SystemID: SystemIDCMDB,
			ID:       Biz,
			InstanceSelections: []iam.RelatedInstanceSelection{
				{
					SystemID: SystemIDCMDB,
					ID:       BizSelection,
				},
			},
		},
	}

	schemeResource = []iam.RelateResourceType{
		{
			SystemID: SystemIDHCM,
			ID:       CloudSelectionScheme,
			InstanceSelections: []iam.RelatedInstanceSelection{
				{
					SystemID: SystemIDHCM,
					ID:       CloudSelectionSchemeSelection,
				},
			},
		},
	}

	mainaccountResource = []iam.RelateResourceType{
		{
			SystemID: SystemIDHCM,
			ID:       MainAccount,
			InstanceSelections: []iam.RelatedInstanceSelection{
				{
					SystemID: SystemIDHCM,
					ID:       MainAccountSelection,
				},
			},
		},
	}
	billCloudVendorResource = []iam.RelateResourceType{
		{
			SystemID: SystemIDHCM,
			ID:       BillCloudVendor,
			InstanceSelections: []iam.RelatedInstanceSelection{
				{
					SystemID: SystemIDHCM,
					ID:       BillCloudVendorSelection,
				},
			},
		},
	}
)

// GenerateStaticActions return need to register action.
func GenerateStaticActions() []iam.ResourceAction {
	resourceActionList := make([]iam.ResourceAction, 0)

	resourceActionList = append(resourceActionList, genResManagementActions()...)
	resourceActionList = append(resourceActionList, genResourceAccessActions()...)
	resourceActionList = append(resourceActionList, genCloudSelectionActions()...)
	resourceActionList = append(resourceActionList, genPlatformManageActions()...)
	resourceActionList = append(resourceActionList, genAccountManageActions()...)

	return resourceActionList
}

func genResManagementActions() []iam.ResourceAction {
	actions := []iam.ResourceAction{{
		ID:                   BizAccess,
		Name:                 ActionIDNameMap[BizAccess],
		NameEn:               "Access Biz",
		Type:                 View,
		RelatedResourceTypes: bizResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   BizIaaSResCreate,
		Name:                 ActionIDNameMap[BizIaaSResCreate],
		NameEn:               "Create Biz IaaS Resource",
		Type:                 Create,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}, {
		ID:                   BizIaaSResOperate,
		Name:                 ActionIDNameMap[BizIaaSResOperate],
		NameEn:               "Operate Biz IaaS Resource",
		Type:                 Edit,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}, {
		ID:                   BizIaaSResDelete,
		Name:                 ActionIDNameMap[BizIaaSResDelete],
		NameEn:               "Delete Biz IaaS Resource",
		Type:                 Delete,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}}

	actions = append(actions, genCLBResManActions()...)

	// 证书管理的Actions
	actions = append(actions, genCertResManActions()...)

	// TODO 开启编排相关功能后放开注释
	// 资源编排的Actions
	// actions = append(actions, genArrangeResManActions()...)
	actions = append(actions, []iam.ResourceAction{{
		ID:                   BizRecycleBinOperate,
		Name:                 ActionIDNameMap[BizRecycleBinOperate],
		NameEn:               "Operate Biz RecycleBin",
		Type:                 Edit,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}, {
		ID:                   BizRecycleBinConfig,
		Name:                 ActionIDNameMap[BizRecycleBinConfig],
		NameEn:               "Config Biz RecycleBin",
		Type:                 Edit,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}, {
		ID:                   BizOperationRecordFind,
		Name:                 ActionIDNameMap[BizOperationRecordFind],
		NameEn:               "Find Biz OperationRecord",
		Type:                 View,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}, {
		ID:                   BizTaskManagementOperate,
		Name:                 ActionIDNameMap[BizTaskManagementOperate],
		NameEn:               "Operate Biz TaskManagement",
		Type:                 Edit,
		RelatedResourceTypes: bizResource,
		RelatedActions:       []iam.ActionID{BizAccess},
		Version:              1,
	}}...)

	return actions
}

func genCLBResManActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   BizCLBResCreate,
			Name:                 ActionIDNameMap[BizCLBResCreate],
			NameEn:               "Create Biz CLB",
			Type:                 Create,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		}, {
			ID:                   BizCLBResOperate,
			Name:                 ActionIDNameMap[BizCLBResOperate],
			NameEn:               "Operate Biz CLB",
			Type:                 Edit,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		}, {
			ID:                   BizCLBResDelete,
			Name:                 ActionIDNameMap[BizCLBResDelete],
			NameEn:               "Delete Biz CLB",
			Type:                 Delete,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		},
	}
}

// genCertResManActions 业务-证书管理的Actions
func genCertResManActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   BizCertResCreate,
			Name:                 ActionIDNameMap[BizCertResCreate],
			NameEn:               "Create Biz Cert",
			Type:                 Create,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		}, {
			ID:                   BizCertResDelete,
			Name:                 ActionIDNameMap[BizCertResDelete],
			NameEn:               "Delete Biz Cert",
			Type:                 Delete,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		},
	}
}

/*
func genArrangeResManActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   BizArrangeResCreate,
			Name:                 ActionIDNameMap[BizArrangeResCreate],
			NameEn:               "Create Biz Arrange",
			Type:                 Create,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		}, {
			ID:                   BizArrangeResOperate,
			Name:                 ActionIDNameMap[BizArrangeResOperate],
			NameEn:               "Operate Biz Arrange",
			Type:                 Edit,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		}, {
			ID:                   BizArrangeResDelete,
			Name:                 ActionIDNameMap[BizArrangeResDelete],
			NameEn:               "Delete Biz Arrange",
			Type:                 Delete,
			RelatedResourceTypes: bizResource,
			RelatedActions:       []iam.ActionID{BizAccess},
			Version:              1,
		},
	}
}
*/

func genResourceAccessActions() []iam.ResourceAction {
	actions := []iam.ResourceAction{{
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
		ID:                   SubAccountEdit,
		Name:                 ActionIDNameMap[SubAccountEdit],
		NameEn:               "Edit Sub Account",
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
	actions = append(actions, genIaaSResAccessActions()...)
	// TODO 开启clb和编排相关功能后放开注释
	actions = append(actions, genCLBResAccessActions()...)
	actions = append(actions, genCertResAccessActions()...)
	actions = append(actions, []iam.ResourceAction{{
		ID:                   RecycleBinAccess,
		Name:                 ActionIDNameMap[RecycleBinAccess],
		NameEn:               "Find RecycleBin",
		Type:                 View,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   RecycleBinOperate,
		Name:                 ActionIDNameMap[RecycleBinOperate],
		NameEn:               "Operate RecycleBin",
		Type:                 Edit,
		RelatedResourceTypes: accountResource,
		RelatedActions:       []iam.ActionID{RecycleBinAccess},
		Version:              1,
	}, {
		ID:                   RecycleBinConfig,
		Name:                 ActionIDNameMap[RecycleBinConfig],
		NameEn:               "Config RecycleBin",
		Type:                 Edit,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   OperationRecordFind,
		Name:                 ActionIDNameMap[OperationRecordFind],
		NameEn:               "Find OperationRecord",
		Type:                 View,
		RelatedResourceTypes: accountResource,
		RelatedActions:       nil,
		Version:              1,
	}}...)
	actions = append(actions, genCosResAccessActions()...)
	return actions
}

func genCloudSelectionActions() []iam.ResourceAction {
	actions := []iam.ResourceAction{{
		ID:                   CloudSelectionRecommend,
		Name:                 ActionIDNameMap[CloudSelectionRecommend],
		NameEn:               "Selection Recommend",
		Type:                 View,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   CloudSelectionSchemeFind,
		Name:                 ActionIDNameMap[CloudSelectionSchemeFind],
		NameEn:               "Find Scheme",
		Type:                 View,
		RelatedResourceTypes: schemeResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   CloudSelectionSchemeEdit,
		Name:                 ActionIDNameMap[CloudSelectionSchemeEdit],
		NameEn:               "Edit Scheme",
		Type:                 Edit,
		RelatedResourceTypes: schemeResource,
		RelatedActions:       []iam.ActionID{CloudSelectionSchemeFind},
		Version:              1,
	}, {
		ID:                   CloudSelectionSchemeDelete,
		Name:                 ActionIDNameMap[CloudSelectionSchemeDelete],
		NameEn:               "Delete Scheme",
		Type:                 Delete,
		RelatedResourceTypes: schemeResource,
		RelatedActions:       []iam.ActionID{CloudSelectionSchemeFind},
		Version:              1,
	}}

	return actions
}

func genIaaSResAccessActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   ResourceFind,
			Name:                 ActionIDNameMap[ResourceFind],
			NameEn:               "Find Resource",
			Type:                 View,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{RecycleBinAccess},
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
			ID:                   IaaSResCreate,
			Name:                 ActionIDNameMap[IaaSResCreate],
			NameEn:               "Create IaaS Resource",
			Type:                 Create,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		}, {
			ID:                   IaaSResOperate,
			Name:                 ActionIDNameMap[IaaSResOperate],
			NameEn:               "Operate IaaS Resource",
			Type:                 Edit,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		}, {
			ID:                   IaaSResDelete,
			Name:                 ActionIDNameMap[IaaSResDelete],
			NameEn:               "Delete IaaS Resource",
			Type:                 Delete,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		},
	}
}

func genCLBResAccessActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   CLBResCreate,
			Name:                 ActionIDNameMap[CLBResCreate],
			NameEn:               "Create CLB",
			Type:                 Create,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		}, {
			ID:                   CLBResOperate,
			Name:                 ActionIDNameMap[CLBResOperate],
			NameEn:               "Operate CLB",
			Type:                 Edit,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		}, {
			ID:                   CLBResDelete,
			Name:                 ActionIDNameMap[CLBResDelete],
			NameEn:               "Delete CLB",
			Type:                 Delete,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		},
	}
}

// genCertResAccessActions 资源-证书管理的Actions
func genCertResAccessActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   CertResCreate,
			Name:                 ActionIDNameMap[CertResCreate],
			NameEn:               "Create Cert",
			Type:                 Create,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		}, {
			ID:                   CertResDelete,
			Name:                 ActionIDNameMap[CertResDelete],
			NameEn:               "Delete Cert",
			Type:                 Delete,
			RelatedResourceTypes: accountResource,
			RelatedActions:       []iam.ActionID{ResourceFind},
			Version:              1,
		},
	}
}

func genPlatformManageActions() []iam.ResourceAction {
	return []iam.ResourceAction{{
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
	}, {
		ID:                   GlobalConfiguration,
		Name:                 ActionIDNameMap[GlobalConfiguration],
		NameEn:               "Global Configuration",
		Type:                 View,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   RootAccountManage,
		Name:                 ActionIDNameMap[RootAccountManage],
		NameEn:               "Root Account Manage",
		Type:                 View,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   AccountBillManage,
		Name:                 ActionIDNameMap[AccountBillManage],
		NameEn:               "Account Bill Manage",
		Type:                 View,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   AccountBillPull,
		Name:                 ActionIDNameMap[AccountBillPull],
		NameEn:               "Account Bill Pull",
		Type:                 View,
		RelatedResourceTypes: billCloudVendorResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   ApplicationManage,
		Name:                 ActionIDNameMap[ApplicationManage],
		NameEn:               "Application Manage",
		Type:                 View,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	},
	}
}

func genAccountManageActions() []iam.ResourceAction {
	// MainAccount
	actions := []iam.ResourceAction{{
		ID:                   MainAccountFind,
		Name:                 ActionIDNameMap[MainAccountFind],
		NameEn:               "Find MainAccount",
		Type:                 View,
		RelatedResourceTypes: mainaccountResource,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   MainAccountCreate,
		Name:                 ActionIDNameMap[MainAccountCreate],
		NameEn:               "Create MainAccount",
		Type:                 Create,
		RelatedResourceTypes: nil,
		RelatedActions:       nil,
		Version:              1,
	}, {
		ID:                   MainAccountEdit,
		Name:                 ActionIDNameMap[MainAccountEdit],
		NameEn:               "Edit MainAccount",
		Type:                 Edit,
		RelatedResourceTypes: mainaccountResource,
		RelatedActions:       nil,
		Version:              1,
	},
	}

	return actions
}

func genCosResAccessActions() []iam.ResourceAction {
	return []iam.ResourceAction{
		{
			ID:                   CosBucketCreate,
			Name:                 ActionIDNameMap[CosBucketCreate],
			NameEn:               "COS Bucket Create",
			Type:                 Create,
			RelatedResourceTypes: accountResource,
			RelatedActions:       nil,
			Version:              1,
		}, {
			ID:                   CosBucketFind,
			Name:                 ActionIDNameMap[CosBucketFind],
			NameEn:               "COS Bucket Find",
			Type:                 View,
			RelatedResourceTypes: accountResource,
			RelatedActions:       nil,
			Version:              1,
		}, {
			ID:                   CosBucketDelete,
			Name:                 ActionIDNameMap[CosBucketDelete],
			NameEn:               "COS Bucket Delete",
			Type:                 Delete,
			RelatedResourceTypes: accountResource,
			RelatedActions:       nil,
			Version:              1,
		},
	}
}
