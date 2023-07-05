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

package csvpc

import (
	"errors"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/assert"
)

// AzureVpcCreateReq ...
type AzureVpcCreateReq struct {
	BkBizID           int64  `json:"bk_biz_id" validate:"omitempty"`
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required,lowercase"`
	Region            string `json:"region" validate:"required,lowercase"`
	Name              string `json:"name" validate:"required,min=1,max=60,lowercase"`
	IPv4Cidr          string `json:"ipv4_cidr" validate:"required,cidrv4"`
	BkCloudID         int64  `json:"bk_cloud_id" validate:"required,min=1"`

	Subnet struct {
		Name     string `json:"name" validate:"required,min=1,max=60,lowercase"`
		IPv4Cidr string `json:"ipv4_cidr" validate:"required,cidrv4"`
	} `json:"subnet" validate:"required"`

	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *AzureVpcCreateReq) Validate(bizRequired bool) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if bizRequired && req.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	// region can be no space lowercase
	if !assert.IsSameCaseNoSpaceString(req.Region) {
		return errf.New(errf.InvalidParameter, "region can only be lowercase")
	}

	return nil
}
