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

package mocktcloud

import (
	"hcm/pkg/adaptor/types/core"
	adtroutetable "hcm/pkg/adaptor/types/route-table"
	adtsubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/rand"

	"go.uber.org/mock/gomock"
)

func (v *vpcPlaybook) applySubnet(mockCloud *MockTCloud) {

	mockCloud.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).DoAndReturn(v.listSubnet).MinTimes(1)

	mockCloud.EXPECT().CreateSubnets(gomock.Any(), gomock.Any()).DoAndReturn(v.createSubnet).MinTimes(1)

	// only return nil, do nothing
	mockCloud.EXPECT().
		UpdateSubnet(gomock.Any(), gomock.AssignableToTypeOf((*adtsubnet.TCloudSubnetUpdateOption)(nil))).
		Return(nil).MinTimes(1)

	mockCloud.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).
		DoAndReturn(func(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
			if err := opt.Validate(); err != nil {
				return err
			}
			return v.subnetStore.Remove(opt.ResourceID)
		}).MinTimes(1)
}

func (v *vpcPlaybook) listSubnet(_ *kit.Kit, opt *core.TCloudListOption) (*adtsubnet.TCloudSubnetListResult,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}
	var values []adtsubnet.TCloudSubnet
	if len(opt.CloudIDs) == 0 {
		v.subnetStore.Filter(func(subnet adtsubnet.TCloudSubnet) bool {
			return subnet.Extension.Region == opt.Region
		})
	} else {
		values = v.subnetStore.GetByKeys(opt.CloudIDs...)
	}
	return &adtsubnet.TCloudSubnetListResult{
		Count:   converter.ValToPtr(uint64(len(values))),
		Details: values,
	}, nil
}

func (v *vpcPlaybook) createSubnet(_ *kit.Kit, opt *adtsubnet.TCloudSubnetsCreateOption) ([]adtsubnet.TCloudSubnet,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	ids := make([]string, len(opt.Subnets))
	// 创建子网
	for _, net := range opt.Subnets {
		createdSubnet := adtsubnet.TCloudSubnet{
			CloudVpcID: opt.CloudVpcID,
			CloudID:    "subnet-" + rand.String(8),
			Name:       net.Name,
			Ipv4Cidr:   []string{net.IPv4Cidr},
			Ipv6Cidr:   nil,
			Memo:       nil,
			Extension: &adtsubnet.TCloudSubnetExtension{
				IsDefault:               false,
				Region:                  opt.Region,
				Zone:                    net.Zone,
				CloudRouteTableID:       &net.CloudRouteTableID,
				CloudNetworkAclID:       nil,
				AvailableIPAddressCount: 0,
				TotalIpAddressCount:     0,
				UsedIpAddressCount:      0,
			},
		}

		// 创建默认路由表
		routeTable := adtroutetable.TCloudRouteTable{
			CloudID:    "rtb-" + rand.String(8),
			Name:       "default",
			CloudVpcID: createdSubnet.CloudVpcID,
			Region:     createdSubnet.Extension.Region,
			Memo:       converter.ValToPtr("default"),
			Extension:  &adtroutetable.TCloudRouteTableExtension{Main: true},
		}
		v.routeTableStore.Add(routeTable.CloudID, routeTable)

		createdSubnet.Extension.CloudRouteTableID = &routeTable.CloudID
		v.subnetStore.Add(createdSubnet.CloudID, createdSubnet)
		ids = append(ids, createdSubnet.CloudID)

	}

	return v.subnetStore.GetByKeys(ids...), nil

}
