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

// Package region defines region service.
package region

import (
	"fmt"

	typesRegion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	protoDsRegion "hcm/pkg/api/data-service/cloud/region"
	protoHcRegion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// TCloudSyncRegion tcloud sync region.
func (r region) TCloudSyncRegion(cts *rest.Contexts, vendor enumor.Vendor) error {
	req := new(protoHcRegion.TCloudRegionSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	// batch get region list from cloudapi.
	list, err := r.BatchGetTCloudRegionList(cts, req)
	if err != nil {
		logs.Errorf("[%s-region] request cloudapi response failed. accountID: %s, err: %v",
			enumor.TCloud, req.AccountID, err)
		return err
	}

	resourceDBMap, err := r.BatchGetTCloudRegionMapFromDB(cts, req, vendor)
	if err != nil {
		logs.Errorf("[%s-region] batch get regiondblist failed. accountID: %s, err: %v",
			enumor.TCloud, req.AccountID, err)
		return err
	}

	err = r.BatchSyncTCloudRegionList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-region] compare api and dblist failed. accountID: %s, err: %v",
			enumor.TCloud, req.AccountID, err)
		return err
	}

	return nil
}

// BatchGetTCloudRegionList batch get region list from cloudapi.
func (r region) BatchGetTCloudRegionList(cts *rest.Contexts, req *protoHcRegion.TCloudRegionSyncReq) (
	*typesRegion.TCloudRegionListResult, error) {

	cli, err := r.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudResp, err := cli.ListRegion(cts.Kit)
	if err != nil {
		logs.Errorf("get tcloud region list failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(cloudResp.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "cloudapi has not available region")
	}

	return &typesRegion.TCloudRegionListResult{
		Details: cloudResp.Details,
	}, nil
}

// BatchGetTCloudRegionMapFromDB batch get region map from db.
func (r region) BatchGetTCloudRegionMapFromDB(cts *rest.Contexts, req *protoHcRegion.TCloudRegionSyncReq,
	vendor enumor.Vendor) (map[string]cloudcore.TCloudRegion, error) {

	page := uint32(0)
	resourceMap := make(map[string]cloudcore.TCloudRegion, 0)
	for {
		count := core.DefaultMaxPageLimit
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
				&filter.AtomRule{
					Field: "status",
					Op:    filter.Equal.Factory(),
					Value: constant.TCloudStateEnable,
				},
			},
		}
		dbQueryReq := &protoDsRegion.TCloudRegionListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := r.cs.DataService().TCloud.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("[%s-region]batch get regionlist db error. accountID: %s, offset: %d, "+
				"limit: %d, err: %v", vendor, req.AccountID, offset, count, err)
			return nil, err
		}

		if len(dbList.Details) == 0 {
			return resourceMap, nil
		}

		for _, item := range dbList.Details {
			resourceMap[item.RegionID] = item
		}

		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}

	return resourceMap, nil
}

// BatchSyncTCloudRegionList batch sync vendor region list.
func (r region) BatchSyncTCloudRegionList(cts *rest.Contexts, req *protoHcRegion.TCloudRegionSyncReq,
	list *typesRegion.TCloudRegionListResult, resourceDBMap map[string]cloudcore.TCloudRegion) error {
	createResources, updateResources, existIDMap, err := r.filterTCloudRegionList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &protoDsRegion.TCloudRegionBatchUpdateReq{
			Regions: updateResources,
		}
		if err = r.cs.DataService().TCloud.Region.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("[%s-region]batch compare db update failed. accountID: %s, err: %v",
				enumor.TCloud, req.AccountID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &protoDsRegion.TCloudRegionCreateReq{
			Regions: createResources,
		}
		if _, err = r.cs.DataService().TCloud.Region.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("[%s-region]batch compare db create failed. accountID: %s, err: %v",
				enumor.TCloud, req.AccountID, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resourceItem := range resourceDBMap {
		if _, ok := existIDMap[resourceItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resourceItem.ID)
		}
	}

	if len(deleteIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err := r.cs.DataService().TCloud.Region.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
			return err
		}
		if err != nil {
			logs.Errorf("[%s-region]batch compare db delete failed. accountID: %s, deleteIDs: %v, "+
				"err: %v", enumor.TCloud, req.AccountID, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterTCloudRegionList filter tcloud region list
func (r region) filterTCloudRegionList(req *protoHcRegion.TCloudRegionSyncReq,
	list *typesRegion.TCloudRegionListResult, resourceDBMap map[string]cloudcore.TCloudRegion) (
	createResources []protoDsRegion.TCloudRegionBatchCreate, updateResources []protoDsRegion.TCloudRegionBatchUpdate,
	existIDMap map[string]bool, err error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi regionlist is empty, accountID: %s", req.AccountID)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.RegionID]; ok {
			if resourceInfo.RegionID == item.RegionID && resourceInfo.RegionName == item.RegionName &&
				resourceInfo.Status == item.RegionState {
				continue
			}

			tmpRes := protoDsRegion.TCloudRegionBatchUpdate{
				ID:         resourceInfo.ID,
				RegionID:   item.RegionID,
				RegionName: item.RegionName,
				Status:     item.RegionState,
			}
			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add resource data
			tmpRes := protoDsRegion.TCloudRegionBatchCreate{
				Vendor:     enumor.TCloud,
				RegionID:   item.RegionID,
				RegionName: item.RegionName,
				Status:     item.RegionState,
			}
			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}
