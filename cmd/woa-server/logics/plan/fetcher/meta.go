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

package fetcher

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	dmtypes "hcm/pkg/dal/dao/types/meta"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetZoneMap get zone id name mapping from data-service.
func (f *ResPlanFetcher) GetZoneMap(kt *kit.Kit) (map[string]string, error) {
	req := &protocloud.ZoneListReq{
		Filter: tools.EqualExpression("vendor", enumor.TCloudZiyan),
		Page:   core.NewDefaultBasePage(),
	}

	zoneMap := make(map[string]string)
	for {
		zones, err := f.client.DataService().TCloudZiyan.Zone.ListZoneExt(kt, req)
		if err != nil {
			return nil, fmt.Errorf("list zone failed, err: %v", err)
		}

		// 排除 disable_cvm 为 true 的 zone（disable_cvm 为空视为 false，即不禁用）
		for _, zone := range zones.Details {
			// 如果 DisableCvm 显式为 true，则跳过该 zone（已禁用）
			if zone.Extension != nil && zone.Extension.DisableCvm {
				continue
			}
			zoneMap[zone.Name] = zone.NameCn
		}

		if len(zones.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return zoneMap, nil
}

// GetRegionAreaMap get region area mapping from data-service.
// region_name 从 region.city_name 获取
// area_name 从 region.area_name 获取
func (f *ResPlanFetcher) GetRegionAreaMap(kt *kit.Kit) (map[string]dmtypes.RegionArea, error) {
	// 查询 region 表获取 region_name 和 area_name
	req := &core.ListReq{
		Filter: tools.EqualExpression("vendor", enumor.TCloudZiyan),
		Page:   core.NewDefaultBasePage(),
	}

	regionAreaMap := make(map[string]dmtypes.RegionArea)
	for {
		regions, err := f.client.DataService().TCloudZiyan.Region.ListRegion(kt, req)
		if err != nil {
			return nil, fmt.Errorf("list region from data-service failed, err: %v", err)
		}

		for _, region := range regions.Details {
			regionAreaMap[region.RegionID] = dmtypes.RegionArea{
				RegionID:   region.RegionID,
				RegionName: region.CityName,
				AreaName:   region.AreaName,
			}
		}

		if uint(len(regions.Details)) < req.Page.Limit {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return regionAreaMap, nil
}

// GetMetaMaps get create resource plan demand needed zoneMap, regionAreaMap and deviceTypeMap.
func (f *ResPlanFetcher) GetMetaMaps(kt *kit.Kit) (map[string]string, map[string]dmtypes.RegionArea,
	map[string]wdt.WoaDeviceTypeTable, error) {

	// get zone id name mapping from data-service.
	zoneMap, err := f.GetZoneMap(kt)
	if err != nil {
		logs.Errorf("get zone map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get region area mapping from data-service.
	regionAreaMap, err := f.GetRegionAreaMap(kt)
	if err != nil {
		logs.Errorf("get region area map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	// get device type mapping.
	deviceTypeMap, err := f.deviceTypesMap.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	return zoneMap, regionAreaMap, deviceTypeMap, nil
}

// GetMetaNameMapsFromIDMap get zone name map and region name map from id map.
func (f *ResPlanFetcher) GetMetaNameMapsFromIDMap(zoneMap map[string]string,
	regionAreaMap map[string]dmtypes.RegionArea) (
	map[string]string, map[string]dmtypes.RegionArea) {

	zoneNameMap := make(map[string]string)
	for id, name := range zoneMap {
		zoneNameMap[name] = id
	}
	regionNameMap := make(map[string]dmtypes.RegionArea)
	for _, item := range regionAreaMap {
		regionNameMap[item.RegionName] = item
	}
	return zoneNameMap, regionNameMap
}
