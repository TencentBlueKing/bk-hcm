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
	"reflect"
	"strings"
	"testing"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/serviced"
)

type stubDiscover struct{ addr string }

func (d stubDiscover) Discover(cc.Name) ([]string, error)              { return []string{d.addr}, nil }
func (d stubDiscover) Services() []cc.Name                             { return []cc.Name{cc.DataServiceName} }
func (d stubDiscover) ByLabels([]string) serviced.Discover             { return d }
func (d stubDiscover) GetServiceAllNodeKeys(cc.Name) ([]string, error) { return nil, nil }

func TestDecideSwitch(t *testing.T) {
	const (
		initS    = enumor.IaCPluginStateInit
		enabled  = enumor.IaCPluginStateEnabled
		disabled = enumor.IaCPluginStateDisabled
		failedS  = enumor.IaCPluginStateFailed
	)

	rec := func(state enumor.IaCPluginState, msg string) *coreplugin.IaCPlugin {
		return &coreplugin.IaCPlugin{
			PluginName:    "tencentcloud",
			PluginVersion: "1.83.34",
			Protocol:      enumor.IaCPluginProtocolTFProvider,
			State:         state,
			Message:       msg,
		}
	}

	cases := []struct {
		name         string
		current      *coreplugin.IaCPlugin
		target       enumor.IaCPluginState
		schemaExists bool
		wantFrom     []enumor.IaCPluginState
		wantErrSub   string
	}{
		{"enabled -> disabled", rec(enabled, ""), disabled, false, []enumor.IaCPluginState{enabled, failedS}, ""},
		{"enabled -> enabled (idempotent)", rec(enabled, ""), enabled, false, nil, ""},
		{"disabled -> disabled (idempotent)", rec(disabled, ""), disabled, false, nil, ""},
		{"disabled -> enabled with schema", rec(disabled, ""), enabled, true, []enumor.IaCPluginState{disabled}, ""},
		{"disabled -> enabled without schema", rec(disabled, ""), enabled, false, nil, "schema does not exist"},
		{"failed -> disabled", rec(failedS, "schema sync failed"), disabled, false, []enumor.IaCPluginState{enabled, failedS}, ""},
		{"failed -> enabled", rec(failedS, "schema sync failed"), enabled, true, nil, "schema sync failed"},
		{"init -> disabled", rec(initS, ""), disabled, true, nil, "not finished"},
		{"init -> enabled", rec(initS, ""), enabled, true, nil, "not finished"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				details := []map[string]any{}
				if c.schemaExists {
					details = append(details, map[string]any{"id": "1"})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"code": 0,
					"data": map[string]any{"details": details},
				})
			}))
			defer srv.Close()

			s := &svc{client: client.NewClientSet(srv.Client(), stubDiscover{addr: srv.URL})}
			gotFrom, err := s.decideSwitch(kit.New(), c.current, c.target)
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
