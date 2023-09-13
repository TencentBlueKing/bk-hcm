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
	"hcm/pkg/adaptor/types/region"
	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/converter"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"go.uber.org/mock/gomock"
)

type regionPlaybook struct {
}

// Name region
func (r *regionPlaybook) Name() string {
	return string(enumor.RegionCloudResType)
}

// Apply to add list region method ability, always return same region list
func (r *regionPlaybook) Apply(mockCloud *MockTCloud, controller *gomock.Controller) {
	regionList := []region.TCloudRegion{
		{RegionID: "ap-mariana", RegionName: "太平洋西北(马里亚纳海沟)", RegionState: "AVAILABLE"},
		{RegionID: "ap-guangzhou", RegionName: "华南地区(广州)", RegionState: "AVAILABLE"},
	}

	zones := map[string][]typeszone.TCloudZone{
		"ap-mariana": {
			newZone("ap-mariana-1", "999991", "马里亚纳一区", "AVAILABLE"),
			newZone("ap-mariana-2", "999992", "马里亚纳二区", "AVAILABLE"),
			newZone("ap-mariana-3", "999993", "马里亚纳三区", "UNAVAILABLE"),
		},
		"ap-guangzhou": {
			newZone("ap-guangzhou-1", "100001", "广州一区", "UNAVAILABLE"),
			newZone("ap-guangzhou-2", "100002", "广州二区", "UNAVAILABLE"),
			newZone("ap-guangzhou-3", "100003", "广州三区", "AVAILABLE"),
			newZone("ap-guangzhou-4", "100004", "广州四区", "AVAILABLE"),
		},
	}
	// 不管传入什么都返回指定的regionList
	mockCloud.EXPECT().ListRegion(gomock.Any()).MinTimes(1).
		Return(&region.TCloudRegionListResult{
			Count:   converter.ValToPtr(uint64(len(regionList))),
			Details: regionList,
		}, nil)

	// 设定请求广州地域可用区的时候返回我们定义的可用区
	mockCloud.EXPECT().ListZone(gomock.Any(), &typeszone.TCloudZoneListOption{Region: "ap-guangzhou"}).
		MinTimes(1).Return(zones["ap-guangzhou"], nil)

	// 针对逻辑比较简单的无状态函数，直接返回通过参数匹配返回指定值即可，
	// EXPECT 方法返回相应的记录器(recorder)
	mockCloud.EXPECT().
		// 指定ListZone方法的第一个参数为任意值，第二个参数和给定参数相等的时候（底层会通过反射比较值）
		ListZone(gomock.Any(), &typeszone.TCloudZoneListOption{Region: "ap-mariana"}).
		// 可以调用任意次，也可以指定特定次数（最多几次、最少几次），默认是一次
		MinTimes(1).
		// 返回指定的值
		Return(zones["ap-mariana"], nil)

}

// NewRegionPlaybook  return region playbook, always returns the same region list
func NewRegionPlaybook() Playbook {
	return &regionPlaybook{}
}

func newZone(zone, id, name, status string) typeszone.TCloudZone {
	return typeszone.TCloudZone{
		ZoneInfo: &cvm.ZoneInfo{
			Zone:      &zone,
			ZoneId:    &id,
			ZoneName:  &name,
			ZoneState: &status,
		},
	}
}
