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

	// AsyncSubSys defines all the async flow or task related sub system.
	AsyncSubSys = "async"

	// CronSubSys defines all the cron related sub system.
	CronSubSys = "cron"

	// HTTPSubSys defines unified http request sub system. Used for
	// cross-service http_request_* metrics (api-server / hc-service /
	// data-service service-side requests, and adaptor cloud API calls).
	HTTPSubSys = "http"

	// CLBSubSys defines clb business related sub system. Used for
	// CLB submit entry metrics, e.g. hcm_clb_submit_*.
	CLBSubSys = "clb"
)

// labels
const (
	LabelProcessName = "process_name"
	LabelHost        = "host"
	// LabelBKCCBizID is the HCM business ID label. It avoids using
	// "bk_biz_id", which is reserved by BK Monitor for the collection biz.
	LabelBKCCBizID = "bkcc_biz_id"
	// LabelVendor vendor标签
	LabelVendor = "vendor"
	// LabelOperation operation标签
	LabelOperation = "operation"
	// LabelState state标签
	LabelState = "state"
	// LabelErrType errType标签
	LabelErrType = "err_type"
	// LabelFlowName flowName标签
	LabelFlowName = "flow_name"
	// LabelActionName actionName标签
	LabelActionName = "action_name"
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

	// 主动注册所有服务公共的 hcm_http_request_* 指标，确保 /metrics 在启动后
	// 立刻暴露 # HELP/# TYPE 元信息，避免依赖首次 Observe 的惰性注册导致的
	// metric 短暂缺失。该指标由 pkg/rest/handler.go middleware、api-server
	// proxy 以及 pkg/adaptor/metric 共同上报，所有服务进程都会暴露。
	EnsureHTTPRequestMetric()
}
