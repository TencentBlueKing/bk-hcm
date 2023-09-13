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
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"

	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types/core"
	adtroutetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"

	"go.uber.org/mock/gomock"
)

func (v *vpcPlaybook) applyRouteTable(mockCloud *MockTCloud) {

	mockCloud.EXPECT().ListRouteTable(gomock.Any(), gomock.Any()).DoAndReturn(v.listRouteTable).MinTimes(1)

	// only return nil, do nothing
	mockCloud.EXPECT().
		UpdateRouteTable(gomock.Any(), gomock.AssignableToTypeOf((*adtroutetable.TCloudRouteTableUpdateOption)(nil))).
		Return(nil).MinTimes(1)

	mockCloud.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
			err := opt.Validate()
			if err != nil {
				return err
			}
			return v.subnetStore.Remove(opt.ResourceID)
		}).MinTimes(1)
}

// listRouteTable 和其他list接口不同，路由表如果结果中没有指定的id，会返回未找到错误
func (v *vpcPlaybook) listRouteTable(_ *kit.Kit,
	opt *core.TCloudListOption) (*adtroutetable.TCloudRouteTableListResult,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}
	var values []adtroutetable.TCloudRouteTable

	if len(opt.CloudIDs) == 0 {
		values = v.routeTableStore.Filter(func(table adtroutetable.TCloudRouteTable) bool {
			return table.Region == opt.Region
		})
	} else {
		// 找不到要返回错误
		for _, id := range opt.CloudIDs {
			if rtb, ok := v.routeTableStore.Get(id); ok {
				values = append(values, rtb)
			} else {
				return nil, errors.NewTencentCloudSDKError(tcloud.ErrNotFound,
					"mock route table not found: "+rtb.CloudVpcID, "request-id")
			}
		}
	}

	return &adtroutetable.TCloudRouteTableListResult{
		Count:   converter.ValToPtr(uint64(len(values))),
		Details: values,
	}, nil

}
