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

package metrics

import (
	"sync"
	"time"

	"hcm/pkg/cc"

	"github.com/prometheus/client_golang/prometheus"
)

// Component label constants for the http_request_* metrics.
//
// These values populate the `component` label. To avoid drift between the
// metric label and the runtime service identity (cc.ServiceName()), the
// service-bound constants are derived directly from cc.Name; the compiler
// will reject any future rename in cc.Name that diverges from these values.
//
// ComponentAdaptor and ComponentAgent stay as standalone literals because
// they have no corresponding cc.Name entry today (adaptor is not a service;
// agent-server is not yet defined in cc).
const (
	ComponentAPIServer   = string(cc.APIServerName)
	ComponentCloudServer = string(cc.CloudServerName)
	ComponentDataService = string(cc.DataServiceName)
	ComponentHCService   = string(cc.HCServiceName)
	ComponentTaskServer  = string(cc.TaskServerName)
	ComponentAuthServer  = string(cc.AuthServerName)
	ComponentWebServer   = string(cc.WebServerName)
	ComponentAccount     = string(cc.AccountServerName)
	ComponentAgent       = string(cc.AgentServerName)
	ComponentAdaptor     = "adaptor"
)

// VendorNone is the value used by the `vendor` label when the打点 happens on
// a service-side HTTP entry point that is not bound to a specific cloud
// vendor (e.g. api-server / hc-service / data-service request entry).
const VendorNone = "none"

// Method label constants used by adaptor cloud API打点. Service-side HTTP
// metrics use the actual HTTP verb (GET/POST/PUT/...).
const (
	MethodSDK  = "SDK"
	MethodCALL = "CALL"
)

// httpRequestMetric holds the unified http_request_* metrics.
//
// Same metric name MUST always be emitted with the same label set across
// every component, otherwise prometheus will reject the registration on
// other side. The label sets are:
//
//	hcm_http_request_cost_seconds: component, endpoint, method, vendor
//	hcm_http_request_total       : component, endpoint, method, vendor
//	hcm_http_request_fail_total  : component, endpoint, method, vendor, err_type
//
// `endpoint` MUST be a low-cardinality value: a route template for HTTP
// services (e.g. "/api/v1/cloud/foo/{id}") or a stable cloud API name for
// adaptor (e.g. "DescribeLoadBalancers"). Raw URL or rid MUST NOT be used.
type httpRequestMetric struct {
	cost      *prometheus.HistogramVec
	total     *prometheus.CounterVec
	failTotal *prometheus.CounterVec
}

var (
	httpReqOnce sync.Once
	httpReq     *httpRequestMetric
)

// initHTTPRequestMetric registers the unified http_request_* metrics on the
// global registerer. Safe to call multiple times - it only registers once
// per process. Service main() / Init code should call EnsureHTTPRequestMetric
// after metrics.InitMetrics(), but registration is also lazily triggered by
// the helper observe functions to be defensive.
func initHTTPRequestMetric() {
	httpReqOnce.Do(func() {
		labels := prometheus.Labels{}

		base := []string{"component", "endpoint", "method", "vendor"}
		failLabels := []string{"component", "endpoint", "method", "vendor", "err_type"}

		m := &httpRequestMetric{}
		m.cost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   Namespace,
			Subsystem:   HTTPSubSys,
			Name:        "request_cost_seconds",
			Help:        "the cost seconds of an http request (server entry or adaptor cloud API call).",
			ConstLabels: labels,
			// Buckets cover both fast in-process service entries and slow
			// cloud API calls (up to ~30s). Aligned with cloudapi/lag_seconds.
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 10, 20, 30},
		}, base)
		Register().MustRegister(m.cost)

		m.total = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   Namespace,
			Subsystem:   HTTPSubSys,
			Name:        "request_total",
			Help:        "the total count of http requests (server entry or adaptor cloud API call).",
			ConstLabels: labels,
		}, base)
		Register().MustRegister(m.total)

		m.failTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   Namespace,
			Subsystem:   HTTPSubSys,
			Name:        "request_fail_total",
			Help:        "the total count of failed http requests (server entry or adaptor cloud API call), by err_type.",
			ConstLabels: labels,
		}, failLabels)
		Register().MustRegister(m.failTotal)

		httpReq = m
	})
}

// EnsureHTTPRequestMetric ensures the http_request_* metrics are registered.
// It is safe to call from package init paths; it is idempotent.
func EnsureHTTPRequestMetric() {
	initHTTPRequestMetric()
}

// ObserveHTTPRequest records one http_request_total + http_request_cost_seconds
// sample. If err_type != ErrTypeOK an extra http_request_fail_total sample is
// emitted as well.
//
// component MUST be one of the Component* constants. endpoint MUST be a
// route template (HTTP services) or a stable cloud API name (adaptor).
// method MUST be the HTTP verb for service-side, or MethodSDK / MethodCALL
// for adaptor. vendor MUST be VendorNone for service-side, or the real cloud
// vendor for adaptor.
func ObserveHTTPRequest(component, endpoint, method, vendor string, cost time.Duration, errType ErrType) {
	if httpReq == nil {
		initHTTPRequestMetric()
	}
	if endpoint == "" {
		// fall back to a stable placeholder; never emit empty label values.
		endpoint = "unknown"
	}
	if vendor == "" {
		vendor = VendorNone
	}
	labels := prometheus.Labels{
		"component": component,
		"endpoint":  endpoint,
		"method":    method,
		"vendor":    vendor,
	}
	httpReq.total.With(labels).Inc()
	httpReq.cost.With(labels).Observe(cost.Seconds())
	if errType != ErrTypeOK {
		failLabels := prometheus.Labels{
			"component": component,
			"endpoint":  endpoint,
			"method":    method,
			"vendor":    vendor,
			"err_type":  errType.String(),
		}
		httpReq.failTotal.With(failLabels).Inc()
	}
}
