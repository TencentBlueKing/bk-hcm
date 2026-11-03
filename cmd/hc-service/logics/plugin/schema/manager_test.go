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
	"testing"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

type stubSchemaGetter struct {
	schema *coreplugin.PluginSchema
	err    error
}

func (g *stubSchemaGetter) GetSchema(kt *kit.Kit) (*coreplugin.PluginSchema, error) {
	return g.schema, g.err
}

func TestSchemaManagerSyncSchemaAndGetSchema(t *testing.T) {
	data := newMockSchemaData()
	cache := newSchemaCacheForTest(data, DefaultSchemaCacheTTL)
	defer cache.Close()
	mgr := newSchemaManagerForTest(data, cache)
	kt := kit.New()

	validSchema := buildTestPluginSchema(t, "tencentcloud", "1.81.5", "initial-provider")

	// 1. Unregistered plugin in iac_plugin -> sync_result=failed
	unreg, err := mgr.SyncSchema(kt, &stubSchemaGetter{schema: validSchema})
	if err != nil {
		t.Fatalf("SyncSchema returned unexpected error: %v", err)
	}
	if unreg.SyncResult != enumor.SchemaSyncFailed || unreg.Schema != nil || len(unreg.SyncMessage) == 0 {
		t.Fatalf("expected failed sync when plugin is not registered, got: %+v", unreg)
	}

	// Register plugin in iac_plugin
	regKey := coreplugin.PluginInfo{Name: "tencentcloud", Version: "1.81.5"}
	data.plugins[regKey] = &coreplugin.IaCPlugin{
		ID:            "p-1",
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		Protocol:      enumor.IaCPluginProtocolHTTP,
		State:         enumor.IaCPluginStateInit,
	}

	// 2. First SyncSchema -> inserts into DB, returns sync_result=success, does NOT write to cache
	res1, err := mgr.SyncSchema(kt, &stubSchemaGetter{schema: validSchema})
	if err != nil {
		t.Fatalf("expected first SyncSchema to succeed, got: %v", err)
	}
	if res1.SyncResult != enumor.SchemaSyncSuccess || res1.SyncMessage != "" ||
		res1.Schema.SchemaHash != validSchema.SchemaHash {
		t.Fatalf("unexpected first SyncSchema result: %+v", res1)
	}
	cache.mu.RLock()
	_, inCacheAfterSync := cache.schemaCache[regKey]
	cache.mu.RUnlock()
	if inCacheAfterSync {
		t.Fatalf("SyncSchema must not write to SchemaCache")
	}

	// 3. Second SyncSchema with identical schema_hash -> idempotent skip, still sync_result=success
	res2, err := mgr.SyncSchema(kt, &stubSchemaGetter{schema: validSchema})
	if err != nil {
		t.Fatalf("expected idempotent SyncSchema to succeed, got: %v", err)
	}
	if res2.SyncResult != enumor.SchemaSyncSuccess || res2.SyncMessage != "" || res2.Schema == nil {
		t.Fatalf("expected success on idempotent SyncSchema, got: %+v", res2)
	}

	// 4. Third SyncSchema with modified schema for same (plugin_name, plugin_version) -> sync_result=failed
	tamperedSchema := buildTestPluginSchema(t, "tencentcloud", "1.81.5", "tampered-provider")
	conflict, err := mgr.SyncSchema(kt, &stubSchemaGetter{schema: tamperedSchema})
	if err != nil {
		t.Fatalf("SyncSchema returned unexpected error: %v", err)
	}
	if conflict.SyncResult != enumor.SchemaSyncFailed || conflict.Schema != nil || len(conflict.SyncMessage) == 0 {
		t.Fatalf("expected failed sync on conflicting schema_hash, got: %+v", conflict)
	}

	// 5. GetSchema via SchemaManager -> loads from DB and returns validSchema
	fetched, err := mgr.GetSchema(kt, regKey)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}
	if fetched.SchemaHash != validSchema.SchemaHash {
		t.Fatalf("expected hash %s, got %s", validSchema.SchemaHash, fetched.SchemaHash)
	}
}
