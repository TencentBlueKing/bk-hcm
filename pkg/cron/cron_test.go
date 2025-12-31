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
	"testing"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/cron/core"
	"hcm/pkg/kit"
	"hcm/pkg/metrics"
)

type testTask struct {
	nextTime time.Time
}

// NewTask create a new task
func newTask() core.Task {
	return &testTask{
		nextTime: time.Now(),
	}
}

// Name return the name of the task.
func (t *testTask) Name() string {
	return "test task"
}

// Next return the next time to run the task.
func (t *testTask) Next() (time.Time, error) {
	lastTime := t.nextTime
	t.nextTime = lastTime.Add(time.Second)
	return lastTime, nil
}

// Do execute the task.
func (t *testTask) Do(kt *kit.Kit) error {
	fmt.Printf("exec test task, now: %v\n", time.Now())
	return nil
}

// GetURL get the url of the task, require every task to have external api in service.
func (t *testTask) GetURL() string {
	return "/test/task"
}

// TestCron test cron
func TestCron(t *testing.T) {
	cc.InitRuntime(cc.TestSetting{})
	if err := Init(context.Background(), metrics.Register()); err != nil {
		t.Fatal(err)
	}
	if err := Register([]core.Task{newTask()}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 5)

	if err := Stop(); err != nil {
		t.Fatal(err)
	}
}
