/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package codec

import (
	"math/big"
	"strings"
	"testing"

	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestEncodeNullAndUnknown(t *testing.T) {
	schema := objectSchema(
		attr("name", `"string"`, false),
		attr("note", `"string"`, false),
		computedAttr("id", `"string"`),
	)
	absent, err := Codec{}.Encode(schema, map[string]any{"name": "vm"})
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := Codec{}.Encode(schema, map[string]any{"name": "vm", "note": nil})
	if err != nil {
		t.Fatal(err)
	}
	if string(absent.GetMsgpack()) != string(explicit.GetMsgpack()) || len(absent.GetJson()) != 0 {
		t.Fatalf("nil and absent differ or json was set")
	}
	decoded, err := Codec{}.Decode(schema, absent)
	if err != nil {
		t.Fatal(err)
	}
	object := decoded.(map[string]any)
	if object["note"] != nil || object["id"] != nil {
		t.Fatalf("missing computed and optional = %#v", object)
	}

	unknown, err := Codec{}.Encode(schema, map[string]any{"name": "vm", "id": Unknown})
	if err != nil || unknown == nil {
		t.Fatal(err)
	}
	if string(unknown.GetMsgpack()) == string(absent.GetMsgpack()) {
		t.Fatal("unknown encoded as null")
	}
	state, err := Codec{}.Decode(schema, unknown)
	if err != nil {
		t.Fatal(err)
	}
	if state.(map[string]any)["id"] != Unknown {
		t.Fatalf("id = %#v", state)
	}
	root, err := Codec{}.Encode(schema, nil)
	if err != nil {
		t.Fatal(err)
	}
	var nilMap map[string]any
	nilRoot, err := Codec{}.Encode(schema, nilMap)
	if err != nil || string(nilRoot.GetMsgpack()) != string(root.GetMsgpack()) {
		t.Fatalf("nil map root = %v err=%v", nilRoot.GetMsgpack(), err)
	}
	decodedRoot, err := Codec{}.Decode(schema, root)
	if err != nil || decodedRoot != nil {
		t.Fatalf("root = %#v err=%v", decodedRoot, err)
	}
	rootType, err := schemaObject(schema)
	if err != nil {
		t.Fatal(err)
	}
	rootValue, err := tftypes.ValueFromMsgPack(root.GetMsgpack(), rootType)
	if err != nil || !rootValue.IsNull() || !rootValue.IsKnown() {
		t.Fatalf("root null=%v known=%v err=%v", rootValue.IsNull(), rootValue.IsKnown(), err)
	}

	id := attributeValue(t, schema, unknown.GetMsgpack(), "id")
	if id.IsNull() || id.IsKnown() {
		t.Fatalf("id null=%v known=%v", id.IsNull(), id.IsKnown())
	}
}

func TestEncodeNestedBlocks(t *testing.T) {
	schema := &tfplugin5.Schema{Block: &tfplugin5.Schema_Block{
		BlockTypes: []*tfplugin5.Schema_NestedBlock{
			nested("network", tfplugin5.Schema_NestedBlock_LIST),
			nested("zone", tfplugin5.Schema_NestedBlock_SINGLE),
			nested("group", tfplugin5.Schema_NestedBlock_GROUP),
			nested("labels", tfplugin5.Schema_NestedBlock_MAP),
			nested("tags", tfplugin5.Schema_NestedBlock_SET),
		},
	}}
	withLimits := &tfplugin5.Schema{Version: 9, Block: &tfplugin5.Schema_Block{
		BlockTypes: []*tfplugin5.Schema_NestedBlock{
			withItems(nested("network", tfplugin5.Schema_NestedBlock_LIST), 2, 4),
			withItems(nested("zone", tfplugin5.Schema_NestedBlock_SINGLE), 1, 1),
			withItems(nested("group", tfplugin5.Schema_NestedBlock_GROUP), 1, 1),
			withItems(nested("labels", tfplugin5.Schema_NestedBlock_MAP), 0, 3),
			withItems(nested("tags", tfplugin5.Schema_NestedBlock_SET), 1, 8),
		},
	}}
	value := map[string]any{
		"network": []any{map[string]any{"id": "net-1"}},
		"zone":    map[string]any{"id": "az-1"},
		"group":   map[string]any{"id": "grp-1"},
		"labels":  map[string]any{"env": map[string]any{"id": "prod"}},
		"tags":    []any{map[string]any{"id": "a"}, map[string]any{"id": "b"}},
	}
	encoded, err := Codec{}.Encode(schema, value)
	if err != nil || encoded == nil {
		t.Fatal(err)
	}
	limited, err := Codec{}.Encode(withLimits, value)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded.GetMsgpack()) != string(limited.GetMsgpack()) {
		t.Fatal("min_items, max_items, or schema version changed the encoding")
	}
}

