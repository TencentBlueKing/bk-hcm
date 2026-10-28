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

package router

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/pool"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/cmd/hc-service/logics/tfruntime/codec"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/plugin/tfplugin5"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/plugin"

	goplugin "github.com/hashicorp/go-plugin"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/grpc"
)

type mockDataStore struct {
	mu        sync.RWMutex
	plugins   map[string][]coreplugin.IaCPlugin
	schemas   map[coreplugin.PluginInfo]*coreplugin.PluginSchema
	instances map[string]*coreplugin.IacPluginInstance
}

func newMockDataStore() *mockDataStore {
	return &mockDataStore{
		plugins:   make(map[string][]coreplugin.IaCPlugin),
		schemas:   make(map[coreplugin.PluginInfo]*coreplugin.PluginSchema),
		instances: make(map[string]*coreplugin.IacPluginInstance),
	}
}

type fakeDiscovery struct{}

func (fakeDiscovery) GetServers() ([]string, error) {
	return []string{"http://127.0.0.1"}, nil
}

type inMemHTTPClient struct {
	handler http.HandlerFunc
}

func (c *inMemHTTPClient) Do(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	c.handler(rec, req)
	return rec.Result(), nil
}

func newMockDataService(store *mockDataStore) *dataservice.Client {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/iac_plugins/list"):
			handlePluginList(w, r, store)
		case strings.HasSuffix(r.URL.Path, "/iac_plugin_schema/list"):
			handleSchemaList(w, r, store)
		case strings.HasSuffix(r.URL.Path, "/iac_plugin_instance/list"):
			handleInstanceList(w, r, store)
		}
	}
	cap := &restcli.Capability{
		Client:   &inMemHTTPClient{handler: handler},
		Discover: fakeDiscovery{},
	}
	return dataservice.NewClient(cap, "v1")
}

func handlePluginList(w http.ResponseWriter, r *http.Request, store *mockDataStore) {
	req := new(core.ListReq)
	_ = json.NewDecoder(r.Body).Decode(req)
	name := extractAtomField(req.Filter, "plugin_name")
	store.mu.RLock()
	list := store.plugins[name]
	store.mu.RUnlock()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": errf.OK,
		"data": &dataplugin.IaCPluginListResult{Details: list},
	})
}

func handleSchemaList(w http.ResponseWriter, r *http.Request, store *mockDataStore) {
	req := new(dataplugin.IacPluginSchemaListReq)
	_ = json.NewDecoder(r.Body).Decode(req)
	key := coreplugin.PluginInfo{
		Name:    extractAtomField(req.Filter, "plugin_name"),
		Version: extractAtomField(req.Filter, "plugin_version"),
	}
	store.mu.RLock()
	s := store.schemas[key]
	store.mu.RUnlock()

	var details []tableplugin.IacPluginSchemaTable
	if s != nil {
		detailBytes, _ := json.Marshal(s.SchemaDetail)
		details = []tableplugin.IacPluginSchemaTable{{
			ID:            "1",
			PluginName:    key.Name,
			PluginVersion: key.Version,
			SchemaHash:    s.SchemaHash,
			SchemaDetail:  types.JsonField(detailBytes),
		}}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": errf.OK,
		"data": &dataplugin.IacPluginSchemaListResult{Details: details},
	})
}

func handleInstanceList(w http.ResponseWriter, r *http.Request, store *mockDataStore) {
	req := new(dataplugin.IacPluginInstanceListReq)
	_ = json.NewDecoder(r.Body).Decode(req)
	id := extractAtomField(req.Filter, "id")
	store.mu.RLock()
	inst := store.instances[id]
	store.mu.RUnlock()

	var details []coreplugin.IacPluginInstance
	if inst != nil {
		details = []coreplugin.IacPluginInstance{*inst}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": errf.OK,
		"data": &dataplugin.IacPluginInstanceListResult{Details: details},
	})
}

func extractAtomField(expr *filter.Expression, field string) string {
	if expr == nil {
		return ""
	}
	for _, r := range expr.Rules {
		rule, ok := r.(*filter.AtomRule)
		if !ok {
			continue
		}
		if rule.Field == field {
			if val, ok := rule.Value.(string); ok {
				return val
			}
			if arr, ok := rule.Value.([]any); ok && len(arr) > 0 {
				if val, ok := arr[0].(string); ok {
					return val
				}
			}
		}
	}
	return ""
}

