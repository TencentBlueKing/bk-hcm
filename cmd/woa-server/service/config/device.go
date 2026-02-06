/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package config device config
package config

import (
	types "hcm/cmd/woa-server/types/config"
	dataproto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	datapmdevicetype "hcm/pkg/api/data-service/tcloud-ziyan-pm-device-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetDeviceType gets config device type list
func (s *service) GetDeviceType(cts *rest.Contexts) (interface{}, error) {
	input := new(protocloud.DistinctDeviceTypeListReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Device().ListDistinctDeviceType(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetCvmDeviceDetail gets config cvm device detail info by condition
func (s *service) GetCvmDeviceDetail(cts *rest.Contexts) (interface{}, error) {
	input := new(protocloud.DeviceTypeListReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Device().ListDeviceType(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateManyDevice creates device configs in batch
func (s *service) CreateManyDevice(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.DeviceTypeBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := s.logics.Device().BatchCreateDeviceType(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to batch create device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpdateDeviceProperty updates device config property
func (s *service) UpdateDeviceProperty(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.DeviceTypeBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.logics.Device().BatchUpdateDeviceType(cts.Kit, req); err != nil {
		logs.Errorf("failed to batch update device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteDevice deletes device config
func (s *service) DeleteDevice(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := s.logics.Device().BatchDeleteDeviceType(cts.Kit, req); err != nil {
		return nil, err
	}
	return nil, nil
}

// GetDvmDeviceType gets config dvm device type list
func (s *service) GetDvmDeviceType(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.GetDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get dvm device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().GetDvmDeviceType(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get dvm device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateDvmDevice creates config dvm device type
func (s *service) CreateDvmDevice(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.DvmDeviceInfo)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create dvm device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM机型-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmType, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.Device().CreateDvmDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to create dvm device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetPmDeviceType gets config physical machine device type list
func (s *service) GetPmDeviceType(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.GetPmDeviceTypeParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get physical machine device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().GetPmDeviceType(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get physical machine device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreatePmDevice creates config physical machine device type
func (s *service) CreatePmDevice(cts *rest.Contexts) (interface{}, error) {
	req := new(datapmdevicetype.CreateTCloudZiyanPmDeviceTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to create physical machine device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// CVM机型-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmType, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.Device().CreatePmDevice(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to create physical machine device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
