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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Plugin process pool metrics: hcm_plugin_pool_*.
//
// 事件类（直方图、计数器）由进程池在事件发生处调用 Observe / Inc 记录；
// 状态类（gauge）由每个池注册一个 collector，在抓取时调用池提供的 stats 函数读取。
// 标签只有插件名、进程状态与结果，不含任何 Provider 配置信息。

// PluginPoolSubSys defines plugin process pool related sub system.
const PluginPoolSubSys = "plugin_pool"

const (
	labelPlugin = "plugin"
	labelResult = "result"
)

type pluginPoolMetric struct {
	acquireCost   *prometheus.HistogramVec
	leaseHold     *prometheus.HistogramVec
	startCost     *prometheus.HistogramVec
	configureCost *prometheus.HistogramVec
	retryTotal    *prometheus.CounterVec
}

var (
	pluginPoolOnce sync.Once
	pluginPool     *pluginPoolMetric
)

func initPluginPoolMetric() {
	pluginPoolOnce.Do(func() {
		m := &pluginPoolMetric{}
		m.acquireCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: PluginPoolSubSys,
			Name:      "acquire_seconds",
			Help:      "the cost seconds from acquire call to return, by result.",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60},
		}, []string{labelPlugin, labelResult})
		Register().MustRegister(m.acquireCost)

		m.leaseHold = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: PluginPoolSubSys,
			Name:      "lease_hold_seconds",
			Help:      "the seconds a lease is held from delivery to release.",
			Buckets:   []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300, 600, 1800},
		}, []string{labelPlugin})
		Register().MustRegister(m.leaseHold)

		m.startCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: PluginPoolSubSys,
			Name:      "start_seconds",
			Help:      "the cost seconds of starting a plugin process, by result.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		}, []string{labelPlugin, labelResult})
		Register().MustRegister(m.startCost)

		m.configureCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: PluginPoolSubSys,
			Name:      "configure_seconds",
			Help:      "the cost seconds of configuring a plugin process, by result.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		}, []string{labelPlugin, labelResult})
		Register().MustRegister(m.configureCost)

		m.retryTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: PluginPoolSubSys,
			Name:      "retries_total",
			Help:      "the total retries counted on requests after a plugin process failure.",
		}, []string{labelPlugin})
		Register().MustRegister(m.retryTotal)

		pluginPool = m
	})
}

// ObservePluginPoolAcquire records one Acquire call.
func ObservePluginPoolAcquire(plugin, result string, cost time.Duration) {
	initPluginPoolMetric()
	pluginPool.acquireCost.WithLabelValues(plugin, result).Observe(cost.Seconds())
}

// ObservePluginPoolLeaseHold records how long one lease was held.
func ObservePluginPoolLeaseHold(plugin string, hold time.Duration) {
	initPluginPoolMetric()
	pluginPool.leaseHold.WithLabelValues(plugin).Observe(hold.Seconds())
}

// ObservePluginPoolStart records one plugin process start.
func ObservePluginPoolStart(plugin, result string, cost time.Duration) {
	initPluginPoolMetric()
	pluginPool.startCost.WithLabelValues(plugin, result).Observe(cost.Seconds())
}

// ObservePluginPoolConfigure records one ConfigureProvider call.
func ObservePluginPoolConfigure(plugin, result string, cost time.Duration) {
	initPluginPoolMetric()
	pluginPool.configureCost.WithLabelValues(plugin, result).Observe(cost.Seconds())
}

// IncPluginPoolRetry records one increase of a request's retry count.
func IncPluginPoolRetry(plugin string) {
	initPluginPoolMetric()
	pluginPool.retryTotal.WithLabelValues(plugin).Inc()
}

// PluginPoolStats 一个进程池在某一时刻的状态。
type PluginPoolStats struct {
	QueueLength int
	// OldestWait 队首请求已等待的时长，队列为空时为 0。
	OldestWait time.Duration
	// Processes 各状态的进程数，key 为状态名。
	Processes           map[string]int
	InflightLeases      int
	MaxProcesses        int
	MaxLeasesPerProcess int
}

// RegisterPluginPoolStats 为一个进程池注册状态类指标，抓取时调用 stats 读取。
//
// 插件名作为常量标签，同一个插件同时只能注册一次，重复注册返回错误。
func RegisterPluginPoolStats(plugin string, stats func() PluginPoolStats) error {
	c := newPluginPoolCollector(plugin, stats)
	return Register().Register(c)
}

// pluginPoolCollector 一个进程池的状态类指标。
type pluginPoolCollector struct {
	stats func() PluginPoolStats

	queueLength    *prometheus.Desc
	oldestWait     *prometheus.Desc
	processes      *prometheus.Desc
	inflightLeases *prometheus.Desc
	maxProcesses   *prometheus.Desc
	maxLeases      *prometheus.Desc
}

func newPluginPoolCollector(plugin string, stats func() PluginPoolStats) *pluginPoolCollector {
	l := prometheus.Labels{labelPlugin: plugin}
	return &pluginPoolCollector{
		stats:          stats,
		queueLength:    newPluginPoolDesc("queue_length", "the number of requests waiting in the queue.", l),
		oldestWait:     newPluginPoolDesc("queue_oldest_wait_seconds", "the seconds the queue head has waited.", l),
		processes:      newPluginPoolDesc("processes", "the number of plugin processes by state.", l, LabelState),
		inflightLeases: newPluginPoolDesc("inflight_leases", "the number of leases not yet released.", l),
		maxProcesses:   newPluginPoolDesc("max_processes", "the configured max processes of the pool.", l),
		maxLeases:      newPluginPoolDesc("max_leases_per_process", "the configured max leases per process.", l),
	}
}

func newPluginPoolDesc(name, help string, constLabels prometheus.Labels, labels ...string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(Namespace, PluginPoolSubSys, name), help, labels, constLabels)
}

// Describe implements prometheus.Collector.
func (c *pluginPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.queueLength
	ch <- c.oldestWait
	ch <- c.processes
	ch <- c.inflightLeases
	ch <- c.maxProcesses
	ch <- c.maxLeases
}

// Collect implements prometheus.Collector.
func (c *pluginPoolCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.stats()

	ch <- prometheus.MustNewConstMetric(c.queueLength, prometheus.GaugeValue, float64(s.QueueLength))
	ch <- prometheus.MustNewConstMetric(c.oldestWait, prometheus.GaugeValue, s.OldestWait.Seconds())
	for state, n := range s.Processes {
		ch <- prometheus.MustNewConstMetric(c.processes, prometheus.GaugeValue, float64(n), state)
	}
	ch <- prometheus.MustNewConstMetric(c.inflightLeases, prometheus.GaugeValue, float64(s.InflightLeases))
	ch <- prometheus.MustNewConstMetric(c.maxProcesses, prometheus.GaugeValue, float64(s.MaxProcesses))
	ch <- prometheus.MustNewConstMetric(c.maxLeases, prometheus.GaugeValue, float64(s.MaxLeasesPerProcess))
}
