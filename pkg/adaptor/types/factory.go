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

// Package types defines cloud resource adaptor protocol.
package types

import (
	"errors"
	"fmt"
	"sync"

	"hcm/pkg/criteria/enumor"
)

// Factory holds all the supported operation on the cloud resources.
type Factory interface {
	AccountInterface
}

// FactoryManager manage all cloud vendor factory.
type FactoryManager struct {
	vendor map[enumor.Vendor]Factory
	lock   *sync.Mutex
}

// NewFactoryManager new factory manager.
func NewFactoryManager() *FactoryManager {
	return &FactoryManager{
		vendor: make(map[enumor.Vendor]Factory, 0),
		lock:   new(sync.Mutex),
	}
}

// RegisterVendor register a cloud vendor with factory
func (fm *FactoryManager) RegisterVendor(v enumor.Vendor, f Factory) error {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	_, exist := fm.vendor[v]
	if exist {
		return fmt.Errorf("%s vendor has already been registered, can not be registered again", v)
	}

	if f == nil {
		return errors.New("the registered vendor factory is nil")
	}

	fm.vendor[v] = f

	return nil
}

// Vendor returns the according vendor's factory, if vendor not exist return nil.
func (fm *FactoryManager) Vendor(v enumor.Vendor) (bool, Factory) {
	f, exist := fm.vendor[v]
	if !exist {
		return false, nil
	}

	return true, f
}