var (
	testKeyOnce sync.Once
	testPubPEM  string
)

func getTestPubPEM(t testing.TB) string {
	t.Helper()
	testKeyOnce.Do(func() {
		priv, err := rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			t.Fatalf("generate rsa key failed: %v", err)
		}
		der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key failed: %v", err)
		}
		testPubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	})
	return testPubPEM
}

func newTestPluginSchema(name, version string, cap coreplugin.Capabilities) *coreplugin.PluginSchema {
	detail := coreplugin.SchemaDetail{
		ProtocolVersion: "1.0",
		Capabilities:    cap,
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema: &tfjson.Schema{Block: &tfjson.SchemaBlock{}},
			ResourceSchemas: map[string]*tfjson.Schema{
				name + "_host": {
					Block: &tfjson.SchemaBlock{
						Attributes: map[string]*tfjson.SchemaAttribute{
							"a":          {AttributeType: cty.Number, Optional: true},
							"bk_biz_id":  {AttributeType: cty.Number, Optional: true},
							"bk_host_id": {AttributeType: cty.Number, Computed: true},
						},
					},
				},
			},
			DataSourceSchemas: map[string]*tfjson.Schema{
				name + "_biz": {Block: &tfjson.SchemaBlock{}},
			},
		},
	}
	info := coreplugin.PluginInfo{Name: name, Version: version}
	hash, _ := coreplugin.ComputeSchemaHash(info, &detail)
	return &coreplugin.PluginSchema{
		Plugin:       info,
		SchemaHash:   hash,
		SchemaDetail: detail,
	}
}

type stubSouthboundClient struct {
	mu           sync.RWMutex
	calls        int32
	healthCalls  int32
	schemaCalls  int32
	healthResult *plugin.HealthData
	healthErr    error
}

