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

package azure

import (
	"fmt"
	"strings"

	"hcm/pkg/tools/converter"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfCreateAzureCvm) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请新增[%s]虚拟机(%s)", a.Vendor().GetNameZh(), a.req.Name), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfCreateAzureCvm) RenderItsmForm() (string, error) {
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

	// 登录用户名
	formItems = append(formItems, formItem{Label: "登录用户名", Value: req.Username})

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

func (a *ApplicationOfCreateAzureCvm) renderBaseInfo() ([]formItem, error) {
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

	// 资源组
	formItems = append(formItems, formItem{Label: "资源组", Value: req.ResourceGroupName})

	// 云地域
	regionInfo, err := a.GetAzureRegion(req.Region)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.Name})

	// Note: 主机申请这里的可用区与Azure的统一可用区不是一个概念
	formItems = append(formItems, formItem{Label: "可用区", Value: req.Zone})

	// 名称
	formItems = append(formItems, formItem{Label: "名称", Value: req.Name})

	// 机型
	instanceTypeInfo, err := a.GetAzureInstanceType(req.AccountID, req.Region, req.InstanceType)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{
		Label: "机型",
		Value: fmt.Sprintf("%s (%d核%sG)",
			req.InstanceType, instanceTypeInfo.CPU, a.ConvertMemoryMBToGB(instanceTypeInfo.Memory)),
	})

	// 镜像
	imageInfo, err := a.GetImage(a.Vendor(), req.CloudImageID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "镜像", Value: imageInfo.Name})

	return formItems, nil
}

func (a *ApplicationOfCreateAzureCvm) renderNetwork() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// VPC
	vpcInfo, err := a.GetVpc(a.Vendor(), req.AccountID, req.CloudVpcID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "VPC", Value: vpcInfo.Name})

	// 子网
	subnetInfo, err := a.GetSubnet(a.Vendor(), req.AccountID, req.CloudVpcID, req.CloudSubnetID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "子网", Value: subnetInfo.Name})

	// 是否自动分配公网IP
	if req.PublicIPAssigned {
		formItems = append(formItems, formItem{Label: "是否自动分配公网IP", Value: "是"})
	} else {
		formItems = append(formItems, formItem{Label: "是否自动分配公网IP", Value: "否"})
	}

	// 所属的蓝鲸云区域
	bkCloudAreaName, err := a.GetCloudAreaName(converter.PtrToVal(req.BkCloudID))
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "所属的蓝鲸云区域", Value: bkCloudAreaName})

	// 安全组
	securityGroups, err := a.ListSecurityGroup(a.Vendor(), req.AccountID, req.CloudSecurityGroupIDs)
	securityGroupNames := make([]string, 0, len(req.CloudSecurityGroupIDs))
	for _, s := range securityGroups {
		securityGroupNames = append(securityGroupNames, s.Name)
	}
	formItems = append(formItems, formItem{Label: "安全组", Value: strings.Join(securityGroupNames, ",")})

	return formItems, nil
}

func (a *ApplicationOfCreateAzureCvm) renderDiskForm() []formItem {
	req := a.req
	formItems := make([]formItem, 0)

	// 系统盘
	formItems = append(formItems, formItem{
		Label: "系统盘",
		Value: fmt.Sprintf("%s, %dGB", DiskTypeNameMap[req.SystemDisk.DiskType], req.SystemDisk.DiskSizeGB),
	})

	// 数据盘
	disks := make([]string, 0, len(req.DataDisk))
	for _, d := range req.DataDisk {
		disks = append(disks, fmt.Sprintf("%s(%dGB,%d个)", DiskTypeNameMap[d.DiskType], d.DiskSizeGB, d.DiskCount))
	}
	formItems = append(formItems, formItem{Label: "数据盘", Value: strings.Join(disks, ",")})

	return formItems
}
