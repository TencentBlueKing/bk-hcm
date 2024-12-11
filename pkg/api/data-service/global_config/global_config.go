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

// Package datagconf global config data service
package datagconf

import (
	"fmt"

	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	"hcm/pkg/criteria/validator"
	tablegconf "hcm/pkg/dal/table/global-config"
)

// ListReq ...
type ListReq struct {
	core.ListReq `json:",inline"`
}

// Validate ListReq
func (req *ListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ListReq.Validate(); err != nil {
		return err
	}

	return nil
}

// ListResp ...
type ListResp core.ListResultT[tablegconf.GlobalConfigTable]

// BatchCreateReq ...
type BatchCreateReq struct {
	Configs []cgconf.GlobalConfig `json:"configs" validate:"required,min=1,max=100"`
}

// Validate BatchCreateReq
func (req *BatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}

// BatchUpdateReq ...
type BatchUpdateReq struct {
	Configs []cgconf.GlobalConfig `json:"configs" validate:"required,min=1,max=100"`
}

// Validate BatchUpdateReq
func (req *BatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, config := range req.Configs {
		if len(config.ID) == 0 {
			return fmt.Errorf("config id can not be empty")
		}
	}

	return nil
}

// BatchDeleteReq ...
type BatchDeleteReq struct {
	core.BatchDeleteReq `json:",inline"`
}

// Validate BatchCreateReq
func (req *BatchDeleteReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.BatchDeleteReq.Validate(); err != nil {
		return err
	}

	return nil
}

// FindReq ...
type FindReq struct {
	Key  string `json:"key" validate:"required"`
	Type string `json:"type" validate:"required"`
}

// Validate FindReq
func (req *FindReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return nil
}
