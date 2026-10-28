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

// Package watch 运行期检查本副本的 tf 插件二进制是否已落后于注册表。
//
// 二进制部署没有滚动替换语义：某个副本被重启并注册了新版本后，仍运行旧二进制的副本不会自己发现。
// 本包按周期读取注册表，本副本版本低于已完成最高版本并连续若干个周期成立时通知退出。
// 只读注册表：不加锁、不写库、不写审计，不会干扰正在进行的注册或升级。
package watch

import (
	"context"
	"time"

	tfdiscovery "hcm/cmd/hc-service/logics/plugin/tf/discovery"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iacplugin"
	"hcm/pkg/logs"
)

const (
	// DefaultInterval 默认检查周期。
	DefaultInterval = 60 * time.Second
	// DefaultThreshold 连续多少个周期落后才退出。
	// 取 3 而不是 1：新版本刚注册完成时，旧副本本来就在被替换的窗口里，不应立刻自杀。
	DefaultThreshold = 3
)

// Options Watcher 的运行参数。
type Options struct {
	// Interval 检查周期，<=0 时取 DefaultInterval。
	Interval time.Duration
	// Threshold 连续落后多少个周期后退出，<=0 时取 DefaultThreshold。
	Threshold int
}

// Watcher 周期性检查本副本版本是否落后。
type Watcher struct {
	cli      *global.IaCPluginClient
	entries  map[string]tfdiscovery.PluginEntry
	onBehind func()
	opts     Options

	// counts 每个插件当前连续落后的周期数，只在 Run 的单个 goroutine 里访问。
	counts map[string]int
}

// New 创建 Watcher。entries 是启动时目录检查的结果（只读）；onBehind 用于通知退出。
func New(cli *global.IaCPluginClient, entries map[string]tfdiscovery.PluginEntry,
	onBehind func(), opts Options) *Watcher {

	if opts.Interval <= 0 {
		opts.Interval = DefaultInterval
	}
	if opts.Threshold <= 0 {
		opts.Threshold = DefaultThreshold
	}

	return &Watcher{
		cli:      cli,
		entries:  entries,
		onBehind: onBehind,
		opts:     opts,
		counts:   make(map[string]int, len(entries)),
	}
}

// Run 阻塞执行检查，直到 ctx 取消或已触发退出。没有任何本地插件时立即返回。
func (w *Watcher) Run(ctx context.Context) {
	if len(w.entries) == 0 {
		return
	}

	ticker := time.NewTicker(w.opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.check(ctx) {
				return
			}
		}
	}
}

// check 执行一轮检查，返回是否已触发退出。
func (w *Watcher) check(ctx context.Context) bool {
	// 后端任务使用后端操作 kit：data-service 会校验请求头里的操作人与 AppCode。
	kt := core.NewBackendKit()
	kt.Ctx = ctx

	for _, entry := range w.entries {
		if ctx.Err() != nil {
			return false
		}

		res, err := w.cli.List(kt, &core.ListReq{
			Filter: tools.EqualExpression("plugin_name", entry.PluginName),
			Page:   core.NewDefaultBasePage(),
		})
		if err != nil {
			// 读取失败只跳过本轮，不计数：把 data-service 的抖动当成"版本落后"会误杀健康副本。
			logs.Warnf("list plugin %s records failed, skip version check this round, err: %v, rid: %s",
				entry.PluginName, err, kt.Rid)
			continue
		}

		behind, registered, err := isPluginVersionBehind(res.Details, entry.PluginVersion)
		if err != nil {
			logs.Warnf("check plugin %s version behind failed, skip this round, err: %v, rid: %s",
				entry.PluginName, err, kt.Rid)
			continue
		}

		if !behind {
			w.counts[entry.PluginName] = 0
			continue
		}

		w.counts[entry.PluginName]++
		logs.Warnf("check plugin %s version behind, local: %s, registered: %s, consecutive: %d/%d, rid: %s",
			entry.PluginName, entry.PluginVersion, registered, w.counts[entry.PluginName], w.opts.Threshold, kt.Rid)

		if w.counts[entry.PluginName] >= w.opts.Threshold {
			logs.Errorf("check plugin %s version stays behind, local: %s, registered: %s, exit, rid: %s",
				entry.PluginName, entry.PluginVersion, registered, kt.Rid)
			if w.onBehind != nil {
				w.onBehind()
			}
			return true
		}
	}

	return false
}

// isPluginVersionBehind 判断本地版本是否低于注册表中已完成（enabled / disabled）的最高版本。
//
// 只统计已完成的记录：init / failed 的更高版本说明升级还没成功，这时旧副本仍是当前唯一可用的版本，不能算落后。
// 本地版本高于已完成最高版本（自己的新版本还没注册完）同样不算落后。
func isPluginVersionBehind(records []coreplugin.IaCPlugin, localVersion string) (behind bool, registered string, err error) {
	filteredRecords := make([]coreplugin.IaCPlugin, 0, len(records))
	for _, rec := range records {
		if rec.State == enumor.IaCPluginStateEnabled || rec.State == enumor.IaCPluginStateDisabled {
			filteredRecords = append(filteredRecords, rec)
		}
	}

	highest, err := iacplugin.Latest(filteredRecords)
	if err != nil {
		return false, "", err
	}
	if highest == nil {
		return false, "", nil
	}

	cmp, err := iacplugin.CompareVersion(localVersion, highest.PluginVersion)
	if err != nil {
		return false, "", err
	}

	return cmp < 0, highest.PluginVersion, nil
}
