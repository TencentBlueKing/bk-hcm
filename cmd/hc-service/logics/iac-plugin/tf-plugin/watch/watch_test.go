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

package watch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/cc"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
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

func rec(version string, state enumor.IaCPluginState) coreplugin.IaCPlugin {
	return coreplugin.IaCPlugin{PluginName: "tc", PluginVersion: version, State: state}
}

func TestIsPluginVersionBehind(t *testing.T) {
	const (
		enabled  = enumor.IaCPluginStateEnabled
		disabled = enumor.IaCPluginStateDisabled
		initS    = enumor.IaCPluginStateInit
		failedS  = enumor.IaCPluginStateFailed
	)

	cases := []struct {
		name       string
		local      string
		records    []coreplugin.IaCPlugin
		wantBehind bool
		wantVer    string
	}{
		{"behind", "1.0.0", []coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", enabled)}, true, "1.1.0"},
		{"same version", "1.0.0", []coreplugin.IaCPlugin{rec("1.0.0", enabled)}, false, "1.0.0"},
		{"local newer, own registration still running", "1.2.0",
			[]coreplugin.IaCPlugin{rec("1.0.0", enabled)}, false, "1.0.0"},
		{"a disabled higher version counts as completed", "1.0.0",
			[]coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", disabled)}, true, "1.1.0"},
		// 升级失败时旧副本仍是唯一可用版本，不能判成落后。
		{"a higher failed version does not count", "1.0.0",
			[]coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", failedS)}, false, "1.0.0"},
		{"a higher init version does not count", "1.0.0",
			[]coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", initS)}, false, "1.0.0"},
		{"no completed record", "1.0.0", []coreplugin.IaCPlugin{rec("1.0.0", initS)}, false, ""},
		{"no record at all", "1.0.0", nil, false, ""},
		{"semver, not string compare", "1.9.0", []coreplugin.IaCPlugin{rec("1.10.0", enabled)}, true, "1.10.0"},
		{"semver, not string compare (not behind)", "1.10.0",
			[]coreplugin.IaCPlugin{rec("1.9.0", enabled)}, false, "1.9.0"},
	}

	for _, c := range cases {
		behind, ver, err := isPluginVersionBehind(c.records, c.local)
		if err != nil {
			t.Errorf("%s: unexpected err: %v", c.name, err)
			continue
		}
		if behind != c.wantBehind || ver != c.wantVer {
			t.Errorf("%s: got behind=%v ver=%q, want behind=%v ver=%q", c.name, behind, ver, c.wantBehind, c.wantVer)
		}
	}

	if _, _, err := isPluginVersionBehind([]coreplugin.IaCPlugin{rec("not-a-version", enabled)}, "1.0.0"); err == nil {
		t.Error("an unparsable recorded version must be an error")
	}
}

func TestWatcherCheck(t *testing.T) {
	type roundResp struct {
		records []coreplugin.IaCPlugin
		fail    bool
	}

	const enabled = enumor.IaCPluginStateEnabled
	rounds := []roundResp{
		// 1: 落后 (1/3)
		{records: []coreplugin.IaCPlugin{rec("1.1.0", enabled)}},
		// 2: 读取失败，不计数，仍为 1/3
		{fail: true},
		// 3: 落后 (2/3)
		{records: []coreplugin.IaCPlugin{rec("1.1.0", enabled)}},
		// 4: 恢复一致，计数清零 (0/3)
		{records: []coreplugin.IaCPlugin{rec("1.0.0", enabled)}},
		// 5~7: 连续 3 次落后触发退出
		{records: []coreplugin.IaCPlugin{rec("1.1.0", enabled)}},
		{records: []coreplugin.IaCPlugin{rec("1.1.0", enabled)}},
		{records: []coreplugin.IaCPlugin{rec("1.1.0", enabled)}},
	}

	idx := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cur := rounds[idx]
		idx++
		if cur.fail {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    errf.Aborted,
				"message": "transient error",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": errf.OK,
			"data": &dataplugin.IaCPluginListResult{Details: cur.records},
		})
	}))
	defer srv.Close()

	cli := global.NewIaCPluginClient(rest.NewClient(&restcli.Capability{
		Client:   srv.Client(),
		Discover: fakeDiscover{host: srv.URL},
	}, "/"))

	called := 0
	w := New(cli, map[string]discovery.Entry{"tc": {PluginName: "tc", PluginVersion: "1.0.0"}},
		func() { called++ }, Options{})

	for i := 0; i < len(rounds)-1; i++ {
		if w.check(context.Background()) {
			t.Fatalf("round %d should not exit early", i+1)
		}
	}
	if !w.check(context.Background()) {
		t.Fatal("final round should trigger exit")
	}
	if called != 1 {
		t.Fatalf("onBehind called=%d, want 1", called)
	}
}
