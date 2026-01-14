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

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud/region"
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

// BatchCreateTCloudZiyanRegion batch create region.
func (svc *regionSvc) BatchCreateTCloudZiyanRegion(cts *rest.Contexts) (interface{}, error) {
	req := new(protoregion.TCloudRegionCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regionIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		regions := make([]tableregion.TCloudZiyanRegionTable, 0, len(req.Regions))
		for _, createReq := range req.Regions {
			tmpRegion := tableregion.TCloudZiyanRegionTable{
				Vendor:     createReq.Vendor,
				RegionID:   createReq.RegionID,
				RegionName: createReq.RegionName,
				AreaName:   createReq.AreaName,
				CityName:   createReq.CityName,
				Status:     createReq.Status,
				Source:     createReq.Source,
				Creator:    cts.Kit.User,
				Reviser:    cts.Kit.User,
			}
			regions = append(regions, tmpRegion)
		}

		regionID, err := svc.dao.TCloudZiyanRegion().BatchCreateWithTx(cts.Kit, txn, regions)
		if err != nil {
			return nil, fmt.Errorf("create tcloud ziyan region failed, err: %v", err)
		}

		return regionID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := regionIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create tcloud ziyan region but return ids type %s is not string array",
			reflect.TypeOf(regionIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateTCloudZiyanRegion batch update region.
func (svc *regionSvc) BatchUpdateTCloudZiyanRegion(cts *rest.Contexts) error {
	req := new(protoregion.TCloudRegionBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
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

	listRes, err := svc.dao.TCloudZiyanRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud ziyan region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return fmt.Errorf("list tcloud ziyan region failed, err: %v", err)
	}

	if listRes.Count != uint64(len(req.Regions)) {
		return fmt.Errorf("list tcloud ziyan region failed, some region(ids=%+v) doesn't exist", ids)
	}

	// update region
	tmpRegion := &tableregion.TCloudZiyanRegionTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Regions {
		tmpRegion.RegionID = updateReq.RegionID
		tmpRegion.RegionName = updateReq.RegionName
		tmpRegion.AreaName = updateReq.AreaName
		tmpRegion.CityName = updateReq.CityName
		tmpRegion.Status = updateReq.Status
		if len(updateReq.Source) > 0 {
			tmpRegion.Source = updateReq.Source
		}

		err = svc.dao.TCloudZiyanRegion().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), tmpRegion)
		if err != nil {
			logs.Errorf("update tcloud ziyan region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return fmt.Errorf("update tcloud ziyan region failed, err: %v", err)
		}
	}

	return nil
}

// GetTCloudZiyanRegion get region details.
func (svc *regionSvc) GetTCloudZiyanRegion(cts *rest.Contexts) (interface{}, error) {
	regionID := cts.PathParameter("id").String()

	dbRegion, err := getTCloudZiyanRegionFromTable(cts.Kit, svc.dao, regionID)
	if err != nil {
		return nil, err
	}

	base := convertTCloudZiyanBaseRegion(dbRegion)
	return base, nil
}

func getTCloudZiyanRegionFromTable(kt *kit.Kit, dao dao.Set, regionID string) (*tableregion.TCloudZiyanRegionTable,
	error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", regionID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.TCloudZiyanRegion().List(kt, opt)
	if err != nil {
		logs.Errorf("list tcloud ziyan region failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list tcloud ziyan region failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list tcloud ziyan region failed, region(id=%s) doesn't exist", regionID)
	}

	return &details[0], nil
}

// ListTCloudZiyanRegion list regions.
func (svc *regionSvc) ListTCloudZiyanRegion(cts *rest.Contexts) (interface{}, error) {
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
	daoRegionResp, err := svc.dao.TCloudZiyanRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud ziyan region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud ziyan region failed, err: %v", err)
	}
	if req.Page.Count {
		return &protoregion.TCloudRegionListResult{Count: daoRegionResp.Count}, nil
	}

	details := make([]protocore.TCloudRegion, 0, len(daoRegionResp.Details))
	for _, region := range daoRegionResp.Details {
		details = append(details, converter.PtrToVal(convertTCloudZiyanBaseRegion(&region)))
	}

	return &protoregion.TCloudRegionListResult{Details: details}, nil
}

func convertTCloudZiyanBaseRegion(dbRegion *tableregion.TCloudZiyanRegionTable) *protocore.TCloudRegion {
	if dbRegion == nil {
		return nil
	}

	return &protocore.TCloudRegion{
		ID:         dbRegion.ID,
		Vendor:     dbRegion.Vendor,
		RegionID:   dbRegion.RegionID,
		RegionName: dbRegion.RegionName,
		AreaName:   dbRegion.AreaName,
		CityName:   dbRegion.CityName,
		Source:     dbRegion.Source,
		Status:     dbRegion.Status,
		Creator:    dbRegion.Creator,
		Reviser:    dbRegion.Reviser,
		CreatedAt:  dbRegion.CreatedAt.String(),
		UpdatedAt:  dbRegion.UpdatedAt.String(),
	}
}

// BatchDeleteTCloudZiyanRegion batch delete regions.
func (svc *regionSvc) BatchDeleteTCloudZiyanRegion(cts *rest.Contexts) error {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.TCloudZiyanRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud ziyan region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return fmt.Errorf("list tcloud ziyan region failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil
	}

	delRegionIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delRegionIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delRegionFilter := tools.ContainersExpression("id", delRegionIDs)
		if err = svc.dao.TCloudZiyanRegion().BatchDeleteWithTx(cts.Kit, txn, delRegionFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		logs.Errorf("delete tcloud ziyan region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}
