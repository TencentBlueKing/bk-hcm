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

package metrics

import (
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/version"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// globalRegister is a global register which is used to collect metrics we need.
// it will be initialized when process is up for safe usage.
// and then be revised later when service is initialized.
var globalRegister prometheus.Registerer

func init() {
	// set default global register
	globalRegister = prometheus.DefaultRegisterer
}

// Register must only be called after the metric service is started.
func Register() prometheus.Registerer {
	return globalRegister
}

// httpHandler used to expose the metrics to prometheus.
var httpHandler http.Handler

// Handler returns the http handler with metrics.
func Handler() http.Handler {
	return httpHandler
}

const (
	// Namespace is the root namespace of the hcm metric
	Namespace = "hcm"

	// RestfulSubSys defines rest server's sub system
	RestfulSubSys = "restful"

	// OrmCmdSubSys defines all the orm command related sub system.
	OrmCmdSubSys = "orm"

	// CloudApiSubSys defines all cloud api related subsystem
	CloudApiSubSys = "cloudapi"
)

// labels
const (
	LabelProcessName = "process_name"
	LabelHost        = "host"
)

// InitMetrics init metrics registerer and http handler
func InitMetrics(endpoint string) {
	registry := prometheus.NewRegistry()

	processName := string(cc.ServiceName())
	label := prometheus.Labels{LabelProcessName: processName, LabelHost: endpoint}

	register := prometheus.WrapRegistererWith(label, registry)

	// set up global register
	globalRegister = register

	register.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	register.MustRegister(collectors.NewGoCollector())

	// metric current service version.
	versionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   Namespace,
			Subsystem:   "version",
			Name:        "info",
			Help:        "The version info of the current service",
			ConstLabels: prometheus.Labels{},
		},
		[]string{"version", "build_time", "git_hash"},
	)
	register.MustRegister(versionGauge)
	versionGauge.With(prometheus.Labels{
		"version":    version.VERSION,
		"build_time": version.BUILDTIME,
		"git_hash":   version.GITHASH,
	}).Set(1)

	// set up metrics http handler
	httpHandler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
