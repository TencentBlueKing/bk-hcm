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

	"hcm/pkg/api/core"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"
)

// NewKit new async kit.
func NewKit() *kit.Kit {
	return core.NewBackendKit()
}

var (
	// DefRetryCount 任务执行失败默认重试次数
	DefRetryCount = uint(3)
	// DefRetryRangeMS 任务执行失败默认重试周期
	DefRetryRangeMS = [2]uint{1000, 15000}
)

const (
	// ErrTaskExecTimeout 任务执行超时
	ErrTaskExecTimeout = "task exec timeout"
	// ErrTaskNodeShutdown 任务节点关闭
	ErrTaskNodeShutdown = "task node shutdown"
	// ErrSomeTaskExecFailed 部分任务执行失败
	ErrSomeTaskExecFailed = "some tasks failed to be executed"

	//  listScheduledFlowLimit 每次调度器查询分配给当前节点的任务流数量
	listScheduledFlowLimit = 20

	// listExpiredTasksLimit 每次WatchDog查询超时任务的数量
	listExpiredTasksLimit = 100
)

// Flow 消费所需的异步任务流。
type Flow struct {
	model.Flow `json:",inline"`

	Kit *kit.Kit `json:"-"`
}

// SetFlowTypePriorityOption 设置优先级，Persistent为false时仅在内存临时生效，否则会持久化到db的global_config表
type SetFlowTypePriorityOption struct {
	FlowType   enumor.FlowName `json:"flow_type" validate:"required"`
	Priority   *int            `json:"priority" validate:"required"`
	Persistent bool            `json:"persistent" validate:"required"`
}

// Validate AddTemplateFlowOption
func (opt *SetFlowTypePriorityOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}
	if err := opt.FlowType.Validate(); err != nil {
		return err
	}
	if converter.PtrToVal(opt.Priority) > FlowTypeMinPriority {
		return fmt.Errorf("priority should be less than or equal to %d", FlowTypeMinPriority)
	}
	return nil
}

// ResetFlowPriorityOption 恢复默认优先级
type ResetFlowPriorityOption struct {
	FlowType enumor.FlowName `json:"flow_type" validate:"required"`
}

// Validate AddTemplateFlowOption
func (opt *ResetFlowPriorityOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}
	if err := opt.FlowType.Validate(); err != nil {
		return err
	}
	return nil
}
