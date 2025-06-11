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

// Package zone ...
package zone

import (
	"fmt"
	"net/http"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/zone"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	tablezone "hcm/pkg/dal/table/cloud/zone"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitZoneService initial the security group service
func InitZoneService(cap *capability.Capability) {
	svc := &zoneSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateZone", http.MethodPost, "/vendors/{vendor}/zones/batch/create", svc.BatchCreateZone)

	h.Add("BatchUpdateZone", http.MethodPatch, "/vendors/{vendor}/zones/batch/update", svc.BatchUpdateZone)

	h.Add("ListZone", http.MethodPost, "/zones/list", svc.ListZone)

	h.Add("BatchDeleteZone", http.MethodDelete, "/zones/batch", svc.BatchDeleteZone)

	h.Load(cap.WebService)
}

type zoneSvc struct {
	dao dao.Set
}

// BatchCreateZone create zone.
func (svc *zoneSvc) BatchCreateZone(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateZone[zone.TCloudZoneExtension](vendor, svc, cts)
	case enumor.Aws:
		return batchCreateZone[zone.AwsZoneExtension](vendor, svc, cts)
	case enumor.HuaWei:
		return batchCreateZone[zone.HuaWeiZoneExtension](vendor, svc, cts)
	case enumor.Gcp:
		return batchCreateZone[zone.GcpZoneExtension](vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// BatchUpdateZone update zone.
func (svc *zoneSvc) BatchUpdateZone(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateZone[zone.TCloudZoneExtension](cts, svc)
	case enumor.Aws:
		return batchUpdateZone[zone.AwsZoneExtension](cts, svc)
	case enumor.HuaWei:
		return batchUpdateZone[zone.HuaWeiZoneExtension](cts, svc)
	case enumor.Azure:
		return batchUpdateZone[zone.GcpZoneExtension](cts, svc)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// BatchDeleteZone delete zone.
func (svc *zoneSvc) BatchDeleteZone(cts *rest.Contexts) (interface{}, error) {

	req := new(protocloud.ZoneBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.Zone().Delete(cts.Kit, req.Filter); err != nil {
		logs.Errorf("delete zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListZone list zone.
func (svc *zoneSvc) ListZone(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ZoneListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.Zone().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list zone failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ZoneListResult{Count: result.Count}, nil
	}

	details := make([]zone.BaseZone, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, zone.BaseZone{
			ID:        one.ID,
			Vendor:    enumor.Vendor(one.Vendor),
			CloudID:   one.CloudID,
			Name:      one.Name,
			Region:    one.Region,
			NameCn:    one.NameCn,
			State:     one.State,
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		})
	}

	return &protocloud.ZoneListResult{Details: details}, nil
}

func batchCreateZone[T zone.ZoneExtension](vendor enumor.Vendor, svc *zoneSvc,
	cts *rest.Contexts) (interface{}, error) {

	req := new(protocloud.ZoneBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		zones := make([]*tablezone.ZoneTable, 0, len(req.Zones))
		for _, zone := range req.Zones {
			extension, err := json.MarshalToString(zone.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			zones = append(zones, &tablezone.ZoneTable{
				Vendor:    vendor,
				CloudID:   zone.CloudID,
				Name:      zone.Name,
				Region:    zone.Region,
				NameCn:    zone.NameCn,
				State:     zone.State,
				Extension: tabletype.JsonField(extension),
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			})
		}

		ids, err := svc.dao.Zone().CreateWithTx(cts.Kit, txn, zones)
		if err != nil {
			return nil, fmt.Errorf("create zone failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create zone but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func batchUpdateZone[T zone.ZoneExtension](cts *rest.Contexts, svc *zoneSvc) (
	interface{}, error) {

	req := new(protocloud.ZoneBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Zones))
	for _, one := range req.Zones {
		ids = append(ids, one.ID)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, zone := range req.Zones {
			update := &tablezone.ZoneTable{
				State:   zone.State,
				Reviser: cts.Kit.User,
			}

			if err := svc.dao.Zone().UpdateByIDWithTx(cts.Kit, txn, zone.ID, update); err != nil {
				logs.Errorf("UpdateByIDWithTx zone failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update zone failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
