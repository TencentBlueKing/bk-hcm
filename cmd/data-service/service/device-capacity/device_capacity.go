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

// Package devicecapacity ...
package devicecapacity

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	coredevicecapacity "hcm/pkg/api/core/device-capacity"
	devicecapacity "hcm/pkg/api/data-service/device-capacity"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tabledevicecapacity "hcm/pkg/dal/table/device-capacity"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// CreateDeviceCapacity create device capacity.
func (svc *service) CreateDeviceCapacity(cts *rest.Contexts) (interface{}, error) {
	req := new(devicecapacity.CreateDeviceCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tabledevicecapacity.DeviceCapacityTable, 0, len(req.Items))
		for _, item := range req.Items {
			model := tabledevicecapacity.DeviceCapacityTable{
				RequireType: item.RequireType,
				Region:      item.Region,
				Zone:        item.Zone,
				DeviceType:  item.DeviceType,
				Capacity:    item.Capacity,
				Extension:   item.Extension,
				Creator:     cts.Kit.User,
				Reviser:     cts.Kit.User,
			}
			models = append(models, model)
		}
		ids, err := svc.dao.DeviceCapacity().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create device capacity failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create device capacity commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := createIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create device capacity but return id type not string, id type: %v",
			reflect.TypeOf(createIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// DeleteDeviceCapacity delete device capacity.
func (svc *service) DeleteDeviceCapacity(cts *rest.Contexts) (interface{}, error) {
	req := new(devicecapacity.DeleteDeviceCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delIDs := make([]string, 0)
		page := core.NewDefaultBasePage()
		for {
			opt := &types.ListOption{
				Fields: []string{"id"},
				Filter: req.Filter,
				Page:   page,
			}
			listResp, err := svc.dao.DeviceCapacity().List(cts.Kit, opt)
			if err != nil {
				logs.Errorf("list device capacity failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("list device capacity failed, err: %v", err)
			}

			for _, one := range listResp.DeviceCapacities {
				delIDs = append(delIDs, one.ID)
			}

			if uint(len(listResp.DeviceCapacities)) < page.Limit {
				break
			}
			page.Start += uint32(page.Limit)
		}

		for _, chunk := range slice.Split(delIDs, int(filter.DefaultMaxInLimit)) {
			delFilter := tools.ContainersExpression("id", chunk)
			if err := svc.dao.DeviceCapacity().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				logs.Errorf("delete device capacity chunk failed, err: %s, chunk: %v, rid: %s", err, chunk, cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete device capacity failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// UpdateDeviceCapacity update device capacity.
func (svc *service) UpdateDeviceCapacity(cts *rest.Contexts) (interface{}, error) {
	req := new(devicecapacity.UpdateDeviceCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Items {
			deviceCapacity := &tabledevicecapacity.DeviceCapacityTable{
				Reviser: cts.Kit.User,
			}

			if one.RequireType != nil {
				deviceCapacity.RequireType = converter.PtrToVal(one.RequireType)
			}
			if one.Region != nil {
				deviceCapacity.Region = converter.PtrToVal(one.Region)
			}
			if one.Zone != nil {
				deviceCapacity.Zone = converter.PtrToVal(one.Zone)
			}
			if one.DeviceType != nil {
				deviceCapacity.DeviceType = converter.PtrToVal(one.DeviceType)
			}
			if one.Capacity != nil {
				deviceCapacity.Capacity = one.Capacity
			}
			if one.Extension != nil {
				deviceCapacity.Extension = converter.PtrToVal(one.Extension)
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.DeviceCapacity().UpdateWithTx(cts.Kit, txn, flt, deviceCapacity); err != nil {
				logs.Errorf("update device capacity failed, err: %v, id: %s, rid: %s", err, one.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update device capacity failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListDeviceCapacity list device capacity.
func (svc *service) ListDeviceCapacity(cts *rest.Contexts) (interface{}, error) {
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
	res, err := svc.dao.DeviceCapacity().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list device capacity failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list device capacity failed, err: %v", err)
	}
	if req.Page.Count {
		return &devicecapacity.ListDeviceCapacityResult{Count: res.Count}, nil
	}

	deviceCapacities := make([]coredevicecapacity.DeviceCapacity, 0, len(res.DeviceCapacities))
	for _, one := range res.DeviceCapacities {
		deviceCapacities = append(deviceCapacities, coredevicecapacity.DeviceCapacity{
			ID:          one.ID,
			RequireType: one.RequireType,
			Region:      one.Region,
			Zone:        one.Zone,
			DeviceType:  one.DeviceType,
			Capacity:    one.Capacity,
			Extension:   one.Extension,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &devicecapacity.ListDeviceCapacityResult{Details: deviceCapacities}, nil
}

// ListCapacityWithDeviceInfo list device capacity with device type details.
func (svc *service) ListCapacityWithDeviceInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(devicecapacity.ListCapacityWithDeviceInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	res, err := svc.dao.DeviceCapacity().ListWithDeviceInfo(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list device capacity with device info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list device capacity with device info failed, err: %v", err)
	}

	if req.Page != nil && req.Page.Count {
		return &devicecapacity.ListCapacityWithDeviceInfoResult{Count: res.Count}, nil
	}

	details := make([]coredevicecapacity.CapacityWithDeviceInfo, 0, len(res.Details))
	for _, one := range res.Details {
		details = append(details, coredevicecapacity.CapacityWithDeviceInfo{
			RequireType:     enumor.RequireType(one.RequireType),
			Region:          one.Region,
			Zone:            one.Zone,
			DeviceFamily:    one.DeviceFamily,
			DeviceType:      one.DeviceType,
			CPUCore:         one.CPUCore,
			Memory:          one.Memory,
			Capacity:        one.Capacity,
			CoreType:        one.CoreType,
			DeviceTypeClass: one.DeviceTypeClass,
		})
	}

	return &devicecapacity.ListCapacityWithDeviceInfoResult{Details: details}, nil
}
