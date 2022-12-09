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

package hcservice

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
)

// AccountCheckReq defines account check request.
// TODO: 结构体信息随便定义，之后自行替换即可
type AccountCheckReq struct {
	Vendor      enumor.Vendor             `json:"vendor,omitempty"`
	Secret      *types.Secret             `json:"secret,omitempty"`
	AccountInfo *types.AccountCheckOption `json:"account_info,omitempty"`
}

// Validate account check req.
func (req AccountCheckReq) Validate() error {
	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	if req.Secret == nil {
		return errf.New(errf.InvalidParameter, "secret is required")
	}

	if req.AccountInfo == nil {
		return errf.New(errf.InvalidParameter, "account into is required")
	}

	return nil
}
