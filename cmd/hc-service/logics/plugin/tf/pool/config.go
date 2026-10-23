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
	defaultMaxLeasesPerProcess = 1
	defaultMaxQueueWait        = time.Minute
	defaultStartTimeout        = 30 * time.Second
	defaultConfigureTimeout    = 30 * time.Second
	defaultCloseTimeout        = 30 * time.Second
	defaultMaxConfigAge        = 30 * time.Minute
)

// PoolConfig 进程池参数。MaxProcesses 与 MaxQueueLength 必填，其余为 0 时取默认值。
type PoolConfig struct {
	// MaxProcesses 池内进程数上限。starting / configuring / running / closing 的进程都计入，
	// 只有确认进程退出后才释放名额。
	MaxProcesses int
	// MaxLeasesPerProcess 单个进程同时交付的 lease 上限，默认 1。
	// provider 是否支持并发 RPC 确认之前不要调大。
	MaxLeasesPerProcess int
	// MaxQueueLength 等待队列中有效请求数的上限，达到后新请求直接返回 ErrTooManyWaiters。
	MaxQueueLength int
	// MaxQueueWait 单个请求的最长排队时间，默认 1 分钟；同时受调用方 kit.Ctx 约束。
	MaxQueueWait time.Duration
	// StartTimeout 单次拉起进程（含握手）的超时，默认 30 秒。
	StartTimeout time.Duration
	// ConfigureTimeout 单次 ConfigureProvider 的超时，默认 30 秒。超时的进程状态不可信，会被关闭。
	ConfigureTimeout time.Duration
	// CloseTimeout 池内部关闭进程（失败、回收）时等待进程退出的时限，默认 30 秒，超时后强制结束。
	// PluginPool.Close 关闭进程时使用调用方传入的 ctx，不受此项约束。
	CloseTimeout time.Duration
	// MaxConfigAge 一次配置可被复用的时长，默认 30 分钟，须短于云凭证有效期。
	// 过期只阻止新的交付，不影响已持有的 lease。
	MaxConfigAge time.Duration
}

func (c *PoolConfig) trySetDefault() {
	if c.MaxLeasesPerProcess == 0 {
		c.MaxLeasesPerProcess = defaultMaxLeasesPerProcess
	}
	if c.MaxQueueWait == 0 {
		c.MaxQueueWait = defaultMaxQueueWait
	}
	if c.StartTimeout == 0 {
		c.StartTimeout = defaultStartTimeout
	}
	if c.ConfigureTimeout == 0 {
		c.ConfigureTimeout = defaultConfigureTimeout
	}
	if c.CloseTimeout == 0 {
		c.CloseTimeout = defaultCloseTimeout
	}
	if c.MaxConfigAge == 0 {
		c.MaxConfigAge = defaultMaxConfigAge
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
	if c.MaxQueueLength <= 0 {
		return errors.New("maxQueueLength should be greater than 0")
	}

	durations := []struct {
		name  string
		value time.Duration
	}{
		{"maxQueueWait", c.MaxQueueWait},
		{"startTimeout", c.StartTimeout},
		{"configureTimeout", c.ConfigureTimeout},
		{"closeTimeout", c.CloseTimeout},
		{"maxConfigAge", c.MaxConfigAge},
	}
	for _, d := range durations {
		if d.value <= 0 {
			return fmt.Errorf("%s should be greater than 0", d.name)
		}
	}

	return nil
}

// configKey 计算插件实例 properties（Provider 配置）的匹配键：规范化编码后取 SHA-256。
//
// 规范化口径：encoding/json 编码，map 的键按字典序输出，因此与键的插入顺序无关。
// 口径定稿前（方案待定 #2）只在这一处修改。配置含凭证，返回的错误不得包含配置内容。
func configKey(config map[string]any) (string, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("provider config cannot be encoded: %T", err)
	}

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
