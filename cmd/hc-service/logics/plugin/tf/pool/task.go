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
	"hcm/cmd/hc-service/logics/tfruntime/codec"
	"hcm/cmd/hc-service/logics/tfruntime/executor"
	"hcm/cmd/hc-service/types"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
)

// 创建、配置、关闭进程都在锁外的协程中执行：
//   - 锁内先把进程改成 starting / configuring / closing，协程结束后回到锁内写回，每个协程只写回一次；
//   - 同一个进程同一时刻只有一个协程在处理它，由进程状态保证；
//   - 协程用 taskCtx 加超时，不跟随请求取消：进程是共享的，请求走了，后面的请求还能用。

// startCreateLocked 登记一个 starting 进程占用容量，在协程里拉起。
// w 为触发创建的请求，拉起后为它配置；为 nil 时是预热，拉起后作为未配置的空闲进程。
func (p *PluginPool) startCreateLocked(w *request) {
	p.procSeq++
	proc := &pluginProcess{id: p.procSeq, state: stateStarting}
	if w != nil {
		proc.binding = &binding{config: w.config, key: w.key}
		proc.pending = []*request{w}
	}
	p.procs[proc.id] = proc

	p.wg.Add(1)
	go p.runCreate(proc)
}

// startConfigureLocked 把空闲的 running 进程配置成请求的配置。
func (p *PluginPool) startConfigureLocked(proc *pluginProcess, w *request) {
	proc.binding = &binding{config: w.config, key: w.key}
	proc.pending = []*request{w}
	proc.state = stateConfiguring
	p.wg.Add(1)
	go p.runConfigure(proc, proc.exec, proc.binding.config)
}

func (p *PluginPool) runCreate(proc *pluginProcess) {
	defer p.wg.Done()

	kt := kit.New()
	kt.Ctx = context.WithValue(p.taskCtx, constant.RidKey, kt.Rid)

	start := p.now()
	manager, exec, err := p.startProcess(kt)
	result := resultSuccess
	if err != nil {
		result = resultFailure
	}
	metrics.ObservePluginPoolStart(p.pluginName, result, p.now().Sub(start))

	p.finishCreate(proc, manager, exec, err, kt)
}

// startProcess 拉起一个进程。握手由进程管理侧限制等待时间，池不再另设超时。
//
// 只要 manager 构造出来就返回它，即使 Start 失败：调用方要靠它关掉可能已拉起一半的进程。
func (p *PluginPool) startProcess(kt *kit.Kit) (*tfruntime.ProcessManager, executor.Executor, error) {
	manager := tfruntime.NewProcessManager(p.pluginName, p.execPath, nil, "", "")
	manager.SetValueCodec(codec.Codec{})

	ctx := kt.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	exec, err := manager.Start(ctx)
	if err == nil && exec == nil {
		err = errors.New("start returned nil executor")
	}
	return manager, exec, err
}

func (p *PluginPool) finishCreate(proc *pluginProcess, manager *tfruntime.ProcessManager, exec executor.Executor,
	err error, kt *kit.Kit) {

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	// starting 的进程只会由这里移出进程表，所以不用核对记录是否还在。
	proc.manager = manager
	proc.exec = exec

	switch {
	case err != nil:
		logs.Errorf("plugin %s start process %d failed, err: %v, rid: %s", p.pluginName, proc.id, err, kt.Rid)
		p.requeuePendingLocked(proc, fmt.Errorf("start plugin %s process failed: %w", p.pluginName, err))
		delete(p.procs, proc.id)

	case p.closed:
		p.failPendingLocked(proc, ErrPoolClosed)
		p.closeProcessLocked(proc)

	case !proc.hasLivePending():
		// 预热，或等它的请求都已取消：不配置，避免把凭证写进没人用的进程。
		proc.binding = nil
		proc.pending = nil
		proc.state = stateRunning
		proc.idleSince = p.now()

	default:
		proc.state = stateConfiguring
		p.wg.Add(1)
		go p.runConfigure(proc, proc.exec, proc.binding.config)
	}

	p.scheduleLocked()
}

func (p *PluginPool) runConfigure(proc *pluginProcess, exec executor.Executor, config map[string]any) {
	defer p.wg.Done()

	kt, cancel := p.taskKit(p.cfg.ConfigureTimeout)
	defer cancel()

	start := p.now()
	_, err := callWithTimeout(kt.Ctx, func() (*types.ConfigureResult, error) {
		return exec.ConfigureProvider(kt.Ctx, config)
	})
	metrics.ObservePluginPoolConfigure(p.pluginName, configureResult(err), p.now().Sub(start))

	p.finishConfigure(proc, err, kt)
}

