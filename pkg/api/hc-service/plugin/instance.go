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

// Package plugin 定义 hc-service 插件实例相关的 HTTP 请求与响应结构。
package plugin

import (
	"errors"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/validator"
)

// PluginInstanceCreateReq hc-service 创建插件实例请求。
type PluginInstanceCreateReq struct {
	PluginName string         `json:"plugin_name" validate:"required,max=128"`
	Name       string         `json:"name" validate:"required,max=128"`
	Properties map[string]any `json:"properties" validate:"omitempty"`
	Extension  map[string]any `json:"extension" validate:"omitempty"`
	Memo       *string        `json:"memo" validate:"omitempty,max=255"`
}

// Validate 校验 PluginInstanceCreateReq。
func (req *PluginInstanceCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// PluginInstanceUpdateReq hc-service 更新插件实例请求。
// 当 Properties 非 nil 时，按全量覆盖语义重新基于活跃版本 Schema 校验并更新 attrs 与 secret_attrs；
// 当 Properties 为 nil 时，仅局部更新元数据字段（name / extension / memo）。
type PluginInstanceUpdateReq struct {
	Name       string         `json:"name,omitempty" validate:"omitempty,max=128"`
	Properties map[string]any `json:"properties,omitempty" validate:"omitempty"`
	Extension  map[string]any `json:"extension,omitempty" validate:"omitempty"`
	Memo       *string        `json:"memo,omitempty" validate:"omitempty,max=255"`
}

// Validate 校验 PluginInstanceUpdateReq。
func (req *PluginInstanceUpdateReq) Validate() error {
	if len(req.Name) == 0 && req.Properties == nil && req.Extension == nil && req.Memo == nil {
		return errors.New("at least one of name, properties, extension, memo must be set")
	}

	return validator.Validate.Struct(req)
}

// PluginInstanceListReq hc-service 列表查询插件实例请求。
type PluginInstanceListReq = core.ListReq

// PluginInstanceListResult hc-service 列表查询插件实例结果。
type PluginInstanceListResult = core.ListResultT[coreplugin.PluginInstance]

// PluginInstanceListResp hc-service 列表查询插件实例响应。
type PluginInstanceListResp = core.BaseResp[PluginInstanceListResult]
