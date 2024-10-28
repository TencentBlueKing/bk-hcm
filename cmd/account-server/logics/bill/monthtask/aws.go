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

	return describer
}

// awsMonthDescriber aws month task describer
type awsMonthDescriber struct {
	RootAccountCloudID           string
	SpArnPrefix                  string
	SpAccountCloudID             string
	CommonExpenseExcludeCloudIDs []string
}

// GetMonthTaskTypes aws month tasks
func (aws *awsMonthDescriber) GetMonthTaskTypes() []enumor.MonthTaskType {
	if aws.SpArnPrefix == "" {
		return []enumor.MonthTaskType{
			enumor.AwsOutsideBillMonthTask,
			// 没有配置sp前缀则不生成对应的sp分账任务
			// enumor.AwsSavingsPlansMonthTask,
			enumor.AwsSupportMonthTask,
		}
	}
	return []enumor.MonthTaskType{
		enumor.AwsOutsideBillMonthTask,
		enumor.AwsSavingsPlansMonthTask,
		enumor.AwsSupportMonthTask,
	}
}

// GetTaskExtension extension for task
func (aws *awsMonthDescriber) GetTaskExtension() (map[string]string, error) {

	return map[string]string{
		constant.AwsCommonExpenseExcludeCloudIDKey: strings.Join(aws.CommonExpenseExcludeCloudIDs, ","),
		constant.AwsSavingsPlanARNPrefixKey:        aws.SpArnPrefix,
		constant.AwsSavingsPlanAccountCloudIDKey:   aws.SpAccountCloudID,
	}, nil
}
