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

// Package dataplugin 定义 data-service 插件相关接口的请求与响应结构。
package dataplugin

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// IacPluginInstanceBatchCreateReq 批量创建 iac_plugin_instance 请求。
type IacPluginInstanceBatchCreateReq struct {
	Items []IacPluginInstanceCreateField `json:"items" validate:"required,min=1"`
}

// Validate 校验 IacPluginInstanceBatchCreateReq。
func (req *IacPluginInstanceBatchCreateReq) Validate() error {
	if len(req.Items) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("items count should <= %d", constant.BatchOperationMaxLimit)
	}

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for i := range req.Items {
		if err := req.Items[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IacPluginInstanceCreateField 单条 iac_plugin_instance 创建字段。
type IacPluginInstanceCreateField struct {
	PluginName  string         `json:"plugin_name" validate:"required,lte=64"`
	Name        string         `json:"name" validate:"required,lte=128"`
	Attrs       map[string]any `json:"attrs"`
	SecretAttrs map[string]any `json:"secret_attrs"`
	Extension   map[string]any `json:"extension"`
	Memo        *string        `json:"memo"`
}

// Validate 校验 IacPluginInstanceCreateField。
func (req *IacPluginInstanceCreateField) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := validator.ValidateMemo(req.Memo, false); err != nil {
		return err
	}

	return nil
}

// -------------------------- Update --------------------------

// IacPluginInstanceBatchUpdateReq 批量更新 iac_plugin_instance 请求。
type IacPluginInstanceBatchUpdateReq struct {
	Items []IacPluginInstanceUpdateField `json:"items" validate:"required,min=1"`
}

// Validate 校验 IacPluginInstanceBatchUpdateReq。
func (req *IacPluginInstanceBatchUpdateReq) Validate() error {
	if len(req.Items) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("items count should <= %d", constant.BatchOperationMaxLimit)
	}

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for i := range req.Items {
		if err := req.Items[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IacPluginInstanceUpdateField 单条 iac_plugin_instance 更新字段。
type IacPluginInstanceUpdateField struct {
	ID          string         `json:"id" validate:"required,lte=64"`
	Name        string         `json:"name" validate:"omitempty,lte=128"`
	Attrs       map[string]any `json:"attrs"`
	SecretAttrs map[string]any `json:"secret_attrs"`
	Extension   map[string]any `json:"extension"`
	Memo        *string        `json:"memo"`
}

// Validate 校验 IacPluginInstanceUpdateField。
func (req *IacPluginInstanceUpdateField) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.Name == "" && req.Attrs == nil && req.SecretAttrs == nil && req.Extension == nil && req.Memo == nil {
		return errors.New("at least one of name, attrs, secret_attrs, extension, memo must be set")
	}

	if err := validator.ValidateMemo(req.Memo, false); err != nil {
		return err
	}

	return nil
}

// -------------------------- Delete --------------------------

// IacPluginInstanceBatchDeleteReq 批量删除 iac_plugin_instance 请求。
type IacPluginInstanceBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate 校验 IacPluginInstanceBatchDeleteReq。
func (req *IacPluginInstanceBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// IacPluginInstanceListReq 列表查询 iac_plugin_instance 请求。
type IacPluginInstanceListReq = core.ListReq

// IacPluginInstanceListResult iac_plugin_instance 列表结果数据。
type IacPluginInstanceListResult = core.ListResultT[coreplugin.IacPluginInstance]

// IacPluginInstanceListResp 列表查询 iac_plugin_instance 响应。
type IacPluginInstanceListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *IacPluginInstanceListResult `json:"data"`
}
