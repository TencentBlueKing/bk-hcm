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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	proto "hcm/pkg/api/cloud-server"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"

	"github.com/emicklei/go-restful/v3"
)

type stubAuthorizer struct {
	err error
}

func (a stubAuthorizer) Authorize(_ *kit.Kit, _ ...meta.ResourceAttribute) ([]meta.Decision, bool, error) {
	return nil, a.err == nil, a.err
}

func (a stubAuthorizer) AuthorizeAny(_ *kit.Kit, _ ...meta.ResourceAttribute) ([]meta.Decision, error) {
	return nil, a.err
}

func (a stubAuthorizer) AuthorizeWithPerm(_ *kit.Kit, _ ...meta.ResourceAttribute) error {
	return a.err
}

func (a stubAuthorizer) ListAuthorizedInstances(_ *kit.Kit, _ *meta.ListAuthResInput) (
	*meta.AuthorizedInstances, error) {
	return nil, a.err
}

func (a stubAuthorizer) ListAuthInstWithFilter(_ *kit.Kit, _ *meta.ListAuthResInput, expr *filter.Expression,
	_ string) (*filter.Expression, bool, error) {
	return expr, false, a.err
}

func (a stubAuthorizer) RegisterResourceCreatorAction(_ *kit.Kit, _ *meta.RegisterResCreatorActionInst) error {
	return a.err
}

func (a stubAuthorizer) GetPermissionToApply(_ *kit.Kit, _ ...meta.ResourceAttribute) (*meta.IamPermission, error) {
	return nil, a.err
}

func (a stubAuthorizer) GetApplyPermUrl(_ *kit.Kit, _ *meta.IamPermission) (string, error) {
	return "", a.err
}

var _ auth.Authorizer = stubAuthorizer{}

type stubDiscover struct{ addr string }

func (d stubDiscover) Discover(cc.Name) ([]string, error)              { return []string{d.addr}, nil }
func (d stubDiscover) Services() []cc.Name                             { return []cc.Name{cc.DataServiceName} }
func (d stubDiscover) ByLabels([]string) serviced.Discover             { return d }
func (d stubDiscover) GetServiceAllNodeKeys(cc.Name) ([]string, error) { return nil, nil }

func newTestSvc(srv *httptest.Server, authErr error) *svc {
	cs := client.NewClientSet(srv.Client(), stubDiscover{addr: srv.URL})
	return &svc{
		client:     cs,
		authorizer: stubAuthorizer{err: authErr},
	}
}

func newTestContext(t *testing.T, body any) *rest.Contexts {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, "http://test/iac_plugins", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	return &rest.Contexts{Kit: kit.New(), Request: restful.NewRequest(httpReq)}
}

func TestListIaCPlugin(t *testing.T) {
	validBody := map[string]any{
		"filter": map[string]any{"op": "and", "rules": []any{}},
		"page":   map[string]any{"count": false, "start": 0, "limit": 10},
	}

	t.Run("no permission does not call data service", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("should not call data service when unauthorized: %s", r.URL.Path)
		}))
		defer srv.Close()

		s := newTestSvc(srv, errf.New(errf.PermissionDenied, "no permission"))
		if _, err := s.ListIaCPlugin(newTestContext(t, validBody)); err == nil {
			t.Fatal("ListIaCPlugin want permission error, got nil")
		}
	})

	t.Run("success returns list result", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    0,
				"message": "",
				"data": map[string]any{
					"count": 0,
					"details": []map[string]any{
						{
							"id":             "00000001",
							"plugin_name":    "tencentcloud",
							"plugin_version": "1.83.34",
							"protocol":       enumor.IaCPluginProtocolTFProvider,
							"state":          enumor.IaCPluginStateEnabled,
							"extension":      map[string]any{},
						},
					},
				},
			})
		}))
		defer srv.Close()

		s := newTestSvc(srv, nil)
		got, err := s.ListIaCPlugin(newTestContext(t, validBody))
		if err != nil {
			t.Fatalf("ListIaCPlugin unexpected err: %v", err)
		}
		res := got.(*proto.IaCPluginListResult)
		if len(res.Details) != 1 || res.Details[0].PluginName != "tencentcloud" {
			t.Fatalf("ListIaCPlugin details = %+v, want 1 tencentcloud record", res.Details)
		}
	})
}