func newStubSouthboundServer(t testing.TB, sbCli *stubSouthboundClient) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/health"):
			atomic.AddInt32(&sbCli.healthCalls, 1)
			sbCli.mu.RLock()
			hErr := sbCli.healthErr
			hRes := sbCli.healthResult
			sbCli.mu.RUnlock()
			if hErr != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": true, "code": 0, "data": hRes})
		case strings.HasSuffix(r.URL.Path, "/schema"):
			atomic.AddInt32(&sbCli.schemaCalls, 1)
			w.WriteHeader(http.StatusInternalServerError)
		case strings.HasSuffix(r.URL.Path, "/apply"):
			atomic.AddInt32(&sbCli.calls, 1)
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{"status":"succeeded"}}`))
		case strings.HasSuffix(r.URL.Path, "/validate"), strings.HasSuffix(r.URL.Path, "/read_data"):
			atomic.AddInt32(&sbCli.calls, 1)
			_, _ = w.Write([]byte(`{"result":true,"code":0,"data":{}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func setupRouterFixture(t testing.TB, cap coreplugin.Capabilities) (*mockDataStore, *stubSouthboundClient, *Service) {
	t.Helper()
	store := newMockDataStore()
	extBytes, _ := json.Marshal(coreplugin.HTTPExtension{
		APIPath:      "/api/bkcmdb",
		CryptoType:   "RSA",
		CryptoConfig: &coreplugin.HTTPCryptoConfig{PublicKey: getTestPubPEM(t)},
	})
	store.plugins["bkcmdb"] = []coreplugin.IaCPlugin{
		{
			ID: "p1", PluginName: "bkcmdb", PluginVersion: "1.0.0",
			Protocol: enumor.IaCPluginProtocolHTTP, State: enumor.IaCPluginStateDisabled, Extension: extBytes,
		},
		{
			ID: "p2", PluginName: "bkcmdb", PluginVersion: "1.2.0",
			Protocol: enumor.IaCPluginProtocolHTTP, State: enumor.IaCPluginStateEnabled, Extension: extBytes,
		},
	}
	store.schemas[coreplugin.PluginInfo{Name: "bkcmdb", Version: "1.2.0"}] = newTestPluginSchema("bkcmdb", "1.2.0", cap)
	store.instances["00000002"] = &coreplugin.IacPluginInstance{
		ID:          "00000002",
		PluginName:  "bkcmdb",
		Attrs:       map[string]any{"bk_paas_host": "https://paas.example.com"},
		SecretAttrs: map[string]any{"token": "tok_123"},
	}
	store.instances["tc-inst"] = &coreplugin.IacPluginInstance{
		ID:          "tc-inst",
		PluginName:  "tencentcloud",
		Attrs:       map[string]any{"region": "ap-guangzhou"},
		SecretAttrs: map[string]any{"secret_key": "tc_secret"},
	}

	dataCli := newMockDataService(store)
	schemaMgr := pluginschema.NewSchemaManager(dataCli, pluginschema.DefaultSchemaCacheTTL)
	sbCli := &stubSouthboundClient{}
	sbSrv := newStubSouthboundServer(t, sbCli)
	gw := &cc.ApiGateway{Endpoints: []string{sbSrv.URL}, AppCode: "hcm", AppSecret: "sec"}
	httpPlugin := httpplugin.New(dataCli, nil, gw, schemaMgr, httpplugin.Options{Enabled: true})
	r := New(dataCli, schemaMgr, httpPlugin, nil, nil)
	return store, sbCli, r
}

func TestRouter_ResolveAndDispatch_PositiveAndOwnershipGate(t *testing.T) {
	fullCap := coreplugin.Capabilities{
		Apply: true, Plan: true, Refresh: true, ReadData: true,
		AsyncApply: true, Stop: true, Validate: true, Import: true, List: true,
	}
	store, sbCli, r := setupRouterFixture(t, fullCap)
	kt := kit.New()

	// 1. Positive resolve & dispatch: picks latest enabled 1.2.0 and merges properties
	commonHost := proto.CommonInstanceReq{
		Plugin: "bkcmdb", PluginInstanceID: "00000002", TypeName: "bkcmdb_host",
	}
	routeCtx, err := r.Resolve(kt, commonHost, enumor.IaCPluginOperationValidate)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.0", routeCtx.Plugin.PluginVersion)
	assert.Equal(t, map[string]any{"bk_paas_host": "https://paas.example.com", "token": "tok_123"}, routeCtx.Properties)

	_, err = r.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonHost,
		ConfigType:        proto.ConfigTypeResource,
		Config:            map[string]any{},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&sbCli.calls))

	// 2. AC-S02: Cross-plugin instance misuse rejected with InvalidParameter, 0 downstream calls
	_, err = r.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{
			Plugin: "bkcmdb", PluginInstanceID: "tc-inst", TypeName: "bkcmdb_host",
		},
		Config: map[string]any{"bk_biz_id": 213},
	})
	assert.Error(t, err)
	assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)
	assert.Equal(t, int32(1), atomic.LoadInt32(&sbCli.calls))

	// 3. Instance not found -> RecordNotFound
	_, err = r.Resolve(kt, proto.CommonInstanceReq{
		Plugin: "bkcmdb", PluginInstanceID: "not-exist", TypeName: "bkcmdb_host",
	}, enumor.IaCPluginOperationExecute)
	assert.Equal(t, errf.RecordNotFound, errf.Error(err).Code)

	// 4. Plugin not found -> RecordNotFound
	_, err = r.Resolve(kt, proto.CommonInstanceReq{
		Plugin: "unknown_plugin", PluginInstanceID: "00000002", TypeName: "bkcmdb_host",
	}, enumor.IaCPluginOperationExecute)
	assert.Equal(t, errf.RecordNotFound, errf.Error(err).Code)

	// 5. Latest plugin not enabled -> Aborted
	store.mu.Lock()
	store.plugins["bkcmdb"][1].State = enumor.IaCPluginStateDisabled
	store.mu.Unlock()
	_, err = r.Resolve(kt, commonHost, enumor.IaCPluginOperationExecute)
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
}

