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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	cvt "hcm/pkg/tools/converter"
)

// GetMonitorDataReq defines request to get monitor data.
type GetMonitorDataReq struct {
	MetricName string      `json:"metric_name" validate:"required"`
	Period     int64       `json:"period" validate:"required"`
	StartTime  interface{} `json:"start_time" validate:"omitempty"` // DateTimeLayout string for tcloud, unix ms for huawei
	EndTime    interface{} `json:"end_time" validate:"omitempty"`   // DateTimeLayout string for tcloud, unix ms for huawei
	Namespace  string      `json:"namespace" validate:"omitempty"`  // used by vendor=huawei, e.g. SYS.ECS/SYS.VPC
	Filter     string      `json:"filter" validate:"omitempty"`     // used by vendor=huawei, e.g. average/max/min
	IDs        []string    `json:"ids" validate:"required,min=1"`
}

// Validate request.
func (req *GetMonitorDataReq) Validate(vendor enumor.Vendor) error {
	if len(req.IDs) > constant.MonitorMaxInstanceLimit {
		return fmt.Errorf("instances count should <= %d", constant.MonitorMaxInstanceLimit)
	}

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.Period <= 0 {
		return fmt.Errorf("period should > 0")
	}

	switch vendor {
	case enumor.TCloud:
		if req.Period < 60 {
			return fmt.Errorf("period should >= 60 for vendor %s", vendor)
		}
		if _, _, err := req.GetTCloudTimeRange(); err != nil {
			return fmt.Errorf("start_time and end_time are required for vendor %s", vendor)
		}
	case enumor.HuaWei:
		startTime, endTime, err := req.GetHuaWeiTimeRange()
		if err != nil {
			return fmt.Errorf("start_time and end_time are required for vendor %s", vendor)
		}
		if startTime <= 0 || endTime <= 0 {
			return fmt.Errorf("start_time and end_time should > 0 for vendor %s", vendor)
		}
		if startTime >= endTime {
			return fmt.Errorf("start_time should < end_time for vendor %s", vendor)
		}
	}

	return nil
}

// GetTCloudTimeRange parses tcloud time range (layout string).
func (req *GetMonitorDataReq) GetTCloudTimeRange() (string, string, error) {
	startTime, ok := req.StartTime.(string)
	if !ok || len(startTime) == 0 {
		return "", "", fmt.Errorf("invalid start_time for tcloud")
	}

	endTime, ok := req.EndTime.(string)
	if !ok || len(endTime) == 0 {
		return "", "", fmt.Errorf("invalid end_time for tcloud")
	}

	return startTime, endTime, nil
}

// GetHuaWeiTimeRange parses huawei time range (unix ms).
func (req *GetMonitorDataReq) GetHuaWeiTimeRange() (int64, int64, error) {
	startTime, err := cvt.ParseTimeToInt64(req.StartTime)
	if err != nil {
		return 0, 0, fmt.Errorf("parse huawei start_time failed: %w", err)
	}

	endTime, err := cvt.ParseTimeToInt64(req.EndTime)
	if err != nil {
		return 0, 0, fmt.Errorf("parse huawei end_time failed: %w", err)
	}

	return startTime, endTime, nil
}

// GetMonitorDataResp defines response of get monitor data.
type GetMonitorDataResp struct {
	DataPoints []*MonitorDataPointResp `json:"data_points"`
}

// MonitorDataPointResp defines a single monitor data point response.
type MonitorDataPointResp struct {
	ID         string                 `json:"id"`
	IP         []string               `json:"ip"`
	Region     string                 `json:"region"`
	InstanceID string                 `json:"instance_id"`
	Timestamps []int64                `json:"timestamps"`
	Values     []float64              `json:"values"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}
