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

package executor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"hcm/pkg/plugin/tfplugin5"
	"hcm/pkg/version"
)

func TestPluginExecutor_GetProviderSchema_Resources(t *testing.T) {
	resource := &tfplugin5.Schema{Version: 2, Block: &tfplugin5.Schema_Block{
		Attributes: []*tfplugin5.Schema_Attribute{stringAttr("name", false)},
		BlockTypes: []*tfplugin5.Schema_NestedBlock{{
			TypeName: "net",
			Nesting:  tfplugin5.Schema_NestedBlock_LIST,
			MinItems: 1,
			Block:    &tfplugin5.Schema_Block{Attributes: []*tfplugin5.Schema_Attribute{stringAttr("zone", false)}},
		}},
	}}
	fake := &fakePlugin{getSchema: func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
		return &tfplugin5.GetProviderSchema_Response{
			Provider: objectSchema(stringAttr("secret_id", true)),
			ResourceSchemas: map[string]*tfplugin5.Schema{
				"tcloud_cvm": resource,
			},
			ServerCapabilities: &tfplugin5.ServerCapabilities{PlanDestroy: true, MoveResourceState: true},
		}, nil
	}}
	executor := newAPIExecutor(fake)
	executor.version = ""

	got, err := executor.GetProviderSchema(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.ProtocolVersion != "1.0" || got.Plugin.Name != "tencentcloud" || got.Plugin.Version != "" {
		t.Fatalf("identity = %+v", got.Plugin)
	}
	attr := got.Provider.Block.Attributes["secret_id"]
	if attr.Type != "string" || !attr.Required || !attr.Sensitive {
		t.Fatalf("provider attribute = %+v", attr)
	}
	if got.Provider.Version != 0 || got.ResourceSchemas["tcloud_cvm"].Version != 2 {
		t.Fatalf("schema versions = provider %d resource %d", got.Provider.Version, got.ResourceSchemas["tcloud_cvm"].Version)
	}
	nested := got.ResourceSchemas["tcloud_cvm"].Block.BlockTypes["net"]
	if nested.Nesting != "list" || nested.MinItems != 1 || nested.Block.Attributes["zone"].Type != "string" {
		t.Fatalf("nested block = %+v", nested)
	}
	if len(got.ResourceSchemas["tcloud_cvm"].Block.Attributes) == 0 || got.ResourceSchemas["tcloud_cvm"].Block.BlockTypes == nil {
		t.Fatal("adapted block maps are missing")
	}
	for _, key := range []string{"apply", "plan", "validate", "import", "refresh"} {
		if !got.Capabilities[key] {
			t.Fatalf("missing capability %s: %v", key, got.Capabilities)
		}
	}
	for _, key := range []string{"read_data", "async_apply", "list", "plan_destroy", "move_resource_state"} {
		if _, ok := got.Capabilities[key]; ok {
			t.Fatalf("capability %s should be absent: %v", key, got.Capabilities)
		}
	}
	if executor.resourceSchemas["tcloud_cvm"] == nil || executor.providerSchema == nil {
		t.Fatal("protocol schemas were not kept")
	}
}

func TestPluginExecutor_GetProviderSchema_DataSources(t *testing.T) {
	fake := &fakePlugin{getSchema: func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
		return &tfplugin5.GetProviderSchema_Response{
			Provider: objectSchema(stringAttr("region", false)),
			DataSourceSchemas: map[string]*tfplugin5.Schema{
				"tcloud_image": objectSchema(stringAttr("id", false)),
			},
		}, nil
	}}
	got, err := newAPIExecutor(fake).GetProviderSchema(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Capabilities["read_data"] {
		t.Fatalf("capabilities = %v", got.Capabilities)
	}
	for _, key := range []string{"apply", "plan", "validate", "import", "refresh"} {
		if _, ok := got.Capabilities[key]; ok {
			t.Fatalf("capability %s should be absent: %v", key, got.Capabilities)
		}
	}
	if got.ResourceSchemas == nil || got.DataSourceSchemas == nil {
		t.Fatal("schema maps should be empty objects, not nil")
	}
}

