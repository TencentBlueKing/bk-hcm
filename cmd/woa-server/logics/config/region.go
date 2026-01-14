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
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
)

// RegionIf provides management interface for operations of region config
type RegionIf interface {
	// GetRegion get region type config list
	GetRegion(kt *kit.Kit) (*types.GetRegionResult, error)
	// GetIdcRegion get region type config list
	GetIdcRegion(kt *kit.Kit) (*types.GetIdcRegionRst, error)
}

// NewRegionOp creates a region interface
func NewRegionOp(client *client.ClientSet) RegionIf {
	return &region{
		client: client,
	}
}

type region struct {
	client *client.ClientSet
}

// GetRegion get region type config list
func (r *region) GetRegion(kt *kit.Kit) (*types.GetRegionResult, error) {
	// 从 data-service 查询 region 列表
	req := &core.ListReq{
		Filter: tools.EqualExpression("vendor", enumor.TCloudZiyan),
		Page:   core.NewCountPage(),
	}
	countRes, err := r.client.DataService().TCloudZiyan.Region.ListRegion(kt, req)
	if err != nil {
		return nil, fmt.Errorf("list region count failed, err: %v", err)
	}

	req.Page = core.NewDefaultBasePage()
	allRegions := make([]*types.Region, 0)
	for {
		apiRegions, err := r.client.DataService().TCloudZiyan.Region.ListRegion(kt, req)
		if err != nil {
			return nil, fmt.Errorf("list region failed, err: %v", err)
		}

		// 转换数据为 types.Region 类型
		for _, item := range apiRegions.Details {
			regionOne := &types.Region{
				Region:         item.RegionID,
				RegionCn:       item.RegionName,
				CmdbRegionName: item.CityName,
			}
			allRegions = append(allRegions, regionOne)
		}

		if uint(len(apiRegions.Details)) < req.Page.Limit {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	rst := &types.GetRegionResult{
		Count: int64(countRes.Count),
		Info:  allRegions,
	}

	return rst, nil
}

// GetIdcRegion get idc region list
func (r *region) GetIdcRegion(kt *kit.Kit) (*types.GetIdcRegionRst, error) {
	filter := make(map[string]interface{})

	// TODO 替换mysql，依然使用静态数据
	insts, err := config.Operation().IdcZone().GetRegionList(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetIdcRegionRst{
		Info: insts,
	}

	return rst, nil
}
