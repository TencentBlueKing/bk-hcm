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

package iam

import (
	"context"

	"hcm/cmd/auth-server/types"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/iam/client"
	"hcm/pkg/kit"
)

// ListInstances query instances based on filter criteria.
func (i *IAM) ListInstances(kt *kit.Kit, resType client.TypeID, option *types.ListInstanceFilter, page types.Page) (
	*types.ListInstanceResult, error) {
	filter, err := option.GetFilter(resType)
	if err != nil {
		return nil, err
	}

	countReq := &dataservice.ListInstancesReq{
		ResourceType: resType,
		Filter:       filter,
		Page:         &core.BasePage{Count: true},
	}
	countResp, err := i.ds.Global.Auth.ListInstances(kt.Ctx, kt.Header(), countReq)
	if err != nil {
		return nil, err
	}

	req := &dataservice.ListInstancesReq{
		ResourceType: resType,
		Filter:       filter,
		Page: &core.BasePage{
			Count: false,
			Start: uint32(page.Offset),
			Limit: page.Limit,
		},
	}
	resp, err := i.ds.Global.Auth.ListInstances(kt.Ctx, kt.Header(), req)
	if err != nil {
		return nil, err
	}

	instances := make([]types.InstanceResource, 0)
	for _, one := range resp.Details {
		instances = append(instances, types.InstanceResource{
			ID: types.InstanceID{
				InstanceID: one.ID,
			},
			DisplayName: one.DisplayName,
		})
	}

	result := &types.ListInstanceResult{
		Count:   countResp.Count,
		Results: instances,
	}
	return result, nil
}

// ListInstancesWithAttributes list resource instances that user is privileged to access by policy, returns id list.
func (i *IAM) ListInstancesWithAttributes(ctx context.Context, opts *client.ListWithAttributes) (idList []string,
	err error) {

	// TODO implement this when attribute auth is enabled
	return make([]string, 0), nil
}
