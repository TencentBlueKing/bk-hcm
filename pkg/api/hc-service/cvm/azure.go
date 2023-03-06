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

package cvm

import (
	"fmt"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// AzureOperateSyncReq cvm oprate sync request.
type AzureOperateSyncReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	Region            string   `json:"region" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"required"`
}

// Validate cvm operate sync request.
func (req *AzureOperateSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	if len(req.CloudIDs) <= 0 {
		return fmt.Errorf("operate sync count should > 0")
	}

	return validator.Validate.Struct(req)
}

// AzureDeleteReq define delete req.
type AzureDeleteReq struct {
	Force bool `json:"force" validate:"required"`
}

// Validate request.
func (req *AzureDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureStopReq azure stop req.
type AzureStopReq struct {
	SkipShutdown bool `json:"skip_shutdown" validate:"required"`
}

// Validate request.
func (req *AzureStopReq) Validate() error {
	return validator.Validate.Struct(req)
}
