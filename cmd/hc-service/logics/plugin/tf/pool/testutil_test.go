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
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	if cfg.MinIdleProcesses == 0 && cfg.MaxProcesses > 0 && cfg.MaxProcesses < defaultMinIdleProcesses {
		cfg.MinIdleProcesses = cfg.MaxProcesses
	}
	if cfg.MaxLeasesPerProcess == 0 {
		cfg.MaxLeasesPerProcess = 1
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

// waitPrewarmed 等池启动时的预热完成：minIdle 个空闲进程已拉起且不再有进程在准备。
func waitPrewarmed(t *testing.T, p *PluginPool, minIdle int) {
	t.Helper()
	require.Eventually(t, func() bool {
		st := p.testStats()
		return st.states[stateRunning] == minIdle && st.settled()
	}, waitTimeout, time.Millisecond, "idle processes are prewarmed when the pool starts")
}

// preparingWithPending 返回正在准备（拉起或配置）且等着交付的请求最多的进程的 pending 数。
func preparingWithPending(p *PluginPool) int {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	most := 0
	for _, proc := range p.procs {
		if proc.state == stateConfiguring || proc.state == stateStarting {
			most = max(most, len(proc.pending))
		}
	}
	return most
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
	require.Eventually(t, func() bool { return p.testStats().waiters == n }, waitTimeout, time.Millisecond)
}

func waitSettled(t *testing.T, p *PluginPool) {
	t.Helper()
	require.Eventually(t, func() bool { return p.testStats().settled() }, waitTimeout, time.Millisecond)
	assertInvariants(t, p)
}

// fakeStarter 测试用进程管理工厂。钩子在 mu 保护下读写，可在测试中途修改。
type fakeStarter struct {
	mu sync.Mutex
	// newErr NewProcessManager 返回的错误。
	newErr error
	// startFn Start 的行为，nil 表示成功。可阻塞（例如等待 m.closed()）以模拟不响应 ctx 的调用。
	startFn func(m *fakeManager) error
	// configureFn ConfigureProvider 的行为，nil 表示成功。
	configureFn func(m *fakeManager, config map[string]any) error
	// closeFn Close 的行为，nil 表示立即退出；返回错误表示未能确认退出。
	closeFn func(m *fakeManager, ctx context.Context) error

	managers []*fakeManager

	starts     atomic.Int64
	configures atomic.Int64
}

func (s *fakeStarter) set(fn func(s *fakeStarter)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s)
}

func (s *fakeStarter) hooks() (func(*fakeManager) error, func(*fakeManager, map[string]any) error,
	func(*fakeManager, context.Context) error) {

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.startFn, s.configureFn, s.closeFn
}

// NewProcessManager 实现 tfruntime.Starter。
func (s *fakeStarter) NewProcessManager(_ *kit.Kit, _, _, _ string) (tfruntime.ProcessManager, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.newErr != nil {
		return nil, s.newErr
	}
	m := &fakeManager{starter: s, id: len(s.managers) + 1, closedCh: make(chan struct{})}
	s.managers = append(s.managers, m)
	return m, nil
}

func (s *fakeStarter) allManagers() []*fakeManager {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*fakeManager(nil), s.managers...)
}

// fakeManager 测试用进程管理。Close 成功后 closed() 关闭，阻塞在其上的调用随之返回。
type fakeManager struct {
	starter    *fakeStarter
	id         int
	closedCh   chan struct{}
	closeOnce  sync.Once
	closeCalls atomic.Int64
}

var _ tfruntime.ProcessManager = (*fakeManager)(nil)

func (m *fakeManager) closed() <-chan struct{} {
	return m.closedCh
}

// Start 实现 tfruntime.ProcessManager。
func (m *fakeManager) Start(_ *kit.Kit) (tfruntime.Executor, error) {
	m.starter.starts.Add(1)

	startFn, _, _ := m.starter.hooks()
	if startFn != nil {
		if err := startFn(m); err != nil {
			return nil, err
		}
	}
	return &fakeExecutor{m: m}, nil
}

