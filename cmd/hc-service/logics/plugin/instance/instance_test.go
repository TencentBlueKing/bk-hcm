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

package instance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	hcplugin "hcm/pkg/api/hc-service/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableplugin "hcm/pkg/dal/table/plugin"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	restcli "hcm/pkg/rest/client"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

func buildTestProviderBlock() *tfjson.SchemaBlock {
	return &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"secret_id": {
				AttributeType: cty.String,
				Required:      true,
				Sensitive:     false,
			},
			"secret_key": {
				AttributeType: cty.String,
				Required:      true,
				Sensitive:     true,
			},
			"region": {
				AttributeType: cty.String,
				Optional:      true,
				Sensitive:     false,
			},
			"account_id": {
				AttributeType: cty.String,
				Computed:      true,
			},
		},
		NestedBlocks: map[string]*tfjson.SchemaBlockType{
			"assume_role": {
				NestingMode: tfjson.SchemaNestingModeList,
				Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"role_arn": {
							AttributeType: cty.String,
							Required:      true,
							Sensitive:     false,
						},
						"session_name": {
							AttributeType: cty.String,
							Optional:      true,
							Sensitive:     false,
						},
						"external_id": {
							AttributeType: cty.String,
							Optional:      true,
							Sensitive:     true,
						},
					},
				},
			},
		},
	}
}

func TestSplitAndMergeProperties(t *testing.T) {
	block := buildTestProviderBlock()
	inputProps := map[string]any{
		"secret_id":  "AKID-001",
		"secret_key": "SK-001",
		"region":     "ap-guangzhou",
		"assume_role": []any{
			map[string]any{
				"role_arn":     "qcs::cam::uin/1:roleName/r1",
				"session_name": "s1",
			},
			map[string]any{
				"role_arn":    "qcs::cam::uin/1:roleName/r2",
				"external_id": "ext-secret-2",
			},
		},
	}

	attrs, secretAttrs, err := splitAndValidateProperties(block, inputProps)
	if err != nil {
		t.Fatalf("splitAndValidateProperties failed: %v", err)
	}

	wantAttrs := map[string]any{
		"secret_id": "AKID-001",
		"region":    "ap-guangzhou",
		"assume_role": []any{
			map[string]any{
				"role_arn":     "qcs::cam::uin/1:roleName/r1",
				"session_name": "s1",
			},
			map[string]any{
				"role_arn": "qcs::cam::uin/1:roleName/r2",
			},
		},
	}
	wantSecretAttrs := map[string]any{
		"secret_key": "SK-001",
		"assume_role": []any{
			map[string]any{},
			map[string]any{
				"external_id": "ext-secret-2",
			},
		},
	}

	if !reflect.DeepEqual(attrs, wantAttrs) {
		t.Fatalf("attrs mismatch:\ngot:  %#v\nwant: %#v", attrs, wantAttrs)
	}
	if !reflect.DeepEqual(secretAttrs, wantSecretAttrs) {
		t.Fatalf("secretAttrs mismatch:\ngot:  %#v\nwant: %#v", secretAttrs, wantSecretAttrs)
	}

	merged := mergeInstanceProperties(attrs, secretAttrs)
	if !reflect.DeepEqual(merged, inputProps) {
		t.Fatalf("merged properties mismatch:\ngot:  %#v\nwant: %#v", merged, inputProps)
	}
}

func TestSplitAndValidatePropertiesRejectInvalid(t *testing.T) {
	block := buildTestProviderBlock()

	// 1. Unknown top-level field
	_, _, err := splitAndValidateProperties(block, map[string]any{
		"secret_id":     "AKID",
		"secret_key":    "SK",
		"unknown_field": "val",
	})
	if err == nil || !strings.Contains(err.Error(), "unknown property") {
		t.Fatalf("expected unknown property error, got: %v", err)
	}

	// 2. Computed read-only field
	_, _, err = splitAndValidateProperties(block, map[string]any{
		"secret_id":  "AKID",
		"secret_key": "SK",
		"account_id": "10001",
	})
	if err == nil || !strings.Contains(err.Error(), "computed read-only") {
		t.Fatalf("expected computed read-only error, got: %v", err)
	}

	// 3. Missing required field
	_, _, err = splitAndValidateProperties(block, map[string]any{
		"secret_id": "AKID",
	})
	if err == nil || !strings.Contains(err.Error(), "required property") {
		t.Fatalf("expected required property error, got: %v", err)
	}

	// 4. Invalid attribute value type
	_, _, err = splitAndValidateProperties(block, map[string]any{
		"secret_id":  "AKID",
		"secret_key": "SK",
		"region":     12345,
	})
	if err == nil || !strings.Contains(err.Error(), "expects string") {
		t.Fatalf("expected type mismatch error, got: %v", err)
	}
}

