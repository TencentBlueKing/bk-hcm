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
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/tools/json"
)

func TestToPluginSchema(t *testing.T) {
	detail := coreplugin.SchemaDetail{
		ProtocolVersion: "1.0",
		Capabilities: coreplugin.Capabilities{
			Apply: true,
		},
	}
	detailBytes, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("marshal schema detail failed: %v", err)
	}

	record := tableplugin.IacPluginSchemaTable{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		SchemaHash:    "sha256:test",
		SchemaDetail:  types.JsonField(detailBytes),
	}
	converted, err := toPluginSchema(record)
	if err != nil {
		t.Fatalf("convert plugin schema failed: %v", err)
	}

	if converted.Plugin.Name != record.PluginName ||
		converted.Plugin.Version != record.PluginVersion ||
		converted.SchemaHash != record.SchemaHash ||
		converted.ProtocolVersion != detail.ProtocolVersion ||
		!converted.Capabilities.Apply {
		t.Fatalf("unexpected converted plugin schema: %+v", converted)
	}
}

func TestToPluginSchemaNilSchemaDetail(t *testing.T) {
	record := tableplugin.IacPluginSchemaTable{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		SchemaHash:    "sha256:test",
	}

	converted, err := toPluginSchema(record)
	if err != nil {
		t.Fatalf("convert empty schema detail failed: %v", err)
	}
	if converted.Plugin.Name != record.PluginName || converted.Capabilities != (coreplugin.Capabilities{}) {
		t.Fatalf("unexpected converted plugin schema: %+v", converted)
	}
}
