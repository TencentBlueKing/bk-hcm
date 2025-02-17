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

	"github.com/aws/aws-sdk-go/service/ec2"
)

// AwsNetworkInterfaceListOption defines aws network interface list option.
type AwsNetworkInterfaceListOption struct {
	Region  string        `json:"region" validate:"required"`
	Filters []*ec2.Filter `json:"filters" validate:"omitempty"`
	Page    *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate AwsNetworkInterfaceListOption.
func (a AwsNetworkInterfaceListOption) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}
	if a.Page != nil {
		return a.Page.Validate()
	}

	return nil
}

// AwsNetworkInterfaceWithCountResp defines Aws network interface with count.
type AwsNetworkInterfaceWithCountResp struct {
	NextToken *string
	Details   []AwsNetworkInterface
}

// AwsNetworkInterface defines Aws network interface.
type AwsNetworkInterface struct {
	*ec2.NetworkInterface
}
