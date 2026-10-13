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

package consumer

import (
	"strconv"

	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// initMetric registers the async consumer's metrics.
func initMetric(register prometheus.Registerer) *metric {
	m := new(metric)
	labels := prometheus.Labels{}

	initLegacyMetrics(m, register, labels)
	initExecutionMetrics(m, register, labels)

	return m
}

func initLegacyMetrics(m *metric, register prometheus.Registerer, labels prometheus.Labels) {
	// 监控taskInitQueueSize
	m.taskInitQueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.AsyncSubSys,
			Name:        "task_init_queue_size",
			Help:        "Current size of the task init queue",
			ConstLabels: labels,
		},
		[]string{"queue_name"},
	)
	register.MustRegister(m.taskInitQueueSize)

	// 监控当前运行中的各flowType数量
	m.flowTypeRunningNum = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.AsyncSubSys,
			Name:        "flow_type_running_num",
			Help:        "Number of running flows by type",
			ConstLabels: labels,
		},
		[]string{"flowType"},
	)
	register.MustRegister(m.flowTypeRunningNum)

	// 监控各flowType实际运行时间(包括了等待执行时间以及协程池阻塞情况)
	m.flowTypeExecTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.AsyncSubSys,
			Name:        "flow_type_exec_duration_milliseconds",
			Help:        "Execution duration of flows by type in milliseconds",
			ConstLabels: labels,
			Buckets: []float64{1, 2, 3, 4, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 400, 800, 1000, 1500,
				6000},
		},
		[]string{"flowType"},
	)
	register.MustRegister(m.flowTypeExecTime)
}

func initExecutionMetrics(m *metric, register prometheus.Registerer, labels prometheus.Labels) {
	// New: flow execution cost / failure metrics.
	flowLabels := []string{metrics.LabelBKCCBizID, metrics.LabelVendor, metrics.LabelOperation, metrics.LabelFlowName,
		metrics.LabelState}
	flowFailLabels := append(flowLabels, metrics.LabelErrType)
	m.flowExecCostSec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.AsyncSubSys,
			Name:        "flow_exec_cost_seconds",
			Help:        "End-to-end flow execution cost in seconds (entry to terminal state).",
			ConstLabels: labels,
			Buckets:     []float64{0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
		},
		flowLabels,
	)
	register.MustRegister(m.flowExecCostSec)

	m.flowFailTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.AsyncSubSys,
			Name:      "flow_fail_total",
			Help: "Total count of flows ending in non-success terminal state, by business dimensions, flow_name, " +
				"state and err_type.",
			ConstLabels: labels,
		},
		flowFailLabels,
	)
	register.MustRegister(m.flowFailTotal)

	// New: task action execution cost / failure metrics.
	taskLabels := []string{metrics.LabelBKCCBizID, metrics.LabelVendor, metrics.LabelOperation, metrics.LabelFlowName,
		metrics.LabelActionName, metrics.LabelState}
	taskFailLabels := append(taskLabels, metrics.LabelErrType)
	m.taskExecCostSec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.AsyncSubSys,
			Name:        "task_exec_cost_seconds",
			Help:        "Single task action attempt execution cost in seconds (act.Run wall time).",
			ConstLabels: labels,
			Buckets:     []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 300, 600},
		},
		taskLabels,
	)
	register.MustRegister(m.taskExecCostSec)

	m.taskFailTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.AsyncSubSys,
			Name:        "task_fail_total",
			Help:        "Total count of task action attempts ending in non-success state, by err_type.",
			ConstLabels: labels,
		},
		taskFailLabels,
	)
	register.MustRegister(m.taskFailTotal)
}

type metric struct {
	taskInitQueueSize  *prometheus.GaugeVec
	flowTypeRunningNum *prometheus.GaugeVec
	flowTypeExecTime   *prometheus.HistogramVec

	// Flow level: end-to-end cost & failure (terminal states only).
	flowExecCostSec *prometheus.HistogramVec
	flowFailTotal   *prometheus.CounterVec

	// Task level: per-attempt cost & failure (act.Run wall time).
	taskExecCostSec *prometheus.HistogramVec
	taskFailTotal   *prometheus.CounterVec
}

type shareDataMetricDims struct {
	bkBizID      int64
	bkBizIDValid bool
	vendor       string
	operation    string
}

func (dims shareDataMetricDims) bkBizIDLabel() string {
	if !dims.bkBizIDValid {
		return "unknown"
	}
	return strconv.FormatInt(dims.bkBizID, 10)
}

func getShareDataMetricDims(shareData *tableasync.ShareData) shareDataMetricDims {
	dims := shareDataMetricDims{
		vendor:    "unknown",
		operation: "unknown",
	}
	if shareData == nil {
		return dims
	}
	if bkBizIDStr, ok := shareData.Get(tableasync.ShareDataKeyBkBizID); ok && bkBizIDStr != "" {
		if bkBizID, err := strconv.ParseInt(bkBizIDStr, 10, 64); err == nil {
			dims.bkBizID = bkBizID
			dims.bkBizIDValid = true
		}
	}
	if vendor, ok := shareData.Get(tableasync.ShareDataKeyVendor); ok && vendor != "" {
		dims.vendor = vendor
	}
	if operation, ok := shareData.Get(tableasync.ShareDataKeyOperationType); ok && operation != "" {
		dims.operation = operation
	}
	return dims
}
