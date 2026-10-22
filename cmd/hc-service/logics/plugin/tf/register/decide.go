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

// Package register 实现 tf_provider 插件的启动注册与升级。
package register

import (
	"encoding/json"
	"fmt"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/iacplugin"
	cvt "hcm/pkg/tools/converter"
)

// Decision 注册判定的完整结果。
type Decision struct {
	Action IaCPluginAction
	// Reason 判定原因，用于日志与退出信息。
	Reason string
	// Record 该版本已有的记录。插入新记录和版本回退退出时为 nil。
	Record *coreplugin.IaCPlugin
}

// Decide 根据该插件名下的全部注册记录与本副本的目录版本，判定下一步动作。
//
// 它是纯函数，不做任何 I/O：records 必须是持插件名锁后读到的最新记录，version 是目录检查得到的版本。
// 判定顺序：
//  1. 记录里存在版本高于 version 的已完成记录（enabled / disabled）→ 版本回退，退出；
//  2. 看 version 对应的记录：不存在、或处于 init / failed → 需要注册（新建或重置后重入）；
//     enabled / disabled → 已注册，只校验 sha256。
//
// 动作只表达"要做什么"，"新建还是重置"由注册执行根据 Record 是否为空决定。
//
// "已完成最高版本" M 只统计 enabled / disabled：init / failed 的高版本不能让旧副本误判为版本回退
func Decide(records []coreplugin.IaCPlugin, version string) (*Decision, error) {
	if _, err := iacplugin.ParseVersion(version); err != nil {
		return nil, err
	}

	// 一次遍历做两件事：
	//  1. 插件名在注册表里唯一，同名的记录必须都是 tf_provider。若该名字已被 http 插件占用，
	//     继续比对会把 tf 的 sha256 写进 http 记录的 extension，必须先失败退出；
	//  2. 只收 enabled / disabled，init / failed 不参与「已完成最高版本」比较。
	filteredRecords := make([]coreplugin.IaCPlugin, 0, len(records))
	for _, rec := range records {
		if rec.Protocol != enumor.IaCPluginProtocolTFProvider {
			return nil, fmt.Errorf("plugin %s is registered with protocol %s, but the local directory is a %s",
				rec.PluginName, rec.Protocol, enumor.IaCPluginProtocolTFProvider)
		}
		if rec.State == enumor.IaCPluginStateEnabled || rec.State == enumor.IaCPluginStateDisabled {
			filteredRecords = append(filteredRecords, rec)
		}
	}

	highest, err := iacplugin.Latest(filteredRecords)
	if err != nil {
		return nil, err
	}
	if highest != nil {
		cmp, err := iacplugin.CompareVersion(version, highest.PluginVersion)
		if err != nil {
			return nil, err
		}
		if cmp < 0 {
			return &Decision{
				Action: IaCPluginActionExit,
				Reason: fmt.Sprintf("local plugin version %s is lower than the highest registered version %s",
					version, highest.PluginVersion),
			}, nil
		}
	}

	match, err := matchPluginVersionRecord(records, version)
	if err != nil {
		return nil, err
	}
	decision := &Decision{Record: match}
	switch {
	case match == nil:
		// 没有记录：新建后注册。
		decision.Action = IaCPluginActionRegister
		decision.Reason = fmt.Sprintf("plugin version %s is not registered", version)
	case match.State == enumor.IaCPluginStateInit:
		// 上次注册中途崩溃：重置后重入，与新建走同一套流程。
		decision.Action = IaCPluginActionRegister
		decision.Reason = "previous registration was interrupted"
	case match.State == enumor.IaCPluginStateFailed:
		decision.Action = IaCPluginActionRegister
		decision.Reason = "previous registration failed"
	case match.State == enumor.IaCPluginStateEnabled || match.State == enumor.IaCPluginStateDisabled:
		decision.Action = IaCPluginActionVerify
		decision.Reason = fmt.Sprintf("plugin version %s is already registered as %s", version, match.State)
	default:
		return nil, fmt.Errorf("plugin record %s has an unknown state %q", match.ID, match.State)
	}

	return decision, nil
}

// matchPluginVersionRecord 返回版本号等于 version 的记录。没有这条记录时返回 nil。
func matchPluginVersionRecord(records []coreplugin.IaCPlugin, version string) (*coreplugin.IaCPlugin, error) {
	for i := range records {
		cmp, err := iacplugin.CompareVersion(records[i].PluginVersion, version)
		if err != nil {
			return nil, fmt.Errorf("plugin record %s has an invalid version: %w", records[i].ID, err)
		}
		if cmp == 0 {
			return cvt.ValToPtr(records[i]), nil
		}
	}
	return nil, nil
}

// VerifyHash 比对记录里本架构的 sha256 与本地计算值。
//
// 只比较 arch 对应的那一项：每个副本只计算自己架构的 hash，其它架构的值由对应架构的副本负责。
func VerifyHash(record *coreplugin.IaCPlugin, arch, localSHA256 string) (IaCPluginHashResult, error) {
	if record == nil {
		return "", fmt.Errorf("plugin record is nil")
	}

	var ext coreplugin.TFProviderExtension
	if len(record.Extension) != 0 {
		if err := json.Unmarshal(record.Extension, &ext); err != nil {
			return "", fmt.Errorf("plugin record %s has an invalid extension: %w", record.ID, err)
		}
	}

	recorded, ok := ext.Architectures[arch]
	if !ok || recorded == "" {
		return IaCPluginHashResultMissing, nil
	}
	if recorded == localSHA256 {
		return IaCPluginHashResultMatch, nil
	}

	return IaCPluginHashResultMismatch, nil
}

// WithArchHash 返回在 extension 基础上补写（或设置）arch 对应 sha256 后的新 extension。
//
// 其它架构已有的值与 extension 中的其它字段原样保留，不会被覆盖。
func WithArchHash(extension json.RawMessage, arch, sha256 string) (json.RawMessage, error) {
	all := make(map[string]json.RawMessage)
	if len(extension) != 0 {
		if err := json.Unmarshal(extension, &all); err != nil {
			return nil, fmt.Errorf("invalid extension: %w", err)
		}
	}

	archs := make(map[string]string)
	if raw, ok := all["architectures"]; ok && len(raw) != 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &archs); err != nil {
			return nil, fmt.Errorf("invalid extension.architectures: %w", err)
		}
	}
	archs[arch] = sha256

	encoded, err := json.Marshal(archs)
	if err != nil {
		return nil, err
	}
	all["architectures"] = encoded

	return json.Marshal(all)
}
