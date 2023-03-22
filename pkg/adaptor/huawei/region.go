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

package huawei

import (
	"hcm/pkg/adaptor/types/region"
	"hcm/pkg/kit"

	dcsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dcs/v2/region"
	ecsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
	eipregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/region"
	imsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/region"
	vpcregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/region"
)

// ListRegion 查看地域
// reference: https://support.huaweicloud.com/api-iam/iam_05_0001.html
func (h *HuaWei) ListRegion(kt *kit.Kit) ([]*region.HuaWeiRegionModel, error) {
	// huawei region need by resource but we can use in public
	regions := make([]*region.HuaWeiRegionModel, 0)

	// ecs: cvm disk networkinterface
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_NORTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_NORTH_4.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_EAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_EAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_SOUTHWEST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.AP_SOUTHEAST_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.AP_SOUTHEAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.AP_SOUTHEAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.AF_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.SA_BRAZIL_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.LA_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_SOUTH_4.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.NA_MEXICO_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.LA_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_NORTH_9.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.CN_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ecs, ecsregion.AP_SOUTHEAST_4.Id))

	// vpc: vpc subnet sg sgRule routetable
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.AF_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_NORTH_4.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_NORTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_EAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_EAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_SOUTHWEST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.AP_SOUTHEAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_NORTH_9.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.AP_SOUTHEAST_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.AP_SOUTHEAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.SA_BRAZIL_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.LA_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.CN_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.LA_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.NA_MEXICO_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Vpc, vpcregion.AP_SOUTHEAST_4.Id))

	// eip: eip
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.AF_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_NORTH_4.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_NORTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_EAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_EAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_SOUTHWEST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.AP_SOUTHEAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.AP_SOUTHEAST_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.AP_SOUTHEAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_NORTH_9.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.LA_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.SA_BRAZIL_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.CN_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.LA_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.NA_MEXICO_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Eip, eipregion.AP_SOUTHEAST_4.Id))

	// ims: publicimage
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.AF_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_NORTH_4.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_NORTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_EAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_EAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_SOUTHWEST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.AP_SOUTHEAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.AP_SOUTHEAST_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.AP_SOUTHEAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.CN_NORTH_9.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.LA_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.SA_BRAZIL_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.LA_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.NA_MEXICO_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Ims, imsregion.AP_SOUTHEAST_4.Id))

	// dcs: zone
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.AF_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_NORTH_4.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_NORTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_EAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_EAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_SOUTH_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_SOUTHWEST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.AP_SOUTHEAST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.AP_SOUTHEAST_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.AP_SOUTHEAST_3.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.RU_NORTHWEST_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.SA_BRAZIL_1.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.LA_NORTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.LA_SOUTH_2.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.CN_NORTH_9.Id))
	regions = append(regions, getHuaWeiModelRegion(Dcs, dcsregion.NA_MEXICO_1.Id))

	return regions, nil
}

func getHuaWeiModelRegion(service string, regionId string) *region.HuaWeiRegionModel {
	region := &region.HuaWeiRegionModel{
		Service:  service,
		RegionID: regionId,
		Type:     "public",
	}
	return region
}
