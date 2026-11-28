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

// Package check 定义 IaC 插件运行期周期检查的统一接口、调度循环以及各协议实现。
// 不同协议（tf_provider、http）各自实现单轮检查逻辑，周期调度、后台 kit 构造与上下文退出由 Run 统一管理。
package check

import (
	"context"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Checker IaC 插件运行期周期检查接口，不同协议插件提供各自实现。
type Checker interface {
	// Name 返回检查器标识（如 tf_provider / http），用于日志区分。
	Name() string
	// Interval 返回检查周期。
	Interval() time.Duration
	// Check 执行一轮检查。返回 true 表示任务应终止调度（例如已触发进程退出）。
	Check(kt *kit.Kit) (stop bool)
}

// Run 按 Checker.Interval() 周期执行 Check，直到 ctx 取消或 Check 返回 true。
func Run(ctx context.Context, c Checker) {
	if c == nil {
		return
	}

	interval := c.Interval()
	if interval <= 0 {
		logs.Warnf("skip iac plugin checker %s with non-positive interval: %s", c.Name(), interval)
		return
	}

	logs.Infof("start iac plugin checker %s, interval: %s", c.Name(), interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logs.Infof("iac plugin checker %s stopped", c.Name())
			return
		case <-ticker.C:
			kt := core.NewBackendKit()
			kt.Ctx = ctx
			if c.Check(kt) {
				logs.Infof("iac plugin checker %s finished", c.Name())
				return
			}
		}
	}
}
