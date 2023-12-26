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

package action

import (
	"fmt"
	"sync"

	"hcm/pkg/criteria/enumor"
)

// manager ...
var manager *Manager

func init() {
	manager = NewManager()
}

// Manager action manager
type Manager struct {
	actionMap  map[enumor.ActionName]Action
	flowTplMap map[enumor.FlowName]FlowTemplate
	rwLock     *sync.RWMutex
}

// NewManager 创建action管理器
func NewManager() *Manager {
	return &Manager{
		actionMap:  make(map[enumor.ActionName]Action),
		flowTplMap: make(map[enumor.FlowName]FlowTemplate),
		rwLock:     &sync.RWMutex{},
	}
}

// RegisterAction 注册到ActionMap
func (am *Manager) RegisterAction(acts ...Action) error {
	am.rwLock.Lock()
	defer am.rwLock.Unlock()

	for _, act := range acts {
		if err := act.Name().Validate(); err != nil {
			return err
		}

		am.actionMap[act.Name()] = act
	}

	return nil
}

// GetAction 执行时根据注册的名字获取执行体
func (am *Manager) GetAction(name enumor.ActionName) (Action, bool) {
	am.rwLock.RLock()
	defer am.rwLock.RUnlock()

	act, ok := am.actionMap[name]
	return act, ok
}

// RegisterFlowTpl register flow templates.
func (am *Manager) RegisterFlowTpl(templates ...FlowTemplate) error {
	am.rwLock.Lock()
	defer am.rwLock.Unlock()

	for _, tpl := range templates {
		// actionID 唯一性校验
		taskMap := make(map[ActIDType]bool)
		for _, one := range tpl.Tasks {
			if taskMap[one.ActionID] {
				return fmt.Errorf("actionID: %s repeat", one.ActionID)
			}

			if err := one.ActionName.Validate(); err != nil {
				return err
			}

			taskMap[one.ActionID] = true
		}

		// dependOn 依赖ActionID存在校验
		for _, task := range tpl.Tasks {
			for _, one := range task.DependOn {
				if !taskMap[one] {
					return fmt.Errorf("dependOn's actionID: %s not exist", one)
				}
			}
		}

		// 校验模板合法性
		if err := tpl.Validate(); err != nil {
			return err
		}
		am.flowTplMap[tpl.Name] = tpl
	}

	return nil
}

// GetFlowTpl get flow template by name.
func (am *Manager) GetFlowTpl(name enumor.FlowName) (FlowTemplate, bool) {
	am.rwLock.RLock()
	defer am.rwLock.RUnlock()

	tpl, ok := am.flowTplMap[name]
	return tpl, ok
}

// RegisterAction register action.
func RegisterAction(acts ...Action) {
	if err := manager.RegisterAction(acts...); err != nil {
		panic(fmt.Sprintf("register action failed, err: %v", err))
	}
}

// GetAction get action by name.
func GetAction(name enumor.ActionName) (Action, bool) {
	return manager.GetAction(name)
}

// RegisterTpl register flow template.
func RegisterTpl(templates ...FlowTemplate) {
	if err := manager.RegisterFlowTpl(templates...); err != nil {
		panic(fmt.Sprintf("register template flow failed, err: %v", err))
	}
}

// GetTpl get flow template by name.
func GetTpl(name enumor.FlowName) (FlowTemplate, bool) {
	return manager.GetFlowTpl(name)
}
