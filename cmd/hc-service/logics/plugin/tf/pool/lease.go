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
	"sync/atomic"
	"time"

	"hcm/cmd/hc-service/logics/tfruntime/executor"
	"hcm/cmd/hc-service/types"
)

// ProviderIF 进程池交给业务调用方的插件方法：只有资源与数据源的业务方法。
//
// 配置（ConfigureProvider）由池在交付前完成，取 Schema 不属于运行期调用，二者都不对调用方开放。
type ProviderIF interface {
	ValidateResourceTypeConfig(ctx context.Context, typeName string, config map[string]any) (
		*types.ValidateResult, error)
	ValidateDataSourceConfig(ctx context.Context, typeName string, config map[string]any) (
		*types.ValidateResult, error)
	UpgradeResourceState(ctx context.Context, typeName string, stateVersion int64, raw *types.RawState) (
		*types.UpgradeResult, error)
	ReadResource(ctx context.Context, typeName string, current map[string]any, private []byte) (
		*types.ReadResourceResult, error)
	PlanResourceChange(ctx context.Context, typeName string, prior, proposed, config map[string]any,
		priorPrivate []byte) (*types.PlanResult, error)
	ApplyResourceChange(ctx context.Context, typeName string, prior, planned, config map[string]any,
		plannedPrivate []byte) (*types.ApplyResult, error)
	ImportResourceState(ctx context.Context, typeName, id string) (*types.ImportResult, error)
	ReadDataSource(ctx context.Context, typeName string, config map[string]any) (*types.ReadDataSourceResult, error)
}

var _ ProviderIF = (*Lease)(nil)

// Lease 一个已按调用方配置配好的插件进程上的一个并发名额。
//
// 持有期内的多次调用都落在同一个进程上（同一次业务的 Plan 与 Apply 必须用同一个 lease）。
// 用完必须调用 Release，建议 defer。Release 之后再调用业务方法直接返回 ErrLeaseReleased，
// 不会打到进程上 —— 该进程此时可能已被重配成别人的配置。
//
// lease 只记录进程 ID，不持有池内进程记录：进程退出后记录被删除，归还时按 ID 查不到即忽略。
type Lease struct {
	pool   *PluginPool
	procID uint64
	exec   executor.Executor
	// acquiredAt 交付时间，归还时用于记录持有时长。
	acquiredAt time.Time
	released   atomic.Bool
}

// Release 归还 lease，幂等、不阻塞。进程是否仍可用由池判断，调用方无需也无法告知。
func (l *Lease) Release() {
	if l == nil || !l.released.CompareAndSwap(false, true) {
		return
	}
	l.pool.release(l)
}

// ValidateResourceTypeConfig 校验资源声明。
func (l *Lease) ValidateResourceTypeConfig(ctx context.Context, typeName string, config map[string]any) (
	*types.ValidateResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ValidateResourceTypeConfig(ctx, typeName, config)
}

// ValidateDataSourceConfig 校验数据源查询条件。
func (l *Lease) ValidateDataSourceConfig(ctx context.Context, typeName string, config map[string]any) (
	*types.ValidateResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ValidateDataSourceConfig(ctx, typeName, config)
}

// UpgradeResourceState 把状态升到当前 schema 版本。
func (l *Lease) UpgradeResourceState(ctx context.Context, typeName string, stateVersion int64, raw *types.RawState) (
	*types.UpgradeResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.UpgradeResourceState(ctx, typeName, stateVersion, raw)
}

// ReadResource 读取资源现状。
func (l *Lease) ReadResource(ctx context.Context, typeName string, current map[string]any, private []byte) (
	*types.ReadResourceResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ReadResource(ctx, typeName, current, private)
}

// PlanResourceChange 计算计划。
func (l *Lease) PlanResourceChange(ctx context.Context, typeName string, prior, proposed, config map[string]any,
	priorPrivate []byte) (*types.PlanResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.PlanResourceChange(ctx, typeName, prior, proposed, config, priorPrivate)
}

// ApplyResourceChange 按计划修改外部系统。
func (l *Lease) ApplyResourceChange(ctx context.Context, typeName string, prior, planned, config map[string]any,
	plannedPrivate []byte) (*types.ApplyResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ApplyResourceChange(ctx, typeName, prior, planned, config, plannedPrivate)
}

// ImportResourceState 按外部标识读入已有资源。
func (l *Lease) ImportResourceState(ctx context.Context, typeName, id string) (*types.ImportResult, error) {
	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ImportResourceState(ctx, typeName, id)
}

// ReadDataSource 读取数据源。
func (l *Lease) ReadDataSource(ctx context.Context, typeName string, config map[string]any) (
	*types.ReadDataSourceResult, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ReadDataSource(ctx, typeName, config)
}
