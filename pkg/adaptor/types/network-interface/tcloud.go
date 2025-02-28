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

package networkinterface

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// TCloudNetworkInterfaceListOption defines tcloud network interface list option.
type TCloudNetworkInterfaceListOption struct {
	Region  string           `json:"region" validate:"required"`
	Filters []*vpc.Filter    `json:"filters" validate:"omitempty"`
	Page    *core.TCloudPage `json:"page" validate:"required"`
}

// Validate tcloud network interface list option.
func (opt TCloudNetworkInterfaceListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}
	if err := opt.Page.Validate(); err != nil {
		return err
	}
	return nil
}

// TCloudNetworkInterfaceWithCountResp defines tcloud network interface with count.
type TCloudNetworkInterfaceWithCountResp struct {
	TotalCount uint64
	Details    []TCloudNetworkInterface
}

// TCloudNetworkInterface defines tcloud network interface.
type TCloudNetworkInterface struct {
	*vpc.NetworkInterface
}