func TestSplitAndValidateAttributeNestedTypeAndCtyComposite(t *testing.T) {
	block := &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"endpoints": {
				Required: true,
				AttributeNestedType: &tfjson.SchemaNestedAttributeType{
					NestingMode: tfjson.SchemaNestingModeSingle,
					Attributes: map[string]*tfjson.SchemaAttribute{
						"sts": {
							AttributeType: cty.String,
							Required:      true,
						},
						"token": {
							AttributeType: cty.String,
							Optional:      true,
							Sensitive:     true,
						},
					},
				},
			},
			"proxy_config": {
				Optional: true,
				AttributeType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
					"host": cty.String,
					"port": cty.Number,
				}, []string{"port"}),
			},
			"retry_tuple": {
				Optional:      true,
				AttributeType: cty.Tuple([]cty.Type{cty.String, cty.Number}),
			},
		},
	}

	input := map[string]any{
		"endpoints": map[string]any{
			"sts":   "https://sts.example.com",
			"token": "secret-token-01",
		},
		"proxy_config": map[string]any{
			"host": "127.0.0.1",
			"port": 8080,
		},
		"retry_tuple": []any{"exponential", 3},
	}

	attrs, secretAttrs, err := splitAndValidateProperties(block, input)
	if err != nil {
		t.Fatalf("splitAndValidateProperties failed: %v", err)
	}

	wantAttrs := map[string]any{
		"endpoints": map[string]any{
			"sts": "https://sts.example.com",
		},
		"proxy_config": map[string]any{
			"host": "127.0.0.1",
			"port": 8080,
		},
		"retry_tuple": []any{"exponential", 3},
	}
	wantSecretAttrs := map[string]any{
		"endpoints": map[string]any{
			"token": "secret-token-01",
		},
	}

	if !reflect.DeepEqual(attrs, wantAttrs) {
		t.Fatalf("attrs mismatch:\ngot:  %#v\nwant: %#v", attrs, wantAttrs)
	}
	if !reflect.DeepEqual(secretAttrs, wantSecretAttrs) {
		t.Fatalf("secretAttrs mismatch:\ngot:  %#v\nwant: %#v", secretAttrs, wantSecretAttrs)
	}
	if merged := mergeInstanceProperties(attrs, secretAttrs); !reflect.DeepEqual(merged, input) {
		t.Fatalf("merged mismatch:\ngot:  %#v\nwant: %#v", merged, input)
	}

	// Reject missing required sub-attribute in nested_type
	_, _, err = splitAndValidateProperties(block, map[string]any{
		"endpoints": map[string]any{"token": "secret-token-01"},
	})
	if err == nil || !strings.Contains(err.Error(), `required property "endpoints.sts" is missing`) {
		t.Fatalf("expected missing endpoints.sts error, got: %v", err)
	}

	// Reject unknown attribute in cty.Object
	_, _, err = splitAndValidateProperties(block, map[string]any{
		"endpoints":    map[string]any{"sts": "https://sts.example.com"},
		"proxy_config": map[string]any{"host": "127.0.0.1", "extra": "bad"},
	})
	if err == nil || !strings.Contains(err.Error(), `unknown property "proxy_config.extra"`) {
		t.Fatalf("expected unknown proxy_config.extra error, got: %v", err)
	}

	// Reject wrong tuple length
	_, _, err = splitAndValidateProperties(block, map[string]any{
		"endpoints":   map[string]any{"sts": "https://sts.example.com"},
		"retry_tuple": []any{"only-one"},
	})
	if err == nil || !strings.Contains(err.Error(), "expects 2 tuple elements") {
		t.Fatalf("expected tuple length error, got: %v", err)
	}
}

