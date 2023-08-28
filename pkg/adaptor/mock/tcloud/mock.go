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
	adaptormock "hcm/pkg/adaptor/mock"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"

	"go.uber.org/mock/gomock"
)

// Playbook 给mock加上具体方法的实现
type Playbook interface {
	Name() string
	Apply(*MockTCloud, *gomock.Controller)
}

var defaultPlaybook = map[string]Playbook{}

func register(applier Playbook) {
	defaultPlaybook[applier.Name()] = applier
}

// NewMockCloud return fake adaptor
func NewMockCloud(playbooks ...Playbook) (*MockTCloud, *gomock.Controller) {
	ctrl := gomock.NewController(&adaptormock.LogReporter{})
	mockCloud := NewMockTCloud(ctrl)

	for _, playbook := range playbooks {
		register(playbook)
	}
	for name, applier := range defaultPlaybook {
		logs.V(3).Infoln(enumor.TCloud, "registering playbook:", name)
		applier.Apply(mockCloud, ctrl)
	}

	return mockCloud, ctrl
}

func init() {
	// register the playbook method
	register(NewRegionPlaybook())
	register(NewCrudVpcPlaybook(nil, nil, nil))
}