func TestSwitchIaCPluginValidationAndAuth(t *testing.T) {
	t.Run("invalid target state rejected before downstream call", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("should not call data service on invalid state: %s", r.URL.Path)
		}))
		defer srv.Close()

		s := newTestSvc(srv, nil)
		body := map[string]any{"plugin_name": "tencentcloud", "state": "init"}
		if _, err := s.SwitchIaCPlugin(newTestContext(t, body)); err == nil {
			t.Fatal("SwitchIaCPlugin(state=init) want err, got nil")
		}
	})

	t.Run("no permission rejected before downstream call", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("should not call data service when unauthorized: %s", r.URL.Path)
		}))
		defer srv.Close()

		s := newTestSvc(srv, errf.New(errf.PermissionDenied, "no permission"))
		body := map[string]any{"plugin_name": "tencentcloud", "state": "disabled"}
		if _, err := s.SwitchIaCPlugin(newTestContext(t, body)); err == nil {
			t.Fatal("SwitchIaCPlugin want permission error, got nil")
		}
	})
}

func TestDecideSwitch(t *testing.T) {
	const (
		initS    = enumor.IaCPluginStateInit
		enabled  = enumor.IaCPluginStateEnabled
		disabled = enumor.IaCPluginStateDisabled
		failedS  = enumor.IaCPluginStateFailed

		tfProto   = enumor.IaCPluginProtocolTFProvider
		httpProto = enumor.IaCPluginProtocolHTTP
	)

	cases := []struct {
		name         string
		current      *coreplugin.IaCPlugin
		target       enumor.IaCPluginState
		schemaExists bool
		wantFrom     []enumor.IaCPluginState
		wantErrSub   string
	}{
		{
			name:         "enabled -> disabled",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: enabled},
			target:       disabled,
			schemaExists: false,
			wantFrom:     []enumor.IaCPluginState{enabled, failedS},
		},
		{
			name:         "enabled -> enabled (idempotent)",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: enabled},
			target:       enabled,
			schemaExists: false,
			wantFrom:     nil,
		},
		{
			name:         "disabled -> disabled (idempotent)",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: disabled},
			target:       disabled,
			schemaExists: false,
			wantFrom:     nil,
		},
		{
			name:         "disabled -> enabled with schema",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: disabled},
			target:       enabled,
			schemaExists: true,
			wantFrom:     []enumor.IaCPluginState{disabled},
		},
		{
			name:         "disabled -> enabled without schema",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: disabled},
			target:       enabled,
			schemaExists: false,
			wantErrSub:   "schema does not exist",
		},
		{
			name:         "failed -> disabled",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: failedS, Message: "schema sync failed"},
			target:       disabled,
			schemaExists: false,
			wantFrom:     []enumor.IaCPluginState{enabled, failedS},
		},
		{
			name:         "failed -> enabled",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: failedS, Message: "schema sync failed"},
			target:       enabled,
			schemaExists: true,
			wantErrSub:   "schema sync failed",
		},
		{
			name:         "init -> disabled",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: tfProto, State: initS},
			target:       disabled,
			schemaExists: true,
			wantErrSub:   "not finished",
		},
		{
			name:         "init -> enabled",
			current:      &coreplugin.IaCPlugin{PluginName: "tencentcloud", PluginVersion: "1.83.34", Protocol: httpProto, State: initS},
			target:       enabled,
			schemaExists: true,
			wantErrSub:   "not finished",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotFrom, err := decideSwitch(c.current, c.target, c.schemaExists)
			if len(c.wantErrSub) > 0 {
				if err == nil || !strings.Contains(err.Error(), c.wantErrSub) {
					t.Fatalf("decideSwitch err = %v, want substring %q", err, c.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("decideSwitch unexpected err: %v", err)
			}
			if !reflect.DeepEqual(gotFrom, c.wantFrom) {
				t.Fatalf("decideSwitch = %v, want %v", gotFrom, c.wantFrom)
			}
		})
	}
}
