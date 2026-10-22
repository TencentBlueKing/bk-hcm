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
		name    string
		records []coreplugin.IaCPlugin
		version string
		want    IaCPluginAction
	}{
		{"first register", nil, "1.83.34", IaCPluginActionRegister},
		{"upgrade: higher than the completed one", []coreplugin.IaCPlugin{rec("1.83.30", enabled)}, "1.83.34", IaCPluginActionRegister},
		{"rollback: lower than a completed one", []coreplugin.IaCPlugin{rec("1.83.34", enabled)}, "1.83.30", IaCPluginActionExit},
		{"rollback against a disabled one", []coreplugin.IaCPlugin{rec("1.83.34", disabled)}, "1.83.30", IaCPluginActionExit},
		{"same version enabled: only verify hash", []coreplugin.IaCPlugin{rec("1.83.34", enabled)}, "1.83.34", IaCPluginActionVerify},
		{"same version disabled: only verify hash", []coreplugin.IaCPlugin{rec("1.83.34", disabled)}, "1.83.34", IaCPluginActionVerify},
		{"interrupted init is re-entered", []coreplugin.IaCPlugin{rec("1.83.34", initS)}, "1.83.34", IaCPluginActionRegister},
		{"failed is re-entered", []coreplugin.IaCPlugin{rec("1.83.34", failed)}, "1.83.34", IaCPluginActionRegister},
		{
			// 升级失败（高版本 failed）不能让仍在服务的旧副本被判成版本回退。
			"a failed higher version must not make an old replica exit",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.30", IaCPluginActionVerify,
		},
		{
			"a higher init must not make an old replica exit",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", initS)}, "1.83.30", IaCPluginActionVerify,
		},
		{
			"upgrade retry: old enabled, new failed",
			[]coreplugin.IaCPlugin{rec("1.83.30", enabled), rec("1.83.34", failed)}, "1.83.34", IaCPluginActionRegister,
		},
		{
			// 字符串比较会认为 1.9.0 > 1.10.0。
			"semver not string compare",
			[]coreplugin.IaCPlugin{rec("1.9.0", enabled)}, "1.10.0", IaCPluginActionRegister,
		},
		{
			"semver not string compare: rollback",
			[]coreplugin.IaCPlugin{rec("1.10.0", enabled)}, "1.9.0", IaCPluginActionExit,
		},
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
	}
}

func TestDecideRecordRef(t *testing.T) {
	records := []coreplugin.IaCPlugin{rec("1.83.30", enabled()), rec("1.83.34", enumor.IaCPluginStateFailed)}

	got, err := Decide(records, "1.83.34")
	if err != nil {
		t.Fatal(err)
	}
	// Record 必须指向同版本那条，而不是最高版本或第一条。
	if got.Record == nil || got.Record.PluginVersion != "1.83.34" {
		t.Fatalf("record should be the same-version one, got %+v", got.Record)
	}

	got, err = Decide(nil, "1.83.34")
	if err != nil || got.Record != nil {
		t.Fatalf("insert has no record, got %+v err=%v", got, err)
	}
}

func enabled() enumor.IaCPluginState { return enumor.IaCPluginStateEnabled }

func TestDecideProtocolMismatch(t *testing.T) {
	httpRec := coreplugin.IaCPlugin{ID: "h", PluginName: "tc", PluginVersion: "1.0.0",
		Protocol: enumor.IaCPluginProtocolHTTP, State: enumor.IaCPluginStateEnabled}

	if _, err := Decide([]coreplugin.IaCPlugin{httpRec}, "1.0.0"); err == nil {
		t.Fatal("a record registered as http must not be treated as a tf provider")
	}
	// 更高版本的 http 记录同样要报错，而不是被当成「版本回退」静默退出。
	httpHigh := httpRec
	httpHigh.PluginVersion = "2.0.0"
	if _, err := Decide([]coreplugin.IaCPlugin{httpHigh}, "1.0.0"); err == nil {
		t.Fatal("a higher http record must be reported, not treated as a rollback")
	}
}

