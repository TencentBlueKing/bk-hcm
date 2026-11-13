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

// put 在 root 下按约定创建 {name}/{version}/{arch}/{file}。
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

// otherArch 返回一个与本机不同的架构目录名，用来放不应被读取的文件。
func otherArch() string {
	if LocalArch() == "linux_arm64" {
		return "linux_amd64"
	}
	return "linux_arm64"
}

func TestScanHighestVersionBySemver(t *testing.T) {
	root := t.TempDir()
	// 字符串比较会选 1.9.0，semver 必须选 1.10.0。
	put(t, root, "tc", "1.9.0", LocalArch(), "bin", "old")
	put(t, root, "tc", "1.10.0", LocalArch(), "bin", "new")

	entries, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries["tc"].PluginVersion != "1.10.0" {
		t.Fatalf("want 1.10.0, got %+v", entries)
	}
	if entries["tc"].SHA256 != sum("new") {
		t.Fatalf("sha256 should be computed from the chosen version's file")
	}
}

func TestScanOnlyLocalArch(t *testing.T) {
	root := t.TempDir()
	put(t, root, "tc", "1.0.0", LocalArch(), "bin", "amd64-content")
	put(t, root, "tc", "1.0.0", otherArch(), "bin", "arm64-content")

	entries, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries["tc"].SHA256 != sum("amd64-content") || entries["tc"].Arch != LocalArch() {
		t.Fatalf("must only read the local arch, got %+v", entries)
	}
}

func TestScanArbitraryFileName(t *testing.T) {
	root := t.TempDir()
	put(t, root, "tc", "1.0.0", LocalArch(), "whatever-name_v9", "x")

	entries, err := Scan(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("err=%v entries=%+v", err, entries)
	}
	if filepath.Base(entries["tc"].ExecPath) != "whatever-name_v9" {
		t.Fatalf("exec path = %s", entries["tc"].ExecPath)
	}
}

func TestScanMultiplePlugins(t *testing.T) {
	root := t.TempDir()
	put(t, root, "tc", "1.0.0", LocalArch(), "bin", "a")
	put(t, root, "aws", "2.3.4", LocalArch(), "bin", "b")

	entries, err := Scan(root)
	if err != nil || len(entries) != 2 {
		t.Fatalf("err=%v entries=%+v", err, entries)
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
	}

	for _, c := range cases {
		root := t.TempDir()
		c.build(root)
		if _, err := Scan(root); err == nil {
			t.Errorf("%s: want error, got nil", c.name)
		}
	}
}

func TestScanEmptyOrMissingRoot(t *testing.T) {
	if _, err := Scan(filepath.Join(t.TempDir(), "not-exist")); err == nil {
		t.Fatal("missing root: want error, got nil")
	}

	if _, err := Scan(t.TempDir()); err == nil {
		t.Fatal("empty root: want error, got nil")
	}

	withFile := t.TempDir()
	if err := os.WriteFile(filepath.Join(withFile, "README"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Scan(withFile); err == nil {
		t.Fatal("root with only a file: want error, got nil")
	}
}

func TestScanPluginDirWithoutVersions(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "tc"), 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := Scan(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("plugin dir without versions should be skipped, err=%v entries=%v", err, entries)
	}
}
