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

package monthtask

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
)

type gcpMonthTaskBaseRunner struct {
	creditReturnMap             map[string]string
	creditReturnAccountCloudIds []string
	excludeAccountCloudIds      []string
}

func (g *gcpMonthTaskBaseRunner) initExtension(opt *MonthTaskActionOption) error {
	g.creditReturnMap = make(map[string]string)
	if opt.Extension == nil {
		return nil
	}

	var creditReturnConfigs []cc.CreditReturn
	creditReturnConfigString := opt.Extension[constant.GcpCreditReturnConfigKey]
	err := json.Unmarshal([]byte(creditReturnConfigString), &creditReturnConfigs)
	if err != nil {
		return fmt.Errorf("fail to unmarshal gcp credit return config: %w", err)
	}

	for _, config := range creditReturnConfigs {
		g.creditReturnMap[config.CreditID] = config.AccountCloudID
		g.creditReturnAccountCloudIds = append(g.creditReturnAccountCloudIds, config.AccountCloudID)
	}
	if opt.Extension[constant.GcpCommonExpenseExcludeCloudIDKey] != "" {
		excludeCloudIDStr := opt.Extension[constant.GcpCommonExpenseExcludeCloudIDKey]
		excluded := strings.Split(excludeCloudIDStr, ",")
		g.excludeAccountCloudIds = excluded
	}
	return nil
}

func newGcpRunner(taskType enumor.MonthTaskType) (MonthTaskRunner, error) {
	switch taskType {
	case enumor.GcpSupportMonthTask:
		return &GcpSupportMonthTask{}, nil
	case enumor.GcpCreditsMonthTask:
		return &GcpCreditMonthTask{}, nil
	default:
		return nil, errors.New("not support task type of gcp: " + string(taskType))
	}

}
