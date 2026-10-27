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

package cloudserver

import (
	"fmt"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// IaCPluginListResult IaC 插件注册记录列表结果，列表请求复用 ListReq。
type IaCPluginListResult = core.ListResultT[coreplugin.IaCPlugin]

// IaCPluginSwitchReq IaC 插件启用 / 停用请求。
type IaCPluginSwitchReq struct {
	PluginName string                `json:"plugin_name" validate:"required,max=64"`
	State      enumor.IaCPluginState `json:"state" validate:"required"`
}

// Validate IaCPluginSwitchReq.
func (req *IaCPluginSwitchReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	switch req.State {
	case enumor.IaCPluginStateEnabled, enumor.IaCPluginStateDisabled:
		return nil
	default:
		return fmt.Errorf("unsupported switch target state: %s, only enabled or disabled is allowed", req.State)
	}
}
