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

// Package plugin 定义插件相关的核心模型。
package plugin

import (
	"hcm/pkg/api/core"
)

// PluginInstance 插件实例对外核心模型（hc-service 返回完整明文 properties，不暴露 attrs/secret_attrs 内部拆分形态）。
type PluginInstance struct {
	ID             string         `json:"id"`
	PluginName     string         `json:"plugin_name"`
	Name           string         `json:"name"`
	Properties     map[string]any `json:"properties"`
	Extension      map[string]any `json:"extension"`
	Memo           *string        `json:"memo"`
	TenantID       string         `json:"tenant_id,omitempty"`
	*core.Revision `json:",inline"`
}

// IacPluginInstance data-service 层插件实例模型（包含同构拆分的 attrs 与叶子已解密的 secret_attrs）。
type IacPluginInstance struct {
	ID             string         `json:"id"`
	PluginName     string         `json:"plugin_name"`
	Name           string         `json:"name"`
	Attrs          map[string]any `json:"attrs"`
	SecretAttrs    map[string]any `json:"secret_attrs"`
	Extension      map[string]any `json:"extension"`
	Memo           *string        `json:"memo"`
	TenantID       string         `json:"tenant_id,omitempty"`
	*core.Revision `json:",inline"`
}
