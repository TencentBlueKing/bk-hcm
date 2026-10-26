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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	defaultMaxLeasesPerProcess = 4
	defaultMinIdleProcesses    = 2
	defaultMaxRetries          = 2
	defaultMaxQueueWait        = time.Minute
	defaultConfigureTimeout    = 30 * time.Second
	defaultIdleReclaimTimeout  = 10 * time.Minute
	defaultMaxConfigAge        = 30 * time.Minute
	defaultPoolCheckInterval   = 30 * time.Second
)

// PoolConfig 进程池参数。MaxProcesses 与 MaxQueueLength 必填，其余为 0 时取默认值。
type PoolConfig struct {
	// MaxProcesses 池内进程数上限，starting / configuring / running / closing 都计入，确认进程退出后才释放。
	MaxProcesses int
	// MaxLeasesPerProcess 单个进程同时交付的 lease（含准备期预约）上限，默认 4。
	MaxLeasesPerProcess int
	// MinIdleProcesses 保持的最少空闲进程数，默认 2，不能超过 MaxProcesses。
	MinIdleProcesses int
	// MaxRetries 进程侧失败（启动失败、配置超时或传输错误、准备期进程退出）后请求的最大重试次数，默认 2。
	MaxRetries int
	// MaxQueueLength 等待队列中有效请求数的上限，达到后新请求返回 ErrTooManyWaiters。
	MaxQueueLength int
	// MaxQueueWait 单个请求的最长排队时间，默认 1 分钟，同时受调用方 kit.Ctx 约束。
	MaxQueueWait time.Duration
	// ConfigureTimeout 一次 ConfigureProvider 的超时，默认 30 秒。超时的进程状态不可信，会被关闭。
	ConfigureTimeout time.Duration
	// IdleReclaimTimeout 空闲超过该时长的进程可被回收（回收后空闲数不低于 MinIdleProcesses），默认 10 分钟。
	IdleReclaimTimeout time.Duration
	// MaxConfigAge 一次配置可被复用的时长，默认 30 分钟，须短于云凭证有效期。过期只阻止新的交付，不影响已持有的 lease。
	MaxConfigAge time.Duration
	// PoolCheckInterval 池内巡检周期，默认 30 秒。
	PoolCheckInterval time.Duration
}

func (c *PoolConfig) trySetDefault() {
	if c.MaxLeasesPerProcess == 0 {
		c.MaxLeasesPerProcess = defaultMaxLeasesPerProcess
	}
	if c.MinIdleProcesses == 0 {
		c.MinIdleProcesses = defaultMinIdleProcesses
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = defaultMaxRetries
	}
	if c.MaxQueueWait == 0 {
		c.MaxQueueWait = defaultMaxQueueWait
	}
	if c.ConfigureTimeout == 0 {
		c.ConfigureTimeout = defaultConfigureTimeout
	}
	if c.IdleReclaimTimeout == 0 {
		c.IdleReclaimTimeout = defaultIdleReclaimTimeout
	}
	if c.MaxConfigAge == 0 {
		c.MaxConfigAge = defaultMaxConfigAge
	}
	if c.PoolCheckInterval == 0 {
		c.PoolCheckInterval = defaultPoolCheckInterval
	}
}

// Validate 校验参数，应在 trySetDefault 之后调用。
func (c PoolConfig) Validate() error {
	if c.MaxProcesses <= 0 {
		return errors.New("maxProcesses should be greater than 0")
	}
	if c.MaxLeasesPerProcess <= 0 {
		return errors.New("maxLeasesPerProcess should be greater than 0")
	}
	if c.MinIdleProcesses < 0 || c.MinIdleProcesses > c.MaxProcesses {
		return errors.New("minIdleProcesses should be between 0 and maxProcesses")
	}
	if c.MaxRetries < 0 {
		return errors.New("maxRetries should not be negative")
	}
	if c.MaxQueueLength <= 0 {
		return errors.New("maxQueueLength should be greater than 0")
	}
	if c.MaxQueueWait <= 0 || c.ConfigureTimeout <= 0 || c.IdleReclaimTimeout <= 0 || c.MaxConfigAge <= 0 ||
		c.PoolCheckInterval <= 0 {
		return errors.New("timeouts and intervals should be greater than 0")
	}
	return nil
}

// configKey 计算 Provider 配置的匹配键：JSON 编码后取 SHA-256。
//
// encoding/json 按字典序输出 map 的键，所以结果与键的插入顺序无关。规范化口径（方案待定 #2）只在这里改。
// 配置含凭证，返回的错误不得包含配置内容。
func configKey(config map[string]any) (string, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("provider config cannot be encoded: %T", err)
	}

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
