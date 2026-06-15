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

package hccvm

import (
	"fmt"
	"strings"
	"time"

	typecvm "hcm/pkg/adaptor/types/cvm"
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
	SkipShutdown bool `json:"skip_shutdown" validate:"omitempty"`
}

// Validate request.
func (req *AzureStopReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureCreateReq azure create req.
type AzureCreateReq struct {
	AccountID            string                  `json:"account_id" validate:"required"`
	ResourceGroupName    string                  `json:"resource_group_name" validate:"required"`
	Region               string                  `json:"region" validate:"required"`
	Name                 string                  `json:"name" validate:"required"`
	Zones                []string                `json:"zones" validate:"omitempty"`
	InstanceType         string                  `json:"instance_type" validate:"required"`
	CloudImageID         string                  `json:"cloud_image_id" validate:"required"`
	Username             string                  `json:"username" validate:"required"`
	Password             string                  `json:"password" validate:"required"`
	CloudSubnetID        string                  `json:"cloud_subnet_id" validate:"required"`
	CloudSecurityGroupID string                  `json:"cloud_security_group_id" validate:"required"`
	OSDisk               *typecvm.AzureOSDisk    `json:"os_disk" validate:"required"`
	DataDisk             []typecvm.AzureDataDisk `json:"data_disk" validate:"omitempty"`
	PublicIPAssigned     bool                    `json:"public_ip_assigned" validate:"omitempty"`
}

// Validate request.
func (req *AzureCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureCreateResp define azure create resp.
type AzureCreateResp struct {
	CloudID string `json:"cloud_id"`
}

// AzureMonitorDataReq defines request to get azure monitor data.
type AzureMonitorDataReq struct {
	AccountID           string   `json:"account_id" validate:"required"`
	Region              string   `json:"region" validate:"required"`
	MetricName          string   `json:"metric_name" validate:"required"`
	Period              int64    `json:"period" validate:"required,min=1"`
	StartTime           string   `json:"start_time" validate:"required"` // RFC3339 UTC
	EndTime             string   `json:"end_time" validate:"required"`   // RFC3339 UTC
	MetricNamespace     string   `json:"metric_namespace" validate:"omitempty"`
	Aggregation         string   `json:"aggregation" validate:"omitempty"`
	AutoAdjustTimegrain *bool    `json:"auto_adjust_timegrain" validate:"omitempty"`
	Top                 *int32   `json:"top" validate:"omitempty"`
	OrderBy             string   `json:"orderby" validate:"omitempty"`
	Filter              string   `json:"filter" validate:"omitempty"`
	ResultType          string   `json:"result_type" validate:"omitempty"`
	InstanceIDs         []string `json:"instance_ids" validate:"required,min=1,max=20"`
}

// Validate request.
func (req *AzureMonitorDataReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return fmt.Errorf("start_time should be RFC3339 format")
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return fmt.Errorf("end_time should be RFC3339 format")
	}
	if !startTime.Before(endTime) {
		return fmt.Errorf("start_time should < end_time")
	}
	_, startOffset := startTime.Zone()
	_, endOffset := endTime.Zone()
	if startOffset != 0 || endOffset != 0 {
		return fmt.Errorf("start_time and end_time should be UTC timezone")
	}
	if req.Top != nil && *req.Top <= 0 {
		return fmt.Errorf("top should > 0")
	}
	if len(req.OrderBy) != 0 && req.Top == nil {
		return fmt.Errorf("orderby requires top")
	}
	if len(req.ResultType) != 0 {
		resultType := strings.ToLower(req.ResultType)
		if resultType != "data" && resultType != "metadata" {
			return fmt.Errorf("result_type should be one of: Data, Metadata")
		}
	}
	if len(req.InstanceIDs) > constant.MonitorMaxInstanceLimit {
		return fmt.Errorf("instances count should <= %d", constant.MonitorMaxInstanceLimit)
	}

	return nil
}

// AzureMonitorDataResp defines response of azure monitor data.
type AzureMonitorDataResp struct {
	DataPoints []*MonitorDataPointResp `json:"data_points"`
}
