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

package consumer

import (
	"fmt"
	"testing"
	"time"

	"hcm/pkg/async/task"
	"hcm/pkg/criteria/enumor"
)

func TestExecutor(t *testing.T) {
	// 初始化
	bd := new(mockBackend)
	exec := NewExecutor(bd, 5, time.Duration(15*time.Second))
	psr := new(mockParser)
	exec.SetGetParserFunc(func() Parser {
		return psr
	})
	task.ActionManagerInstance.RegisterAction(NewPrintTask())

	exec.Start()

	tasks := buildTestTaskData(1)

	for index := range tasks {
		exec.Push(tasks[index])
	}

	time.Sleep(10 * time.Second)
}

func buildTestTaskData(count int) []*task.Task {

	ans := make([]*task.Task, 0)
	for i := 0; i < count; i++ {
		asyncKit := NewKit()
		ans = append(ans, &task.Task{
			ID:          fmt.Sprintf("task-%d", i),
			FlowID:      fmt.Sprintf("flow-%d", i),
			FlowName:    fmt.Sprintf("flow-%d", i),
			ActionName:  string(enumor.TestPrintTask),
			Params:      nil,
			RetryCount:  0,
			TimeoutSecs: 5,
			DependOn:    make([]string, 0),
			State:       enumor.TaskPending,
			Memo:        "",
			ShareData:   nil,
			Kit:         asyncKit,
		})
	}

	return ans
}