func TestRouter_SchemaTypeAndCapabilitiesGating(t *testing.T) {
	partialCap := coreplugin.Capabilities{
		Apply: true, Plan: true, Refresh: true, ReadData: true, Stop: true, AsyncApply: false,
	}
	_, sbCli, r := setupRouterFixture(t, partialCap)
	kt := kit.New()
	commonHost := proto.CommonInstanceReq{
		Plugin: "bkcmdb", PluginInstanceID: "00000002", TypeName: "bkcmdb_host",
	}

	// Positive resource and data_source routing
	_, err := r.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: commonHost,
		Config:            map[string]any{"a": 1},
	})
	assert.NoError(t, err)

	_, err = r.ReadData(kt, &proto.ReadDataReq{
		CommonInstanceReq: proto.CommonInstanceReq{
			Plugin: "bkcmdb", PluginInstanceID: "00000002", TypeName: "bkcmdb_biz",
		},
		Config: map[string]any{},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&sbCli.calls))

	// Negative type_name not in schema
	_, err = r.Resolve(kt, proto.CommonInstanceReq{
		Plugin: "bkcmdb", PluginInstanceID: "00000002", TypeName: "bkcmdb_unknown",
	}, enumor.IaCPluginOperationExecute)
	assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)

	// Negative ReadData on resource type
	_, err = r.Resolve(kt, commonHost, enumor.IaCPluginOperationReadData)
	assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)

	// Negative Import when capability.import == false
	_, err = r.Resolve(kt, commonHost, enumor.IaCPluginOperationImport)
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)

	// Negative Stop when stop == true but async_apply == false
	_, err = r.Resolve(kt, commonHost, enumor.IaCPluginOperationStop)
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
	assert.Equal(t, int32(2), atomic.LoadInt32(&sbCli.calls))
}

func TestRouter_GetSchemaZeroSouthboundAndHealth(t *testing.T) {
	fullCap := coreplugin.Capabilities{Apply: true, Plan: true, Refresh: true, ReadData: true}
	store, sbCli, r := setupRouterFixture(t, fullCap)
	store.schemas[coreplugin.PluginInfo{Name: "bkcmdb", Version: "1.0.0"}] = newTestPluginSchema("bkcmdb", "1.0.0", fullCap)
	sbCli.healthResult = &plugin.HealthData{PluginVersion: "1.2.0"}
	kt := kit.New()

	// AC-P02: Call GetSchema 100 times (both default enabled version and explicit historical version)
	for i := 0; i < 100; i++ {
		s, err := r.GetSchema(kt, &proto.GetSchemaReq{Plugin: "bkcmdb"})
		assert.NoError(t, err)
		assert.Equal(t, "1.2.0", s.Plugin.Version)
	}
	histSchema, err := r.GetSchema(kt, &proto.GetSchemaReq{Plugin: "bkcmdb", PluginVersion: "1.0.0"})
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", histSchema.Plugin.Version)
	assert.Equal(t, int32(0), atomic.LoadInt32(&sbCli.schemaCalls))

	// GetSchema non-existent version -> RecordNotFound
	_, err = r.GetSchema(kt, &proto.GetSchemaReq{Plugin: "bkcmdb", PluginVersion: "9.9.9"})
	assert.Equal(t, errf.RecordNotFound, errf.Error(err).Code)

	// Health positive -> 1.2.0
	hRes, err := r.Health(kt, &proto.HealthReq{Plugin: "bkcmdb"})
	assert.NoError(t, err)
	assert.Equal(t, "1.2.0", hRes.PluginVersion)

	// Health negative when southbound fails -> Aborted (2000006)
	sbCli.healthErr = errors.New("gateway 503")
	_, err = r.Health(kt, &proto.HealthReq{Plugin: "bkcmdb"})
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
}

// testProviderGRPCPlugin 实现 go-plugin GRPCPlugin 服务端，供子进程模式下启动真实 tfplugin5 ProviderServer。
type testProviderGRPCPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
}

func (p *testProviderGRPCPlugin) GRPCServer(_ *goplugin.GRPCBroker, s *grpc.Server) error {
	tfplugin5.RegisterProviderServer(s, &testTFProviderServer{})
	return nil
}

func (p *testProviderGRPCPlugin) GRPCClient(
	_ context.Context, _ *goplugin.GRPCBroker, _ *grpc.ClientConn) (any, error) {

	return nil, errors.New("not used on server side")
}

type testTFProviderServer struct {
	tfplugin5.UnimplementedProviderServer
}

