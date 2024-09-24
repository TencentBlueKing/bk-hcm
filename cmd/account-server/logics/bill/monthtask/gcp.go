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
	"fmt"
	"strings"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/json"
)

func init() {
	monthTaskDescriberRegistry[enumor.Gcp] = &GcpMonthDescriber{}
}

// GcpMonthDescriber gcp month task describer
type GcpMonthDescriber struct {
}

// GetTaskExtension ...
func (gcp *GcpMonthDescriber) GetTaskExtension(rootAccountCloudID string) (map[string]string, error) {

	// set exclude account id
	excludeCloudIds := cc.AccountServer().BillAllocation.GcpCommonExpense.ExcludeAccountCloudIDs
	var creditReturnConfig string
	var err error
	// matching saving plan allocation option
	for _, creditConfigs := range cc.AccountServer().BillAllocation.GcpCredits {
		if creditConfigs.RootAccountCloudID != rootAccountCloudID {
			continue
		}
		creditReturnConfig, err = json.MarshalToString(creditConfigs.ReturnConfigs)
		if err != nil {
			return nil, fmt.Errorf("fail to marshal gcp credit return config to json :%w", err)
		}
	}

	return map[string]string{
		constant.GcpCommonExpenseExcludeCloudIDKey: strings.Join(excludeCloudIds, ","),
		constant.GcpCreditReturnConfigKey:          creditReturnConfig,
	}, nil
}

// GetMonthTaskTypes gcp month tasks
func (gcp *GcpMonthDescriber) GetMonthTaskTypes() []enumor.MonthTaskType {
	return []enumor.MonthTaskType{enumor.GcpCreditsMonthTask, enumor.GcpSupportMonthTask}
}
