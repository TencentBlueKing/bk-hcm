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

// Package action ...
package action

import (
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
)

// Action 异步任务必须实现的运行接口。
type Action interface {
	// Name 返回异步任务名称
	Name() enumor.ActionName
	// Run 异步任务运行操作
	Run(kt run.ExecuteKit, params interface{}) (result interface{}, err error)
}

// RollbackAction Action如果支持回滚操作，实现该接口。会在Action执行失败、Action执行一半崩溃后，进行调用。
// State: running -> rollback -> pending
type RollbackAction interface {
	Rollback(kt run.ExecuteKit, params interface{}) error
}

// ParameterAction 如果任务运行需要依赖请求参数，需要通过该接口返回参数结构，会将任务实例中的参数内容解析到这个返回参数上。
type ParameterAction interface {
	// ParameterNew 返回新的参数结构。返回参数可以实现 Decoder 接口，自定义解码方式。
	ParameterNew() (params interface{})
}
