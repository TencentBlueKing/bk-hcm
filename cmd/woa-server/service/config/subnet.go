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

// Package config subnet config
package config

import (
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetSubnet gets subnet config list
func (s *service) GetSubnet(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetSubnetParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	subnetReq := &types.GetAllSubnetReq{
		Region:     input.Region,
		Zones:      []string{input.Zone},
		CloudVpcID: input.Vpc,
	}
	rst, err := s.logics.Subnet().GetAllSubnet(cts.Kit, subnetReq)
	if err != nil {
		logs.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetSubnetList gets subnet detail config list
func (s *service) GetSubnetList(cts *rest.Contexts) (interface{}, error) {
	input := new(types.GetSubnetListParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rst, err := s.logics.Subnet().GetSubnetList(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get subnet list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// UpdateSubnetProperty updates subnet config property
func (s *service) UpdateSubnetProperty(cts *rest.Contexts) (interface{}, error) {
	input := new(types.UpdateSubnetPropertyParam)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to update subnet, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err := input.Validate()
	if err != nil {
		logs.Errorf("failed to update subnet, err: %v, input: %+v, rid: %s", err, input, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// CVM子网-菜单粒度鉴权
	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.ZiyanCvmSubnet, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	data := input.Property
	// cannot update device id
	delete(data, "id")

	if err = s.logics.Subnet().UpdateSubnetBatch(cts.Kit, input.Ids, input.Property); err != nil {
		logs.Errorf("failed to update subnet, err: %v, ids: %v, rid: %s", err, input.Ids, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
