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

package aws

import (
	"fmt"
	"time"

	typecloudwatch "hcm/pkg/adaptor/types/cloudwatch"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

const (
	awsCloudWatchStatSum     = "Sum"
	awsCloudWatchStatAverage = "Average"
)

// GetMonitorData gets monitor data from aws cloudwatch.
// reference: https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
func (a *Aws) GetMonitorData(kt *kit.Kit, opt *typecvm.AwsMonitorDataOption) (*typecvm.AwsMonitorDataResult, error) {
	if err := validateMonitorDataOption(opt); err != nil {
		logs.Errorf("validate aws monitor data option failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	startTime, endTime, err := parseMonitorTimeRange(opt)
	if err != nil {
		logs.Errorf("parse aws monitor time range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	sourceMetricName, isTrafficMetric := mapMetricName(opt.MetricName)
	queryIDToInstanceID, metricDataQueries := buildMonitorMetricDataQueries(opt, sourceMetricName)

	resp, err := a.GetMetricData(kt, &typecloudwatch.AwsGetMetricDataOption{
		Region:            opt.Region,
		MetricDataQueries: metricDataQueries,
		StartTime:         startTime,
		EndTime:           endTime,
	})
	if err != nil {
		logs.Errorf("get aws monitor data failed, region: %s, metric_name: %s, err: %v, rid: %s",
			opt.Region, opt.MetricName, err, kt.Rid)
		return nil, err
	}

	return buildMonitorDataResult(kt, opt, resp, queryIDToInstanceID, sourceMetricName, isTrafficMetric), nil
}

func validateMonitorDataOption(opt *typecvm.AwsMonitorDataOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "monitor data option is required")
	}
	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return nil
}

func parseMonitorTimeRange(opt *typecvm.AwsMonitorDataOption) (time.Time, time.Time, error) {
	startTime, err := time.Parse(time.RFC3339, opt.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, errf.NewFromErr(errf.InvalidParameter, err)
	}

	endTime, err := time.Parse(time.RFC3339, opt.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, errf.NewFromErr(errf.InvalidParameter, err)
	}
	return startTime, endTime, nil
}

func buildMonitorMetricDataQueries(opt *typecvm.AwsMonitorDataOption, sourceMetricName string) (
	map[string]string, []typecloudwatch.MetricDataQuery) {

	stat := chooseCloudWatchStat(sourceMetricName)
	queryIDToInstanceID := make(map[string]string, len(opt.InstanceIDs))
	metricDataQueries := make([]typecloudwatch.MetricDataQuery, 0, len(opt.InstanceIDs))
	for idx, instanceID := range opt.InstanceIDs {
		queryID := fmt.Sprintf("m%d", idx)
		queryIDToInstanceID[queryID] = instanceID
		metricDataQueries = append(metricDataQueries, typecloudwatch.MetricDataQuery{
			ID:         queryID,
			Namespace:  constant.AwsCvmNamespace,
			MetricName: sourceMetricName,
			Dimensions: []typecloudwatch.Dimension{{
				Name:  constant.AwsCvmInstanceIDKey,
				Value: instanceID,
			}},
			Stat:   stat,
			Period: opt.Period,
		})
	}

	return queryIDToInstanceID, metricDataQueries
}

func buildMonitorDataResult(kt *kit.Kit, opt *typecvm.AwsMonitorDataOption, resp []*typecloudwatch.MetricDataResult,
	queryIDToInstanceID map[string]string, sourceMetricName string,
	isTrafficMetric bool) *typecvm.AwsMonitorDataResult {

	result := &typecvm.AwsMonitorDataResult{DataPoints: make([]*typecvm.MonitorDataPoint, 0, len(resp))}
	for _, data := range resp {
		instanceID, ok := queryIDToInstanceID[data.ID]
		if !ok {
			logs.Warnf("aws monitor result query id not found, query_id: %s, rid: %s", data.ID, kt.Rid)
			continue
		}

		result.DataPoints = append(result.DataPoints, &typecvm.MonitorDataPoint{
			Dimensions: []*typecvm.MonitorDimension{{
				Name:  constant.AwsCvmInstanceIDKey,
				Value: instanceID,
			}},
			Timestamps: data.Timestamps,
			Values:     data.Values,
			Extensions: buildMonitorExtensions(opt.MetricName, sourceMetricName, isTrafficMetric),
		})
	}

	return result
}

func buildMonitorExtensions(metricName, sourceMetricName string, isTrafficMetric bool) map[string]interface{} {
	extensions := map[string]interface{}{
		"namespace":          constant.AwsCvmNamespace,
		"metric_name":        metricName,
		"source_metric_name": sourceMetricName,
		"stat":               chooseCloudWatchStat(sourceMetricName),
	}
	if isTrafficMetric {
		extensions["semantic_phase"] = "phase1_total_traffic_mapping"
		extensions["traffic_scope"] = "total"
		extensions["unit"] = "Bytes"
	}
	return extensions
}

func mapMetricName(metricName string) (string, bool) {
	switch metricName {
	case constant.MetricLanIntraffic, constant.MetricWanIntraffic:
		return constant.AwsMetricNetworkIn, true
	case constant.MetricLanOuttraffic, constant.MetricWanOuttraffic:
		return constant.AwsMetricNetworkOut, true
	default:
		return metricName, false
	}
}

func chooseCloudWatchStat(metricName string) string {
	switch metricName {
	case constant.AwsMetricNetworkIn, constant.AwsMetricNetworkOut:
		return awsCloudWatchStatSum
	default:
		return awsCloudWatchStatAverage
	}
}
