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
	monthTaskDescriberRegistry[enumor.Aws] = &AwsMonthDescriber{}
}

// AwsMonthDescriber aws month task describer
type AwsMonthDescriber struct {
}

// GetMonthTaskTypes aws month tasks
func (aws *AwsMonthDescriber) GetMonthTaskTypes() []enumor.MonthTaskType {
	return []enumor.MonthTaskType{enumor.AwsSavingsPlansMonthTask, enumor.AwsSupportMonthTask}
}

// GetTaskExtension extension for task
func (aws *AwsMonthDescriber) GetTaskExtension(rootAccountCloudID string) (map[string]string, error) {
	// set exclude account id
	excludeCloudIds := cc.AccountServer().BillAllocation.AwsCommonExpense.ExcludeAccountCloudIDs
	var spArnPrefix, spAccountCloudID string
	// matching saving plan allocation option
	for _, spOpt := range cc.AccountServer().BillAllocation.AwsSavingsPlans {
		if spOpt.RootAccountCloudID != rootAccountCloudID {
			continue
		}
		spAccountCloudID = spOpt.SpPurchaseAccountCloudID
		spArnPrefix = spOpt.SpArnPrefix
	}

	return map[string]string{
		constant.AwsCommonExpenseExcludeCloudIDKey: strings.Join(excludeCloudIds, ","),
		constant.AwsSavingsPlanARNPrefixKey:        spArnPrefix,
		constant.AwsSavingsPlanAccountCloudIDKey:   spAccountCloudID,
	}, nil
}
