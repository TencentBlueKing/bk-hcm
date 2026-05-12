/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package metric is used to collect cloud api metrics.
package metric

import (
	"net/http"
	"strings"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// restMetric is used to collect cloud api metrics.
var cloudApiMetric *metric

// InitCloudApiMetrics ..
func InitCloudApiMetrics(reg prometheus.Registerer) {
	m := new(metric)

	labels := prometheus.Labels{}

	m.lagSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CloudApiSubSys,
		Name:        "lag_seconds",
		Help:        "the lag seconds to request the cloud API",
		ConstLabels: labels,
		Buckets:     []float64{0.05, 0.075, 0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.7, 1, 2, 3, 4, 5, 10, 20, 30},
	}, []string{"vendor", "http_code", "api_name", "region", "endpoint"})
	reg.MustRegister(m.lagSec)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.CloudApiSubSys,
			Name:        "total_err_count",
			Help:        "the total error count to request the restful API",
			ConstLabels: labels,
		}, []string{"vendor", "http_code", "api_name", "region", "endpoint"})
	reg.MustRegister(m.errCounter)

	cloudApiMetric = m
}

type metric struct {
	// lagSec record the cost time of request cloud API.
	lagSec *prometheus.HistogramVec

	// errCounter record the total error count request cloud API.
	errCounter *prometheus.CounterVec
}

// GetTCloudRecordRoundTripper get record round tripper for tcloud.
//
// This RoundTripper is wired into the TCloud SDK's HTTP transport (see
// pkg/adaptor/tcloud/client.go), so the vendor dimension is fixed to
// enumor.TCloud by design. For other clouds (huawei / aws / azure / gcp)
// that don't expose a RoundTripper hook, use ObserveCloudAPI directly at
// the SDK call site with the matching enumor.Vendor.
func GetTCloudRecordRoundTripper(next http.RoundTripper) promhttp.RoundTripperFunc {
	if next == nil {
		next = http.DefaultTransport
	}
	const vendor = string(enumor.TCloud)
	return func(req *http.Request) (*http.Response, error) {
		action := strings.Join(req.Header["X-TC-Action"], ",")
		region := strings.Join(req.Header["X-TC-Region"], ",")
		start := time.Now()
		code := "nil"
		ret, err := next.RoundTrip(req)
		if ret != nil {
			code = ret.Status
		}

		labels := prometheus.Labels{
			"vendor":    vendor,
			"endpoint":  req.Host,
			"region":    region,
			"api_name":  action,
			"http_code": code,
		}
		if err != nil || (ret != nil && ret.StatusCode != http.StatusOK) {
			cloudApiMetric.errCounter.With(labels).Inc()
		}
		cost := time.Since(start)
		cloudApiMetric.lagSec.With(labels).Observe(cost.Seconds())

		// Also emit the unified hcm_http_request_* metrics so cross-vendor
		// dashboards can consume a single metric family. We intentionally
		// keep the cloudapi_* metrics above for backward compatibility (kept
		// label set unchanged) while adding the unified family with a
		// bounded label cardinality (no http_code / region label here).
		ObserveCloudAPI(vendor, action, cost, classifyCloudErr(err, ret))
		return ret, err
	}
}

// ObserveCloudAPI records a unified hcm_http_request_* sample for one
// adaptor cloud API call. It can be called directly by SDK call-site
// instrumentation in adaptors that don't use a RoundTripper hook (e.g.
// huawei / aws / azure / gcp).
//
// vendor MUST be one of enumor.Vendor values (e.g. "tcloud", "huawei").
// apiName MUST be a stable cloud API name (e.g. "DescribeLoadBalancers").
// Raw URL paths or vendor-specific opaque ids MUST NOT be used.
func ObserveCloudAPI(vendor, apiName string, cost time.Duration, errType metrics.ErrType) {
	metrics.ObserveHTTPRequest(metrics.ComponentAdaptor, apiNameOrUnknown(apiName),
		metrics.MethodSDK, vendor, cost, errType)
}

func apiNameOrUnknown(name string) string {
	if name == "" {
		return "unknown"
	}
	return name
}

// classifyCloudErr maps the (err, http.Response) pair from a RoundTrip into a
// normalized metrics.ErrType. Non-2xx responses are classified as cloud_error
// because the cloud API returned a structured error envelope; transport
// errors fall through metrics.ClassifyError for timeout / network detection.
func classifyCloudErr(err error, resp *http.Response) metrics.ErrType {
	if err != nil {
		return metrics.ClassifyError(err)
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		return metrics.ErrTypeCloudError
	}
	return metrics.ErrTypeOK
}
