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
	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// TODO: 任务流、任务成功或失败等情况metrics打点
func initMetric(register prometheus.Registerer) *metric {
	m := new(metric)

	labels := prometheus.Labels{}
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
			Buckets:     []float64{1, 2, 3, 4, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 400, 800, 1000, 1500, 6000},
		},
		[]string{"flowType"},
	)
	register.MustRegister(m.flowTypeExecTime)
	return m
}

type metric struct {
	taskInitQueueSize  *prometheus.GaugeVec
	flowTypeRunningNum *prometheus.GaugeVec
	flowTypeExecTime   *prometheus.HistogramVec
}
