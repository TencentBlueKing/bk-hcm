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

package region

import (
	typesregion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/api/core"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	apiregion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// SyncHuaWeiRegion sync all region
func (r *region) SyncHuaWeiRegion(cts *rest.Contexts) (interface{}, error) {

	req := new(apiregion.HuaWeiRegionSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := r.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	regions, err := client.ListRegion(cts.Kit)
	if err != nil {
		logs.Errorf("request adaptor to list huawei region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*typesregion.HuaWeiRegionModel)
	for _, region := range regions {
		cloudMap[region.Service+"_"+region.RegionID] = region
	}

	dsMap, err := r.getHuaWeiRegionAllDS(cts, req)
	if err != nil {
		logs.Errorf("request getHuaWeiRegionAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	addMap := make(map[string]*typesregion.HuaWeiRegionModel)
	updateMap := make(map[string]*typesregion.HuaWeiRegionModel)
	for k, v := range cloudMap {
		if _, ok := dsMap[k]; !ok {
			addMap[k] = v
		} else {
			updateMap[k] = v
		}
	}

	deleteMap := make(map[string]*typesregion.HuaWeiRegionModel)
	for k, v := range dsMap {
		if _, ok := cloudMap[k]; !ok {
			deleteMap[k] = v
		}
	}

	if len(deleteMap) > 0 {
		err := r.syncHuaWeiRegionDelete(cts, deleteMap)
		if err != nil {
			logs.Errorf("request syncHuaWeiRegionDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		err := r.syncHuaWeiRegionUpdate(cts, updateMap, dsMap)
		if err != nil {
			logs.Errorf("request syncHuaWeiRegionUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	if len(addMap) > 0 {
		err := r.syncHuaWeiRegionAdd(cts, addMap)
		if err != nil {
			logs.Errorf("request syncHuaWeiRegionAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (r *region) getHuaWeiRegionAllDS(cts *rest.Contexts,
	req *apiregion.HuaWeiRegionSyncReq) (map[string]*typesregion.HuaWeiRegionModel, error) {

	start := 0
	dsMap := make(map[string]*typesregion.HuaWeiRegionModel, 0)
	for {
		dataReq := &protoregion.HuaWeiRegionListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: constant.SyncTimingListHuaWeiRegion,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: core.DefaultMaxPageLimit},
		}

		results, err := r.dataCli.HuaWei.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				region := new(typesregion.HuaWeiRegionModel)
				region.ID = detail.ID
				region.RegionID = detail.RegionID
				region.Service = detail.Service
				region.Type = detail.Type
				dsMap[detail.Service+"_"+detail.RegionID] = region
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsMap, nil
}

func (r *region) syncHuaWeiRegionDelete(cts *rest.Contexts, deleteMap map[string]*typesregion.HuaWeiRegionModel) error {

	for _, v := range deleteMap {
		deleteReq := &protoregion.HuaWeiRegionBatchDeleteReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "service",
						Op:    filter.Equal.Factory(),
						Value: v.Service,
					},
					&filter.AtomRule{
						Field: "region_id",
						Op:    filter.Equal.Factory(),
						Value: v.RegionID,
					},
				},
			},
		}
		err := r.dataCli.HuaWei.Region.BatchDeleteRegion(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *region) syncHuaWeiRegionUpdate(cts *rest.Contexts, cloudMap map[string]*typesregion.HuaWeiRegionModel,
	dsMap map[string]*typesregion.HuaWeiRegionModel) error {

	list := make([]protoregion.HuaWeiRegionBatchUpdate, 0)
	for k, v := range cloudMap {
		if _, ok := dsMap[k]; ok {
			if v.Type == dsMap[k].Type {
				continue
			}

			one := protoregion.HuaWeiRegionBatchUpdate{
				ID:   dsMap[k].ID,
				Type: cloudMap[k].Type,
			}
			list = append(list, one)
		}
	}

	updateReq := &protoregion.HuaWeiRegionBatchUpdateReq{
		Regions: list,
	}

	if len(updateReq.Regions) > 0 {
		if err := r.dataCli.HuaWei.Region.BatchUpdateRegion(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateRegion failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}
	return nil
}

func (r *region) syncHuaWeiRegionAdd(cts *rest.Contexts, cloudMap map[string]*typesregion.HuaWeiRegionModel) error {

	list := make([]protoregion.HuaWeiRegionBatchCreate, 0)
	for _, v := range cloudMap {
		one := protoregion.HuaWeiRegionBatchCreate{
			RegionID: v.RegionID,
			Type:     v.Type,
			Service:  v.Service,
		}
		list = append(list, one)
	}

	createReq := &protoregion.HuaWeiRegionBatchCreateReq{
		Regions: list,
	}
	_, err := r.dataCli.HuaWei.Region.BatchCreateRegion(cts.Kit.Ctx, cts.Kit.Header(),
		createReq)
	if err != nil {
		return err
	}

	return nil
}
