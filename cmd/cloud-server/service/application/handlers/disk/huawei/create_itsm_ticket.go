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
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfCreateHuaWeiDisk) RenderItsmTitle() (string, error) {
	name := "未命名"
	if a.req.DiskName != nil && *a.req.DiskName != "" {
		name = *a.req.DiskName
	}
	return fmt.Sprintf("申请新增[%s]云盘(%s)", a.Vendor().GetNameZh(), name), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfCreateHuaWeiDisk) RenderItsmForm() (string, error) {
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
	formItems = append(formItems, formItem{Label: "云厂商", Value: a.Vendor().GetNameZh()})

	// 云地域
	regionInfo, err := a.GetHuaWeiRegion(req.Region)
	if err != nil {
		return "", err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.LocalesZhCn})

	// 可用区
	zoneInfo, err := a.GetZone(a.Vendor(), req.Region, req.Zone)
	if err != nil {
		return "", err
	}
	formItems = append(formItems, formItem{Label: "可用区", Value: zoneInfo.Name})

	diskItems := []formItem{
		{Label: "云硬盘类型", Value: DiskTypeValue[req.DiskType]},
		{Label: "大小", Value: strconv.FormatInt(int64(req.DiskSize), 10)},
		{Label: "购买数量", Value: strconv.FormatInt(int64(req.DiskCount), 10)},
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
