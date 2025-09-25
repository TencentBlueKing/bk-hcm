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
	"sync"
	"time"

	"hcm/pkg/logs"
)

// TimeWindow 使用环形缓冲区实现固定容量队列
type TimeWindow struct {
	sync.Mutex
	queue    []taskTypeExecTime
	capacity uint
	duration time.Duration
	head     uint // 队头索引，指向下一个出队位置
	tail     uint // 队尾索引，指向下一个插入位置
	size     uint // 当前元素数量
}

type taskTypeExecTime struct {
	execTime  float64   // 本次执行耗时,单位秒
	entryTime time.Time // 入队时间戳
}

// NewTimeWindow 创建时间窗口，capacity为队列容量，duration为时间窗口大小，单位分钟
func NewTimeWindow(capacity uint, duration uint) *TimeWindow {
	return &TimeWindow{
		queue:    make([]taskTypeExecTime, capacity),
		capacity: capacity,
		duration: time.Duration(duration) * time.Minute,
	}
}

// Push 记录入队时刻，超量自动覆盖最旧
func (w *TimeWindow) Push(execTime float64) {
	w.Lock()
	defer w.Unlock()
	if execTime < 0 {
		logs.Errorf("execTime(%f) < 0", execTime)
		return
	}
	if w.size == w.capacity {
		w.head = (w.head + 1) % w.capacity // 覆盖最旧元素
	} else {
		w.size++
	}
	w.queue[w.tail] = taskTypeExecTime{execTime: execTime, entryTime: time.Now()}
	w.tail = (w.tail + 1) % w.capacity
}

// GetAvg 计算时间窗口内的平均执行时间
// 返回值：
//   - avgExecTime: 平均执行时间（秒）
//   - neverExec: 是否从未执行过任务
func (w *TimeWindow) GetAvg() (avgExecTime float64, neverExec bool) {
	w.Lock()
	defer w.Unlock()

	// 检查队列是否为空：该任务类型在整个服务生命周期内从未被执行
	// 一旦执行过后队列中将始终保留至少一条执行时间的记录
	if w.size == 0 {
		return 0, true
	}

	var inSum float64  // 时间窗口内的执行时间总和
	var outSum float64 // 时间窗口外的执行时间总和
	var inCount uint   // 时间窗口内的任务数量
	var outCount uint  // 时间窗口外的任务数量
	now := time.Now()

	// 遍历环形队列中的所有任务记录，按时间分类统计
	// 由于是FIFO队列，数据分布只有三种情况：
	// 1. 队头 <- [过期数据] <- [有效数据] <- 队尾
	// 2. 队头 <- [过期数据] <- 队尾
	// 3. 队头 <- [有效数据] <- 队尾
	for i := uint(0); i < w.size; i++ {
		idx := (w.head + i) % w.capacity
		task := w.queue[idx]

		// 判断任务记录是否在时间窗口内
		if now.Sub(task.entryTime) <= w.duration {
			// 时间窗口内的数据
			inSum += task.execTime
			inCount++
		} else {
			// 时间窗口外的过期数据
			outSum += task.execTime
			outCount++
		}
	}

	// 优先返回时间窗口内的平均值，并清理过期数据
	if inCount > 0 {
		// 移动队头指针，跳过已过期的数据
		w.head = (w.head + outCount) % w.capacity
		// 更新队列大小，只保留有效数据
		w.size = inCount
		return inSum / float64(inCount), false
	}

	// 如果时间窗口内没有数据，返回过期数据的平均值作为参考
	// 注意：这里保留过期数据不清理，作为历史参考
	return outSum / float64(outCount), false
}
