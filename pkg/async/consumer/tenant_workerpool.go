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
	"fmt"
	"sync"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
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

// executeWithTenant 从租户表拿到所有状态为enable的租户id并分发到chan，由工作协程池中的协程消费处理
func (wp *tenantWorkerPool) executeWithTenant() error {
	tenantIDs, err := wp.listTenantIDs()
	if err != nil {
		logs.Errorf("tenantWorkerPool failed to list tenants, err: %v", err)
		return err
	}

	for _, tenantID := range tenantIDs {
		wp.submit(tenantID)
	}
	return nil
}

// listTenantIDs 获取所有租户ID
func (wp *tenantWorkerPool) listTenantIDs() ([]string, error) {
	kt := NewKit()
	tenantIDs := make([]string, 0)
	page := core.NewDefaultBasePage()
	for {
		result, err := actcli.GetDataService().Global.Tenant.List(kt, &core.ListReq{
			Page:   page,
			Fields: []string{"tenant_id"},
			Filter: tools.EqualExpression("status", "enable"),
		})
		if err != nil {
			logs.Errorf("list tenant failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list tenant failed, err: %v", err)
		}

		for _, t := range result.Details {
			tenantIDs = append(tenantIDs, t.TenantID)
		}

		// 如果当前页数据不足一页，说明后面没有更多数据了
		if uint(len(result.Details)) < page.Limit {
			break
		}
		page.Start += uint32(page.Limit)
	}
	return tenantIDs, nil
}

// submit 将租户ID提交到任务通道中，由工作协程池中的协程处理
func (wp *tenantWorkerPool) submit(tenantID string) {
	wp.taskChan <- tenantID
}

// shutdownPoolGracefully 关闭协程池，封装内部关闭逻辑，并等待所有协程退出
func (wp *tenantWorkerPool) shutdownPoolGracefully() {
	close(wp.taskChan)
	wp.wg.Wait()
}
