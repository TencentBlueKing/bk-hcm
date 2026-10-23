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

// Package pool 实现 tf_provider 插件的进程池。
//
// 调用方用 Acquire 申请 lease，拿到的 lease 对应一个已按调用方 Provider 配置配好的插件进程上的一个并发名额，
// 用完 Release。池只交付 lease，不执行业务请求。
//
// 调度：所有请求共用一条严格 FIFO 队列，调度器每次只为队首选进程，选不出即结束本轮，不越过队首。
// 因此服务顺序只由入队顺序决定，不存在饥饿。
//
// 并发模型：池内只有一把锁 mu，锁内只做内存读写；创建、配置、关闭进程等慢操作在锁外的异步任务中执行，
// 完成后回到锁内写回。约定：以 Locked 结尾的函数要求调用方已持有 poolLock。
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Options 创建进程池的参数。
type Options struct {
	PluginName    string
	PluginVersion string
	// ExecPath 插件可执行文件路径。
	ExecPath string
	// Starter 进程管理工厂，每次创建进程调用一次。
	Starter tfruntime.Starter
	Config  PoolConfig
}

// PluginPool 一个插件的进程池。
type PluginPool struct {
	pluginName    string
	pluginVersion string
	execPath      string
	starter       tfruntime.Starter
	cfg           PoolConfig
	// now 可替换的时钟，只在构造后、使用前替换（用于测试配置过期）。
	now func() time.Time

	// taskCtx 异步任务（创建、配置、内部关闭）的根 context：不跟随请求取消，
	// 只在 Close 超时后取消，用于中断仍在进行的任务。
	taskCtx     context.Context
	cancelTasks context.CancelFunc
	// wg 统计全部异步任务，Close 据此等待。
	wg sync.WaitGroup

	poolLock sync.Mutex
	procs    map[uint64]*pluginProcess
	queue    *queue[*request]
	closed   bool
	// procSeq 进程 ID 高水位，只增不复用，避免已发出的 lease 归还到新进程。
	procSeq uint64
}

// NewPluginPool 创建进程池。池创建后不预热进程，第一个请求到达时按需创建。
func NewPluginPool(opts Options) (*PluginPool, error) {
	if opts.PluginName == "" {
		return nil, errors.New("plugin name should not be empty")
	}
	if opts.ExecPath == "" {
		return nil, errors.New("plugin exec path should not be empty")
	}
	if opts.Starter == nil {
		return nil, errors.New("plugin process starter should not be nil")
	}

	cfg := opts.Config
	cfg.trySetDefault()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pool config: %w", err)
	}

	taskCtx, cancel := context.WithCancel(context.Background())
	return &PluginPool{
		pluginName:    opts.PluginName,
		pluginVersion: opts.PluginVersion,
		execPath:      opts.ExecPath,
		starter:       opts.Starter,
		cfg:           cfg,
		now:           time.Now,
		taskCtx:       taskCtx,
		cancelTasks:   cancel,
		procs:         make(map[uint64]*pluginProcess),
		queue:         newQueue[*request](),
	}, nil
}

// Acquire 申请一个已按 config 配好的进程名额。
//
// 请求进入队尾，按入队顺序被分配。以下情况返回错误：池已关闭（ErrPoolClosed）、队列已满（ErrTooManyWaiters）、
// 排队超过 MaxQueueWait（ErrAcquireTimeout）、kt.Ctx 结束（返回 ctx 的错误）、为该请求准备进程失败。
// config 含凭证，返回的错误不包含 config 内容。
func (p *PluginPool) Acquire(kt *kit.Kit, config map[string]any) (*Lease, error) {
	if kt == nil {
		return nil, errors.New("kit should not be nil")
	}

	// 匹配键只依赖 config，在锁外算完。编码失败时请求还没有入队。
	key, err := configKey(config)
	if err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}

	w, err := p.enqueue(kt, config, key)
	if err != nil {
		return nil, err
	}
	// 入队后必须放开锁再等。调度器要拿同一把锁把结果写入 w.ch，持锁等待会死锁。
	return p.waitLease(kt, w)
}

