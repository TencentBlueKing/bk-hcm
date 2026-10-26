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

package pool

import (
	"time"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/cmd/hc-service/logics/tfruntime/executor"
)

// procState 插件进程在池中的状态。
type procState string

const (
	// stateStarting 已占用容量，正在拉起进程，尚无执行器。
	stateStarting procState = "starting"
	// stateConfiguring 正在 ConfigureProvider（含重配），不能交付、不能重配、不能回收。
	stateConfiguring procState = "configuring"
	// stateRunning 可交付或可配置；空闲 / 使用中 / 已满由 binding 与 inflight 推导。
	stateRunning procState = "running"
	// stateClosing 正在关闭，确认进程退出前仍占用容量。
	stateClosing procState = "closing"
)

// binding 进程上的 Provider 配置。
//
// 进程的 binding 为 nil 表示未配置；boundAt 为零值表示正在配置成 key、尚未生效；
// boundAt 非零表示已生效，是否过期由 MaxConfigAge 判断。
// 配置开始时就写入目标（旧配置从此不可信），成功时补 boundAt，失败时整体清空。
type binding struct {
	// config 含凭证，不得进入日志与错误信息。
	config  map[string]any
	key     string
	boundAt time.Time
}

// configuring 是否正在配置、尚未生效。
func (b *binding) configuring() bool {
	return b != nil && b.boundAt.IsZero()
}

// valid 配置是否已生效且未过期。
func (b *binding) valid(now time.Time, maxAge time.Duration) bool {
	return b != nil && !b.boundAt.IsZero() && now.Before(b.boundAt.Add(maxAge))
}

// pluginProcess 池内一个插件进程的记录，所有字段受 PluginPool.poolLock 保护。
//
// 每次释放锁时成立：
//   - 0 <= inflight <= MaxLeasesPerProcess，starting / configuring 期间 inflight 为 0；
//   - pending 只在 starting / configuring 期间非空，长度不超过 MaxLeasesPerProcess；
//   - running 时 binding 为 nil 或已生效。
type pluginProcess struct {
	// id 单调递增、不复用：lease 按 id 归还，进程退出后同一个 id 不会指到新进程。
	id    uint64
	state procState

	binding *binding
	// inflight 已交付、未归还的 lease 数。
	inflight int
	// pending 本次准备（创建或配置）成功后要交付的请求，按到达顺序，含触发者。已取消的留在其中，交付时跳过。
	pending []*request
	// idleSince 最近一次变为空闲的时间。
	idleSince time.Time

	// manager 创建任务拿到后写入，之后不再改变；starting 期间为 nil。
	manager *tfruntime.ProcessManager
	// exec 进程拉起后写入；starting 期间为 nil。
	exec executor.Executor
}

// hasLivePending 是否还有未取消的请求在等这次准备。
func (proc *pluginProcess) hasLivePending() bool {
	for _, w := range proc.pending {
		if w.state != requestDone {
			return true
		}
	}
	return false
}

// idle 是否空闲：running 且没有在途 lease，不论是否已配置。
func (proc *pluginProcess) idle() bool {
	return proc.state == stateRunning && proc.inflight == 0
}