type mockDataServiceEnv struct {
	pluginsEnabled bool
	instanceRecord *coreplugin.IacPluginInstance
	lastCreateReq  *dataplugin.IacPluginInstanceBatchCreateReq
	lastUpdateReq  *dataplugin.IacPluginInstanceBatchUpdateReq
	deleted        bool
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

func newMockPluginInstanceLogic(t *testing.T, env *mockDataServiceEnv) *PluginInstanceManager {
	t.Helper()

	pluginInfo := coreplugin.PluginInfo{
		Name:    "tencentcloud",
		Version: "1.81.5",
	}
	schemaDetail := coreplugin.SchemaDetail{
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema: &tfjson.Schema{
				Version: 0,
				Block:   buildTestProviderBlock(),
			},
		},
		ProtocolVersion: "1.0",
		Capabilities: coreplugin.Capabilities{
			Apply:   true,
			Plan:    true,
			Refresh: true,
		},
	}
	schemaHash, err := coreplugin.ComputeSchemaHash(pluginInfo, &schemaDetail)
	if err != nil {
		t.Fatalf("compute schema hash failed: %v", err)
	}
	schemaDetailBytes, err := json.Marshal(schemaDetail)
	if err != nil {
		t.Fatalf("marshal schema detail failed: %v", err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/iac_plugins/list"):
			var details []coreplugin.IaCPlugin
			if env.pluginsEnabled {
				details = []coreplugin.IaCPlugin{{
					ID:            "p-1",
					PluginName:    "tencentcloud",
					PluginVersion: "1.81.5",
					Protocol:      enumor.IaCPluginProtocolHTTP,
					State:         enumor.IaCPluginStateEnabled,
				}}
			}
			writeOKResp(w, &dataplugin.IaCPluginListResult{Details: details})

		case strings.HasSuffix(path, "/iac_plugin_schema/list"):
			writeOKResp(w, &dataplugin.IacPluginSchemaListResult{
				Details: []tableplugin.IacPluginSchemaTable{{
					ID:            "ps-1",
					PluginName:    "tencentcloud",
					PluginVersion: "1.81.5",
					SchemaHash:    schemaHash,
					SchemaDetail:  tabletypes.JsonField(schemaDetailBytes),
				}},
			})

		case strings.HasSuffix(path, "/iac_plugin_instance/batch/create"):
			req := new(dataplugin.IacPluginInstanceBatchCreateReq)
			_ = json.NewDecoder(r.Body).Decode(req)
			env.lastCreateReq = req
			if len(req.Items) > 0 {
				item := req.Items[0]
				env.instanceRecord = &coreplugin.IacPluginInstance{
					ID:          "pi-001",
					PluginName:  item.PluginName,
					Name:        item.Name,
					Attrs:       item.Attrs,
					SecretAttrs: item.SecretAttrs,
					Extension:   item.Extension,
					Memo:        item.Memo,
					TenantID:    "default",
				}
			}
			writeOKResp(w, &core.BatchCreateResult{IDs: []string{"pi-001"}})

		case strings.HasSuffix(path, "/iac_plugin_instance/batch/update"):
			req := new(dataplugin.IacPluginInstanceBatchUpdateReq)
			_ = json.NewDecoder(r.Body).Decode(req)
			env.lastUpdateReq = req
			if env.instanceRecord != nil && len(req.Items) > 0 {
				item := req.Items[0]
				if len(item.Name) > 0 {
					env.instanceRecord.Name = item.Name
				}
				if item.Memo != nil {
					env.instanceRecord.Memo = item.Memo
				}
				if item.Attrs != nil {
					env.instanceRecord.Attrs = item.Attrs
				}
				if item.SecretAttrs != nil {
					env.instanceRecord.SecretAttrs = item.SecretAttrs
				}
				if item.Extension != nil {
					env.instanceRecord.Extension = item.Extension
				}
			}
			writeOKResp(w, nil)

		case strings.HasSuffix(path, "/iac_plugin_instance/batch"):
			env.deleted = true
			env.instanceRecord = nil
			writeOKResp(w, nil)

		case strings.HasSuffix(path, "/iac_plugin_instance/list"):
			var details []coreplugin.IacPluginInstance
			if env.instanceRecord != nil {
				details = append(details, *env.instanceRecord)
			}
			writeOKResp(w, &dataplugin.IacPluginInstanceListResult{
				Count:   uint64(len(details)),
				Details: details,
			})
		}
	}

	cap := &restcli.Capability{
		Client:   &inMemoryHTTPClient{handler: handler},
		Discover: fakeDiscovery{},
	}
	dataCli := dataservice.NewClient(cap, "v1")
	schemaMgr := schema.NewSchemaManager(dataCli, time.Minute)
	return NewPluginInstanceManager(dataCli, schemaMgr)
}

