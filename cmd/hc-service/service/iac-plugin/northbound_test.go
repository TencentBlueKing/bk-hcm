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

package iacplugin

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	"hcm/cmd/hc-service/logics/iac-plugin/router"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/cmd/hc-service/service/capability"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"

	"github.com/emicklei/go-restful/v3"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

var (
	nbKeyOnce sync.Once
	nbPubPEM  string
)

func getNorthboundTestPubPEM(t *testing.T) string {
	t.Helper()
	nbKeyOnce.Do(func() {
		priv, err := rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			t.Fatalf("generate rsa key failed: %v", err)
		}
		der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key failed: %v", err)
		}
		nbPubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	})
	return nbPubPEM
}

func newNorthboundStubSouthboundServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/health"):
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{"plugin_version":"1.2.0"}}`))
		case strings.HasSuffix(r.URL.Path, "/stop"):
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{"status":"failed","message":"stopped"}}`))
		case strings.HasSuffix(r.URL.Path, "/apply"), strings.HasSuffix(r.URL.Path, "/result"):
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{"status":"succeeded"}}`))
		case strings.HasSuffix(r.URL.Path, "/import"), strings.HasSuffix(r.URL.Path, "/refresh"),
			strings.HasSuffix(r.URL.Path, "/read_data"):
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{"state":{"id":"1"}}}`))
		default:
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

type inMemDSClient struct{ handler http.HandlerFunc }

func (c *inMemDSClient) Do(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	c.handler(rec, req)
	return rec.Result(), nil
}

type stubDisc struct{}

func (stubDisc) GetServers() ([]string, error) { return []string{"http://127.0.0.1"}, nil }

func buildTestRouterService(t *testing.T) *router.Service {
	t.Helper()
	extBytes, _ := json.Marshal(coreplugin.HTTPExtension{
		APIPath:      "/api/bkcmdb",
		CryptoType:   "RSA",
		CryptoConfig: &coreplugin.HTTPCryptoConfig{PublicKey: getNorthboundTestPubPEM(t)},
	})
	detail := coreplugin.SchemaDetail{
		ProtocolVersion: "1.0",
		Capabilities: coreplugin.Capabilities{
			Apply: true, Plan: true, Refresh: true, ReadData: true,
			AsyncApply: true, Stop: true, Validate: true, Import: true, List: true,
		},
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema: &tfjson.Schema{Block: &tfjson.SchemaBlock{}},
			ResourceSchemas: map[string]*tfjson.Schema{
				"bkcmdb_host": {Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"a": {AttributeType: cty.Number, Optional: true},
					},
				}},
			},
			DataSourceSchemas: map[string]*tfjson.Schema{
				"bkcmdb_biz": {Block: &tfjson.SchemaBlock{}},
			},
		},
	}
	info := coreplugin.PluginInfo{Name: "bkcmdb", Version: "1.2.0"}
	hash, _ := coreplugin.ComputeSchemaHash(info, &detail)
	pluginRec := coreplugin.IaCPlugin{
		ID: "p1", PluginName: "bkcmdb", PluginVersion: "1.2.0",
		Protocol: enumor.IaCPluginProtocolHTTP, State: enumor.IaCPluginStateEnabled, Extension: extBytes,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/iac_plugins/list"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": errf.OK,
				"data": &dataplugin.IaCPluginListResult{Details: []coreplugin.IaCPlugin{pluginRec}},
			})
		case strings.HasSuffix(r.URL.Path, "/iac_plugin_schema/list"):
			detailBytes, _ := json.Marshal(detail)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": errf.OK,
				"data": &dataplugin.IacPluginSchemaListResult{Details: []tableplugin.IacPluginSchemaTable{{
					ID: "1", PluginName: "bkcmdb", PluginVersion: "1.2.0",
					SchemaHash: hash, SchemaDetail: types.JsonField(detailBytes),
				}}},
			})
		case strings.HasSuffix(r.URL.Path, "/iac_plugin_instance/list"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": errf.OK,
				"data": &dataplugin.IacPluginInstanceListResult{Details: []coreplugin.IacPluginInstance{{
					ID: "i1", PluginName: "bkcmdb",
					Attrs: map[string]any{"host": "https://cmdb.example.com"},
				}}},
			})
		}
	}
	dataCli := dataservice.NewClient(&restcli.Capability{
		Client:   &inMemDSClient{handler: handler},
		Discover: stubDisc{},
	}, "v1")
	schemaMgr := pluginschema.NewSchemaManager(dataCli, pluginschema.DefaultSchemaCacheTTL)
	sbSrv := newNorthboundStubSouthboundServer(t)
	gw := &cc.ApiGateway{Endpoints: []string{sbSrv.URL}, AppCode: "hcm", AppSecret: "sec"}
	httpPlugin := httpplugin.New(dataCli, nil, gw, schemaMgr, httpplugin.Options{Enabled: true})
	return router.New(dataCli, schemaMgr, httpPlugin, nil, nil)
}

