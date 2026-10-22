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
	"time"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

// SchemaManager hc-service 侧插件 Schema 同步与读取。
type SchemaManager struct {
	dataCli *dataservice.Client
	cache   *SchemaCache
}

// NewSchemaManager 基于 data-service Client 与内部 SchemaCache 创建 SchemaManager。
func NewSchemaManager(dataCli *dataservice.Client, cacheTTL time.Duration) *SchemaManager {
	return &SchemaManager{
		dataCli: dataCli,
		cache:   NewSchemaCache(dataCli, cacheTTL),
	}
}

// SyncSchema 通过 getter 获取 Schema，对照 iac_plugin 与南向规则校验，
// 幂等且防篡改地写入 iac_plugin_schema；本阶段不写入本地缓存。
func (m *SchemaManager) SyncSchema(kt *kit.Kit, getter SchemaGetter) (*SyncSchemaResult, error) {
	if getter == nil {
		logs.Errorf("sync plugin schema failed: schema getter is nil, rid: %s", kt.Rid)
		return newSyncFailed("schema getter is nil"), nil
	}

	pluginSchema, err := getter.GetSchema(kt)
	if err != nil {
		logs.Errorf("get plugin schema via getter failed, err: %v, rid: %s", err, kt.Rid)
		return newSyncFailed(err.Error()), nil
	}
	if pluginSchema == nil {
		logs.Errorf("get plugin schema via getter returned nil schema, rid: %s", kt.Rid)
		return newSyncFailed("plugin schema is nil"), nil
	}

	regPlugin, found, err := m.getPlugin(kt, pluginSchema.PluginName, pluginSchema.PluginVersion)
	if err != nil {
		logs.Errorf("query iac_plugin registration failed, plugin: %s@%s, err: %v, rid: %s",
			pluginSchema.PluginName, pluginSchema.PluginVersion, err, kt.Rid)
		return newSyncFailed(err.Error()), nil
	}
	if !found || regPlugin == nil {
		logs.Errorf("plugin registration not found in iac_plugin, plugin: %s@%s, rid: %s",
			pluginSchema.PluginName, pluginSchema.PluginVersion, kt.Rid)
		return newSyncFailed(fmt.Sprintf("plugin %s@%s is not registered in iac_plugin",
			pluginSchema.PluginName, pluginSchema.PluginVersion)), nil
	}

	if err := coreplugin.ValidatePluginSchema(pluginSchema, regPlugin.PluginName, regPlugin.PluginVersion,
		regPlugin.Protocol); err != nil {
		logs.Errorf("validate plugin schema failed, plugin: %s@%s, protocol: %s, err: %v, rid: %s",
			pluginSchema.PluginName, pluginSchema.PluginVersion, regPlugin.Protocol, err, kt.Rid)
		return newSyncFailed(err.Error()), nil
	}

	existingHash, exists, err := m.cache.querySchemaHash(kt, pluginSchema.PluginName, pluginSchema.PluginVersion)
	if err != nil {
		logs.Errorf("query existing plugin schema_hash failed, plugin: %s@%s, err: %v, rid: %s",
			pluginSchema.PluginName, pluginSchema.PluginVersion, err, kt.Rid)
		return newSyncFailed(err.Error()), nil
	}

	if exists {
		return m.compareExistingSchemaHash(kt, pluginSchema, existingHash)
	}

	if _, createErr := m.createSchema(kt, pluginSchema); createErr != nil {
		// Handle potential concurrent insert race on unique key (plugin_name, plugin_version)
		recheckHash, recheckExists, recheckErr := m.cache.querySchemaHash(kt, pluginSchema.PluginName,
			pluginSchema.PluginVersion)
		if recheckErr == nil && recheckExists {
			return m.compareExistingSchemaHash(kt, pluginSchema, recheckHash)
		}

		logs.Errorf("create plugin schema failed, plugin: %s@%s, err: %v, rid: %s",
			pluginSchema.PluginName, pluginSchema.PluginVersion, createErr, kt.Rid)
		return newSyncFailed(createErr.Error()), nil
	}

	logs.Infof("sync plugin schema created new record, plugin: %s@%s, schema_hash: %s, rid: %s",
		pluginSchema.PluginName, pluginSchema.PluginVersion, pluginSchema.SchemaHash, kt.Rid)
	return newSyncSuccess(pluginSchema), nil
}