func writeOKResp(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    errf.OK,
		"message": "",
		"data":    data,
	})
}

func TestCreatePluginInstanceRejectDisabledPlugin(t *testing.T) {
	env := &mockDataServiceEnv{pluginsEnabled: false}
	logic := newMockPluginInstanceLogic(t, env)
	kt := kit.New()

	_, err := logic.CreatePluginInstance(kt, &hcplugin.PluginInstanceCreateReq{
		PluginName: "tencentcloud",
		Name:       "prod",
		Properties: map[string]any{
			"secret_id":  "AKID",
			"secret_key": "SK",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "state=enabled not found") {
		t.Fatalf("expected active plugin not found error, got: %v", err)
	}
}

func TestPluginInstanceCRUDFullOverwriteAndMetadataPartialUpdate(t *testing.T) {
	env := &mockDataServiceEnv{pluginsEnabled: true}
	logic := newMockPluginInstanceLogic(t, env)
	kt := kit.New()

	// 1. Create instance
	createRes, err := logic.CreatePluginInstance(kt, &hcplugin.PluginInstanceCreateReq{
		PluginName: "tencentcloud",
		Name:       "prod-gz",
		Properties: map[string]any{
			"secret_id":  "AKID-1",
			"secret_key": "SK-1",
			"region":     "ap-guangzhou",
		},
		Extension: map[string]any{"bk_biz_id": float64(1001)},
	})
	if err != nil || createRes.ID != "pi-001" {
		t.Fatalf("CreatePluginInstance failed: res=%+v, err=%v", createRes, err)
	}

	// 2. Get instance -> properties merged
	got, err := logic.GetPluginInstance(kt, "pi-001")
	if err != nil {
		t.Fatalf("GetPluginInstance failed: %v", err)
	}
	if got.Properties["secret_id"] != "AKID-1" ||
		got.Properties["secret_key"] != "SK-1" ||
		got.Properties["region"] != "ap-guangzhou" {
		t.Fatalf("unexpected merged properties on Get: %#v", got.Properties)
	}

	// 3. Partial metadata update (Properties == nil) -> does not overwrite attrs/secret_attrs
	newMemo := "updated-memo"
	if err := logic.UpdatePluginInstance(kt, "pi-001", &hcplugin.PluginInstanceUpdateReq{
		Name: "prod-gz-renamed",
		Memo: &newMemo,
	}); err != nil {
		t.Fatalf("partial metadata UpdatePluginInstance failed: %v", err)
	}
	if len(env.lastUpdateReq.Items) == 0 ||
		env.lastUpdateReq.Items[0].Attrs != nil ||
		env.lastUpdateReq.Items[0].SecretAttrs != nil {
		t.Fatalf("metadata-only update should not send attrs/secret_attrs, got: %+v", env.lastUpdateReq)
	}

	// 4. Full overwrite of properties (region omitted) -> region is removed from attrs
	if err := logic.UpdatePluginInstance(kt, "pi-001", &hcplugin.PluginInstanceUpdateReq{
		Properties: map[string]any{
			"secret_id":  "AKID-2",
			"secret_key": "SK-2",
		},
	}); err != nil {
		t.Fatalf("full properties UpdatePluginInstance failed: %v", err)
	}

	gotAfterUpdate, err := logic.GetPluginInstance(kt, "pi-001")
	if err != nil {
		t.Fatalf("GetPluginInstance after update failed: %v", err)
	}
	if _, hasRegion := gotAfterUpdate.Properties["region"]; hasRegion {
		t.Fatalf("expected region to be removed after full overwrite, got: %#v", gotAfterUpdate.Properties)
	}
	if gotAfterUpdate.Properties["secret_key"] != "SK-2" || gotAfterUpdate.Name != "prod-gz-renamed" {
		t.Fatalf("unexpected instance state after full overwrite: %+v", gotAfterUpdate)
	}

	// 5. Delete instance
	if err := logic.DeletePluginInstance(kt, "pi-001"); err != nil || !env.deleted {
		t.Fatalf("DeletePluginInstance failed: err=%v, deleted=%v", err, env.deleted)
	}
}
