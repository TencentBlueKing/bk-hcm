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
	"fmt"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/cidr"
)

// TCloudVpcCreateReq ...
type TCloudVpcCreateReq struct {
	BkBizID   int64  `json:"bk_biz_id" validate:"omitempty"`
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Name      string `json:"name" validate:"required,min=1,max=60"`
	IPv4Cidr  string `json:"ipv4_cidr" validate:"required,cidrv4"`

	Subnet struct {
		Name     string `json:"name" validate:"required,min=1,max=60"`
		IPv4Cidr string `json:"ipv4_cidr" validate:"required,cidrv4"`
		Zone     string `json:"zone" validate:"required"`
	} `json:"subnet" validate:"required"`

	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *TCloudVpcCreateReq) Validate(bizRequired bool) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if bizRequired && req.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	if err := cidr.IsSubnetContained(req.IPv4Cidr, req.Subnet.IPv4Cidr); err != nil {
		return fmt.Errorf("is subnet contained failed, err: %v", err)
	}

	return nil
}