func (m *SchemaManager) compareExistingSchemaHash(kt *kit.Kit, incoming *coreplugin.PluginSchema,
	existingHash string) (*SyncSchemaResult, error) {

	if existingHash == incoming.SchemaHash {
		logs.Infof("sync plugin schema skipped idempotently, plugin: %s@%s, schema_hash: %s, rid: %s",
			incoming.PluginName, incoming.PluginVersion, incoming.SchemaHash, kt.Rid)
		return newSyncSuccess(incoming), nil
	}

	logs.Errorf("plugin schema_hash conflict detected, refuse to overwrite existing schema, "+
		"plugin: %s@%s, existing_hash: %s, incoming_hash: %s, rid: %s",
		incoming.PluginName, incoming.PluginVersion, existingHash, incoming.SchemaHash, kt.Rid)
	return newSyncFailed(fmt.Sprintf("plugin schema_hash conflict for %s@%s: existing %s, incoming %s",
		incoming.PluginName, incoming.PluginVersion, existingHash, incoming.SchemaHash)), nil
}

func newSyncSuccess(schema *coreplugin.PluginSchema) *SyncSchemaResult {
	return &SyncSchemaResult{
		SyncResult:  enumor.SchemaSyncSuccess,
		SyncMessage: "",
		Schema:      schema,
	}
}

func newSyncFailed(message string) *SyncSchemaResult {
	return &SyncSchemaResult{
		SyncResult:  enumor.SchemaSyncFailed,
		SyncMessage: message,
		Schema:      nil,
	}
}

// GetSchema 通过 SchemaCache 按 (pluginName, pluginVersion) 读取 Schema。
func (m *SchemaManager) GetSchema(kt *kit.Kit, pluginName, pluginVersion string) (*coreplugin.PluginSchema, error) {
	schema, err := m.cache.Get(kt, pluginName, pluginVersion)
	if err != nil {
		logs.Errorf("get plugin schema failed, plugin: %s@%s, err: %v, rid: %s",
			pluginName, pluginVersion, err, kt.Rid)
		return nil, err
	}
	return schema, nil
}

// getPlugin 按 (plugin_name, plugin_version) 查询插件注册记录。
func (m *SchemaManager) getPlugin(kt *kit.Kit, pluginName, pluginVersion string) (
	*coreplugin.IaCPlugin, bool, error) {

	if m.dataCli == nil || m.dataCli.Global == nil || m.dataCli.Global.IaCPlugin == nil {
		return nil, false, errf.New(errf.InvalidParameter, "data-service iac_plugin client is not initialized")
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("plugin_name", pluginName),
			tools.RuleEqual("plugin_version", pluginVersion),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	res, err := m.dataCli.Global.IaCPlugin.List(kt, req)
	if err != nil {
		logs.Errorf("list iac_plugin from data-service failed, plugin: %s@%s, err: %v, rid: %s",
			pluginName, pluginVersion, err, kt.Rid)
		return nil, false, err
	}

	if res == nil || len(res.Details) == 0 {
		return nil, false, nil
	}

	return &res.Details[0], true, nil
}

func (m *SchemaManager) createSchema(kt *kit.Kit, schema *coreplugin.PluginSchema) (string, error) {
	if m.dataCli == nil || m.dataCli.Global == nil || m.dataCli.Global.IacPluginSchema == nil {
		return "", errf.New(errf.InvalidParameter, "data-service iac_plugin_schema client is not initialized")
	}
	if schema == nil {
		return "", errf.New(errf.InvalidParameter, "plugin schema is nil")
	}

	rawBytes, err := json.Marshal(schema.RawSchema)
	if err != nil {
		logs.Errorf("marshal raw schema failed, plugin: %s@%s, err: %v, rid: %s",
			schema.PluginName, schema.PluginVersion, err, kt.Rid)
		return "", err
	}

	req := &dataplugin.IacPluginSchemaCreateReq{
		PluginName:    schema.PluginName,
		PluginVersion: schema.PluginVersion,
		SchemaHash:    schema.SchemaHash,
		SchemaDetail:  types.JsonField(rawBytes),
	}

	res, err := m.dataCli.Global.IacPluginSchema.Create(kt, req)
	if err != nil {
		logs.Errorf("create iac_plugin_schema in data-service failed, plugin: %s@%s, err: %v, rid: %s",
			schema.PluginName, schema.PluginVersion, err, kt.Rid)
		return "", err
	}

	return res.ID, nil
}
