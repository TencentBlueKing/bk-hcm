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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/filter"

	tfjson "github.com/hashicorp/terraform-json"
)

type mockSchemaData struct {
	mu            sync.RWMutex
	plugins       map[cacheKey]*coreplugin.IaCPlugin
	schemas       map[cacheKey]*coreplugin.PluginSchema
	hashCallCount int32
	fullCallCount int32
	fullQueryHook func()
}

func newMockSchemaData() *mockSchemaData {
	return &mockSchemaData{
		plugins: make(map[cacheKey]*coreplugin.IaCPlugin),
		schemas: make(map[cacheKey]*coreplugin.PluginSchema),
	}
}

type fakeDiscovery struct{}

func (fakeDiscovery) GetServers() ([]string, error) {
	return []string{"http://127.0.0.1"}, nil
}

type inMemoryHTTPClient struct {
	handler http.HandlerFunc
}

func (c *inMemoryHTTPClient) Do(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	c.handler(rec, req)
	return rec.Result(), nil
}

func newMockDataServiceClient(data *mockSchemaData) *dataservice.Client {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		switch {
		case strings.HasSuffix(path, "/iac_plugins/list"):
			req := new(core.ListReq)
			_ = json.NewDecoder(r.Body).Decode(req)
			key := extractCacheKeyFromFilter(req.Filter)
			data.mu.RLock()
			p, ok := data.plugins[key]
			data.mu.RUnlock()

			var details []coreplugin.IaCPlugin
			if ok && p != nil {
				details = []coreplugin.IaCPlugin{*p}
			}
			writeMockOK(w, &dataplugin.IaCPluginListResult{Details: details})

		case strings.HasSuffix(path, "/iac_plugin_schema/list"):
			req := new(dataplugin.IacPluginSchemaListReq)
			_ = json.NewDecoder(r.Body).Decode(req)
			key := extractCacheKeyFromFilter(req.Filter)

			isHashOnly := len(req.Fields) > 0
			if isHashOnly {
				atomic.AddInt32(&data.hashCallCount, 1)
			} else {
				atomic.AddInt32(&data.fullCallCount, 1)
				if data.fullQueryHook != nil {
					data.fullQueryHook()
				}
			}

			data.mu.RLock()
			s, ok := data.schemas[key]
			data.mu.RUnlock()

			var details []tableplugin.IacPluginSchemaTable
			if ok && s != nil {
				row := tableplugin.IacPluginSchemaTable{
					ID:            "1",
					PluginName:    s.PluginName,
					PluginVersion: s.PluginVersion,
					SchemaHash:    s.SchemaHash,
				}
				if !isHashOnly {
					rawBytes, _ := json.Marshal(s.RawSchema)
					row.SchemaDetail = types.JsonField(rawBytes)
				}
				details = []tableplugin.IacPluginSchemaTable{row}
			}
			writeMockOK(w, &dataplugin.IacPluginSchemaListResult{Details: details})

		case strings.HasSuffix(path, "/iac_plugin_schema/batch/create"):
			req := new(dataplugin.IacPluginSchemaBatchCreateReq)
			_ = json.NewDecoder(r.Body).Decode(req)
			if len(req.Items) == 0 {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    errf.InvalidParameter,
					"message": "items is empty",
				})
				return
			}
			item := req.Items[0]
			key := cacheKey{PluginName: item.PluginName, PluginVersion: item.PluginVersion}

			data.mu.Lock()
			if _, exists := data.schemas[key]; exists {
				data.mu.Unlock()
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    errf.Aborted,
					"message": "duplicate key",
				})
				return
			}
			var raw coreplugin.RawSchema
			_ = json.Unmarshal([]byte(item.SchemaDetail), &raw)
			data.schemas[key] = &coreplugin.PluginSchema{
				PluginName:    item.PluginName,
				PluginVersion: item.PluginVersion,
				SchemaHash:    item.SchemaHash,
				RawSchema:     raw,
			}
			data.mu.Unlock()

			writeMockOK(w, &core.BatchCreateResult{IDs: []string{"1"}})
		}
	}

	cap := &restcli.Capability{
		Client:   &inMemoryHTTPClient{handler: handler},
		Discover: fakeDiscovery{},
	}
	return dataservice.NewClient(cap, "v1")
}

