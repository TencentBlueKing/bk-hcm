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

	"hcm/pkg/adaptor/types/region"
	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/uuid"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"go.uber.org/mock/gomock"
)

var regionList = []region.TCloudRegion{
	{RegionID: "ap-mariana", RegionName: "太平洋西北(马里亚纳海沟)", RegionState: "AVAILABLE"},
	{RegionID: "ap-guangzhou", RegionName: "华南地区(广州)", RegionState: "AVAILABLE"},
}

var zoneMap = map[string][]typeszone.TCloudZone{
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

type regionPlaybook struct {
	regionList []region.TCloudRegion
	zoneMap    map[string][]typeszone.TCloudZone
}

// Name region
func (r *regionPlaybook) Name() string {
	return string(enumor.RegionCloudResType)
}

// Apply to add list region method ability, always return same region list
func (r *regionPlaybook) Apply(mockCloud *MockTCloud, controller *gomock.Controller) {

	// 不管传入什么都返回指定的regionList,最少要调用一次
	mockCloud.EXPECT().ListRegion(gomock.Any()).MinTimes(1).
		Return(&region.TCloudRegionListResult{
			Count:   converter.ValToPtr(uint64(len(r.regionList))),
			Details: r.regionList,
		}, nil)

	// 设定请求广州地域可用区的时候返回自定义的可用区
	mockCloud.EXPECT().
		// 指定ListZone方法的第一个参数为任意值，第二个参数和给定参数相等的时候
		ListZone(gomock.Any(), &typeszone.TCloudZoneListOption{Region: "ap-guangzhou"}).
		// 最少调用一次
		MinTimes(1).
		Return(r.zoneMap["ap-guangzhou"], nil)

	mockCloud.EXPECT().
		ListZone(gomock.Any(), &typeszone.TCloudZoneListOption{Region: "ap-mariana"}).
		MinTimes(1).
		Return(r.zoneMap["ap-mariana"], nil)

}

// NewRegionPlaybook  return region playbook, always returns the same region list
func NewRegionPlaybook() Playbook {
	return &regionPlaybook{regionList: regionList, zoneMap: zoneMap}
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

// RegionValidate ...
func RegionValidate(regionID string) error {
	regions := slice.Filter(regionList, func(r region.TCloudRegion) bool { return r.RegionID == regionID })

	if len(regions) != 1 {
		return errors.NewTencentCloudSDKError("InvalidParameterValue",
			"参数 `X-TC-Region` 取值错误。", uuid.UUID())
	}
	if regions[0].RegionState == "AVAILABLE" {
		return errors.NewTencentCloudSDKError("InvalidParameterValue",
			"region 不可用。", uuid.UUID())
	}
	return nil
}

// ZoneValidate ...
func ZoneValidate(regionID, zoneStr string) error {

	err := RegionValidate(regionID)
	if err != nil {
		return err
	}
	zones := slice.Filter(zoneMap[regionID], func(z typeszone.TCloudZone) bool {
		return assert.IsPtrStringEqual(z.ZoneId, &regionID)
	})
	if len(zones) != 1 {
		return errors.NewTencentCloudSDKError("InvalidParameterValue",
			"zone 错误", uuid.UUID())
	}
	if converter.PtrToVal(zones[0].ZoneState) == "AVAILABLE" {
		return errors.NewTencentCloudSDKError("InvalidParameterValue",
			"zone 不可用。", uuid.UUID())
	}
	return nil
}
