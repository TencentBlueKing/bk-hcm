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

package metric

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/metrics"
)

// restMetric is used to collect cloud api metrics.
var cloudApiMetric *metric

func init() {
	m := new(metric)
	labels := prometheus.Labels{}

	m.lagMS = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CloudApiSubSys,
		Name:        "lag_seconds",
		Help:        "the lag seconds to request the cloud API",
		ConstLabels: labels,
		Buckets:     []float64{0.05, 0.075, 0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.7, 1, 2, 3, 4, 5, 10, 20, 30},
	}, []string{"vendor", "http_code", "api_name", "region", "endpoint"})
	metrics.Register().MustRegister(m.lagMS)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.RestfulSubSys,
			Name:        "total_err_count",
			Help:        "the total error count to request the restful API",
			ConstLabels: labels,
		}, []string{"vendor", "http_code", "api_name", "region", "endpoint"})
	metrics.Register().MustRegister(m.errCounter)

	cloudApiMetric = m
}

type metric struct {
	// lagMS record the cost time of request cloud API.
	lagMS *prometheus.HistogramVec

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
		cost := time.Since(start).Seconds()
		cloudApiMetric.lagMS.With(
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