func TestPluginExecutor_GetProviderSchema_Hash(t *testing.T) {
	schemaFor := func(attr string) *fakePlugin {
		return &fakePlugin{getSchema: func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
			return &tfplugin5.GetProviderSchema_Response{
				Provider: objectSchema(stringAttr("region", false)),
				ResourceSchemas: map[string]*tfplugin5.Schema{
					"tcloud_cvm": objectSchema(stringAttr(attr, false)),
				},
				Diagnostics: []*tfplugin5.Diagnostic{{
					Severity: tfplugin5.Diagnostic_WARNING,
					Summary:  "ignored by hash",
				}},
			}, nil
		}}
	}
	firstExec := newAPIExecutor(schemaFor("name"))
	first, err := firstExec.GetProviderSchema(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	firstExec.version = "v9"
	firstExec.name = "other"
	second, err := firstExec.GetProviderSchema(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first.SchemaHash == "" || first.SchemaHash != second.SchemaHash {
		t.Fatalf("hash changed without schema change: %s %s", first.SchemaHash, second.SchemaHash)
	}
	changed, err := newAPIExecutor(schemaFor("zone")).GetProviderSchema(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if changed.SchemaHash == first.SchemaHash {
		t.Fatal("hash stayed the same after an attribute change")
	}
}

func TestPluginExecutor_ConfigureProvider_LoadsSchema(t *testing.T) {
	fake := &fakePlugin{getSchema: resourceSchema(t)}
	loadEchoPrepare(fake)
	executor := newAPIExecutor(fake)
	executor.SetCodec(&jsonCodec{})
	ctx := context.Background()
	if _, err := executor.ConfigureProvider(ctx, map[string]any{"secret_id": "ak-test-value"}); err != nil {
		t.Fatal(err)
	}
	if fake.count("GetSchema") != 1 || fake.count("Configure") != 1 {
		t.Fatalf("schema=%d configure=%d", fake.count("GetSchema"), fake.count("Configure"))
	}
	if _, err := executor.ConfigureProvider(ctx, map[string]any{"secret_id": "ak-test-value"}); err != nil {
		t.Fatal(err)
	}
	if fake.count("GetSchema") != 1 {
		t.Fatalf("schema calls after a second configure = %d", fake.count("GetSchema"))
	}
	if _, err := executor.ReadResource(ctx, "tcloud_cvm", map[string]any{"id": "ins-1"}, nil); err != nil {
		t.Fatal(err)
	}
	if fake.count("ReadResource") != 1 {
		t.Fatalf("read calls = %d", fake.count("ReadResource"))
	}
}

func TestPluginExecutor_ConfigureProvider_SchemaError(t *testing.T) {
	fake := &fakePlugin{getSchema: func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
		return nil, errors.New("schema down")
	}}
	executor := newAPIExecutor(fake)
	executor.SetCodec(&jsonCodec{})
	if _, err := executor.ConfigureProvider(context.Background(), map[string]any{"secret_id": "ak-test-value"}); err == nil || err.Error() != "schema down" {
		t.Fatalf("err = %v", err)
	}
	if fake.count("PrepareProviderConfig") != 0 || fake.count("Configure") != 0 {
		t.Fatalf("prepare=%d configure=%d", fake.count("PrepareProviderConfig"), fake.count("Configure"))
	}
}

func TestPluginExecutor_ConfigureProvider_PrepareError(t *testing.T) {
	fake := &fakePlugin{
		getSchema: providerOnlySchema(t),
		prepare: func(*tfplugin5.PrepareProviderConfig_Request) (*tfplugin5.PrepareProviderConfig_Response, error) {
			return &tfplugin5.PrepareProviderConfig_Response{Diagnostics: errorDiagnostic("bad config")}, nil
		},
	}
	executor := newAPIExecutor(fake)
	executor.SetCodec(&jsonCodec{})
	if _, err := executor.GetProviderSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err := executor.ConfigureProvider(context.Background(), map[string]any{"secret_id": "ak-test-value"})
	if !errors.Is(err, errDiagnostic) {
		t.Fatalf("err = %v", err)
	}
	if result == nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].GetSeverity() != tfplugin5.Diagnostic_ERROR {
		t.Fatalf("diagnostics = %+v", result)
	}
	if fake.count("Configure") != 0 {
		t.Fatalf("configure calls = %d", fake.count("Configure"))
	}
}

