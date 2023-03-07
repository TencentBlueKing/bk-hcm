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

package networkcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/api/data-service/cloud"
	reltypes "hcm/pkg/dal/dao/types"
	"hcm/pkg/tools/json"
)

func toProtoNetworkInterfaceExtWithCvmIDs[T coreni.NetworkInterfaceExtension](
	data *reltypes.ListCvmRelsJoinNetworkInterfaceDetails) ([]*cloud.NetworkInterfaceExtWithCvmID[T], error) {

	details := make([]*cloud.NetworkInterfaceExtWithCvmID[T], len(data.Details))
	for idx, d := range data.Details {
		extResult, err := toProtoNetworkInterfaceExtWithCvmID[T](d)
		if err != nil {
			return nil, err
		}
		details[idx] = extResult
	}
	return details, nil
}

func toProtoNetworkInterfaceExtWithCvmID[T coreni.NetworkInterfaceExtension](d *reltypes.NetworkInterfaceWithCvmID) (
	*cloud.NetworkInterfaceExtWithCvmID[T], error) {

	var extension = new(T)
	err := json.UnmarshalFromString(string(d.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}

	return &cloud.NetworkInterfaceExtWithCvmID[T]{
		NetworkInterface: coreni.NetworkInterface[T]{
			BaseNetworkInterface: coreni.BaseNetworkInterface{
				ID:            d.ID,
				Vendor:        d.Vendor,
				Name:          d.Name,
				AccountID:     d.AccountID,
				Region:        d.Region,
				Zone:          d.Zone,
				CloudID:       d.CloudID,
				VpcID:         d.VpcID,
				CloudVpcID:    d.CloudVpcID,
				SubnetID:      d.SubnetID,
				CloudSubnetID: d.CloudSubnetID,
				PrivateIPv4:   d.PrivateIPv4,
				PrivateIPv6:   d.PrivateIPv6,
				PublicIPv4:    d.PublicIPv4,
				PublicIPv6:    d.PublicIPv6,
				BkBizID:       d.BkBizID,
				InstanceID:    d.InstanceID,
				Revision: &core.Revision{
					Creator:   d.Creator,
					Reviser:   d.Reviser,
					CreatedAt: d.CreatedAt,
					UpdatedAt: d.UpdatedAt,
				},
			},
			Extension: extension,
		},
		CvmID:        d.CvmID,
		RelCreator:   d.RelCreator,
		RelCreatedAt: d.RelCreatedAt,
	}, nil
}