// enqueue 把请求放入队尾并立刻调度。池已关闭或队列已满时不入队。
func (p *PluginPool) enqueue(kt *kit.Kit, config map[string]any, key string) (*request, error) {
	w := &request{config: config, key: key, kt: kt, ch: make(chan result, 1)}

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}
	if p.queue.Len() >= p.cfg.MaxQueueLength {
		return nil, ErrTooManyWaiters
	}

	w.enqueuedAt = p.now()
	w.state = requestQueued
	p.queue.PushBack(w)
	//触发调度
	p.scheduleLocked()
	return w, nil
}

// waitLease 等待这次入队的结果。kt.Ctx 为空时不受调用方取消，只受 MaxQueueWait 约束。
//
// ctx 结束或排队超时都要取消请求：此时 lease 可能已经写入 w.ch，不取消就会泄漏名额。
func (p *PluginPool) waitLease(kt *kit.Kit, w *request) (*Lease, error) {
	ctx := kt.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	timer := time.NewTimer(p.cfg.MaxQueueWait)
	defer timer.Stop()

	select {
	case r := <-w.ch:
		return r.lease, r.err
	case <-ctx.Done():
		p.cancel(w)
		return nil, ctx.Err()
	case <-timer.C:
		p.cancel(w)
		return nil, ErrAcquireTimeout
	}
}

// cancel 调用方放弃等待。请求在队列中则出队；在进程 pending 中只改为 requestDone，交付时跳过；
// 已交付但调用方没取走则立即归还。交付与取消都在锁内以 requestDone 互斥，因此既不漏还也不重复还。
func (p *PluginPool) cancel(w *request) {
	var orphan *Lease

	p.poolLock.Lock()
	switch {
	case w.state != requestDone:
		if w.state == requestQueued {
			p.queue.Remove(w)
		}
		w.state = requestDone
		// 被取消的可能是队首，后面的请求也许已经可以派发。
		p.scheduleLocked()
	case len(w.ch) > 0:
		orphan = (<-w.ch).lease
	}
	p.poolLock.Unlock()

	// Release 会重新取锁，必须在解锁之后调用。
	if orphan != nil {
		orphan.Release()
	}
}

// release 归还进程上的一个名额，由 Lease.Release 调用。
func (p *PluginPool) release(procID uint64) {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	proc := p.procs[procID]
	if proc == nil {
		// 进程已退出，记录已删除。
		return
	}
	if proc.inflight <= 0 {
		logs.Errorf("plugin %s process %d released with no inflight lease, state: %s", p.pluginName, proc.id,
			proc.state)
		return
	}

	proc.inflight--
	if proc.inflight == 0 && proc.state == stateRunning {
		proc.idleSince = p.now()
	}
	p.scheduleLocked()
}

// Close 停止接收新请求，排队中与准备中的请求返回 ErrPoolClosed，并关闭全部进程。
//
// 有 lease 的进程由进程管理的 Close 等待在途调用返回后结束；处在两次调用之间的 lease，其后续调用会返回错误。
// 正在创建或配置的进程等任务写回后关闭。ctx 结束时中断仍在进行的任务并返回 ctx 的错误，
// 进程管理的 Close 在 ctx 结束时会强制结束进程。再次调用直接返回 nil，不再等待。
func (p *PluginPool) Close(ctx context.Context) error {
	if !p.shutdown(ctx) {
		return nil
	}

	logs.Infof("plugin %s process pool is closing", p.pluginName)
	return p.waitClosed(ctx)
}

// shutdown 停止接收新请求，并按进程状态开始关闭。池已经在关闭时返回 false。
func (p *PluginPool) shutdown(ctx context.Context) bool {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.closed {
		return false
	}
	p.closed = true

	for _, w := range p.queue.Drain() {
		finishRequestLocked(w, result{err: ErrPoolClosed})
	}
	for _, proc := range p.procs {
		switch proc.state {
		case stateRunning:
			p.closeProcessLocked(proc, ctx, 0)
		case stateStarting, stateConfiguring:
			// 创建和配置还在执行，写回时看到池已关闭再关进程。等它们的请求现在结束。
			p.failPendingLocked(proc, ErrPoolClosed)
		}
	}
	return true
}

// waitClosed 等待全部关闭协程结束。wg.Wait 不能放进 select，所以用 done 通知。
// ctx 结束时取消 taskCtx，让仍在进行的创建、配置和内部关闭返回。
func (p *PluginPool) waitClosed(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		p.cancelTasks()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		p.cancelTasks()
		return ctx.Err()
	}
}
