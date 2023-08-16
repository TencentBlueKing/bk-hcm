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

package actions

import (
	"fmt"
	"time"

	"hcm/cmd/task-server/logics/async/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

func init() {
	task.ActionManagerInstance.RegisterAction(NewCreateSG())
}

type CreateSG struct {
	ShareData map[string]interface{}
}

func NewCreateSG() *CreateSG {
	return &CreateSG{
		ShareData: make(map[string]interface{}),
	}
}

func (c *CreateSG) Name() string {
	return string(enumor.TestCreateSG)
}

func (c *CreateSG) RunBefore(kt *kit.Kit, params interface{}) error {
	return nil
}

func (c *CreateSG) Run(kt *kit.Kit, params interface{}) error {
	fmt.Println("run create sg", time.Now())
	return nil
}

func (c *CreateSG) RunBeforeSuccess(kt *kit.Kit, params interface{}) error {
	return nil
}

func (c *CreateSG) RunBeforeFailed(kt *kit.Kit, params interface{}) error {
	return nil
}

func (c *CreateSG) RetryBefore(kt *kit.Kit, params interface{}) error {
	return nil
}
