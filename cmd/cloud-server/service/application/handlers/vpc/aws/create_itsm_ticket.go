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

package aws

import (
	"fmt"
	"strings"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfCreateAwsVpc) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请新增[%s]VPC[%s]", a.Vendor().GetNameZh(), a.req.Name), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfCreateAwsVpc) RenderItsmForm() (string, error) {
	req := a.req

	formItems := make([]formItem, 0)

	// 基本通用信息
	baseInfoFormItems, err := a.renderBaseInfo()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, baseInfoFormItems...)

	// VPC
	vpcFormItems, err := a.renderVpc()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, vpcFormItems...)

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

func (a *ApplicationOfCreateAwsVpc) renderBaseInfo() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 业务
	bizName, err := a.GetBizName(req.BkBizID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "业务", Value: bizName})

	// 云账号
	accountInfo, err := a.GetAccount(req.AccountID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云账号", Value: accountInfo.Name})

	// 云厂商
	formItems = append(formItems, formItem{Label: "云厂商", Value: a.Vendor().GetNameZh()})

	// 云地域
	regionInfo, err := a.GetAwsRegion(req.Region)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.RegionName})

	return formItems, nil
}

func (a *ApplicationOfCreateAwsVpc) renderVpc() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 名称
	formItems = append(formItems, formItem{Label: "名称", Value: req.Name})

	// IPv4 CIDR
	formItems = append(formItems, formItem{Label: "IPv4 CIDR", Value: req.IPv4Cidr})

	// 租期
	instanceTenancyNameMap := map[string]string{"default": "默认", "dedicated": "专用"}
	formItems = append(formItems, formItem{
		Label: "租期",
		Value: instanceTenancyNameMap[req.InstanceTenancy],
	})

	return formItems, nil
}
