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

package task

import (
	"sync"

	"hcm/pkg/kit"
)

var ActionManagerInstance *ActionManager

func init() {
	ActionManagerInstance = NewActionManager()
}

// Action 任务执行体接口定义
type Action interface {
	// Name action name
	Name() string
	// RunBefore run before func
	RunBefore(kt *kit.Kit, params interface{}) error
	// Run run func
	Run(kt *kit.Kit, params interface{}) error
	// RunBeforeSuccess run after success func
	RunBeforeSuccess(kt *kit.Kit, params interface{}) error
	// RunBeforeFailed run after failed func
	RunBeforeFailed(kt *kit.Kit, params interface{}) error
	// RetryBefore retry before func
	RetryBefore(kt *kit.Kit, params interface{}) error
}

// ActionManager action manager
type ActionManager struct {
	actionMap map[string]Action
	rwLock    *sync.RWMutex
}

func NewActionManager() *ActionManager {
	return &ActionManager{
		actionMap: make(map[string]Action),
		rwLock:    &sync.RWMutex{},
	}
}

// RegisterAction 注册到ActionMap
func (am *ActionManager) RegisterAction(acts ...Action) {
	am.rwLock.Lock()
	defer am.rwLock.Unlock()

	for _, act := range acts {
		am.actionMap[act.Name()] = act
	}
}

// GetAction 执行时根据注册的名字获取执行体
func (am *ActionManager) GetAction(name string) (Action, bool) {
	am.rwLock.RLock()
	defer am.rwLock.RUnlock()

	act, ok := am.actionMap[name]
	return act, ok
}
