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

package handlers

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"
)

// GetVpc 查询VPC
func (a *BaseApplicationHandler) GetVpc(vendor enumor.Vendor, accountID, cloudVpcID string) (
	*corecloud.BaseVpc, error) {

	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Vpc.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&core.ListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s vpc by cloud_id(%s)", vendor, cloudVpcID)
	}

	return &resp.Details[0], nil
}

// GetVpcByID 通过id查询VPC
func (a *BaseApplicationHandler) GetVpcByID(vendor enumor.Vendor, id string) (*corecloud.BaseVpc, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: id},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Vpc.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&core.ListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s vpc by id(%s)", vendor, id)
	}

	return &resp.Details[0], nil
}

// GetGcpVpcWithExt 查询gcp并带有扩展信息
func (a *BaseApplicationHandler) GetGcpVpcWithExt(vpcID string) (*corecloud.Vpc[corecloud.GcpVpcExtension], error) {

	// 查询
	resp, err := a.Client.DataService().Gcp.Vpc.ListVpcExt(a.Cts.Kit,
		&core.ListReq{Filter: tools.EqualExpression("id", vpcID), Page: &core.BasePage{Limit: 1}})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found gcp vpc by id(%s)", vpcID)
	}
	return &resp.Details[0], nil
}

// GetEip 获取eip信息
func (a *BaseApplicationHandler) GetEip(vendor enumor.Vendor, accountID string, cloudEipID string) (
	*eip.EipResult, error) {

	reqFilter := tools.ExpressionAnd(
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("account_id", accountID),
		tools.RuleEqual("cloud_id", cloudEipID),
	)
	// 查询
	resp, err := a.Client.DataService().Global.ListEip(a.Cts.Kit, &core.ListReq{
		Filter: reqFilter,
		Page:   a.getPageOfOneLimit(),
	},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s eip by cloud_id(%s)", vendor, cloudEipID)
	}

	return resp.Details[0], nil
}
