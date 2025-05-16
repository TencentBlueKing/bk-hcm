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

import "sync"

// workerPool 协程池
type workerPool struct {
	workerNum uint
	taskChan  chan string
	wg        sync.WaitGroup
}

// newWorkerPool 创建协程池
func newWorkerPool(workers uint) *workerPool {
	return &workerPool{
		workerNum: workers,
		taskChan:  make(chan string, workers),
	}
}

// run 启动工作协程，且传入的工作函数应当能够接收租户id
func (wp *workerPool) run(workerFunc func(tenantID string)) {
	for i := 0; i < int(wp.workerNum); i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for tenantID := range wp.taskChan {
				workerFunc(tenantID)
			}
		}()
	}
}

// submit 将租户ID提交到任务通道中，由工作协程池中的协程处理
func (wp *workerPool) submit(tenantID string) {
	wp.taskChan <- tenantID
}

// closeAndWait 关闭协程池，封装内部关闭逻辑，并等待所有协程退出
func (wp *workerPool) closeAndWait() {
	close(wp.taskChan)
	wp.wg.Wait()
}
