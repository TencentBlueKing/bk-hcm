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
	mockCloud.EXPECT().ListZone(gomock.Any(), gomock.Eq(&typeszone.TCloudZoneListOption{Region: "ap-mariana"})).
		AnyTimes().Return(zones["ap-mariana"], nil)

	mockCloud.EXPECT().ListZone(gomock.Any(), gomock.Eq(&typeszone.TCloudZoneListOption{Region: "ap-guangzhou"})).
		AnyTimes().Return(zones["ap-guangzhou"], nil)

	mockCloud.EXPECT().ListRegion(gomock.Any()).AnyTimes().
		Return(&region.TCloudRegionListResult{
			Count:   converter.ValToPtr(uint64(len(regionList))),
			Details: regionList,
		}, nil)
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