func extractCacheKeyFromFilter(expr *filter.Expression) cacheKey {
	var key cacheKey
	if expr == nil {
		return key
	}
	for _, r := range expr.Rules {
		rule, ok := r.(*filter.AtomRule)
		if !ok {
			continue
		}
		val, _ := rule.Value.(string)
		switch rule.Field {
		case "plugin_name":
			key.PluginName = val
		case "plugin_version":
			key.PluginVersion = val
		}
	}
	return key
}

func writeMockOK(w http.ResponseWriter, data interface{}) {
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    errf.OK,
		"message": "",
		"data":    data,
	})
}

func newSchemaCacheForTest(data *mockSchemaData, ttl time.Duration) *SchemaCache {
	return NewSchemaCache(newMockDataServiceClient(data), ttl)
}

func newSchemaManagerForTest(data *mockSchemaData, cache *SchemaCache) *SchemaManager {
	return &SchemaManager{
		dataCli: newMockDataServiceClient(data),
		cache:   cache,
	}
}

func buildTestPluginSchema(t *testing.T, name, version, desc string) *coreplugin.PluginSchema {
	t.Helper()
	raw := coreplugin.RawSchema{
		ProtocolVersion: "1.0",
		Capabilities: map[string]bool{
			coreplugin.CapabilityApply:   true,
			coreplugin.CapabilityPlan:    true,
			coreplugin.CapabilityRefresh: true,
		},
		Provider: &tfjson.Schema{
			Version: 0,
			Block: &tfjson.SchemaBlock{
				Description: desc,
			},
		},
		ResourceSchemas: map[string]*tfjson.Schema{
			"tencentcloud_vpc": {
				Version: 0,
				Block:   &tfjson.SchemaBlock{},
			},
		},
	}
	hash, err := coreplugin.ComputeRawSchemaHash(&raw)
	if err != nil {
		t.Fatalf("compute raw schema hash failed: %v", err)
	}
	return &coreplugin.PluginSchema{
		PluginName:    name,
		PluginVersion: version,
		SchemaHash:    hash,
		RawSchema:     raw,
	}
}

// waitSchemaCached 轮询同包可见的缓存，直到异步写入到达期望的 schema_hash 且尚未过期。
func waitSchemaCached(t *testing.T, cache *SchemaCache, key cacheKey, hash string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		cache.mu.RLock()
		cached, ok := cache.schemaCache[key]
		expireAt, hasExpire := cache.expireMap[key]
		cache.mu.RUnlock()
		if ok && cached != nil && cached.SchemaHash == hash && hasExpire && time.Now().Before(expireAt) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("schema cache was not updated for %s@%s hash %s", key.PluginName, key.PluginVersion, hash)
}

