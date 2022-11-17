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

package client

import (
	"time"

	"hcm/pkg/rest/discovery"

	"github.com/prometheus/client_golang/prometheus"
)

// Capability http request limit.
type Capability struct {
	// Client http client.
	Client HTTPClient

	// Discover get request address.
	Discover discovery.Interface

	// the max tolerance api request latency time, if exceeded this time, then
	// this request will be logged and warned.
	ToleranceLatencyTime time.Duration

	// MetricOpts metric option.
	MetricOpts MetricOption
}

// MetricOption metrics options.
type MetricOption struct {
	// prometheus metric register
	Register prometheus.Registerer
	// if not set, use default buckets value
	DurationBuckets []float64
}
