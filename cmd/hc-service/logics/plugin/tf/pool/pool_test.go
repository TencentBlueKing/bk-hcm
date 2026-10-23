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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/kit"
)

const waitTimeout = 3 * time.Second

func newTestPool(t *testing.T, s *fakeStarter, cfg PoolConfig) *PluginPool {
	t.Helper()

	if cfg.MaxQueueLength == 0 {
		cfg.MaxQueueLength = 100
	}
	p, err := NewPluginPool(Options{PluginName: "tcloud", PluginVersion: "1.0.0", ExecPath: "/bin/fake", Starter: s,
		Config: cfg})
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
		defer cancel()
		_ = p.Close(ctx)
	})
	return p
}

func providerConfig(secretID string) map[string]any {
	return map[string]any{"secret_id": secretID, "secret_key": "key-of-" + secretID}
}

func mustAcquire(t *testing.T, p *PluginPool, secretID string) *Lease {
	t.Helper()

	l, err := p.Acquire(kit.New(), providerConfig(secretID))
	require.NoError(t, err)
	require.NotNil(t, l)
	return l
}

// identityOf 返回 lease 所在进程的 "进程 ID / 当前配置"。
func identityOf(t *testing.T, l *Lease) string {
	t.Helper()

	resp, err := l.PlanResourceChange(kit.New(), &tfruntime.PlanResourceChangeRequest{TypeName: "x"})
	require.NoError(t, err)
	return string(resp.PlannedPrivate)
}

type acquireResult struct {
	lease *Lease
	err   error
}

func acquireAsync(p *PluginPool, kt *kit.Kit, secretID string) <-chan acquireResult {
	ch := make(chan acquireResult, 1)
	go func() {
		l, err := p.Acquire(kt, providerConfig(secretID))
		ch <- acquireResult{lease: l, err: err}
	}()
	return ch
}

func waitResult(t *testing.T, ch <-chan acquireResult) acquireResult {
	t.Helper()

	select {
	case r := <-ch:
		return r
	case <-time.After(waitTimeout):
		t.Fatal("acquire did not return in time")
		return acquireResult{}
	}
}

func assertPending(t *testing.T, ch <-chan acquireResult) {
	t.Helper()

	select {
	case r := <-ch:
		t.Fatalf("acquire should still be waiting, got lease=%v err=%v", r.lease != nil, r.err)
	case <-time.After(50 * time.Millisecond):
	}
}

func waitWaiters(t *testing.T, p *PluginPool, n int) {
	t.Helper()
	require.Eventually(t, func() bool { return p.stats().waiters == n }, waitTimeout, time.Millisecond)
}

func waitSettled(t *testing.T, p *PluginPool) {
	t.Helper()
	require.Eventually(t, func() bool { return p.stats().settled() }, waitTimeout, time.Millisecond)
	assertInvariants(t, p)
}

func TestNewValidatesOptions(t *testing.T) {
	cfg := PoolConfig{MaxProcesses: 1, MaxQueueLength: 1}
	s := &fakeStarter{}

	_, err := NewPluginPool(Options{ExecPath: "/bin/fake", Starter: s, Config: cfg})
	assert.Error(t, err)
	_, err = NewPluginPool(Options{PluginName: "tcloud", Starter: s, Config: cfg})
	assert.Error(t, err)
	_, err = NewPluginPool(Options{PluginName: "tcloud", ExecPath: "/bin/fake", Config: cfg})
	assert.Error(t, err)
	_, err = NewPluginPool(Options{PluginName: "tcloud", ExecPath: "/bin/fake", Starter: s})
	assert.Error(t, err)
}

func TestAcquireCreatesConfiguresAndReuses(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2})

	l1 := mustAcquire(t, p, "A")
	assert.Equal(t, "1/A", identityOf(t, l1))
	assert.EqualValues(t, 1, s.starts.Load())
	assert.EqualValues(t, 1, s.configures.Load())
	assertInvariants(t, p)

	l1.Release()
	assert.Equal(t, 0, p.stats().inflight)

	l2 := mustAcquire(t, p, "A")
	assert.Equal(t, "1/A", identityOf(t, l2), "same config reuses the configured process")
	assert.EqualValues(t, 1, s.starts.Load())
	assert.EqualValues(t, 1, s.configures.Load(), "no reconfigure for the same config")
	l2.Release()
	assertInvariants(t, p)
}

