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
	"time"

	"hcm/cmd/hc-service/logics/plugin/query"
	"hcm/cmd/hc-service/logics/plugin/tf/discovery"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// removedPoolCloseTimeout 关闭一个已停用插件的池时，最长等待多久。
const removedPoolCloseTimeout = time.Minute

// PluginRegistryOptions 创建 PluginRegistry 的参数。
type PluginRegistryOptions struct {
	// Pool 插件池的参数。目前所有插件池使用同一份，后续可能按插件类型差异化配置。
	Pool PoolConfig
	// CheckInterval 读取插件表、建池或关池的周期。
	CheckInterval time.Duration
}

// PluginRegistry 本副本所有 tf 插件的进程池，按插件名索引（内存结构，与数据库的 iac_plugin 插件表不是一回事）。
// 上游请求按插件名用 Get 取得对应的池。
//
// 本副本目录里有、且插件表里该插件已启用的才有池。后台按 CheckInterval 读取插件表：
// 新启用的插件建池并预热，被停用的插件关池。读取插件表失败只记日志、跳过本轮，不会因此关池。
type PluginRegistry struct {
	dataCli *dataservice.Client
	entries []discovery.Entry
	opts    PluginRegistryOptions

	// ctx 后台检查的 context，Close 时取消。
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.RWMutex
	pools  map[string]*PluginPool
	closed bool
}

// NewPluginRegistry 创建 PluginRegistry，此时不建池。池参数不合法时返回错误，调用方应让服务不启动。
func NewPluginRegistry(dataCli *dataservice.Client, entries []discovery.Entry,
	opts PluginRegistryOptions) (*PluginRegistry, error) {

	if dataCli == nil {
		return nil, errors.New("data service client should not be nil")
	}
	if opts.CheckInterval <= 0 {
		return nil, errors.New("registry check interval should be greater than 0")
	}

	opts.Pool.trySetDefault()
	if err := opts.Pool.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pool config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &PluginRegistry{
		dataCli: dataCli,
		entries: entries,
		opts:    opts,
		ctx:     ctx,
		cancel:  cancel,
		pools:   make(map[string]*PluginPool),
	}, nil
}

// StartSync 启动后台同步：立刻同步一次，之后按周期同步。
func (r *PluginRegistry) StartSync() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	r.wg.Add(1)
	go r.syncLoop()
}

// GetPool 返回插件对应的池。插件未启用、或本副本没有该插件时返回 ErrPluginNotEnabled。
func (r *PluginRegistry) GetPool(pluginName string) (*PluginPool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.pools[pluginName]
	if !ok {
		return nil, ErrPluginNotEnabled
	}
	return p, nil
}

// Close 停止后台检查并关闭全部池。ctx 结束时不再等待，返回 ctx 的错误；重复调用返回 nil。
func (r *PluginRegistry) Close(ctx context.Context) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	pools := r.pools
	r.pools = make(map[string]*PluginPool)
	r.mu.Unlock()

	r.cancel()
	r.wg.Wait()

	return closePools(ctx, pools)
}

func (r *PluginRegistry) syncLoop() {
	defer r.wg.Done()

	r.syncOnce()

	ticker := time.NewTicker(r.opts.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.syncOnce()
		}
	}
}

// syncOnce 读取插件表，让池与已启用的插件保持一致。
func (r *PluginRegistry) syncOnce() {
	if len(r.entries) == 0 {
		return
	}

	// 后端任务使用后端操作 kit：data-service 会校验请求头里的操作人与 AppCode。
	kt := core.NewBackendKit()
	kt.Ctx = r.ctx

	names := slice.Map(r.entries, func(e discovery.Entry) string { return e.PluginName })
	expr := tools.ExpressionAnd(
		tools.RuleEqual("protocol", enumor.IaCPluginProtocolTFProvider),
		tools.RuleIn("plugin_name", names),
	)
	records, err := query.List(kt, r.dataCli, expr)
	if err != nil {
		logs.Warnf("list tf plugins failed, skip pool sync this round, err: %v, rid: %s", err, kt.Rid)
		return
	}

	enabled, err := enabledPlugins(records)
	if err != nil {
		logs.Warnf("find enabled tf plugins failed, skip pool sync this round, err: %v, rid: %s", err, kt.Rid)
		return
	}

	if err = r.syncPools(kt, enabled); err != nil {
		logs.Errorf("close pools failed, err: %v, rid: %s", err, kt.Rid)
	}
}

// syncPools 让池与 enabled 一致：给已启用却没有池的本地插件启动池，关掉不再启用的池。
func (r *PluginRegistry) syncPools(kt *kit.Kit, enabled map[string]bool) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}

	// 新增已启用的池
	for _, entry := range r.entries {
		if !enabled[entry.PluginName] || r.pools[entry.PluginName] != nil {
			continue
		}
		p, err := NewPluginPool(Options{
			PluginName:    entry.PluginName,
			PluginVersion: entry.PluginVersion,
			ExecPath:      entry.ExecPath,
			Config:        r.opts.Pool,
		})
		if err != nil {
			logs.Errorf("create pool of plugin %s failed, err: %v, rid: %s", entry.PluginName, err, kt.Rid)
			continue
		}

		r.pools[entry.PluginName] = p
		logs.Infof("created process pool of plugin %s@%s, rid: %s", entry.PluginName, entry.PluginVersion, kt.Rid)
	}

	// 删除不再启用的池
	toClose := make(map[string]*PluginPool)
	for name, p := range r.pools {
		if !enabled[name] {
			toClose[name] = p
			delete(r.pools, name)
		}
	}
	r.mu.Unlock()

	if len(toClose) == 0 {
		return nil
	}

	closeCtx, cancel := context.WithTimeout(r.ctx, removedPoolCloseTimeout)
	defer cancel()
	return closePools(closeCtx, toClose)
}

// closePools 关闭一批池，返回合并后的错误。
func closePools(ctx context.Context, pools map[string]*PluginPool) error {
	var errs []error
	for name, p := range pools {
		if err := p.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("close pool of plugin %s: %w", name, err))
		}
	}
	return errors.Join(errs...)
}

// enabledPlugins 返回插件表里处于启用状态的 tf 插件名。
//
// 每个插件只看已完成（enabled / disabled）的最高版本：升级中的新版本 init / failed 时，当前版本仍在服务，
// 不能因此关池；已完成的最高版本是 disabled 则说明被管理员停用。
// 记录里有不合法的版本号时返回错误，调用方不应在不确定的数据上建池或关池。
func enabledPlugins(records []coreplugin.IaCPlugin) (map[string]bool, error) {
	completed := make(map[string][]coreplugin.IaCPlugin)
	for _, rec := range records {
		if rec.Protocol != enumor.IaCPluginProtocolTFProvider {
			continue
		}
		if rec.State != enumor.IaCPluginStateEnabled && rec.State != enumor.IaCPluginStateDisabled {
			continue
		}
		completed[rec.PluginName] = append(completed[rec.PluginName], rec)
	}

	enabled := make(map[string]bool)
	for name, recs := range completed {
		highest, err := iacplugin.Latest(recs)
		if err != nil {
			return nil, fmt.Errorf("plugin %s: %w", name, err)
		}
		if highest.State == enumor.IaCPluginStateEnabled {
			enabled[name] = true
		}
	}
	return enabled, nil
}