func (p *PluginPool) finishConfigure(proc *pluginProcess, err error, kt *kit.Kit) {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.procs[proc.id] != proc {
		// 配置期间进程退出，巡检已删除记录并把请求放回队列。
		return
	}

	switch {
	case p.closed:
		p.failPendingLocked(proc, ErrPoolClosed)
		p.closeProcessLocked(proc)

	case err == nil:
		now := p.now()
		proc.binding.boundAt = now
		proc.state = stateRunning
		proc.idleSince = now
		for _, w := range proc.pending {
			p.deliverLocked(proc, w)
		}
		proc.pending = nil
		logs.Infof("plugin %s process %d configured, key: %.12s, inflight: %d, rid: %s", p.pluginName, proc.id,
			proc.binding.key, proc.inflight, kt.Rid)

	case isBusinessError(err):
		// 插件拒绝了这份配置（缺密钥、鉴权失败等）：换进程也会失败，请求直接失败，进程完好、继续复用。
		logs.Errorf("plugin %s rejected config on process %d, err: %v, rid: %s", p.pluginName, proc.id, err, kt.Rid)
		p.failPendingLocked(proc, fmt.Errorf("configure plugin %s failed: %w", p.pluginName, err))
		proc.binding = nil
		proc.state = stateRunning
		proc.idleSince = p.now()

	default:
		// 超时或传输错误：进程状态不可信，关掉它，请求放回队列重试。
		logs.Errorf("plugin %s configure process %d failed, err: %v, rid: %s", p.pluginName, proc.id, err, kt.Rid)
		p.requeuePendingLocked(proc, fmt.Errorf("configure plugin %s process failed: %w", p.pluginName, err))
		p.closeProcessLocked(proc)
	}

	p.scheduleLocked()
}

// failPendingLocked 让等这次准备的请求都以 err 失败，不重试。
func (p *PluginPool) failPendingLocked(proc *pluginProcess, err error) {
	for _, w := range proc.pending {
		finishRequestLocked(w, result{err: err})
	}
	proc.pending = nil
}

// requeuePendingLocked 把等这次准备的请求按原顺序放回队尾重试；超过 MaxRetries 的以 err 失败。
func (p *PluginPool) requeuePendingLocked(proc *pluginProcess, err error) {
	for _, w := range proc.pending {
		if w.state == requestDone {
			continue
		}

		w.attempts++
		metrics.IncPluginPoolRetry(p.pluginName)
		switch {
		case p.closed:
			finishRequestLocked(w, result{err: ErrPoolClosed})
		case w.attempts > p.cfg.MaxRetries:
			finishRequestLocked(w, result{err: err})
		default:
			w.state = requestQueued
			p.queue.PushBack(w)
		}
	}
	proc.pending = nil
}

// closeProcessLocked 把进程转为 closing，在协程里关闭；确认退出后才删除记录、释放容量。
func (p *PluginPool) closeProcessLocked(proc *pluginProcess) {
	proc.state = stateClosing
	proc.binding = nil

	if proc.manager == nil {
		// 进程管理对象都没创建出来，没有要关的进程。
		delete(p.procs, proc.id)
		return
	}

	p.wg.Add(1)
	go p.runClose(proc)
}

func (p *PluginPool) runClose(proc *pluginProcess) {
	defer p.wg.Done()

	err := safeCall(func() error { return proc.manager.Close() })
	exited := err == nil || processExited(proc.manager)

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if !exited {
		// 进程可能还活着：保留 closing 记录继续占用容量，由巡检确认退出后删除。
		logs.Errorf("plugin %s close process %d failed, err: %v", p.pluginName, proc.id, err)
		return
	}

	logs.Infof("plugin %s process %d closed", p.pluginName, proc.id)
	delete(p.procs, proc.id)
	p.scheduleLocked()
}

// taskKit 为后台协程创建 kit：新 rid，ctx 继承 taskCtx 并带超时。
func (p *PluginPool) taskKit(timeout time.Duration) (*kit.Kit, context.CancelFunc) {
	kt := kit.New()
	ctx, cancel := context.WithTimeout(p.taskCtx, timeout)
	kt.Ctx = context.WithValue(ctx, constant.RidKey, kt.Rid)
	return kt, cancel
}

// isBusinessError 是否为业务错误（插件返回 Error 级 Diagnostic 或配置无法按 Schema 编码）。
func isBusinessError(err error) bool {
	return errors.Is(err, executor.ErrDiagnostic) || errors.Is(err, executor.ErrInvalidConfig)
}

// callResult callWithTimeout 中 fn 的返回值。
type callResult[T any] struct {
	val T
	err error
}

// callWithTimeout 在协程中执行 fn，ctx 结束时不再等待、直接返回超时错误。
//
// 用于不一定响应 ctx 的插件调用：超时后 fn 可能还在跑，它的结果被丢弃；调用方须关闭对应进程让它返回。
func callWithTimeout[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	done := make(chan callResult[T], 1)
	go func() {
		var r callResult[T]
		r.err = safeCall(func() error {
			var err error
			r.val, err = fn()
			return err
		})
		done <- r
	}()

	select {
	case r := <-done:
		return r.val, r.err
	case <-ctx.Done():
		var zero T
		return zero, fmt.Errorf("call did not return in time: %w", ctx.Err())
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

// processExited 进程是否已退出；Exited 发生 panic 时按未退出处理。
func processExited(manager *tfruntime.ProcessManager) bool {
	exited := false
	err := safeCall(func() error {
		exited = manager.Exited()
		return nil
	})
	return err == nil && exited
}
