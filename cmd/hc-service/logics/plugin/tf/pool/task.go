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
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// 异步任务约定：
//   - 任务在锁内改好进程状态（starting / configuring / closing），在锁外执行慢操作，完成后回到锁内写回；
//   - 每个进程同一时刻至多一个异步任务（状态本身即独占标志），每个任务恰好写回一次；
//   - 任务使用池的 taskCtx 加超时，不跟随请求取消：进程是共享的，请求走了，后面的请求还能用；
//   - 外部调用（Start、ConfigureProvider、Close）的 panic 被转成错误，按失败写回；
//   - Start 与 ConfigureProvider 不响应 ctx 时，超时后仍按失败写回并关闭进程，由进程管理的 Close 让其返回。

// startCreateLocked 规则 4：登记一个 starting 进程占用容量，异步拉起后再为请求配置。
func (p *PluginPool) startCreateLocked(w *request) {
	p.procSeq++
	proc := &pluginProcess{
		id:      p.procSeq,
		state:   stateStarting,
		binding: &binding{config: w.config, key: w.key},
		pending: []*request{w},
	}
	p.procs[proc.id] = proc

	p.wg.Add(1)
	go p.runCreate(proc, w.kt.Rid)
}

// startConfigureLocked 规则 2 / 3：把空闲的 running 进程配置成请求的配置。
func (p *PluginPool) startConfigureLocked(proc *pluginProcess, w *request) {
	proc.state = stateConfiguring
	proc.binding = &binding{config: w.config, key: w.key}
	proc.pending = []*request{w}

	p.wg.Add(1)
	go p.runConfigure(proc, proc.exec, proc.binding.config, w.kt.Rid)
}

func (p *PluginPool) runCreate(proc *pluginProcess, rid string) {
	defer p.wg.Done()

	kt, cancel := p.taskKit(p.cfg.StartTimeout)
	defer cancel()

	logs.Infof("plugin %s creating process %d, task rid: %s, rid: %s", p.pluginName, proc.id, kt.Rid, rid)

	// NewProcessManager 只构造对象、不拉起进程，同步调用；拿到 manager 之后，Start 超时也能把进程关掉。
	var manager tfruntime.ProcessManager
	err := safeCall(func() error {
		m, err := p.starter.NewProcessManager(kt, p.pluginName, p.pluginVersion, p.execPath)
		if err != nil {
			return fmt.Errorf("new process manager: %w", err)
		}
		if m == nil {
			return errors.New("new process manager returned nil")
		}
		manager = m
		return nil
	})

	var exec tfruntime.Executor
	if err == nil {
		err = callWithTimeout(kt.Ctx, func() error {
			e, err := manager.Start(kt)
			if err != nil {
				return err
			}
			if e == nil {
				return errors.New("start returned nil executor")
			}
			exec = e
			return nil
		})
	}

	p.finishCreate(proc, manager, exec, err, rid)
}

func (p *PluginPool) finishCreate(proc *pluginProcess, manager tfruntime.ProcessManager, exec tfruntime.Executor,
	err error, rid string) {

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	proc.manager = manager
	if err != nil {
		logs.Errorf("plugin %s start process %d failed, err: %v, rid: %s", p.pluginName, proc.id, err, rid)
		p.failProcessLocked(proc, fmt.Errorf("start plugin %s process failed: %w", p.pluginName, err))
		p.scheduleLocked()
		return
	}
	proc.exec = exec

	if p.closed {
		p.failProcessLocked(proc, ErrPoolClosed)
		return
	}

	if proc.binding == nil || !proc.hasLivePending() {
		// 等它的请求都已取消：不配置，避免无谓地把凭证写进进程，作为未配置的空闲进程留给后面的队首。
		logs.Infof("plugin %s process %d started without live waiters, keep it idle, rid: %s", p.pluginName,
			proc.id, rid)
		proc.binding = nil
		proc.pending = nil
		proc.state = stateRunning
		proc.idleSince = p.now()
		p.scheduleLocked()
		return
	}

	proc.state = stateConfiguring
	p.wg.Add(1)
	go p.runConfigure(proc, proc.exec, proc.binding.config, rid)
}

func (p *PluginPool) runConfigure(proc *pluginProcess, exec tfruntime.Executor, config map[string]any, rid string) {
	defer p.wg.Done()

	kt, cancel := p.taskKit(p.cfg.ConfigureTimeout)
	defer cancel()

	err := callWithTimeout(kt.Ctx, func() error {
		_, err := exec.ConfigureProvider(kt, &tfruntime.ConfigureProviderRequest{Config: config})
		return err
	})

	p.finishConfigure(proc, err, kt.Rid, rid)
}

