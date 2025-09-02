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
	"sort"
	"sync"
	"time"

	"hcm/pkg/logs"
)

// TaskInitQueue 基于task执行时间排序的有界阻塞优先队列
type TaskInitQueue struct {
	sync.Mutex
	queue    []*InitPayloadWithScore
	head     uint // 队头索引，指向下一个出队位置
	tail     uint // 队尾索引，指向下一个插入位置
	size     uint // 当前元素数量
	capacity uint
	notEmpty *sync.Cond
	notFull  *sync.Cond
	closed   bool
	done     chan struct{}
	mc       *metric
}

// InitPayloadWithScore ...
type InitPayloadWithScore struct {
	*InitPayload
	score float64
}

// NewTaskInitQueue ...
func NewTaskInitQueue(capacity uint, mc *metric) *TaskInitQueue {
	q := &TaskInitQueue{
		queue:    make([]*InitPayloadWithScore, capacity),
		capacity: capacity,
		closed:   false,
		done:     make(chan struct{}),
		mc:       mc,
	}
	q.notEmpty = sync.NewCond(q)
	q.notFull = sync.NewCond(q)
	// 定时重新根据等待时间和执行时间重排序task
	go q.sortByScore()
	return q
}

// Push 添加元素到优先队列，队列满时阻塞入队协程
func (s *TaskInitQueue) Push(payload *InitPayload) error {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return fmt.Errorf("send on closed task init queue")
	}

	if payload == nil || payload.flow == nil || payload.task == nil {
		logs.Errorf("push nil payload to task init queue")
		return fmt.Errorf("push nil payload to task init queue")
	}

	// 等到队列有空位才继续执行下去
	for s.size >= s.capacity && !s.closed {
		s.notFull.Wait()
	}

	// 再次检查，防止在等待期间队列被关闭
	if s.closed {
		return fmt.Errorf("send on closed task init queue")
	}

	// 入队
	s.queue[s.tail] = &InitPayloadWithScore{
		InitPayload: payload,
		score:       0,
	}
	s.tail = (s.tail + 1) % s.capacity
	s.size++
	s.notEmpty.Signal()
	s.mc.taskInitQueueSize.WithLabelValues("init_queue").Set(float64(s.size))
	return nil
}

// Pop 从优先队列取出元素，队列空时阻塞出队协程，如果队列已关闭且为空，返回nil和false
func (s *TaskInitQueue) Pop() (*InitPayload, bool) {
	s.Lock()
	defer s.Unlock()

	// 等到队列有元素才继续执行下去
	for s.size == 0 && !s.closed {
		s.notEmpty.Wait()
	}

	// 如果队列已关闭且为空，返回nil
	if s.closed && s.size == 0 {
		return nil, false
	}

	// 出队
	payload := s.queue[s.head]
	s.head = (s.head + 1) % s.capacity
	s.size--
	s.notFull.Signal()
	s.mc.taskInitQueueSize.WithLabelValues("init_queue").Set(float64(s.size))
	return payload.InitPayload, true
}

// Size 通过锁准确地返回当前队列大小
func (s *TaskInitQueue) Size() uint {
	s.Lock()
	defer s.Unlock()
	return s.size
}

// Capacity 返回队列容量
func (s *TaskInitQueue) Capacity() uint {
	return s.capacity
}

// Close 关闭队列，唤醒所有等待的goroutine
func (s *TaskInitQueue) Close() {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return
	}

	s.closed = true
	close(s.done)
	s.notEmpty.Broadcast()
	s.notFull.Broadcast()
}

// sortByScore 每 1 秒对队列元素按 score 降序重排
func (s *TaskInitQueue) sortByScore() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done: // 队列关闭，立即退出
			return
		case <-ticker.C:
			s.Lock()
			if s.size <= 1 { // 0 或 1 个元素无需排序
				s.Unlock()
				continue
			}

			curTime := time.Now()
			// 1. 收集逻辑区间元素、计算 score
			elements := make([]*InitPayloadWithScore, 0, s.size)
			for i := uint(0); i < s.size; i++ {
				idx := (s.head + i) % s.capacity
				item := s.queue[idx]
				waitTime := curTime.Sub(item.entryTime).Seconds()
				norWait := waitTime / (waitTime + 1)
				norExec := 1 / (1 + item.task.ExecTime)
				item.score = norWait + norExec
				elements = append(elements, item)
			}

			// 2. 排序（降序）
			sort.Slice(elements, func(i, j int) bool {
				return elements[i].score > elements[j].score
			})

			// 3. 写回循环队列，保持 head 不变
			for i := uint(0); i < s.size; i++ {
				s.queue[(s.head+i)%s.capacity] = elements[i]
			}

			s.Unlock()
		}
	}
}
