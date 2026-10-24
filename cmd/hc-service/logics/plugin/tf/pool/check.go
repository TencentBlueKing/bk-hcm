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
	"slices"
	"time"

	"hcm/pkg/logs"
)

// checkLoop 按 PoolCheckInterval 巡检，池关闭后退出。
func (p *PluginPool) checkLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.cfg.PoolCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCheck:
			return
		case <-ticker.C:
			p.checkOnce()
		}
	}
}

// checkOnce 一轮巡检：
//  1. 锁外检查进程是否退出、连接是否可用；
//  2. 删除已退出的进程，关闭连接不可用的 running 进程；
//  3. 补足或回收空闲进程；
//  4. 调度一次，兜底可能遗漏的唤醒。
func (p *PluginPool) checkOnce() {
	procs := p.processesWithManager()

	var exited, unhealthy []*pluginProcess
	for _, proc := range procs {
		if processExited(proc.manager) {
			exited = append(exited, proc)
			continue
		}
		// CheckHealth 不阻塞，对所有存活进程都调；结果只对写回时仍是 running 的进程生效。
		if err := safeCall(proc.manager.CheckHealth); err != nil {
			logs.Warnf("plugin %s process %d health check failed, err: %v", p.pluginName, proc.id, err)
			unhealthy = append(unhealthy, proc)
		}
	}

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.closed {
		return
	}

	for _, proc := range exited {
		if p.procs[proc.id] != proc {
			continue
		}
		logs.Warnf("plugin %s process %d exited in state %s", p.pluginName, proc.id, proc.state)
		delete(p.procs, proc.id)
		p.requeuePendingLocked(proc, fmt.Errorf("plugin %s process exited", p.pluginName))
	}
	for _, proc := range unhealthy {
		if p.procs[proc.id] == proc && proc.state == stateRunning {
			p.closeProcessLocked(p.taskCtx, proc)
		}
	}

	p.reconcileIdleLocked(p.now())
	p.scheduleLocked()
}

// processesWithManager 返回已有进程管理对象的进程。starting 的进程还没有 manager，由创建协程自己处理。
func (p *PluginPool) processesWithManager() []*pluginProcess {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.closed {
		return nil
	}

	procs := make([]*pluginProcess, 0, len(p.procs))
	for _, proc := range p.procs {
		if proc.manager != nil {
			procs = append(procs, proc)
		}
	}
	return procs
}

// reconcileIdleLocked 让空闲进程数回到 MinIdleProcesses：少了补建，多了回收超时的。
//
// 空闲进程指 running 且没有在途 lease 的进程，不论是否已配置；还没被请求认领的 starting 进程拉起后就是空闲进程，也计入。
// 回收只处理超出下限的部分，空闲最久的先回收，所以回收后空闲数不会低于下限。
func (p *PluginPool) reconcileIdleLocked(now time.Time) {
	idle := 0
	var expired []*pluginProcess
	for _, proc := range p.procs {
		switch {
		case proc.idle():
			idle++
			if now.Sub(proc.idleSince) >= p.cfg.IdleReclaimTimeout {
				expired = append(expired, proc)
			}
		case proc.state == stateStarting && !proc.hasLivePending():
			idle++
		}
	}

	if idle < p.cfg.MinIdleProcesses {
		n := min(p.cfg.MinIdleProcesses-idle, p.cfg.MaxProcesses-len(p.procs))
		for i := 0; i < n; i++ {
			p.startCreateLocked(nil)
		}
		return
	}

	slices.SortFunc(expired, compareIdle)
	n := min(idle-p.cfg.MinIdleProcesses, len(expired))
	for _, proc := range expired[:n] {
		p.closeProcessLocked(p.taskCtx, proc)
	}
}
