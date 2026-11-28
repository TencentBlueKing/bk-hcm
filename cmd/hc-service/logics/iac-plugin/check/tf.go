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

package check

import (
	"time"

	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

var _ Checker = (*TFChecker)(nil)

const (
	// DefaultTFInterval 默认 tf 插件版本检查周期。
	DefaultTFInterval = 60 * time.Second
	// DefaultTFThreshold 连续多少个周期落后才退出。
	// 取 3 而不是 1：新版本刚注册完成时，旧副本本来就在被替换的窗口里，不应立刻自杀。
	DefaultTFThreshold = 3
)

// TFOptions tf 插件版本检查器的运行参数。
type TFOptions struct {
	// Interval 检查周期，<=0 时取 DefaultTFInterval。
	Interval time.Duration
	// Threshold 连续落后多少个周期后退出，<=0 时取 DefaultTFThreshold。
	Threshold int
}

// TFChecker 周期性检查本地 tf 插件版本是否落后于注册表，持续落后时通知进程退出。
type TFChecker struct {
	cli      *global.IaCPluginClient
	entries  map[string]discovery.PluginEntry
	onBehind func()
	opts     TFOptions

	// counts 每个插件当前连续落后的周期数
	counts map[string]int
}

// NewTFChecker 创建 tf 插件版本检查器。entries 是启动时目录扫描的结果（只读）；onBehind 用于通知退出。
func NewTFChecker(cli *global.IaCPluginClient, entries map[string]discovery.PluginEntry,
	onBehind func(), opts TFOptions) *TFChecker {

	if opts.Interval <= 0 {
		opts.Interval = DefaultTFInterval
	}
	if opts.Threshold <= 0 {
		opts.Threshold = DefaultTFThreshold
	}

	return &TFChecker{
		cli:      cli,
		entries:  entries,
		onBehind: onBehind,
		opts:     opts,
		counts:   make(map[string]int, len(entries)),
	}
}

// Name 返回检查器标识。
func (c *TFChecker) Name() string {
	return string(enumor.IaCPluginProtocolTFProvider)
}

// Interval 返回版本检查周期。
func (c *TFChecker) Interval() time.Duration {
	return c.opts.Interval
}

// Check 执行一轮检查，返回是否已触发退出。没有任何本地插件时直接返回 true 终止调度。
func (c *TFChecker) Check(kt *kit.Kit) bool {
	if len(c.entries) == 0 {
		return true
	}

	for _, entry := range c.entries {
		if kt.Ctx.Err() != nil {
			return false
		}

		behind, registered, err := c.isPluginVersionBehind(kt, entry)
		if err != nil {
			// 读取或校验失败只跳过本轮，不计数：避免因 data-service 抖动误杀健康实例。
			logs.Warnf("check plugin %s version behind failed, skip this round, err: %v, rid: %s",
				entry.PluginName, err, kt.Rid)
			continue
		}

		if !behind {
			c.counts[entry.PluginName] = 0
			continue
		}

		c.counts[entry.PluginName]++
		logs.Warnf("check plugin %s version behind, local: %s, registered: %s, consecutive: %d/%d, rid: %s",
			entry.PluginName, entry.PluginVersion, registered, c.counts[entry.PluginName], c.opts.Threshold, kt.Rid)

		if c.counts[entry.PluginName] >= c.opts.Threshold {
			logs.Errorf("check plugin %s version stays behind, local: %s, registered: %s, exit, rid: %s",
				entry.PluginName, entry.PluginVersion, registered, kt.Rid)
			if c.onBehind != nil {
				c.onBehind()
			}
			return true
		}
	}

	return false
}

// isPluginVersionBehind 查询注册表并判断本地版本是否落后于当前最高版本。
// 仅当已完成记录（enabled / disabled）中的最高版本处于 enabled 且高于本地版本时返回 true。
func (c *TFChecker) isPluginVersionBehind(kt *kit.Kit,
	entry discovery.PluginEntry) (behind bool, registered string, err error) {

	highest, err := c.cli.GetLatest(kt, entry.PluginName,
		tools.ContainersExpression("state", []enumor.IaCPluginState{
			enumor.IaCPluginStateEnabled,
			enumor.IaCPluginStateDisabled,
		}))
	if err != nil {
		return false, "", err
	}
	if highest == nil || highest.State != enumor.IaCPluginStateEnabled {
		return false, "", nil
	}

	cmp, err := iacplugin.CompareVersion(entry.PluginVersion, highest.PluginVersion)
	if err != nil {
		return false, "", err
	}

	return cmp < 0, highest.PluginVersion, nil
}
