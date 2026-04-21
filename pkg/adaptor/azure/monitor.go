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

package azure

import (
	"fmt"
	"strings"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

// GetMonitorData gets monitor data from azure monitor.
// reference: https://learn.microsoft.com/rest/api/monitor/metrics/list
func (az *Azure) GetMonitorData(kt *kit.Kit, opt *typecvm.AzureMonitorDataOption) (
	*typecvm.AzureMonitorDataResult, error) {

	if err := validateAzureMonitorOption(opt); err != nil {
		logs.Errorf("validate azure monitor option failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	client, err := az.clientSet.monitorMetricsClient()
	if err != nil {
		logs.Errorf("new azure monitor metrics client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	sourceMetricName, isFallback := mapAzureMetricName(opt.MetricName)
	result := &typecvm.AzureMonitorDataResult{
		DataPoints: make([]*typecvm.MonitorDataPoint, 0, len(opt.InstanceIDs)),
	}

	options := buildAzureMetricsListOptions(opt, sourceMetricName)
	for _, instanceID := range opt.InstanceIDs {
		resp, err := client.List(kt.Ctx, instanceID, options)
		if err != nil {
			logs.Errorf("get azure monitor data failed, instance_id: %s, metric_name: %s, err: %v, rid: %s",
				instanceID, sourceMetricName, err, kt.Rid)
			return nil, err
		}

		dataPoint := buildAzureMonitorDataPoint(instanceID, sourceMetricName, opt, resp, isFallback)
		result.DataPoints = append(result.DataPoints, dataPoint)
	}

	return result, nil
}

func validateAzureMonitorOption(opt *typecvm.AzureMonitorDataOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "monitor data option is required")
	}
	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return nil
}

func buildAzureMetricsListOptions(opt *typecvm.AzureMonitorDataOption,
	sourceMetricName string) *armmonitor.MetricsClientListOptions {

	// 构建时间范围，格式为 "startTime/endTime"，符合 Azure Monitor Metrics API 的 timespan 参数要求
	timespan := fmt.Sprintf("%s/%s", opt.StartTime, opt.EndTime)
	// 构建采样间隔，格式为 ISO 8601 持续时间格式，如 PT60S 表示 60 秒
	interval := fmt.Sprintf("PT%dS", opt.Period)
	// 根据指标类型选择聚合方式，流量类指标默认 Total，其余默认 Average
	aggregation := chooseAzureAggregation(opt)
	namespace := opt.MetricNamespace
	if len(namespace) == 0 {
		namespace = constant.AzureMetricNamespaceDefault
	}
	resultType := parseAzureResultType(opt.ResultType)

	req := &armmonitor.MetricsClientListOptions{
		Timespan:            converter.ValToPtr(timespan),
		Interval:            converter.ValToPtr(interval),
		Metricnames:         converter.ValToPtr(sourceMetricName),
		Aggregation:         converter.ValToPtr(aggregation),
		Metricnamespace:     converter.ValToPtr(namespace),
		AutoAdjustTimegrain: opt.AutoAdjustTimegrain,
		Top:                 opt.Top,
		Orderby:             converter.ValToPtr(opt.OrderBy),
		Filter:              converter.ValToPtr(opt.Filter),
		ResultType:          resultType,
	}

	if len(opt.OrderBy) == 0 {
		req.Orderby = nil
	}
	if len(opt.Filter) == 0 {
		req.Filter = nil
	}

	return req
}

func parseAzureResultType(resultType typecvm.AzureResultType) *armmonitor.ResultType {
	if len(resultType) == 0 {
		return nil
	}
	switch strings.ToLower(string(resultType)) {
	case "metadata":
		rt := armmonitor.ResultTypeMetadata
		return &rt
	default:
		rt := armmonitor.ResultTypeData
		return &rt
	}
}

func chooseAzureAggregation(opt *typecvm.AzureMonitorDataOption) string {
	if len(opt.Aggregation) != 0 {
		return opt.Aggregation
	}

	metricName := strings.ToLower(opt.MetricName)
	if metricName == strings.ToLower(constant.MetricLanIntraffic) ||
		metricName == strings.ToLower(constant.MetricLanOuttraffic) ||
		metricName == strings.ToLower(constant.MetricWanIntraffic) ||
		metricName == strings.ToLower(constant.MetricWanOuttraffic) {
		return constant.AzureMonitorAggregationTotal
	}
	return constant.AzureMonitorAggregationAverage
}

func mapAzureMetricName(metricName string) (string, bool) {
	switch metricName {
	case constant.MetricLanIntraffic, constant.MetricWanIntraffic:
		return constant.AzureMetricNetworkInTotal, true
	case constant.MetricLanOuttraffic, constant.MetricWanOuttraffic:
		return constant.AzureMetricNetworkOutTotal, true
	default:
		return metricName, false
	}
}

func buildAzureMonitorDataPoint(instanceID, sourceMetricName string, opt *typecvm.AzureMonitorDataOption,
	resp armmonitor.MetricsClientListResponse, isFallback bool) *typecvm.MonitorDataPoint {

	resourceID := extractAzureResourceID(resp, instanceID)
	dataPoint := &typecvm.MonitorDataPoint{
		Dimensions: []*typecvm.MonitorDimension{{
			Name:  constant.AzureCvmInstanceIDKey,
			Value: resourceID,
		}},
		Timestamps: make([]int64, 0),
		Values:     make([]float64, 0),
		Extensions: buildAzureMonitorExtensions(resp, opt, sourceMetricName, isFallback),
	}

	if resp.Value == nil {
		return dataPoint
	}
	aggregation := chooseAzureAggregation(opt)
	for _, metric := range resp.Value {
		if metric == nil || metric.Timeseries == nil {
			continue
		}
		for _, ts := range metric.Timeseries {
			if ts == nil || ts.Data == nil {
				continue
			}
			for _, dp := range ts.Data {
				if dp == nil || dp.TimeStamp == nil {
					continue
				}
				metricValue, ok := pickAzureMetricValue(dp, aggregation)
				if !ok {
					continue
				}
				dataPoint.Timestamps = append(dataPoint.Timestamps, dp.TimeStamp.Unix())
				dataPoint.Values = append(dataPoint.Values, metricValue)
			}
		}
	}

	return dataPoint
}

func extractAzureResourceID(resp armmonitor.MetricsClientListResponse, fallback string) string {
	for _, metric := range resp.Value {
		if metric == nil || metric.Timeseries == nil {
			continue
		}
		for _, ts := range metric.Timeseries {
			if ts == nil || ts.Metadatavalues == nil {
				continue
			}
			for _, meta := range ts.Metadatavalues {
				if meta == nil || meta.Name == nil || meta.Name.Value == nil || meta.Value == nil {
					continue
				}
				if strings.EqualFold(*meta.Name.Value, constant.AzureCvmInstanceIDKey) {
					return *meta.Value
				}
			}
		}
	}
	return fallback
}

func buildAzureMonitorExtensions(resp armmonitor.MetricsClientListResponse, opt *typecvm.AzureMonitorDataOption,
	sourceMetricName string, isFallback bool) map[string]interface{} {

	extensions := map[string]interface{}{
		"vendor":             enumor.Azure,
		"metric_name":        opt.MetricName,
		"source_metric_name": sourceMetricName,
		"aggregation":        chooseAzureAggregation(opt),
		"semantic_phase":     "native_monitor_query",
		"traffic_scope":      "native",
		"is_fallback":        false,
	}

	extensions["cost"] = converter.PtrToVal(resp.Cost)
	extensions["granularity"] = converter.PtrToVal(resp.Interval)
	extensions["namespace"] = converter.PtrToVal(resp.Namespace)
	extensions["resource_region"] = converter.PtrToVal(resp.Resourceregion)
	if len(opt.MetricNamespace) != 0 {
		extensions["namespace"] = opt.MetricNamespace
	}

	if isFallback {
		extensions["semantic_phase"] = "phase1_total_traffic_mapping"
		extensions["traffic_scope"] = "total"
		extensions["is_fallback"] = true
	}

	extensions["unit"] = extractAzureMetricUnit(resp)
	return extensions
}

func extractAzureMetricUnit(resp armmonitor.MetricsClientListResponse) string {
	if resp.Value == nil {
		return ""
	}
	for _, metric := range resp.Value {
		if metric != nil && metric.Unit != nil {
			return string(*metric.Unit)
		}
	}
	return ""
}

func pickAzureMetricValue(dp *armmonitor.MetricValue, aggregation string) (float64, bool) {
	switch strings.ToLower(aggregation) {
	case "average":
		if dp.Average != nil {
			return *dp.Average, true
		}
	case "minimum":
		if dp.Minimum != nil {
			return *dp.Minimum, true
		}
	case "maximum":
		if dp.Maximum != nil {
			return *dp.Maximum, true
		}
	case "count":
		if dp.Count != nil {
			return *dp.Count, true
		}
	case "total":
		if dp.Total != nil {
			return *dp.Total, true
		}
	default:
		if dp.Total != nil {
			return *dp.Total, true
		}
	}

	// fallback priority
	if dp.Total != nil {
		return *dp.Total, true
	}
	if dp.Average != nil {
		return *dp.Average, true
	}
	if dp.Maximum != nil {
		return *dp.Maximum, true
	}
	if dp.Minimum != nil {
		return *dp.Minimum, true
	}
	if dp.Count != nil {
		return *dp.Count, true
	}

	return 0, false
}
