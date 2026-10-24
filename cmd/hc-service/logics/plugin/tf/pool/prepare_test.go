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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/kit"
)

func TestStartFailureRetriesUntilExhaustedAndFreesCapacity(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) { s.startFn = func(*fakeManager) error { return errors.New("handshake failed") } })
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxRetries: 2})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start plugin tcloud process failed")
	assert.NotContains(t, err.Error(), "key-of-A")
	// 初始 1 次 + 重试 2 次 = 3 次启动尝试。
	assert.EqualValues(t, 3, s.starts.Load())

	waitSettled(t, p)
	assert.Equal(t, 0, p.testStats().total(), "failed process is removed")
	ms := s.allManagers()
	require.Len(t, ms, 3)
	for _, m := range ms {
		assert.EqualValues(t, 1, m.closeCalls.Load(), "partially started process is closed")
	}

	s.set(func(s *fakeStarter) { s.startFn = nil })
	l := mustAcquire(t, p, "A")
	l.Release()
}

func TestStartTimeoutClosesProcess(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.startFn = func(m *fakeManager) error {
			<-m.closed()
			return errExecutorClosed
		}
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxRetries: 1, StartTimeout: 60 * time.Millisecond})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	waitSettled(t, p)
	assert.Equal(t, 0, p.testStats().total())
}

// TestConfigureBusinessErrorFailsBatchAndKeepsProcessIdle 业务错误：整批 pending 直接失败不重试，进程回到 running 空闲。
func TestConfigureBusinessErrorFailsBatchAndKeepsProcessIdle(t *testing.T) {
	gate := make(chan struct{})
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error {
			<-gate
			return &tfruntime.DiagnosticsError{
				Diagnostics: []tfruntime.Diagnostic{
					{Severity: tfruntime.SeverityError, Summary: "missing secret"},
				},
			}
		}
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 3, MaxRetries: 2})

	ch1 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.configures.Load() == 1 }, waitTimeout, time.Millisecond)
	ch2 := acquireAsync(p, kit.New(), "A")
	ch3 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool {
		p.poolLock.Lock()
		defer p.poolLock.Unlock()
		return len(p.procs[1].pending) == 3
	}, waitTimeout, time.Millisecond)

	close(gate)
	for _, ch := range []<-chan acquireResult{ch1, ch2, ch3} {
		r := waitResult(t, ch)
		require.Error(t, r.err)
		var de *tfruntime.DiagnosticsError
		assert.True(t, errors.As(r.err, &de), "business error keeps its type through the pool")
		assert.NotContains(t, r.err.Error(), "key-of-A")
	}

	waitSettled(t, p)
	assert.EqualValues(t, 1, s.configures.Load(), "business error must not retry")
	st := p.testStats()
	assert.Equal(t, 1, st.states[stateRunning], "process returns to running idle on business error")

	// 随后好配置可直接复用这台进程。
	s.set(func(s *fakeStarter) { s.configureFn = nil })
	l := mustAcquire(t, p, "B")
	assert.Equal(t, "1/B", identityOf(t, l))
	assert.EqualValues(t, 1, s.starts.Load())
	l.Release()
}

// TestConfigureTransportErrorRetriesAtQueueTail 传输错误：进程被关闭，等它的请求放回队尾，排在已在排队的请求后面。
func TestConfigureTransportErrorRetriesAtQueueTail(t *testing.T) {
	gate := make(chan struct{})
	var attempt atomic.Int64
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error {
			if attempt.Add(1) == 1 {
				<-gate
				return errors.New("broken pipe")
			}
			return nil
		}
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 2, MaxRetries: 2})

	chA1 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.configures.Load() == 1 }, waitTimeout, time.Millisecond)
	chA2 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool {
		p.poolLock.Lock()
		defer p.poolLock.Unlock()
		return p.procs[1] != nil && len(p.procs[1].pending) == 2
	}, waitTimeout, time.Millisecond)
	// B 在 A1、A2 之后入队；A1、A2 失败后放回队尾，应排在 B 后面。
	chB := acquireAsync(p, kit.New(), "B")
	waitWaiters(t, p, 1)

	close(gate)
	rb := waitResult(t, chB)
	require.NoError(t, rb.err)
	assert.Equal(t, "2/B", identityOf(t, rb.lease), "B is served first, on the replacement process")
	assertPending(t, chA1)
	assertPending(t, chA2)

	rb.lease.Release()
	ra1 := waitResult(t, chA1)
	ra2 := waitResult(t, chA2)
	require.NoError(t, ra1.err)
	require.NoError(t, ra2.err)
	assert.Equal(t, "2/A", identityOf(t, ra1.lease))
	ra1.lease.Release()
	ra2.lease.Release()
}

