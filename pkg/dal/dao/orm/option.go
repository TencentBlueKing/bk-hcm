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

package orm

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

type options struct {
	// ingressLimiter db request limiter.
	ingressLimiter *rate.Limiter
	// logLimiter write db request log limiter.
	logLimiter *rate.Limiter
	// mc db request metrics.
	mc *metric
	// slowRequestMS db slow request time, beyond this time, the db request will be logged. unit: millisecond
	slowRequestMS time.Duration
}

// Option orm option func defines.
type Option func(opt *options)

// IngressLimiter set db request limiter related params.
func IngressLimiter(qps, burst uint) Option {
	return func(opt *options) {
		opt.ingressLimiter = rate.NewLimiter(rate.Limit(qps), int(burst))
	}
}

// LogLimiter write db request log limiter related params.
func LogLimiter(qps, burst uint) Option {
	return func(opt *options) {
		opt.logLimiter = rate.NewLimiter(rate.Limit(qps), int(burst))
	}
}

// MetricsRegisterer set metrics registerer.
func MetricsRegisterer(register prometheus.Registerer) Option {
	return func(opt *options) {
		opt.mc = initMetric(register)
	}
}

// SlowRequestMS set db slow request time.
func SlowRequestMS(ms uint) Option {
	return func(opt *options) {
		opt.slowRequestMS = time.Duration(ms) * time.Millisecond
	}
}
