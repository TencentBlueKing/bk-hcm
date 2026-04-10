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

package instancetype

import (
	typescw "hcm/pkg/adaptor/types/cloudwatch"
	proto "hcm/pkg/api/hc-service/instance-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListAssumeRoleMetricDataForAws queries CloudWatch metric time-series data via AssumeRole.
func (i *instanceTypeAdaptor) ListAssumeRoleMetricDataForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleGetMetricDataReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// Get and validate CloudID from main_account table
	cloudID, err := i.getCloudIDFromMainAccount(cts.Kit, req.MainAccountID, req.RootAccountID)
	if err != nil {
		return nil, err
	}

	client, err := i.adaptor.AwsWithAssumeRole(cts.Kit, req.RootAccountID, cloudID, req.RoleChain, req.ExternalID)
	if err != nil {
		return nil, err
	}

	queries := make([]typescw.MetricDataQuery, 0, len(req.MetricDataQueries))
	for _, q := range req.MetricDataQueries {
		dims := make([]typescw.Dimension, 0, len(q.Dimensions))
		for _, d := range q.Dimensions {
			dims = append(dims, typescw.Dimension{Name: d.Name, Value: d.Value})
		}
		queries = append(queries, typescw.MetricDataQuery{
			ID:         q.ID,
			Namespace:  q.Namespace,
			MetricName: q.MetricName,
			Dimensions: dims,
			Stat:       q.Stat,
			Period:     q.Period,
		})
	}

	opt := &typescw.AwsGetMetricDataOption{
		Region:            req.Region,
		MetricDataQueries: queries,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
	}

	results, err := client.GetMetricData(cts.Kit, opt)
	if err != nil {
		logs.Errorf("get aws assume role cloudwatch metric data failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	data := make([]*proto.MetricDataResultItem, 0, len(results))
	for _, r := range results {
		msgs := make([]proto.MetricDataMessageItem, 0, len(r.Messages))
		for _, m := range r.Messages {
			msgs = append(msgs, proto.MetricDataMessageItem{Code: m.Code, Value: m.Value})
		}
		data = append(data, &proto.MetricDataResultItem{
			ID:         r.ID,
			Label:      r.Label,
			StatusCode: r.StatusCode,
			Messages:   msgs,
			Timestamps: r.Timestamps,
			Values:     r.Values,
		})
	}

	return data, nil
}

// ListAssumeRoleMetricsForAws lists available CloudWatch metrics via AssumeRole.
func (i *instanceTypeAdaptor) ListAssumeRoleMetricsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleListMetricsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// Get and validate CloudID from main_account table
	cloudID, err := i.getCloudIDFromMainAccount(cts.Kit, req.MainAccountID, req.RootAccountID)
	if err != nil {
		return nil, err
	}

	client, err := i.adaptor.AwsWithAssumeRole(cts.Kit, req.RootAccountID, cloudID, req.RoleChain, req.ExternalID)
	if err != nil {
		return nil, err
	}

	dims := make([]typescw.Dimension, 0, len(req.Dimensions))
	for _, d := range req.Dimensions {
		dims = append(dims, typescw.Dimension{Name: d.Name, Value: d.Value})
	}

	opt := &typescw.AwsListMetricsOption{
		Region:     req.Region,
		Namespace:  req.Namespace,
		MetricName: req.MetricName,
		Dimensions: dims,
	}

	results, err := client.ListMetrics(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list aws assume role cloudwatch metrics failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results, nil
}
