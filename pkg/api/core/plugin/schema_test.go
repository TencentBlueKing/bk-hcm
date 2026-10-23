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

package plugin

import (
	stdjson "encoding/json"
	"strings"
	"testing"

	"hcm/pkg/tools/json"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

func newValidTestPluginSchema(t *testing.T) *PluginSchema {
	t.Helper()
	detail := SchemaDetail{
		ProtocolVersion: "1.0",
		Capabilities: Capabilities{
			Apply:      true,
			Plan:       true,
			Refresh:    true,
			ReadData:   true,
			AsyncApply: true,
			Stop:       true,
			Validate:   true,
		},
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema: &tfjson.Schema{
				Version: 0,
				Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"secret_id": {
							AttributeType: cty.String,
							Required:      true,
							Description:   "TencentCloud Secret ID",
						},
					},
				},
			},
			ResourceSchemas: map[string]*tfjson.Schema{
				"tencentcloud_vpc": {
					Version: 0,
					Block: &tfjson.SchemaBlock{
						Attributes: map[string]*tfjson.SchemaAttribute{
							"name": {
								AttributeType: cty.String,
								Required:      true,
							},
						},
					},
				},
			},
			DataSourceSchemas: map[string]*tfjson.Schema{
				"tencentcloud_availability_zones": {
					Version: 0,
					Block:   &tfjson.SchemaBlock{},
				},
			},
		},
	}

	pluginInfo := PluginInfo{
		Name:    "tencentcloud",
		Version: "1.81.5",
	}
	hash, err := ComputeSchemaHash(pluginInfo, &detail)
	if err != nil {
		t.Fatalf("compute schema hash failed: %v", err)
	}

	return &PluginSchema{
		Plugin:       pluginInfo,
		SchemaHash:   hash,
		SchemaDetail: detail,
	}
}

func TestPluginSchemaInlineSerialization(t *testing.T) {
	ps := newValidTestPluginSchema(t)

	// 1. Test hcm/pkg/tools/json marshal & unmarshal
	data, err := json.Marshal(ps)
	if err != nil {
		t.Fatalf("json.Marshal PluginSchema failed: %v", err)
	}

	var topMap map[string]interface{}
	if err := json.Unmarshal(data, &topMap); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}

	expectedTopKeys := []string{
		"plugin", "schema_hash",
		"protocol_version", "capabilities", "provider",
		"resource_schemas", "data_source_schemas",
	}
	for _, key := range expectedTopKeys {
		if _, ok := topMap[key]; !ok {
			t.Errorf("expected key %q at top level of serialized PluginSchema, got map: %v", key, topMap)
		}
	}
	for _, nestedKey := range []string{"SchemaDetail", "ProviderSchema"} {
		if _, hasNested := topMap[nestedKey]; hasNested {
			t.Errorf("unexpected nested %s key in serialized PluginSchema", nestedKey)
		}
	}

	// 2. Test stdlib encoding/json round-trip
	stdBytes, err := stdjson.Marshal(ps)
	if err != nil {
		t.Fatalf("stdjson.Marshal failed: %v", err)
	}
	var roundTrip PluginSchema
	if err := stdjson.Unmarshal(stdBytes, &roundTrip); err != nil {
		t.Fatalf("stdjson.Unmarshal failed: %v", err)
	}
	if roundTrip.Plugin != ps.Plugin || roundTrip.SchemaHash != ps.SchemaHash ||
		roundTrip.ProtocolVersion != ps.ProtocolVersion {
		t.Fatalf("round-trip mismatch: got %+v", roundTrip)
	}

	// 3. Test SchemaDetail standalone serialization excludes outer columns
	detailBytes, err := json.Marshal(ps.SchemaDetail)
	if err != nil {
		t.Fatalf("json.Marshal SchemaDetail failed: %v", err)
	}
	var detailMap map[string]interface{}
	if err := json.Unmarshal(detailBytes, &detailMap); err != nil {
		t.Fatalf("json.Unmarshal SchemaDetail map failed: %v", err)
	}
	for _, forbiddenKey := range []string{"plugin", "plugin_name", "plugin_version", "schema_hash"} {
		if _, exists := detailMap[forbiddenKey]; exists {
			t.Errorf("SchemaDetail JSON must not contain %q", forbiddenKey)
		}
	}
}

