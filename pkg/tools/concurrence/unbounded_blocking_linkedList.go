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

package concurrence

import (
	"fmt"
	"sync"

	"hcm/pkg/tools/linkedList"
)

// UnboundedBlockingLinkedList 无界阻塞pop链表
type UnboundedBlockingLinkedList[T any] struct {
	sync.Mutex
	notEmpty *sync.Cond
	list     *linkedList.LinkedList[T]
	closed   bool
}

// NewUnboundedBlockingLinkedList 创建一个无界阻塞pop链表
func NewUnboundedBlockingLinkedList[T any]() *UnboundedBlockingLinkedList[T] {
	s := &UnboundedBlockingLinkedList[T]{
		list:   &linkedList.LinkedList[T]{},
		closed: false,
	}
	s.notEmpty = sync.NewCond(s)
	return s
}

// Push 入队，永不阻塞
func (s *UnboundedBlockingLinkedList[T]) Push(data T) error {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return fmt.Errorf("send on closed task init queue")
	}

	s.list.PushBack(data)
	s.notEmpty.Signal() // 唤醒一个阻塞在 Pop 的 goroutine
	return nil
}

// Pop 弹出队列头部元素，如果队列为空则阻塞
func (s *UnboundedBlockingLinkedList[T]) Pop() (T, bool) {
	s.Lock()
	defer s.Unlock()
	for s.list.IsEmpty() && !s.closed {
		s.notEmpty.Wait()
	}
	// 如果队列已关闭且为空，返回nil
	if s.closed && s.list.IsEmpty() {
		var t T
		return t, false
	}
	return s.list.Pop()
}

// IsEmpty ...
func (s *UnboundedBlockingLinkedList[T]) IsEmpty() bool {
	s.Lock()
	defer s.Unlock()
	return s.list.IsEmpty()
}

// Len ...
func (s *UnboundedBlockingLinkedList[T]) Len() int {
	s.Lock()
	defer s.Unlock()
	return s.list.Len()
}

// Close 关闭队列，唤醒所有等待的goroutine
func (s *UnboundedBlockingLinkedList[T]) Close() {
	s.Lock()
	defer s.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	s.notEmpty.Broadcast()
}
