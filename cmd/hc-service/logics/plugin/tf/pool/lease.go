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
	"sync/atomic"
	"time"

	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/pkg/kit"
)

// ProviderIF 进程池交给业务调用方的插件方法：只有资源与数据源的业务方法。
//
// 配置（ConfigureProvider）由池在交付前完成，取 Schema 不属于运行期调用，二者都不对调用方开放。
type ProviderIF interface {
	ValidateResourceTypeConfig(kt *kit.Kit, req *tfruntime.ValidateResourceTypeConfigRequest) (
		*tfruntime.ValidateResourceTypeConfigResponse, error)
	ValidateDataSourceConfig(kt *kit.Kit, req *tfruntime.ValidateDataSourceConfigRequest) (
		*tfruntime.ValidateDataSourceConfigResponse, error)
	UpgradeResourceState(kt *kit.Kit, req *tfruntime.UpgradeResourceStateRequest) (
		*tfruntime.UpgradeResourceStateResponse, error)
	ReadResource(kt *kit.Kit, req *tfruntime.ReadResourceRequest) (*tfruntime.ReadResourceResponse, error)
	PlanResourceChange(kt *kit.Kit, req *tfruntime.PlanResourceChangeRequest) (
		*tfruntime.PlanResourceChangeResponse, error)
	ApplyResourceChange(kt *kit.Kit, req *tfruntime.ApplyResourceChangeRequest) (
		*tfruntime.ApplyResourceChangeResponse, error)
	ImportResourceState(kt *kit.Kit, req *tfruntime.ImportResourceStateRequest) (
		*tfruntime.ImportResourceStateResponse, error)
	ReadDataSource(kt *kit.Kit, req *tfruntime.ReadDataSourceRequest) (*tfruntime.ReadDataSourceResponse, error)
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
	exec   tfruntime.Executor
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
func (l *Lease) ValidateResourceTypeConfig(kt *kit.Kit, req *tfruntime.ValidateResourceTypeConfigRequest) (
	*tfruntime.ValidateResourceTypeConfigResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ValidateResourceTypeConfig(kt, req)
}

// ValidateDataSourceConfig 校验数据源查询条件。
func (l *Lease) ValidateDataSourceConfig(kt *kit.Kit, req *tfruntime.ValidateDataSourceConfigRequest) (
	*tfruntime.ValidateDataSourceConfigResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ValidateDataSourceConfig(kt, req)
}

// UpgradeResourceState 把状态升到当前 schema 版本。
func (l *Lease) UpgradeResourceState(kt *kit.Kit, req *tfruntime.UpgradeResourceStateRequest) (
	*tfruntime.UpgradeResourceStateResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.UpgradeResourceState(kt, req)
}

// ReadResource 读取资源现状。
func (l *Lease) ReadResource(kt *kit.Kit, req *tfruntime.ReadResourceRequest) (
	*tfruntime.ReadResourceResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ReadResource(kt, req)
}

// PlanResourceChange 计算计划。
func (l *Lease) PlanResourceChange(kt *kit.Kit, req *tfruntime.PlanResourceChangeRequest) (
	*tfruntime.PlanResourceChangeResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.PlanResourceChange(kt, req)
}

// ApplyResourceChange 按计划修改外部系统。
func (l *Lease) ApplyResourceChange(kt *kit.Kit, req *tfruntime.ApplyResourceChangeRequest) (
	*tfruntime.ApplyResourceChangeResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ApplyResourceChange(kt, req)
}

// ImportResourceState 按外部标识读入已有资源。
func (l *Lease) ImportResourceState(kt *kit.Kit, req *tfruntime.ImportResourceStateRequest) (
	*tfruntime.ImportResourceStateResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ImportResourceState(kt, req)
}

// ReadDataSource 读取数据源。
func (l *Lease) ReadDataSource(kt *kit.Kit, req *tfruntime.ReadDataSourceRequest) (
	*tfruntime.ReadDataSourceResponse, error) {

	if l.released.Load() {
		return nil, ErrLeaseReleased
	}
	return l.exec.ReadDataSource(kt, req)
}
