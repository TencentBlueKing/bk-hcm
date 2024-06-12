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

package mainaccount

import (
	"fmt"
	"strings"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfCreateMainAccount) RenderItsmTitle() (string, error) {
	fieldName := a.req.Vendor.GetMainAccountNameFieldName()
	return fmt.Sprintf("申请创建[%s]二级账号(%s)", a.Vendor().GetNameZh(), a.req.Extension[fieldName]), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfCreateMainAccount) RenderItsmForm() (string, error) {
	req := a.req

	// 获取业务名字
	bkName, err := a.GetBizName(req.BkBizID)
	if err != nil {
		return "", fmt.Errorf("list biz name failed, bk_biz_ids: %v, err: %w", req.BkBizID, err)
	}

	// 公共参数
	formItems := []formItem{
		{Label: "云厂商", Value: req.Vendor.GetNameZh()},
		{Label: "站点类型", Value: req.Site.GetMainAccountSiteTypeName()},
		{Label: "账号名称", Value: req.Extension[req.Vendor.GetMainAccountNameFieldName()]},
		{Label: "账号邮箱", Value: req.Email},
		{Label: "使用业务", Value: bkName},
		{Label: "负责人", Value: strings.Join(req.Managers, ",")},
		{Label: "备份负责人", Value: strings.Join(req.BakManagers, ",")},
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
