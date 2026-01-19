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

// Package metric ...
package metric

import (
	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// TaskName task name
	TaskName = "task_name"
)

var (
	execCounter  *prometheus.CounterVec
	execError    *prometheus.CounterVec
	execDuration *prometheus.HistogramVec
)

// Init initial the metric
func Init(reg prometheus.Registerer) error {
	execCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.CronSubSys,
			Name:      "exec_count",
			Help:      "the total count to exec cron task",
		}, []string{TaskName})
	reg.MustRegister(execCounter)

	execError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.CronSubSys,
			Name:      "exec_error",
			Help:      "the total count to exec cron task error",
		}, []string{TaskName})
	reg.MustRegister(execError)

	execDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Subsystem: metrics.CronSubSys,
		Name:      "exec_duration_seconds",
		Help:      "the duration to exec cron task",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 10, 20, 30, 45, 90,
			120, 180, 300, 600, 1800, 3600, 7200, 10800, 21600, 43200, 86400},
	}, []string{TaskName})
	reg.MustRegister(execDuration)

	return nil
}

// ExecCounter return the execCounter
func ExecCounter() *prometheus.CounterVec {
	return execCounter
}

// ExecError return the execError
func ExecError() *prometheus.CounterVec {
	return execError
}

// ExecDuration return the execDuration
func ExecDuration() *prometheus.HistogramVec {
	return execDuration
}
