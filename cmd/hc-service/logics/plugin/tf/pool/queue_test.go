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

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type item struct{ n int }

func items(n int) []*item {
	out := make([]*item, n)
	for i := range out {
		out[i] = &item{n: i}
	}
	return out
}

func drainNums(q *queue[*item]) []int {
	var out []int
	for _, v := range q.Drain() {
		out = append(out, v.n)
	}
	return out
}

func TestQueueOrder(t *testing.T) {
	q := newQueue[*item]()
	_, ok := q.Peek()
	assert.False(t, ok)
	_, ok = q.PopFront()
	assert.False(t, ok)

	for _, v := range items(3) {
		q.PushBack(v)
	}
	assert.Equal(t, 3, q.Len())

	v, ok := q.Peek()
	require.True(t, ok)
	assert.Equal(t, 0, v.n)
	assert.Equal(t, 3, q.Len(), "peek must not pop")

	v, _ = q.PopFront()
	assert.Equal(t, 0, v.n)
	assert.Equal(t, []int{1, 2}, drainNums(q))
	assert.Equal(t, 0, q.Len())
}

func TestQueueRemove(t *testing.T) {
	q := newQueue[*item]()
	vs := items(5)
	for _, v := range vs {
		q.PushBack(v)
	}

	q.Remove(vs[0]) // 队首
	q.Remove(vs[2]) // 中间
	q.Remove(vs[4]) // 队尾
	q.Remove(vs[2]) // 重复取消无副作用
	assert.Equal(t, 2, q.Len())

	v, ok := q.Peek()
	require.True(t, ok)
	assert.Equal(t, 1, v.n, "peek skips removed head")
	assert.Equal(t, 2, q.Len())

	assert.Equal(t, []int{1, 3}, drainNums(q))
	assert.Equal(t, 0, q.Len())
	assert.Empty(t, q.removed)

	// 全部取消后队列为空。
	q.PushBack(vs[0])
	q.Remove(vs[0])
	assert.Equal(t, 0, q.Len())
	_, ok = q.PopFront()
	assert.False(t, ok)
	assert.Empty(t, q.removed)
}

func TestQueueCompactKeepsOrder(t *testing.T) {
	q := newQueue[*item]()
	vs := items(1000)
	for _, v := range vs {
		q.PushBack(v)
	}

	var want []int
	for i, v := range vs {
		if i%3 == 0 {
			want = append(want, i)
			continue
		}
		q.Remove(v)
	}

	assert.Equal(t, len(want), q.Len())
	assert.Less(t, len(q.buf), 1000, "compaction should have run")
	assert.Equal(t, want, drainNums(q))
}

// TestQueueAgainstModel 随机操作序列与参照模型逐步比对，并检查缓冲区占用有界。
func TestQueueAgainstModel(t *testing.T) {
	rng := rand.New(rand.NewSource(20261023))
	q := newQueue[*item]()
	var model []*item
	next := 0

	for step := 0; step < 200000; step++ {
		switch op := rng.Intn(10); {
		case op < 5:
			v := &item{n: next}
			next++
			q.PushBack(v)
			model = append(model, v)
		case op < 7:
			v, ok := q.PopFront()
			if len(model) == 0 {
				require.False(t, ok, "step %d", step)
				break
			}
			require.True(t, ok, "step %d", step)
			require.Same(t, model[0], v, "step %d", step)
			model = model[1:]
		case op < 9:
			if len(model) == 0 {
				break
			}
			i := rng.Intn(len(model))
			q.Remove(model[i])
			model = slices.Delete(model, i, i+1)
		default:
			v, ok := q.Peek()
			require.Equal(t, len(model) > 0, ok, "step %d", step)
			if ok {
				require.Same(t, model[0], v, "step %d", step)
			}
		}

		require.Equal(t, len(model), q.Len(), "step %d", step)
		live := len(model)
		bound := 2*(2*live+queueCompactThreshold) + queueShrinkThreshold
		require.LessOrEqual(t, len(q.buf), bound, "step %d: buffer grows unbounded", step)
	}

	got := q.Drain()
	require.Equal(t, len(model), len(got))
	for i := range got {
		require.Same(t, model[i], got[i])
	}
}

func TestQueuePoppedSlotsCleared(t *testing.T) {
	q := newQueue[*item]()
	vs := items(10)
	for _, v := range vs {
		q.PushBack(v)
	}
	for i := 0; i < 5; i++ {
		_, _ = q.PopFront()
	}

	// 已出队的位置必须清零，避免切片继续持有已完成的请求。
	for i := 0; i < q.head; i++ {
		assert.Nil(t, q.buf[i])
	}
}
