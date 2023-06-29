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

package cssubnet

import (
	"encoding/json"

	"hcm/pkg/criteria/validator"
)

// SubnetCreateReq define subnet create req.
type SubnetCreateReq struct {
	BkBizID    int64   `json:"bk_biz_id" validate:"required"`
	AccountID  string  `json:"account_id" validate:"required"`
	CloudVpcID string  `json:"cloud_vpc_id" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	Memo       *string `json:"memo" validate:"omitempty"`
	Data       json.RawMessage
}

// Validate TCloudSubnetCreateReq.
func (c SubnetCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}
