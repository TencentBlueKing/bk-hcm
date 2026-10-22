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

// Package iacplugin IaC 插件注册与升级的公共规则：状态转换、版本比较。
package iacplugin

import (
	"fmt"
	"strings"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"

	"golang.org/x/mod/semver"
)

// allowedTransitions 合法的状态转换表（不含"状态不变"）。
var allowedTransitions = map[enumor.IaCPluginState]map[enumor.IaCPluginState]struct{}{
	enumor.IaCPluginStateInit: {
		enumor.IaCPluginStateEnabled: {},
		enumor.IaCPluginStateFailed:  {},
	},
	enumor.IaCPluginStateEnabled: {
		enumor.IaCPluginStateFailed:   {},
		enumor.IaCPluginStateDisabled: {},
	},
	enumor.IaCPluginStateFailed: {
		enumor.IaCPluginStateInit:     {},
		enumor.IaCPluginStateDisabled: {},
	},
	enumor.IaCPluginStateDisabled: {
		enumor.IaCPluginStateEnabled: {},
	},
}

// CanTransit 判断状态转换是否合法。相同状态视为无变化，返回 true。
func CanTransit(from, to enumor.IaCPluginState) bool {
	if from == to {
		return true
	}

	_, ok := allowedTransitions[from][to]
	return ok
}

// ParseVersion 校验并规范化插件版本号：必须是三段式 主.次.修订，可带预发布后缀和 v 前缀。
// 返回带 v 前缀的规范形式，供 semver 包比较使用。
func ParseVersion(version string) (string, error) {
	v := version
	if len(v) > 0 && v[0] != 'v' {
		v = "v" + v
	}

	if !semver.IsValid(v) {
		return "", fmt.Errorf("invalid plugin version: %q", version)
	}

	// semver.IsValid 接受 v1、v1.2 这类简写，这里要求必须是完整的三段式。
	if !hasThreeSegments(v) {
		return "", fmt.Errorf("invalid plugin version: %q, expected major.minor.patch", version)
	}

	return v, nil
}

// hasThreeSegments 判断 v 的主体部分（去掉预发布和构建信息）是否恰好有三段。
// semver.IsValid 已经通过，这里只数主.次.修订。
func hasThreeSegments(v string) bool {
	body, _, _ := strings.Cut(v[1:], "-")
	body, _, _ = strings.Cut(body, "+")
	return len(strings.Split(body, ".")) == 3
}

// CompareVersion 按 semver 比较两个版本号，a<b 返回 -1，相等返回 0，a>b 返回 1。任一版本不合法返回错误。
func CompareVersion(a, b string) (int, error) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}

	vb, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}

	return semver.Compare(va, vb), nil
}

// Latest 从同名插件的记录中选出版本最高的一条；plugins 为空返回 nil。
// 记录里存在不合法版本号时返回错误，调用方不应在不确定的版本上继续。
func Latest(plugins []coreplugin.IaCPlugin) (*coreplugin.IaCPlugin, error) {
	var latest *coreplugin.IaCPlugin
	var latestVersion string
	for i := range plugins {
		version, err := ParseVersion(plugins[i].PluginVersion)
		if err != nil {
			return nil, err
		}

		if latest == nil || semver.Compare(version, latestVersion) > 0 {
			latest = cvt.ValToPtr(plugins[i])
			latestVersion = version
		}
	}

	return latest, nil
}

// InheritState 计算 version 同步成功后应取的状态：上一版本（版本低于 version 的记录中最高的一条）为 disabled
// 则继承 disabled，否则为 enabled；没有上一版本（首次注册）时为 enabled。tf_provider 与 http 插件共用。
// 版本号不合法的记录不参与比较：调用方在判定阶段已校验过版本号，这里保守按不继承处理。
func InheritState(records []coreplugin.IaCPlugin, version string) enumor.IaCPluginState {
	lower := make([]coreplugin.IaCPlugin, 0, len(records))
	for _, rec := range records {
		cmp, err := CompareVersion(rec.PluginVersion, version)
		if err != nil {
			continue
		}
		if cmp < 0 {
			lower = append(lower, rec)
		}
	}

	previous, err := Latest(lower)
	if err == nil && previous != nil && previous.State == enumor.IaCPluginStateDisabled {
		return enumor.IaCPluginStateDisabled
	}

	return enumor.IaCPluginStateEnabled
}
