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

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
)

// BaseResp http response.
type BaseResp struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// Response is a http standard response
type Response struct {
	Code        int32               `json:"code"`
	Message     string              `json:"message"`
	Permissions *meta.IamPermission `json:"permission,omitempty"`
	Data        interface{}         `json:"data"`
}

// NewBaseResp new BaseResp.
func NewBaseResp(code int32, msg string) *BaseResp {
	return &BaseResp{
		Code:    code,
		Message: msg,
	}
}

// WriteResp writer response to http.ResponseWriter.
func WriteResp(w http.ResponseWriter, resp interface{}) {
	bytes, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		logs.ErrorDepthf(1, "response marshal failed, err: %v", err)
		return
	}

	_, err = fmt.Fprintf(w, string(bytes))
	if err != nil {
		logs.ErrorDepthf(1, "write resp to ResponseWriter failed, err: %v", err)
		return
	}

	return
}

// FileDownloadResp define file download resp.
type FileDownloadResp interface {
	ContentType() string
	ContentDisposition() string
	Filepath() string
	IsDeleteFile() bool
}

// FileResp ...
type FileResp struct {
	ContentTypeStr        string
	ContentDispositionStr string
	FilePath              string
}

// ContentType ...
func (f *FileResp) ContentType() string {
	return f.ContentTypeStr
}

// ContentDisposition ...
func (f *FileResp) ContentDisposition() string {
	return f.ContentDispositionStr
}

// Filepath return file path.
func (f *FileResp) Filepath() string {
	return f.FilePath
}

// IsDeleteFile ...
func (f *FileResp) IsDeleteFile() bool {
	return false
}
