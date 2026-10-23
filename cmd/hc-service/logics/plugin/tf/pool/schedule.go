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

	"hcm/pkg/logs"
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
	// actionCreate 规则 4：进程数未到上限，新建进程再配置。
	actionCreate
)

// scheduleLocked 取一个、派一个：只看队首，为它选进程并派发，选不出即结束本轮，不越过队首。
//
// 一轮可以连续派发多个请求；慢操作都交给异步任务，本函数只做内存读写。
// 整轮只取一次当前时间，避免配置在一轮中途过期导致前后判断不一致。
func (p *PluginPool) scheduleLocked() {
	if p.closed {
		return
	}

	now := p.now()
	for {
		w, ok := p.queue.Peek()
		//没有可处理的请求，结束本轮调度
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
		default:
			logs.Errorf("plugin %s pool dispatched unknown action %d, rid: %s", p.pluginName, act, w.kt.Rid)
		}
	}
}

// pickLocked 为配置键 key 选进程，单次遍历进程表，耗时与队列长度无关。
//
// 优先级：
//
//	1a. running、配置已生效未过期且 key 一致、inflight 未满：直接交付；多个候选取 inflight 少的，再取 boundAt 更新的；
//	1b. starting / configuring、正在配置成同一个 key、pending 未满：加入 pending 预约；多个候选取 pending 少的；
//	2.  running、inflight 为 0、未配置或配置已过期：配置；
//	3.  running、inflight 为 0、配置了别的且未过期：重配；
//	4.  以上都没有且进程数未到上限：新建；
//	5.  都不满足：不派发。
//
// 规则 2、3 要求 inflight 为 0：有在途 lease 时重配，会让这些 lease 后续的调用用上别人的配置。
// 规则 2、3 的同类候选取空闲最久的；所有平局按进程 ID 取小，保证结果确定。
func (p *PluginPool) pickLocked(now time.Time, key string) (*pluginProcess, action) {
	var matched, reserved, unbound, other *pluginProcess

	for _, proc := range p.procs {
		switch proc.state {
		case stateRunning:
			if proc.binding.valid(now, p.cfg.MaxConfigAge) {
				switch {
				case proc.binding.key == key:
					if proc.inflight < p.cfg.MaxLeasesPerProcess && lessLoaded(proc, matched) {
						matched = proc
					}
				case proc.inflight == 0:
					if lessIdle(proc, other) {
						other = proc
					}
				}
				continue
			}

			if proc.inflight == 0 && lessIdle(proc, unbound) {
				unbound = proc
			}
		case stateStarting, stateConfiguring:
			if proc.binding.configuring() && proc.binding.key == key &&
				len(proc.pending) < p.cfg.MaxLeasesPerProcess && lessPending(proc, reserved) {
				reserved = proc
			}
		}
	}

	switch {
	//规则 1a：已按同样配置配好且有空位，直接交付。
	case matched != nil:
		return matched, actionDeliver
	//规则 1b：正在按同样配置准备且预约未满，加入 pending。
	case reserved != nil:
		return reserved, actionReserve
	//规则 2：空闲进程（未配置、已过期，或配置了别的）配置成请求的配置。
	case unbound != nil:
		return unbound, actionConfigure
	//规则 3：running、inflight 为 0、配置了别的且未过期：重配。
	case other != nil:
		return other, actionConfigure
	//规则 4：进程数未到上限，新建进程再配置。
	case len(p.procs) < p.cfg.MaxProcesses:
		return nil, actionCreate
	default:
		return nil, actionHold
	}
}

// deliverLocked 把进程上的一个名额交给请求。已结束的请求不占名额。
func (p *PluginPool) deliverLocked(proc *pluginProcess, w *request) {
	if w.state == requestDone {
		return
	}

	proc.inflight++
	lease := &Lease{pool: p, procID: proc.id, exec: proc.exec, acquiredAt: p.now()}
	finishRequestLocked(w, result{lease: lease})
}

// lessLoaded a 是否比当前候选 b 更适合直接交付：inflight 更少优先，inflight 相同取 boundAt 更新的，平局取 ID 小的。
// b 为 nil 时为真。
func lessLoaded(a, b *pluginProcess) bool {
	if b == nil {
		return true
	}
	if a.inflight != b.inflight {
		return a.inflight < b.inflight
	}
	if !a.binding.boundAt.Equal(b.binding.boundAt) {
		return a.binding.boundAt.After(b.binding.boundAt)
	}
	return a.id < b.id
}

// lessPending a 是否比当前候选 b 更适合加入预约：pending 更少，平局取 ID 小的。b 为 nil 时为真。
func lessPending(a, b *pluginProcess) bool {
	if b == nil {
		return true
	}
	if len(a.pending) != len(b.pending) {
		return len(a.pending) < len(b.pending)
	}
	return a.id < b.id
}

// lessIdle a 是否比当前候选 b 更适合被配置：空闲更久，平局取 ID 小的。b 为 nil 时为真。
func lessIdle(a, b *pluginProcess) bool {
	if b == nil {
		return true
	}
	if !a.idleSince.Equal(b.idleSince) {
		return a.idleSince.Before(b.idleSince)
	}
	return a.id < b.id
}
