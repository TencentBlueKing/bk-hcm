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

// Package concurrence ...
package concurrence

import "sync"

// ConcurrentMapCounter 并发安全的 map[string]int64 计数器
type ConcurrentMapCounter struct {
	mu sync.RWMutex
	m  map[string]int64
}

// NewConcurrentMapCounter 预先创建好内部 map
func NewConcurrentMapCounter() *ConcurrentMapCounter {
	return &ConcurrentMapCounter{m: make(map[string]int64)}
}

// Inc 对 key 加 delta（可为负）。若 key 不存在则视为 0。
func (c *ConcurrentMapCounter) Inc(key string, delta int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] += delta
	return c.m[key]
}

// GetValueAndSum 返回 key 当前值以及所有值之和；若 key 不存在则返回 0 和 0。
func (c *ConcurrentMapCounter) GetValueAndSum(key string) (int64, int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sum := int64(0)
	for _, val := range c.m {
		sum += val
	}
	return c.m[key], sum
}

// Snapshot 返回此刻的完整副本。
func (c *ConcurrentMapCounter) Snapshot() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int64, len(c.m))
	for k, v := range c.m {
		out[k] = v
	}
	return out
}
