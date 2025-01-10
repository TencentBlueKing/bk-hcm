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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/runtime/filter"
)

// GetSubnet 查询子网
func (a *BaseApplicationHandler) GetSubnet(vendor enumor.Vendor, accountID, cloudVpcID, cloudSubnetID,
	region string) (*corecloud.BaseSubnet, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudSubnetID},
			filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Subnet.List(
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
		return nil, fmt.Errorf("not found %s subnet by cloud_id(%s)", vendor, cloudSubnetID)
	}

	return &resp.Details[0], nil
}

// GetSubnetsByCloudVpcID 通过vpc id查询子网
func (a *BaseApplicationHandler) GetSubnetsByCloudVpcID(
	vendor enumor.Vendor, accountID, cloudVpcID string,
) ([]corecloud.BaseSubnet, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Subnet.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&core.ListReq{
			Filter: reqFilter,
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Details, nil
}
