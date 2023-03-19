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

package gcp

import (
	"fmt"
	"strings"

	"hcm/cmd/cloud-server/service/application/handlers"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/thirdparty/esb/itsm"
)

// CreateITSMTicket 使用请求数据创建申请
func (a *ApplicationOfCreateGcpCvm) CreateITSMTicket(serviceID int64, callbackUrl string) (string, error) {
	// 渲染ITSM表单内容
	contentDisplay, err := a.renderITSMForm()
	if err != nil {
		return "", fmt.Errorf("render itsm application form error: %w", err)
	}

	params := &itsm.CreateTicketParams{
		ServiceID:      serviceID,
		Creator:        a.Cts.Kit.User,
		CallbackURL:    callbackUrl,
		Title:          fmt.Sprintf("申请新增虚拟机[%s]", a.req.Name),
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

func (a *ApplicationOfCreateGcpCvm) renderBaseInfo() ([]formItem, error) {
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
	regionInfo, err := a.GetGcpRegion(req.Region)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.RegionName})

	// 可用区
	zoneInfo, err := a.GetZone(a.vendor, req.Region, req.Zone)
	formItems = append(formItems, formItem{Label: "可用区", Value: zoneInfo.Name})

	// 名称
	formItems = append(formItems, formItem{Label: "名称", Value: req.Name})

	// 机型
	instanceTypeInfo, err := a.GetGcpInstanceType(req.AccountID, req.Zone, req.InstanceType)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{
		Label: "机型",
		Value: fmt.Sprintf("%s (%d核%sG)",
			req.InstanceType, instanceTypeInfo.CPU, a.ConvertMemoryMBToGB(instanceTypeInfo.Memory)),
	})

	// 镜像
	imageInfo, err := a.GetImage(a.vendor, req.CloudImageID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "镜像", Value: imageInfo.Name})

	return formItems, nil
}

func (a *ApplicationOfCreateGcpCvm) renderNetwork() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// VPC
	vpcInfo, err := a.GetVpc(a.vendor, req.AccountID, req.CloudVpcID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "VPC", Value: vpcInfo.Name})

	// 子网
	subnetInfo, err := a.GetSubnet(a.vendor, req.AccountID, req.CloudVpcID, req.CloudSubnetID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "子网", Value: subnetInfo.Name})

	// 所属的蓝鲸云区域
	bkCloudAreaName, err := a.GetCloudAreaName(vpcInfo.BkCloudID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "所属的蓝鲸云区域", Value: bkCloudAreaName})

	return formItems, nil
}

func (a *ApplicationOfCreateGcpCvm) renderDiskForm() []formItem {
	req := a.req
	formItems := make([]formItem, 0)

	// 系统盘
	formItems = append(formItems, formItem{
		Label: "系统盘",
		Value: fmt.Sprintf("%s, %dGB", DiskTypeNameMap[req.SystemDisk.DiskType], req.SystemDisk.DiskSizeGB),
	})

	// 数据盘
	disks := make([]string, 0, len(req.DataDisk))
	modeNameMap := map[typecvm.GcpDiskMode]string{typecvm.ReadOnly: "只读", typecvm.ReadWrite: "读写"}
	autoDeleteNameMap := map[bool]string{true: "删除磁盘", false: "保留磁盘"}
	for _, d := range req.DataDisk {
		disks = append(
			disks,
			fmt.Sprintf(
				"%s(%dGB,%d个,%s,%s)",
				DiskTypeNameMap[d.DiskType], d.DiskSizeGB, d.DiskCount,
				modeNameMap[d.Mode], autoDeleteNameMap[d.AutoDelete],
			),
		)
	}
	formItems = append(formItems, formItem{Label: "数据盘", Value: strings.Join(disks, ",")})

	return formItems
}

func (a *ApplicationOfCreateGcpCvm) renderITSMForm() (string, error) {
	req := a.req

	formItems := make([]formItem, 0)

	// 基本通用信息
	baseInfoFormItems, err := a.renderBaseInfo()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, baseInfoFormItems...)

	// 网络
	networkFormItems, err := a.renderNetwork()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, networkFormItems...)

	// 硬盘
	formItems = append(formItems, a.renderDiskForm()...)

	// 购买数量
	formItems = append(formItems, formItem{Label: "购买数量", Value: fmt.Sprintf("%d", req.RequiredCount)})

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
