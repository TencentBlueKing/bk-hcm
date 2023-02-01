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
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableregion "hcm/pkg/dal/table/cloud/region"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// InitAwsRegionService initialize the region service.
func InitAwsRegionService(cap *capability.Capability) {
	svc := &awsRegionSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateRegion", "POST", "/vendors/{vendor}/regions/batch/create", svc.BatchCreateRegion)
	h.Add("BatchUpdateRegion", "PATCH", "/vendors/{vendor}/regions/batch", svc.BatchDeleteRegion)
	h.Add("BatchUpdateRegionBaseInfo", "PATCH", "/regions/base/batch", svc.BatchUpdateRegionBaseInfo)
	h.Add("GetRegion", "GET", "/vendors/{vendor}/regions/{id}", svc.GetRegion)
	h.Add("ListRegion", "POST", "/regions/list", svc.ListVpc)
	h.Add("DeleteRegion", "DELETE", "/regions/batch", svc.BatchDeleteRegion)

	h.Load(cap.WebService)
}

type awsRegionSvc struct {
	dao dao.Set
}

// BatchCreateRegion batch create region.
func (svc *awsRegionSvc) BatchCreateRegion(cts *rest.Contexts) (interface{}, error) {
	req := new(protoregion.TCloudRegionCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regionIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		regions := make([]tableregion.AwsRegionTable, 0, len(req.Regions))
		for _, createReq := range req.Regions {
			tmpRegion := tableregion.AwsRegionTable{
				Vendor:      createReq.Vendor,
				RegionID:    createReq.RegionID,
				RegionName:  createReq.RegionName,
				IsAvailable: createReq.IsAvailable,
				Creator:     cts.Kit.User,
				Reviser:     cts.Kit.User,
			}
			regions = append(regions, tmpRegion)
		}

		regionID, err := svc.dao.AwsRegion().BatchCreateWithTx(cts.Kit, txn, regions)
		if err != nil {
			return nil, fmt.Errorf("create aws region failed, err: %v", err)
		}

		return regionID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := regionIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create aws region but return ids type %s is not string array",
			reflect.TypeOf(regionIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateRegion batch update region.
func (svc *awsRegionSvc) BatchUpdateRegion(cts *rest.Contexts) (interface{}, error) {
	req := new(protoregion.AwsRegionBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Regions))
	for _, region := range req.Regions {
		ids = append(ids, region.ID)
	}

	// check if all regions exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.AwsRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list region failed, err: %v", err)
	}
	if listRes.Count != uint64(len(req.Regions)) {
		return nil, fmt.Errorf("list region failed, some region(ids=%+v) doesn't exist", ids)
	}

	// update region
	tmpRegion := &tableregion.AwsRegionTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Regions {
		tmpRegion.Vendor = updateReq.Vendor
		tmpRegion.RegionID = updateReq.RegionID
		tmpRegion.RegionName = updateReq.RegionName
		tmpRegion.IsAvailable = updateReq.IsAvailable

		err = svc.dao.AwsRegion().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), tmpRegion)
		if err != nil {
			logs.Errorf("update region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update region failed, err: %v", err)
		}
	}
	return nil, nil
}

// BatchUpdateRegionBaseInfo batch update region base info.
func (svc *awsRegionSvc) BatchUpdateRegionBaseInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protoregion.AwsRegionBaseInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0)
	for _, region := range req.Regions {
		ids = append(ids, region.IDs...)
	}

	// check if all regions exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.TcloudRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list region failed, err: %v", err)
	}
	if listRes.Count != uint64(len(ids)) {
		return nil, fmt.Errorf("list region failed, some region(ids=%+v) doesn't exist", ids)
	}

	// update region
	tmpRegion := &tableregion.AwsRegionTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Regions {
		tmpRegion.Vendor = updateReq.Data.Vendor
		tmpRegion.RegionID = updateReq.Data.RegionID
		tmpRegion.RegionName = updateReq.Data.RegionName
		tmpRegion.IsAvailable = updateReq.Data.IsAvailable

		err = svc.dao.AwsRegion().Update(cts.Kit, tools.ContainersExpression("id", updateReq.IDs), tmpRegion)
		if err != nil {
			logs.Errorf("update region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update region failed, err: %v", err)
		}
	}

	return nil, nil
}

// GetRegion get region details.
func (svc *awsRegionSvc) GetRegion(cts *rest.Contexts) (interface{}, error) {
	regionID := cts.PathParameter("id").String()

	dbRegion, err := getAwsRegionFromTable(cts.Kit, svc.dao, regionID)
	if err != nil {
		return nil, err
	}

	base := convertAwsBaseRegion(dbRegion)
	return base, nil
}

func getAwsRegionFromTable(kt *kit.Kit, dao dao.Set, regionID string) (*tableregion.AwsRegionTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", regionID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.AwsRegion().List(kt, opt)
	if err != nil {
		logs.Errorf("list region failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list region failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list region failed, region(id=%s) doesn't exist", regionID)
	}

	return &details[0], nil
}

// ListVpc list regions.
func (svc *awsRegionSvc) ListVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	daoRegionResp, err := svc.dao.AwsRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list region failed, err: %v", err)
	}
	if req.Page.Count {
		return &protoregion.AwsRegionListResult{Count: daoRegionResp.Count}, nil
	}

	details := make([]protocore.AwsRegion, 0, len(daoRegionResp.Details))
	for _, region := range daoRegionResp.Details {
		details = append(details, converter.PtrToVal(convertAwsBaseRegion(&region)))
	}

	return &protoregion.AwsRegionListResult{Details: details}, nil
}

func convertAwsBaseRegion(dbRegion *tableregion.AwsRegionTable) *protocore.AwsRegion {
	if dbRegion == nil {
		return nil
	}

	return &protocore.AwsRegion{
		ID:          dbRegion.ID,
		Vendor:      dbRegion.Vendor,
		RegionID:    dbRegion.RegionID,
		RegionName:  dbRegion.RegionName,
		IsAvailable: dbRegion.IsAvailable,
		Creator:     dbRegion.Creator,
		Reviser:     dbRegion.Reviser,
		CreatedAt:   dbRegion.CreatedAt,
		UpdatedAt:   dbRegion.UpdatedAt,
	}
}

// BatchDeleteRegion batch delete regions.
func (svc *awsRegionSvc) BatchDeleteRegion(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.AwsRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list region failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delRegionIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delRegionIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delRegionFilter := tools.ContainersExpression("id", delRegionIDs)
		if err := svc.dao.AwsRegion().BatchDeleteWithTx(cts.Kit, txn, delRegionFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		logs.Errorf("delete region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
