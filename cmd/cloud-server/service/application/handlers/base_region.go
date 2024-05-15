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
	corecloudregion "hcm/pkg/api/core/cloud/region"
	"hcm/pkg/runtime/filter"
)

// GetTCloudRegion 查询云地域信息
func (a *BaseApplicationHandler) GetTCloudRegion(region string) (*corecloudregion.TCloudRegion, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "region_id", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().TCloud.Region.ListRegion(
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
		return nil, fmt.Errorf("not found tcloud region by region_id(%s)", region)
	}

	return &resp.Details[0], nil
}

// GetAwsRegion 查询云地域信息
func (a *BaseApplicationHandler) GetAwsRegion(region string) (*corecloudregion.AwsRegion, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "region_id", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Aws.Region.ListRegion(
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
		return nil, fmt.Errorf("not found aws region by region_id(%s)", region)
	}

	return &resp.Details[0], nil
}

// GetHuaWeiRegion 查询云地域信息
func (a *BaseApplicationHandler) GetHuaWeiRegion(region string) (*corecloudregion.HuaWeiRegion, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "region_id", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().HuaWei.Region.ListRegion(
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
		return nil, fmt.Errorf("not found huawei region by region_id(%s)", region)
	}

	return &resp.Details[0], nil
}

// GetGcpRegion 查询云地域信息
func (a *BaseApplicationHandler) GetGcpRegion(region string) (*corecloudregion.GcpRegion, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "region_id", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Gcp.Region.ListRegion(
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
		return nil, fmt.Errorf("not found gcp region by region_id(%s)", region)
	}

	return &resp.Details[0], nil
}

// GetAzureRegion 查询云地域信息
func (a *BaseApplicationHandler) GetAzureRegion(region string) (*corecloudregion.AzureRegion, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "name", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Azure.Region.ListRegion(
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
		return nil, fmt.Errorf("not found azure region by name(%s)", region)
	}

	return &resp.Details[0], nil
}
