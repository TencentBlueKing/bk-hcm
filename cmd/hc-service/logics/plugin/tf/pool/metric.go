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
	"errors"

	"hcm/pkg/metrics"
)

// 指标定义在 pkg/metrics/plugin_pool.go，这里只有结果取值的归类与状态快照。

// acquire_seconds 的 result 取值。
const (
	acquireSuccess       = "success"
	acquireBusinessError = "business_error"
	acquireProcessError  = "process_error"
	acquireTimeout       = "timeout"
	acquireCanceled      = "canceled"
	acquireQueueFull     = "queue_full"
	acquireClosed        = "closed"
	acquireInvalidConfig = "invalid_config"
)

// start_seconds 与 configure_seconds 的 result 取值。
const (
	resultSuccess        = "success"
	resultFailure        = "failure"
	resultBusinessError  = "business_error"
	resultTransportError = "transport_error"
)

// acquireFailure 把准备进程失败的错误归类为 acquire_seconds 的 result。
func acquireFailure(err error) string {
	switch {
	case errors.Is(err, ErrPoolClosed):
		return acquireClosed
	case isBusinessError(err):
		return acquireBusinessError
	default:
		return acquireProcessError
	}
}

// configureResult 把 ConfigureProvider 的结果归类为 configure_seconds 的 result。
func configureResult(err error) string {
	switch {
	case err == nil:
		return resultSuccess
	case isBusinessError(err):
		return resultBusinessError
	default:
		return resultTransportError
	}
}

// stats 在抓取时被调用，返回池的状态快照。
//
// 必须持锁：进程表是 map，与调度并发读写会直接导致进程崩溃，不只是读到旧值。
// 锁内只遍历进程表（不超过 MaxProcesses 个）、读队列长度与队首，开销是微秒级，抓取间隔是秒级，对调度无可感知的影响。
func (p *PluginPool) stats() metrics.PluginPoolStats {
	s := metrics.PluginPoolStats{
		Processes: map[string]int{
			string(stateStarting): 0, string(stateConfiguring): 0, string(stateRunning): 0, string(stateClosing): 0,
		},
		MaxProcesses:        p.cfg.MaxProcesses,
		MaxLeasesPerProcess: p.cfg.MaxLeasesPerProcess,
	}

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	s.QueueLength = p.queue.Len()
	if w, ok := p.queue.Peek(); ok {
		s.OldestWait = p.now().Sub(w.enqueuedAt)
	}
	for _, proc := range p.procs {
		s.Processes[string(proc.state)]++
		s.InflightLeases += proc.inflight
	}
	return s
}
