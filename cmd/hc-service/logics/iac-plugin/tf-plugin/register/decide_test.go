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
	"testing"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	tabletypes "hcm/pkg/dal/table/types"
)

func rec(version string, state enumor.IaCPluginState) coreplugin.IaCPlugin {
	return coreplugin.IaCPlugin{ID: version, PluginName: "tc", PluginVersion: version, State: state,
		Protocol: enumor.IaCPluginProtocolTFProvider}
}

func TestDecide(t *testing.T) {
	const (
		enabled  = enumor.IaCPluginStateEnabled
		disabled = enumor.IaCPluginStateDisabled
		initS    = enumor.IaCPluginStateInit
		failed   = enumor.IaCPluginStateFailed
	)

	cases := []struct {
		name       string
		records    []coreplugin.IaCPlugin
		version    string
		want       enumor.IaCPluginAction
		wantRecVer string
	}{
		{"first register", nil, "1.83.34", enumor.IaCPluginActionRegister, ""},
		{"upgrade", []coreplugin.IaCPlugin{rec("1.83.30", enabled)}, "1.83.34", enumor.IaCPluginActionRegister, ""},
		{"rollback against enabled", []coreplugin.IaCPlugin{rec("1.83.34", enabled)}, "1.83.30", enumor.IaCPluginActionExit, ""},
		{"rollback against disabled", []coreplugin.IaCPlugin{rec("1.83.34", disabled)}, "1.83.30", enumor.IaCPluginActionExit, ""},
		{"same version enabled", []coreplugin.IaCPlugin{rec("1.83.34", enabled)}, "1.83.34", enumor.IaCPluginActionVerify, "1.83.34"},
		{"same version disabled", []coreplugin.IaCPlugin{rec("1.83.34", disabled)}, "1.83.34", enumor.IaCPluginActionVerify, "1.83.34"},
		{"re-enter init", []coreplugin.IaCPlugin{rec("1.83.34", initS)}, "1.83.34", enumor.IaCPluginActionRegister, "1.83.34"},
		{"re-enter failed", []coreplugin.IaCPlugin{rec("1.83.34", failed)}, "1.83.34", enumor.IaCPluginActionRegister, "1.83.34"},
		// 升级失败或进行中时，仍在服务的旧副本不能被判成版本回退。
		{"higher failed does not exit old replica",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.30", enumor.IaCPluginActionVerify, "1.83.30"},
		{"higher init does not exit old replica",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", initS)}, "1.83.30", enumor.IaCPluginActionVerify, "1.83.30"},
		{"upgrade retry points to same-version failed record",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.34", enumor.IaCPluginActionRegister, "1.83.34"},
		{"semver upgrade 1.9.0 -> 1.10.0", []coreplugin.IaCPlugin{rec("1.9.0", enabled)}, "1.10.0", enumor.IaCPluginActionRegister, ""},
		{"semver rollback 1.10.0 -> 1.9.0", []coreplugin.IaCPlugin{rec("1.10.0", enabled)}, "1.9.0", enumor.IaCPluginActionExit, ""},
	}

	for _, c := range cases {
		got, err := Decide(c.records, c.version)
		if err != nil {
			t.Errorf("%s: unexpected err: %v", c.name, err)
			continue
		}
		if got.Action != c.want {
			t.Errorf("%s: action = %s (%s), want %s", c.name, got.Action, got.Reason, c.want)
		}
		gotRecVer := ""
		if got.Record != nil {
			gotRecVer = got.Record.PluginVersion
		}
		if gotRecVer != c.wantRecVer {
			t.Errorf("%s: record version = %q, want %q", c.name, gotRecVer, c.wantRecVer)
		}
	}
}

func TestDecideErrors(t *testing.T) {
	httpRec := coreplugin.IaCPlugin{ID: "h", PluginName: "tc", PluginVersion: "1.0.0",
		Protocol: enumor.IaCPluginProtocolHTTP, State: enumor.IaCPluginStateEnabled}
	httpHigh := httpRec
	httpHigh.PluginVersion = "2.0.0"

	cases := []struct {
		name    string
		records []coreplugin.IaCPlugin
		version string
	}{
		{"same version http protocol mismatch", []coreplugin.IaCPlugin{httpRec}, "1.0.0"},
		{"higher http protocol mismatch", []coreplugin.IaCPlugin{httpHigh}, "1.0.0"},
		{"invalid local version", nil, "latest"},
		{"invalid recorded version", []coreplugin.IaCPlugin{rec("not-a-version", enumor.IaCPluginStateEnabled)}, "1.0.0"},
		{"unknown state", []coreplugin.IaCPlugin{{ID: "x", PluginVersion: "1.0.0", State: "weird"}}, "1.0.0"},
	}
	for _, c := range cases {
		if _, err := Decide(c.records, c.version); err == nil {
			t.Errorf("%s: want error, got nil", c.name)
		}
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

func TestVerifyHash(t *testing.T) {
	cases := []struct {
		name string
		ext  tabletypes.JsonField
		want enumor.IaCPluginHashResult
	}{
		{"match", extOf(t, map[string]string{"linux_amd64": "aaa"}), enumor.IaCPluginHashResultMatch},
		{"mismatch", extOf(t, map[string]string{"linux_amd64": "bbb"}), enumor.IaCPluginHashResultMismatch},
		{"other arch only", extOf(t, map[string]string{"linux_arm64": "ccc"}), enumor.IaCPluginHashResultMissing},
		{"empty arch hash", extOf(t, map[string]string{"linux_amd64": ""}), enumor.IaCPluginHashResultMissing},
		{"no architectures key", tabletypes.JsonField(`{}`), enumor.IaCPluginHashResultMissing},
		{"empty extension", "", enumor.IaCPluginHashResultMissing},
	}

	for _, c := range cases {
		got, err := VerifyHash(&coreplugin.IaCPlugin{ID: "1", Extension: c.ext}, "linux_amd64", "aaa")
		if err != nil || got != c.want {
			t.Errorf("%s: got %v err=%v, want %v", c.name, got, err, c.want)
		}
	}

	if _, err := VerifyHash(&coreplugin.IaCPlugin{ID: "1", Extension: tabletypes.JsonField(`not json`)}, "linux_amd64", "aaa"); err == nil {
		t.Error("invalid extension want err")
	}
	if _, err := VerifyHash(nil, "linux_amd64", "aaa"); err == nil {
		t.Error("nil record want err")
	}
}

func TestWithArchHash(t *testing.T) {
	// 补写本架构时，其它架构和其它扩展字段必须原样保留。
	in := tabletypes.JsonField(`{"architectures":{"linux_arm64":"ccc"},"other":{"k":1}}`)
	out, err := WithArchHash(in, "linux_amd64", "aaa")
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

	for _, empty := range []tabletypes.JsonField{"", tabletypes.JsonField(`{}`), tabletypes.JsonField(`{"architectures":null}`)} {
		out, err = WithArchHash(empty, "linux_amd64", "aaa")
		if err != nil {
			t.Fatalf("from %s: %v", empty, err)
		}
		res, err := VerifyHash(&coreplugin.IaCPlugin{ID: "1", Extension: out}, "linux_amd64", "aaa")
		if err != nil || res != enumor.IaCPluginHashResultMatch {
			t.Fatalf("from %s: round trip got %v err=%v", empty, res, err)
		}
	}

	if _, err = WithArchHash(tabletypes.JsonField(`[1]`), "linux_amd64", "aaa"); err == nil {
		t.Error("non-object extension want err")
	}
}
