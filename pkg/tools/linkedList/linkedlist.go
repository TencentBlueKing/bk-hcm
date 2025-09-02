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

package linkedList

import (
	"container/list"
)

// LinkedList 泛型链表
type LinkedList[T any] struct {
	list.List
}

// Push 方法将元素添加到链表尾部
func (s *LinkedList[T]) Push(value T) {
	s.PushBack(value)
}

// Pop 方法从链表头部移除并返回元素
func (s *LinkedList[T]) Pop() (T, bool) {
	if s.Len() == 0 {
		var zero T
		return zero, false
	}
	element := s.Front()
	s.Remove(element)
	return element.Value.(T), true
}

// IsEmpty 方法检查链表是否为空
func (s *LinkedList[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Len 方法返回链表的长度
func (s *LinkedList[T]) Len() int {
	return s.List.Len()
}
