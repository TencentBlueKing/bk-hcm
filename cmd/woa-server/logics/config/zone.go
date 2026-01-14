/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"

	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/maps"
)

// ZoneIf provides management interface for operations of zone config
type ZoneIf interface {
	// GetZone get zone type config list
	GetZone(kt *kit.Kit, req *types.GetZoneParam) (*types.GetZoneResult, error)

	// GetIdcZone get idc zone type config list
	GetIdcZone(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetIdcZoneRst, error)
	// CreateIdcZone creates idc zone type config
	CreateIdcZone(kt *kit.Kit, input *types.IdcZone) (mapstr.MapStr, error)
}

// NewZoneOp creates a zone interface
func NewZoneOp(client *client.ClientSet) ZoneIf {
	return &zone{
		client: client,
	}
}

type zone struct {
	client *client.ClientSet
}

// buildZoneFilter 根据 GetZoneParam 构建 filter.Expression
func buildZoneFilter(req *types.GetZoneParam) *filter.Expression {
	rules := []filter.RuleFactory{
		tools.RuleEqual("vendor", enumor.TCloudZiyan),
	}

	if req == nil {
		return &filter.Expression{
			Op:    filter.And,
			Rules: rules,
		}
	}

	// 处理 Region 字段
	if len(req.Region) > 0 {
		rules = append(rules, tools.RuleIn("region", req.Region))
	}
	// 处理 Zone 字段
	if len(req.Zone) > 0 {
		rules = append(rules, tools.RuleIn("name", req.Zone))
	}

	return &filter.Expression{
		Op:    filter.And,
		Rules: rules,
	}
}

// GetZone get zone type config list
func (z *zone) GetZone(kt *kit.Kit, req *types.GetZoneParam) (*types.GetZoneResult, error) {
	// 从 data-service 查询 zone 列表
	filterExpr := buildZoneFilter(req)
	apiReq := &protocloud.ZoneListReq{
		Filter: filterExpr,
		Page:   core.NewCountPage(),
	}
	countRes, err := z.client.DataService().TCloudZiyan.Zone.ListZoneExt(kt, apiReq)
	if err != nil {
		return nil, fmt.Errorf("count zone failed, err: %v", err)
	}

	apiReq.Page = core.NewDefaultBasePage()
	allZones := make([]*types.Zone, 0)
	regionSet := make(map[string]struct{})
	for {
		apiZones, err := z.client.DataService().TCloudZiyan.Zone.ListZoneExt(kt, apiReq)
		if err != nil {
			return nil, fmt.Errorf("list zone failed, err: %v", err)
		}

		// 收集所有 region 值
		for _, item := range apiZones.Details {
			regionSet[item.Region] = struct{}{}
		}

		// 转换数据为 types.Zone
		for _, item := range apiZones.Details {
			zoneOne := &types.Zone{
				Zone:           item.Name,
				ZoneCn:         item.NameCn,
				Region:         item.Region,
				RegionCn:       "", // 稍后从 region 中获取
				CmdbRegionName: item.Extension.CityName,
				CmdbZoneName:   item.Extension.LogicCampusName,
			}
			allZones = append(allZones, zoneOne)
		}

		if uint(len(apiZones.Details)) < apiReq.Page.Limit {
			break
		}
		apiReq.Page.Start += uint32(apiReq.Page.Limit)
	}

	// 从 region API 查询 RegionCn
	regionMap, err := z.getRegionNameMap(kt, regionSet)
	if err != nil {
		logs.Errorf("failed to get region name, region: %+v, err: %v, rid: %s", regionSet, err, kt.Rid)
		return nil, err
	}
	// 填充 RegionCn，注意有部分zone没有对应的region（例如ap-mumbai）
	for _, zoneOne := range allZones {
		regionName, ok := regionMap[zoneOne.Region]
		if !ok {
			logs.Warnf("failed to get region name, region: %s, rid: %s", zoneOne.Region, kt.Rid)
			continue
		}
		zoneOne.RegionCn = regionName
	}

	rst := &types.GetZoneResult{
		Count: int64(countRes.Count),
		Info:  allZones,
	}
	return rst, nil
}

// getRegionNameMap 根据 region 集合查询 region name 映射
func (z *zone) getRegionNameMap(kt *kit.Kit, regionSet map[string]struct{}) (map[string]string, error) {
	if len(regionSet) == 0 {
		return make(map[string]string), nil
	}
	// 构建 region filter
	regionFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.RuleEqual("vendor", enumor.TCloudZiyan),
			tools.RuleIn("region_id", maps.Keys(regionSet)),
		},
	}

	// 查询 region 列表
	regionReq := &core.ListReq{
		Filter: regionFilter,
		Page:   core.NewDefaultBasePage(),
	}

	regionMap := make(map[string]string)
	for {
		apiRegions, err := z.client.DataService().TCloudZiyan.Region.ListRegion(kt, regionReq)
		if err != nil {
			return nil, fmt.Errorf("list region failed, err: %v", err)
		}

		// 构建 region_id -> region_name 映射
		for _, item := range apiRegions.Details {
			regionMap[item.RegionID] = item.RegionName
		}

		if uint(len(apiRegions.Details)) < regionReq.Page.Limit {
			break
		}
		regionReq.Page.Start += uint32(regionReq.Page.Limit)
	}

	return regionMap, nil
}

// GetIdcZone get idc zone type config list
func (z *zone) GetIdcZone(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetIdcZoneRst, error) {
	insts, err := config.Operation().IdcZone().FindManyZone(kt.Ctx, cond)
	if err != nil {
		return nil, err
	}

	rst := &types.GetIdcZoneRst{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// CreateIdcZone creates idc zone type config
func (z *zone) CreateIdcZone(kt *kit.Kit, input *types.IdcZone) (mapstr.MapStr, error) {
	id, err := config.Operation().IdcZone().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create idc zone, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().IdcZone().CreateZone(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create idc zone, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}
