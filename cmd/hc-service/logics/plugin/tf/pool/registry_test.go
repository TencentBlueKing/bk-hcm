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
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/cmd/hc-service/logics/plugin/tf/discovery"
	coreplugin "hcm/pkg/api/core/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

func tfRecord(name, version string, state enumor.IaCPluginState) coreplugin.IaCPlugin {
	return coreplugin.IaCPlugin{PluginName: name, PluginVersion: version, State: state,
		Protocol: enumor.IaCPluginProtocolTFProvider}
}

func TestEnabledPlugins(t *testing.T) {
	enabled, disabled, initial, failed := enumor.IaCPluginStateEnabled, enumor.IaCPluginStateDisabled,
		enumor.IaCPluginStateInit, enumor.IaCPluginStateFailed

	cases := []struct {
		name    string
		records []coreplugin.IaCPlugin
		want    []string
	}{
		{"enabled", []coreplugin.IaCPlugin{tfRecord("a", "1.0.0", enabled)}, []string{"a"}},
		{"disabled", []coreplugin.IaCPlugin{tfRecord("a", "1.0.0", disabled)}, nil},
		{
			// 新版本还在注册，当前版本仍在服务，不能关池。
			"upgrading keeps the completed version", []coreplugin.IaCPlugin{
				tfRecord("a", "1.0.0", enabled), tfRecord("a", "1.1.0", initial)}, []string{"a"},
		},
		{
			"failed upgrade keeps the completed version", []coreplugin.IaCPlugin{
				tfRecord("a", "1.0.0", enabled), tfRecord("a", "1.1.0", failed)}, []string{"a"},
		},
		{
			"highest completed version decides", []coreplugin.IaCPlugin{
				tfRecord("a", "1.0.0", enabled), tfRecord("a", "1.1.0", disabled)}, nil,
		},
		{
			"version order is semantic, not lexical", []coreplugin.IaCPlugin{
				tfRecord("a", "1.9.0", disabled), tfRecord("a", "1.10.0", enabled)}, []string{"a"},
		},
		{"never completed", []coreplugin.IaCPlugin{tfRecord("a", "1.0.0", initial)}, nil},
		{
			"plugins are independent", []coreplugin.IaCPlugin{
				tfRecord("a", "1.0.0", enabled), tfRecord("b", "1.0.0", disabled)}, []string{"a"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := enabledPlugins(c.records)
			require.NoError(t, err)
			for _, name := range c.want {
				assert.True(t, got[name], "%s should be enabled", name)
			}
			for name, on := range got {
				assert.True(t, !on || slices.Contains(c.want, name), "%s should not be enabled", name)
			}
		})
	}

	// 版本号不合法时不做判断，调用方据此跳过本轮，而不是在不确定的数据上关池。
	_, err := enabledPlugins([]coreplugin.IaCPlugin{tfRecord("a", "not-a-version", enabled)})
	assert.Error(t, err)
}

// newTestPluginRegistry 创建只含一个本地插件 tcloud 的 PluginRegistry。
func newTestPluginRegistry(t *testing.T, s *fakeStarter) *PluginRegistry {
	t.Helper()

	r, err := NewPluginRegistry(&dataservice.Client{}, []discovery.Entry{{PluginName: "tcloud", PluginVersion: "1.0.0",
		ExecPath: "/bin/fake"}}, s, PluginRegistryOptions{
		Pool:          PoolConfig{MaxProcesses: 2, MaxQueueLength: 1},
		CheckInterval: time.Hour,
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = r.Close(context.Background()) })
	return r
}

