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

// Package dataplugin defines IaC plugin data-service API request and response structures.
package dataplugin

import (
	"encoding/json"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// IaCPluginMessageMaxLen 是 iac_plugin.message 的列宽，对应建表的 varchar(255)。
const IaCPluginMessageMaxLen = 255

// IaCPluginCreateReq IaC 插件注册记录创建请求。
type IaCPluginCreateReq struct {
	PluginName    string                   `json:"plugin_name" validate:"required,max=64"`
	PluginVersion string                   `json:"plugin_version" validate:"required,max=64"`
	Protocol      enumor.IaCPluginProtocol `json:"protocol" validate:"required"`
	State         enumor.IaCPluginState    `json:"state" validate:"required"`
	Message       string                   `json:"message" validate:"max=255"`
	Extension     json.RawMessage          `json:"extension" validate:"required"`
	Memo          *string                  `json:"memo" validate:"omitempty,max=255"`
}

// Validate IaCPluginCreateReq.
func (req *IaCPluginCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Protocol.Validate(); err != nil {
		return err
	}

	return req.State.Validate()
}

// IaCPluginUpdateReq IaC 插件注册记录更新请求。
// 创建后 plugin_name、plugin_version、protocol 不可修改。
// ExpectState 不为空时，仅当记录当前状态等于它才会更新，否则返回 errf.Aborted。
type IaCPluginUpdateReq struct {
	ID          string                `json:"id" validate:"required"`
	ExpectState enumor.IaCPluginState `json:"expect_state,omitempty" validate:"omitempty"`
	State       enumor.IaCPluginState `json:"state,omitempty" validate:"omitempty"`
	// Message 为 nil 表示不更新；指向空字符串表示清空。
	Message   *string         `json:"message,omitempty" validate:"omitempty,max=255"`
	Extension json.RawMessage `json:"extension,omitempty" validate:"omitempty"`
	Memo      *string         `json:"memo,omitempty" validate:"omitempty,max=255"`
}

// Validate IaCPluginUpdateReq.
func (req *IaCPluginUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.ExpectState) != 0 {
		if err := req.ExpectState.Validate(); err != nil {
			return err
		}
	}

	if len(req.State) != 0 {
		if err := req.State.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IaCPluginBatchUpdateStateReq 按插件名批量修改状态：仅当前状态属于 FromStates 的记录会被更新。
type IaCPluginBatchUpdateStateReq struct {
	PluginName string                  `json:"plugin_name" validate:"required"`
	FromStates []enumor.IaCPluginState `json:"from_states" validate:"required,min=1"`
	ToState    enumor.IaCPluginState   `json:"to_state" validate:"required"`
}

// Validate IaCPluginBatchUpdateStateReq.
func (req *IaCPluginBatchUpdateStateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, one := range req.FromStates {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	return req.ToState.Validate()
}

// IaCPluginListResult IaC 插件注册记录列表结果，列表请求复用 core.ListReq。
type IaCPluginListResult = core.ListResultT[coreplugin.IaCPlugin]
