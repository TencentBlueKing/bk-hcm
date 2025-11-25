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

package gcp

import (
	"errors"
	"fmt"
	"time"

	typesMonitoring "hcm/pkg/adaptor/types/monitoring"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListMonitorTimeSeries lists time series data from GCP Cloud Monitoring.
func (g *Gcp) ListMonitorTimeSeries(kt *kit.Kit, projectID string, opt *typesMonitoring.GcpListTimeSeriesOption) (
	*typesMonitoring.GcpListTimeSeriesResult, error) {

	if projectID == "" {
		return nil, errf.New(errf.InvalidParameter, "project id is required")
	}
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list time series option is required")
	}

	if err := opt.Validate(); err != nil {
		logs.Errorf("validate list time series option failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.monitoringClient(kt)
	if err != nil {
		logs.Errorf("create monitoring client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	defer client.Close()

	// Build and send request
	req, err := g.buildListTimeSeriesRequest(kt, projectID, opt)
	if err != nil {
		logs.Errorf("build list time series request failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// Call API - ListTimeSeries returns an iterator
	// The iterator will make HTTP request on first Next() call
	// and iter.Response will contain the full ListTimeSeriesResponse with all TimeSeries in that page
	iter := client.ListTimeSeries(kt.Ctx, req)

	// Trigger the API call by calling Next() once
	// This will populate iter.Response with the complete response for the current page
	// Note: Next() returns a single TimeSeries, but iter.Response contains the full page
	if _, err := iter.Next(); err != nil && !errors.Is(err, iterator.Done) {
		logs.Errorf("call list time series API failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// Extract the full response from iterator
	// iter.Response is populated after the first Next() call and contains the complete page
	return g.collectTimeSeriesResults(kt, iter)
}

// buildListTimeSeriesRequest builds the ListTimeSeriesRequest from options.
func (g *Gcp) buildListTimeSeriesRequest(kt *kit.Kit, projectID string, opt *typesMonitoring.GcpListTimeSeriesOption) (
	*monitoringpb.ListTimeSeriesRequest, error) {

	// Build time interval
	interval, err := buildMonitorTimeInterval(kt, opt.Interval)
	if err != nil {
		logs.Errorf("build monitor time interval failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// Build request
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:      fmt.Sprintf("projects/%s", projectID),
		Filter:    opt.Filter,
		Interval:  interval,
		PageSize:  int32(opt.PageSize),
		PageToken: opt.PageToken,
	}

	// Set view
	view, ok := monitoringpb.ListTimeSeriesRequest_TimeSeriesView_value[string(opt.View)]
	if !ok {
		err := fmt.Errorf("invalid view: %s", opt.View)
		logs.Errorf("build list time series request failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	req.View = monitoringpb.ListTimeSeriesRequest_TimeSeriesView(view)

	// Set aggregation if specified
	if opt.Aggregation != nil {
		aggregation, err := buildMonitorAggregation(kt, opt.Aggregation)
		if err != nil {
			logs.Errorf("build monitor aggregation failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		req.Aggregation = aggregation
	}

	if opt.SecondaryAggregation != nil {
		secondaryAggregation, err := buildMonitorAggregation(kt, opt.SecondaryAggregation)
		if err != nil {
			logs.Errorf("build monitor secondary aggregation failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		req.SecondaryAggregation = secondaryAggregation
	}

	return req, nil
}

// collectTimeSeriesResults collects time series results directly from iterator response
// and converts to BK-HCM format. This avoids using iterator.NewPager which may cause pagination issues.
func (g *Gcp) collectTimeSeriesResults(kt *kit.Kit, iter *monitoring.TimeSeriesIterator) (
	*typesMonitoring.GcpListTimeSeriesResult, error) {

	timeSeries := make([]*typesMonitoring.GcpTimeSeries, 0)
	executionErrors := make([]*typesMonitoring.GcpExecutionError, 0)
	unit := ""
	nextPageToken := ""

	// Extract data directly from iterator's Response
	// The Response is populated after the first Next() call
	if iter.Response != nil {
		if resp, ok := iter.Response.(*monitoringpb.ListTimeSeriesResponse); ok {
			// Get all time series from response
			for _, ts := range resp.TimeSeries {
				converted := convertMonitorTimeSeriesFromProto(ts)
				timeSeries = append(timeSeries, converted)
			}

			// Extract metadata
			unit = resp.Unit
			nextPageToken = resp.NextPageToken
			for _, err := range resp.ExecutionErrors {
				executionErrors = append(executionErrors, &typesMonitoring.GcpExecutionError{
					Code:    err.Code,
					Message: err.Message,
				})
			}
		} else {
			logs.Warnf("iterator response is not ListTimeSeriesResponse type, actual type: %T, rid: %s",
				iter.Response, kt.Rid)
		}
	} else {
		logs.Warnf("iterator response is nil, rid: %s", kt.Rid)
	}

	return &typesMonitoring.GcpListTimeSeriesResult{
		TimeSeries:      timeSeries,
		NextPageToken:   nextPageToken,
		ExecutionErrors: executionErrors,
		Unit:            unit,
	}, nil
}

// buildMonitorTimeInterval converts BK-HCM GcpTimeInterval to GCP TimeInterval.
func buildMonitorTimeInterval(kt *kit.Kit, interval *typesMonitoring.GcpTimeInterval) (
	*monitoringpb.TimeInterval, error) {

	result := &monitoringpb.TimeInterval{}

	// Parse end time (required)
	endTime, err := time.Parse(time.RFC3339, interval.EndTime)
	if err != nil {
		logs.Errorf("parse end time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	result.EndTime = timestamppb.New(endTime)

	// Parse start time (optional)
	if interval.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, interval.StartTime)
		if err != nil {
			logs.Errorf("parse start time failed, err: %v, rid: %s", err, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		result.StartTime = timestamppb.New(startTime)
	}

	return result, nil
}

// buildMonitorAggregation converts BK-HCM GcpMonitoringAggregation to GCP Aggregation.
func buildMonitorAggregation(kt *kit.Kit, agg *typesMonitoring.GcpMonitoringAggregation) (
	*monitoringpb.Aggregation, error) {

	if agg == nil {
		return nil, errors.New("aggregation is nil")
	}

	result := &monitoringpb.Aggregation{}

	// Parse alignment period if specified
	if agg.AlignmentPeriod != "" {
		duration, err := time.ParseDuration(agg.AlignmentPeriod)
		if err != nil {
			logs.Errorf("parse alignment period failed, err: %v, rid: %s", err, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		result.AlignmentPeriod = durationpb.New(duration)
	}

	// Set per-series aligner
	if agg.PerSeriesAligner != "" {
		aligner, ok := monitoringpb.Aggregation_Aligner_value[string(agg.PerSeriesAligner)]
		if !ok {
			err := fmt.Errorf("invalid per_series_aligner: %s", agg.PerSeriesAligner)
			logs.Errorf("build monitor aggregation failed, err: %v, rid: %s", err, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		result.PerSeriesAligner = monitoringpb.Aggregation_Aligner(aligner)
	}

	// Set cross-series reducer
	if agg.CrossSeriesReducer != "" {
		reducer, ok := monitoringpb.Aggregation_Reducer_value[string(agg.CrossSeriesReducer)]
		if !ok {
			err := fmt.Errorf("invalid cross_series_reducer: %s", agg.CrossSeriesReducer)
			logs.Errorf("build monitor aggregation failed, err: %v, rid: %s", err, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		result.CrossSeriesReducer = monitoringpb.Aggregation_Reducer(reducer)
	}

	// Set group by fields
	if len(agg.GroupByFields) > 0 {
		result.GroupByFields = agg.GroupByFields
	}

	return result, nil
}

// convertMonitorTimeSeriesFromProto converts protobuf TimeSeries to BK-HCM GcpTimeSeries.
func convertMonitorTimeSeriesFromProto(pb *monitoringpb.TimeSeries) *typesMonitoring.GcpTimeSeries {
	if pb == nil {
		return nil
	}

	ts := &typesMonitoring.GcpTimeSeries{
		Unit: pb.Unit,
	}

	// Convert metric
	if pb.Metric != nil {
		ts.Metric = &typesMonitoring.GcpMetric{
			Type:   pb.Metric.Type,
			Labels: pb.Metric.Labels,
		}
	}

	// Convert resource
	if pb.Resource != nil {
		ts.Resource = &typesMonitoring.GcpMonitoredResource{
			Type:   pb.Resource.Type,
			Labels: pb.Resource.Labels,
		}
	}

	// Convert metric kind
	ts.MetricKind = typesMonitoring.GcpMetricKind(pb.MetricKind.String())

	// Convert value type
	ts.ValueType = typesMonitoring.GcpValueType(pb.ValueType.String())

	// Convert points
	if len(pb.Points) > 0 {
		ts.Points = make([]*typesMonitoring.GcpMonitoringPoint, 0, len(pb.Points))
		for _, p := range pb.Points {
			ts.Points = append(ts.Points, convertMonitorPointFromProto(p))
		}
	}

	return ts
}

// convertMonitorPointFromProto converts protobuf Point to BK-HCM GcpMonitoringPoint.
func convertMonitorPointFromProto(pb *monitoringpb.Point) *typesMonitoring.GcpMonitoringPoint {
	if pb == nil {
		return nil
	}

	point := &typesMonitoring.GcpMonitoringPoint{}

	// Convert interval
	if pb.Interval != nil {
		point.Interval = &typesMonitoring.GcpTimeInterval{}
		if pb.Interval.StartTime != nil {
			point.Interval.StartTime = pb.Interval.StartTime.AsTime().Format(time.RFC3339)
		}
		if pb.Interval.EndTime != nil {
			point.Interval.EndTime = pb.Interval.EndTime.AsTime().Format(time.RFC3339)
		}
	}

	// Convert typed value
	if pb.Value != nil {
		point.Value = &typesMonitoring.GcpTypedValue{}

		switch v := pb.Value.Value.(type) {
		case *monitoringpb.TypedValue_DoubleValue:
			val := v.DoubleValue
			point.Value.DoubleValue = &val
		case *monitoringpb.TypedValue_Int64Value:
			val := v.Int64Value
			point.Value.Int64Value = &val
		case *monitoringpb.TypedValue_BoolValue:
			val := v.BoolValue
			point.Value.BoolValue = &val
		case *monitoringpb.TypedValue_StringValue:
			val := v.StringValue
			point.Value.StringValue = &val
		case *monitoringpb.TypedValue_DistributionValue:
			if v.DistributionValue != nil {
				dist := &typesMonitoring.GcpTimeSeriesDistribution{
					Count:                 v.DistributionValue.Count,
					Mean:                  v.DistributionValue.Mean,
					SumOfSquaredDeviation: v.DistributionValue.SumOfSquaredDeviation,
					BucketCounts:          v.DistributionValue.BucketCounts,
				}
				if v.DistributionValue.Range != nil {
					dist.Range = &typesMonitoring.GcpRange{
						Min: v.DistributionValue.Range.Min,
						Max: v.DistributionValue.Range.Max,
					}
				}
				point.Value.DistributionValue = dist
			}
		}
	}

	return point
}