// TestPluginRegistrySyncPools 验证 PluginRegistry 按插件表里已启用的插件启动池、关闭停用的池，并通过 Get 按插件名取池。
func TestPluginRegistrySyncPools(t *testing.T) {
	r := newTestPluginRegistry(t, &fakeStarter{})
	kt := kit.New()
	on, off := map[string]bool{"tcloud": true}, map[string]bool{"tcloud": false}

	_, err := r.GetPool("tcloud")
	assert.ErrorIs(t, err, ErrPluginNotEnabled, "no pool before it is enabled")

	require.NoError(t, r.syncPools(kt, on))
	p, err := r.GetPool("tcloud")
	require.NoError(t, err)
	assert.Equal(t, "tcloud", p.pluginName)

	// 重复检查不重复建池。
	require.NoError(t, r.syncPools(kt, on))
	same, err := r.GetPool("tcloud")
	require.NoError(t, err)
	assert.Same(t, p, same)

	// 本副本没有的插件即使已启用也不建池。
	require.NoError(t, r.syncPools(kt, map[string]bool{"tcloud": true, "remote-only": true}))
	_, err = r.GetPool("remote-only")
	assert.ErrorIs(t, err, ErrPluginNotEnabled)

	// 被停用：从 r.pools 删掉并关闭。
	require.NoError(t, r.syncPools(kt, off))
	_, err = r.GetPool("tcloud")
	assert.ErrorIs(t, err, ErrPluginNotEnabled)
	_, err = p.Acquire(kit.New(), providerConfig("A"))
	assert.ErrorIs(t, err, ErrPoolClosed, "the disabled pool is closed")

	// 重新启用后建新池。
	require.NoError(t, r.syncPools(kt, on))
	again, err := r.GetPool("tcloud")
	require.NoError(t, err)
	assert.NotSame(t, p, again)
}

// TestPluginRegistryPrewarmsOnCreate PluginRegistry 建出的池在启动时就预热，不用等第一个请求。
func TestPluginRegistryPrewarmsOnCreate(t *testing.T) {
	s := &fakeStarter{}
	r := newTestPluginRegistry(t, s)

	require.NoError(t, r.syncPools(kit.New(), map[string]bool{"tcloud": true}))

	require.Eventually(t, func() bool { return s.starts.Load() == int64(defaultMinIdleProcesses) },
		waitTimeout, time.Millisecond, "idle processes are started right after the pool is created")
}

// TestPluginRegistryCloseClosesPools 关闭 PluginRegistry 会关闭全部池，且可重复调用。
func TestPluginRegistryCloseClosesPools(t *testing.T) {
	r := newTestPluginRegistry(t, &fakeStarter{})
	require.NoError(t, r.syncPools(kit.New(), map[string]bool{"tcloud": true}))
	p, err := r.GetPool("tcloud")
	require.NoError(t, err)

	require.NoError(t, r.Close(context.Background()))
	require.NoError(t, r.Close(context.Background()), "close is idempotent")

	_, err = p.Acquire(kit.New(), providerConfig("A"))
	assert.ErrorIs(t, err, ErrPoolClosed)
	_, err = r.GetPool("tcloud")
	assert.ErrorIs(t, err, ErrPluginNotEnabled)
}

// TestPluginRegistrySyncAfterClose 检查进行到一半时 PluginRegistry 被关闭：刚启动的池不登记，并在同步里关掉。
func TestPluginRegistrySyncAfterClose(t *testing.T) {
	r := newTestPluginRegistry(t, &fakeStarter{})
	require.NoError(t, r.Close(context.Background()))

	require.NoError(t, r.syncPools(kit.New(), map[string]bool{"tcloud": true}))

	_, err := r.GetPool("tcloud")
	assert.ErrorIs(t, err, ErrPluginNotEnabled)
}

// TestNewPluginRegistryRejectsInvalidConfig 池参数不合法时创建注册表就失败，服务据此拒绝启动，而不是等到建池时才发现。
func TestNewPluginRegistryRejectsInvalidConfig(t *testing.T) {
	cli := &dataservice.Client{}
	valid := PluginRegistryOptions{Pool: PoolConfig{MaxProcesses: 4, MaxQueueLength: 10}, CheckInterval: time.Second}

	_, err := NewPluginRegistry(cli, nil, &fakeStarter{}, valid)
	assert.NoError(t, err)

	tooManyIdle := valid
	tooManyIdle.Pool.MinIdleProcesses = 5
	_, err = NewPluginRegistry(cli, nil, &fakeStarter{}, tooManyIdle)
	assert.Error(t, err, "minIdleProcesses above maxProcesses")

	noInterval := valid
	noInterval.CheckInterval = 0
	_, err = NewPluginRegistry(cli, nil, &fakeStarter{}, noInterval)
	assert.Error(t, err)

	_, err = NewPluginRegistry(cli, nil, nil, valid)
	assert.Error(t, err, "starter is required")
	_, err = NewPluginRegistry(nil, nil, &fakeStarter{}, valid)
	assert.Error(t, err, "data service client is required")
}
