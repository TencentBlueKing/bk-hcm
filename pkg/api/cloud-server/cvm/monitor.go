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

package cscvm

import (
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// GetMonitorDataReq defines request to get monitor data.
type GetMonitorDataReq struct {
	MetricName string   `json:"metric_name" validate:"required"`
	Period     int64    `json:"period" validate:"required"`
	StartTime  string   `json:"start_time" validate:"required"` // DateTimeLayout format: 2006-01-02 15:04:05
	EndTime    string   `json:"end_time" validate:"required"`   // DateTimeLayout format: 2006-01-02 15:04:05
	IDs        []string `json:"ids" validate:"required,min=1"`
}

// Validate request.
func (req *GetMonitorDataReq) Validate() error {
	if len(req.IDs) > constant.MonitorMaxInstanceLimit {
		return fmt.Errorf("instances count should <= %d", constant.MonitorMaxInstanceLimit)
	}

	return validator.Validate.Struct(req)
}

// GetMonitorDataResp defines response of get monitor data.
type GetMonitorDataResp struct {
	DataPoints []*MonitorDataPointResp `json:"data_points"`
}

// MonitorDataPointResp defines a single monitor data point response.
type MonitorDataPointResp struct {
	ID         string    `json:"id"`
	IP         []string  `json:"ip"`
	Region     string    `json:"region"`
	InstanceID string    `json:"instance_id"`
	Timestamps []int64   `json:"timestamps"`
	Values     []float64 `json:"values"`
}
