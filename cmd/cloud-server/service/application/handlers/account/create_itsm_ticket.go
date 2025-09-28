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

package account

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfAddAccount) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请新增[%s]账号(%s)", a.Vendor().GetNameZh(), a.req.Name), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfAddAccount) RenderItsmForm() (string, error) {
	req := a.req

	// 公共基础信息
	formItems := []formItem{
		{Label: "账号类型", Value: enumor.AccountTypeNameMap[req.Type]},
		{Label: "名称", Value: req.Name},
		{Label: "云厂商", Value: req.Vendor.GetNameZh()},
		{Label: "站点类型", Value: enumor.AccountSiteTypeNameMap[req.Site]},
	}

	// 云特性信息
	switch req.Vendor {
	case enumor.TCloud:
		formItems = append(formItems, []formItem{
			{Label: "主账号ID", Value: req.Extension["cloud_main_account_id"]},
			{Label: "子账号ID", Value: req.Extension["cloud_sub_account_id"]},
			{Label: "SecretId", Value: req.Extension["cloud_secret_id"]},
		}...)
	case enumor.Aws:
		formItems = append(formItems, []formItem{
			{Label: "账号ID", Value: req.Extension["cloud_account_id"]},
			{Label: "IAM用户名称", Value: req.Extension["cloud_iam_username"]},
			{Label: "SecretId/密钥ID", Value: req.Extension["cloud_secret_id"]},
		}...)
	case enumor.HuaWei:
		formItems = append(formItems, []formItem{
			{Label: "主账号名", Value: req.Extension["cloud_main_account_name"]},
			{Label: "账号ID", Value: req.Extension["cloud_sub_account_id"]},
			{Label: "账号名称", Value: req.Extension["cloud_sub_account_name"]},
			{Label: "IAM用户ID", Value: req.Extension["cloud_iam_user_id"]},
			{Label: "IAM用户名称", Value: req.Extension["cloud_iam_username"]},
			{Label: "SecretId/密钥ID", Value: req.Extension["cloud_secret_id"]},
		}...)
	case enumor.Gcp:
		formItems = append(formItems, []formItem{
			{Label: "项目 ID", Value: req.Extension["cloud_project_id"]},
			{Label: "项目名称", Value: req.Extension["cloud_project_name"]},
			{Label: "服务账号ID", Value: req.Extension["cloud_service_account_id"]},
			{Label: "服务账号名称", Value: req.Extension["cloud_service_account_name"]},
			{Label: "服务账号密钥ID", Value: req.Extension["cloud_service_secret_id"]},
		}...)
	case enumor.Azure:
		formItems = append(formItems, []formItem{
			{Label: "租户 ID", Value: req.Extension["cloud_tenant_id"]},
			{Label: "订阅 ID", Value: req.Extension["cloud_subscription_id"]},
			{Label: "订阅名称", Value: req.Extension["cloud_subscription_name"]},
			{Label: "应用程序(客户端) ID", Value: req.Extension["cloud_application_id"]},
			{Label: "应用程序名称", Value: req.Extension["cloud_application_name"]},
			{Label: "客户端密钥ID", Value: req.Extension["cloud_client_secret_id"]},
		}...)
	}

	// 负责人
	formItems = append(formItems, formItem{Label: "责任人", Value: strings.Join(req.Managers, ",")})

	// 查询业务名称
	if req.BkBizIDs[0] == constant.AttachedAllBiz {
		formItems = append(formItems, formItem{Label: "使用业务", Value: "全部"})
	} else {
		bizNames, err := a.ListBizNames(req.BkBizIDs)
		if err != nil {
			return "", fmt.Errorf("list biz name failed, bk_biz_ids: %v, err: %w", req.BkBizIDs, err)
		}
		formItems = append(formItems, formItem{Label: "使用业务", Value: strings.Join(bizNames, ",")})
	}

	// 备注
	if req.Memo != nil && *req.Memo != "" {
		formItems = append(formItems, formItem{Label: "备注", Value: *req.Memo})
	}

	// 转换为ITSM表单内容数据
	content := make([]string, 0, len(formItems))
	for _, i := range formItems {
		content = append(content, fmt.Sprintf("%s: %s", i.Label, i.Value))
	}
	return strings.Join(content, "\n"), nil
}
