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

package devicetype

import (
	coredevicetype "hcm/pkg/api/core/cloud/device-type"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListDeviceType list device type.
func (svc *service) ListDeviceType(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.DeviceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}

	listResult, err := svc.dao.DeviceType().List(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("list device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.DeviceTypeListResult{Count: listResult.Count}, nil
	}

	// Convert DeviceTypeTable to DeviceType
	details := make([]coredevicetype.DeviceType, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		details = append(details, coredevicetype.ConvTableToDeviceType(one))
	}

	return &protocloud.DeviceTypeListResult{Details: details}, nil
}

// ListDistinctDeviceType list distinct device type.
func (svc *service) ListDistinctDeviceType(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.DistinctDeviceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}

	listResult, err := svc.dao.DeviceType().ListDistinctDeviceType(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("list distinct device type failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.DistinctDeviceTypeListResult{Count: listResult.Count}, nil
	}

	details := make([]coredevicetype.DistinctDeviceType, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		details = append(details, coredevicetype.ConvTableToDistinctDeviceType(one))
	}

	return &protocloud.DistinctDeviceTypeListResult{Details: details}, nil
}
