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

package itsm

import (
	"bytes"
	"fmt"
	"html/template"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// MigrateTemplate ITSM 流程注册模板
type MigrateTemplate struct {
	Name    string
	Content string
}

// MigrateTemplates 有序的 ITSM 流程注册模板列表，按注册顺序排列
var MigrateTemplates = []MigrateTemplate{
	{Name: "processInitTemplate", Content: processInitTemplate},
	{Name: "processInitTemplate20260520", Content: processInitTemplate20260520},
}

// SystemMigrate 系统流程导入，使用指定的模板内容注册
func (i *itsm) SystemMigrate(kt *kit.Kit, systemID string, templateContent string) error {

	tmpl, err := template.New("itsm_init_temp").Parse(templateContent)
	if err != nil {
		logs.Errorf("failed to parse itsm init template, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 渲染模版中的租户ID
	renderParams := map[string]string{
		"systemID": systemID,
		"tenantID": kt.TenantID,
	}

	var fileBytes bytes.Buffer
	err = tmpl.Execute(&fileBytes, renderParams)
	if err != nil {
		logs.Errorf("failed to execute itsm init template, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	uploadFile := &apigateway.UploadFileReq{
		FieldName: "file",
		FileName:  "process_init.json",
		FileBytes: fileBytes.Bytes(),
	}

	code, msg, res, err := apigateway.ApiGatewayCallOriginal[CreateTicketReq, CreateTicketResult](i.client,
		i.bkUserCli, i.config, rest.POST, kt, nil, uploadFile, "/system/migrate/")

	if err != nil {
		return err
	}

	// itsm成功时状态码为20000
	if code != success {
		err := fmt.Errorf("failed to call api gateway to migrate system, code: %d, msg: %s", code, msg)
		logs.Errorf("%s, result: %+v, rid: %s", err, res, kt.Rid)
		return err
	}

	return nil
}