func TestEncodeRejectsInvalidValues(t *testing.T) {
	secret := "secret-value-should-not-leak"
	schema := objectSchema(attr("name", `"bool"`, false), attr("items", `["list","string"]`, false))
	extra, err := Codec{}.Encode(schema, map[string]any{"extra": secret})
	if err == nil || extra != nil || !strings.Contains(err.Error(), "extra") || strings.Contains(err.Error(), secret) {
		t.Fatalf("extra = %v err=%v", extra, err)
	}
	mismatched, err := Codec{}.Encode(schema, map[string]any{"name": secret})
	if err == nil || mismatched != nil || !strings.Contains(err.Error(), "name") || strings.Contains(err.Error(), secret) {
		t.Fatalf("mismatch = %v err=%v", mismatched, err)
	}
	element, err := Codec{}.Encode(schema, map[string]any{"items": []any{true}})
	if err == nil || element != nil || !strings.Contains(err.Error(), "items[0]") {
		t.Fatalf("element = %v err=%v", element, err)
	}

	dynamic, err := Codec{}.Encode(objectSchema(attr("payload", `"dynamic"`, false)), map[string]any{})
	if err == nil || dynamic != nil || !strings.Contains(err.Error(), "payload") {
		t.Fatalf("dynamic = %v err=%v", dynamic, err)
	}
	invalid, err := Codec{}.Encode(&tfplugin5.Schema{Block: &tfplugin5.Schema_Block{
		BlockTypes: []*tfplugin5.Schema_NestedBlock{nested("network", tfplugin5.Schema_NestedBlock_INVALID)},
	}}, map[string]any{})
	if err == nil || invalid != nil || !strings.Contains(err.Error(), "network") {
		t.Fatalf("invalid nesting = %v err=%v", invalid, err)
	}
	conflict, err := Codec{}.Encode(&tfplugin5.Schema{Block: &tfplugin5.Schema_Block{
		Attributes: []*tfplugin5.Schema_Attribute{attr("network", `"string"`, false)},
		BlockTypes: []*tfplugin5.Schema_NestedBlock{nested("network", tfplugin5.Schema_NestedBlock_LIST)},
	}}, map[string]any{"network": "net"})
	if err == nil || conflict != nil || !strings.Contains(err.Error(), "network") {
		t.Fatalf("conflict = %v err=%v", conflict, err)
	}
}