func TestReconfigureIdleProcess(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	mustAcquire(t, p, "A").Release()

	l := mustAcquire(t, p, "B")
	assert.Equal(t, "1/B", identityOf(t, l), "idle process is reconfigured instead of creating")
	assert.EqualValues(t, 1, s.starts.Load())
	assert.EqualValues(t, 2, s.configures.Load())
	l.Release()
}

func TestPreferConfiguredProcess(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2})

	la := mustAcquire(t, p, "A")
	lb := mustAcquire(t, p, "B")
	idB := identityOf(t, lb)
	la.Release()
	lb.Release()

	l := mustAcquire(t, p, "B")
	assert.Equal(t, idB, identityOf(t, l), "process already configured with B is preferred")
	assert.EqualValues(t, 2, s.configures.Load())
	l.Release()
}

func TestBusyProcessIsNotReconfigured(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 2})

	la := mustAcquire(t, p, "A")
	chB := acquireAsync(p, kit.New(), "B")
	assertPending(t, chB)
	assert.Equal(t, "1/A", identityOf(t, la), "in-flight lease keeps its config")

	la.Release()
	r := waitResult(t, chB)
	require.NoError(t, r.err)
	assert.Equal(t, "1/B", identityOf(t, r.lease))
	r.lease.Release()
}

// TestStrictFIFOHeadBlocks 队首缺资源时，后面本可填空位的请求也要等（技术方案 4.3 的例子）。
func TestStrictFIFOHeadBlocks(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2, MaxLeasesPerProcess: 2})

	la := mustAcquire(t, p, "A")  // P1 = A，inflight 1
	lb1 := mustAcquire(t, p, "B") // P2 = B，inflight 1
	lb2 := mustAcquire(t, p, "B") // P2 = B，inflight 2（满）
	require.Equal(t, identityOf(t, lb1), identityOf(t, lb2))

	chC := acquireAsync(p, kit.New(), "C") // 队首：无匹配、无空闲、容量已满
	waitWaiters(t, p, 1)
	chA := acquireAsync(p, kit.New(), "A") // P1 有空位，但不能越过队首
	waitWaiters(t, p, 2)
	assertPending(t, chA)

	lb1.Release() // P2 仍有在途 lease，不能重配
	assertPending(t, chC)
	assertPending(t, chA)

	lb2.Release() // P2 空闲 → 队首 C 重配 P2，随后 A 落到 P1
	rc := waitResult(t, chC)
	ra := waitResult(t, chA)
	require.NoError(t, rc.err)
	require.NoError(t, ra.err)
	assert.Equal(t, identityOf(t, la), identityOf(t, ra.lease))
	assert.True(t, strings.HasSuffix(identityOf(t, rc.lease), "/C"))

	la.Release()
	rc.lease.Release()
	ra.lease.Release()
	waitSettled(t, p)
}

func TestCanceledHeadUnblocksNext(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 2})

	la := mustAcquire(t, p, "A")

	ctx, cancel := context.WithCancel(context.Background())
	kt := kit.New()
	kt.Ctx = ctx
	chB := acquireAsync(p, kt, "B") // 队首，等不到资源
	waitWaiters(t, p, 1)
	chA := acquireAsync(p, kit.New(), "A")
	waitWaiters(t, p, 2)
	assertPending(t, chA)

	cancel()
	rb := waitResult(t, chB)
	assert.ErrorIs(t, rb.err, context.Canceled)

	ra := waitResult(t, chA)
	require.NoError(t, ra.err, "cancel of the head lets the next request through")
	assert.Equal(t, identityOf(t, la), identityOf(t, ra.lease))

	la.Release()
	ra.lease.Release()
	assert.Equal(t, 0, p.stats().waiters)
	assertInvariants(t, p)
}

func TestQueueFullAndTimeout(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxQueueLength: 1, MaxQueueWait: 100 * time.Millisecond})

	la := mustAcquire(t, p, "A")

	chB := acquireAsync(p, kit.New(), "B")
	waitWaiters(t, p, 1)

	_, err := p.Acquire(kit.New(), providerConfig("C"))
	assert.ErrorIs(t, err, ErrTooManyWaiters)

	rb := waitResult(t, chB)
	assert.ErrorIs(t, rb.err, ErrAcquireTimeout)
	assert.Equal(t, 0, p.stats().waiters, "timed out request leaves the queue")

	la.Release()
	assertInvariants(t, p)
}

