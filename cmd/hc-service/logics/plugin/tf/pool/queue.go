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

package pool

const (
	// queueShrinkThreshold 队首已出队的位置达到该数量、且占切片一半以上时，把剩余元素搬回切片开头。
	queueShrinkThreshold = 64
	// queueCompactThreshold 已取消但仍留在切片里的元素达到该数量、且多于有效元素时，整体清理一次。
	queueCompactThreshold = 64
)

// queue 是基于切片的先进先出队列，支持按元素取消（延迟清理）。
//
// 使用约定：
//   - 调用方必须持有外部锁；
//   - 同一元素同一时刻至多在队列中出现一次；
//   - Remove 只能传入当前仍在队列中的元素。
type queue[T comparable] struct {
	buf  []T
	head int
	// removed 记录已取消、但仍留在 buf[head:] 中的元素。
	removed map[T]struct{}
}

func newQueue[T comparable]() *queue[T] {
	return &queue[T]{removed: make(map[T]struct{})}
}

// PushBack 把元素追加到队尾。
func (q *queue[T]) PushBack(v T) {
	q.buf = append(q.buf, v)
}

// Peek 返回队首的有效元素，不出队；若队首是已取消元素则顺手移出。
func (q *queue[T]) Peek() (T, bool) {
	for q.head < len(q.buf) {
		v := q.buf[q.head]
		if _, dead := q.removed[v]; !dead {
			return v, true
		}
		delete(q.removed, v)
		q.dropFront()
	}

	var zero T
	return zero, false
}

// PopFront 弹出并返回队首的有效元素。
func (q *queue[T]) PopFront() (T, bool) {
	v, ok := q.Peek()
	if ok {
		q.dropFront()
	}
	return v, ok
}

// Remove 取消一个仍在队列中的元素：先在 removed 中标记，等它走到队首或触发压缩时再从切片移出。
func (q *queue[T]) Remove(v T) {
	if _, dead := q.removed[v]; dead {
		return
	}
	q.removed[v] = struct{}{}

	if n := len(q.removed); n >= queueCompactThreshold && n > q.Len() {
		q.compact()
	}
}

// Len 返回有效元素数，不含已取消的元素。
func (q *queue[T]) Len() int {
	return len(q.buf) - q.head - len(q.removed)
}

// Drain 按顺序取出全部有效元素并清空队列。
func (q *queue[T]) Drain() []T {
	out := make([]T, 0, q.Len())
	for v, ok := q.PopFront(); ok; v, ok = q.PopFront() {
		out = append(out, v)
	}

	q.buf, q.head = nil, 0
	clear(q.removed)
	return out
}

// dropFront 丢弃当前队首 buf[head]，并在队首空位过多时把剩余元素搬回切片开头。
func (q *queue[T]) dropFront() {
	var zero T
	q.buf[q.head] = zero
	q.head++

	if q.head == len(q.buf) {
		q.buf, q.head = q.buf[:0], 0
		return
	}
	if q.head >= queueShrinkThreshold && q.head*2 >= len(q.buf) {
		n := copy(q.buf, q.buf[q.head:])
		clear(q.buf[n:])
		q.buf, q.head = q.buf[:n], 0
	}
}

// compact 移除全部已取消元素，保持有效元素的相对顺序。
func (q *queue[T]) compact() {
	n := 0
	for _, v := range q.buf[q.head:] {
		if _, dead := q.removed[v]; dead {
			continue
		}
		q.buf[n] = v
		n++
	}

	clear(q.buf[n:])
	q.buf, q.head = q.buf[:n], 0
	clear(q.removed)
}