func TestRoundTrip(t *testing.T) {
	schema := objectSchema(
		attr("name", `"string"`, false),
		attr("cpu", `"number"`, false),
		attr("ready", `"bool"`, false),
		attr("disks", `["list","string"]`, false),
		attr("zones", `["set","string"]`, false),
	)
	encoded, err := Codec{}.Encode(schema, map[string]any{
		"name":  "vm-1",
		"cpu":   big.NewFloat(2),
		"ready": true,
		"disks": []any{"data", "root"},
		"zones": []any{"b", "a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Codec{}.Decode(schema, encoded)
	if err != nil {
		t.Fatal(err)
	}
	object := decoded.(map[string]any)
	if object["name"] != "vm-1" || object["ready"] != true {
		t.Fatalf("object = %#v", object)
	}
	if object["cpu"].(*big.Float).Cmp(big.NewFloat(2)) != 0 {
		t.Fatalf("cpu = %v", object["cpu"])
	}
	disks := object["disks"].([]any)
	if len(disks) != 2 || disks[0] != "data" || disks[1] != "root" {
		t.Fatalf("disks = %#v", disks)
	}
	zones := object["zones"].([]any)
	if !sameStrings(zones, "a", "b") {
		t.Fatalf("zones = %#v", zones)
	}

	fromFloat, err := Codec{}.Encode(schema, map[string]any{"cpu": float64(2)})
	if err != nil {
		t.Fatal(err)
	}
	floatDecoded, err := Codec{}.Decode(schema, fromFloat)
	if err != nil {
		t.Fatal(err)
	}
	if floatDecoded.(map[string]any)["cpu"].(*big.Float).Cmp(big.NewFloat(2)) != 0 {
		t.Fatal("float64 did not round-trip as *big.Float")
	}
}

func TestDecodeRejectsMissingSchemaAndJSON(t *testing.T) {
	schema := objectSchema(attr("name", `"string"`, false))
	got, err := Codec{}.Decode(nil, &tfplugin5.DynamicValue{Msgpack: []byte{1}})
	if err == nil || got != nil {
		t.Fatalf("nil schema = %#v err=%v", got, err)
	}
	got, err = Codec{}.Decode(&tfplugin5.Schema{}, &tfplugin5.DynamicValue{Msgpack: []byte{1}})
	if err == nil || got != nil {
		t.Fatalf("empty block = %#v err=%v", got, err)
	}
	encoded, err := Codec{}.Encode(schema, map[string]any{"name": "vm"})
	if err != nil {
		t.Fatal(err)
	}
	jsonOnly := &tfplugin5.DynamicValue{Json: encoded.GetMsgpack()}
	got, err = Codec{}.Decode(schema, jsonOnly)
	if err == nil || got != nil {
		t.Fatalf("json only = %#v err=%v", got, err)
	}
}

func objectSchema(attrs ...*tfplugin5.Schema_Attribute) *tfplugin5.Schema {
	return &tfplugin5.Schema{Block: &tfplugin5.Schema_Block{Attributes: attrs}}
}

func attr(name, typ string, required bool) *tfplugin5.Schema_Attribute {
	return &tfplugin5.Schema_Attribute{
		Name:     name,
		Type:     []byte(typ),
		Required: required,
		Optional: !required,
	}
}

func computedAttr(name, typ string) *tfplugin5.Schema_Attribute {
	return &tfplugin5.Schema_Attribute{Name: name, Type: []byte(typ), Computed: true}
}

func nested(name string, mode tfplugin5.Schema_NestedBlock_NestingMode) *tfplugin5.Schema_NestedBlock {
	return &tfplugin5.Schema_NestedBlock{
		TypeName: name,
		Nesting:  mode,
		Block: &tfplugin5.Schema_Block{Attributes: []*tfplugin5.Schema_Attribute{
			attr("id", `"string"`, false),
		}},
	}
}

func withItems(block *tfplugin5.Schema_NestedBlock, min, max int64) *tfplugin5.Schema_NestedBlock {
	block.MinItems = min
	block.MaxItems = max
	return block
}

func attributeValue(t *testing.T, schema *tfplugin5.Schema, raw []byte, name string) tftypes.Value {
	t.Helper()
	typ, err := schemaObject(schema)
	if err != nil {
		t.Fatal(err)
	}
	value, err := tftypes.ValueFromMsgPack(raw, typ)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]tftypes.Value
	if err = value.As(&object); err != nil {
		t.Fatal(err)
	}
	return object[name]
}

func sameStrings(got []any, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := map[string]int{}
	for _, item := range got {
		text, ok := item.(string)
		if !ok {
			return false
		}
		seen[text]++
	}
	for _, item := range want {
		seen[item]--
		if seen[item] < 0 {
			return false
		}
	}
	return true
}
