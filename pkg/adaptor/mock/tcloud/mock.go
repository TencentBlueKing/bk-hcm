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

package mocktcloud

import (
	"sync"

	adaptormock "hcm/pkg/adaptor/mock"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"

	"go.uber.org/mock/gomock"
)

var once sync.Once

// Playbook a playbook will use gomock EXPECT method to apply implements of method to mock instance
type Playbook interface {
	// Name  playbook identifier
	Name() string
	// Apply use for apply method implements to mock
	Apply(*MockTCloud, *gomock.Controller)
}

var ctrl *gomock.Controller
var mockCloud *MockTCloud

// GetMockCloud return a fake tcloud adaptor
func GetMockCloud() *MockTCloud {
	once.Do(initMock)
	return mockCloud
}

func getPlaybooks() []Playbook {
	// gomock是通过slice来记录同一个方法的多个Call实例的，先记录的调用的有更高的优先级
	return []Playbook{
		/* add playbook here */
		NewRegionPlaybook(),
		NewCrudVpcPlaybook(),
	}
}

func initMock() {
	ctrl = gomock.NewController(&adaptormock.LogReporter{})
	mockCloud = NewMockTCloud(ctrl)

	defaultPlaybook := getPlaybooks()
	for i, playbook := range defaultPlaybook {
		logs.V(3).Infof("[%s] registering %d th playbook: %s", enumor.TCloud, i, playbook.Name())
		playbook.Apply(mockCloud, ctrl)
	}

}

// Finish 检查mock调用次数是否符合预期，主要检查是否存在没有被调用的方法
func Finish() {
	if ctrl != nil {
		ctrl.Finish()
		ctrl = nil
	}
}