func testProtoSchema() *tfplugin5.Schema {
	return &tfplugin5.Schema{
		Version: 1,
		Block: &tfplugin5.Schema_Block{
			Attributes: []*tfplugin5.Schema_Attribute{
				{Name: "a", Type: []byte(`"number"`), Optional: true},
				{Name: "bk_biz_id", Type: []byte(`"number"`), Optional: true},
				{Name: "bk_host_id", Type: []byte(`"number"`), Computed: true},
				{Name: "id", Type: []byte(`"string"`), Optional: true, Computed: true},
				{Name: "region", Type: []byte(`"string"`), Optional: true},
				{Name: "secret_key", Type: []byte(`"string"`), Optional: true, Sensitive: true},
			},
		},
	}
}

func (s *testTFProviderServer) GetSchema(_ context.Context,
	_ *tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error) {

	sch := testProtoSchema()
	return &tfplugin5.GetProviderSchema_Response{
		Provider:          sch,
		ResourceSchemas:   map[string]*tfplugin5.Schema{"tencentcloud_host": sch},
		DataSourceSchemas: map[string]*tfplugin5.Schema{"tencentcloud_biz": sch},
	}, nil
}

func (s *testTFProviderServer) PrepareProviderConfig(_ context.Context,
	req *tfplugin5.PrepareProviderConfig_Request) (*tfplugin5.PrepareProviderConfig_Response, error) {

	return &tfplugin5.PrepareProviderConfig_Response{PreparedConfig: req.GetConfig()}, nil
}

func (s *testTFProviderServer) Configure(_ context.Context,
	_ *tfplugin5.Configure_Request) (*tfplugin5.Configure_Response, error) {

	return &tfplugin5.Configure_Response{}, nil
}

func (s *testTFProviderServer) ValidateResourceTypeConfig(_ context.Context,
	_ *tfplugin5.ValidateResourceTypeConfig_Request) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {

	return &tfplugin5.ValidateResourceTypeConfig_Response{}, nil
}

func (s *testTFProviderServer) ValidateDataSourceConfig(_ context.Context,
	_ *tfplugin5.ValidateDataSourceConfig_Request) (*tfplugin5.ValidateDataSourceConfig_Response, error) {

	return &tfplugin5.ValidateDataSourceConfig_Response{}, nil
}

func (s *testTFProviderServer) ReadResource(_ context.Context,
	req *tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error) {

	return &tfplugin5.ReadResource_Response{
		NewState: req.GetCurrentState(),
		Private:  req.GetPrivate(),
	}, nil
}

func (s *testTFProviderServer) PlanResourceChange(_ context.Context,
	req *tfplugin5.PlanResourceChange_Request) (*tfplugin5.PlanResourceChange_Response, error) {

	return &tfplugin5.PlanResourceChange_Response{
		PlannedState:   req.GetProposedNewState(),
		PlannedPrivate: req.GetPriorPrivate(),
	}, nil
}

func (s *testTFProviderServer) ApplyResourceChange(_ context.Context,
	req *tfplugin5.ApplyResourceChange_Request) (*tfplugin5.ApplyResourceChange_Response, error) {

	return &tfplugin5.ApplyResourceChange_Response{
		NewState: req.GetPlannedState(),
		Private:  req.GetPlannedPrivate(),
	}, nil
}

func (s *testTFProviderServer) ImportResourceState(_ context.Context,
	req *tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error) {

	dv, err := (codec.Codec{}).Encode(testProtoSchema(), map[string]any{"id": req.GetId()})
	if err != nil {
		return nil, err
	}
	return &tfplugin5.ImportResourceState_Response{
		ImportedResources: []*tfplugin5.ImportResourceState_ImportedResource{{
			TypeName: req.GetTypeName(),
			State:    dv,
		}},
	}, nil
}

func (s *testTFProviderServer) ReadDataSource(_ context.Context,
	_ *tfplugin5.ReadDataSource_Request) (*tfplugin5.ReadDataSource_Response, error) {

	dv, err := (codec.Codec{}).Encode(testProtoSchema(), map[string]any{"id": "ds-1"})
	if err != nil {
		return nil, err
	}
	return &tfplugin5.ReadDataSource_Response{State: dv}, nil
}

