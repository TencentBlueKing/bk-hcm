/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2026 THL A29 Limited,
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

package aws

import (
	"testing"

	typecloudwatch "hcm/pkg/adaptor/types/cloudwatch"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/require"
)

func TestMapMetricName_Phase1TrafficMapping(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		wantSource  string
		wantTraffic bool
	}{
		{name: "lan out to network out", metricName: constant.MetricLanOuttraffic, wantSource: constant.AwsMetricNetworkOut, wantTraffic: true},
		{name: "wan out to network out", metricName: constant.MetricWanOuttraffic, wantSource: constant.AwsMetricNetworkOut, wantTraffic: true},
		{name: "lan in to network in", metricName: constant.MetricLanIntraffic, wantSource: constant.AwsMetricNetworkIn, wantTraffic: true},
		{name: "wan in to network in", metricName: constant.MetricWanIntraffic, wantSource: constant.AwsMetricNetworkIn, wantTraffic: true},
		{name: "other metric unchanged", metricName: "CPUUtilization", wantSource: "CPUUtilization", wantTraffic: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, isTraffic := mapMetricName(tt.metricName)
			require.Equal(t, tt.wantSource, source)
			require.Equal(t, tt.wantTraffic, isTraffic)
		})
	}
}

func TestBuildMonitorDataResult_AddsPhase1Extensions(t *testing.T) {
	opt := &typecvm.AwsMonitorDataOption{
		MetricName:  constant.MetricLanOuttraffic,
		InstanceIDs: []string{"i-1"},
		StartTime:   "2026-04-09T00:00:00Z",
		EndTime:     "2026-04-09T01:00:00Z",
		Period:      300,
		Region:      "ap-east-1",
	}

	resp := []*typecloudwatch.MetricDataResult{
		{
			ID:         "m0",
			Timestamps: []int64{1712620800},
			Values:     []float64{1024},
		},
	}

	got := buildMonitorDataResult(
		&kit.Kit{Rid: "test-rid"},
		opt,
		resp,
		map[string]string{"m0": "i-1"},
		constant.AwsMetricNetworkOut,
		true,
	)

	require.Len(t, got.DataPoints, 1)
	dp := got.DataPoints[0]
	require.Len(t, dp.Dimensions, 1)
	require.Equal(t, constant.AwsCvmInstanceIDKey, dp.Dimensions[0].Name)
	require.Equal(t, "i-1", dp.Dimensions[0].Value)
	require.Equal(t, constant.AwsMetricNetworkOut, dp.Extensions["source_metric_name"])
	require.Equal(t, "phase1_total_traffic_mapping", dp.Extensions["semantic_phase"])
	require.Equal(t, "total", dp.Extensions["traffic_scope"])
	require.Equal(t, "Bytes", dp.Extensions["unit"])
}
