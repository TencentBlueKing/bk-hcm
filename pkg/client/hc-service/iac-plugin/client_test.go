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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	coreplugin "hcm/pkg/api/core/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"

	"github.com/stretchr/testify/assert"
)

type testDiscovery struct {
	server string
}

func (d testDiscovery) GetServers() ([]string, error) {
	return []string{d.server}, nil
}

func newTestNorthboundClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	httpCli, err := restcli.NewClient(nil)
	assert.NoError(t, err)
	cap := &restcli.Capability{
		Client:   httpCli,
		Discover: testDiscovery{server: srv.URL},
	}
	return NewClient(rest.NewClient(cap, "/api/v1/hc")), srv
}

func writeJSONOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    errf.OK,
		"message": "",
		"data":    data,
	})
}

func TestNorthboundClient_GetEndpoints(t *testing.T) {
	cli, srv := newTestNorthboundClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		switch r.URL.Path {
		case "/api/v1/hc/iac_plugins/health":
			assert.Equal(t, "bkcmdb", r.URL.Query().Get("plugin"))
			writeJSONOK(w, &proto.HealthResult{PluginVersion: "1.2.0"})
		case "/api/v1/hc/iac_plugins/schema":
			assert.Equal(t, "bkcmdb", r.URL.Query().Get("plugin"))
			assert.Equal(t, "1.2.0", r.URL.Query().Get("plugin_version"))
			writeJSONOK(w, &coreplugin.PluginSchema{
				Plugin:     coreplugin.PluginInfo{Name: "bkcmdb", Version: "1.2.0"},
				SchemaHash: "sha256:abc",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()

	kt := kit.New()
	health, err := cli.Health(kt, "bkcmdb")
	assert.NoError(t, err)
	assert.Equal(t, "1.2.0", health.PluginVersion)

	schema, err := cli.GetSchema(kt, "bkcmdb", "1.2.0")
	assert.NoError(t, err)
	assert.Equal(t, "bkcmdb", schema.Plugin.Name)
	assert.Equal(t, "1.2.0", schema.Plugin.Version)
}

func TestNorthboundClient_PostLifecycleEndpoints(t *testing.T) {
	cli, srv := newTestNorthboundClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		switch r.URL.Path {
		case "/api/v1/hc/iac_plugins/execute",
			"/api/v1/hc/iac_plugins/result",
			"/api/v1/hc/iac_plugins/stop":
			writeJSONOK(w, &proto.ExecuteResult{
				Status: enumor.IaCPluginTaskStatusSucceeded,
				State:  map[string]any{"bk_host_id": float64(10086)},
			})
		case "/api/v1/hc/iac_plugins/plan":
			writeJSONOK(w, &proto.PlanResult{
				State:        map[string]any{"bk_host_id": float64(10086)},
				ReplacePaths: []string{"bk_host_innerip"},
			})
		case "/api/v1/hc/iac_plugins/validate":
			writeJSONOK(w, &proto.ValidateResult{
				Diagnostics: []coreplugin.Diagnostic{{Message: "ok"}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()

	kt := kit.New()
	commonReq := proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "inst-1", TypeName: "bkcmdb_host"}

	execRes, err := cli.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: commonReq,
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, execRes.Status)

	resResult, err := cli.GetResult(kt, &proto.GetResultReq{CommonInstanceReq: commonReq, ApplyID: "ap-1"})
	assert.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, resResult.Status)

	stopRes, err := cli.Stop(kt, &proto.StopReq{CommonInstanceReq: commonReq, ApplyID: "ap-1"})
	assert.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, stopRes.Status)

	planRes, err := cli.Plan(kt, &proto.PlanReq{
		CommonInstanceReq: commonReq,
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"bk_host_innerip"}, planRes.ReplacePaths)

	valRes, err := cli.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonReq,
		ConfigType:        proto.ConfigTypeResource,
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.NoError(t, err)
	assert.Len(t, valRes.Diagnostics, 1)
}

func TestNorthboundClient_PostStateAndQueryEndpoints(t *testing.T) {
	cli, srv := newTestNorthboundClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		switch r.URL.Path {
		case "/api/v1/hc/iac_plugins/import", "/api/v1/hc/iac_plugins/refresh":
			writeJSONOK(w, &proto.StateResult{
				State: map[string]any{"bk_host_id": float64(10086)},
			})
		case "/api/v1/hc/iac_plugins/read_data":
			writeJSONOK(w, &proto.ReadDataResult{
				State: map[string]any{"bk_biz_id": float64(213)},
			})
		case "/api/v1/hc/iac_plugins/list":
			writeJSONOK(w, &proto.ListResult{
				Items:         []coreplugin.ListItem{{ImportID: "10086", State: map[string]any{"id": float64(1)}}},
				NextPageToken: "next-tok",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	defer srv.Close()

	kt := kit.New()
	commonReq := proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "inst-1", TypeName: "bkcmdb_host"}

	impRes, err := cli.Import(kt, &proto.ImportReq{CommonInstanceReq: commonReq, ImportID: "10086"})
	assert.NoError(t, err)
	assert.Equal(t, float64(10086), impRes.State["bk_host_id"])

	refRes, err := cli.Refresh(kt, &proto.RefreshReq{
		CommonInstanceReq: commonReq,
		State:             map[string]any{"bk_host_id": 10086},
	})
	assert.NoError(t, err)
	assert.Equal(t, float64(10086), refRes.State["bk_host_id"])

	readRes, err := cli.ReadData(kt, &proto.ReadDataReq{
		CommonInstanceReq: commonReq,
		Config:            map[string]any{"bk_biz_name": "demo"},
	})
	assert.NoError(t, err)
	assert.Equal(t, float64(213), readRes.State["bk_biz_id"])

	listRes, err := cli.List(kt, &proto.ListReq{CommonInstanceReq: commonReq, Limit: 10})
	assert.NoError(t, err)
	assert.Len(t, listRes.Items, 1)
	assert.Equal(t, "next-tok", listRes.NextPageToken)
}
