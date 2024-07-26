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
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/uuid"
)

// BuildMonthTask build month task
func BuildMonthTask(
	t enumor.MonthTaskType, rootAccountID string,
	vendor enumor.Vendor, billYear, billMonth int) ts.CustomFlowTask {

	return ts.CustomFlowTask{
		ActionID:   action.ActIDType(uuid.UUID()),
		ActionName: enumor.ActionMonthTaskAction,
		Params: MonthTaskActionOption{
			Type:          t,
			RootAccountID: rootAccountID,
			Vendor:        vendor,
			BillYear:      billYear,
			BillMonth:     billMonth,
		},
	}
}
