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

// Task progress metrics: hcm_async_task_manage_* and hcm_async_task_detail_*.
//
// First-iteration scope (per design.md "Resolved Clarifications"):
//   - only end-state cost and failure counters are emitted;
//   - status Gauge / periodic aggregation are deferred;
//   - task_detail cost uses (updated_at - created_at) by callers.
//
// Label cardinality is intentionally bounded: bkcc_biz_id (subscribed
// business spaces), vendor (~5 enum values), operation
// (enumor.TaskOperation enum, ~10 values), state (terminal enum, ~5 values),
// err_type (closed enum).
// task_action_id is intentionally NOT used as a label because it is a UUID
// allocated per task instance and would explode label cardinality; the
// `operation` label provides the equivalent type-level grouping.
type taskProgressMetric struct {
	manageCost *prometheus.HistogramVec
	manageFail *prometheus.CounterVec
	detailCost *prometheus.HistogramVec
	detailFail *prometheus.CounterVec
}

var (
	taskProgressOnce sync.Once
	taskProgress     *taskProgressMetric
)

func initTaskProgressMetric() {
	taskProgressOnce.Do(func() {
		labels := prometheus.Labels{}
		base := []string{LabelBKCCBizID, LabelVendor, LabelOperation, LabelState}
		failLabels := []string{LabelBKCCBizID, LabelVendor, LabelOperation, LabelState, LabelErrType}
		costBuckets := []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600, 7200}

		m := &taskProgressMetric{}
		m.manageCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   Namespace,
			Subsystem:   AsyncSubSys,
			Name:        "task_manage_exec_cost_seconds",
			Help:        "End-to-end task management execution cost in seconds (created_at to terminal state).",
			ConstLabels: labels,
			Buckets:     costBuckets,
		}, base)
		Register().MustRegister(m.manageCost)

		m.manageFail = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   Namespace,
			Subsystem:   AsyncSubSys,
			Name:        "task_manage_fail_total",
			Help:        "Total count of task managements ending in non-success terminal state.",
			ConstLabels: labels,
		}, failLabels)
		Register().MustRegister(m.manageFail)

		m.detailCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   Namespace,
			Subsystem:   AsyncSubSys,
			Name:        "task_detail_exec_cost_seconds",
			Help:        "Per-detail task execution cost in seconds (updated_at - created_at).",
			ConstLabels: labels,
			Buckets:     costBuckets,
		}, base)
		Register().MustRegister(m.detailCost)

		m.detailFail = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   Namespace,
			Subsystem:   AsyncSubSys,
			Name:        "task_detail_fail_total",
			Help:        "Total count of task details ending in non-success terminal state.",
			ConstLabels: labels,
		}, failLabels)
		Register().MustRegister(m.detailFail)

		taskProgress = m
	})
}

// EnsureTaskProgressMetric is idempotent and registers the task_progress metrics.
func EnsureTaskProgressMetric() {
	initTaskProgressMetric()
}

// ObserveTaskManagement records one task management observation. cost is the
// elapsed time from management creation until it reached the given terminal
// state. errType is ignored when state indicates success; when state is a
// non-success terminal value (e.g. failed / cancel / deliver_partial) the
// fail_total counter is incremented as well.
func ObserveTaskManagement(bkBizID int64, vendor, operation, state string, cost time.Duration, errType ErrType) {
	if taskProgress == nil {
		initTaskProgressMetric()
	}
	bizLabel := strconv.FormatInt(bkBizID, 10)
	vendor = nonEmpty(vendor)
	operation = nonEmpty(operation)
	state = nonEmpty(state)

	labels := prometheus.Labels{
		LabelBKCCBizID: bizLabel,
		LabelVendor:    vendor,
		LabelOperation: operation,
		LabelState:     state,
	}
	taskProgress.manageCost.With(labels).Observe(cost.Seconds())
	if errType != ErrTypeOK {
		failLabels := prometheus.Labels{
			LabelBKCCBizID: bizLabel,
			LabelVendor:    vendor,
			LabelOperation: operation,
			LabelState:     state,
			LabelErrType:   errType.String(),
		}
		taskProgress.manageFail.With(failLabels).Inc()
	}
}

// ObserveTaskDetail records one task detail observation. cost is intended to
// be (updated_at - created_at) of the detail row.
func ObserveTaskDetail(bkBizID int64, vendor, operation, state string, cost time.Duration, errType ErrType) {
	if taskProgress == nil {
		initTaskProgressMetric()
	}
	bizLabel := strconv.FormatInt(bkBizID, 10)
	vendor = nonEmpty(vendor)
	operation = nonEmpty(operation)
	state = nonEmpty(state)

	labels := prometheus.Labels{
		LabelBKCCBizID: bizLabel,
		LabelVendor:    vendor,
		LabelOperation: operation,
		LabelState:     state,
	}
	taskProgress.detailCost.With(labels).Observe(cost.Seconds())
	if errType != ErrTypeOK {
		failLabels := prometheus.Labels{
			LabelBKCCBizID: bizLabel,
			LabelVendor:    vendor,
			LabelOperation: operation,
			LabelState:     state,
			LabelErrType:   errType.String(),
		}
		taskProgress.detailFail.With(failLabels).Inc()
	}
}

func nonEmpty(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
