/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package cos tencent cloud cos api
package cos

import (
	"hcm/pkg/criteria/validator"
)

// GenerateTemporalUrlReq ...
type GenerateTemporalUrlReq struct {
	Filename   string `json:"filename" validate:"omitempty"`
	TTLSeconds int64  `json:"ttl_seconds" validate:"required,max=3600"`
}

// Validate ...
func (r *GenerateTemporalUrlReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GenerateTemporalUrlResult ...
type GenerateTemporalUrlResult struct {
	AK    string `json:"ak"`
	SK    string `json:"sk"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

// UploadFileReq ...
type UploadFileReq struct {
	Filename   string `json:"filename" validate:"required"`
	FileBase64 string `json:"file_base64" validate:"required"`
}

// Validate ...
func (r *UploadFileReq) Validate() error {
	return validator.Validate.Struct(r)
}
