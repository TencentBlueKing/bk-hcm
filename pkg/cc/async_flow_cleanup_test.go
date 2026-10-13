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

package cc

import (
	"strings"
	"testing"

	"hcm/pkg/criteria/constant"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// TestAsyncFlowAndTaskCleanupTrySetDefault 缺省配置经 trySetDefault 后取默认值。
func TestAsyncFlowAndTaskCleanupTrySetDefault(t *testing.T) {
	cfg := new(AsyncFlowAndTaskCleanup)
	cfg.trySetDefault()

	assert.True(t, cvt.PtrToVal(cfg.Enabled))
	assert.Equal(t, constant.DefaultAsyncFlowCleanupIntervalMin, cvt.PtrToVal(cfg.IntervalMin))
	assert.Equal(t, constant.DefaultAsyncFlowCleanupRetentionDays, cvt.PtrToVal(cfg.RetentionDays))
	assert.Equal(t, constant.DefaultAsyncFlowCleanupBatchIntervalMs, cvt.PtrToVal(cfg.BatchIntervalMs))
}

// TestAsyncFlowAndTaskCleanupTrySetDefaultKeepUserValue 用户已配置的值不被默认值覆盖，非法值同样原样保留。
func TestAsyncFlowAndTaskCleanupTrySetDefaultKeepUserValue(t *testing.T) {
	cfg := &AsyncFlowAndTaskCleanup{
		Enabled:         cvt.ValToPtr(false),
		IntervalMin:     cvt.ValToPtr(30),
		RetentionDays:   cvt.ValToPtr(90),
		BatchIntervalMs: cvt.ValToPtr(300),
	}
	cfg.trySetDefault()

	assert.False(t, cvt.PtrToVal(cfg.Enabled))
	assert.Equal(t, 30, cvt.PtrToVal(cfg.IntervalMin))
	assert.Equal(t, 90, cvt.PtrToVal(cfg.RetentionDays))
	assert.Equal(t, 300, cvt.PtrToVal(cfg.BatchIntervalMs))

	// 显式配置的非法值必须活着走到 validate，否则下界校验永远不可达
	illegal := &AsyncFlowAndTaskCleanup{
		IntervalMin:     cvt.ValToPtr(0),
		RetentionDays:   cvt.ValToPtr(-1),
		BatchIntervalMs: cvt.ValToPtr(0),
	}
	illegal.trySetDefault()

	assert.Equal(t, 0, cvt.PtrToVal(illegal.IntervalMin))
	assert.Equal(t, -1, cvt.PtrToVal(illegal.RetentionDays))
	assert.Equal(t, 0, cvt.PtrToVal(illegal.BatchIntervalMs))
}

// TestAsyncFlowAndTaskCleanupValidate 开关打开时的非法配置逐项校验。
func TestAsyncFlowAndTaskCleanupValidate(t *testing.T) {
	newValidCfg := func() AsyncFlowAndTaskCleanup {
		return AsyncFlowAndTaskCleanup{
			Enabled:         cvt.ValToPtr(true),
			IntervalMin:     cvt.ValToPtr(constant.DefaultAsyncFlowCleanupIntervalMin),
			RetentionDays:   cvt.ValToPtr(constant.DefaultAsyncFlowCleanupRetentionDays),
			BatchIntervalMs: cvt.ValToPtr(constant.DefaultAsyncFlowCleanupBatchIntervalMs),
		}
	}

	assert.NoError(t, newValidCfg().validate())

	cases := []struct {
		name     string
		modify   func(cfg *AsyncFlowAndTaskCleanup)
		keyword  string
		hasError bool
	}{
		{"intervalMin is zero", func(c *AsyncFlowAndTaskCleanup) { c.IntervalMin = cvt.ValToPtr(0) },
			"intervalMin", true},
		{"intervalMin is negative", func(c *AsyncFlowAndTaskCleanup) { c.IntervalMin = cvt.ValToPtr(-1) },
			"intervalMin", true},
		{"retentionDays is zero", func(c *AsyncFlowAndTaskCleanup) { c.RetentionDays = cvt.ValToPtr(0) },
			"retentionDays", true},
		{"batchIntervalMs is zero", func(c *AsyncFlowAndTaskCleanup) { c.BatchIntervalMs = cvt.ValToPtr(0) },
			"batchIntervalMs", true},
		{"batchIntervalMs is negative", func(c *AsyncFlowAndTaskCleanup) { c.BatchIntervalMs = cvt.ValToPtr(-1) },
			"batchIntervalMs", true},
		// 不设下限：低于默认 100ms 的正值同样合法，调低限速由运维自行承担风险
		{"batchIntervalMs below default is legal", func(c *AsyncFlowAndTaskCleanup) {
			c.BatchIntervalMs = cvt.ValToPtr(1)
		}, "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := newValidCfg()
			c.modify(&cfg)

			err := cfg.validate()
			if !c.hasError {
				assert.NoError(t, err)
				return
			}

			if !assert.Error(t, err) {
				return
			}
			assert.True(t, strings.Contains(err.Error(), c.keyword),
				"error %q should contain field name %q", err.Error(), c.keyword)
		})
	}
}

// TestAsyncFlowAndTaskCleanupValidateDisabled 开关关闭时其余项非法也校验通过。
func TestAsyncFlowAndTaskCleanupValidateDisabled(t *testing.T) {
	cfg := AsyncFlowAndTaskCleanup{
		Enabled:         cvt.ValToPtr(false),
		IntervalMin:     cvt.ValToPtr(0),
		RetentionDays:   cvt.ValToPtr(-1),
		BatchIntervalMs: cvt.ValToPtr(0),
	}

	assert.NoError(t, cfg.validate())
}

