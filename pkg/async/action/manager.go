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
	flowTplMap map[enumor.FlowTplName]FlowTemplate
	rwLock     *sync.RWMutex
}

// NewManager 创建action管理器
func NewManager() *Manager {
	return &Manager{
		actionMap:  make(map[enumor.ActionName]Action),
		flowTplMap: make(map[enumor.FlowTplName]FlowTemplate),
		rwLock:     &sync.RWMutex{},
	}
}

// RegisterAction 注册到ActionMap
func (am *Manager) RegisterAction(acts ...Action) {
	am.rwLock.Lock()
	defer am.rwLock.Unlock()

	for _, act := range acts {
		am.actionMap[act.Name()] = act
	}
}

// GetAction 执行时根据注册的名字获取执行体
func (am *Manager) GetAction(name enumor.ActionName) (Action, bool) {
	am.rwLock.RLock()
	defer am.rwLock.RUnlock()

	act, ok := am.actionMap[name]
	return act, ok
}

// RegisterFlowTpl register flow templates.
func (am *Manager) RegisterFlowTpl(templates ...FlowTemplate) {
	am.rwLock.Lock()
	defer am.rwLock.Unlock()

	for _, tpl := range templates {
		am.flowTplMap[tpl.Name] = tpl
	}
}

// GetFlowTpl get flow template by name.
func (am *Manager) GetFlowTpl(name enumor.FlowTplName) (FlowTemplate, bool) {
	am.rwLock.RLock()
	defer am.rwLock.RUnlock()

	tpl, ok := am.flowTplMap[name]
	return tpl, ok
}

// RegisterAction register action.
func RegisterAction(acts ...Action) {
	manager.RegisterAction(acts...)
}

// GetAction get action by name.
func GetAction(name enumor.ActionName) (Action, bool) {
	return manager.GetAction(name)
}

// RegisterTpl register flow template.
func RegisterTpl(templates ...FlowTemplate) {
	manager.RegisterFlowTpl(templates...)
}

// GetTpl get flow template by name.
func GetTpl(name enumor.FlowTplName) (FlowTemplate, bool) {
	return manager.GetFlowTpl(name)
}