func TestLeaseReleaseIsIdempotent(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, MaxLeasesPerProcess: 2})

	l1 := mustAcquire(t, p, "A")
	l2 := mustAcquire(t, p, "A")
	assert.Equal(t, 2, p.stats().inflight)

	l1.Release()
	l1.Release()
	assert.Equal(t, 1, p.stats().inflight, "double release must not double decrement")

	_, err := l1.PlanResourceChange(kit.New(), &tfruntime.PlanResourceChangeRequest{})
	assert.ErrorIs(t, err, ErrLeaseReleased)
	_, err = l1.ApplyResourceChange(kit.New(), &tfruntime.ApplyResourceChangeRequest{})
	assert.ErrorIs(t, err, ErrLeaseReleased)

	l2.Release()
	assert.Equal(t, 0, p.stats().inflight)
	assertInvariants(t, p)
}

func TestConfigExpiryReconfiguresSameProcess(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 2, MaxConfigAge: time.Minute})

	var mu sync.Mutex
	now := baseTime
	p.now = func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return now
	}

	l := mustAcquire(t, p, "A")
	id := identityOf(t, l)
	l.Release()

	mu.Lock()
	now = now.Add(2 * time.Minute)
	mu.Unlock()

	l = mustAcquire(t, p, "A")
	assert.Equal(t, id, identityOf(t, l), "expired config is refreshed on the same process")
	assert.EqualValues(t, 1, s.starts.Load())
	assert.EqualValues(t, 2, s.configures.Load())
	l.Release()
}

func TestStartFailureFreesCapacity(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) { s.startFn = func(*fakeManager) error { return errors.New("handshake failed") } })
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start plugin tcloud process failed")
	assert.NotContains(t, err.Error(), "key-of-A")

	waitSettled(t, p)
	assert.Equal(t, 0, p.stats().total(), "failed process is removed")
	ms := s.allManagers()
	require.Len(t, ms, 1)
	assert.EqualValues(t, 1, ms[0].closeCalls.Load(), "partially started process is closed")

	s.set(func(s *fakeStarter) { s.startFn = nil })
	l := mustAcquire(t, p, "A")
	l.Release()
}

func TestNewProcessManagerFailure(t *testing.T) {
	s := &fakeStarter{newErr: errors.New("bad exec path")}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	waitSettled(t, p)
	assert.Equal(t, 0, p.stats().total())
}

func TestConfigureBusinessErrorIsDistinguishable(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error {
			return &tfruntime.DiagnosticsError{
				Diagnostics: []tfruntime.Diagnostic{
					{Severity: tfruntime.SeverityError, Summary: "missing secret"},
				},
			}
		}
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	var de *tfruntime.DiagnosticsError
	assert.True(t, errors.As(err, &de), "business error keeps its type through the pool")
	assert.NotContains(t, err.Error(), "key-of-A")

	waitSettled(t, p)
	assert.Equal(t, 0, p.stats().total(), "stage 1 closes the process on any configure failure")
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
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, ConfigureTimeout: 100 * time.Millisecond})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	waitSettled(t, p)
	assert.Equal(t, 0, p.stats().total())

	s.set(func(s *fakeStarter) { s.configureFn = nil })
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
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1, StartTimeout: 100 * time.Millisecond})

	_, err := p.Acquire(kit.New(), providerConfig("A"))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	waitSettled(t, p)
	assert.Equal(t, 0, p.stats().total())
}

func TestPanicInPluginCallIsContained(t *testing.T) {
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error { panic("boom") }
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

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
	st := p.stats()
	assert.Equal(t, 1, st.states[stateRunning], "process is kept after the request is gone")
	assert.EqualValues(t, 0, s.configures.Load(), "no waiter, no configure")

	l := mustAcquire(t, p, "A")
	assert.Equal(t, "1/A", identityOf(t, l))
	assert.EqualValues(t, 1, s.starts.Load())
	l.Release()
}

func TestCloseFailsWaitersAndClosesProcesses(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	la := mustAcquire(t, p, "A")
	chB := acquireAsync(p, kit.New(), "B")
	waitWaiters(t, p, 1)

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	require.NoError(t, p.Close(ctx))

	assert.ErrorIs(t, waitResult(t, chB).err, ErrPoolClosed)
	_, err := p.Acquire(kit.New(), providerConfig("A"))
	assert.ErrorIs(t, err, ErrPoolClosed)

	ms := s.allManagers()
	require.Len(t, ms, 1)
	assert.EqualValues(t, 1, ms[0].closeCalls.Load())
	assert.Equal(t, 0, p.stats().total())

	_, err = la.PlanResourceChange(kit.New(), &tfruntime.PlanResourceChangeRequest{})
	assert.Error(t, err, "calls on a closed process fail")
	la.Release()

	require.NoError(t, p.Close(ctx), "close is idempotent")
	assertInvariants(t, p)
}

