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

package accountsecret

import (
	"encoding/json"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// AccountSecretUpdateReq defines account secret update request.
type AccountSecretUpdateReq struct {
	Type      *enumor.AccountSecretType `json:"type,omitempty"`
	Extension *json.RawMessage          `json:"extension,omitempty"`
}

// Validate account secret update request.
func (req *AccountSecretUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.Type != nil {
		if err := req.Type.Validate(); err != nil {
			return err
		}
	}

	return nil
}
