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

package azure

import (
	"testing"
	"time"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/stretchr/testify/require"
)

func TestMapAzureMetricName_Phase1TrafficMapping(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		wantMetric string
		fallback   bool
	}{
		{name: "lan in", metricName: constant.MetricLanIntraffic, wantMetric: constant.AzureMetricNetworkInTotal, fallback: true},
		{name: "wan in", metricName: constant.MetricWanIntraffic, wantMetric: constant.AzureMetricNetworkInTotal, fallback: true},
		{name: "lan out", metricName: constant.MetricLanOuttraffic, wantMetric: constant.AzureMetricNetworkOutTotal, fallback: true},
		{name: "wan out", metricName: constant.MetricWanOuttraffic, wantMetric: constant.AzureMetricNetworkOutTotal, fallback: true},
		{name: "normal metric", metricName: "Percentage CPU", wantMetric: "Percentage CPU", fallback: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetric, gotFallback := mapAzureMetricName(tt.metricName)
			require.Equal(t, tt.wantMetric, gotMetric)
			require.Equal(t, tt.fallback, gotFallback)
		})
	}
}

func TestChooseAzureAggregation_Default(t *testing.T) {
	req := &typecvm.AzureMonitorDataOption{MetricName: constant.MetricLanOuttraffic}
	require.Equal(t, constant.AzureMonitorAggregationTotal, chooseAzureAggregation(req))

	req = &typecvm.AzureMonitorDataOption{MetricName: "Percentage CPU"}
	require.Equal(t, constant.AzureMonitorAggregationAverage, chooseAzureAggregation(req))
}

func TestBuildAzureMonitorDataPoint_FillValuesAndExtensions(t *testing.T) {
	ts := time.Date(2026, 4, 9, 0, 5, 0, 0, time.UTC)
	resp := armmonitor.MetricsClientListResponse{
		Response: armmonitor.Response{
			Cost:           ptrInt32(123),
			Interval:       ptrString("PT5M"),
			Namespace:      ptrString(constant.AzureMetricNamespaceDefault),
			Resourceregion: ptrString("eastus"),
			Value: []*armmonitor.Metric{
				{
					Unit: ptrUnit(armmonitor.UnitBytes),
					Timeseries: []*armmonitor.TimeSeriesElement{
						{
							Metadatavalues: []*armmonitor.MetadataValue{
								{
									Name:  &armmonitor.LocalizableString{Value: ptrString(constant.AzureCvmInstanceIDKey)},
									Value: ptrString("/subscriptions/s1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"),
								},
							},
							Data: []*armmonitor.MetricValue{
								{TimeStamp: &ts, Total: ptrFloat64(2048)},
							},
						},
					},
				},
			},
		},
	}

	dp := buildAzureMonitorDataPoint(
		"/subscriptions/s1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
		constant.AzureMetricNetworkOutTotal,
		&typecvm.AzureMonitorDataOption{MetricName: constant.MetricLanOuttraffic},
		resp,
		true,
	)

	require.Len(t, dp.Dimensions, 1)
	require.Equal(t, constant.AzureCvmInstanceIDKey, dp.Dimensions[0].Name)
	require.Equal(t, "/subscriptions/s1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
		dp.Dimensions[0].Value)
	require.Equal(t, []int64{ts.Unix()}, dp.Timestamps)
	require.Equal(t, []float64{2048}, dp.Values)
	require.Equal(t, "Bytes", dp.Extensions["unit"])
	require.Equal(t, int32(123), dp.Extensions["cost"])
	require.Equal(t, "PT5M", dp.Extensions["granularity"])
	require.Equal(t, constant.AzureMetricNamespaceDefault, dp.Extensions["namespace"])
	require.Equal(t, "eastus", dp.Extensions["resource_region"])
	require.Equal(t, true, dp.Extensions["is_fallback"])
	require.Equal(t, "phase1_total_traffic_mapping", dp.Extensions["semantic_phase"])
	require.Equal(t, "total", dp.Extensions["traffic_scope"])
}

func TestExtractAzureResourceID_Fallback(t *testing.T) {
	resp := armmonitor.MetricsClientListResponse{
		Response: armmonitor.Response{
			Value: []*armmonitor.Metric{
				{
					Timeseries: []*armmonitor.TimeSeriesElement{{}},
				},
			},
		},
	}

	got := extractAzureResourceID(resp, "fallback-id")
	require.Equal(t, "fallback-id", got)
}

func ptrString(v string) *string { return &v }
func ptrFloat64(v float64) *float64 { return &v }
func ptrInt32(v int32) *int32 { return &v }
func ptrUnit(v armmonitor.Unit) *armmonitor.Unit { return &v }
