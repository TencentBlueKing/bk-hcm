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
	"fmt"
	"testing"
)

// checkInvariantsLocked 校验池的不变式（技术方案 6.11），调用方持有 p.poolLock。
//
//	I1  0 <= inflight <= MaxLeasesPerProcess
//	I3  进程数不超过 MaxProcesses
//	I4  starting / configuring 有且只有一个任务；running 没有任务
//	I6  请求至多处于一处：队列、某进程的 pending、已结束
//	I8  starting / configuring 期间 inflight == 0、binding 为空或正在配置、pending 不超过上限
//	I10 queue.Len() 等于队列中的有效请求数
func (p *PluginPool) checkInvariantsLocked() error {
	if len(p.procs) > p.cfg.MaxProcesses {
		return fmt.Errorf("I3: %d processes exceed max %d", len(p.procs), p.cfg.MaxProcesses)
	}

	seen := make(map[*request]string)
	valid := 0
	for _, w := range p.queue.buf[p.queue.head:] {
		if _, dead := p.queue.removed[w]; dead {
			if w.state != requestDone {
				return fmt.Errorf("I6: removed request %p is %s", w, w.state)
			}
			continue
		}
		valid++
		if w.state != requestQueued {
			return fmt.Errorf("I6: request %p in queue is %s", w, w.state)
		}
		if where, ok := seen[w]; ok {
			return fmt.Errorf("I6: request %p in queue twice (also %s)", w, where)
		}
		seen[w] = "queue"
	}
	if valid != p.queue.Len() {
		return fmt.Errorf("I10: queue.Len()=%d, valid=%d", p.queue.Len(), valid)
	}

	for id, proc := range p.procs {
		if err := p.checkProcessLocked(id, proc, seen); err != nil {
			return err
		}
	}

	if p.closed && p.queue.Len() != 0 {
		return fmt.Errorf("closed pool still has %d waiters", p.queue.Len())
	}
	return nil
}

func (p *PluginPool) checkProcessLocked(id uint64, proc *pluginProcess, seen map[*request]string) error {
	if proc.id != id {
		return fmt.Errorf("process key %d != id %d", id, proc.id)
	}
	if proc.inflight < 0 || proc.inflight > p.cfg.MaxLeasesPerProcess {
		return fmt.Errorf("I1: process %d inflight %d out of range", id, proc.inflight)
	}

	switch proc.state {
	case stateStarting, stateConfiguring:
		if proc.inflight != 0 {
			return fmt.Errorf("I8: preparing process %d has inflight %d", id, proc.inflight)
		}
		if proc.binding != nil && !proc.binding.configuring() {
			return fmt.Errorf("I8: preparing process %d has an effective binding", id)
		}
		if len(proc.pending) > p.cfg.MaxLeasesPerProcess {
			return fmt.Errorf("I8: process %d pending %d exceeds capacity", id, len(proc.pending))
		}
	case stateRunning:
		if len(proc.pending) != 0 {
			return fmt.Errorf("running process %d has pending requests", id)
		}
		if proc.binding.configuring() {
			return fmt.Errorf("running process %d binding is not effective", id)
		}
		if proc.exec == nil {
			return fmt.Errorf("running process %d has no executor", id)
		}
		if p.closed {
			return fmt.Errorf("closed pool still has running process %d", id)
		}
	case stateClosing:
		if len(proc.pending) != 0 {
			return fmt.Errorf("closing process %d has pending requests", id)
		}
	default:
		return fmt.Errorf("process %d has unknown state %s", id, proc.state)
	}

	for _, w := range proc.pending {
		if w.state == requestQueued {
			return fmt.Errorf("I6: pending request %p of process %d is queued", w, id)
		}
		if where, ok := seen[w]; ok {
			return fmt.Errorf("I6: request %p in process %d pending and %s", w, id, where)
		}
		seen[w] = fmt.Sprintf("process %d pending", id)
	}
	return nil
}

func assertInvariants(t *testing.T, p *PluginPool) {
	t.Helper()

	p.poolLock.Lock()
	err := p.checkInvariantsLocked()
	p.poolLock.Unlock()
	if err != nil {
		t.Fatalf("invariant violated: %v", err)
	}
}

// poolStats 测试用快照。
type poolStats struct {
	states   map[procState]int
	inflight int
	waiters  int
}

func (p *PluginPool) stats() poolStats {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	s := poolStats{states: make(map[procState]int), waiters: p.queue.Len()}
	for _, proc := range p.procs {
		s.states[proc.state]++
		s.inflight += proc.inflight
	}
	return s
}

// settled 没有进程处于准备或关闭中。
func (s poolStats) settled() bool {
	return s.states[stateStarting] == 0 && s.states[stateConfiguring] == 0 && s.states[stateClosing] == 0
}

func (s poolStats) total() int {
	n := 0
	for _, c := range s.states {
		n += c
	}
	return n
}