// TestRequeuePending 直接验证放回队尾的规则：保持原相对顺序、跳过已取消的、超过重试次数的以原错误失败。
// 顺序只能在放回的那一刻观察到，经过调度后同 key 的请求会被一起交付，所以这里不走完整流程。
func TestRequeuePending(t *testing.T) {
	p := newTestPool(t, &fakeStarter{}, PoolConfig{MaxProcesses: 1, MaxRetries: 1})

	newRequest := func(rid string, attempts int, state requestState) *request {
		kt := kit.New()
		kt.Rid = rid
		return &request{kt: kt, attempts: attempts, state: state, ch: make(chan result, 1)}
	}
	first := newRequest("first", 0, requestPending)
	second := newRequest("second", 0, requestPending)
	canceled := newRequest("canceled", 0, requestDone)
	exhausted := newRequest("exhausted", 1, requestPending)
	proc := &pluginProcess{id: 1, pending: []*request{first, canceled, exhausted, second}}

	cause := errors.New("start failed")
	p.poolLock.Lock()
	p.requeuePendingLocked(proc, cause)
	queued := p.queue.Drain()
	p.poolLock.Unlock()

	assert.Empty(t, proc.pending)
	require.Len(t, queued, 2)
	assert.Same(t, first, queued[0], "original relative order is kept")
	assert.Same(t, second, queued[1])
	for _, w := range queued {
		assert.Equal(t, 1, w.attempts)
		assert.Equal(t, requestQueued, w.state)
	}

	require.Len(t, exhausted.ch, 1)
	assert.ErrorIs(t, (<-exhausted.ch).err, cause, "exhausted request fails with the cause")
	assert.Equal(t, requestDone, exhausted.state)
	assert.Empty(t, canceled.ch, "canceled request is skipped, not failed or requeued")
}

func TestConfigureTimeoutClosesProcess(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		// 不响应 ctx，只有进程被关闭才返回。
		s.configureFn = func(m *fakeManager, _ map[string]any) error {
			<-m.closed()
			return errExecutorClosed
		}
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxRetries: 1, ConfigureTimeout: 60 * time.Millisecond})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	waitSettled(t, p)
	assert.Equal(t, 0, p.testStats().total())

	s.set(func(s *fakeStarter) { s.configureFn = nil })
	l := mustAcquire(t, p, "A")
	l.Release()
}

func TestPanicInPluginCallIsContained(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error { panic("boom") }
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxRetries: 1})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic")

	waitSettled(t, p)
	s.set(func(s *fakeStarter) { s.configureFn = nil })
	l := mustAcquire(t, p, "A")
	l.Release()
}

func TestCancelDuringStartKeepsProcessUnconfigured(t *testing.T) {
	gate := make(chan struct{})
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) { s.startFn = func(*fakeManager) error { <-gate; return nil } })
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	ctx, cancel := context.WithCancel(context.Background())
	kt := kit.New()
	kt.Ctx = ctx
	ch := acquireAsync(p, kt, "A")
	require.Eventually(t, func() bool { return s.starts.Load() == 1 }, waitTimeout, time.Millisecond)

	cancel()
	assert.ErrorIs(t, waitResult(t, ch).err, context.Canceled)

	close(gate)
	waitSettled(t, p)
	st := p.testStats()
	assert.Equal(t, 1, st.states[stateRunning], "process is kept after the request is gone")
	assert.EqualValues(t, 0, s.configures.Load(), "no waiter, no configure")

	l := mustAcquire(t, p, "A")
	assert.Equal(t, "1/A", identityOf(t, l))
	assert.EqualValues(t, 1, s.starts.Load())
	l.Release()
}

func TestReserveDuringPreparationSharesProcess(t *testing.T) {
	gate := make(chan struct{})
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error { <-gate; return nil }
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2, MaxLeasesPerProcess: 2})

	ch1 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.configures.Load() == 1 }, waitTimeout, time.Millisecond)

	ch2 := acquireAsync(p, kit.New(), "A")
	waitWaiters(t, p, 0)
	assert.EqualValues(t, 1, s.starts.Load(), "second same-key request reserves the preparing process")

	close(gate)
	r1 := waitResult(t, ch1)
	r2 := waitResult(t, ch2)
	require.NoError(t, r1.err)
	require.NoError(t, r2.err)
	assert.Equal(t, "1/A", identityOf(t, r1.lease))
	assert.Equal(t, "1/A", identityOf(t, r2.lease))
	r1.lease.Release()
	r2.lease.Release()
}

// TestReserveCanceledWaiterSkippedAndFullFallsThrough 预约中取消的请求在交付时被跳过；预约满时走后续规则新建第二个进程。
func TestReserveCanceledWaiterSkippedAndFullFallsThrough(t *testing.T) {
	gate := make(chan struct{})
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error { <-gate; return nil }
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2, MaxLeasesPerProcess: 2})

	ch1 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.configures.Load() == 1 }, waitTimeout, time.Millisecond)

	ctx2, cancel2 := context.WithCancel(context.Background())
	kt2 := kit.New()
	kt2.Ctx = ctx2
	ch2 := acquireAsync(p, kt2, "A")
	require.Eventually(t, func() bool {
		p.poolLock.Lock()
		defer p.poolLock.Unlock()
		return p.procs[1] != nil && len(p.procs[1].pending) == 2
	}, waitTimeout, time.Millisecond)

	// 第三个同 key 请求：进程 1 的 pending 已满（2 个），按规则 4 新建进程 2。
	ch3 := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.starts.Load() == 2 }, waitTimeout, time.Millisecond)

	// 取消在进程 1 pending 里的第二个请求。
	cancel2()
	assert.ErrorIs(t, waitResult(t, ch2).err, context.Canceled)

	close(gate)
	r1 := waitResult(t, ch1)
	r3 := waitResult(t, ch3)
	require.NoError(t, r1.err)
	require.NoError(t, r3.err)
	assert.Equal(t, "1/A", identityOf(t, r1.lease))
	assert.Equal(t, "2/A", identityOf(t, r3.lease))
	assert.Equal(t, 2, p.testStats().inflight, "canceled reserved request must not occupy inflight")
	r1.lease.Release()
	r3.lease.Release()
	waitSettled(t, p)
}