func (s *testTFProviderServer) Stop(_ context.Context,
	_ *tfplugin5.Stop_Request) (*tfplugin5.Stop_Response, error) {

	return &tfplugin5.Stop_Response{}, nil
}

func TestMain(m *testing.M) {
	if os.Getenv("PLUGIN_PROTOCOL_VERSIONS") != "" ||
		os.Getenv(constant.TFPluginMagicCookieKey) == constant.TFPluginMagicCookieValue {
		goplugin.Serve(&goplugin.ServeConfig{
			HandshakeConfig: goplugin.HandshakeConfig{
				ProtocolVersion:  constant.TFPluginCoreProtocolVersion,
				MagicCookieKey:   constant.TFPluginMagicCookieKey,
				MagicCookieValue: constant.TFPluginMagicCookieValue,
			},
			VersionedPlugins: map[int]goplugin.PluginSet{
				constant.TFPluginProtocolVersion: {
					constant.TFPluginProviderName: &testProviderGRPCPlugin{},
				},
			},
			GRPCServer: goplugin.DefaultGRPCServer,
		})
		return
	}
	tmpLogDir, _ := os.MkdirTemp("", "router-test-logs-")
	logs.InitLogger(logs.LogConfig{
		LogDir:         tmpLogDir,
		LogMaxSize:     500,
		LogLineMaxSize: 2,
		LogMaxNum:      5,
		ToStdErr:       false,
		AlsoToStdErr:   false,
		Verbosity:      3,
	})
	code := m.Run()
	_ = os.RemoveAll(tmpLogDir)
	os.Exit(code)
}

func newTFTestSchema(name, version string, cap coreplugin.Capabilities) *coreplugin.PluginSchema {
	block := &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"a":          {AttributeType: cty.Number, Optional: true},
			"bk_biz_id":  {AttributeType: cty.Number, Optional: true},
			"bk_host_id": {AttributeType: cty.Number, Computed: true},
			"id":         {AttributeType: cty.String, Optional: true, Computed: true},
			"region":     {AttributeType: cty.String, Optional: true},
			"secret_key": {AttributeType: cty.String, Optional: true, Sensitive: true},
		},
	}
	detail := coreplugin.SchemaDetail{
		ProtocolVersion: "5.3",
		Capabilities:    cap,
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema:      &tfjson.Schema{Block: block},
			ResourceSchemas:   map[string]*tfjson.Schema{name + "_host": {Block: block}},
			DataSourceSchemas: map[string]*tfjson.Schema{name + "_biz": {Block: block}},
		},
	}
	info := coreplugin.PluginInfo{Name: name, Version: version}
	hash, _ := coreplugin.ComputeSchemaHash(info, &detail)
	return &coreplugin.PluginSchema{Plugin: info, SchemaHash: hash, SchemaDetail: detail}
}

