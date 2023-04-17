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
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
)

// Deliver 执行资源交付
func (a *ApplicationOfCreateHuaWeiVpc) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// 创建主机
	result, err := a.Client.HCService().HuaWei.Vpc.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		a.toHcProtoVpcCreateReq(),
	)
	if err != nil || result == nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	return enumor.Completed, map[string]interface{}{"vpc_id": result.ID}, nil
}

func (a *ApplicationOfCreateHuaWeiVpc) toHcProtoVpcCreateReq() *hcproto.VpcCreateReq[hcproto.HuaWeiVpcCreateExt] {
	req := a.req

	return &hcproto.VpcCreateReq[hcproto.HuaWeiVpcCreateExt]{
		BaseVpcCreateReq: &hcproto.BaseVpcCreateReq{
			AccountID: req.AccountID,
			Name:      req.Name,
			Category:  enumor.BizVpcCategory,
			Memo:      req.Memo,
			BkCloudID: req.BkCloudID,
			BkBizID:   req.BkBizID,
		},
		Extension: &hcproto.HuaWeiVpcCreateExt{
			Region:   req.Region,
			IPv4Cidr: req.IPv4Cidr,
			Subnets: []hcproto.SubnetCreateReq[hcproto.HuaWeiSubnetCreateExt]{
				{
					BaseSubnetCreateReq: &hcproto.BaseSubnetCreateReq{
						AccountID: req.AccountID,
						Name:      req.Subnet.Name,
						Memo:      req.Memo,
						BkBizID:   req.BkBizID,
					},
					Extension: &hcproto.HuaWeiSubnetCreateExt{
						Region:     req.Region,
						IPv4Cidr:   req.Subnet.IPv4Cidr,
						Ipv6Enable: *req.Subnet.IPv6Enable,
						GatewayIp:  req.Subnet.GatewayIP,
					},
				},
			},
		},
	}
}
