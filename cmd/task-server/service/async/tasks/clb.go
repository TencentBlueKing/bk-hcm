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

package tasks

import (
	"fmt"
	"sync"

	"hcm/pkg/async"
)

var clbOnce sync.Once

type clb struct {
}

func init() {
	clbOnce.Do(func() {
		clb := new(clb)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "create_clb", clb.CreateClb)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "undo_create_clb", clb.UndoCreateClb)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "before_check_clb", clb.BeforeCheckClb)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "check_clb", clb.CheckClb)
	})
}

var clbSyncMap sync.Map

// CreateClb test
func (clb *clb) CreateClb(key string) (string, error) {
	// TODO: 实际是调用hc-service等接口

	clbSyncMap.LoadOrStore(key, "27")

	return key, fmt.Errorf("zyx-test error")
}

// UndoCreateClb test
func (clb *clb) UndoCreateClb(key string) (string, error) {
	// TODO: 实际是调用hc-service等接口

	clbSyncMap.Delete(key)

	_, ok := clbSyncMap.LoadOrStore(key, "27")

	fmt.Println("#######", ok)

	return key, nil
}

// BeforeCheckClb test
func (clb *clb) BeforeCheckClb(key string) (string, error) {
	// TODO: 实际是调用hc-service等接口

	clbSyncMap.LoadOrStore(key, "27")

	return key, nil
}

// CheckClb test
func (clb *clb) CheckClb(key string) (bool, error) {
	// TODO: 实际是调用hc-service等接口

	_, ok := clbSyncMap.LoadOrStore(key, "27")

	return ok, nil
}
