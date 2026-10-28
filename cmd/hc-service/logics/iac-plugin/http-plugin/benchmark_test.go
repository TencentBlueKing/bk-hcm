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

package httpplugin_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	"hcm/cmd/hc-service/logics/iac-plugin/router"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	restcli "hcm/pkg/rest/client"

	tfjson "github.com/hashicorp/terraform-json"
)

var (
	p99CheckOnce sync.Once
	cachedP99Ms  float64
	benchKeyOnce sync.Once
	benchPubPEM  string
)

func generateBenchRSAPubPEM(b testing.TB) string {
	b.Helper()
	benchKeyOnce.Do(func() {
		priv, err := rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			b.Fatalf("generate rsa key failed: %v", err)
		}
		der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			b.Fatalf("marshal public key failed: %v", err)
		}
		benchPubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	})
	return benchPubPEM
}

type benchHTTP struct{ handler http.HandlerFunc }

func (c *benchHTTP) Do(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	c.handler(rec, req)
	return rec.Result(), nil
}

type benchDisc struct{}

func (benchDisc) GetServers() ([]string, error) { return []string{"http://127.0.0.1"}, nil }

func setupBenchRouterAndExecutor(b testing.TB, pubPEM, secretVal string) (*router.Service, *httpplugin.Service) {
	b.Helper()
	extBytes, _ := json.Marshal(coreplugin.HTTPExtension{
		APIPath:      "/api/bkcmdb",
		CryptoType:   "RSA",
		CryptoConfig: &coreplugin.HTTPCryptoConfig{PublicKey: pubPEM},
	})
	detail := coreplugin.SchemaDetail{
		ProtocolVersion: "1.0",
		Capabilities:    coreplugin.Capabilities{Apply: true, Plan: true, Refresh: true, ReadData: true},
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema: &tfjson.Schema{Block: &tfjson.SchemaBlock{}},
			ResourceSchemas: map[string]*tfjson.Schema{
				"bkcmdb_host": {Block: &tfjson.SchemaBlock{}},
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
					ID: "inst-1", PluginName: "bkcmdb",
					Attrs:       map[string]any{"host": "https://cmdb.example.com"},
					SecretAttrs: map[string]any{"secret_key": secretVal},
				}}},
			})
		}
	}
	dataCli := dataservice.NewClient(&restcli.Capability{
		Client:   &benchHTTP{handler: handler},
		Discover: benchDisc{},
	}, "v1")
	schemaMgr := pluginschema.NewSchemaManager(dataCli, pluginschema.DefaultSchemaCacheTTL)
	httpPlugin := httpplugin.New(dataCli, nil, nil, schemaMgr, httpplugin.Options{})
	return router.New(dataCli, schemaMgr, httpPlugin, nil, nil), httpPlugin
}

// BenchmarkNorthboundRouteAndEncrypt benchmarks 1,000 iterations of Router.Resolve + RSA-OAEP launch_config encryption
// under SchemaCache hit and mock data-service, verifying P99 latency < 10ms (AC-P01).
func BenchmarkNorthboundRouteAndEncrypt(b *testing.B) {
	pubPEM := generateBenchRSAPubPEM(b)
	r, exec := setupBenchRouterAndExecutor(b, pubPEM, "tok_bench_123")
	kt := kit.New()

	// Warm up SchemaCache and RSA encryptor cache.
	benchCommon := proto.CommonInstanceReq{
		Plugin: "bkcmdb", PluginInstanceID: "inst-1", TypeName: "bkcmdb_host",
	}
	warmCtx, err := r.Resolve(kt, benchCommon, enumor.IaCPluginOperationExecute)
	if err != nil {
		b.Fatalf("warm up resolve failed: %v", err)
	}
	if _, _, err := exec.EncryptLaunchConfig(kt, warmCtx.Plugin, warmCtx.Properties); err != nil {
		b.Fatalf("warm up encrypt failed: %v", err)
	}

	p99CheckOnce.Do(func() {
		const sampleCount = 1000
		durations := make([]time.Duration, 0, sampleCount)
		for i := 0; i < sampleCount; i++ {
			start := time.Now()
			ctx, err := r.Resolve(kt, benchCommon, enumor.IaCPluginOperationExecute)
			if err != nil {
				b.Fatalf("resolve failed: %v", err)
			}
			if _, _, err := exec.EncryptLaunchConfig(kt, ctx.Plugin, ctx.Properties); err != nil {
				b.Fatalf("encrypt failed: %v", err)
			}
			durations = append(durations, time.Since(start))
		}

		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		p99 := durations[sampleCount*99/100]
		cachedP99Ms = float64(p99.Microseconds()) / 1000.0
		if p99 >= 10*time.Millisecond {
			b.Fatalf("expected P99 < 10ms over 1,000 iterations, got %v", p99)
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, err := r.Resolve(kt, benchCommon, enumor.IaCPluginOperationExecute)
		if err != nil {
			b.Fatalf("resolve failed: %v", err)
		}
		if _, _, err := exec.EncryptLaunchConfig(kt, ctx.Plugin, ctx.Properties); err != nil {
			b.Fatalf("encrypt failed: %v", err)
		}
	}
	b.ReportMetric(cachedP99Ms, "p99_ms")
}
