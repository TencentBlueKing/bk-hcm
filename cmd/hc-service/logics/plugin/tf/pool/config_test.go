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
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoolConfigDefaultAndValidate(t *testing.T) {
	cfg := PoolConfig{MaxProcesses: 2, MaxQueueLength: 10}
	cfg.trySetDefault()
	require.NoError(t, cfg.Validate())
	assert.Equal(t, defaultMaxLeasesPerProcess, cfg.MaxLeasesPerProcess)
	assert.Equal(t, defaultMaxQueueWait, cfg.MaxQueueWait)
	assert.Equal(t, defaultMaxConfigAge, cfg.MaxConfigAge)

	cases := []struct {
		name string
		edit func(c *PoolConfig)
	}{
		{"no max processes", func(c *PoolConfig) { c.MaxProcesses = 0 }},
		{"no queue length", func(c *PoolConfig) { c.MaxQueueLength = 0 }},
		{"negative leases", func(c *PoolConfig) { c.MaxLeasesPerProcess = -1 }},
		{"negative wait", func(c *PoolConfig) { c.MaxQueueWait = -time.Second }},
		{"negative start timeout", func(c *PoolConfig) { c.StartTimeout = -time.Second }},
		{"negative configure timeout", func(c *PoolConfig) { c.ConfigureTimeout = -time.Second }},
		{"negative close timeout", func(c *PoolConfig) { c.CloseTimeout = -time.Second }},
		{"negative config age", func(c *PoolConfig) { c.MaxConfigAge = -time.Second }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bad := PoolConfig{MaxProcesses: 2, MaxQueueLength: 10}
			c.edit(&bad)
			bad.trySetDefault()
			assert.Error(t, bad.Validate())
		})
	}
}

func TestConfigKey(t *testing.T) {
	a := map[string]any{"secret_id": "id", "region": "ap-guangzhou",
		"assume_role": map[string]any{"role_arn": "arn", "session_name": "s"}}
	b := map[string]any{"assume_role": map[string]any{"session_name": "s", "role_arn": "arn"},
		"region": "ap-guangzhou", "secret_id": "id"}

	ka, err := configKey(a)
	require.NoError(t, err)
	kb, err := configKey(b)
	require.NoError(t, err)
	assert.Equal(t, ka, kb, "key must not depend on map insertion order")
	assert.Len(t, ka, 64)

	kc, err := configKey(map[string]any{"secret_id": "other", "region": "ap-guangzhou"})
	require.NoError(t, err)
	assert.NotEqual(t, ka, kc)

	kNil, err := configKey(nil)
	require.NoError(t, err)
	kEmpty, err := configKey(map[string]any{})
	require.NoError(t, err)
	assert.NotEqual(t, kNil, kEmpty)

	_, err = configKey(map[string]any{"secret_key": "s3cr3t", "bad": math.NaN()})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cr3t")
}
