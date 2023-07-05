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
	"hcm/cmd/cloud-server/service/application/handlers/vpc/logics"
	"hcm/cmd/cloud-server/service/common"
	"hcm/pkg/criteria/enumor"
)

// Deliver 执行资源交付
func (a *ApplicationOfCreateHuaWeiVpc) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// 创建vpc
	result, err := a.Client.HCService().HuaWei.Vpc.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		common.ConvHuaWeiVpcCreateReq(a.req),
	)
	if err != nil || result == nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	// 交付vpc到业务下
	deliverVpcResult, err := logics.DeliverVpc(a.Cts.Kit, a.req.BkBizID,
		a.Client.DataService(), a.Audit, result.ID)
	if err != nil {
		return enumor.DeliverError, deliverVpcResult, err
	}

	// 查询vpc
	vpcInfo, err := a.GetVpcByID(a.Vendor(), result.ID)
	if err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	// 查询子网
	subnetsInfo, err := a.GetSubnetsByCloudVpcID(a.Vendor(), a.req.AccountID, vpcInfo.CloudID)
	if err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	if len(subnetsInfo) > 0 {
		// 交付子网到业务下
		subnetIDs := make([]string, 0, len(subnetsInfo))
		for _, one := range subnetsInfo {
			subnetIDs = append(subnetIDs, one.ID)
		}
		deliverSubnetResult, err := logics.DeliverSubnet(a.Cts.Kit, a.req.BkBizID,
			a.Client.DataService(), a.Audit, subnetIDs)
		if err != nil {
			return enumor.DeliverError, deliverSubnetResult, err
		}
	}

	// 查询路由表
	routetableInfo, err := a.GetRouteTables(a.Vendor(), a.req.AccountID, vpcInfo.CloudID)
	if err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	if len(routetableInfo) > 0 {
		// 交付路由表到业务下
		routetableIDs := make([]string, 0, len(routetableInfo))
		for _, one := range routetableInfo {
			routetableIDs = append(routetableIDs, one.ID)
		}
		deliverRouteTableResult, err := logics.DeliverRouteTable(a.Cts.Kit, a.req.BkBizID,
			a.Client.DataService(), a.Audit, routetableIDs)
		if err != nil {
			return enumor.DeliverError, deliverRouteTableResult, err
		}
	}

	return enumor.Completed, map[string]interface{}{"vpc_id": result.ID}, nil
}
