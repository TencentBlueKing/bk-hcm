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
	typescw "hcm/pkg/adaptor/types/cloudwatch"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// GetMetricData queries CloudWatch metric time-series data.
// reference: https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
func (a *Aws) GetMetricData(kt *kit.Kit, opt *typescw.AwsGetMetricDataOption) ([]*typescw.MetricDataResult, error) {
	client, err := a.clientSet.cloudWatchClient(opt.Region)
	if err != nil {
		return nil, err
	}

	queries := make([]*cloudwatch.MetricDataQuery, 0, len(opt.MetricDataQueries))
	for _, q := range opt.MetricDataQueries {
		dims := make([]*cloudwatch.Dimension, 0, len(q.Dimensions))
		for _, d := range q.Dimensions {
			dims = append(dims, &cloudwatch.Dimension{
				Name:  aws.String(d.Name),
				Value: aws.String(d.Value),
			})
		}
		queries = append(queries, &cloudwatch.MetricDataQuery{
			Id: aws.String(q.ID),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Namespace:  aws.String(q.Namespace),
					MetricName: aws.String(q.MetricName),
					Dimensions: dims,
				},
				Stat:   aws.String(q.Stat),
				Period: aws.Int64(q.Period),
			},
		})
	}

	startTime := opt.StartTime
	endTime := opt.EndTime
	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: queries,
		StartTime:         &startTime,
		EndTime:           &endTime,
	}

	// Use a map to merge results across pages — AWS may split the same Id across pages.
	mergedMap := make(map[string]*typescw.MetricDataResult)
	var orderedIDs []string

	for {
		resp, err := client.GetMetricDataWithContext(kt.Ctx, input)
		if err != nil {
			logs.Errorf("get aws cloudwatch metric data failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, result := range resp.MetricDataResults {
			id := aws.StringValue(result.Id)
			item, exists := mergedMap[id]
			if !exists {
				item = &typescw.MetricDataResult{ID: id}
				mergedMap[id] = item
				orderedIDs = append(orderedIDs, id)
			}
			if result.Label != nil {
				item.Label = aws.StringValue(result.Label)
			}
			if result.StatusCode != nil {
				item.StatusCode = aws.StringValue(result.StatusCode)
			}
			for _, msg := range result.Messages {
				item.Messages = append(item.Messages, typescw.MetricDataMessage{
					Code:  aws.StringValue(msg.Code),
					Value: aws.StringValue(msg.Value),
				})
			}
			for _, t := range result.Timestamps {
				item.Timestamps = append(item.Timestamps, t.Unix())
			}
			for _, v := range result.Values {
				item.Values = append(item.Values, aws.Float64Value(v))
			}
		}

		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}

	data := make([]*typescw.MetricDataResult, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		data = append(data, mergedMap[id])
	}

	return data, nil
}

// ListMetrics lists available CloudWatch metrics.
// Returns raw AWS SDK cloudwatch.Metric objects for transparent pass-through.
// reference: https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_ListMetrics.html
func (a *Aws) ListMetrics(kt *kit.Kit, opt *typescw.AwsListMetricsOption) ([]*cloudwatch.Metric, error) {
	client, err := a.clientSet.cloudWatchClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &cloudwatch.ListMetricsInput{}
	if opt.Namespace != "" {
		input.Namespace = aws.String(opt.Namespace)
	}
	if opt.MetricName != "" {
		input.MetricName = aws.String(opt.MetricName)
	}
	if len(opt.Dimensions) > 0 {
		filters := make([]*cloudwatch.DimensionFilter, 0, len(opt.Dimensions))
		for _, d := range opt.Dimensions {
			filters = append(filters, &cloudwatch.DimensionFilter{
				Name:  aws.String(d.Name),
				Value: aws.String(d.Value),
			})
		}
		input.Dimensions = filters
	}

	data := make([]*cloudwatch.Metric, 0)
	for {
		resp, err := client.ListMetricsWithContext(kt.Ctx, input)
		if err != nil {
			logs.Errorf("list aws cloudwatch metrics failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		data = append(data, resp.Metrics...)

		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}

	return data, nil
}