func TestSchemaCacheHitMissExpireAndMismatch(t *testing.T) {
	data := newMockSchemaData()
	s1 := buildTestPluginSchema(t, "tencentcloud", "1.81.5", "v1")
	key := cacheKey{PluginName: "tencentcloud", PluginVersion: "1.81.5"}
	data.schemas[key] = s1

	cache := newSchemaCacheForTest(data, 7*24*time.Hour)
	defer cache.Close()

	kt := kit.New()

	// 1. Cold start cache miss -> loads full schema from DB and asynchronously writes to cache
	got, err := cache.Get(kt, "tencentcloud", "1.81.5")
	if err != nil {
		t.Fatalf("Get on cold start failed: %v", err)
	}
	if got.SchemaHash != s1.SchemaHash {
		t.Fatalf("expected hash %s, got %s", s1.SchemaHash, got.SchemaHash)
	}
	waitSchemaCached(t, cache, key, s1.SchemaHash)
	if atomic.LoadInt32(&data.fullCallCount) != 1 {
		t.Fatalf("expected 1 full DB query on miss, got %d", data.fullCallCount)
	}

	// 2. Cache hit & not expired & hash matches -> only queries schema_hash, refreshes TTL
	got2, err := cache.Get(kt, "tencentcloud", "1.81.5")
	if err != nil {
		t.Fatalf("Get on cache hit failed: %v", err)
	}
	if got2.SchemaHash != s1.SchemaHash {
		t.Fatalf("expected hash %s, got %s", s1.SchemaHash, got2.SchemaHash)
	}
	if atomic.LoadInt32(&data.hashCallCount) != 1 || atomic.LoadInt32(&data.fullCallCount) != 1 {
		t.Fatalf("expected 1 hash call and still 1 full call, got hash=%d full=%d",
			data.hashCallCount, data.fullCallCount)
	}

	// 3. Cache hit & not expired, but DB schema_hash changed -> reloads full schema from DB & updates cache
	s2 := buildTestPluginSchema(t, "tencentcloud", "1.81.5", "v2-updated")
	data.mu.Lock()
	data.schemas[key] = s2
	data.mu.Unlock()

	got3, err := cache.Get(kt, "tencentcloud", "1.81.5")
	if err != nil {
		t.Fatalf("Get on hash mismatch failed: %v", err)
	}
	if got3.SchemaHash != s2.SchemaHash {
		t.Fatalf("expected updated schema hash %s, got %s", s2.SchemaHash, got3.SchemaHash)
	}
	waitSchemaCached(t, cache, key, s2.SchemaHash)
	if atomic.LoadInt32(&data.fullCallCount) != 2 {
		t.Fatalf("expected 2 full DB queries after hash mismatch, got %d", data.fullCallCount)
	}

	// 4. Cache expired -> triggers eviction, reloads full schema from DB & updates cache
	cache.mu.Lock()
	cache.expireMap[key] = time.Now().Add(-time.Second)
	cache.mu.Unlock()

	got4, err := cache.Get(kt, "tencentcloud", "1.81.5")
	if err != nil {
		t.Fatalf("Get after expiration failed: %v", err)
	}
	if got4.SchemaHash != s2.SchemaHash {
		t.Fatalf("expected hash %s, got %s", s2.SchemaHash, got4.SchemaHash)
	}
	waitSchemaCached(t, cache, key, s2.SchemaHash)
	if atomic.LoadInt32(&data.fullCallCount) != 3 {
		t.Fatalf("expected 3 full DB queries after expiration, got %d", data.fullCallCount)
	}

	// 5. DB record deleted -> returns RecordNotFound error and evicts cache entry
	data.mu.Lock()
	delete(data.schemas, key)
	data.mu.Unlock()

	if _, err = cache.Get(kt, "tencentcloud", "1.81.5"); err == nil {
		t.Fatalf("expected error when schema record is missing in DB")
	}
}

func TestSchemaCacheSingleflightConcurrency(t *testing.T) {
	data := newMockSchemaData()
	s1 := buildTestPluginSchema(t, "tencentcloud", "1.81.5", strings.Repeat("x", 16))
	key := cacheKey{PluginName: "tencentcloud", PluginVersion: "1.81.5"}
	data.schemas[key] = s1

	data.fullQueryHook = func() {
		time.Sleep(30 * time.Millisecond)
	}

	cache := newSchemaCacheForTest(data, DefaultSchemaCacheTTL)
	defer cache.Close()

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			res, err := cache.Get(kit.New(), "tencentcloud", "1.81.5")
			if err != nil {
				t.Errorf("concurrent Get failed: %v", err)
				return
			}
			if res.SchemaHash != s1.SchemaHash {
				t.Errorf("unexpected hash: %s", res.SchemaHash)
			}
		}()
	}
	wg.Wait()

	if calls := atomic.LoadInt32(&data.fullCallCount); calls != 1 {
		t.Fatalf("expected singleflight to coalesce 10 concurrent misses into 1 DB query, got %d", calls)
	}
}
