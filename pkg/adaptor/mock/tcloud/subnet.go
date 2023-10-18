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
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types/core"
	adtroutetable "hcm/pkg/adaptor/types/route-table"
	adtsubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/rand"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
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
		DoAndReturn(v.deleteSubnet).MinTimes(1)
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

	// 查找默认路由表，顺便校验vpc存在性
	defaultRT := v.routeTableStore.Find(func(table adtroutetable.TCloudRouteTable) bool {
		return table.Extension.Main && table.CloudVpcID == opt.CloudVpcID
	})
	if defaultRT == nil {
		return nil, errors.NewTencentCloudSDKError(tcloud.ErrNotFound,
			"[mock] can not find default route table for "+opt.CloudVpcID+", check vpc exists", "request-id")
	}
	ids := make([]string, len(opt.Subnets))
	// 创建子网请求
	for _, net := range opt.Subnets {
		// 未指定子网id则换成vpc默认子网id
		if len(net.CloudRouteTableID) == 0 {
			net.CloudRouteTableID = defaultRT.CloudID
		} else {
			// 尝试查找子网id
			if _, exists := v.routeTableStore.Get(net.CloudRouteTableID); !exists {
				return nil, errors.NewTencentCloudSDKError(tcloud.ErrNotFound,
					"[mock] can not find route table "+net.CloudRouteTableID, "request-id")
			}
		}
		createdSubnet := adtsubnet.TCloudSubnet{
			CloudVpcID: opt.CloudVpcID,
			CloudID:    rand.Prefix("subnet-", 8),
			Name:       net.Name,
			Ipv4Cidr:   []string{net.IPv4Cidr},
			Ipv6Cidr:   nil,
			Memo:       net.Memo,
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

		v.subnetStore.Add(createdSubnet.CloudID, createdSubnet)
		ids = append(ids, createdSubnet.CloudID)

	}

	return v.subnetStore.GetByKeys(ids...), nil

}

func (v *vpcPlaybook) deleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	if _, exists := v.subnetStore.Get(opt.ResourceID); !exists {
		return errf.Newf(errf.RecordNotFound, "not found ")
	}
	return v.subnetStore.Remove(opt.ResourceID)
}