func newTestContainer(t *testing.T) *restful.Container {
	t.Helper()
	cc.InitRuntime(cc.TestSetting{})
	ws := new(restful.WebService)
	ws.Path("/api/v1/hc")
	ws.Produces(restful.MIME_JSON)
	InitService(&capability.Capability{
		WebService:   ws,
		ClientSet:    &client.ClientSet{},
		PluginRouter: buildTestRouterService(t),
	})
	return restful.NewContainer().Add(ws)
}

func doTestRequest(t *testing.T, container *restful.Container, method, path, rawBody string) rest.BaseResp {
	t.Helper()
	var bodyReader *bytes.Reader
	if rawBody != "" {
		bodyReader = bytes.NewReader([]byte(rawBody))
	}
	if bodyReader == nil {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	kt := kit.New()
	kt.User = "tester"
	kt.AppCode = "test-app"
	req.Header = kt.Header()
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	var resp rest.BaseResp
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp
}

func TestNorthboundHandlers_Success(t *testing.T) {
	container := newTestContainer(t)
	testCases := []struct {
		name, method, path, body string
	}{
		{"GET health", http.MethodGet, "/api/v1/hc/iac_plugins/health?plugin=bkcmdb", ""},
		{"GET schema", http.MethodGet, "/api/v1/hc/iac_plugins/schema?plugin=bkcmdb&plugin_version=1.2.0", ""},
		{"POST execute", http.MethodPost, "/api/v1/hc/iac_plugins/execute",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","config":{"a":1}}`},
		{"POST result", http.MethodPost, "/api/v1/hc/iac_plugins/result",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","apply_id":"ap1"}`},
		{"POST stop", http.MethodPost, "/api/v1/hc/iac_plugins/stop",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","apply_id":"ap1"}`},
		{"POST plan", http.MethodPost, "/api/v1/hc/iac_plugins/plan",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","config":{"a":1}}`},
		{"POST validate", http.MethodPost, "/api/v1/hc/iac_plugins/validate",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","config_type":"resource","config":{"a":1}}`},
		{"POST import", http.MethodPost, "/api/v1/hc/iac_plugins/import",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","import_id":"imp1"}`},
		{"POST refresh", http.MethodPost, "/api/v1/hc/iac_plugins/refresh",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","state":{"id":1}}`},
		{"POST read_data", http.MethodPost, "/api/v1/hc/iac_plugins/read_data",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_biz","config":{"a":1}}`},
		{"POST list", http.MethodPost, "/api/v1/hc/iac_plugins/list",
			`{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","limit":10}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doTestRequest(t, container, tc.method, tc.path, tc.body)
			assert.Equal(t, errf.OK, resp.Code)
		})
	}
}

func TestNorthboundHandlers_DecodeAndValidateErrors(t *testing.T) {
	container := newTestContainer(t)
	testCases := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int32
	}{
		{
			name:     "health missing plugin",
			method:   http.MethodGet,
			path:     "/api/v1/hc/iac_plugins/health",
			wantCode: errf.InvalidParameter,
		},
		{
			name:     "schema invalid version",
			method:   http.MethodGet,
			path:     "/api/v1/hc/iac_plugins/schema?plugin=bkcmdb&plugin_version=bad",
			wantCode: errf.InvalidParameter,
		},
		{
			name:     "execute malformed json",
			method:   http.MethodPost,
			path:     "/api/v1/hc/iac_plugins/execute",
			body:     `{"plugin":`,
			wantCode: errf.DecodeRequestFailed,
		},
		{
			name:     "execute both prior_state and config nil",
			method:   http.MethodPost,
			path:     "/api/v1/hc/iac_plugins/execute",
			body:     `{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host"}`,
			wantCode: errf.InvalidParameter,
		},
		{
			name:     "validate missing config_type",
			method:   http.MethodPost,
			path:     "/api/v1/hc/iac_plugins/validate",
			body:     `{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","config":{"a":1}}`,
			wantCode: errf.InvalidParameter,
		},
		{
			name:     "validate invalid config_type",
			method:   http.MethodPost,
			path:     "/api/v1/hc/iac_plugins/validate",
			body:     `{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","config_type":"provider","config":{"a":1}}`,
			wantCode: errf.InvalidParameter,
		},
		{
			name:     "list limit out of bounds",
			method:   http.MethodPost,
			path:     "/api/v1/hc/iac_plugins/list",
			body:     `{"plugin":"bkcmdb","plugin_instance_id":"i1","type_name":"bkcmdb_host","limit":501}`,
			wantCode: errf.InvalidParameter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doTestRequest(t, container, tc.method, tc.path, tc.body)
			assert.Equal(t, tc.wantCode, resp.Code)
		})
	}
}
