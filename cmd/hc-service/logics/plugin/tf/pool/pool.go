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
//
// 并发：池内只有一把锁 poolLock，锁内只做内存读写；创建、配置、关闭进程等慢操作在锁外的协程中执行，
// 完成后回到锁内写回结果。以 Locked 结尾的函数要求调用方已持有 poolLock。
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
	"hcm/pkg/metrics"
)

// Options 创建进程池的参数。
type Options struct {
	PluginName    string
	PluginVersion string
	// ExecPath 插件可执行文件路径。
	ExecPath string
	// Starter 进程管理工厂，每创建一个进程调用一次。
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
	// now 时钟
	now func() time.Time

	// taskCtx 创建、配置、关闭协程的根 context，不跟随请求取消，只在 Close 时取消。
	taskCtx     context.Context
	cancelTasks context.CancelFunc
	// stopCheck 关闭后巡检协程退出。
	stopCheck chan struct{}
	// wg 统计所有后台协程，Close 据此等待。
	wg sync.WaitGroup

	poolLock sync.Mutex
	procs    map[uint64]*pluginProcess
	queue    *queue[*request]
	closed   bool
	// procSeq 进程 ID 高水位，只增不复用。
	procSeq uint64
}

// NewPluginPool 创建进程池并启动巡检。进程按需创建，空闲进程由巡检补足。
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
	p := &PluginPool{
		pluginName:    opts.PluginName,
		pluginVersion: opts.PluginVersion,
		execPath:      opts.ExecPath,
		starter:       opts.Starter,
		cfg:           cfg,
		now:           time.Now,
		taskCtx:       taskCtx,
		cancelTasks:   cancel,
		stopCheck:     make(chan struct{}),
		procs:         make(map[uint64]*pluginProcess),
		queue:         newQueue[*request](),
	}

	if err := metrics.RegisterPluginPoolStats(p.pluginName, p.stats); err != nil {
		// 指标注册失败（如同名插件已经注册过）不影响池工作，只是缺少状态类指标。
		logs.Errorf("register plugin %s pool stats metrics failed, err: %v", p.pluginName, err)
	}

	p.wg.Add(1)
	go p.checkLoop()
	return p, nil
}

// Acquire 申请一个已按 config 配好的进程名额。
//
// 请求进入队尾，按入队顺序被分配。以下情况返回错误：池已关闭（ErrPoolClosed）、队列已满（ErrTooManyWaiters）、
// 排队超过 MaxQueueWait（ErrAcquireTimeout）、kt.Ctx 结束（ctx 的错误）、为请求准备进程失败。
// config 含凭证，返回的错误不包含 config 内容；传入后调用方不得再修改 config。
func (p *PluginPool) Acquire(kt *kit.Kit, config map[string]any) (*Lease, error) {
	if kt == nil {
		return nil, errors.New("kit should not be nil")
	}

	start := p.now()
	lease, result, err := p.acquire(kt, config)
	metrics.ObservePluginPoolAcquire(p.pluginName, result, p.now().Sub(start))
	return lease, err
}

// acquire 执行一次申请，额外返回 acquire_seconds 的 result。
func (p *PluginPool) acquire(kt *kit.Kit, config map[string]any) (*Lease, string, error) {
	key, err := configKey(config)
	if err != nil {
		return nil, acquireInvalidConfig, fmt.Errorf("invalid provider config: %w", err)
	}

	w, err := p.enqueue(kt, config, key)
	switch {
	case errors.Is(err, ErrPoolClosed):
		return nil, acquireClosed, err
	case err != nil:
		return nil, acquireQueueFull, err
	}
	return p.wait(kt, w)
}

// enqueue 把请求放入队尾并立刻调度。
func (p *PluginPool) enqueue(kt *kit.Kit, config map[string]any, key string) (*request, error) {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}
	if p.queue.Len() >= p.cfg.MaxQueueLength {
		return nil, ErrTooManyWaiters
	}

	w := &request{
		config:     config,
		key:        key,
		kt:         kt,
		enqueuedAt: p.now(),
		state:      requestQueued,
		ch:         make(chan result, 1),
	}
	p.queue.PushBack(w)
	p.scheduleLocked()
	return w, nil
}

