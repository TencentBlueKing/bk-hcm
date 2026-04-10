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

package instancetype

import (
	"encoding/json"
	"errors"
	"time"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// DimensionParam represents a CloudWatch dimension filter (Name + Value pair).
type DimensionParam struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// MetricDataQueryParam represents a single metric data query aligned with
// the CloudWatch GetMetricData API structure.
type MetricDataQueryParam struct {
	ID         string           `json:"id" validate:"required"`
	Namespace  string           `json:"namespace" validate:"required"`
	MetricName string           `json:"metric_name" validate:"required"`
	Dimensions []DimensionParam `json:"dimensions"`
	Stat       string           `json:"stat" validate:"required"`
	Period     int64            `json:"period" validate:"required,min=1"`
}

// AwsAssumeRoleGetMetricDataReq is the request for querying CloudWatch metric
// time-series data via AssumeRole cross-account access.
type AwsAssumeRoleGetMetricDataReq struct {
	RootAccountID     string                 `json:"root_account_id" validate:"required"`
	MainAccountID     string                 `json:"main_account_id" validate:"required"`
	RoleChain         []string               `json:"role_chain" validate:"required,min=1"`
	Region            string                 `json:"region" validate:"required"`
	ExternalID        string                 `json:"external_id,omitempty"`
	MetricDataQueries []MetricDataQueryParam `json:"metric_data_queries" validate:"required,min=1,dive"`
	StartTime         time.Time              `json:"start_time" validate:"required"`
	EndTime           time.Time              `json:"end_time" validate:"required"`
}

// Validate validates the request fields.
func (req *AwsAssumeRoleGetMetricDataReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	if !req.EndTime.After(req.StartTime) {
		return errors.New("end_time must be after start_time")
	}
	return nil
}

// MetricDataMessageItem holds a warning or error message associated with a metric data query.
type MetricDataMessageItem struct {
	Code  string `json:"code,omitempty"`
	Value string `json:"value,omitempty"`
}

// MetricDataResultItem holds the time-series result for a single metric query.
type MetricDataResultItem struct {
	ID         string                  `json:"id"`
	Label      string                  `json:"label,omitempty"`
	StatusCode string                  `json:"status_code,omitempty"`
	Messages   []MetricDataMessageItem `json:"messages,omitempty"`
	Timestamps []int64                 `json:"timestamps"`
	Values     []float64               `json:"values"`
}

// AwsAssumeRoleGetMetricDataResp wraps the GetMetricData response.
type AwsAssumeRoleGetMetricDataResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*MetricDataResultItem `json:"data"`
}

// AwsAssumeRoleListMetricsReq is the request for listing available CloudWatch
// metrics via AssumeRole cross-account access.
type AwsAssumeRoleListMetricsReq struct {
	RootAccountID string           `json:"root_account_id" validate:"required"`
	MainAccountID string           `json:"main_account_id" validate:"required"`
	RoleChain     []string         `json:"role_chain" validate:"required,min=1"`
	Region        string           `json:"region" validate:"required"`
	ExternalID    string           `json:"external_id,omitempty"`
	Namespace     string           `json:"namespace,omitempty"`
	MetricName    string           `json:"metric_name,omitempty"`
	Dimensions    []DimensionParam `json:"dimensions,omitempty"`
}

// Validate validates the request fields.
func (req *AwsAssumeRoleListMetricsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleListMetricsResp wraps the raw AWS CloudWatch ListMetrics response
// for transparent pass-through.
type AwsAssumeRoleListMetricsResp struct {
	rest.BaseResp `json:",inline"`
	Data          json.RawMessage `json:"data"`
}
