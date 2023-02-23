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

package eip

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// TCloudEipListOption ...
type TCloudEipListOption struct {
	Region string `validate:"required"`
	Page   *core.TCloudPage
}

// ToDescribeAddressesRequest ...
func (o *TCloudEipListOption) ToDescribeAddressesRequest() (*vpc.DescribeAddressesRequest, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	req := vpc.NewDescribeAddressesRequest()
	if o.Page != nil {
		req.Offset = common.Int64Ptr(int64(o.Page.Offset))
		req.Limit = common.Int64Ptr(int64(o.Page.Limit))
	}
	return req, nil
}

// Validate ...
func (o *TCloudEipListOption) Validate() error {
	if err := validator.Validate.Struct(o); err != nil {
		return err
	}
	if o.Page != nil {
		if err := o.Page.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// TCloudEipListResult ...
type TCloudEipListResult struct {
	Count   *uint64
	Details []*TCloudEip
}

// TCloudEip ...
type TCloudEip struct {
	CloudID            string
	Name               *string
	Region             string
	InstanceId         *string
	Status             *string
	PublicIp           *string
	PrivateIp          *string
	Bandwidth          *uint64
	InternetChargeType *string
}
