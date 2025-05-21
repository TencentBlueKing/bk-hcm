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

package actioncvm

import (
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/uuid"
)

// BuildCreateCvmTasks build create cvm tasks.
func BuildCreateCvmTasks(totalCount int64, limit int64, assignCvmOpt *AssignCvmOption,
	convTask func(actionID action.ActIDType, count int64) ts.CustomFlowTask) []ts.CustomFlowTask {

	ctrFunc := counter.NewNumStringCounter(1, 10)
	actionIDs := make([]action.ActIDType, 0)
	tasks := make([]ts.CustomFlowTask, 0)
	for ; totalCount > 0; totalCount -= limit {
		requiredCount := int64(0)
		if totalCount > limit {
			requiredCount = limit
		} else {
			requiredCount = totalCount
		}

		actionID := action.ActIDType(ctrFunc())
		actionIDs = append(actionIDs, actionID)
		tasks = append(tasks, convTask(actionID, requiredCount))
	}

	if assignCvmOpt != nil {
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(uuid.UUID()),
			ActionName: enumor.ActionAssignCvm,
			Params:     assignCvmOpt,
			DependOn:   actionIDs,
		})
	}

	return tasks
}
