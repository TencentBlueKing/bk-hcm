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

package tcloud

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"hcm/pkg/ad/provider"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
)

// NewProvider new provider.
func NewProvider(s *types.BaseSecret) (*TCloud, error) {

	if s == nil {
		return nil, errf.New(errf.InvalidParameter, "secret is required")
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &TCloud{clientSet: newClientSet(s, profile.NewClientProfile())}, nil
}

// TCloud define tcloud provider.
type TCloud struct {
	clientSet *clientSet
}

// Vendor return vendor.
func (tcloud TCloud) Vendor() enumor.Vendor {
	return enumor.TCloud
}

var _ provider.Provider = new(TCloud)
