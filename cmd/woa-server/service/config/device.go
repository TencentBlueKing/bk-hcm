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
	"errors"
	"fmt"
	"strconv"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetDeviceWithCapacity gets config device detail info with capacity
func (s *service) GetDeviceWithCapacity(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.GetDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().GetDeviceWithCapacity(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetDevice gets all available config device detail info
func (s *service) GetDevice(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.GetDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().GetDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetDeviceType gets config device type list
func (s *service) GetDeviceType(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.GetDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().GetDeviceType(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetDeviceTypeDetail gets config device type with detail info
func (s *service) GetDeviceTypeDetail(cts *rest.Contexts) (interface{}, error) {
	// TODO: input validation
	input := new(types.GetDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get device type detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().GetDeviceTypeDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device type detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetCvmDeviceDetail gets config cvm device detail info by condition
func (s *service) GetCvmDeviceDetail(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get cvm device config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get cvm device config, key: %s, err: %v, rid: %s", errKey, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Device().GetCvmDeviceDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get device list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateDevice creates device config
func (s *service) CreateDevice(cts *rest.Contexts) (interface{}, error) {
	inputData := new(types.DeviceInfo)
	if err := cts.DecodeInto(inputData); err != nil {
		logs.Errorf("failed to create device, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Device().CreateDevice(cts.Kit, inputData)
	if err != nil {
		logs.Errorf("failed to create device, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 在CRP device_type中同步该机型
	if err = s.planLogics.SyncDeviceTypesFromCRP(cts.Kit, []string{inputData.DeviceType}); err != nil {
		logs.Errorf("failed to sync res plan device type, err: %v, input: %+v, rid: %s", err, inputData,
			cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// CreateManyDevice creates device configs in batch
func (s *service) CreateManyDevice(cts *rest.Contexts) (interface{}, error) {
	input := new(types.CreateManyDeviceParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create device in batch, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to create device in batch, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	crpDeviceTypeMap, err := s.logics.Device().ListDeviceTypeInfoFromCrp(cts.Kit, []string{input.DeviceType})
	if err != nil {
		logs.Errorf("failed to get device type from CRP, err: %v, deviceType: %s, rid: %s", err, input.DeviceType,
			cts.Kit.Rid)
		return nil, err
	}

	crpDeviceInfo, exists := crpDeviceTypeMap[input.DeviceType]
	if !exists && !input.ForceCreate {
		logs.Errorf("device type not exist in CRP, input: %+v, rid: %s", input, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DeviceTypeAbsentInCRP, errors.New("device type not exist in CRP"))
	}

	// 当机型存在于crp时，那么创建时以crp的实例族为准
	if exists {
		input.DeviceGroup = crpDeviceInfo.InstanceGroup
		// 判断输入的 CPU核数，内存大小，大小核心 和crp的是否一致
		if input.Cpu != int64(crpDeviceInfo.CPUAmount) || input.Mem != int64(crpDeviceInfo.RamAmount) ||
			input.DeviceSize != enumor.GetCoreTypeByCRPCoreTypeID(crpDeviceInfo.CoreType) {
			return nil, fmt.Errorf("input device config not match CRP device config, input: %+v, crp: %+v",
				input, crpDeviceInfo)
		}
		input.DeviceTypeClass = crpDeviceInfo.InstanceTypeClass
		input.TechnicalClass = crpDeviceInfo.CvmInstanceTypeClass
	}
	if err = s.logics.Device().CreateManyDevice(cts.Kit, input); err != nil {
		logs.Errorf("failed to create device in batch, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	// 在CRP device_type中同步该机型
	if err = s.planLogics.SyncDeviceTypesFromCRP(cts.Kit, []string{input.DeviceType}); err != nil {
		logs.Errorf("failed to sync res plan device type, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateDevice updates device config
func (s *service) UpdateDevice(cts *rest.Contexts) (interface{}, error) {
	input := make(map[string]interface{})
	if err := cts.DecodeInto(&input); err != nil {
		logs.Errorf("failed to update device, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM机型-菜单粒度鉴权
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmType, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	if err = s.logics.Device().UpdateDevice(cts.Kit, instId, input); err != nil {
		logs.Errorf("failed to update device, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateDeviceProperty updates device config property
func (s *service) UpdateDeviceProperty(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpdateDevicePropertyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to update cvm device config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to update cvm device config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// CVM机型-菜单粒度鉴权
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmType, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	cond := map[string]interface{}{
		"id": map[string]interface{}{
			pkg.BKDBIN: input.Ids,
		},
	}

	data := input.Property
	// cannot update device id
	delete(data, "id")

	if err = s.logics.Device().UpdateDeviceBatch(cts.Kit, cond, input.Property); err != nil {
		logs.Errorf("failed to update cvm device config, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteDevice deletes device config
func (s *service) DeleteDevice(cts *rest.Contexts) (interface{}, error) {
	instId, err := strconv.ParseInt(cts.Request.PathParameter("id"), 10, 64)
	if err != nil {
		logs.Errorf("failed to parse id, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM机型-菜单粒度鉴权
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmType, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	if err = s.logics.Device().DeleteDevice(cts.Kit, instId); err != nil {
		logs.Errorf("failed to delete device, err: %v, rid: %s", err, cts.Kit.Rid)
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
	input := new(types.GetDeviceParam)
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
	// TODO: input validation
	input := new(types.PmDeviceInfo)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create physical machine device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM机型-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmType, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.Device().CreatePmDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to create physical machine device type, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
