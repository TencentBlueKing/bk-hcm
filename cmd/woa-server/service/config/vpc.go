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

// Package config vpc config
package config

import (
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// GetVpc gets vpc config list
func (s *service) GetVpc(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetVpcParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Vpc().GetVpc(cts.Kit, []string{input.Region})
	if err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetVpcList gets vpc id list
func (s *service) GetVpcList(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetVpcListParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rst, err := s.logics.Vpc().GetVpcList(cts.Kit, input.Regions)
	if err != nil {
		logs.Errorf("failed to get vpc list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpsertRegionDftVpc upsert region default vpc
func (s *service) UpsertRegionDftVpc(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpsertRegionDftVpcReq)
	if err := cts.DecodeInto(input); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	if err := s.logics.Vpc().UpsertRegionDftVpc(cts.Kit, input.RegionDftVpcInfos); err != nil {
		logs.Errorf("failed to upsert region default vpc, err: %v, input: %v, rid: %s", err, converter.PtrToVal(input),
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