// wait 等待请求的结果，额外返回 acquire_seconds 的 result。必须在锁外等：调度器要拿同一把锁写入结果。
func (p *PluginPool) wait(kt *kit.Kit, w *request) (*Lease, string, error) {
	ctx := kt.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	timer := time.NewTimer(p.cfg.MaxQueueWait)
	defer timer.Stop()

	select {
	case r := <-w.ch:
		if r.err != nil {
			return nil, acquireFailure(r.err), r.err
		}
		return r.lease, acquireSuccess, nil
	case <-ctx.Done():
		p.cancel(w)
		return nil, acquireCanceled, ctx.Err()
	case <-timer.C:
		p.cancel(w)
		return nil, acquireTimeout, ErrAcquireTimeout
	}
}

// cancel 调用方放弃等待。
//
// 请求在队列中则移出队列；在进程 pending 中只标记结束，交付时跳过；已交付但调用方没取走则立即归还。
// 交付与取消都在锁内检查 requestDone，二者只有一个生效，所以名额既不漏还也不重复还。
func (p *PluginPool) cancel(w *request) {
	var orphan *Lease

	p.poolLock.Lock()
	if w.state != requestDone {
		if w.state == requestQueued {
			p.queue.Remove(w)
		}
		w.state = requestDone
		// 被取消的可能是队首，后面的请求也许已能派发。
		p.scheduleLocked()
	} else if len(w.ch) > 0 {
		orphan = (<-w.ch).lease
	}
	p.poolLock.Unlock()

	// Release 会再取锁，必须在解锁后调用。
	if orphan != nil {
		orphan.Release()
	}
}

// release 归还 lease 占用的名额，由 Lease.Release 调用。进程已不在表中时忽略。
func (p *PluginPool) release(l *Lease) {
	metrics.ObservePluginPoolLeaseHold(p.pluginName, p.now().Sub(l.acquiredAt))

	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	proc := p.procs[l.procID]
	if proc == nil {
		return
	}
	if proc.inflight <= 0 {
		logs.Errorf("plugin %s process %d released with no inflight lease, state: %s", p.pluginName, proc.id,
			proc.state)
		return
	}

	proc.inflight--
	if proc.idle() {
		proc.idleSince = p.now()
	}
	p.scheduleLocked()
}

// Close 停止接收新请求，排队中与准备中的请求返回 ErrPoolClosed，并关闭全部进程。
//
// 有 lease 的进程由进程管理的 Close 等在途RPC调用返回后结束，处在两次调用之间的 lease 后续调用会失败。
// 正在创建或配置的进程等协程写回后关闭。ctx 结束时取消仍在进行的创建、配置并返回 ctx 的错误。
// 重复调用直接返回 nil。
func (p *PluginPool) Close(ctx context.Context) error {
	if !p.shutdown(ctx) {
		return nil
	}
	logs.Infof("plugin %s process pool is closing", p.pluginName)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.cancelTasks()
		return nil
	case <-ctx.Done():
		p.cancelTasks()
		return ctx.Err()
	}
}

// shutdown 标记关闭并开始关闭各进程。已经关闭过时返回 false。
func (p *PluginPool) shutdown(ctx context.Context) bool {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if p.closed {
		return false
	}
	p.closed = true
	close(p.stopCheck)

	for _, w := range p.queue.Drain() {
		finishRequestLocked(w, result{err: ErrPoolClosed})
	}
	for _, proc := range p.procs {
		switch proc.state {
		case stateRunning:
			p.closeProcessLocked(ctx, proc)
		case stateStarting, stateConfiguring:
			// 协程写回时看到池已关闭会关掉进程；等它的请求现在就结束。
			p.failPendingLocked(proc, ErrPoolClosed)
		}
	}
	return true
}
