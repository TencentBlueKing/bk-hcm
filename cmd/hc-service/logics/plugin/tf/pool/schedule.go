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
	"cmp"
	"time"
)

// action 调度器为队首选出的动作。
type action uint8

const (
	// actionHold 没有可用资源，结束本轮。
	actionHold action = iota
	// actionDeliver 规则 1a：已按同样配置配好且有空位，直接交付。
	actionDeliver
	// actionReserve 规则 1b：正在按同样配置准备且预约未满，加入 pending。
	actionReserve
	// actionConfigure 规则 2 / 3：空闲进程（未配置、已过期，或配置了别的）配置成请求的配置。
	actionConfigure
	// actionCreate 规则 4：进程数未到上限，新建进程。
	actionCreate
)

// scheduleLocked 只看队首：为它选进程并派发，再看下一个；选不出就结束，不越过队首。
//
// 慢操作都交给协程，这里只做内存读写。整轮只取一次当前时间，避免配置在一轮中途过期导致判断不一致。
func (p *PluginPool) scheduleLocked() {
	if p.closed {
		return
	}

	now := p.now()
	for {
		w, ok := p.queue.Peek()
		if !ok {
			return
		}

		proc, act := p.pickLocked(now, w.key)
		if act == actionHold {
			return
		}

		p.queue.PopFront()
		w.state = requestPending

		switch act {
		case actionDeliver:
			p.deliverLocked(proc, w)
		case actionReserve:
			proc.pending = append(proc.pending, w)
		case actionConfigure:
			p.startConfigureLocked(proc, w)
		case actionCreate:
			p.startCreateLocked(w)
		}
	}
}

// pickLocked 为配置键 key 选进程，只遍历进程表，耗时与队列长度无关。
//
// 优先级：
//
//	1a. running、配置已生效且 key 一致、inflight 未满：直接交付；取 inflight 少的，再取配置更新的；
//	1b. starting / configuring、正在配置成同一个 key、pending 未满：预约；取 pending 少的；
//	2.  running 空闲、未配置或配置已过期：配置；取空闲最久的；
//	3.  running 空闲、配置了别的：重配；取空闲最久的；
//	4.  进程数未到上限：新建；
//	5.  都不满足：等待。
//
// 规则 2、3 要求没有在途 lease，否则这些 lease 后续的调用会用上别人的配置。平局一律取进程 ID 小的。
func (p *PluginPool) pickLocked(now time.Time, key string) (*pluginProcess, action) {
	var matched, reserved, unbound, other *pluginProcess

	for _, proc := range p.procs {
		switch proc.state {
		case stateRunning:
			switch {
			case !proc.binding.valid(now, p.cfg.MaxConfigAge):
				unbound = pickIdle(proc, unbound)
			case proc.binding.key == key:
				matched = p.pickMatched(proc, matched)
			default:
				other = pickIdle(proc, other)
			}

		case stateStarting, stateConfiguring:
			reserved = p.pickReserved(proc, key, reserved)
		}
	}

	switch {
	case matched != nil:
		return matched, actionDeliver
	case reserved != nil:
		return reserved, actionReserve
	case unbound != nil:
		return unbound, actionConfigure
	case other != nil:
		return other, actionConfigure
	case len(p.procs) < p.cfg.MaxProcesses:
		return nil, actionCreate
	default:
		return nil, actionHold
	}
}

// deliverLocked 把进程上的一个名额交给请求。已结束的请求不占名额。
func (p *PluginPool) deliverLocked(proc *pluginProcess, w *request) {
	lease := &Lease{pool: p, procID: proc.id, exec: proc.exec, acquiredAt: p.now()}
	if finishRequestLocked(w, result{lease: lease}) {
		proc.inflight++
	}
}

// pickMatched 若 proc 未满且比 cur 更适合直接交付（inflight 少的优先，再取配置更新的），返回 proc。
func (p *PluginPool) pickMatched(proc, cur *pluginProcess) *pluginProcess {
	if proc.inflight >= p.cfg.MaxLeasesPerProcess {
		return cur
	}
	if cur == nil || cmp.Or(cmp.Compare(proc.inflight, cur.inflight),
		cur.binding.boundAt.Compare(proc.binding.boundAt), cmp.Compare(proc.id, cur.id)) < 0 {
		return proc
	}
	return cur
}

// pickReserved 若 proc 正在配置成 key、预约未满且比 cur 的 pending 更少，返回 proc。
func (p *PluginPool) pickReserved(proc *pluginProcess, key string, cur *pluginProcess) *pluginProcess {
	if !proc.binding.configuring() || proc.binding.key != key || len(proc.pending) >= p.cfg.MaxLeasesPerProcess {
		return cur
	}
	if cur == nil || cmp.Or(cmp.Compare(len(proc.pending), len(cur.pending)), cmp.Compare(proc.id, cur.id)) < 0 {
		return proc
	}
	return cur
}

// pickIdle 若 proc 空闲且比 cur 空闲更久，返回 proc。
func pickIdle(proc, cur *pluginProcess) *pluginProcess {
	if proc.inflight == 0 && (cur == nil || compareIdle(proc, cur) < 0) {
		return proc
	}
	return cur
}

// compareIdle 空闲最久的排在前面，平局按进程 ID。
func compareIdle(a, b *pluginProcess) int {
	return cmp.Or(a.idleSince.Compare(b.idleSince), cmp.Compare(a.id, b.id))
}
