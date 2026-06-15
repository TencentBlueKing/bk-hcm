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

package application

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// SubAccountSecretStatusBatchUpdateReq define sub account secret status batch update request.
type SubAccountSecretStatusBatchUpdateReq struct {
	SubAccountBaseReq `json:",inline"`
	SubAccountSecrets []SubAccountSecretStatusUpdateReq `json:"sub_account_secrets" validate:"required,min=1,max=100"`
}

// Validate sub account secret status batch update request.
func (req *SubAccountSecretStatusBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ValidateBase(); err != nil {
		return err
	}

	for i, item := range req.SubAccountSecrets {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("sub_account_secrets[%d] validate failed, err: %w", i, err)
		}
	}

	return nil
}

// SubAccountSecretStatusUpdateReq define single sub account secret status update request.
type SubAccountSecretStatusUpdateReq struct {
	ID     string                        `json:"id" validate:"required"`
	Status enumor.SubAccountSecretStatus `json:"status" validate:"required"`
}

// Validate sub account secret status update request.
func (req *SubAccountSecretStatusUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return req.Status.Validate()
}