func setupTFProviderRouterFixture(t testing.TB) (*mockDataStore, *Service, *pool.PluginRegistry) {
	t.Helper()
	t.Setenv("TMPDIR", "/tmp")
	setting := &cc.HCServiceSetting{}
	setting.TFPlugin.MinPort = 10000
	setting.TFPlugin.MaxPort = 25000
	cc.InitRuntime(setting)
	store := newMockDataStore()
	tfCap := coreplugin.Capabilities{
		Apply: true, Plan: true, Refresh: true, ReadData: true,
		Validate: true, Import: true, AsyncApply: false, Stop: false, List: false,
	}
	store.plugins["tencentcloud"] = []coreplugin.IaCPlugin{
		{
			ID: "tf-1", PluginName: "tencentcloud", PluginVersion: "1.81.5",
			Protocol: enumor.IaCPluginProtocolTFProvider, State: enumor.IaCPluginStateEnabled,
		},
		{
			ID: "tf-2", PluginName: "tencentcloud", PluginVersion: "1.82.0",
			Protocol: enumor.IaCPluginProtocolTFProvider, State: enumor.IaCPluginStateEnabled,
		},
	}
	store.schemas[coreplugin.PluginInfo{Name: "tencentcloud", Version: "1.81.5"}] =
		newTFTestSchema("tencentcloud", "1.81.5", tfCap)
	store.schemas[coreplugin.PluginInfo{Name: "tencentcloud", Version: "1.82.0"}] =
		newTFTestSchema("tencentcloud", "1.82.0", tfCap)
	store.instances["tc-inst"] = &coreplugin.IacPluginInstance{
		ID:          "tc-inst",
		PluginName:  "tencentcloud",
		Attrs:       map[string]any{"region": "ap-guangzhou"},
		SecretAttrs: map[string]any{"secret_key": "SK_TF_SUPER_SECRET_999"},
	}

	dataCli := newMockDataService(store)
	schemaMgr := pluginschema.NewSchemaManager(dataCli, pluginschema.DefaultSchemaCacheTTL)
	selfExec, err := os.Executable()
	require.NoError(t, err)
	entry := discovery.Entry{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		ExecPath:      selfExec,
	}
	plugins := map[string]discovery.Entry{"tencentcloud": entry}
	reg, err := pool.NewPluginRegistry(dataCli, []discovery.Entry{entry}, pool.PluginRegistryOptions{
		Pool: pool.PoolConfig{
			MaxProcesses:        2,
			MaxLeasesPerProcess: 8,
			MinIdleProcesses:    1,
			MaxQueueLength:      64,
		},
		CheckInterval: time.Hour,
	})
	require.NoError(t, err)
	reg.StartSync()
	t.Cleanup(func() { _ = reg.Close(context.Background()) })
	require.Eventually(t, func() bool {
		_, err := reg.GetPool("tencentcloud")
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)

	r := New(dataCli, schemaMgr, nil, reg, plugins)
	return store, r, reg
}

func TestRouter_TFProviderLocalVersionLockHealthAndGates(t *testing.T) {
	store, r, _ := setupTFProviderRouterFixture(t)
	kt := kit.New()
	commonRes := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "tc-inst", TypeName: "tencentcloud_host",
	}
	commonDS := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "tc-inst", TypeName: "tencentcloud_biz",
	}

	// 1. 滚动升级多版本（1.81.5 与 1.82.0）并存时，严格锁定本副本 localVersion 1.81.5
	routeCtx, err := r.Resolve(kt, commonRes, enumor.IaCPluginOperationExecute)
	require.NoError(t, err)
	assert.Equal(t, "1.81.5", routeCtx.Plugin.PluginVersion)
	assert.Equal(t, "1.81.5", routeCtx.Schema.Plugin.Version)

	schemaRes, err := r.GetSchema(kt, &proto.GetSchemaReq{Plugin: "tencentcloud"})
	require.NoError(t, err)
	assert.Equal(t, "1.81.5", schemaRes.Plugin.Version)

	// 2. 本地插件 Health 探活：enabled 时返回本副本 localVersion 1.81.5
	hRes, err := r.Health(kt, &proto.HealthReq{Plugin: "tencentcloud"})
	require.NoError(t, err)
	assert.Equal(t, "1.81.5", hRes.PluginVersion)

	// 3. Validate 按 config_type 精确门控：资源传 data 或数据源传 resource 均拦截
	_, err = r.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonRes,
		ConfigType:        proto.ConfigTypeDataSource,
		Config:            map[string]any{},
	})
	assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)

	_, err = r.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonDS,
		ConfigType:        proto.ConfigTypeResource,
		Config:            map[string]any{},
	})
	assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)

	// 4. 调用不支持的 GetResult / Stop / List 被能力位前置拦截
	_, err = r.GetResult(kt, &proto.GetResultReq{CommonInstanceReq: commonRes, ApplyID: "ap-1"})
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
	_, err = r.Stop(kt, &proto.StopReq{CommonInstanceReq: commonRes, ApplyID: "ap-1"})
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
	_, err = r.List(kt, &proto.ListReq{CommonInstanceReq: commonRes, Limit: 10})
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)

	// 5. 6 个运行态接口经真实 PluginPool 与 TFPluginAdaptor 分发成功
	execRes, err := r.Execute(kt, &proto.ExecuteReq{CommonInstanceReq: commonRes, Config: map[string]any{"a": 1}})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, execRes.Status)

	_, err = r.Plan(kt, &proto.PlanReq{CommonInstanceReq: commonRes, Config: map[string]any{"a": 1}})
	require.NoError(t, err)
	_, err = r.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonRes, ConfigType: proto.ConfigTypeResource, Config: map[string]any{"a": 1},
	})
	require.NoError(t, err)
	_, err = r.Import(kt, &proto.ImportReq{CommonInstanceReq: commonRes, ImportID: "imp-1"})
	require.NoError(t, err)
	_, err = r.Refresh(kt, &proto.RefreshReq{CommonInstanceReq: commonRes, State: map[string]any{"id": "1"}})
	require.NoError(t, err)
	_, err = r.ReadData(kt, &proto.ReadDataReq{CommonInstanceReq: commonDS, Config: map[string]any{"a": 1}})
	require.NoError(t, err)

	// 6. 当本副本 localVersion 处于 disabled 状态时，Health 与运行态请求均返回 Aborted
	store.mu.Lock()
	store.plugins["tencentcloud"][0].State = enumor.IaCPluginStateDisabled
	store.mu.Unlock()
	_, err = r.Health(kt, &proto.HealthReq{Plugin: "tencentcloud"})
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
	_, err = r.Execute(kt, &proto.ExecuteReq{CommonInstanceReq: commonRes, Config: map[string]any{"a": 1}})
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
}

