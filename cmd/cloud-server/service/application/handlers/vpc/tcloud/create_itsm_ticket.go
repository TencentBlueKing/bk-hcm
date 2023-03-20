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

package tcloud

import (
	"fmt"
	"strings"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/pkg/thirdparty/esb/itsm"
)

// CreateITSMTicket 使用请求数据创建申请
func (a *ApplicationOfCreateTCloudVpc) CreateITSMTicket(serviceID int64, callbackUrl string) (string, error) {
	// 渲染ITSM表单内容
	contentDisplay, err := a.renderITSMForm()
	if err != nil {
		return "", fmt.Errorf("render itsm application form error: %w", err)
	}

	params := &itsm.CreateTicketParams{
		ServiceID:      serviceID,
		Creator:        a.Cts.Kit.User,
		CallbackURL:    callbackUrl,
		Title:          fmt.Sprintf("申请新增VPC[%s]", a.req.Name),
		ContentDisplay: contentDisplay,
		// ITSM流程里使用变量引用的方式设置各个节点审批人
		VariableApprovers: []itsm.VariableApprover{
			{
				Variable:  "platform_manager",
				Approvers: a.platformManagers,
			},
		},
	}

	sn, err := a.EsbClient.Itsm().CreateTicket(a.Cts.Kit.Ctx, params)
	if err != nil {
		return "", fmt.Errorf("call itsm create ticket api failed, err: %v", err)
	}

	return sn, nil
}

type formItem struct {
	Label string
	Value string
}

func (a *ApplicationOfCreateTCloudVpc) renderBaseInfo() ([]formItem, error) {
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
	formItems = append(formItems, formItem{Label: "云厂商", Value: handlers.VendorNameMap[a.vendor]})

	// 云地域
	regionInfo, err := a.GetTCloudRegion(req.Region)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.RegionName})

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudVpc) renderVpc() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 名称
	formItems = append(formItems, formItem{Label: "名称", Value: req.Name})

	// IPv4 CIDR
	formItems = append(formItems, formItem{Label: "IPv4 CIDR", Value: req.IPv4Cidr})

	// 所属的蓝鲸云区域
	bkCloudAreaName, err := a.GetCloudAreaName(req.BkCloudID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "所属的蓝鲸云区域", Value: bkCloudAreaName})

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudVpc) renderSubnet() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 名称
	formItems = append(formItems, formItem{Label: "子网名称", Value: req.Subnet.Name})

	// IPv4 CIDR
	formItems = append(formItems, formItem{Label: "子网IPv4 CIDR", Value: req.Subnet.IPv4Cidr})

	// 可用区
	zoneInfo, err := a.GetZone(a.vendor, req.Region, req.Subnet.Zone)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "子网可用区", Value: zoneInfo.Name})

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudVpc) renderITSMForm() (string, error) {
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

	// 子网
	subnetFormItems, err := a.renderSubnet()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, subnetFormItems...)

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
