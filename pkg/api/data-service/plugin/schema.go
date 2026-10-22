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

// Package dataplugin 定义 IaC 插件与插件 Schema 的 data-service API 请求/响应结构。
package dataplugin

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// IacPluginSchemaBatchCreateReq 批量创建 iac_plugin_schema 请求。
type IacPluginSchemaBatchCreateReq struct {
	Items []IacPluginSchemaCreateField `json:"items" validate:"required,min=1"`
}

// Validate 校验 IacPluginSchemaBatchCreateReq。
func (req *IacPluginSchemaBatchCreateReq) Validate() error {
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

// IacPluginSchemaCreateField 单条 iac_plugin_schema 创建字段。
type IacPluginSchemaCreateField struct {
	PluginName    string          `json:"plugin_name" validate:"required,lte=64"`
	PluginVersion string          `json:"plugin_version" validate:"required,lte=64"`
	SchemaHash    string          `json:"schema_hash" validate:"required,lte=80"`
	SchemaDetail  types.JsonField `json:"schema_detail" validate:"required"`
}

// Validate 校验 IacPluginSchemaCreateField。
func (req *IacPluginSchemaCreateField) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := coreplugin.ValidateSchemaHashFormat(req.SchemaHash); err != nil {
		return err
	}

	return nil
}

// -------------------------- Update --------------------------

// IacPluginSchemaBatchUpdateReq 批量更新 iac_plugin_schema 请求。
type IacPluginSchemaBatchUpdateReq struct {
	Items []IacPluginSchemaUpdateField `json:"items" validate:"required,min=1"`
}

// Validate 校验 IacPluginSchemaBatchUpdateReq。
func (req *IacPluginSchemaBatchUpdateReq) Validate() error {
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

// IacPluginSchemaUpdateField 单条 iac_plugin_schema 更新字段。
type IacPluginSchemaUpdateField struct {
	ID           string          `json:"id" validate:"required,lte=64"`
	SchemaHash   string          `json:"schema_hash" validate:"omitempty,lte=80"`
	SchemaDetail types.JsonField `json:"schema_detail"`
}

// Validate 校验 IacPluginSchemaUpdateField。
func (req *IacPluginSchemaUpdateField) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.SchemaHash) == 0 && len(req.SchemaDetail) == 0 {
		return errors.New("at least one of schema_hash, schema_detail must be set")
	}

	if len(req.SchemaHash) != 0 {
		if err := coreplugin.ValidateSchemaHashFormat(req.SchemaHash); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Delete --------------------------

// IacPluginSchemaBatchDeleteReq 批量删除 iac_plugin_schema 请求。
type IacPluginSchemaBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate 校验 IacPluginSchemaBatchDeleteReq。
func (req *IacPluginSchemaBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// IacPluginSchemaListReq 列表查询 iac_plugin_schema 请求。
type IacPluginSchemaListReq = core.ListReq

// IacPluginSchemaListResult iac_plugin_schema 列表结果数据。
type IacPluginSchemaListResult = core.ListResultT[tableplugin.IacPluginSchemaTable]

// IacPluginSchemaListResp 列表查询 iac_plugin_schema 响应。
type IacPluginSchemaListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *IacPluginSchemaListResult `json:"data"`
}
