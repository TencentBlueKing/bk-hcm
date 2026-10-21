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

// Package plugin 定义 IaC 插件与插件 Schema 的核心数据模型及校验规则。
package plugin

import (
	"hcm/pkg/criteria/enumor"

	tfjson "github.com/hashicorp/terraform-json"
)

const (
	// ProtocolTFProvider Terraform Provider 插件协议。
	ProtocolTFProvider = enumor.IaCPluginProtocolTFProvider
	// ProtocolHTTP HTTP 插件协议。
	ProtocolHTTP = enumor.IaCPluginProtocolHTTP
)

// PluginInfo 表示 Schema 中的插件身份与版本标识。
type PluginInfo struct {
	// Name 全局唯一插件名称，同时作为所有 type_name 的前缀。
	Name string `json:"name"`
	// Version 插件版本号，与 /health 返回的 plugin_version 一致。
	Version string `json:"version"`
}

// Capabilities 表示插件声明支持的南向接口能力位。
type Capabilities struct {
	// Apply 表示是否实现了 POST /apply。
	Apply bool `json:"apply,omitempty"`
	// Plan 表示是否实现了 POST /plan。
	Plan bool `json:"plan,omitempty"`
	// Refresh 表示是否实现了 POST /refresh。
	Refresh bool `json:"refresh,omitempty"`
	// ReadData 表示是否实现了 POST /read_data。
	ReadData bool `json:"read_data,omitempty"`
	// AsyncApply 表示是否支持异步 apply（status=running + POST /result）。
	AsyncApply bool `json:"async_apply,omitempty"`
	// Stop 表示是否实现了 POST /stop。
	Stop bool `json:"stop,omitempty"`
	// Validate 表示是否实现了 POST /validate。
	Validate bool `json:"validate,omitempty"`
	// Import 表示是否实现了 POST /import。
	Import bool `json:"import,omitempty"`
	// List 表示是否实现了 POST /list。
	List bool `json:"list,omitempty"`
}

// SchemaDetail 持久化到 iac_plugin_schema.schema_detail JSON 列的 Schema 详情载荷。
// 内嵌 tfjson.ProviderSchema 以保持与 Terraform provider schema JSON 格式一致。
type SchemaDetail struct {
	tfjson.ProviderSchema `json:",inline"`

	// ProtocolVersion 南向载荷格式版本号。
	ProtocolVersion string `json:"protocol_version,omitempty"`
	// Capabilities 声明已实现的南向接口能力。
	Capabilities Capabilities `json:"capabilities"`
}

// PluginSchema GetSchema / 南向协议返回的完整插件 Schema。
type PluginSchema struct {
	// Plugin 包含插件名称与版本。
	Plugin PluginInfo `json:"plugin"`
	// SchemaHash 整个 Schema 的内容指纹。
	SchemaHash string `json:"schema_hash"`
	// SchemaDetail 内联展开的 Schema 详情载荷。
	SchemaDetail `json:",inline"`
}