func TestDecideInvalid(t *testing.T) {
	if _, err := Decide(nil, "latest"); err == nil {
		t.Error("invalid local version want err")
	}

	if _, err := Decide([]coreplugin.IaCPlugin{rec("not-a-version", enumor.IaCPluginStateEnabled)}, "1.0.0"); err == nil {
		t.Error("invalid recorded version want err")
	}

	bad := coreplugin.IaCPlugin{ID: "x", PluginVersion: "1.0.0", State: "weird"}
	if _, err := Decide([]coreplugin.IaCPlugin{bad}, "1.0.0"); err == nil {
		t.Error("unknown state want err")
	}
}

func extOf(t *testing.T, archs map[string]string) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(map[string]interface{}{"architectures": archs})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestVerifyHash(t *testing.T) {
	cases := []struct {
		name string
		ext  json.RawMessage
		want IaCPluginHashResult
	}{
		{"match", extOf(t, map[string]string{"linux_amd64": "aaa"}), IaCPluginHashResultMatch},
		{"mismatch", extOf(t, map[string]string{"linux_amd64": "bbb"}), IaCPluginHashResultMismatch},
		{"arch missing: only other arch recorded", extOf(t, map[string]string{"linux_arm64": "ccc"}), IaCPluginHashResultMissing},
		{"empty arch value counts as missing", extOf(t, map[string]string{"linux_amd64": ""}), IaCPluginHashResultMissing},
		{"no architectures key", json.RawMessage(`{}`), IaCPluginHashResultMissing},
		{"empty extension", nil, IaCPluginHashResultMissing},
	}

	for _, c := range cases {
		got, err := VerifyHash(&coreplugin.IaCPlugin{ID: "1", Extension: c.ext}, "linux_amd64", "aaa")
		if err != nil || got != c.want {
			t.Errorf("%s: got %v err=%v, want %v", c.name, got, err, c.want)
		}
	}

	if _, err := VerifyHash(&coreplugin.IaCPlugin{ID: "1", Extension: json.RawMessage(`not json`)}, "linux_amd64", "aaa"); err == nil {
		t.Error("invalid extension want err")
	}
	if _, err := VerifyHash(nil, "linux_amd64", "aaa"); err == nil {
		t.Error("nil record want err")
	}
}

func TestWithArchHash(t *testing.T) {
	// 补写本架构时：其它架构的值与 extension 其它字段都必须原样保留。
	in := json.RawMessage(`{"architectures":{"linux_arm64":"ccc"},"other":{"k":1}}`)
	out, err := WithArchHash(in, "linux_amd64", "aaa")
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Architectures map[string]string      `json:"architectures"`
		Other         map[string]interface{} `json:"other"`
	}
	if err = json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got.Architectures["linux_amd64"] != "aaa" || got.Architectures["linux_arm64"] != "ccc" {
		t.Fatalf("architectures = %v", got.Architectures)
	}
	if got.Other["k"] != float64(1) {
		t.Fatalf("other fields must be preserved, got %v", got.Other)
	}

	// 从空 extension 开始。
	for _, empty := range []json.RawMessage{nil, json.RawMessage(`{}`), json.RawMessage(`{"architectures":null}`)} {
		out, err = WithArchHash(empty, "linux_amd64", "aaa")
		if err != nil {
			t.Fatalf("from %s: %v", empty, err)
		}
		res, err := VerifyHash(&coreplugin.IaCPlugin{ID: "1", Extension: out}, "linux_amd64", "aaa")
		if err != nil || res != IaCPluginHashResultMatch {
			t.Fatalf("from %s: round trip got %v err=%v", empty, res, err)
		}
	}

	if _, err = WithArchHash(json.RawMessage(`[1]`), "linux_amd64", "aaa"); err == nil {
		t.Error("non-object extension want err")
	}
}
