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
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func generateTestRSA3072Key(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 3072)
	assert.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	assert.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
	return priv, string(pubPEM)
}

func TestNewLaunchConfigEncryptor(t *testing.T) {
	priv, pubPEM := generateTestRSA3072Key(t)
	validCfg, err := json.Marshal(RSACryptoConfig{PublicKey: pubPEM})
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		cryptoType   CryptoType
		cryptoConfig json.RawMessage
		wantErr      bool
	}{
		{name: "valid RSA crypto_config", cryptoType: CryptoTypeRSA, cryptoConfig: validCfg, wantErr: false},
		{name: "unsupported crypto_type", cryptoType: "AES", cryptoConfig: validCfg, wantErr: true},
		{name: "empty crypto_config", cryptoType: CryptoTypeRSA, cryptoConfig: nil, wantErr: true},
		{name: "invalid JSON in crypto_config", cryptoType: CryptoTypeRSA, cryptoConfig: json.RawMessage(`{invalid`),
			wantErr: true},
		{name: "missing public_key in crypto_config", cryptoType: CryptoTypeRSA, cryptoConfig: json.RawMessage(`{}`),
			wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCryptoConfig(tc.cryptoType, tc.cryptoConfig)
			encryptor, encErr := NewLaunchConfigEncryptor(tc.cryptoType, tc.cryptoConfig)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Error(t, encErr)
				assert.Nil(t, encryptor)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, encErr)
			b64, err := encryptor.Encrypt(map[string]any{"access_token": "tok-1", "bk_tenant_id": "default"})
			assert.NoError(t, err)
			assert.Equal(t, 512, len(b64))

			cipherBytes, err := base64.StdEncoding.DecodeString(b64)
			assert.NoError(t, err)
			plainBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, cipherBytes, nil)
			assert.NoError(t, err)
			assert.JSONEq(t, `{"access_token":"tok-1","bk_tenant_id":"default"}`, string(plainBytes))
		})
	}
}

