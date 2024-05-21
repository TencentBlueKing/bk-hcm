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
	"hcm/pkg/api/core"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/kit"
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
	listScheduledFlowLimit = 10

	// listExpiredTasksLimit 每次WatchDog查询超时任务的数量
	listExpiredTasksLimit = 100
)

// Flow 消费所需的异步任务流。
type Flow struct {
	model.Flow `json:",inline"`

	Kit *kit.Kit `json:"-"`
}
