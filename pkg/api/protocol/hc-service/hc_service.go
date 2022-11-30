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
	"errors"

	"hcm/pkg/criteria/enumor"
)

// AccountCheckReq defines account check request.
// TODO: 结构体信息随便定义，之后自行替换即可
type AccountCheckReq struct {
	Vendor    enumor.Vendor `json:"vendor"`
	SecretID  string        `json:"secret_id"`
	SecretKey string        `json:"secret_key"`
}

// Validate account check req.
func (req AccountCheckReq) Validate() error {
	if err := req.Vendor.Validate(); err != nil {
		return err
	}

	if len(req.SecretID) == 0 {
		return errors.New("secret id is required")
	}

	if len(req.SecretKey) == 0 {
		return errors.New("secret key is required")
	}

	return nil
}
