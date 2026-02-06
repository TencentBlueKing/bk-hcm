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

package tcloudziyanpmdevicetype

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	coretcloudziyanpmdevicetype "hcm/pkg/api/core/tcloud-ziyan-pm-device-type"
	datatcloudziyanpmdevicetype "hcm/pkg/api/data-service/tcloud-ziyan-pm-device-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	ziyanpmdt "hcm/pkg/dal/table/tcloud-ziyan-pm-device-type"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// CreateTCloudZiyanPmDeviceType create tcloud ziyan pm device type.
func (svc *service) CreateTCloudZiyanPmDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(datatcloudziyanpmdevicetype.CreateTCloudZiyanPmDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]ziyanpmdt.TCloudZiyanPmDeviceTypeTable, 0, len(req.Items))
		for _, item := range req.Items {
			model := ziyanpmdt.TCloudZiyanPmDeviceTypeTable{
				DeviceType: item.DeviceType,
				Raid:       item.Raid,
				CpuCore:    item.CpuCore,
				Memory:     item.Memory,
				Disable:    item.Disable,
				Creator:    cts.Kit.User,
				Reviser:    cts.Kit.User,
			}
			models = append(models, model)
		}
		ids, err := svc.dao.TCloudZiyanPmDeviceType().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create tcloud ziyan pm device type failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create tcloud ziyan pm device type commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := createIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create tcloud ziyan pm device type but return id type not string, id type: %v",
			reflect.TypeOf(createIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// DeleteTCloudZiyanPmDeviceType delete tcloud ziyan pm device type.
func (svc *service) DeleteTCloudZiyanPmDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(datatcloudziyanpmdevicetype.DeleteTCloudZiyanPmDeviceTypeReq)
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
			listResp, err := svc.dao.TCloudZiyanPmDeviceType().List(cts.Kit, opt)
			if err != nil {
				logs.Errorf("list tcloud ziyan pm device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("list tcloud ziyan pm device type failed, err: %v", err)
			}

			for _, one := range listResp.DeviceTypes {
				delIDs = append(delIDs, one.ID)
			}

			if uint(len(listResp.DeviceTypes)) < page.Limit {
				break
			}
			page.Start += uint32(page.Limit)
		}

		for _, chunk := range slice.Split(delIDs, int(filter.DefaultMaxInLimit)) {
			delFilter := tools.ContainersExpression("id", chunk)
			if err := svc.dao.TCloudZiyanPmDeviceType().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				logs.Errorf("delete tcloud ziyan pm device type chunk failed, err: %s, chunk: %v, rid: %s", err, chunk,
					cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete tcloud ziyan pm device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// UpdateTCloudZiyanPmDeviceType update tcloud ziyan pm device type.
func (svc *service) UpdateTCloudZiyanPmDeviceType(cts *rest.Contexts) (interface{}, error) {
	req := new(datatcloudziyanpmdevicetype.UpdateTCloudZiyanPmDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Items {
			deviceType := &ziyanpmdt.TCloudZiyanPmDeviceTypeTable{
				Reviser: cts.Kit.User,
			}

			if one.DeviceType != nil {
				deviceType.DeviceType = converter.PtrToVal(one.DeviceType)
			}
			if one.Raid != nil {
				deviceType.Raid = converter.PtrToVal(one.Raid)
			}
			if one.CpuCore != nil {
				deviceType.CpuCore = converter.PtrToVal(one.CpuCore)
			}
			if one.Memory != nil {
				deviceType.Memory = converter.PtrToVal(one.Memory)
			}
			if one.Disable != nil {
				deviceType.Disable = converter.PtrToVal(one.Disable)
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.TCloudZiyanPmDeviceType().UpdateWithTx(cts.Kit, txn, flt, deviceType); err != nil {
				logs.Errorf("update tcloud ziyan pm device type failed, err: %v, rid: %s, id: %s", err, cts.Kit.Rid,
					one.ID)
				return nil, fmt.Errorf("update tcloud ziyan pm device type failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListTCloudZiyanPmDeviceType list tcloud ziyan pm device type.
func (svc *service) ListTCloudZiyanPmDeviceType(cts *rest.Contexts) (interface{}, error) {
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
	res, err := svc.dao.TCloudZiyanPmDeviceType().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud ziyan pm device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud ziyan pm device type failed, err: %v", err)
	}
	if req.Page.Count {
		return &datatcloudziyanpmdevicetype.ListTCloudZiyanPmDeviceTypeResult{Count: res.Count}, nil
	}

	deviceTypes := make([]coretcloudziyanpmdevicetype.TCloudZiyanPmDeviceType, 0, len(res.DeviceTypes))
	for _, one := range res.DeviceTypes {
		deviceTypes = append(deviceTypes, coretcloudziyanpmdevicetype.TCloudZiyanPmDeviceType{
			ID:         one.ID,
			DeviceType: one.DeviceType,
			Raid:       one.Raid,
			CpuCore:    one.CpuCore,
			Memory:     one.Memory,
			Disable:    one.Disable,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &datatcloudziyanpmdevicetype.ListTCloudZiyanPmDeviceTypeResult{Details: deviceTypes}, nil
}