func TestCloseWhileStarting(t *testing.T) {
	gate := make(chan struct{})
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) { s.startFn = func(*fakeManager) error { <-gate; return nil } })
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	ch := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.starts.Load() == 1 }, waitTimeout, time.Millisecond)

	closed := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
		defer cancel()
		closed <- p.Close(ctx)
	}()

	assert.ErrorIs(t, waitResult(t, ch).err, ErrPoolClosed, "waiter fails immediately on close")
	select {
	case <-closed:
		t.Fatal("close must wait for the starting process")
	case <-time.After(50 * time.Millisecond):
	}

	close(gate)
	require.NoError(t, <-closed)
	ms := s.allManagers()
	require.Len(t, ms, 1)
	assert.EqualValues(t, 1, ms[0].closeCalls.Load(), "process started after close is closed")
	assert.EqualValues(t, 0, s.configures.Load())
	assert.Equal(t, 0, p.stats().total())
}

func TestCloseTimeout(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})
	mustAcquire(t, p, "A").Release()

	s.set(func(s *fakeStarter) {
		s.closeFn = func(_ *fakeManager, ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, p.Close(ctx), context.DeadlineExceeded)

	waitSettled(t, p)
	assert.Equal(t, 0, p.stats().total())
}

func TestCloseWhileConfiguring(t *testing.T) {
	gate := make(chan struct{})
	s := &fakeStarter{}
	s.set(func(s *fakeStarter) {
		s.configureFn = func(*fakeManager, map[string]any) error { <-gate; return nil }
	})
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})

	ch := acquireAsync(p, kit.New(), "A")
	require.Eventually(t, func() bool { return s.configures.Load() == 1 }, waitTimeout, time.Millisecond)

	closed := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
		defer cancel()
		closed <- p.Close(ctx)
	}()
	assert.ErrorIs(t, waitResult(t, ch).err, ErrPoolClosed)

	close(gate)
	require.NoError(t, <-closed)
	assert.Equal(t, 0, p.stats().total(), "process configured after close is closed, not delivered")
	assert.EqualValues(t, 1, s.allManagers()[0].closeCalls.Load())
}

func TestLeaseForwardsBusinessMethods(t *testing.T) {
	s := &fakeStarter{}
	p := newTestPool(t, s, PoolConfig{MaxProcesses: 1})
	l := mustAcquire(t, p, "A")
	kt := kit.New()

	calls := map[string]func() error{
		"ValidateResourceTypeConfig": func() error {
			_, err := l.ValidateResourceTypeConfig(kt, &tfruntime.ValidateResourceTypeConfigRequest{})
			return err
		},
		"ValidateDataSourceConfig": func() error {
			_, err := l.ValidateDataSourceConfig(kt, &tfruntime.ValidateDataSourceConfigRequest{})
			return err
		},
		"UpgradeResourceState": func() error {
			_, err := l.UpgradeResourceState(kt, &tfruntime.UpgradeResourceStateRequest{})
			return err
		},
		"ReadResource": func() error {
			_, err := l.ReadResource(kt, &tfruntime.ReadResourceRequest{})
			return err
		},
		"PlanResourceChange": func() error {
			_, err := l.PlanResourceChange(kt, &tfruntime.PlanResourceChangeRequest{})
			return err
		},
		"ApplyResourceChange": func() error {
			_, err := l.ApplyResourceChange(kt, &tfruntime.ApplyResourceChangeRequest{})
			return err
		},
		"ImportResourceState": func() error {
			_, err := l.ImportResourceState(kt, &tfruntime.ImportResourceStateRequest{})
			return err
		},
		"ReadDataSource": func() error {
			_, err := l.ReadDataSource(kt, &tfruntime.ReadDataSourceRequest{})
			return err
		},
	}

	for name, call := range calls {
		assert.NoError(t, call(), "%s while held", name)
	}
	l.Release()
	for name, call := range calls {
		assert.ErrorIs(t, call(), ErrLeaseReleased, "%s after release", name)
	}
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
