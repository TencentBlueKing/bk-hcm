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

package template

import (
	"fmt"
	"os"

	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitTemplateService ...
func InitTemplateService(c *capability.Capability) {
	svr := &service{
		client: c.ApiClient,
	}

	h := rest.NewHandler()
	h.Add("GetTemplate", "GET", "/templates/{filename}", svr.GetTemplate)

	h.Load(c.WebService)
}

type service struct {
	client *client.ClientSet
}

// GetTemplate get template file
func (u *service) GetTemplate(cts *rest.Contexts) (interface{}, error) {

	filename := cts.PathParameter("filename").String()
	if filename == "" {
		return nil, fmt.Errorf("filename is empty")
	}

	if err := validateTemplateFilename(filename); err != nil {
		logs.Errorf("failed to validate template filename: %s, err: %v, rid: %s", filename, err, cts.Kit.Rid)
		return nil, err
	}

	templateDirPath := cc.WebServer().TemplatePath
	files, err := os.ReadDir(templateDirPath)
	if err != nil {
		logs.Errorf("failed to read directory: %s, err: %v, rid: %s", templateDirPath, err, cts.Kit.Rid)
		return nil, err
	}
	var filepath string
	for _, file := range files {
		if file.Name() == filename {
			filepath = templateDirPath + "/" + filename
		}
	}
	if filepath == "" {
		return nil, fmt.Errorf("file not found: %s", filename)
	}

	return &FileResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filename),
		FilePath:              filepath,
	}, nil
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
