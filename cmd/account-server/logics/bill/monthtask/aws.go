/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package monthtask

import (
	"strings"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

func init() {
	monthTaskDescriberRegistry[enumor.Aws] = NewAwsMonthDescriber
}

// NewAwsMonthDescriber ...
func NewAwsMonthDescriber(rootAccountCloudID string) MonthTaskDescriber {
	describer := &awsMonthDescriber{RootAccountCloudID: rootAccountCloudID}
	// set exclude account id
	describer.CommonExpenseExcludeCloudIDs = cc.AccountServer().BillAllocation.AwsCommonExpense.ExcludeAccountCloudIDs
	// matching saving plan allocation option
	for _, spOpt := range cc.AccountServer().BillAllocation.AwsSavingsPlans {
		if spOpt.RootAccountCloudID != rootAccountCloudID {
			continue
		}
		describer.SpAccountCloudID = spOpt.SpPurchaseAccountCloudID
		describer.SpArnPrefix = spOpt.SpArnPrefix
	}

	// 过滤符合条件的账号抵扣配置
	describer.RootAccountDeductItemTypes = cc.AccountServer().BillAllocation.AwsDeductAccountItems.DeductItemTypes
	filterDeductItemTypes := make(map[string][]string)
	for rootAccountKey, rootAccountItem := range describer.RootAccountDeductItemTypes {
		if rootAccountKey != rootAccountCloudID {
			continue
		}
		filterDeductItemTypes = rootAccountItem
	}
	awsDeductItemTypes, err := json.Marshal(filterDeductItemTypes)
	if err != nil {
		logs.Warnf("fail to json marshal awsDeductAccountItems config, err: %v", err)
	}
	describer.DeductItemTypes = string(awsDeductItemTypes)

	return describer
}

// awsMonthDescriber aws month task describer
type awsMonthDescriber struct {
	RootAccountCloudID           string
	SpArnPrefix                  string
	SpAccountCloudID             string
	CommonExpenseExcludeCloudIDs []string
	RootAccountDeductItemTypes   map[string]map[string][]string // 根账号需要抵扣的项目列表
	DeductItemTypes              string                         // 需要抵扣的账单明细项目类型列表，比如税费Tax
}

// GetMonthTaskTypes aws month tasks
// 这里的MonthTaskType配置，有严格顺序，修改时需要注意下
func (aws *awsMonthDescriber) GetMonthTaskTypes() []enumor.MonthTaskType {
	monthTaskTypes := []enumor.MonthTaskType{
		enumor.AwsOutsideBillMonthTask,
	}

	if aws.SpArnPrefix != "" {
		monthTaskTypes = append(monthTaskTypes, enumor.AwsSavingsPlansMonthTask)
	}

	// 根账号配置了需要抵扣的项目
	if len(aws.RootAccountDeductItemTypes) > 0 {
		if _, ok := aws.RootAccountDeductItemTypes[aws.RootAccountCloudID]; ok {
			monthTaskTypes = append(monthTaskTypes, enumor.DeductMonthTask)
		}
	}

	monthTaskTypes = append(monthTaskTypes, enumor.AwsSupportMonthTask)
	return monthTaskTypes
}

// GetTaskExtension extension for task
func (aws *awsMonthDescriber) GetTaskExtension() (map[string]string, error) {

	return map[string]string{
		constant.AwsCommonExpenseExcludeCloudIDKey: strings.Join(aws.CommonExpenseExcludeCloudIDs, ","),
		constant.AwsSavingsPlanARNPrefixKey:        aws.SpArnPrefix,
		constant.AwsSavingsPlanAccountCloudIDKey:   aws.SpAccountCloudID,
		constant.AwsAccountDeductItemTypesKey:      aws.DeductItemTypes,
	}, nil
}
