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
	if w.size == w.capacity {
		w.head = (w.head + 1) % w.capacity // 覆盖最旧元素
	} else {
		w.size++
	}
	w.queue[w.tail] = taskTypeExecTime{execTime: execTime, entryTime: time.Now()}
	w.tail = (w.tail + 1) % w.capacity
}

// GetAvg 计算时间窗口内的平均执行时间
func (w *TimeWindow) GetAvg() (avgExecTime float64, neverExec bool) {
	w.Lock()
	defer w.Unlock()
	// 该任务类型在整个服务生命周期内从未被执行。一旦执行过后队列中将始终保留至少一条执行时间的记录
	if w.size == 0 {
		return 0, true
	}

	// 计算时间窗口内执行时间
	var insum float64
	// 计算时间窗口外执行时间
	var outsum float64
	var inCount uint
	var outCount uint
	now := time.Now()

	// 先遍历统计数据，因为是先进先出的队列，所以只有三种情况：
	// 1、队头<-时间窗口外的数据，时间窗口内的数据<-队尾
	// 2、队头<-时间窗口外的数据<-队尾
	// 3、队头<-时间窗口内的数据<-队尾
	for i := uint(0); i < w.size; i++ {
		idx := (w.head + i) % w.capacity
		task := w.queue[idx]
		if now.Sub(task.entryTime) <= w.duration {
			insum += task.execTime
			inCount++
		} else {
			outsum += task.execTime
			outCount++
		}
	}

	// 如果存在时间窗口内的数据，则返回前清理时间窗口外的数据
	if inCount > 0 {
		w.head = (w.head + outCount) % w.capacity
		w.size = inCount
		return insum / float64(inCount), false
	}
	// 否则直接返回时间窗口外的数据并且保留
	return outsum / float64(outCount), false
}
