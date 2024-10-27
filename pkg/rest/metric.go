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

package rest

import (
	"net/http"
	"strings"
	"time"

	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// restMetric is used to collect restfull metrics.
var restMetric *metric

func initMetric() {
	m := new(metric)
	labels := prometheus.Labels{}

	m.lagMS = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.RestfulSubSys,
		Name:        "lag_milliseconds",
		Help:        "the lags(milliseconds) to request the restful API",
		ConstLabels: labels,
		Buckets:     []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 3, 4, 5, 10, 30, 50, 100},
	}, []string{"alias", "biz"})
	metrics.Register().MustRegister(m.lagMS)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.RestfulSubSys,
			Name:        "total_err_count",
			Help:        "the total error count to request the restful API",
			ConstLabels: labels,
		}, []string{"alias", "biz"})
	metrics.Register().MustRegister(m.errCounter)
	ThirdLags = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   "third",
		Name:        "lag_seconds",
		Help:        "请求第三方API延迟信息",
		ConstLabels: labels,
		Buckets:     []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 3, 4, 5, 10, 30, 50, 100},
	}, []string{"method", "code", "action", "region", "target"})
	metrics.Register().MustRegister(ThirdLags)

	if common.DefaultHttpClient == nil {
		transport := &http.Transport{}
		common.DefaultHttpClient = &http.Client{Transport: recordRoundTrip(transport)}
	}
	restMetric = m
}

type metric struct {
	// lagMS record the cost time of request the restful API.
	lagMS *prometheus.HistogramVec

	// errCounter record the total error count when request restful API.
	errCounter *prometheus.CounterVec
}

// ThirdLags 请求外部的延迟记录
var ThirdLags *prometheus.HistogramVec

func recordRoundTrip(next http.RoundTripper) promhttp.RoundTripperFunc {
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
		ThirdLags.
			With(prometheus.Labels{
				"action": action,
				"method": req.Method,
				"code":   code,
				"region": region,
				"target": req.Host}).
			Observe(cost)
		return ret, err
	}
}
