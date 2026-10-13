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
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// CLB submit metrics. See openspec/changes/add-sops-fine-grained-metrics/specs/spec.md
// for the canonical label set definition. Label cardinality is bounded
// because operation_type is an enum (~10 values), vendor is an enum (~5),
// and bkcc_biz_id is bounded by the number of subscribed business spaces.
type clbSubmitMetric struct {
	cost      *prometheus.HistogramVec
	total     *prometheus.CounterVec
	failTotal *prometheus.CounterVec
}

var (
	clbSubmitOnce sync.Once
	clbSubmit     *clbSubmitMetric
)

func initCLBSubmitMetric() {
	clbSubmitOnce.Do(func() {
		labels := prometheus.Labels{}
		base := []string{LabelBKCCBizID, "vendor", "operation_type"}
		failLabels := []string{LabelBKCCBizID, "vendor", "operation_type", "err_type"}

		m := &clbSubmitMetric{}
		m.cost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   Namespace,
			Subsystem:   CLBSubSys,
			Name:        "submit_cost_seconds",
			Help:        "the cost seconds of one CLB submit request (entry-to-flow-creation).",
			ConstLabels: labels,
			Buckets:     []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		}, base)
		Register().MustRegister(m.cost)

		m.total = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   Namespace,
			Subsystem:   CLBSubSys,
			Name:        "submit_total",
			Help:        "the total count of CLB submit requests by bkcc_biz_id / vendor / operation_type.",
			ConstLabels: labels,
		}, base)
		Register().MustRegister(m.total)

		m.failTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   Namespace,
			Subsystem:   CLBSubSys,
			Name:        "submit_fail_total",
			Help:        "the total count of failed CLB submit requests by bkcc_biz_id / vendor / operation_type / err_type.",
			ConstLabels: labels,
		}, failLabels)
		Register().MustRegister(m.failTotal)

		clbSubmit = m
	})
}

// EnsureCLBSubmitMetric is idempotent and registers the clb_submit_* metrics.
func EnsureCLBSubmitMetric() {
	initCLBSubmitMetric()
}

// ObserveCLBSubmit records one CLB submit observation: total, cost histogram,
// and (on failure) fail_total. errType==ErrTypeOK indicates success and skips
// the fail_total emission.
func ObserveCLBSubmit(bkBizID int64, vendor, operationType string, cost time.Duration, errType ErrType) {
	if clbSubmit == nil {
		initCLBSubmitMetric()
	}
	if vendor == "" {
		vendor = "unknown"
	}
	if operationType == "" {
		operationType = "unknown"
	}
	bizLabel := strconv.FormatInt(bkBizID, 10)
	labels := prometheus.Labels{
		LabelBKCCBizID:   bizLabel,
		"vendor":         vendor,
		"operation_type": operationType,
	}
	clbSubmit.total.With(labels).Inc()
	clbSubmit.cost.With(labels).Observe(cost.Seconds())
	if errType != ErrTypeOK {
		failLabels := prometheus.Labels{
			LabelBKCCBizID:   bizLabel,
			"vendor":         vendor,
			"operation_type": operationType,
			"err_type":       errType.String(),
		}
		clbSubmit.failTotal.With(failLabels).Inc()
	}
}
