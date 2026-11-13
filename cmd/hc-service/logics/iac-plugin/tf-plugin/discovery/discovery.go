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

// Package discovery 扫描 hc-service 本地的 tf_provider 插件目录。
//
// 目录约定：{根目录}/{插件名}/{版本}/{os}_{arch}/{唯一的可执行文件}。
// 每个副本只读取自己架构的目录，扫描结果在启动时确定，之后只读；运行期间替换目录不会被重新感知。
package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"hcm/pkg/iacplugin"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// Entry 一个插件在本副本上的目录检查结果，创建后只读。
type Entry struct {
	// PluginName 插件名，取自目录名。
	PluginName string
	// PluginVersion 插件版本，取自目录名，已通过 semver 三段式校验。
	PluginVersion string
	// Arch 本副本架构，格式 {os}_{arch}。
	Arch string
	// ExecPath 本架构目录下唯一的可执行文件路径。
	ExecPath string
	// SHA256 ExecPath 对应文件的 sha256，十六进制小写。
	SHA256 string
}

// LocalArch 返回本副本的架构标识，格式 {os}_{arch}，与插件目录中的架构目录名一致。
func LocalArch() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

// Scan 扫描插件根目录，为每个插件产出一条条目（取 semver 最高的版本），以插件名为 key 返回。
//
// 根目录不存在，或其中没有插件目录时返回错误。调用方应当让进程退出，
// 而不是带着一个缺失或残缺的插件目录继续提供服务。
// 版本目录名非法、本架构目录缺失、可执行文件缺失或不唯一同样返回错误。
func Scan(root string) (map[string]Entry, error) {
	pluginNames, err := subDirNames(root)
	if err != nil {
		return nil, fmt.Errorf("read plugin root dir %s failed, err: %w", root, err)
	}
	if len(pluginNames) == 0 {
		return nil, fmt.Errorf("plugin root dir %s has no plugin directory", root)
	}

	entries := make(map[string]Entry, len(pluginNames))
	for _, name := range pluginNames {
		entry, err := scanPlugin(filepath.Join(root, name), name, LocalArch())
		if err != nil {
			return nil, err
		}
		if entry == nil {
			continue
		}
		entries[name] = cvt.PtrToVal(entry)
	}

	return entries, nil
}

// subDirNames 返回目录下的子目录名。普通文件不属于插件或版本，直接忽略。
func subDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// scanPlugin 扫描单个插件目录，返回最高版本的条目。没有任何版本目录时返回 nil。
func scanPlugin(pluginPath, name, arch string) (*Entry, error) {
	versions, err := subDirNames(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("read plugin dir %s failed, err: %w", pluginPath, err)
	}
	for _, version := range versions {
		if _, err = iacplugin.ParseVersion(version); err != nil {
			return nil, fmt.Errorf("plugin %s has an invalid version directory %q: %w", name, version, err)
		}
	}
	if len(versions) == 0 {
		return nil, nil
	}

	highest, err := highestPluginVersion(versions)
	if err != nil {
		return nil, err
	}
	for _, version := range versions {
		if version != highest {
			logs.Infof("plugin %s ignores lower version %s, using %s", name, version, highest)
		}
	}

	execPath, err := findExecutable(filepath.Join(pluginPath, highest, arch))
	if err != nil {
		return nil, fmt.Errorf("plugin %s version %s: %w", name, highest, err)
	}
	sum, err := fileSHA256(execPath)
	if err != nil {
		return nil, fmt.Errorf("plugin %s version %s: %w", name, highest, err)
	}

	return &Entry{
		PluginName:    name,
		PluginVersion: highest,
		Arch:          arch,
		ExecPath:      execPath,
		SHA256:        sum,
	}, nil
}

// highestPluginVersion 返回 versions 中 semver 最大的一个。versions 不能为空，且都已通过 ParseVersion。
func highestPluginVersion(versions []string) (string, error) {
	highest := versions[0]
	for _, version := range versions[1:] {
		cmp, err := iacplugin.CompareVersion(version, highest)
		if err != nil {
			return "", err
		}
		if cmp > 0 {
			highest = version
		}
	}
	return highest, nil
}

// findExecutable 返回架构目录下唯一的文件。不要求文件名符合模板；缺失或不唯一都是错误。
func findExecutable(archDir string) (string, error) {
	entries, err := os.ReadDir(archDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("architecture directory %s does not exist", archDir)
		}
		return "", fmt.Errorf("read architecture directory %s failed, err: %w", archDir, err)
	}

	if len(entries) != 1 {
		return "", fmt.Errorf("architecture directory %s should contain exactly one file, found %d",
			archDir, len(entries))
	}
	if entries[0].IsDir() {
		return "", fmt.Errorf("architecture directory %s should only contain one file, found directory %s",
			archDir, entries[0].Name())
	}

	return filepath.Join(archDir, entries[0].Name()), nil
}

// fileSHA256 流式计算文件的 sha256，不会把整个文件读进内存（插件可达数百 MB）。
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s failed, err: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read %s failed, err: %w", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
