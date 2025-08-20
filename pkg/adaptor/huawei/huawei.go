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

// Package huawei ...
package huawei

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
)

const (
	// Ecs cvm disk network interface
	Ecs = "ecs"
	// Vpc vpc subnet sg sgRule route table
	Vpc = "vpc"
	// Eip eip
	Eip = "eip"
	// Ims public image
	Ims = "ims"
	// Dcs zone
	Dcs = "dcs"
)

// NewHuaWei new huawei.
func NewHuaWei(s *types.BaseSecret) (*HuaWei, error) {
	if err := validateSecret(s); err != nil {
		return nil, err
	}
	return &HuaWei{clientSet: newClientSet(s)}, nil
}

// HuaWei is huawei operator.
type HuaWei struct {
	clientSet *clientSet
}

func validateSecret(s *types.BaseSecret) error {
	if s == nil {
		return errf.New(errf.InvalidParameter, "secret is required")
	}

	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

// sliceToPtr convert slice to pointer.
func sliceToPtr[T any](slice []T) *[]T {
	ptrArr := make([]T, len(slice))
	for idx, val := range slice {
		ptrArr[idx] = val
	}
	return &ptrArr
}
