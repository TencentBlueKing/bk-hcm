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

package monitoring

import (
	typesMonitoring "hcm/pkg/adaptor/types/monitoring"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// GcpListTimeSeriesReq defines the request for listing GCP time series.
type GcpListTimeSeriesReq struct {
	// RootAccountID is the root account ID for authentication.
	RootAccountID string `json:"root_account_id" validate:"required"`

	// MainAccountID is the main account ID for getting cloud project ID.
	MainAccountID string `json:"main_account_id" validate:"required"`

	// Embed the option struct for all time series parameters
	typesMonitoring.GcpListTimeSeriesOption
}

// Validate validates the GcpListTimeSeriesReq.
func (r *GcpListTimeSeriesReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	// Validate the embedded option
	if err := r.GcpListTimeSeriesOption.Validate(); err != nil {
		return err
	}

	return nil
}

// GcpListTimeSeriesResp defines the response for listing GCP time series.
type GcpListTimeSeriesResp struct {
	rest.BaseResp `json:",inline"`
	Data          *typesMonitoring.GcpListTimeSeriesResult `json:"data"`
}
