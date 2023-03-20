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

package huawei

import (
	"fmt"
	"strconv"
	"strings"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/pkg/thirdparty/esb/itsm"
)

type formItem struct {
	Label string
	Value string
}

// CreateITSMTicket ...
func (a *ApplicationOfCreateHuaWeiDisk) CreateITSMTicket(serviceID int64, callbackUrl string) (string, error) {
	// 渲染ITSM表单内容
	contentDisplay, err := a.renderITSMForm()
	if err != nil {
		return "", fmt.Errorf("render itsm application form error: %w", err)
	}

	params := &itsm.CreateTicketParams{
		ServiceID:      serviceID,
		Creator:        a.Cts.Kit.User,
		CallbackURL:    callbackUrl,
		Title:          fmt.Sprintf("申请新增%s云盘[%s]", a.Vendor(), a.req.DiskName),
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

func (a *ApplicationOfCreateHuaWeiDisk) renderITSMForm() (string, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 业务
	bizName, err := a.GetBizName(req.BkBizID)
	if err != nil {
		return "", err
	}
	formItems = append(formItems, formItem{Label: "业务", Value: bizName})

	// 云账号
	accountInfo, err := a.GetAccount(req.AccountID)
	if err != nil {
		return "", err
	}
	formItems = append(formItems, formItem{Label: "云账号", Value: accountInfo.Name})

	// 云厂商
	formItems = append(formItems, formItem{Label: "云厂商", Value: handlers.VendorNameMap[a.Vendor()]})

	// 云地域
	regionInfo, err := a.GetHuaWeiRegion(req.Region)
	if err != nil {
		return "", err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.LocalesZhCn})

	// 可用区
	zoneInfo, err := a.GetZone(a.Vendor(), req.Region, req.Zone)
	formItems = append(formItems, formItem{Label: "可用区", Value: zoneInfo.Name})

	diskItems := []formItem{
		{Label: "云硬盘类型", Value: DiskTypeValue[req.DiskType]},
		{Label: "大小", Value: strconv.FormatInt(int64(*req.DiskCount), 10)},
		{Label: "购买数量", Value: strconv.FormatInt(int64(*req.DiskCount), 10)},
		{Label: "计费模式", Value: ChargeTypeValue[*req.DiskChargeType]},
	}

	if *req.DiskChargeType == "prePaid" {
		diskItems = append(
			diskItems,
			formItem{
				Label: "购买时长",
				Value: fmt.Sprintf(
					"%s %s",
					strconv.FormatInt(int64(*req.DiskChargePrepaid.PeriodNum), 10),
					ChargePeriodType[*req.DiskChargePrepaid.PeriodType],
				),
			},
		)

		if req.DiskChargePrepaid.IsAutoRenew != nil && *req.DiskChargePrepaid.IsAutoRenew == "ture" {
			diskItems = append(diskItems, formItem{Label: "自动续费", Value: "是"})
		} else {
			diskItems = append(diskItems, formItem{Label: "自动续费", Value: "否"})
		}
	}

	if req.Memo != nil && *req.Memo != "" {
		diskItems = append(diskItems, formItem{Label: "描述", Value: *req.Memo})
	}

	formItems = append(formItems, diskItems...)
	// 转换为ITSM表单内容数据
	content := make([]string, 0, len(formItems))
	for _, i := range formItems {
		content = append(content, fmt.Sprintf("%s: %s", i.Label, i.Value))
	}
	return strings.Join(content, "\n"), nil
}
