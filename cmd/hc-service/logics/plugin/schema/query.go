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

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

func (c *SchemaCache) querySchemaHash(kt *kit.Kit, pluginName, pluginVersion string) (string, bool, error) {
	if c == nil || c.dataCli == nil || c.dataCli.Global == nil || c.dataCli.Global.IacPluginSchema == nil {
		return "", false, errf.New(errf.InvalidParameter, "data-service iac_plugin_schema client is not initialized")
	}

	req := &dataplugin.IacPluginSchemaListReq{
		Fields: []string{"id", "plugin_name", "plugin_version", "schema_hash"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("plugin_name", pluginName),
			tools.RuleEqual("plugin_version", pluginVersion),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	res, err := c.dataCli.Global.IacPluginSchema.List(kt, req)
	if err != nil {
		logs.Errorf("list iac_plugin_schema hash from data-service failed, plugin: %s@%s, err: %v, rid: %s",
			pluginName, pluginVersion, err, kt.Rid)
		return "", false, err
	}

	if res == nil || len(res.Details) == 0 {
		return "", false, nil
	}

	return res.Details[0].SchemaHash, true, nil
}

func (c *SchemaCache) queryFullSchema(kt *kit.Kit, pluginName, pluginVersion string) (
	*coreplugin.PluginSchema, bool, error) {

	if c == nil || c.dataCli == nil || c.dataCli.Global == nil || c.dataCli.Global.IacPluginSchema == nil {
		return nil, false, errf.New(errf.InvalidParameter, "data-service iac_plugin_schema client is not initialized")
	}

	req := &dataplugin.IacPluginSchemaListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("plugin_name", pluginName),
			tools.RuleEqual("plugin_version", pluginVersion),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	res, err := c.dataCli.Global.IacPluginSchema.List(kt, req)
	if err != nil {
		logs.Errorf("list full iac_plugin_schema from data-service failed, plugin: %s@%s, err: %v, rid: %s",
			pluginName, pluginVersion, err, kt.Rid)
		return nil, false, err
	}

	if res == nil || len(res.Details) == 0 {
		return nil, false, nil
	}

	pluginSchema, err := toPluginSchema(res.Details[0])
	if err != nil {
		logs.Errorf("convert iac_plugin_schema record to PluginSchema failed, plugin: %s@%s, err: %v, rid: %s",
			pluginName, pluginVersion, err, kt.Rid)
		return nil, false, err
	}

	return pluginSchema, true, nil
}

func toPluginSchema(record tableplugin.IacPluginSchemaTable) (*coreplugin.PluginSchema, error) {
	var raw coreplugin.RawSchema
	if len(record.SchemaDetail) > 0 {
		if err := json.Unmarshal([]byte(record.SchemaDetail), &raw); err != nil {
			return nil, fmt.Errorf("unmarshal raw schema failed, err: %w", err)
		}
	}

	return &coreplugin.PluginSchema{
		PluginName:    record.PluginName,
		PluginVersion: record.PluginVersion,
		SchemaHash:    record.SchemaHash,
		RawSchema:     raw,
	}, nil
}
