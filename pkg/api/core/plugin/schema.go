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

const (
	// CapabilityApply 插件支持 apply。
	CapabilityApply = "apply"
	// CapabilityPlan 插件支持 plan。
	CapabilityPlan = "plan"
	// CapabilityRefresh 插件支持 refresh。
	CapabilityRefresh = "refresh"
	// CapabilityReadData 插件支持 read_data。
	CapabilityReadData = "read_data"
	// CapabilityAsyncApply 插件支持 async_apply。
	CapabilityAsyncApply = "async_apply"
	// CapabilityStop 插件支持 stop。
	CapabilityStop = "stop"
	// CapabilityValidate 插件支持 validate。
	CapabilityValidate = "validate"
	// CapabilityImport 插件支持 import。
	CapabilityImport = "import"
	// CapabilityList 插件支持 list。
	CapabilityList = "list"
)

// PluginSchema GetSchema / 南向协议返回的完整插件 Schema。
type PluginSchema struct {
	PluginName    string `json:"plugin_name"`
	PluginVersion string `json:"plugin_version"`
	SchemaHash    string `json:"schema_hash"`
	RawSchema     `json:",inline"`
}

// RawSchema 持久化到 iac_plugin_schema.schema_detail JSON 列的内层 Schema 载荷。
type RawSchema struct {
	ProtocolVersion         string                            `json:"protocol_version,omitempty"`
	Capabilities            map[string]bool                   `json:"capabilities"`
	Provider                *tfjson.Schema                    `json:"provider"`
	ResourceSchemas         map[string]*tfjson.Schema         `json:"resource_schemas,omitempty"`
	DataSourceSchemas       map[string]*tfjson.Schema         `json:"data_source_schemas,omitempty"`
	ActionSchemas           map[string]*tfjson.Schema         `json:"action_schemas,omitempty"`
	ResourceIdentitySchemas map[string]*tfjson.IdentitySchema `json:"resource_identity_schemas,omitempty"`
}
