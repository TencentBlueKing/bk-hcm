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

package register

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
)

func TestMain(m *testing.M) {
	cc.InitRuntime(new(cc.HCServiceSetting))
	m.Run()
}

type fakeDiscover struct {
	host string
}

func (f fakeDiscover) GetServers() ([]string, error) {
	return []string{f.host}, nil
}

func newTestService(t *testing.T, handler http.HandlerFunc) *Service {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	dataCli := &dataservice.Client{
		Global: global.NewClient(rest.NewClient(&restcli.Capability{
			Client:   srv.Client(),
			Discover: fakeDiscover{host: srv.URL},
		}, "/")),
	}
	return New(dataCli, nil, nil, Options{})
}

func rec(version string, state enumor.IaCPluginState) coreplugin.IaCPlugin {
	return coreplugin.IaCPlugin{
		ID:            version,
		PluginName:    "tc",
		PluginVersion: version,
		State:         state,
		Protocol:      enumor.IaCPluginProtocolTFProvider,
		Extension:     `{"architectures":{"linux_amd64":"aaa"}}`,
	}
}

func extOf(t *testing.T, archs map[string]string) tabletypes.JsonField {
	t.Helper()
	b, err := json.Marshal(map[string]interface{}{"architectures": archs})
	if err != nil {
		t.Fatal(err)
	}
	return tabletypes.JsonField(b)
}

