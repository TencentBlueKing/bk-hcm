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

package cloudwatch

import "time"

// Dimension represents a CloudWatch dimension (Name + Value pair).
type Dimension struct {
	Name  string
	Value string
}

// MetricDataQuery represents a single metric data query aligned with
// the CloudWatch GetMetricData API structure.
type MetricDataQuery struct {
	ID         string
	Namespace  string
	MetricName string
	Dimensions []Dimension
	Stat       string
	Period     int64
}

// AwsGetMetricDataOption is the option for querying CloudWatch metric time-series data.
type AwsGetMetricDataOption struct {
	Region            string
	MetricDataQueries []MetricDataQuery
	StartTime         time.Time
	EndTime           time.Time
}

// MetricDataMessage holds a warning or error message associated with a metric data query.
type MetricDataMessage struct {
	Code  string
	Value string
}

// MetricDataResult holds the time-series result for a single metric query.
type MetricDataResult struct {
	ID         string
	Label      string
	StatusCode string
	Messages   []MetricDataMessage
	Timestamps []int64
	Values     []float64
}

// AwsListMetricsOption is the option for listing available CloudWatch metrics.
type AwsListMetricsOption struct {
	Region     string
	Namespace  string
	MetricName string
	Dimensions []Dimension
}
