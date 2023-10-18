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
	adaptormock "hcm/pkg/adaptor/mock"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	adtroutetable "hcm/pkg/adaptor/types/route-table"
	adtsubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/rand"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"go.uber.org/mock/gomock"
)

// vpcPlaybook in memory crud vpc playbook, add ability to mock vpc, subnet, route table related functions.
type vpcPlaybook struct {
	vpcStore        *adaptormock.Store[string, types.TCloudVpc]
	subnetStore     *adaptormock.Store[string, adtsubnet.TCloudSubnet]
	routeTableStore *adaptormock.Store[string, adtroutetable.TCloudRouteTable]
}

// NewCrudVpcPlaybook  vpc, subnet, route table all in one.
func NewCrudVpcPlaybook() Playbook {

	return &vpcPlaybook{
		vpcStore:        adaptormock.NewCloudResStore[types.TCloudVpc](),
		subnetStore:     adaptormock.NewCloudResStore[adtsubnet.TCloudSubnet](),
		routeTableStore: adaptormock.NewCloudResStore[adtroutetable.TCloudRouteTable](),
	}
}

// Name vpc playbook
func (v *vpcPlaybook) Name() string {
	return string(enumor.VpcCloudResType)
}

// Apply mock method
func (v *vpcPlaybook) Apply(mockCloud *MockTCloud, ctrl *gomock.Controller) {

	v.applyVpc(mockCloud)

	v.applySubnet(mockCloud)

	v.applyRouteTable(mockCloud)

}

func (v *vpcPlaybook) applyVpc(mockCloud *MockTCloud) {

	// 相当于将对mock对象ListVpc方法的调用转化成对v.listVpc的调用
	mockCloud.EXPECT().ListVpc(gomock.Any(), gomock.Any()).DoAndReturn(v.listVpc).MinTimes(1)
	mockCloud.EXPECT().CreateVpc(gomock.Any(), gomock.Any()).DoAndReturn(v.createVpc).MinTimes(1)

	// we do not support operation of update vpc, just check input type and then return nil, do nothing
	mockCloud.EXPECT().
		UpdateVpc(gomock.Any(), gomock.AssignableToTypeOf((*types.TCloudVpcUpdateOption)(nil))).
		Return(nil).MinTimes(1)

	mockCloud.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).DoAndReturn(v.deleteVpc).MinTimes(1)
}

// listVpc 如果输入参数为空，则返回全部数据，如果参数非空，则返回匹配参数
func (v *vpcPlaybook) listVpc(_ *kit.Kit, opt *core.TCloudListOption) (*types.TCloudVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	var found []types.TCloudVpc
	if len(opt.CloudIDs) == 0 {
		// return all that matches region
		found = v.vpcStore.Filter(func(vpc types.TCloudVpc) bool {
			return vpc.Region == opt.Region
		})
	} else {
		// 根据id 查找
		found = v.vpcStore.GetByKeys(opt.CloudIDs...)
	}
	return &types.TCloudVpcListResult{
		Count:   converter.ValToPtr(uint64(len(found))),
		Details: found,
	}, nil
}

// createVpc when creating vpc, the default route table will be created
func (v *vpcPlaybook) createVpc(_ *kit.Kit, opt *types.TCloudVpcCreateOption) (*types.TCloudVpc, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	cloudVpc := types.TCloudVpc{
		CloudID: rand.Prefix("vpc-", 8),
		Name:    opt.Name,
		Region:  opt.Extension.Region,
		Memo:    opt.Memo,
		Extension: &cloud.TCloudVpcExtension{
			Cidr:            []cloud.TCloudCidr{{Cidr: opt.Extension.IPv4Cidr, Type: enumor.Ipv4}},
			IsDefault:       false,
			EnableMulticast: false,
			DnsServerSet:    []string{"10.0.0.1", "183.60.82.98"},
			DomainName:      "aa.bb.cc",
		},
	}
	v.vpcStore.Add(cloudVpc.CloudID, cloudVpc)

	// 创建默认路由表
	routeTable := adtroutetable.TCloudRouteTable{
		CloudID:    rand.Prefix("rtb-", 8),
		Name:       "default",
		CloudVpcID: cloudVpc.CloudID,
		Region:     cloudVpc.Region,
		Memo:       converter.ValToPtr("default route table of " + cloudVpc.CloudID),
		Extension:  &adtroutetable.TCloudRouteTableExtension{Main: true},
	}
	v.routeTableStore.Add(routeTable.CloudID, routeTable)

	return &cloudVpc, nil
}

func (v *vpcPlaybook) deleteVpc(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}
	if _, exists := v.vpcStore.Get(opt.ResourceID); !exists {
		return errors.NewTencentCloudSDKError("NotFound", "not found vpc: "+opt.ResourceID, "xxx")
	}
	// 删除关联路由表
	tables := v.routeTableStore.Filter(func(table adtroutetable.TCloudRouteTable) bool {
		return table.CloudVpcID == opt.ResourceID
	})
	for _, table := range tables {
		err := v.routeTableStore.Remove(table.CloudID)
		if err != nil {
			return err
		}
	}

	// 删除关联子网
	subnets := v.subnetStore.Filter(func(n adtsubnet.TCloudSubnet) bool {
		return n.CloudVpcID == opt.ResourceID
	})
	for _, net := range subnets {
		err := v.deleteSubnet(kt, &core.BaseRegionalDeleteOption{Region: opt.Region,
			BaseDeleteOption: core.BaseDeleteOption{ResourceID: net.CloudID}})
		if err != nil {
			return err
		}
	}
	return v.vpcStore.Remove(opt.ResourceID)
}
