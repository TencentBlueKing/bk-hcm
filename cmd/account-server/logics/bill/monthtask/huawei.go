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
	monthTaskDescriberRegistry[enumor.HuaWei] = NewHuaweiMonthDescriber
}

// NewHuaweiMonthDescriber ...
func NewHuaweiMonthDescriber(rootAccountCloudID string) MonthTaskDescriber {
	describer := &huaweiMonthDescriber{
		RootAccountCloudID: rootAccountCloudID,
	}
	// set exclude account id
	commonExpenseConfig := cc.AccountServer().BillAllocation.HuaweiCommonExpense
	describer.CommonExpenseExcludeCloudIDs = commonExpenseConfig.ExcludeAccountCloudIDs

	return describer
}

// huaweiMonthDescriber huawei month task describer
type huaweiMonthDescriber struct {
	RootAccountCloudID           string
	CommonExpenseExcludeCloudIDs []string
}

// GetMonthTaskTypes huawei month tasks
func (huawei *huaweiMonthDescriber) GetMonthTaskTypes() []enumor.MonthTaskType {
	// vendor huawei only have support month task type
	return []enumor.MonthTaskType{enumor.HuaweiSupportMonthTask}
}

// GetTaskExtension extension for task
func (huawei *huaweiMonthDescriber) GetTaskExtension() (map[string]string, error) {

	return map[string]string{
		constant.HuaweiCommonExpenseExcludeCloudIDKey: strings.Join(huawei.CommonExpenseExcludeCloudIDs, ","),
	}, nil
}