func newValidSampleSchema() *PluginSchema {
	return &PluginSchema{
		Plugin:     PluginInfo{Name: "bkcmdb", Version: "1.2.0"},
		SchemaHash: "sha256:3f9a",
		SchemaDetail: SchemaDetail{
			ProtocolVersion: "1.0",
			Capabilities: Capabilities{
				Apply:      true,
				Plan:       true,
				Refresh:    true,
				ReadData:   true,
				AsyncApply: true,
				Stop:       true,
				Validate:   true,
				Import:     true,
				List:       true,
			},
			ProviderSchema: tfjson.ProviderSchema{
				ConfigSchema: &tfjson.Schema{
					Version: 0,
					Block: &tfjson.SchemaBlock{
						Attributes: map[string]*tfjson.SchemaAttribute{
							"access_token": {AttributeType: cty.String, Required: true, Sensitive: true},
						},
					},
				},
				ResourceSchemas: map[string]*tfjson.Schema{
					"bkcmdb_host": {
						Version: 0,
						Block: &tfjson.SchemaBlock{
							Attributes: map[string]*tfjson.SchemaAttribute{
								"bk_host_innerip": {AttributeType: cty.String, Required: true},
								"bk_host_id":      {AttributeType: cty.Number, Computed: true},
							},
						},
					},
				},
				DataSourceSchemas: map[string]*tfjson.Schema{
					"bkcmdb_biz": {
						Version: 0,
						Block: &tfjson.SchemaBlock{
							Attributes: map[string]*tfjson.SchemaAttribute{
								"bk_biz_name": {AttributeType: cty.String, Required: true},
								"bk_biz_id":   {AttributeType: cty.Number, Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

func TestPluginSchema_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		mutate  func(s *PluginSchema)
		wantErr bool
	}{
		{name: "valid schema", mutate: func(s *PluginSchema) {}, wantErr: false},
		{name: "missing protocol_version", mutate: func(s *PluginSchema) { s.ProtocolVersion = "" }, wantErr: true},
		{name: "missing plugin.name", mutate: func(s *PluginSchema) { s.Plugin.Name = "" }, wantErr: true},
		{name: "resource type_name wrong prefix", mutate: func(s *PluginSchema) {
			s.ResourceSchemas["other_host"] = s.ResourceSchemas["bkcmdb_host"]
		}, wantErr: true},
		{name: "resource_schemas non-empty without apply capability", mutate: func(s *PluginSchema) {
			s.Capabilities.Apply = false
		}, wantErr: true},
		{name: "resource_schemas empty with resource capabilities true", mutate: func(s *PluginSchema) {
			s.ResourceSchemas = nil
		}, wantErr: true},
		{name: "data_source_schemas empty with read_data true", mutate: func(s *PluginSchema) {
			s.DataSourceSchemas = nil
			s.Capabilities.ReadData = true
		}, wantErr: true},
		{name: "stop capability true without async_apply", mutate: func(s *PluginSchema) {
			s.Capabilities.AsyncApply = false
			s.Capabilities.Stop = true
		}, wantErr: true},
		{name: "missing provider block", mutate: func(s *PluginSchema) {
			s.ConfigSchema = nil
		}, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema := newValidSampleSchema()
			tc.mutate(schema)
			err := schema.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPathStepAndDiagnostic(t *testing.T) {
	assert.Equal(t, "fields", cvt.PtrToVal(NewAttributeStep("fields").Attribute))
	assert.Equal(t, "env", cvt.PtrToVal(NewKeyStep("env").Key))
	assert.Equal(t, 0, cvt.PtrToVal(NewIndexStep(0).Index))

	passWithWarning := &ValidateData{
		Diagnostics: []Diagnostic{
			{
				Diagnostic: tfjson.Diagnostic{Severity: tfjson.DiagnosticSeverityWarning},
				Path:       []PathStep{NewAttributeStep("operator")},
				Message:    "deprecated",
			},
		},
	}
	assert.True(t, passWithWarning.IsPassed())

	failWithDefaultError := &ValidateData{
		Diagnostics: []Diagnostic{
			{Path: []PathStep{NewAttributeStep("bk_biz_id")}, Message: "must be positive"},
		},
	}
	assert.False(t, failWithDefaultError.IsPassed())
}

func TestRequestValidateErrors(t *testing.T) {
	assert.Error(t, (*ApplyReq)(nil).Validate())
	assert.Error(t, (&ApplyReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&PlanReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&GetResultReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&StopReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&ValidateReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&ImportReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&RefreshReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&ReadDataReq{TypeName: "a", LaunchConfig: "b"}).Validate())
	assert.Error(t, (&ListReq{TypeName: "a", LaunchConfig: "b", Limit: 0}).Validate())
	assert.Error(t, (&ListReq{TypeName: "a", LaunchConfig: "b", Limit: 501}).Validate())

	_, err := NewClient("", &cc.ApiGateway{}, nil)
	assert.Error(t, err)
	_, err = NewClient("ftp://invalid", &cc.ApiGateway{}, nil)
	assert.Error(t, err)
	_, err = NewClient("http://localhost", nil, nil)
	assert.Error(t, err)
}

func newTestServerAndClient(t *testing.T) (*httptest.Server, Client) {
	t.Helper()
	mux := http.NewServeMux()
	registerTestHandlers(t, mux)

	srv := httptest.NewServer(mux)
	cfg := &cc.ApiGateway{
		Endpoints: []string{srv.URL},
		AppCode:   "bk_hcm",
		AppSecret: "secret-123",
	}
	cli, err := NewClient(srv.URL+"/api/bk-hcm-plugin-cmdb/prod", cfg, nil)
	assert.NoError(t, err)
	return srv, cli
}

func registerTestHandlers(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	prefix := "/api/bk-hcm-plugin-cmdb/prod"

	mux.HandleFunc(prefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"plugin_version":"1.2.0"}}`))
	})
	mux.HandleFunc(prefix+"/schema", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		resp := map[string]any{"result": true, "code": 0, "message": "", "data": newValidSampleSchema()}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc(prefix+"/apply", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		var req ApplyReq
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.TypeName == "itsm_ticket" {
			_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"status":"running","apply_id":"REQ001","state":{"sn":"REQ001"}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"status":"failed","message":"transfer module 5002 failed","state":{"bk_host_id":10086,"bk_module_id":5001}}}`))
	})
	mux.HandleFunc(prefix+"/result", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"status":"succeeded","apply_id":"REQ001","state":{"sn":"REQ001","approve_result":"approved"}}}`))
	})
	mux.HandleFunc(prefix+"/stop", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"status":"failed","message":"ticket revoked","apply_id":"REQ001","state":{"sn":"REQ001"}}}`))
	})
	registerTestReadAndPlanHandlers(prefix, mux)
}

func registerTestReadAndPlanHandlers(prefix string, mux *http.ServeMux) {
	mux.HandleFunc(prefix+"/plan", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"state":{"bk_host_innerip":"10.0.0.9"},"replace_paths":["bk_host_innerip"]}}`))
	})
	mux.HandleFunc(prefix+"/validate", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"diagnostics":[{"path":[{"attribute":"fields"},{"index":0},{"attribute":"key"}],"message":"cannot be empty"}]}}`))
	})
	mux.HandleFunc(prefix+"/import", func(w http.ResponseWriter, r *http.Request) {
		var req ImportReq
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.ImportID == "not-found" {
			_, _ = w.Write([]byte(`{"result":false,"code":40400,"message":"host not found","data":null}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"state":{"bk_host_id":10086}}}`))
	})
	mux.HandleFunc(prefix+"/refresh", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"state":null}}`))
	})
	mux.HandleFunc(prefix+"/read_data", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"state":{"bk_biz_name":"hcm","bk_biz_id":213}}}`))
	})
	mux.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":true,"code":0,"message":"","data":{"items":[{"import_id":"10086","state":{"bk_host_id":10086}}],"next_page_token":"1"}}`))
	})
}

