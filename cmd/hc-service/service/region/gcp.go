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

	adcore "hcm/pkg/adaptor/types/core"
	typesRegion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	protoDsRegion "hcm/pkg/api/data-service/cloud/region"
	protoHcRegion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// GcpSyncRegion gcp sync region.
func (r region) GcpSyncRegion(cts *rest.Contexts, vendor enumor.Vendor) error {
	req := new(protoHcRegion.GcpRegionSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	// batch get region list from cloudapi.
	list, err := r.BatchGetGcpRegionList(cts, req)
	if err != nil {
		logs.Errorf("%s-region request cloudapi response failed. accountID: %s, err: %v",
			enumor.Gcp, req.AccountID, err)
		return err
	}

	resourceDBMap, err := r.BatchGetGcpRegionMapFromDB(cts, req, vendor)
	if err != nil {
		logs.Errorf("%s-region batch get vpcdblist failed. accountID: %s, err: %v",
			enumor.Gcp, req.AccountID, err)
		return err
	}

	err = r.BatchSyncGcpRegionList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("%s-region compare api and dblist failed. accountID: %s, err: %v",
			enumor.Gcp, req.AccountID, err)
		return err
	}

	logs.Infof("%s-region region sync success. accountID: %s", enumor.Gcp, req.AccountID)

	return nil
}

// BatchGetGcpRegionList batch get region list from cloudapi.
func (r region) BatchGetGcpRegionList(cts *rest.Contexts, req *protoHcRegion.GcpRegionSyncReq) (
	*typesRegion.GcpRegionListResult, error) {

	cli, err := r.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("batch get gcp region list client failed, accountID: %s, err: %v, rid: %s",
			req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	nextToken := ""
	list := new(typesRegion.GcpRegionListResult)
	for {
		opt := new(adcore.GcpListOption)
		opt.Page = &adcore.GcpPage{
			PageSize: int64(adcore.GcpQueryLimit),
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		tmpList, tmpErr := cli.ListRegion(cts.Kit, opt)
		if tmpErr != nil {
			logs.Errorf("%s-region batch get cloud api failed. accountID: %s, nextToken: %s, err: %v",
				enumor.Gcp, req.AccountID, nextToken, tmpErr)
			return nil, tmpErr
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(tmpList.NextPageToken) == 0 {
			break
		}

		nextToken = tmpList.NextPageToken
	}

	return list, nil
}

// BatchGetGcpRegionMapFromDB batch get region map from db.
func (r region) BatchGetGcpRegionMapFromDB(cts *rest.Contexts, req *protoHcRegion.GcpRegionSyncReq,
	vendor enumor.Vendor) (map[string]cloudcore.GcpRegion, error) {

	page := uint32(0)
	resourceMap := make(map[string]cloudcore.GcpRegion, 0)
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
			},
		}
		dbQueryReq := &protoDsRegion.GcpRegionListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := r.cs.DataService().Gcp.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-region batch get regionlist db error. accountID: %s, offset: %d, "+
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

// BatchSyncGcpRegionList batch sync vendor region list.
func (r region) BatchSyncGcpRegionList(cts *rest.Contexts, req *protoHcRegion.GcpRegionSyncReq,
	list *typesRegion.GcpRegionListResult, resourceDBMap map[string]cloudcore.GcpRegion) error {
	createResources, updateResources, existIDMap, err := r.filterGcpRegionList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &protoDsRegion.GcpRegionBatchUpdateReq{
			Regions: updateResources,
		}
		if err = r.cs.DataService().Gcp.Region.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("%s-region batch compare db update failed. accountID: %s, err: %v",
				enumor.Gcp, req.AccountID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &protoDsRegion.GcpRegionCreateReq{
			Regions: createResources,
		}
		if _, err = r.cs.DataService().Gcp.Region.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("%s-region batch compare db create failed. accountID: %s, err: %v",
				enumor.Gcp, req.AccountID, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	if len(existIDMap) > 0 {
		for _, resourceItem := range resourceDBMap {
			if _, ok := existIDMap[resourceItem.RegionID]; !ok {
				deleteIDs = append(deleteIDs, resourceItem.ID)
			}
		}
	}

	if len(deleteIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err := r.cs.DataService().Gcp.Region.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
			return err
		}
		if err != nil {
			logs.Errorf("%s-region batch compare db delete failed. accountID: %s, deleteIDs: %v, "+
				"err: %v", enumor.Gcp, req.AccountID, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterGcpRegionList filter gcp region list
func (r region) filterGcpRegionList(req *protoHcRegion.GcpRegionSyncReq,
	list *typesRegion.GcpRegionListResult, resourceDBMap map[string]cloudcore.GcpRegion) (
	createResources []protoDsRegion.GcpRegionBatchCreate, updateResources []protoDsRegion.GcpRegionBatchUpdate,
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
				existIDMap[item.RegionID] = true
				continue
			}

			tmpRes := protoDsRegion.GcpRegionBatchUpdate{
				ID:         resourceInfo.ID,
				RegionID:   item.RegionID,
				RegionName: item.RegionName,
				Status:     item.RegionState,
				SelfLink:   item.SelfLink,
			}
			updateResources = append(updateResources, tmpRes)
			existIDMap[item.RegionID] = true
		} else {
			// need add resource data
			tmpRes := protoDsRegion.GcpRegionBatchCreate{
				Vendor:     enumor.Gcp,
				RegionID:   item.RegionID,
				RegionName: item.RegionName,
				Status:     item.RegionState,
				SelfLink:   item.SelfLink,
			}
			createResources = append(createResources, tmpRes)
			existIDMap[resourceInfo.RegionID] = true
		}
	}

	return createResources, updateResources, existIDMap, nil
}
