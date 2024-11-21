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

// GetTCloudRecordRoundTripper get record round tripper for tcloud
func GetTCloudRecordRoundTripper(next http.RoundTripper) promhttp.RoundTripperFunc {
	if next == nil {
		next = http.DefaultTransport
	}
	return func(req *http.Request) (*http.Response, error) {
		action := strings.Join(req.Header["X-TC-Action"], ",")
		region := strings.Join(req.Header["X-TC-Region"], ",")
		start := time.Now()
		code := "nil"
		ret, err := next.RoundTrip(req)
		if ret != nil {
			code = ret.Status
		}

		if err != nil || (ret != nil && ret.StatusCode != http.StatusOK) {
			cloudApiMetric.errCounter.With(prometheus.Labels{
				"vendor":    string(enumor.TCloud),
				"endpoint":  req.Host,
				"region":    region,
				"api_name":  action,
				"http_code": code,
			}).Inc()
		}
		cost := time.Since(start).Seconds()
		cloudApiMetric.lagSec.With(
			prometheus.Labels{
				"vendor":    string(enumor.TCloud),
				"endpoint":  req.Host,
				"region":    region,
				"api_name":  action,
				"http_code": code,
			}).Observe(cost)
		return ret, err
	}
}
