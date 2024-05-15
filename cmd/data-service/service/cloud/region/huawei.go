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
	coreregion "hcm/pkg/api/core/cloud/region"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableregion "hcm/pkg/dal/table/cloud/region"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateHuaWeiRegion update azure region.
func (svc *regionSvc) BatchUpdateHuaWeiRegion(cts *rest.Contexts) (interface{}, error) {

	req := new(protoregion.HuaWeiRegionBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Regions {
			rule := &tableregion.HuaWeiRegionTable{
				Type:        one.Type,
				LocalesZhCn: one.LocalesZhCn,
				Reviser:     cts.Kit.User,
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.HuaWeiRegion().UpdateWithTx(cts.Kit, txn, flt, rule); err != nil {
				logs.Errorf("update azure resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update azure resource group failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchCreateHuaWeiRegion create region.
func (svc *regionSvc) BatchCreateHuaWeiRegion(cts *rest.Contexts) (interface{}, error) {
	req := new(protoregion.HuaWeiRegionBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	regions := make([]*tableregion.HuaWeiRegionTable, 0, len(req.Regions))
	for _, region := range req.Regions {
		regions = append(regions, &tableregion.HuaWeiRegionTable{
			RegionID:    region.RegionID,
			LocalesZhCn: region.LocalesZhCn,
			Type:        region.Type,
			Service:     region.Service,
			Creator:     cts.Kit.User,
			Reviser:     cts.Kit.User,
		})
	}

	regionIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		regionIDs, err := svc.dao.HuaWeiRegion().CreateWithTx(cts.Kit, txn, regions)
		if err != nil {
			return nil, fmt.Errorf("batch create huawei region failed, err: %v", err)
		}
		return regionIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := regionIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create huawei region but return id type is not string, id type: %v",
			reflect.TypeOf(regionIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchDeleteHuaWeiRegion delete huawei region.
func (svc *regionSvc) BatchDeleteHuaWeiRegion(cts *rest.Contexts) (interface{}, error) {

	req := new(protoregion.HuaWeiRegionBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.HuaWeiRegion().DeleteWithTx(cts.Kit, req.Filter); err != nil {
		logs.Errorf("delete huawei region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListHuaWeiRegion list huawei region with filter
func (svc *regionSvc) ListHuaWeiRegion(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.HuaWeiRegion().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list huawei region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list huawei region failed, err: %v", err)
	}

	if req.Page.Count {
		return &protoregion.HuaWeiRegionListResult{Count: result.Count}, nil
	}

	details := make([]coreregion.HuaWeiRegion, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, coreregion.HuaWeiRegion{
			ID:          one.ID,
			Service:     one.Service,
			RegionID:    one.RegionID,
			Type:        one.Type,
			LocalesPtBr: one.LocalesPtBr,
			LocalesZhCn: one.LocalesZhCn,
			LocalesEnUs: one.LocalesEnUs,
			LocalesEsUs: one.LocalesEsUs,
			LocalesEsEs: one.LocalesEsEs,
			Creator:     one.Creator,
			Reviser:     one.Reviser,
			CreatedAt:   one.CreatedAt.String(),
			UpdatedAt:   one.UpdatedAt.String(),
		})
	}

	return &protoregion.HuaWeiRegionListResult{Details: details}, nil
}