func TestRegisterWithRecords(t *testing.T) {
	const (
		enabled  = enumor.IaCPluginStateEnabled
		disabled = enumor.IaCPluginStateDisabled
		initS    = enumor.IaCPluginStateInit
		failed   = enumor.IaCPluginStateFailed
	)

	// wantAction: "verify"（直接通过无写库）、"create"（新建 init）、"reset"（重置已有记录为 init）、
	// "conflict"（版本回退 errPluginVersionConflict）、"err"（参数/协议/状态非法错误）。
	httpRec := coreplugin.IaCPlugin{ID: "h", PluginName: "tc", PluginVersion: "1.0.0",
		Protocol: enumor.IaCPluginProtocolHTTP, State: enabled}
	httpHigh := httpRec
	httpHigh.PluginVersion = "2.0.0"

	cases := []struct {
		name       string
		records    []coreplugin.IaCPlugin
		version    string
		wantAction string
		wantReset  string
	}{
		{"first register", nil, "1.83.34", "create", ""},
		{"upgrade", []coreplugin.IaCPlugin{rec("1.83.30", enabled)}, "1.83.34", "create", ""},
		{"rollback against enabled", []coreplugin.IaCPlugin{rec("1.83.34", enabled)}, "1.83.30", "conflict", ""},
		{"rollback against disabled", []coreplugin.IaCPlugin{rec("1.83.34", disabled)}, "1.83.30", "conflict", ""},
		{"same version enabled", []coreplugin.IaCPlugin{rec("1.83.34", enabled)}, "1.83.34", "verify", ""},
		{"same version disabled", []coreplugin.IaCPlugin{rec("1.83.34", disabled)}, "1.83.34", "verify", ""},
		{"re-enter init", []coreplugin.IaCPlugin{rec("1.83.34", initS)}, "1.83.34", "reset", "1.83.34"},
		{"re-enter failed", []coreplugin.IaCPlugin{rec("1.83.34", failed)}, "1.83.34", "reset", "1.83.34"},
		// 升级失败或进行中时，仍在服务的旧副本不能被判成版本回退。
		{"higher failed does not exit old replica",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.30", "verify", ""},
		{"higher init does not exit old replica",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", initS)}, "1.83.30", "verify", ""},
		{"upgrade retry points to same-version failed record",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.34", "reset", "1.83.34"},
		// 未注册的中间版本与低版本失败记录不得低于全量最高版本（含 init / failed）。
		{"intermediate version blocked by higher init",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", initS)}, "1.83.32", "conflict", ""},
		{"intermediate version blocked by higher failed",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.32", "conflict", ""},
		{"lower failed blocked by higher init",
			[]coreplugin.IaCPlugin{rec("1.83.30", failed), rec("1.83.34", initS)}, "1.83.30", "conflict", ""},
		{"existing enabled rollback against higher enabled",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", enabled)}, "1.83.30", "conflict", ""},
		{"existing enabled allowed when highest completed is disabled",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.32", enabled), rec("1.83.34", disabled)},
			"1.83.30", "verify", ""},
		{"semver upgrade 1.9.0 -> 1.10.0", []coreplugin.IaCPlugin{rec("1.9.0", enabled)}, "1.10.0", "create", ""},
		{"semver rollback 1.10.0 -> 1.9.0", []coreplugin.IaCPlugin{rec("1.10.0", enabled)}, "1.9.0", "conflict", ""},
		{"same version http protocol mismatch", []coreplugin.IaCPlugin{httpRec}, "1.0.0", "err", ""},
		{"higher http protocol mismatch", []coreplugin.IaCPlugin{httpHigh}, "1.0.0", "err", ""},
		{"invalid local version", nil, "latest", "err", ""},
		{"invalid recorded version", []coreplugin.IaCPlugin{rec("not-a-version", enabled)}, "1.0.0", "err", ""},
		{"unknown state", []coreplugin.IaCPlugin{{
			ID: "x", PluginName: "tc", PluginVersion: "1.0.0",
			Protocol: enumor.IaCPluginProtocolTFProvider, State: "weird",
		}}, "1.0.0", "err", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var calledPath string
			var updateReq dataplugin.IaCPluginUpdateReq

			svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
				calledPath = r.URL.Path
				if r.URL.Path == "/iac_plugins/update" {
					_ = json.NewDecoder(r.Body).Decode(&updateReq)
				}
				// 截断在 saveInitRecord，不拉起本地 TF 子进程。
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"code":    errf.Aborted,
					"message": "stop after saveInitRecord",
				})
			})

			entry := discovery.PluginEntry{
				PluginName:    "tc",
				PluginVersion: c.version,
				Arch:          "linux_amd64",
				SHA256:        "aaa",
			}
			err := svc.registerWithRecords(core.NewBackendKit(), entry, c.records)

			switch c.wantAction {
			case "verify":
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if calledPath != "" {
					t.Fatalf("expected no write RPC on matched verify, got %s", calledPath)
				}
			case "create":
				if calledPath != "/iac_plugins/create" {
					t.Fatalf("calledPath = %q, want /iac_plugins/create (err=%v)", calledPath, err)
				}
			case "reset":
				if calledPath != "/iac_plugins/update" {
					t.Fatalf("calledPath = %q, want /iac_plugins/update (err=%v)", calledPath, err)
				}
				if updateReq.ID != c.wantReset || updateReq.State != enumor.IaCPluginStateInit {
					t.Fatalf("updateReq = %+v, want ID=%s State=init", updateReq, c.wantReset)
				}
			case "conflict":
				if !errors.Is(err, errPluginVersionConflict) {
					t.Fatalf("err = %v, want errPluginVersionConflict", err)
				}
				if calledPath != "" {
					t.Fatalf("expected no write RPC on conflict, got %s", calledPath)
				}
			case "err":
				if err == nil || errors.Is(err, errPluginVersionConflict) {
					t.Fatalf("err = %v, want non-conflict error", err)
				}
				if calledPath != "" {
					t.Fatalf("expected no write RPC on validation error, got %s", calledPath)
				}
			}
		})
	}
}

func TestVerifyRecord(t *testing.T) {
	entry := discovery.PluginEntry{
		PluginName:    "tc",
		PluginVersion: "1.0.0",
		Arch:          "linux_amd64",
		SHA256:        "aaa",
	}

	t.Run("match", func(t *testing.T) {
		updated := false
		svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
			updated = true
		})
		r := &coreplugin.IaCPlugin{ID: "1", State: enumor.IaCPluginStateEnabled,
			Extension: extOf(t, map[string]string{"linux_amd64": "aaa"})}
		if err := svc.verifyRecord(core.NewBackendKit(), entry, r); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if updated {
			t.Fatal("should not update when sha256 matches")
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		updated := false
		svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
			updated = true
		})
		r := &coreplugin.IaCPlugin{ID: "1", State: enumor.IaCPluginStateEnabled,
			Extension: extOf(t, map[string]string{"linux_amd64": "bbb"})}
		err := svc.verifyRecord(core.NewBackendKit(), entry, r)
		if !errors.Is(err, errPluginVersionConflict) {
			t.Fatalf("err = %v, want errPluginVersionConflict", err)
		}
		if updated {
			t.Fatal("should not update on sha256 mismatch")
		}
	})

	missingCases := []struct {
		name string
		ext  tabletypes.JsonField
	}{
		{"other arch only", extOf(t, map[string]string{"linux_arm64": "ccc"})},
		{"empty arch hash", extOf(t, map[string]string{"linux_amd64": ""})},
		{"no architectures key", `{}`},
		{"empty extension", ""},
	}
	for _, c := range missingCases {
		t.Run("backfill "+c.name, func(t *testing.T) {
			var gotReq dataplugin.IaCPluginUpdateReq
			svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&gotReq)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"code": errf.OK})
			})
			r := &coreplugin.IaCPlugin{ID: "1", State: enumor.IaCPluginStateEnabled, Extension: c.ext}
			if err := svc.verifyRecord(core.NewBackendKit(), entry, r); err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			var ext coreplugin.TFProviderExtension
			if err := json.Unmarshal([]byte(gotReq.Extension), &ext); err != nil {
				t.Fatalf("invalid updated extension: %v", err)
			}
			if ext.Architectures["linux_amd64"] != "aaa" {
				t.Fatalf("architectures = %v, want linux_amd64=aaa", ext.Architectures)
			}
		})
	}

	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	if err := svc.verifyRecord(core.NewBackendKit(), entry,
		&coreplugin.IaCPlugin{ID: "1", Extension: `not json`}); err == nil {
		t.Error("invalid extension want err")
	}
	if err := svc.verifyRecord(core.NewBackendKit(), entry, nil); err == nil {
		t.Error("nil record want err")
	}
}