// Close 实现 tfruntime.ProcessManager。
func (m *fakeManager) Close(ctx context.Context) error {
	m.closeCalls.Add(1)

	_, _, closeFn := m.starter.hooks()
	if closeFn != nil {
		if err := closeFn(m, ctx); err != nil {
			return err
		}
	}
	m.closeOnce.Do(func() { close(m.closedCh) })
	return nil
}

// Exited 实现 tfruntime.ProcessManager。
func (m *fakeManager) Exited() bool {
	select {
	case <-m.closedCh:
		return true
	default:
		return false
	}
}

// CheckHealth 实现 tfruntime.ProcessManager。
func (m *fakeManager) CheckHealth() error {
	return nil
}

var errExecutorClosed = errors.New("fake executor: process closed")

// fakeExecutor 测试用执行器。记录当前生效的配置（secret_id），业务调用返回"进程 ID / 当前配置"，
// 测试据此校验 lease 落在哪个进程、该进程此刻是否配置成调用方的配置。
type fakeExecutor struct {
	m *fakeManager

	mu      sync.Mutex
	current string
}

var _ tfruntime.Executor = (*fakeExecutor)(nil)

func (e *fakeExecutor) check() error {
	if e.m.Exited() {
		return errExecutorClosed
	}
	return nil
}

// identity 返回 "进程 ID / 当前配置"，供测试断言。
func (e *fakeExecutor) identity() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return []byte(fmt.Sprintf("%d/%s", e.m.id, e.current))
}

// GetProviderSchema 实现 tfruntime.Executor。
func (e *fakeExecutor) GetProviderSchema(_ *kit.Kit) (*tfruntime.ProviderSchema, error) {
	return &tfruntime.ProviderSchema{}, e.check()
}

// ConfigureProvider 实现 tfruntime.Executor。
func (e *fakeExecutor) ConfigureProvider(_ *kit.Kit, req *tfruntime.ConfigureProviderRequest) (
	*tfruntime.ConfigureProviderResponse, error) {

	e.m.starter.configures.Add(1)
	if err := e.check(); err != nil {
		return nil, err
	}

	_, configureFn, _ := e.m.starter.hooks()
	if configureFn != nil {
		if err := configureFn(e.m, req.Config); err != nil {
			return nil, err
		}
	}

	e.mu.Lock()
	e.current, _ = req.Config["secret_id"].(string)
	e.mu.Unlock()
	return &tfruntime.ConfigureProviderResponse{}, nil
}

// ValidateResourceTypeConfig 实现 tfruntime.Executor。
func (e *fakeExecutor) ValidateResourceTypeConfig(_ *kit.Kit, _ *tfruntime.ValidateResourceTypeConfigRequest) (
	*tfruntime.ValidateResourceTypeConfigResponse, error) {

	return &tfruntime.ValidateResourceTypeConfigResponse{}, e.check()
}

// ValidateDataSourceConfig 实现 tfruntime.Executor。
func (e *fakeExecutor) ValidateDataSourceConfig(_ *kit.Kit, _ *tfruntime.ValidateDataSourceConfigRequest) (
	*tfruntime.ValidateDataSourceConfigResponse, error) {

	return &tfruntime.ValidateDataSourceConfigResponse{}, e.check()
}

// UpgradeResourceState 实现 tfruntime.Executor。
func (e *fakeExecutor) UpgradeResourceState(_ *kit.Kit, _ *tfruntime.UpgradeResourceStateRequest) (
	*tfruntime.UpgradeResourceStateResponse, error) {

	return &tfruntime.UpgradeResourceStateResponse{}, e.check()
}

// ReadResource 实现 tfruntime.Executor。
func (e *fakeExecutor) ReadResource(_ *kit.Kit, _ *tfruntime.ReadResourceRequest) (
	*tfruntime.ReadResourceResponse, error) {

	return &tfruntime.ReadResourceResponse{}, e.check()
}

