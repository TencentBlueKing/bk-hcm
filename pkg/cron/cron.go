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

// Package cron ...
package cron

import (
	"context"
	"fmt"
	"sync"

	"hcm/pkg/cron/core"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cron core.Scheduler
	mu   sync.RWMutex
)

// Init init cron scheduler
func Init(ctx context.Context, reg prometheus.Registerer) error {
	mu.Lock()
	defer mu.Unlock()

	if cron != nil {
		logs.Warnf("cron scheduler already initialized, skipping")
		return nil
	}

	var err error
	cron, err = core.NewScheduler(ctx, reg)
	if err != nil {
		logs.Errorf("new cron scheduler failed, err: %v", err)
		return err
	}
	if err = cron.Start(); err != nil {
		logs.Errorf("start cron scheduler failed, err: %v", err)
		cron = nil
		return err
	}

	return nil
}

// Register register tasks.
func Register(tasks []core.Task) error {
	mu.RLock()
	defer mu.RUnlock()

	if cron == nil {
		logs.Errorf("cron scheduler not initialized, cannot register tasks")
		return fmt.Errorf("cron scheduler not initialized, cannot register tasks")
	}

	if err := cron.Register(tasks); err != nil {
		logs.Errorf("register tasks failed, err: %v", err)
		return err
	}

	return nil
}

// Stop stop cron scheduler.
func Stop() error {
	mu.Lock()
	defer mu.Unlock()

	if cron == nil {
		logs.Warnf("cron scheduler not initialized, cannot stop")
		return nil
	}

	if err := cron.Stop(); err != nil {
		logs.Errorf("stop cron scheduler failed, err: %v", err)
		return err
	}
	cron = nil // 停止后清理，允许重新初始化
	return nil
}
