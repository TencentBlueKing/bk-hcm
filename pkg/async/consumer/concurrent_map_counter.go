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

// Package consumer ...
package consumer

import (
	"sync"
	"sync/atomic"
)

// ConcurrentMapCounter 并发安全的 map[string]int 计数器
type ConcurrentMapCounter struct {
	m sync.Map // key -> *uint64
}

// Inc 对 key 加 delta（可为负）
func (c *ConcurrentMapCounter) Inc(key string, delta int64) uint64 {
	// 第一次取不到就建指针
	ptrI, _ := c.m.LoadOrStore(key, new(uint64))
	ptr := ptrI.(*uint64)

	// 原子加
	newVal := atomic.AddUint64(ptr, uint64(delta))
	return newVal
}

// Get 返回 key 当前的计数值以及该 key 是否存在。若 key 不存在则返回 0
func (c *ConcurrentMapCounter) Get(key string) uint64 {
	ptrI, ok := c.m.Load(key)
	if !ok {
		return 0
	}
	return atomic.LoadUint64(ptrI.(*uint64))
}

// Snapshot 返回此刻 map 的完整快照 map[string]uint64。返回的 map 与内部数据已无任何共享，调用方可以随意读写。
func (c *ConcurrentMapCounter) Snapshot() map[string]uint64 {
	out := make(map[string]uint64)
	c.m.Range(func(key, value any) bool {
		k := key.(string)
		v := atomic.LoadUint64(value.(*uint64))
		out[k] = v
		return true
	})
	return out
}