// PlanResourceChange 实现 tfruntime.Executor，PlannedPrivate 返回 identity。
func (e *fakeExecutor) PlanResourceChange(_ *kit.Kit, _ *tfruntime.PlanResourceChangeRequest) (
	*tfruntime.PlanResourceChangeResponse, error) {

	if err := e.check(); err != nil {
		return nil, err
	}
	return &tfruntime.PlanResourceChangeResponse{PlannedPrivate: e.identity()}, nil
}

// ApplyResourceChange 实现 tfruntime.Executor，Private 返回 identity。
func (e *fakeExecutor) ApplyResourceChange(_ *kit.Kit, _ *tfruntime.ApplyResourceChangeRequest) (
	*tfruntime.ApplyResourceChangeResponse, error) {

	if err := e.check(); err != nil {
		return nil, err
	}
	return &tfruntime.ApplyResourceChangeResponse{Private: e.identity()}, nil
}

// ImportResourceState 实现 tfruntime.Executor。
func (e *fakeExecutor) ImportResourceState(_ *kit.Kit, _ *tfruntime.ImportResourceStateRequest) (
	*tfruntime.ImportResourceStateResponse, error) {

	return &tfruntime.ImportResourceStateResponse{}, e.check()
}

// ReadDataSource 实现 tfruntime.Executor。
func (e *fakeExecutor) ReadDataSource(_ *kit.Kit, _ *tfruntime.ReadDataSourceRequest) (
	*tfruntime.ReadDataSourceResponse, error) {

	return &tfruntime.ReadDataSourceResponse{}, e.check()
}

// checkInvariantsLocked 校验池的不变式（技术方案 6.11），调用方持有 p.poolLock。
//
//	I1  0 <= inflight <= MaxLeasesPerProcess，且全池 inflight 总和等于在途 lease 数
//	I2  每个请求的结果只写入 ch 一次（由 requestDone 与 ch 缓冲 1 保证）
//	I3  进程数不超过 MaxProcesses
//	I6  请求至多处于一处：队列、某进程的 pending、已结束
//	I8  starting / configuring 期间 inflight == 0；若有 binding，boundAt 必须为零值，pending 非空且不超过上限；预热进程允许 binding == nil
//	I9  若队首请求的 key 与某个 running 进程已生效配置一致且 inflight < MaxLeasesPerProcess，pickLocked 必选中 actionDeliver（且不会停留在队首）
//	I10 queue.Len() 等于队列中的有效请求数
func (p *PluginPool) checkInvariantsLocked() error {
	if len(p.procs) > p.cfg.MaxProcesses {
		return fmt.Errorf("I3: %d processes exceed max %d", len(p.procs), p.cfg.MaxProcesses)
	}

	seen := make(map[*request]string)
	valid := 0
	var headReq *request
	for _, w := range p.queue.buf[p.queue.head:] {
		if _, dead := p.queue.removed[w]; dead {
			if w.state != requestDone {
				return fmt.Errorf("I6: removed request %p is %s", w, w.state)
			}
			continue
		}
		if headReq == nil {
			headReq = w
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

	// I9：若队列非空且存在立即可交付的同 key running 进程，pickLocked 必须返回 actionDeliver。
	if headReq != nil && !p.closed {
		now := p.now()
		hasMatchedFree := false
		for _, proc := range p.procs {
			if proc.state == stateRunning && proc.binding.valid(now, p.cfg.MaxConfigAge) &&
				proc.binding.key == headReq.key && proc.inflight < p.cfg.MaxLeasesPerProcess {
				hasMatchedFree = true
				break
			}
		}
		if hasMatchedFree {
			if _, act := p.pickLocked(now, headReq.key); act != actionDeliver {
				return fmt.Errorf("I9: matched free process exists for head key, but pickLocked returned %d", act)
			}
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
		if proc.binding != nil {
			if !proc.binding.configuring() {
				return fmt.Errorf("I8: preparing process %d has an effective binding", id)
			}
			if len(proc.pending) == 0 && !p.closed {
				return fmt.Errorf("I8: preparing process %d with binding has empty pending", id)
			}
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

func (p *PluginPool) testStats() poolStats {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	s := poolStats{
		states:  make(map[procState]int),
		waiters: p.queue.Len(),
	}
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
