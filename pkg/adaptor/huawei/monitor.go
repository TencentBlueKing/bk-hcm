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

package huawei

import (
	"fmt"
	"strconv"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

const (
	// huaWeiMonitorMaxDimensionsPerMetric 指标维度，目前最大可添加4个维度
	huaWeiMonitorMaxDimensionsPerMetric = 4
	// huaWeiMonitorMaxMetricsPerBatch 指标数据。数组长度最大500
	huaWeiMonitorMaxMetricsPerBatch = 500
)

// GetMonitorData get monitor data from huawei.
// reference: https://support.huaweicloud.com/api-ces/ces_03_0034.html
func (h *HuaWei) GetMonitorData(kt *kit.Kit, opt *typecvm.HuaWeiMonitorDataOption) (
	*typecvm.HuaWeiMonitorDataResult, error) {

	if err := validateMonitorOption(opt); err != nil {
		logs.Errorf("validate monitor option failed, vendor: %s, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, err
	}

	queryMetrics := buildQueryMetrics(opt)
	if err := validateQueryMetrics(queryMetrics); err != nil {
		logs.Errorf("validate query metrics failed, vendor: %s, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, err
	}

	responseMetrics, err := h.fetchBatchMonitorData(kt, queryMetrics, opt)
	if err != nil {
		logs.Errorf("fetch batch monitor data failed, vendor: %s, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, err
	}

	return h.buildMonitorResult(kt, responseMetrics, opt.Filter), nil
}

// validateMonitorOption validates monitor option parameters
func validateMonitorOption(opt *typecvm.HuaWeiMonitorDataOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "monitor data option is required")
	}
	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return nil
}

// normalizeMonitorParams normalizes monitor query parameters with defaults
func normalizeMonitorParams(opt *typecvm.HuaWeiMonitorDataOption) (namespace, filter, dimensionName string) {
	namespace = opt.Namespace
	if len(namespace) == 0 {
		namespace = constant.HuaWeiCvmNamespace
	}

	filter = opt.Filter
	if len(filter) == 0 {
		filter = constant.HuaWeiMonitorDefaultFilter
	}

	dimensionName = opt.Dimension
	if len(dimensionName) == 0 {
		dimensionName = constant.HuaWeiCvmInstanceIDKey
		if namespace == constant.HuaWeiVpcNamespace {
			dimensionName = constant.HuaWeiPublicIPIDKey
		}
	}

	return namespace, filter, dimensionName
}

// buildQueryMetrics builds query metrics info list from options
func buildQueryMetrics(opt *typecvm.HuaWeiMonitorDataOption) []cesmodel.MetricInfo {
	namespace, _, dimensionName := normalizeMonitorParams(opt)

	queryMetrics := make([]cesmodel.MetricInfo, 0, len(opt.InstanceIDs))
	for _, instanceID := range opt.InstanceIDs {
		queryMetrics = append(queryMetrics, cesmodel.MetricInfo{
			Namespace:  namespace,
			MetricName: opt.MetricName,
			Dimensions: []cesmodel.MetricsDimension{{
				Name:  dimensionName,
				Value: instanceID,
			}},
		})
	}

	return queryMetrics
}

// validateQueryMetrics validates dimensions count for query metrics
func validateQueryMetrics(queryMetrics []cesmodel.MetricInfo) error {
	for _, metric := range queryMetrics {
		if len(metric.Dimensions) > huaWeiMonitorMaxDimensionsPerMetric {
			return errf.Newf(errf.InvalidParameter, "dimensions count should <= %d",
				huaWeiMonitorMaxDimensionsPerMetric)
		}
	}
	return nil
}

// fetchBatchMonitorData fetches monitor data in batches
func (h *HuaWei) fetchBatchMonitorData(kt *kit.Kit, queryMetrics []cesmodel.MetricInfo,
	opt *typecvm.HuaWeiMonitorDataOption) ([]cesmodel.BatchMetricData, error) {

	_, filter, _ := normalizeMonitorParams(opt)
	responseMetrics := make([]cesmodel.BatchMetricData, 0)

	for offset := 0; offset < len(queryMetrics); offset += huaWeiMonitorMaxMetricsPerBatch {
		end := offset + huaWeiMonitorMaxMetricsPerBatch
		if end > len(queryMetrics) {
			end = len(queryMetrics)
		}

		batchData, err := h.fetchSingleBatch(kt, queryMetrics[offset:end], opt, filter)
		if err != nil {
			return nil, err
		}
		responseMetrics = append(responseMetrics, batchData...)
	}

	return responseMetrics, nil
}

// fetchSingleBatch fetches a single batch of monitor data
func (h *HuaWei) fetchSingleBatch(kt *kit.Kit, metrics []cesmodel.MetricInfo, opt *typecvm.HuaWeiMonitorDataOption,
	filter string) ([]cesmodel.BatchMetricData, error) {

	client, err := h.clientSet.cesClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new huawei ces client failed, err: %v", err)
	}

	req := &cesmodel.BatchListMetricDataRequest{
		ContentType: "application/json; charset=UTF-8",
		Body: &cesmodel.BatchListMetricDataRequestBody{
			Metrics: metrics,
			Period:  strconv.FormatInt(opt.Period, 10),
			Filter:  filter,
			From:    opt.StartTime,
			To:      opt.EndTime,
		},
	}

	resp, err := client.BatchListMetricData(req)
	if err != nil {
		logs.Errorf("get huawei monitor data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Metrics != nil {
		return *resp.Metrics, nil
	}
	return nil, nil
}

// buildMonitorResult builds monitor result from response metrics
func (h *HuaWei) buildMonitorResult(kt *kit.Kit, responseMetrics []cesmodel.BatchMetricData,
	filter string) *typecvm.HuaWeiMonitorDataResult {

	result := &typecvm.HuaWeiMonitorDataResult{
		DataPoints: make([]*typecvm.MonitorDataPoint, 0, len(responseMetrics)),
	}

	for _, metric := range responseMetrics {
		dataPoint := h.convertToMonitorDataPoint(kt, &metric, filter)
		result.DataPoints = append(result.DataPoints, dataPoint)
	}

	return result
}

// convertToMonitorDataPoint converts batch metric data to monitor data point
func (h *HuaWei) convertToMonitorDataPoint(kt *kit.Kit, metric *cesmodel.BatchMetricData, filter string) *typecvm.MonitorDataPoint {
	dimensionsCount := 0
	if metric.Dimensions != nil {
		dimensionsCount = len(*metric.Dimensions)
	}

	dataPoint := &typecvm.MonitorDataPoint{
		Dimensions: make([]*typecvm.MonitorDimension, 0, dimensionsCount),
		Timestamps: make([]int64, 0, len(metric.Datapoints)),
		Values:     make([]float64, 0, len(metric.Datapoints)),
		Extensions: map[string]interface{}{
			"namespace":   metric.Namespace,
			"metric_name": metric.MetricName,
			"unit":        metric.Unit,
			"filter":      filter,
		},
	}

	h.fillDimensions(dataPoint, metric.Dimensions)
	h.fillDatapoints(kt, dataPoint, metric)

	return dataPoint
}

// fillDimensions fills dimensions to data point
func (h *HuaWei) fillDimensions(dataPoint *typecvm.MonitorDataPoint, dimensions *[]cesmodel.MetricsDimension) {
	if dimensions == nil {
		return
	}

	for _, dim := range *dimensions {
		dataPoint.Dimensions = append(dataPoint.Dimensions, &typecvm.MonitorDimension{
			Name:  dim.Name,
			Value: dim.Value,
		})
	}

	if len(dataPoint.Dimensions) > 0 {
		dataPoint.Extensions["dimensions"] = dataPoint.Dimensions
	}
}

// fillDatapoints fills datapoints to data point
func (h *HuaWei) fillDatapoints(kt *kit.Kit, dataPoint *typecvm.MonitorDataPoint, metric *cesmodel.BatchMetricData) {
	for _, dp := range metric.Datapoints {
		value, ok := getHuaWeiMetricValue(dp)
		if !ok {
			logs.Warnf("huawei monitor datapoint has no valid value, metric_name: %s, rid: %s",
				metric.MetricName, kt.Rid)
			continue
		}

		dataPoint.Timestamps = append(dataPoint.Timestamps, dp.Timestamp)
		dataPoint.Values = append(dataPoint.Values, value)
	}
}

func getHuaWeiMetricValue(dp cesmodel.DatapointForBatchMetric) (float64, bool) {
	if dp.Average != nil {
		return *dp.Average, true
	}
	if dp.Max != nil {
		return *dp.Max, true
	}
	if dp.Min != nil {
		return *dp.Min, true
	}
	if dp.Sum != nil {
		return *dp.Sum, true
	}
	if dp.Variance != nil {
		return *dp.Variance, true
	}

	return 0, false
}
