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

package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, root, name, version, arch, file, content string) {
	t.Helper()
	dir := filepath.Join(root, name, version, arch)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func otherArch() string {
	if LocalArch() == "linux_arm64" {
		return "linux_amd64"
	}
	return "linux_arm64"
}

func TestScan(t *testing.T) {
	root := t.TempDir()
	// semver 选 v1.10.0（去 v 前缀入库）、忽略其它架构、支持任意可执行文件名、多插件并存、跳过无版本目录。
	put(t, root, "tc", "1.9.0", LocalArch(), "bin", "old")
	put(t, root, "tc", "v1.10.0", LocalArch(), "whatever-name_v9", "new")
	put(t, root, "tc", "v1.10.0", otherArch(), "bin", "other-arch")
	put(t, root, "aws", "2.3.4", LocalArch(), "bin", "aws-bin")
	if err := os.MkdirAll(filepath.Join(root, "empty-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}

	entries, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 plugins, got %+v", entries)
	}

	tc := entries["tc"]
	wantExec := filepath.Join(root, "tc", "v1.10.0", LocalArch(), "whatever-name_v9")
	if tc.PluginVersion != "1.10.0" || tc.Arch != LocalArch() || tc.ExecPath != wantExec || tc.SHA256 != sum("new") {
		t.Fatalf("unexpected tc entry: %+v", tc)
	}
	if entries["aws"].PluginVersion != "2.3.4" || entries["aws"].SHA256 != sum("aws-bin") {
		t.Fatalf("unexpected aws entry: %+v", entries["aws"])
	}
}

func TestScanInvalid(t *testing.T) {
	cases := []struct {
		name  string
		build func(root string)
	}{
		{"illegal version dir", func(r string) { put(t, r, "tc", "latest", LocalArch(), "bin", "x") }},
		{"two-segment version dir", func(r string) { put(t, r, "tc", "1.2", LocalArch(), "bin", "x") }},
		{"local arch dir missing", func(r string) { put(t, r, "tc", "1.0.0", otherArch(), "bin", "x") }},
		{"no file in arch dir", func(r string) {
			if err := os.MkdirAll(filepath.Join(r, "tc", "1.0.0", LocalArch()), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"more than one file", func(r string) {
			put(t, r, "tc", "1.0.0", LocalArch(), "a", "x")
			put(t, r, "tc", "1.0.0", LocalArch(), "b", "y")
		}},
		{"subdirectory in arch dir", func(r string) {
			if err := os.MkdirAll(filepath.Join(r, "tc", "1.0.0", LocalArch(), "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"empty root", func(r string) {}},
		{"root with only a file", func(r string) {
			if err := os.WriteFile(filepath.Join(r, "README"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	}

	for _, c := range cases {
		root := t.TempDir()
		c.build(root)
		if _, err := Scan(root); err == nil {
			t.Errorf("%s: want error, got nil", c.name)
		}
	}

	if _, err := Scan(filepath.Join(t.TempDir(), "not-exist")); err == nil {
		t.Error("missing root: want error, got nil")
	}
}
