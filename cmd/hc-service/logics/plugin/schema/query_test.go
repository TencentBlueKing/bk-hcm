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
	raw := coreplugin.RawSchema{
		ProtocolVersion: "1.0",
		Capabilities: map[string]bool{
			coreplugin.CapabilityApply: true,
		},
	}
	rawBytes, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal raw schema failed: %v", err)
	}

	record := tableplugin.IacPluginSchemaTable{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		SchemaHash:    "sha256:test",
		SchemaDetail:  types.JsonField(rawBytes),
	}
	converted, err := toPluginSchema(record)
	if err != nil {
		t.Fatalf("convert plugin schema failed: %v", err)
	}

	if converted.PluginName != record.PluginName ||
		converted.PluginVersion != record.PluginVersion ||
		converted.SchemaHash != record.SchemaHash ||
		converted.ProtocolVersion != raw.ProtocolVersion {
		t.Fatalf("unexpected converted plugin schema: %+v", converted)
	}
}

func TestToPluginSchemaNilRawSchema(t *testing.T) {
	record := tableplugin.IacPluginSchemaTable{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		SchemaHash:    "sha256:test",
	}

	converted, err := toPluginSchema(record)
	if err != nil {
		t.Fatalf("convert empty raw schema failed: %v", err)
	}
	if converted.PluginName != record.PluginName || converted.Capabilities != nil {
		t.Fatalf("unexpected converted plugin schema: %+v", converted)
	}
}
