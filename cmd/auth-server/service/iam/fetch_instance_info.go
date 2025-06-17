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
	"hcm/cmd/auth-server/types"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// FetchInstanceInfo obtain resource instance details in batch.
func (i *IAM) FetchInstanceInfo(kt *kit.Kit, resType iam.TypeID, ft *types.FetchInstanceInfoFilter) (
	[]map[string]interface{}, error) {

	// TODO: f.Attrs need to deal with, if add attribute authentication.

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: ft.IDs,
			},
		},
	}

	req := &dataservice.ListInstancesReq{
		ResourceType: resType,
		Filter:       expr,
		Page:         &core.BasePage{Count: false, Limit: uint(len(ft.IDs))},
	}
	resp, err := i.ds.Global.Auth.ListInstances(kt.Ctx, kt.Header(), req)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for _, one := range resp.Details {
		result := make(map[string]interface{}, 0)
		result[types.IDField] = types.InstanceID{
			InstanceID: one.ID,
		}
		result[types.NameField] = one.DisplayName
		results = append(results, result)
	}

	return results, nil
}
