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

package routetable

import (
	"fmt"

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud/route-table"
	reltypes "hcm/pkg/dal/dao/types"
	routetable "hcm/pkg/dal/table/cloud/route-table"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func toProtoRouteTableExt[T protocore.RouteTableExtension](
	data *reltypes.RouteTableListResult) ([]*protocore.RouteTable[T], error) {

	details := make([]*protocore.RouteTable[T], len(data.Details))
	for idx, d := range data.Details {
		extResult, err := toProtoRouteTableExtWithID[T](d)
		if err != nil {
			return nil, err
		}
		details[idx] = extResult
	}
	return details, nil
}

func toProtoRouteTableExtWithID[T protocore.RouteTableExtension](d routetable.RouteTableTable) (
	*protocore.RouteTable[T], error) {

	var extension = new(T)
	err := json.UnmarshalFromString(string(d.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}

	return &protocore.RouteTable[T]{
		BaseRouteTable: protocore.BaseRouteTable{
			ID:         d.ID,
			Vendor:     d.Vendor,
			AccountID:  d.AccountID,
			CloudID:    d.CloudID,
			CloudVpcID: d.CloudVpcID,
			Name:       converter.PtrToVal(d.Name),
			Region:     d.Region,
			Memo:       d.Memo,
			VpcID:      d.VpcID,
			BkBizID:    d.BkBizID,
			Revision: &core.Revision{
				Creator:   d.Creator,
				Reviser:   d.Reviser,
				CreatedAt: d.CreatedAt.String(),
				UpdatedAt: d.UpdatedAt.String(),
			},
		},
		Extension: extension,
	}, nil
}
