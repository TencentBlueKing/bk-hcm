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

// Package config implements cvm restrict config
package config

import (
	types "hcm/cmd/woa-server/types/config"
	devicecapacity "hcm/pkg/api/data-service/device-capacity"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetCvmDiskType gets cvm disk type config list
func (s *service) GetCvmDiskType(_ *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"count": 2,
		"info": []mapstr.MapStr{
			{
				"disk_type": "CLOUD_SSD",
				"disk_name": "SSD云硬盘",
			},
			{
				"disk_type": "CLOUD_PREMIUM",
				"disk_name": "高性能云盘",
			},
		},
	}

	return rst, nil
}

// GetCapacity get cvm capacity list
func (s *service) GetCapacity(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetCapacityParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get resource apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to get resource apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Capacity().GetCapacity(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get resource apply capacity, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpsertCapacity upsert cvm capacity
func (s *service) UpsertCapacity(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpdateCapacityParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to upsert resource apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if _, err := input.Validate(); err != nil {
		logs.Errorf("failed to upsert resource apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := s.logics.Capacity().UpsertCapacity(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to upsert resource apply capacity, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchGetCapacity 批量获取CVM容量信息
func (s *service) BatchGetCapacity(cts *rest.Contexts) (interface{}, error) {
	input := new(types.BatchGetCapacityParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to decode batch get capacity request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate batch get capacity request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Capacity().BatchGetCapacity(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to batch get resource apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// ListCapacityWithDeviceInfo 查询设备库存及其机型详细信息
func (s *service) ListCapacityWithDeviceInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(devicecapacity.ListCapacityWithDeviceInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode list capacity with device info request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Capacity().ListCapacityWithDeviceInfo(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to list capacity with device info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