func TestValidatePluginSchemaAndCapabilities(t *testing.T) {
	expectedPlugin := PluginInfo{Name: "tencentcloud", Version: "1.81.5"}
	validSchema := newValidTestPluginSchema(t)
	if err := ValidatePluginSchema(validSchema, expectedPlugin, ProtocolHTTP); err != nil {
		t.Fatalf("expected valid schema to pass validation, got err: %v", err)
	}

	// Changing plugin version without recomputing schema_hash must fail hash check
	versionChanged := newValidTestPluginSchema(t)
	versionChanged.Plugin.Version = "1.81.6"
	if err := ValidatePluginSchema(versionChanged, PluginInfo{Name: "tencentcloud", Version: "1.81.6"},
		ProtocolHTTP); err == nil {
		t.Fatalf("expected version change without hash recomputation to fail")
	}

	// tf_provider allows empty protocol_version
	tfSchema := newValidTestPluginSchema(t)
	tfSchema.ProtocolVersion = ""
	hash, err := ComputeSchemaHash(tfSchema.Plugin, &tfSchema.SchemaDetail)
	if err != nil {
		t.Fatalf("compute hash failed: %v", err)
	}
	tfSchema.SchemaHash = hash
	if err := ValidatePluginSchema(tfSchema, expectedPlugin, ProtocolTFProvider); err != nil {
		t.Fatalf("expected tf_provider with empty protocol_version to pass, got: %v", err)
	}

	// http rejects empty protocol_version
	if err := ValidatePluginSchema(tfSchema, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected http protocol with empty protocol_version to fail")
	}

	// schema_hash format invalid
	badFormat := newValidTestPluginSchema(t)
	badFormat.SchemaHash = "sha256:INVALID"
	if err := ValidatePluginSchema(badFormat, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected invalid schema_hash format to fail")
	}

	// unsupported schema_hash algorithm fails
	unsupportedAlgo := newValidTestPluginSchema(t)
	unsupportedAlgo.SchemaHash = "md5:" + strings.Repeat("a", 32)
	if err := ValidatePluginSchema(unsupportedAlgo, expectedPlugin, ProtocolHTTP); err == nil ||
		!strings.Contains(err.Error(), "unsupported schema_hash algorithm") {
		t.Fatalf("expected unsupported schema_hash algorithm to fail, got: %v", err)
	}

	// sha512 prefix algorithm is dispatched and validated properly
	sha512Schema := newValidTestPluginSchema(t)
	sha512Hash, err := ComputeSchemaHashByAlgo(SchemaHashAlgoSHA512, sha512Schema.Plugin, &sha512Schema.SchemaDetail)
	if err != nil {
		t.Fatalf("compute sha512 schema hash failed: %v", err)
	}
	sha512Schema.SchemaHash = sha512Hash
	if err := ValidatePluginSchema(sha512Schema, expectedPlugin, ProtocolHTTP); err != nil {
		t.Fatalf("expected valid sha512 schema_hash to pass, got: %v", err)
	}

	// schema_hash does not match SchemaDetail content
	badDigest := newValidTestPluginSchema(t)
	badDigest.SchemaHash = "sha256:" + strings.Repeat("a", 64)
	if err := ValidatePluginSchema(badDigest, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected mismatched schema_hash digest to fail")
	}

	// resource_schemas non-empty requires apply, plan, refresh
	missingPlan := newValidTestPluginSchema(t)
	missingPlan.Capabilities.Plan = false
	missingPlan.SchemaHash, _ = ComputeSchemaHash(missingPlan.Plugin, &missingPlan.SchemaDetail)
	if err := ValidatePluginSchema(missingPlan, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected missing plan capability with non-empty resource_schemas to fail")
	}

	// data_source_schemas empty forbids read_data=true
	emptyDSWithReadData := newValidTestPluginSchema(t)
	emptyDSWithReadData.DataSourceSchemas = nil
	emptyDSWithReadData.Capabilities.ReadData = true
	emptyDSWithReadData.SchemaHash, _ = ComputeSchemaHash(emptyDSWithReadData.Plugin,
		&emptyDSWithReadData.SchemaDetail)
	if err := ValidatePluginSchema(emptyDSWithReadData, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected read_data=true when data_source_schemas is empty to fail")
	}

	// stop=true requires async_apply=true
	stopWithoutAsync := newValidTestPluginSchema(t)
	stopWithoutAsync.Capabilities.AsyncApply = false
	stopWithoutAsync.SchemaHash, _ = ComputeSchemaHash(stopWithoutAsync.Plugin, &stopWithoutAsync.SchemaDetail)
	if err := ValidatePluginSchema(stopWithoutAsync, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected stop=true with async_apply=false to fail")
	}

	// nil provider.block fails
	nilProviderBlock := newValidTestPluginSchema(t)
	nilProviderBlock.ConfigSchema = &tfjson.Schema{}
	nilProviderBlock.SchemaHash, _ = ComputeSchemaHash(nilProviderBlock.Plugin, &nilProviderBlock.SchemaDetail)
	if err := ValidatePluginSchema(nilProviderBlock, expectedPlugin, ProtocolHTTP); err == nil {
		t.Fatalf("expected nil provider.block to fail")
	}
}
