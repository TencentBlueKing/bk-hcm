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
	"testing"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
)

func TestVersion(t *testing.T) {
	valid := []string{"1.83.34", "v1.83.34", "1.10.0-rc.1", "0.0.1"}
	for _, v := range valid {
		if _, err := ParseVersion(v); err != nil {
			t.Errorf("ParseVersion(%q) unexpected err: %v", v, err)
		}
	}

	invalid := []string{"", "latest", "1", "1.2", "v1.2", "1.2.3.4", "a.b.c", "1.2.x"}
	for _, v := range invalid {
		if _, err := ParseVersion(v); err == nil {
			t.Errorf("ParseVersion(%q) want err, got nil", v)
		}
	}

	// 字符串比较会得出 1.9.0 > 1.10.0，semver 必须相反。
	cases := []struct {
		a, b string
		want int
	}{
		{"1.9.0", "1.10.0", -1},
		{"1.10.0", "1.9.0", 1},
		{"1.83.34", "v1.83.34", 0},
		{"1.0.0-rc.1", "1.0.0", -1},
	}
	for _, c := range cases {
		got, err := CompareVersion(c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("CompareVersion(%q, %q) = %d, %v; want %d", c.a, c.b, got, err, c.want)
		}
	}

	if _, err := CompareVersion("1.0.0", "latest"); err == nil {
		t.Error("CompareVersion with invalid version want err")
	}
}

func TestLatest(t *testing.T) {
	mk := func(vs ...string) []coreplugin.IaCPlugin {
		out := make([]coreplugin.IaCPlugin, 0, len(vs))
		for _, v := range vs {
			out = append(out, coreplugin.IaCPlugin{PluginVersion: v})
		}
		return out
	}

	got, err := Latest(mk("1.83.30", "1.83.34", "1.9.0"))
	if err != nil || got == nil || got.PluginVersion != "1.83.34" {
		t.Fatalf("Latest = %+v, %v; want 1.83.34", got, err)
	}

	got, err = Latest(nil)
	if err != nil || got != nil {
		t.Fatalf("Latest(nil) = %+v, %v; want nil, nil", got, err)
	}

	if _, err = Latest(mk("1.0.0", "latest")); err == nil {
		t.Fatal("Latest with an invalid version want err")
	}
}

func TestInheritState(t *testing.T) {
	rec := func(version string, state enumor.IaCPluginState) coreplugin.IaCPlugin {
		return coreplugin.IaCPlugin{PluginName: "p", PluginVersion: version, State: state}
	}
	const (
		initS    = enumor.IaCPluginStateInit
		enabled  = enumor.IaCPluginStateEnabled
		disabled = enumor.IaCPluginStateDisabled
		failedS  = enumor.IaCPluginStateFailed
	)

	cases := []struct {
		name    string
		records []coreplugin.IaCPlugin
		version string
		want    enumor.IaCPluginState
	}{
		{name: "首个版本", records: []coreplugin.IaCPlugin{rec("1.0.0", initS)}, version: "1.0.0", want: enabled},
		{name: "继承启用", records: []coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", initS)},
			version: "1.1.0", want: enabled},
		{name: "继承停用", records: []coreplugin.IaCPlugin{rec("1.0.0", disabled), rec("1.1.0", initS)},
			version: "1.1.0", want: disabled},
		{name: "上一版本失败按启用", records: []coreplugin.IaCPlugin{rec("1.0.0", failedS), rec("1.1.0", initS)},
			version: "1.1.0", want: enabled},
		{name: "只看紧邻的上一版本",
			records: []coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", disabled), rec("1.2.0", initS)},
			version: "1.2.0", want: disabled},
		{name: "更高版本不参与",
			records: []coreplugin.IaCPlugin{rec("1.0.0", enabled), rec("1.1.0", initS), rec("1.2.0", disabled)},
			version: "1.1.0", want: enabled},
		{name: "非法版本记录不参与",
			records: []coreplugin.IaCPlugin{rec("bad", disabled), rec("1.1.0", initS)},
			version: "1.1.0", want: enabled},
	}

	for _, c := range cases {
		if got := InheritState(c.records, c.version); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}
