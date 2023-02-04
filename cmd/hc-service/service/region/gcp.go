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
	adcore "hcm/pkg/adaptor/types/core"
	typesRegion "hcm/pkg/adaptor/types/region"
	dataservice "hcm/pkg/api/data-service"
	protoDsRegion "hcm/pkg/api/data-service/cloud/region"
	protoHcRegion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GcpSyncRegion gcp sync region.
func (r region) GcpSyncRegion(cts *rest.Contexts, vendor enumor.Vendor) (interface{}, error) {
	cloudResp, err := r.BatchGetGcpRegionList(cts)
	if err != nil {
		logs.Errorf("get gcp region list failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	tmpRegions := make([]protoDsRegion.GcpRegionBatchCreate, 0)
	for _, item := range cloudResp.Details {
		tmpRegions = append(tmpRegions, protoDsRegion.GcpRegionBatchCreate{
			Vendor:     vendor,
			RegionID:   item.RegionID,
			RegionName: item.RegionName,
			Status:     item.RegionState,
			SelfLink:   item.SelfLink,
		})
	}

	if len(tmpRegions) == 0 {
		return nil, errf.New(errf.RecordNotFound, "cloudapi has not available region")
	}

	// batch forbidden gcp region state.
	updateStateReq := &protoDsRegion.GcpRegionBatchUpdateReq{
		Regions: []protoDsRegion.GcpRegionBatchUpdate{{Status: constant.GcpStateDisable}},
	}
	err = r.cs.DataService().Gcp.Region.BatchForbiddenRegionState(cts.Kit.Ctx, cts.Kit.Header(), updateStateReq)
	if err != nil {
		logs.Errorf("batch forbidden gcp region state failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// batch create gcp region.
	createReq := &protoDsRegion.GcpRegionCreateReq{
		Regions: tmpRegions,
	}
	resp, err := r.cs.DataService().Gcp.Region.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("batch create gcp region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// batch delete gcp region.
	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("status", constant.GcpStateDisable),
	}
	err = r.cs.DataService().Gcp.Region.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		logs.Errorf("batch delete gcp region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return resp, nil
}

// BatchGetGcpRegionList batch get region list from cloudapi.
func (r region) BatchGetGcpRegionList(cts *rest.Contexts) (*typesRegion.GcpRegionListResult, error) {
	req := new(protoHcRegion.GcpRegionSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

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
			logs.Errorf("[%s-region]batch get cloud api failed. accountID: %s, nextToken: %s, err: %v",
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
