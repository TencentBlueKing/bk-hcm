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

	cvt "hcm/pkg/tools/converter"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfCreateTCloudLB) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请新增[%s]负载均衡(%s)", a.Vendor().GetNameZh(), cvt.PtrToVal(a.req.Name)), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfCreateTCloudLB) RenderItsmForm() (string, error) {
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

	// 计费
	formItems = append(formItems, a.renderInstanceChargeForm()...)

	// 购买数量
	count := cvt.PtrToVal(req.RequireCount)
	if count == 0 {
		count = 1
	}
	formItems = append(formItems, formItem{Label: "购买数量", Value: fmt.Sprintf("%d", count)})

	// 备注
	if req.Memo != "" {
		formItems = append(formItems, formItem{Label: "备注", Value: req.Memo})
	}

	// 转换为ITSM表单内容数据
	content := make([]string, 0, len(formItems))
	for _, i := range formItems {
		content = append(content, fmt.Sprintf("%s: %s", i.Label, i.Value))
	}
	return strings.Join(content, "\n"), nil
}

func (a *ApplicationOfCreateTCloudLB) renderBaseInfo() ([]formItem, error) {
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
	regionInfo, err := a.GetTCloudRegion(req.Region)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.RegionName})

	// 可用区
	zones := append(req.Zones, req.BackupZones...)
	if len(zones) > 0 {
		zoneInfos, err := a.GetZones(a.Vendor(), req.Region, zones)
		if err != nil {
			return formItems, err
		}
		zoneStr := strings.Builder{}
		for i, info := range zoneInfos {
			if i > 0 {
				zoneStr.WriteRune(',')
			}
			zoneStr.WriteString(info.Name)
			if len(info.NameCn) != 0 {
				zoneStr.WriteRune('(')
				zoneStr.WriteString(info.NameCn)
				zoneStr.WriteRune(')')
			}
		}
		formItems = append(formItems, formItem{Label: "可用区", Value: zoneStr.String()})
	}

	// 名称
	formItems = append(formItems, formItem{Label: "名称", Value: cvt.PtrToVal(req.Name)})

	// 规格
	slaType := "共享型"
	if req.SlaType != nil {
		slaType = *req.SlaType
	}
	formItems = append(formItems, formItem{Label: "规格", Value: slaType})

	// 运营商
	isp := "BGP"
	if req.VipIsp != nil {
		isp = *req.VipIsp
	}
	formItems = append(formItems, formItem{Label: "运营商", Value: isp})

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudLB) renderNetwork() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 内网公网类型
	formItems = append(formItems, formItem{Label: "类型", Value: LoadBalancerTypeMap[req.LoadBalancerType]})

	formItems = append(formItems, formItem{Label: "IP版本", Value: IPVersionNameMap[req.AddressIPVersion]})

	// VPC
	vpcName := "未指定-默认VPC"
	if req.CloudVpcID != nil {
		vpcInfo, err := a.GetVpc(a.Vendor(), req.AccountID, cvt.PtrToVal(req.CloudVpcID))
		if err != nil {
			return formItems, err
		}
		vpcName = fmt.Sprintf("%s(%s)", vpcInfo.CloudID, vpcInfo.Name)
	}
	formItems = append(formItems, formItem{Label: "VPC", Value: vpcName})

	// subnet
	subnetName := "未指定"
	if req.CloudVpcID != nil && req.CloudSubnetID != nil {
		// 子网
		subnetInfo, err := a.GetSubnet(a.Vendor(), req.AccountID, cvt.PtrToVal(req.CloudVpcID),
			cvt.PtrToVal(req.CloudSubnetID))
		if err != nil {
			return formItems, err
		}
		subnetName = fmt.Sprintf("%s(%s)", subnetInfo.CloudID, subnetInfo.Name)
	}
	formItems = append(formItems, formItem{Label: "子网", Value: subnetName})

	// EIP信息
	if req.CloudEipID != nil {
		eipInfo, err := a.GetEip(a.Vendor(), req.AccountID, cvt.PtrToVal(req.CloudEipID))
		if err != nil {
			return formItems, err
		}
		formItems = append(formItems, formItem{
			Label: "EIP",
			Value: fmt.Sprintf("%s(%s)", eipInfo.PublicIp, eipInfo.CloudID),
		})
	}

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudLB) renderInstanceChargeForm() []formItem {
	req := a.req
	formItems := make([]formItem, 0)

	payMode := "按量计费"
	if req.InternetChargeType != nil {
		payMode = LoadBalancerNetworkChargeTypeNameMap[*req.InternetChargeType]
	}
	if payMode == "" {
		payMode = string(*req.InternetChargeType)
	}
	// 计费模式
	formItems = append(formItems, formItem{Label: "网络计费模式", Value: payMode})

	// 是否自动续费
	if req.AutoRenew != nil && *req.AutoRenew {
		formItems = append(formItems, formItem{Label: "是否自动续费", Value: "是"})
	} else {
		formItems = append(formItems, formItem{Label: "是否自动续费", Value: "否"})
	}

	return formItems
}
