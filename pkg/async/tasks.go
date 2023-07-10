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

package async

import "sync"

const (
	HcmTaskManager = "HCM"
	XXXTaskManager = "XXX"
)

var (
	taskManager *TaskManager
	once        sync.Once
)

// TaskManager store all tasks mapping
type TaskManager struct {
	Lock  sync.RWMutex
	Tasks map[string]map[string]interface{}
}

// GetTaskManager get task manager
func GetTaskManager() *TaskManager {
	once.Do(func() {
		taskManager = &TaskManager{
			Lock:  sync.RWMutex{},
			Tasks: make(map[string]map[string]interface{}),
		}
	})
	return taskManager
}

// RegisterTask register task by type
func (t *TaskManager) RegisterTask(taskManagerType, taskName string, i interface{}) {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	if _, exsit := t.Tasks[taskManagerType]; !exsit {
		t.Tasks[taskManagerType] = make(map[string]interface{})
	}

	t.Tasks[taskManagerType][taskName] = i
}

// GetAllTasksByManagerType get all tasks by type
func (t *TaskManager) GetAllTasksByManagerType(taskManagerType string) map[string]interface{} {
	t.Lock.RLock()
	defer t.Lock.RUnlock()

	if _, exsit := t.Tasks[taskManagerType]; !exsit {
		return nil
	}

	return t.Tasks[taskManagerType]
}
