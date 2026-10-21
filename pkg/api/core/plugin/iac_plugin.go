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

// Package plugin defines the core types of an IaC plugin record.
package plugin

import (
	"encoding/json"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// IaCPlugin IaC 插件注册记录，每个插件的每个版本一条。平台级，不区分租户。
type IaCPlugin struct {
	ID             string                   `json:"id"`
	PluginName     string                   `json:"plugin_name"`
	PluginVersion  string                   `json:"plugin_version"`
	Protocol       enumor.IaCPluginProtocol `json:"protocol"`
	State          enumor.IaCPluginState    `json:"state"`
	Message        string                   `json:"message"`
	Extension      json.RawMessage          `json:"extension"`
	Memo           *string                  `json:"memo"`
	*core.Revision `json:",inline"`
}
