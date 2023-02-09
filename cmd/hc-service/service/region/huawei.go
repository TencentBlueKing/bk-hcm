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
	"hcm/pkg/api/core"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	apiregion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
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

	cloudAllIDs := make(map[string]bool)

	cloudMap := make(map[string]*HuaWeiRegionSync)
	cloudIDs := make([]string, 0, len(regions))
	for _, data := range regions {
		regionSync := new(HuaWeiRegionSync)
		regionSync.IsUpdate = false
		regionSync.Region = data
		cloudMap[data.Id] = regionSync
		cloudIDs = append(cloudIDs, data.Id)
		cloudAllIDs[data.Id] = true
	}

	updateIDs, dsMap, err := r.getHuaWeiRegionDSSync(cts, cloudIDs, req)
	if err != nil {
		logs.Errorf("request getHuaWeiRegionDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(updateIDs) > 0 {
		err := r.syncHuaWeiRegionUpdate(cts, updateIDs, cloudMap, dsMap)
		if err != nil {
			logs.Errorf("request syncHuaWeiRegionUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	addIDs := make([]string, 0)
	for _, id := range updateIDs {
		if _, ok := cloudMap[id]; ok {
			cloudMap[id].IsUpdate = true
		}
	}
	for k, v := range cloudMap {
		if !v.IsUpdate {
			addIDs = append(addIDs, k)
		}
	}

	if len(addIDs) > 0 {
		err := r.syncHuaWeiRegionAdd(cts, addIDs, req, cloudMap)
		if err != nil {
			logs.Errorf("request syncHuaWeiRegionAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := r.getHuaWeiRegionAllDS(cts, req)
	if err != nil {
		logs.Errorf("request getHuaWeiRegionAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		err := r.syncRegionDelete(cts, deleteIDs)
		if err != nil {
			logs.Errorf("request syncRegionDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (r *region) syncHuaWeiRegionUpdate(cts *rest.Contexts, updateIDs []string, cloudMap map[string]*HuaWeiRegionSync,
	dsMap map[string]*HuaWeiDSRegionSync) error {

	list := make([]protoregion.HuaWeiRegionBatchUpdate, 0, len(updateIDs))
	for _, id := range updateIDs {
		if cloudMap[id].Region.Type == dsMap[id].Region.Type &&
			cloudMap[id].Region.Locales.ZhCn == dsMap[id].Region.LocalesZhCn {
			continue
		}

		one := protoregion.HuaWeiRegionBatchUpdate{
			ID:          dsMap[id].Region.ID,
			Type:        cloudMap[id].Region.Type,
			LocalesZhCn: cloudMap[id].Region.Locales.ZhCn,
		}
		list = append(list, one)
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

func (r *region) syncHuaWeiRegionAdd(cts *rest.Contexts, addIDs []string, req *apiregion.HuaWeiRegionSyncReq,
	cloudMap map[string]*HuaWeiRegionSync) error {

	list := make([]protoregion.HuaWeiRegionBatchCreate, 0, len(addIDs))
	for _, id := range addIDs {
		one := protoregion.HuaWeiRegionBatchCreate{
			RegionID:    cloudMap[id].Region.Id,
			Type:        cloudMap[id].Region.Type,
			LocalesZhCn: cloudMap[id].Region.Locales.ZhCn,
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

func (r *region) syncRegionDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	deleteReq := &protoregion.HuaWeiRegionBatchDeleteReq{
		Filter: tools.ContainersExpression("region_id", deleteCloudIDs),
	}

	err := r.dataCli.HuaWei.Region.BatchDeleteRegion(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return err
	}

	return nil
}

func (r *region) getHuaWeiRegionDSSync(cts *rest.Contexts, cloudIDs []string,
	req *apiregion.HuaWeiRegionSyncReq) ([]string, map[string]*HuaWeiDSRegionSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*HuaWeiDSRegionSync)

	start := 0
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
					&filter.AtomRule{
						Field: "region_id",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
				},
			},
			Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		}

		results, err := r.dataCli.HuaWei.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.RegionID)
				dsRegionSync := new(HuaWeiDSRegionSync)
				dsRegionSync.Region = detail
				dsMap[detail.RegionID] = dsRegionSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return updateIDs, dsMap, nil
}

func (r *region) getHuaWeiRegionAllDS(cts *rest.Contexts, req *apiregion.HuaWeiRegionSyncReq) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
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
			Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		}

		results, err := r.dataCli.HuaWei.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.RegionID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsIDs, nil
}
