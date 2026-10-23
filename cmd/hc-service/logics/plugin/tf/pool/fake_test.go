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

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/kit"
)

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
