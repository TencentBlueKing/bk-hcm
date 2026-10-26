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
func newTestPluginRegistry(t *testing.T) *PluginRegistry {
	t.Helper()

	r, err := NewPluginRegistry(&dataservice.Client{}, []discovery.Entry{{PluginName: "tcloud", PluginVersion: "1.0.0",
		ExecPath: "/bin/fake"}}, PluginRegistryOptions{
		Pool:          PoolConfig{MaxProcesses: 2, MaxQueueLength: 1},
		CheckInterval: time.Hour,
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = r.Close(context.Background()) })
	return r
}

// TestPluginRegistrySyncAfterClose PluginRegistry 关闭后再执行同步不会登记新池。
func TestPluginRegistrySyncAfterClose(t *testing.T) {
	r := newTestPluginRegistry(t)
	require.NoError(t, r.Close(context.Background()))

	require.NoError(t, r.syncPools(kit.New(), map[string]bool{"tcloud": true}))

	_, err := r.GetPool("tcloud")
	assert.ErrorIs(t, err, ErrPluginNotEnabled)
}

// TestNewPluginRegistryRejectsInvalidConfig 池参数不合法时创建注册表就失败，服务据此拒绝启动，而不是等到建池时才发现。
func TestNewPluginRegistryRejectsInvalidConfig(t *testing.T) {
	cli := &dataservice.Client{}
	valid := PluginRegistryOptions{Pool: PoolConfig{MaxProcesses: 4, MaxQueueLength: 10}, CheckInterval: time.Second}

	_, err := NewPluginRegistry(cli, nil, valid)
	assert.NoError(t, err)

	tooManyIdle := valid
	tooManyIdle.Pool.MinIdleProcesses = 5
	_, err = NewPluginRegistry(cli, nil, tooManyIdle)
	assert.Error(t, err, "minIdleProcesses above maxProcesses")

	noInterval := valid
	noInterval.CheckInterval = 0
	_, err = NewPluginRegistry(cli, nil, noInterval)
	assert.Error(t, err)

	_, err = NewPluginRegistry(nil, nil, valid)
	assert.Error(t, err, "data service client is required")
}