func TestClient_HealthSchemaAndApplyLifecycle(t *testing.T) {
	srv, cli := newTestServerAndClient(t)
	defer srv.Close()
	kt := kit.New()

	health, err := cli.Health(kt)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.0", health.PluginVersion)

	schema, err := cli.GetSchema(kt)
	assert.NoError(t, err)
	assert.NoError(t, schema.Validate())

	// Async apply -> running
	asyncRes, err := cli.Apply(kt, &ApplyReq{
		TypeName:      "itsm_ticket",
		LaunchConfig:  "cipher",
		ProposedState: map[string]any{"service_id": 368},
	})
	assert.NoError(t, err)
	assert.Equal(t, TaskStatusRunning, asyncRes.Status)
	assert.Equal(t, "REQ001", asyncRes.ApplyID)

	// Task failed with partial state and data.message (result=true, code=0)
	failedRes, err := cli.Apply(kt, &ApplyReq{
		TypeName:      "bkcmdb_host",
		LaunchConfig:  "cipher",
		ProposedState: map[string]any{"bk_host_innerip": "10.0.0.1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, TaskStatusFailed, failedRes.Status)
	assert.Equal(t, "transfer module 5002 failed", failedRes.Message)
	assert.Equal(t, float64(10086), failedRes.State["bk_host_id"])

	// GetResult & Stop
	pollRes, err := cli.GetResult(kt, &GetResultReq{TypeName: "itsm_ticket", LaunchConfig: "cipher", ApplyID: "REQ001"})
	assert.NoError(t, err)
	assert.Equal(t, TaskStatusSucceeded, pollRes.Status)

	stopRes, err := cli.Stop(kt, &StopReq{TypeName: "itsm_ticket", LaunchConfig: "cipher", ApplyID: "REQ001"})
	assert.NoError(t, err)
	assert.Equal(t, TaskStatusFailed, stopRes.Status)
	assert.Equal(t, "ticket revoked", stopRes.Message)
}

func TestClient_PlanValidateReadAndList(t *testing.T) {
	srv, cli := newTestServerAndClient(t)
	defer srv.Close()
	kt := kit.New()

	planData, err := cli.Plan(kt, &PlanReq{
		TypeName:      "bkcmdb_host",
		LaunchConfig:  "cipher",
		PriorState:    map[string]any{"bk_host_innerip": "10.0.0.1"},
		ProposedState: map[string]any{"bk_host_innerip": "10.0.0.9"},
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"bk_host_innerip"}, planData.ReplacePaths)

	valData, err := cli.Validate(kt, &ValidateReq{
		TypeName:     "itsm_ticket",
		LaunchConfig: "cipher",
		Config:       map[string]any{"service_id": 368},
	})
	assert.NoError(t, err)
	assert.False(t, valData.IsPassed())
	assert.Len(t, valData.Diagnostics, 1)

	impData, err := cli.Import(kt, &ImportReq{TypeName: "bkcmdb_host", LaunchConfig: "cipher", ImportID: "10086"})
	assert.NoError(t, err)
	assert.Equal(t, float64(10086), impData.State["bk_host_id"])

	_, err = cli.Import(kt, &ImportReq{TypeName: "bkcmdb_host", LaunchConfig: "cipher", ImportID: "not-found"})
	assert.Error(t, err)

	refData, err := cli.Refresh(kt, &RefreshReq{
		TypeName:     "bkcmdb_host",
		LaunchConfig: "cipher",
		State:        map[string]any{"bk_host_id": 10086},
	})
	assert.NoError(t, err)
	assert.Nil(t, refData.State)

	rdData, err := cli.ReadData(kt, &ReadDataReq{
		TypeName:     "bkcmdb_biz",
		LaunchConfig: "cipher",
		Config:       map[string]any{"bk_biz_name": "hcm"},
	})
	assert.NoError(t, err)
	assert.Equal(t, float64(213), rdData.State["bk_biz_id"])

	listData, err := cli.List(kt, &ListReq{TypeName: "bkcmdb_host", LaunchConfig: "cipher", Limit: 2})
	assert.NoError(t, err)
	assert.Len(t, listData.Items, 1)
	assert.Equal(t, "1", listData.NextPageToken)
}
