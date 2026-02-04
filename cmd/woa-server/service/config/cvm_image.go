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

// Package config cvm image config
package config

import (
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetCvmImage gets cvm image config list
func (s *service) GetCvmImage(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetCvmImageParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get cvm image list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.CvmImage().GetCvmImage(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get cvm image list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// BatchEnableImageToApplyCVM 批量允许镜像用于申领CVM
func (s *service) BatchEnableImageToApplyCVM(cts *rest.Contexts) (interface{}, error) {
	return s.batchOpImageToApplyCVM(cts, s.logics.CvmImage().BatchEnableImageCvm)
}

// BatchDisableImageToApplyCVM 批量禁止镜像用于申领CVM
func (s *service) BatchDisableImageToApplyCVM(cts *rest.Contexts) (interface{}, error) {
	return s.batchOpImageToApplyCVM(cts, s.logics.CvmImage().BatchDisableImageCvm)
}

// batchOpImageToApplyCVM 批量操作镜像用于申领CVM的通用处理函数
func (s *service) batchOpImageToApplyCVM(cts *rest.Contexts, opFunc func(*kit.Kit, []string) error) (
	interface{}, error) {

	req := new(types.BatchOpImageToApplyCVMReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode batch operate image cvm request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate batch operate image cvm request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := opFunc(cts.Kit, req.ImageIDs); err != nil {
		logs.Errorf("failed to batch operate image cvm, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
