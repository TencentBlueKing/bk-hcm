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

package cvm

import (
	"fmt"
	"slices"

	types "hcm/cmd/woa-server/types/cvm"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/cvmapi"
)

// CreateApplyOrder creates apply order(CVM生产-创建单据)
func (s *service) CreateApplyOrder(cts *rest.Contexts) (interface{}, error) {
	input := new(types.CvmCreateReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err = s.validateDeviceTypeForGreenAndRoll(cts.Kit, input); err != nil {
		logs.Errorf("failed to validate device type for green channel or roll server apply, err: %v, input: %+v, rid: %s",
			err, input, cts.Kit.Rid)
		return nil, err
	}

	// CVM生产-菜单粒度鉴权
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmCreate, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.CreateApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to create apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

func (s *service) validateDeviceTypeForGreenAndRoll(kt *kit.Kit, input *types.CvmCreateReq) error {
	requireType := enumor.RequireType(input.RequireType)
	if requireType != enumor.RequireTypeGreenChannel && requireType != enumor.RequireTypeRollServer {
		return nil
	}

	var deviceType string
	if input.Spec != nil {
		deviceType = input.Spec.DeviceType
	}
	if deviceType == "" {
		logs.Errorf("device type is required for green channel or roll server apply, input: %v, rid: %s",
			input, kt.Rid)
		return fmt.Errorf("device type is required for green channel or roll server apply")
	}

	deviceReq := &protocloud.DistinctDeviceTypeListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
				tools.RuleEqual("device_type", deviceType)),
			Page: core.NewDefaultBasePage(),
		},
	}
	resp, err := s.configLogics.Device().ListDistinctDeviceType(kt, deviceReq)
	if err != nil {
		logs.Errorf("failed to get device type info, err: %v, deviceTypes: %v, rid: %s", err, deviceType, kt.Rid)
		return err
	}
	if len(resp.Details) == 0 {
		logs.Errorf("device type not found, deviceType: %s, rid: %s", deviceType, kt.Rid)
		return fmt.Errorf("device type not found, deviceType: %s", deviceType)
	}
	deviceInfo := resp.Details[0]
	if deviceInfo.DeviceTypeClass == cvmapi.SpecialType {
		return fmt.Errorf("device type %s is not supported for green channel or roll server apply", deviceType)
	}

	if requireType != enumor.RequireTypeGreenChannel {
		return nil
	}

	// 获取小额绿通的配置
	gcConfigs, err := s.gcLogics.GetConfigs(kt)
	if err != nil {
		logs.Errorf("get green channel configs failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	cvmApplyConfigs := gcConfigs.CvmApplyConfig
	if !cvmApplyConfigs.Enabled {
		return nil
	}
	if !(slices.Contains(cvmApplyConfigs.DeviceGroups, deviceInfo.DeviceFamily) && deviceInfo.CpuCore <=
		cvmApplyConfigs.CpuMaxLimit) {
		return fmt.Errorf("device type %s is not supported for green channel", deviceType)
	}

	return nil
}

// GetApplyOrderById gets apply order info by order id
func (s *service) GetApplyOrderById(cts *rest.Contexts) (interface{}, error) {
	input := new(types.CvmOrderReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM生产-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmCreate, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.GetApplyOrderById(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyOrder gets apply order info
func (s *service) GetApplyOrder(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetApplyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate get apply order, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, err
	}

	// CVM生产-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmCreate, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.GetApplyOrder(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyDevice gets apply order launched devices
func (s *service) GetApplyDevice(cts *rest.Contexts) (interface{}, error) {
	input := new(types.CvmDeviceReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get apply order launched devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM生产-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmCreate, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.GetApplyDevice(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get apply order launched devices, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetCapacity gets cvm apply capacity
func (s *service) GetCapacity(cts *rest.Contexts) (interface{}, error) {
	input := new(types.CvmCapacityReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get cvm apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// CVM生产-菜单粒度鉴权
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmCreate, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	rst, err := s.logics.GetCapacity(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get cvm apply capacity, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetApplyStatusCfg get apply status config
func (s *service) GetApplyStatusCfg(_ *rest.Contexts) (interface{}, error) {
	// TODO: store in db
	rst := mapstr.MapStr{
		"info": []mapstr.MapStr{
			{
				"status":      types.ApplyStatusInit,
				"description": "未执行",
			},
			{
				"status":      types.ApplyStatusRunning,
				"description": "执行中",
			},
			{
				"status":      types.ApplyStatusSuccess,
				"description": "成功",
			},
			{
				"status":      types.ApplyStatusFailed,
				"description": "失败",
			},
		},
	}

	return rst, nil
}