func TestWithArchHash(t *testing.T) {
	// 补写本架构时，其它架构和其它扩展字段必须原样保留。
	in := tabletypes.JsonField(`{"architectures":{"linux_arm64":"ccc"},"other":{"k":1}}`)
	out, err := withArchHash(in, "linux_amd64", "aaa")
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Architectures map[string]string      `json:"architectures"`
		Other         map[string]interface{} `json:"other"`
	}
	if err = json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Architectures["linux_amd64"] != "aaa" || got.Architectures["linux_arm64"] != "ccc" {
		t.Fatalf("architectures = %v", got.Architectures)
	}
	if got.Other["k"] != float64(1) {
		t.Fatalf("other fields must be preserved, got %v", got.Other)
	}

	for _, empty := range []tabletypes.JsonField{"", `{}`, `{"architectures":null}`} {
		out, err = withArchHash(empty, "linux_amd64", "aaa")
		if err != nil {
			t.Fatalf("from %s: %v", empty, err)
		}
		var ext coreplugin.TFProviderExtension
		if err = json.Unmarshal([]byte(out), &ext); err != nil {
			t.Fatalf("from %s: unmarshal %v", empty, err)
		}
		if ext.Architectures["linux_amd64"] != "aaa" {
			t.Fatalf("from %s: architectures = %v", empty, ext.Architectures)
		}
	}

	if _, err = withArchHash(`[1]`, "linux_amd64", "aaa"); err == nil {
		t.Error("non-object extension want err")
	}
}
