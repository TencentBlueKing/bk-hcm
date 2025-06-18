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

package consumer

import (
	"hcm/pkg/logs"
	"sync"
)

// tenantWorkerPool 租户协程池，用于并发处理租户ID的任务。每个协程从chan中获取租户ID并执行指定的消费租户ID的工作函数
type tenantWorkerPool struct {
	workerNum  uint
	taskChan   chan string
	workerFunc func(tenantID string)
	wg         sync.WaitGroup
}

// newTenantWorkerPool 创建协程池并立即启动workerNum个工作协程，workerFunc是工作协程的执行函数，要求能够接收租户id
func newTenantWorkerPool(workerNum uint, workerFunc func(tenantID string)) *tenantWorkerPool {
	pool := &tenantWorkerPool{
		workerNum:  workerNum,
		taskChan:   make(chan string, workerNum),
		workerFunc: workerFunc,
	}

	for i := 0; i < int(pool.workerNum); i++ {
		pool.wg.Add(1)
		go pool.worker()
	}
	return pool
}

// worker 工作协程的执行逻辑，启动后会阻塞在chan上等待租户id传入
func (wp *tenantWorkerPool) worker() {
	for tenantID := range wp.taskChan {
		wp.workerFunc(tenantID)
	}
	wp.wg.Done()
}

// submit 将租户ID提交到任务通道中，由工作协程池中的协程处理
func (wp *tenantWorkerPool) submit(tenantID string) {
	wp.taskChan <- tenantID
}

// feedTenantID 从租户表拿到所有状态为enable的租户id并分发到chan，由工作协程池中的协程消费处理
func (wp *tenantWorkerPool) feedTenantID() error {
	tenantIDs, err := listTenantIDs()
	if err != nil {
		logs.Errorf("tenantWorkerPool failed to list tenants, err: %v", err)
		return err
	}

	for _, tenantID := range tenantIDs {
		wp.submit(tenantID)
	}
	return nil
}

// shutdownPoolGracefully 关闭协程池，封装内部关闭逻辑，并等待所有协程退出
func (wp *tenantWorkerPool) shutdownPoolGracefully() {
	close(wp.taskChan)
	wp.wg.Wait()
}
