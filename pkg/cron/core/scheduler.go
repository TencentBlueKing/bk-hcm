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

// Package core ...
package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"hcm/pkg/api/core"
	cronmetric "hcm/pkg/cron/metric"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
)

// Scheduler defines the interface for the Scheduler.
type Scheduler interface {
	// Start starts the scheduler.
	Start() error
	// Stop stops the scheduler gracefully.
	Stop() error
	// Register registers the scheduler task.
	Register(tasks []Task) error
}

type scheduler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	taskChan  chan Task
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.RWMutex
}

// NewScheduler creates a new scheduler.
func NewScheduler(ctx context.Context, reg prometheus.Registerer) (Scheduler, error) {
	err := cronmetric.Init(reg)
	if err != nil {
		logs.Errorf("init cron metric err: %v", err)
		return nil, fmt.Errorf("failed to init cron metrics: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	return &scheduler{
		ctx:      ctx,
		cancel:   cancel,
		taskChan: make(chan Task, 100), // 添加缓冲区避免阻塞
	}, nil
}

// Start starts the scheduler.
func (s *scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return nil
	}
	s.isRunning = true

	s.wg.Add(1)
	go s.run()

	logs.Infof("scheduler started successfully")
	return nil
}

// run is the main scheduler loop
func (s *scheduler) run() {
	defer s.wg.Done()
	defer logs.Infof("scheduler main loop stopped")

	for {
		select {
		case task, ok := <-s.taskChan:
			if !ok {
				return // channel closed, exit gracefully
			}
			s.wg.Add(1)
			go s.executeTask(task)
		case <-s.ctx.Done():
			return // context cancelled, exit gracefully
		}
	}
}

// executeTask handles the execution of a single task
func (s *scheduler) executeTask(task Task) {
	defer s.wg.Done()

	taskName := task.Name()

	logs.Infof("starting task execution: %s", taskName)

	for {
		select {
		case <-s.ctx.Done():
			logs.Infof("task %s stopped due to context cancellation", taskName)
			return
		default:
			if err := s.executeSingleRun(task); err != nil {
				logs.Errorf("task %s execution failed: %v", taskName, err)
				// 添加错误后的延迟，避免快速重试导致资源浪费
				select {
				case <-time.After(time.Second):
					// 等待1秒后重试
				case <-s.ctx.Done():
					return
				}
			}
		}
	}
}

// executeSingleRun executes a single run of the task
func (s *scheduler) executeSingleRun(task Task) error {
	taskName := task.Name()
	labels := map[string]string{cronmetric.TaskName: taskName}

	nextTime, err := task.Next()
	if err != nil {
		cronmetric.ExecError().With(labels).Inc()
		logs.Errorf("get next time failed, err: %v, task: %s", err, taskName)
		return fmt.Errorf("get next time failed: %w, task: %s", err, taskName)
	}

	now := time.Now()
	if nextTime.After(now) {
		waitDuration := nextTime.Sub(now)

		select {
		case <-time.After(waitDuration):
			// Wait until next execution time
		case <-s.ctx.Done():
			return nil
		}
	}

	cronmetric.ExecCounter().With(labels).Inc()
	start := time.Now()

	kt := core.NewBackendKit().NewSubKitWithCtx(s.ctx)
	if err = task.Do(kt); err != nil {
		cronmetric.ExecError().With(labels).Inc()
		logs.Errorf("task execution failed, err: %v, task: %s, rid: %s", err, taskName, kt.Rid)
		return fmt.Errorf("task execution failed: %w, task: %s", err, taskName)
	}

	executionTime := time.Since(start)
	cronmetric.ExecDuration().With(labels).Observe(executionTime.Seconds())
	return nil
}

// Stop stops the scheduler gracefully.
func (s *scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}
	s.isRunning = false

	logs.Infof("stopping scheduler gracefully...")

	// 1. Cancel context to signal all goroutines to stop
	s.cancel()

	// 2. Close task channel to prevent new registrations
	close(s.taskChan)

	// 3. Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logs.Infof("scheduler stopped gracefully")
		return nil
	case <-time.After(30 * time.Second):
		logs.Warnf("scheduler stop timeout, forcing shutdown")
		return fmt.Errorf("scheduler stop timeout")
	}
}

// Register registers the scheduler task.
func (s *scheduler) Register(tasks []Task) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isRunning {
		logs.Errorf("scheduler is not running, cannot register tasks")
		return fmt.Errorf("scheduler is not running, cannot register tasks")
	}

	for _, task := range tasks {
		taskName := task.Name()
		select {
		case s.taskChan <- task:
			logs.Infof("task %s registered successfully", taskName)
		case <-s.ctx.Done():
			logs.Errorf("task registration failed during scheduler stop")
			return fmt.Errorf("task registration failed during scheduler stop")
		default:
			logs.Errorf("task channel is full, dropping task: %s", taskName)
			return fmt.Errorf("task channel is full, dropping task: %s", taskName)
		}
	}

	return nil
}
