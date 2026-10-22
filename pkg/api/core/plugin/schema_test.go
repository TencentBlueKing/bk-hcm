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
	raw := RawSchema{
		ProtocolVersion: "1.0",
		Capabilities: map[string]bool{
			CapabilityApply:      true,
			CapabilityPlan:       true,
			CapabilityRefresh:    true,
			CapabilityReadData:   true,
			CapabilityAsyncApply: true,
			CapabilityStop:       true,
			CapabilityValidate:   true,
		},
		Provider: &tfjson.Schema{
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
			"tencentcloud_Availability_zones": {
				Version: 0,
				Block:   &tfjson.SchemaBlock{},
			},
		},
	}

	hash, err := ComputeRawSchemaHash(&raw)
	if err != nil {
		t.Fatalf("compute raw schema hash failed: %v", err)
	}

	return &PluginSchema{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		SchemaHash:    hash,
		RawSchema:     raw,
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
		"plugin_name", "plugin_version", "schema_hash",
		"protocol_version", "capabilities", "provider",
		"resource_schemas", "data_source_schemas",
	}
	for _, key := range expectedTopKeys {
		if _, ok := topMap[key]; !ok {
			t.Errorf("expected key %q at top level of serialized PluginSchema, got map: %v", key, topMap)
		}
	}
	if _, hasRaw := topMap["RawSchema"]; hasRaw {
		t.Errorf("unexpected nested RawSchema key in serialized PluginSchema")
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
	if roundTrip.PluginName != ps.PluginName || roundTrip.SchemaHash != ps.SchemaHash ||
		roundTrip.ProtocolVersion != ps.ProtocolVersion {
		t.Fatalf("round-trip mismatch: got %+v", roundTrip)
	}

	// 3. Test RawSchema standalone serialization excludes outer columns
	rawBytes, err := json.Marshal(ps.RawSchema)
	if err != nil {
		t.Fatalf("json.Marshal RawSchema failed: %v", err)
	}
	var rawMap map[string]interface{}
	if err := json.Unmarshal(rawBytes, &rawMap); err != nil {
		t.Fatalf("json.Unmarshal RawSchema map failed: %v", err)
	}
	for _, forbiddenKey := range []string{"plugin_name", "plugin_version", "schema_hash"} {
		if _, exists := rawMap[forbiddenKey]; exists {
			t.Errorf("RawSchema JSON must not contain %q", forbiddenKey)
		}
	}
}

func TestValidatePluginSchemaAndCapabilities(t *testing.T) {
	validSchema := newValidTestPluginSchema(t)
	if err := ValidatePluginSchema(validSchema, "tencentcloud", "1.81.5", ProtocolHTTP); err != nil {
		t.Fatalf("expected valid schema to pass validation, got err: %v", err)
	}

	// tf_provider allows empty protocol_version
	tfSchema := newValidTestPluginSchema(t)
	tfSchema.ProtocolVersion = ""
	hash, err := ComputeRawSchemaHash(&tfSchema.RawSchema)
	if err != nil {
		t.Fatalf("compute hash failed: %v", err)
	}
	tfSchema.SchemaHash = hash
	if err := ValidatePluginSchema(tfSchema, "tencentcloud", "1.81.5", ProtocolTFProvider); err != nil {
		t.Fatalf("expected tf_provider with empty protocol_version to pass, got: %v", err)
	}

	// http rejects empty protocol_version
	if err := ValidatePluginSchema(tfSchema, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected http protocol with empty protocol_version to fail")
	}

	// schema_hash format invalid
	badFormat := newValidTestPluginSchema(t)
	badFormat.SchemaHash = "sha256:INVALID"
	if err := ValidatePluginSchema(badFormat, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected invalid schema_hash format to fail")
	}

	// schema_hash does not match RawSchema content
	badDigest := newValidTestPluginSchema(t)
	badDigest.SchemaHash = "sha256:" + strings.Repeat("a", 64)
	if err := ValidatePluginSchema(badDigest, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected mismatched schema_hash digest to fail")
	}

	// resource_schemas non-empty requires apply, plan, refresh
	missingPlan := newValidTestPluginSchema(t)
	missingPlan.Capabilities[CapabilityPlan] = false
	missingPlan.SchemaHash, _ = ComputeRawSchemaHash(&missingPlan.RawSchema)
	if err := ValidatePluginSchema(missingPlan, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected missing plan capability with non-empty resource_schemas to fail")
	}

	// data_source_schemas empty forbids read_data declaration
	emptyDSWithReadData := newValidTestPluginSchema(t)
	emptyDSWithReadData.DataSourceSchemas = nil
	emptyDSWithReadData.Capabilities[CapabilityReadData] = false
	emptyDSWithReadData.SchemaHash, _ = ComputeRawSchemaHash(&emptyDSWithReadData.RawSchema)
	if err := ValidatePluginSchema(emptyDSWithReadData, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected read_data declared when data_source_schemas is empty to fail")
	}

	// stop=true requires async_apply=true
	stopWithoutAsync := newValidTestPluginSchema(t)
	stopWithoutAsync.Capabilities[CapabilityAsyncApply] = false
	stopWithoutAsync.SchemaHash, _ = ComputeRawSchemaHash(&stopWithoutAsync.RawSchema)
	if err := ValidatePluginSchema(stopWithoutAsync, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected stop=true with async_apply=false to fail")
	}

	// nil provider.block fails
	nilProviderBlock := newValidTestPluginSchema(t)
	nilProviderBlock.Provider = &tfjson.Schema{}
	nilProviderBlock.SchemaHash, _ = ComputeRawSchemaHash(&nilProviderBlock.RawSchema)
	if err := ValidatePluginSchema(nilProviderBlock, "tencentcloud", "1.81.5", ProtocolHTTP); err == nil {
		t.Fatalf("expected nil provider.block to fail")
	}
}