func TestRouter_TFProviderSecretRedactionAnd50GoroutinesConcurrent(t *testing.T) {
	const secretMarker = "SK_TF_SUPER_SECRET_999"
	logDir := t.TempDir()
	logs.InitLogger(logs.LogConfig{
		LogDir:         logDir,
		LogMaxSize:     500,
		LogLineMaxSize: 2,
		LogMaxNum:      5,
		ToStdErr:       false,
		AlsoToStdErr:   false,
		Verbosity:      3,
	})

	_, r, reg := setupTFProviderRouterFixture(t)
	commonRes := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "tc-inst", TypeName: "tencentcloud_host",
	}
	commonDS := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "tc-inst", TypeName: "tencentcloud_biz",
	}

	// 1. 50 协程并发调用 Health、GetSchema 与全部 6 个运行态接口（池容量仅 2*8=16，验证 Lease 100% 归还才能全部完成）
	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			kt := kit.New()
			_, err := r.Health(kt, &proto.HealthReq{Plugin: "tencentcloud"})
			assert.NoError(t, err)
			_, err = r.GetSchema(kt, &proto.GetSchemaReq{Plugin: "tencentcloud"})
			assert.NoError(t, err)
			_, err = r.Execute(kt, &proto.ExecuteReq{CommonInstanceReq: commonRes, Config: map[string]any{"a": 1}})
			assert.NoError(t, err)
			_, err = r.Plan(kt, &proto.PlanReq{CommonInstanceReq: commonRes, Config: map[string]any{"a": 1}})
			assert.NoError(t, err)
			_, err = r.Validate(kt, &proto.ValidateReq{
				CommonInstanceReq: commonRes, ConfigType: proto.ConfigTypeResource, Config: map[string]any{"a": 1},
			})
			assert.NoError(t, err)
			_, err = r.Import(kt, &proto.ImportReq{CommonInstanceReq: commonRes, ImportID: "imp-1"})
			assert.NoError(t, err)
			_, err = r.Refresh(kt, &proto.RefreshReq{CommonInstanceReq: commonRes, State: map[string]any{"id": "1"}})
			assert.NoError(t, err)
			_, err = r.ReadData(kt, &proto.ReadDataReq{CommonInstanceReq: commonDS, Config: map[string]any{"a": 1}})
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// 2. 关闭进程池后触发 Acquire 失败分支，验证错误与日志中不包含敏感凭证
	require.NoError(t, reg.Close(context.Background()))
	_, failErr := r.Execute(kit.New(), &proto.ExecuteReq{
		CommonInstanceReq: commonRes,
		Config:            map[string]any{"a": 1},
	})
	require.Error(t, failErr)
	assert.Equal(t, errf.Aborted, errf.Error(failErr).Code)
	assert.NotContains(t, failErr.Error(), secretMarker)

	logs.CloseLogs()
	entries, _ := os.ReadDir(logDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(logDir, entry.Name()))
		require.NoError(t, readErr)
		assert.NotContains(t, string(raw), secretMarker)
	}
}