func TestPluginExecutor_ConfigureProvider_Config(t *testing.T) {
	var configured *tfplugin5.Configure_Request
	fake := &fakePlugin{
		getSchema: providerOnlySchema(t),
		configure: func(in *tfplugin5.Configure_Request) (*tfplugin5.Configure_Response, error) {
			configured = in
			return &tfplugin5.Configure_Response{}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := newAPIExecutor(fake)
	executor.SetCodec(&jsonCodec{})
	if _, err := executor.GetProviderSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := executor.ConfigureProvider(context.Background(), map[string]any{"secret_id": "ak-test-value"}); err != nil {
		t.Fatal(err)
	}
	if configured.GetTerraformVersion() != version.VERSION {
		t.Fatalf("terraform version = %s", configured.GetTerraformVersion())
	}
	if !strings.Contains(string(configured.GetConfig().GetJson()), `"secret_id":"ak-test-value"`) {
		t.Fatalf("configure config = %s", configured.GetConfig().GetJson())
	}
}

type jsonCodec struct {
	err   error
	calls int
}

func (c *jsonCodec) Encode(_ *tfplugin5.Schema, value any) (*tfplugin5.DynamicValue, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &tfplugin5.DynamicValue{Json: raw}, nil
}

func (c *jsonCodec) Decode(_ *tfplugin5.Schema, value *tfplugin5.DynamicValue) (any, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	var out any
	if err := json.Unmarshal(value.GetJson(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func stringAttr(name string, sensitive bool) *tfplugin5.Schema_Attribute {
	return &tfplugin5.Schema_Attribute{
		Name:      name,
		Type:      []byte(`"string"`),
		Optional:  !sensitive,
		Required:  sensitive,
		Sensitive: sensitive,
	}
}

func objectSchema(attrs ...*tfplugin5.Schema_Attribute) *tfplugin5.Schema {
	return &tfplugin5.Schema{Block: &tfplugin5.Schema_Block{Attributes: attrs}}
}

func newAPIExecutor(fake *fakePlugin) *PluginExecutor {
	return &PluginExecutor{rpc: fake, name: "tencentcloud"}
}

func loadEchoPrepare(fake *fakePlugin) {
	fake.prepare = func(in *tfplugin5.PrepareProviderConfig_Request) (*tfplugin5.PrepareProviderConfig_Response, error) {
		return &tfplugin5.PrepareProviderConfig_Response{PreparedConfig: in.GetConfig()}, nil
	}
}

func configureExecutor(t *testing.T, executor *PluginExecutor) {
	t.Helper()
	if _, err := executor.ConfigureProvider(context.Background(), map[string]any{"secret_id": "ak-test-value"}); err != nil {
		t.Fatal(err)
	}
}

func errorDiagnostic(summary string) []*tfplugin5.Diagnostic {
	return []*tfplugin5.Diagnostic{{Severity: tfplugin5.Diagnostic_ERROR, Summary: summary}}
}

func providerOnlySchema(t *testing.T) func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
	t.Helper()
	return func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
		return &tfplugin5.GetProviderSchema_Response{
			Provider: objectSchema(stringAttr("secret_id", true)),
		}, nil
	}
}

func resourceSchema(t *testing.T) func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
	t.Helper()
	return func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
		return &tfplugin5.GetProviderSchema_Response{
			Provider: objectSchema(stringAttr("secret_id", true)),
			ResourceSchemas: map[string]*tfplugin5.Schema{
				"tcloud_cvm": objectSchema(stringAttr("id", false)),
			},
		}, nil
	}
}

func dataSourceSchema(t *testing.T) func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
	t.Helper()
	return func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {
		return &tfplugin5.GetProviderSchema_Response{
			Provider: objectSchema(stringAttr("secret_id", true)),
			DataSourceSchemas: map[string]*tfplugin5.Schema{
				"tcloud_image": objectSchema(stringAttr("id", false)),
			},
		}, nil
	}
}

func preparedExecutor(t *testing.T, fake *fakePlugin, codec ValueCodec) *PluginExecutor {
	t.Helper()
	executor := newAPIExecutor(fake)
	executor.SetCodec(codec)
	if _, err := executor.GetProviderSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	configureExecutor(t, executor)
	return executor
}