func (p *PluginPool) finishConfigure(proc *pluginProcess, err error, taskRid, rid string) {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	key := proc.binding.key
	if err != nil {
		logs.Errorf("plugin %s configure process %d failed, err: %v, task rid: %s, rid: %s", p.pluginName,
			proc.id, err, taskRid, rid)
		p.failProcessLocked(proc, fmt.Errorf("configure plugin %s process failed: %w", p.pluginName, err))
		p.scheduleLocked()
		return
	}

	if p.closed {
		p.failProcessLocked(proc, ErrPoolClosed)
		return
	}

	now := p.now()
	proc.binding.boundAt = now
	proc.state = stateRunning

	pending := proc.pending
	proc.pending = nil
	for _, w := range pending {
		if w.state == requestDone {
			continue
		}
		// if proc.inflight >= p.cfg.MaxLeasesPerProcess {
		// 	// pending 长度不超过 MaxLeasesPerProcess，兜底，正常情况不会执行到这；放回队列而不是丢弃。
		// 	logs.Errorf("plugin %s process %d has more pending requests than capacity, rid: %s", p.pluginName,
		// 		proc.id, w.kt.Rid)
		// 	p.queue.PushBack(w)
		// 	w.state = requestQueued
		// 	continue
		// }
		p.deliverLocked(proc, w)
	}
	if proc.inflight == 0 {
		proc.idleSince = now
	}

	logs.Infof("plugin %s process %d configured, key: %.12s, inflight: %d, task rid: %s, rid: %s", p.pluginName,
		proc.id, key, proc.inflight, taskRid, rid)
	p.scheduleLocked()
}

// failPendingLocked 让本次准备的全部等待请求以 cause 失败（已结束的跳过），并清空 pending。
func (p *PluginPool) failPendingLocked(proc *pluginProcess, cause error) {
	pending := proc.pending
	proc.pending = nil
	for _, w := range pending {
		finishRequestLocked(w, result{err: cause})
	}
}

// failProcessLocked 进程侧失败：等待请求失败，进程状态不可信，关闭它。
func (p *PluginPool) failProcessLocked(proc *pluginProcess, cause error) {
	p.failPendingLocked(proc, cause)
	p.closeProcessLocked(proc, p.taskCtx, p.cfg.CloseTimeout)
}

// closeProcessLocked 把进程标为 closing，并在锁外关闭。调用方已持有 poolLock。
//
// timeout 大于 0：在传入的 ctx（内部关闭用 taskCtx）上再加这个时限，到点强制结束进程。
// timeout 为 0：只用传入的 ctx（PluginPool.Close 的调用方 ctx），池不能取消它。
//
// 关闭完成后才从进程表删除。删除前进程仍占着容量，避免实际进程数超过上限。
func (p *PluginPool) closeProcessLocked(proc *pluginProcess, ctx context.Context, timeout time.Duration) {
	proc.state = stateClosing
	proc.binding = nil

	// timeout 为 0 时没有自己的计时器，cancel 是空函数，避免误取消调用方的 ctx。
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	if proc.manager == nil {
		// 进程还没拉起来，没有可关闭的对象。停掉计时器并立刻释放容量。
		cancel()
		delete(p.procs, proc.id)
		return
	}

	manager := proc.manager
	p.wg.Add(1)
	// manager.Close 会等到在途调用结束，ctx 结束才强制杀进程，不能占着 poolLock 等。
	go func() {
		defer p.wg.Done()
		defer cancel()

		if err := safeCall(func() error { return manager.Close(ctx) }); err != nil {
			logs.Errorf("plugin %s close process %d failed, err: %v", p.pluginName, proc.id, err)
		} else {
			logs.Infof("plugin %s process %d closed", p.pluginName, proc.id)
		}

		// Close 已返回，进程已退出或已被强制结束。这时才删记录并让出容量。
		p.poolLock.Lock()
		delete(p.procs, proc.id)
		p.scheduleLocked()
		p.poolLock.Unlock()
	}()
}

// taskKit 为异步任务创建 kit：新 rid，ctx 继承 taskCtx 并带超时。
func (p *PluginPool) taskKit(timeout time.Duration) (*kit.Kit, context.CancelFunc) {
	kt := kit.New()
	ctx, cancel := context.WithTimeout(p.taskCtx, timeout)
	kt.Ctx = context.WithValue(ctx, constant.RidKey, kt.Rid)
	return kt, cancel
}

// callWithTimeout 在独立协程中执行 fn；若 fn 未在 ctx 截止前返回，则直接返回超时错误，
// 随后由调用方通过 manager.Close 关闭子进程以唤醒阻塞中的 fn。
func callWithTimeout(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- safeCall(fn)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("call did not return in time: %w", ctx.Err())
	}
}

// safeCall 执行 fn，把 panic 转成错误。
func safeCall(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logs.Errorf("plugin process call panic: %v\n%s", r, debug.Stack())
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn()
}
