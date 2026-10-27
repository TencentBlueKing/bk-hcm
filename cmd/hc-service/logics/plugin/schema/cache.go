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

package schema

import (
	"fmt"
	"sync"
	"time"

	coreplugin "hcm/pkg/api/core/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"golang.org/x/sync/singleflight"
)

// SchemaCache 进程内懒加载 Schema 缓存，支持滑动 TTL 与 DB schema_hash 校验。
type SchemaCache struct {
	mu          sync.RWMutex
	schemaCache map[coreplugin.PluginInfo]*coreplugin.PluginSchema
	expireMap   map[coreplugin.PluginInfo]time.Time
	ttl         time.Duration
	evictCh     chan struct{}
	stopCh      chan struct{}
	stopOnce    sync.Once
	sfGroup     singleflight.Group
	dataCli     *dataservice.Client
}

// NewSchemaCache 基于 data-service Client 创建 SchemaCache 并启动过期清理协程。
func NewSchemaCache(dataCli *dataservice.Client, ttl time.Duration) *SchemaCache {
	if ttl <= 0 {
		ttl = DefaultSchemaCacheTTL
	}

	c := &SchemaCache{
		schemaCache: make(map[coreplugin.PluginInfo]*coreplugin.PluginSchema),
		expireMap:   make(map[coreplugin.PluginInfo]time.Time),
		ttl:         ttl,
		evictCh:     make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
		dataCli:     dataCli,
	}

	go c.evictLoop()
	return c
}

// Close 停止 SchemaCache 后台过期清理协程。
func (c *SchemaCache) Close() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// Get 按 PluginInfo（name, version）获取 Schema，含命中/TTL/哈希校验。
func (c *SchemaCache) Get(kt *kit.Kit, plugin coreplugin.PluginInfo) (*coreplugin.PluginSchema, error) {
	if len(plugin.Name) == 0 || len(plugin.Version) == 0 {
		return nil, errf.New(errf.InvalidParameter, "plugin.name and plugin.version are required")
	}

	c.mu.RLock()
	cached, hit := c.schemaCache[plugin]
	expireAt := c.expireMap[plugin]
	c.mu.RUnlock()

	now := time.Now()
	expired := hit && !now.Before(expireAt)
	if expired {
		logs.Infof("plugin schema cache entry expired, plugin: %s@%s, triggering async eviction, rid: %s",
			plugin.Name, plugin.Version, kt.Rid)
		c.triggerEvict()
	}

	if hit && !expired {
		dbHash, found, err := c.querySchemaHash(kt, plugin)
		if err != nil {
			logs.Errorf("query plugin schema_hash from db failed, plugin: %s@%s, err: %v, rid: %s",
				plugin.Name, plugin.Version, err, kt.Rid)
			return nil, err
		}

		if !found {
			c.deleteKey(plugin)
			logs.Errorf("plugin schema not found in db during cache hash check, plugin: %s@%s, rid: %s",
				plugin.Name, plugin.Version, kt.Rid)
			return nil, errf.Newf(errf.RecordNotFound, "plugin schema not found for %s@%s",
				plugin.Name, plugin.Version)
		}

		if cached.SchemaHash == dbHash {
			c.refreshExpire(plugin, cached.SchemaHash)
			return cached, nil
		}

		logs.Warnf("plugin schema_hash mismatch between cache and db, plugin: %s@%s, "+
			"cached: %s, db: %s, reloading full schema from db, rid: %s",
			plugin.Name, plugin.Version, cached.SchemaHash, dbHash, kt.Rid)
	}

	// 缓存未命中 / 已过期 / schema_hash 与 DB 不一致：经 singleflight 回源查全量 Schema。
	return c.loadFromDBAndAsyncCache(kt, plugin)
}

func (c *SchemaCache) loadFromDBAndAsyncCache(kt *kit.Kit, plugin coreplugin.PluginInfo) (
	*coreplugin.PluginSchema, error) {

	sfKey := fmt.Sprintf("%s:%s", plugin.Name, plugin.Version)

	val, err, _ := c.sfGroup.Do(sfKey, func() (interface{}, error) {
		fullSchema, found, queryErr := c.queryFullSchema(kt, plugin)
		if queryErr != nil {
			logs.Errorf("load full plugin schema from db failed, plugin: %s@%s, err: %v, rid: %s",
				plugin.Name, plugin.Version, queryErr, kt.Rid)
			return nil, queryErr
		}

		if !found {
			c.deleteKey(plugin)
			logs.Errorf("plugin schema not found in db, plugin: %s@%s, rid: %s",
				plugin.Name, plugin.Version, kt.Rid)
			return nil, errf.Newf(errf.RecordNotFound, "plugin schema not found for %s@%s",
				plugin.Name, plugin.Version)
		}

		// 写缓存放在 singleflight 闭包内，同一次回源只更新一次，调用方不必等待写锁。
		go c.asyncUpdateCache(plugin, fullSchema)
		return fullSchema, nil
	})
	if err != nil {
		return nil, err
	}

	dbSchema, ok := val.(*coreplugin.PluginSchema)
	if !ok || dbSchema == nil {
		logs.Errorf("unexpected singleflight result for plugin schema %s@%s, rid: %s",
			plugin.Name, plugin.Version, kt.Rid)
		return nil, errf.New(errf.Aborted, "unexpected nil plugin schema loaded from db")
	}

	return dbSchema, nil
}

func (c *SchemaCache) asyncUpdateCache(plugin coreplugin.PluginInfo, schema *coreplugin.PluginSchema) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.schemaCache[plugin] = schema
	c.expireMap[plugin] = time.Now().Add(c.ttl)
}

func (c *SchemaCache) refreshExpire(plugin coreplugin.PluginInfo, expectedHash string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cur, ok := c.schemaCache[plugin]; ok && cur.SchemaHash == expectedHash {
		c.expireMap[plugin] = time.Now().Add(c.ttl)
	}
}

func (c *SchemaCache) deleteKey(plugin coreplugin.PluginInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.schemaCache, plugin)
	delete(c.expireMap, plugin)
}

func (c *SchemaCache) triggerEvict() {
	select {
	case c.evictCh <- struct{}{}:
	default:
	}
}

func (c *SchemaCache) evictLoop() {
	for {
		select {
		case <-c.stopCh:
			return
		case <-c.evictCh:
			c.evictExpired()
		}
	}
}

func (c *SchemaCache) evictExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, exp := range c.expireMap {
		if !now.Before(exp) {
			delete(c.schemaCache, k)
			delete(c.expireMap, k)
		}
	}
}