// TestAsyncFlowAndTaskCleanupEnabledPointer 显式配置 false 与完全不配置该段，默认值处理结果不同。
func TestAsyncFlowAndTaskCleanupEnabledPointer(t *testing.T) {
	type wrapper struct {
		AsyncFlowAndTaskCleanup AsyncFlowAndTaskCleanup `yaml:"asyncFlowAndTaskCleanup"`
	}

	explicit := new(wrapper)
	err := yaml.Unmarshal([]byte("asyncFlowAndTaskCleanup:\n  enabled: false\n"), explicit)
	assert.NoError(t, err)
	explicit.AsyncFlowAndTaskCleanup.trySetDefault()
	assert.False(t, cvt.PtrToVal(explicit.AsyncFlowAndTaskCleanup.Enabled))

	absent := new(wrapper)
	err = yaml.Unmarshal([]byte("network:\n  bindIP: 127.0.0.1\n"), absent)
	assert.NoError(t, err)
	absent.AsyncFlowAndTaskCleanup.trySetDefault()
	assert.True(t, cvt.PtrToVal(absent.AsyncFlowAndTaskCleanup.Enabled))
}

// baseTaskServerYaml 是能通过 TaskServerSetting.Validate 前置各段校验的最小配置，
// 用于把断言聚焦在 asyncFlowAndTaskCleanup 段上。
const baseTaskServerYaml = `
network:
  bindIP: 127.0.0.1
service:
  etcd:
    endpoints:
      - 127.0.0.1:2379
database:
  resource:
    endpoints:
      - 127.0.0.1:3306
    database: hcm
cmdb:
  endpoints:
    - http://127.0.0.1
  appCode: hcm
  appSecret: secret
`

// TestTaskServerSettingLoadOrderRejectIllegalCleanupConfig 复现 pkg/cc/load.go 的真实加载顺序：
// 反序列化 yaml -> trySetDefault -> Validate。
// 这条路径是 AC-011「非法配置导致启动失败」的唯一保障：若 trySetDefault 把显式配置的非法值
// 静默改写成默认值，Validate 里的下界分支就永不可达，本用例会失败。
func TestTaskServerSettingLoadOrderRejectIllegalCleanupConfig(t *testing.T) {
	cases := []struct {
		name    string
		section string
		keyword string
	}{
		{"intervalMin explicit zero", "asyncFlowAndTaskCleanup:\n  intervalMin: 0\n", "intervalMin"},
		{"intervalMin explicit negative", "asyncFlowAndTaskCleanup:\n  intervalMin: -1\n", "intervalMin"},
		{"retentionDays explicit zero", "asyncFlowAndTaskCleanup:\n  retentionDays: 0\n", "retentionDays"},
		{"batchIntervalMs explicit zero", "asyncFlowAndTaskCleanup:\n  batchIntervalMs: 0\n", "batchIntervalMs"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := new(TaskServerSetting)
			assert.NoError(t, yaml.Unmarshal([]byte(baseTaskServerYaml+c.section), s))

			s.trySetDefault()

			err := s.Validate()
			if !assert.Error(t, err, "illegal %s should be rejected by Validate", c.keyword) {
				return
			}
			assert.True(t, strings.Contains(err.Error(), c.keyword),
				"error %q should contain field name %q", err.Error(), c.keyword)
		})
	}
}

// TestTaskServerSettingLoadOrderAcceptAbsentCleanupConfig 整段缺省或只配开关时，
// 走完加载顺序应当校验通过，且三个数值字段取默认值。
func TestTaskServerSettingLoadOrderAcceptAbsentCleanupConfig(t *testing.T) {
	cases := []struct {
		name    string
		section string
	}{
		{"section absent", ""},
		{"section present but empty", "asyncFlowAndTaskCleanup:\n  enabled: true\n"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := new(TaskServerSetting)
			assert.NoError(t, yaml.Unmarshal([]byte(baseTaskServerYaml+c.section), s))

			s.trySetDefault()
			assert.NoError(t, s.Validate())

			cfg := s.AsyncFlowAndTaskCleanup
			assert.True(t, cvt.PtrToVal(cfg.Enabled))
			assert.Equal(t, constant.DefaultAsyncFlowCleanupIntervalMin, cvt.PtrToVal(cfg.IntervalMin))
			assert.Equal(t, constant.DefaultAsyncFlowCleanupRetentionDays, cvt.PtrToVal(cfg.RetentionDays))
			assert.Equal(t, constant.DefaultAsyncFlowCleanupBatchIntervalMs, cvt.PtrToVal(cfg.BatchIntervalMs))
		})
	}
}

// TestTaskServerSettingLoadOrderDisabledSkipCleanupValidate 开关显式关闭时，
// 即使其余项非法也不应阻断启动。
func TestTaskServerSettingLoadOrderDisabledSkipCleanupValidate(t *testing.T) {
	s := new(TaskServerSetting)
	section := "asyncFlowAndTaskCleanup:\n  enabled: false\n  intervalMin: 0\n  batchIntervalMs: 0\n"
	assert.NoError(t, yaml.Unmarshal([]byte(baseTaskServerYaml+section), s))

	s.trySetDefault()
	assert.NoError(t, s.Validate())
	assert.False(t, cvt.PtrToVal(s.AsyncFlowAndTaskCleanup.Enabled))
}
