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
	"hcm/pkg/metrics"
	"hcm/pkg/tools/uuid"

	"github.com/prometheus/client_golang/prometheus"
)

func initMetric(register prometheus.Registerer) *metric {
	m := new(metric)
	labels := prometheus.Labels{
		"connection_id": uuid.UUID(),
	}

	m.cmdLagMS = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.OrmCmdSubSys,
		Name:        "cmd_lag_milliseconds",
		Help:        "the lags(milliseconds) to exec a ORM command",
		ConstLabels: labels,
		Buckets:     []float64{1, 2, 3, 4, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 400, 800, 1000, 1500, 2000},
	}, []string{"cmd"})
	register.MustRegister(m.cmdLagMS)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.OrmCmdSubSys,
			Name:        "total_err_count",
			Help:        "the total error count when exec a ORM command",
			ConstLabels: labels,
		}, []string{"cmd"})
	register.MustRegister(m.errCounter)

	return m
}

type metric struct {
	// cmdLagMS record the cost time to exec an orm command.
	cmdLagMS *prometheus.HistogramVec

	// errCounter record the total error count when exec an orm command.
	errCounter *prometheus.CounterVec
}
