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

// Package constant 账单状态package
package constant

// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版99:重新生成cur成本报告100:正常)
const (
	StatusDefault              = 0
	StatusCreateBucket         = 1
	StatusSetBucketPolicy      = 2
	StatusCreateCur            = 3
	StatusCheckYml             = 4
	StatusCreateCloudFormation = 5
	StatusReCurReport          = 99
	StatusSuccess              = 100
)

const (
	// BillExportFolderPrefix 账单导出文件夹前缀
	BillExportFolderPrefix = "bill_export"
)

// AwsSavingsPlanAccountCloudIDKey ...
const AwsSavingsPlanAccountCloudIDKey = "aws_savings_plan_account_cloud_id"

// AwsSavingsPlanARNPrefixKey ...
const AwsSavingsPlanARNPrefixKey = "aws_savings_plan_arn_prefix"

// AwsCommonExpenseExcludeCloudIDKey ...
const AwsCommonExpenseExcludeCloudIDKey = "aws_common_expense_exclude_account_cloud_id"

// AwsAccountDeductItemTypesKey AWS需要抵扣的账单明细项目类型列表，比如税费Tax
const AwsAccountDeductItemTypesKey = "aws_account_deduct_item_types"

// GcpCommonExpenseExcludeCloudIDKey ...
const GcpCommonExpenseExcludeCloudIDKey = "gcp_common_expense_exclude_account_cloud_id"

// GcpCreditReturnConfigKey ...
const GcpCreditReturnConfigKey = "gcp_credit_return_config"

// AwsLineItemTypeSavingPlanCoveredUsage aws savings plan cost line item type
const AwsLineItemTypeSavingPlanCoveredUsage = "SavingsPlanCoveredUsage"

// HuaweiCommonExpenseExcludeCloudIDKey ...
const HuaweiCommonExpenseExcludeCloudIDKey = "huawei_common_expense_exclude_account_cloud_id"

const (
	// BillOutsideMonthBillName outside bill month bill
	BillOutsideMonthBillName = "OutsideMonthBill"
	// BillCommonExpenseReverseName common expense reverse
	BillCommonExpenseReverseName = "CommonExpenseReverse"
	// BillCommonExpenseName common expense
	BillCommonExpenseName = "CommonExpense"

	// AwsSavingsPlansCostCode aws savings plans cost code,
	AwsSavingsPlansCostCode = "SavingsPlanCost"
	// AwsSavingsPlansCostCodeReverse aws savings plans cost code reverse
	AwsSavingsPlansCostCodeReverse = "SavingsPlanCostReverse"
	// AwsDeductCostCodeReverse aws deduct cost code reverse
	AwsDeductCostCodeReverse = "DeductCostReverse"

	// GcpCreditReturnCost Gcp credit return cost, negative value, e.g. -10.00000
	GcpCreditReturnCost = "Credit"

	// GcpCreditReturnCostReverse Gcp credit return cost reverse, positive value, e.g. 10.00000
	GcpCreditReturnCostReverse = "CreditReverse"
)

// HuaweiBillTypePurchase 华为账单类型-新购
const HuaweiBillTypePurchase = int32(1)

// HuaweiBillChargeModeMonthlyYearly 华为账单计费模式-包年包月
const HuaweiBillChargeModeMonthlyYearly = "1"
