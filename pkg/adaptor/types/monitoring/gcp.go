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

package monitoring

import (
	"errors"
	"fmt"
	"time"

	"hcm/pkg/criteria/validator"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
)

// GcpListTimeSeriesOption defines options for listing GCP time series data.
type GcpListTimeSeriesOption struct {
	// Filter is required. A monitoring filter that specifies which time series should be returned.
	// The filter must specify a single metric type, and can additionally specify metric labels and
	// other information. For example:
	//   metric.type = "compute.googleapis.com/instance/cpu/utilization" AND
	//   resource.type = "gce_instance"
	Filter string `json:"filter" validate:"required"`

	// Interval is the time interval for which results should be returned.
	Interval *GcpTimeInterval `json:"interval" validate:"required"`

	// Aggregation defines how to combine multiple time series into a single time series.
	// Optional. If not set, raw data is returned.
	Aggregation *GcpMonitoringAggregation `json:"aggregation,omitempty"`

	// SecondaryAggregation applies a second aggregation after the primary aggregation is applied.
	// Optional. Only valid if Aggregation is specified.
	SecondaryAggregation *GcpMonitoringAggregation `json:"secondary_aggregation,omitempty"`

	// View specifies which information is returned about the time series.
	// Required. Possible values: TimeSeriesViewFull, TimeSeriesViewHeaders
	View GcpTimeSeriesView `json:"view" validate:"required"`

	// PageSize is the maximum number of results to return in a single response.
	// Maximum value: 100,000.
	PageSize int `json:"page_size,omitempty" validate:"omitempty,gt=0,lte=100000"`

	// PageToken is used to request the next page of results.
	PageToken string `json:"page_token,omitempty"`
}

// Validate validates the GcpListTimeSeriesOption.
func (o *GcpListTimeSeriesOption) Validate() error {
	if err := validator.Validate.Struct(o); err != nil {
		return err
	}

	if o.Interval != nil {
		if err := o.Interval.Validate(); err != nil {
			return err
		}
	}

	if o.Aggregation != nil {
		if err := o.Aggregation.Validate(); err != nil {
			return err
		}
	}

	if o.SecondaryAggregation != nil {
		if o.Aggregation == nil {
			return fmt.Errorf("secondary_aggregation can only be specified if aggregation is specified")
		}
		if err := o.SecondaryAggregation.Validate(); err != nil {
			return err
		}
	}

	if _, ok := monitoringpb.ListTimeSeriesRequest_TimeSeriesView_value[string(o.View)]; !ok {
		return fmt.Errorf("invalid view: %s", o.View)
	}

	return nil
}

// GcpTimeInterval represents a time interval for querying time series data.
type GcpTimeInterval struct {
	// StartTime is the beginning of the time interval.
	// Optional. If not specified, the interval is a point in time.
	// Format: RFC3339 (e.g., "2023-10-27T09:00:00Z")
	StartTime string `json:"start_time,omitempty"`

	// EndTime is the end of the time interval.
	// Required. Format: RFC3339 (e.g., "2023-10-27T10:00:00Z")
	EndTime string `json:"end_time" validate:"required"`
}

// Validate validates the GcpTimeInterval.
func (t *GcpTimeInterval) Validate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	// Parse and validate RFC3339 timestamps
	endTime, err := time.Parse(time.RFC3339, t.EndTime)
	if err != nil {
		return err
	}

	if t.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, t.StartTime)
		if err != nil {
			return err
		}

		if !startTime.Before(endTime) {
			return errors.New("start_time must be before end_time")
		}
	}

	return nil
}

// GcpMonitoringAggregation describes how to combine multiple time series.
type GcpMonitoringAggregation struct {
	// AlignmentPeriod is the alignment period for per-time series alignment.
	// Format: duration string (e.g., "60s", "1m")
	AlignmentPeriod string `json:"alignment_period,omitempty"`

	// PerSeriesAligner specifies the aligner to apply to each time series.
	// Optional. See GcpAggregationAligner constants for valid values.
	PerSeriesAligner GcpAggregationAligner `json:"per_series_aligner,omitempty"`

	// CrossSeriesReducer defines how to combine time series into a single time series.
	// Optional. See GcpAggregationReducer constants for valid values.
	CrossSeriesReducer GcpAggregationReducer `json:"cross_series_reducer,omitempty"`

	// GroupByFields lists the field names to use for grouping.
	GroupByFields []string `json:"group_by_fields,omitempty"`
}

// Validate validates the GcpMonitoringAggregation.
func (a *GcpMonitoringAggregation) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if a.PerSeriesAligner != "" {
		if _, ok := monitoringpb.Aggregation_Aligner_value[string(a.PerSeriesAligner)]; !ok {
			return fmt.Errorf("invalid per_series_aligner: %s", a.PerSeriesAligner)
		}
	}

	if a.CrossSeriesReducer != "" {
		if _, ok := monitoringpb.Aggregation_Reducer_value[string(a.CrossSeriesReducer)]; !ok {
			return fmt.Errorf("invalid cross_series_reducer: %s", a.CrossSeriesReducer)
		}
	}

	return nil
}

// GcpListTimeSeriesResult represents the result of listing time series.
type GcpListTimeSeriesResult struct {
	// TimeSeries is the list of time series data.
	TimeSeries []*GcpTimeSeries `json:"time_series"`

	// NextPageToken is used to retrieve the next page of results.
	NextPageToken string `json:"next_page_token,omitempty"`

	// ExecutionErrors contains any errors that occurred during execution.
	ExecutionErrors []*GcpExecutionError `json:"execution_errors,omitempty"`

	// Unit is the unit of measurement for the metric (applies to all time series if they share the same unit).
	Unit string `json:"unit,omitempty"`
}

// GcpExecutionError represents an error that occurred during query execution.
type GcpExecutionError struct {
	// Code is the error code (Google RPC code).
	Code int32 `json:"code,omitempty"`

	// Message is the error message.
	Message string `json:"message,omitempty"`
}

// GcpTimeSeries represents a single time series data with snake_case JSON tags.
type GcpTimeSeries struct {
	// Metric information
	Metric *GcpMetric `json:"metric,omitempty"`

	// Resource information
	Resource *GcpMonitoredResource `json:"resource,omitempty"`

	// MetricKind is the metric kind (GAUGE, DELTA, CUMULATIVE)
	MetricKind GcpMetricKind `json:"metric_kind,omitempty"`

	// ValueType is the value type (DOUBLE, INT64, BOOL, STRING, DISTRIBUTION)
	ValueType GcpValueType `json:"value_type,omitempty"`

	// Points is the list of data points
	Points []*GcpMonitoringPoint `json:"points,omitempty"`

	// Unit is the unit of measurement
	Unit string `json:"unit,omitempty"`
}

// GcpMetric describes a metric.
type GcpMetric struct {
	// Type is the metric type identifier
	Type string `json:"type,omitempty"`

	// Labels are the metric labels
	Labels map[string]string `json:"labels,omitempty"`
}

// GcpMonitoredResource describes a monitored resource.
type GcpMonitoredResource struct {
	// Type is the resource type
	Type string `json:"type,omitempty"`

	// Labels are the resource labels
	Labels map[string]string `json:"labels,omitempty"`
}

// GcpMonitoringPoint represents a single data point.
type GcpMonitoringPoint struct {
	// Interval is the time interval for this point
	Interval *GcpTimeInterval `json:"interval,omitempty"`

	// Value is the data value
	Value *GcpTypedValue `json:"value,omitempty"`
}

// GcpTypedValue represents a typed value that can hold different types.
type GcpTypedValue struct {
	// DoubleValue for DOUBLE type
	DoubleValue *float64 `json:"double_value,omitempty"`

	// Int64Value for INT64 type
	Int64Value *int64 `json:"int64_value,omitempty"`

	// BoolValue for BOOL type
	BoolValue *bool `json:"bool_value,omitempty"`

	// StringValue for STRING type
	StringValue *string `json:"string_value,omitempty"`

	// DistributionValue for DISTRIBUTION type
	DistributionValue *GcpTimeSeriesDistribution `json:"distribution_value,omitempty"`
}

// GcpTimeSeriesDistribution represents a distribution value.
type GcpTimeSeriesDistribution struct {
	// Count is the number of values in the population
	Count int64 `json:"count,omitempty"`

	// Mean is the arithmetic mean of the values
	Mean float64 `json:"mean,omitempty"`

	// SumOfSquaredDeviation is the sum of squared deviations from the mean
	SumOfSquaredDeviation float64 `json:"sum_of_squared_deviation,omitempty"`

	// Range is the range of values [min, max]
	Range *GcpRange `json:"range,omitempty"`

	// BucketCounts contains the counts for each histogram bucket
	BucketCounts []int64 `json:"bucket_counts,omitempty"`
}

// GcpRange represents a range of values.
type GcpRange struct {
	// Min is the minimum value
	Min float64 `json:"min,omitempty"`

	// Max is the maximum value
	Max float64 `json:"max,omitempty"`
}

// GcpMetricKind describes how a metric is reported.
type GcpMetricKind string

const (
	// GcpMetricKindUnspecified - Do not use this default value.
	GcpMetricKindUnspecified GcpMetricKind = "METRIC_KIND_UNSPECIFIED"
	// GcpMetricKindGauge - An instantaneous measurement of a value.
	// GcpExample: current CPU usage (0-100%)
	GcpMetricKindGauge GcpMetricKind = "GAUGE"
	// GcpMetricKindDelta - The change in a value during a time interval.
	// GcpExample: number of API requests in the last 5 minutes
	GcpMetricKindDelta GcpMetricKind = "DELTA"
	// GcpMetricKindCumulative - A value accumulated over a time interval.
	// Example: total requests since service started
	GcpMetricKindCumulative GcpMetricKind = "CUMULATIVE"
)

// GcpValueType describes the type of metric values.
type GcpValueType string

const (
	// GcpValueTypeUnspecified - Do not use this default value.
	GcpValueTypeUnspecified GcpValueType = "VALUE_TYPE_UNSPECIFIED"
	// GcpValueTypeBool - A boolean value. Only valid with GAUGE metric kind.
	GcpValueTypeBool GcpValueType = "BOOL"
	// GcpValueTypeInt64 - A signed 64-bit integer.
	GcpValueTypeInt64 GcpValueType = "INT64"
	// GcpValueTypeDouble - A double precision floating point number.
	GcpValueTypeDouble GcpValueType = "DOUBLE"
	// GcpValueTypeString - A text string. Only valid with GAUGE metric kind.
	GcpValueTypeString GcpValueType = "STRING"
	// GcpValueTypeDistribution - A distribution value (histogram).
	GcpValueTypeDistribution GcpValueType = "DISTRIBUTION"
)

// GcpTimeSeriesView controls which information is returned in the response.
type GcpTimeSeriesView string

const (
	// GcpTimeSeriesViewFull - Returns complete time series data with metric metadata and data points.
	GcpTimeSeriesViewFull GcpTimeSeriesView = "FULL"
	// GcpTimeSeriesViewHeaders - Returns only metric metadata (headers) without data points.
	GcpTimeSeriesViewHeaders GcpTimeSeriesView = "HEADERS"
)

// GcpAggregationAligner describes how per-series data should be aligned.
type GcpAggregationAligner string

const (
	// GcpAggregationAlignNone - No alignment - keep raw timestamps.
	GcpAggregationAlignNone GcpAggregationAligner = "ALIGN_NONE"
	// GcpAggregationAlignDelta - Difference between the last point and first point in the window.
	GcpAggregationAlignDelta GcpAggregationAligner = "ALIGN_DELTA"
	// GcpAggregationAlignRate - Per-second rate of change (for DELTA and CUMULATIVE metrics).
	GcpAggregationAlignRate GcpAggregationAligner = "ALIGN_RATE"
	// GcpAggregationAlignInterpolate - Linear interpolation at window boundaries.
	GcpAggregationAlignInterpolate GcpAggregationAligner = "ALIGN_INTERPOLATE"
	// GcpAggregationAlignNextOlder - The closest older data point in the window.
	GcpAggregationAlignNextOlder GcpAggregationAligner = "ALIGN_NEXT_OLDER"
	// GcpAggregationAlignMin - Minimum value in the window.
	GcpAggregationAlignMin GcpAggregationAligner = "ALIGN_MIN"
	// GcpAggregationAlignMax - Maximum value in the window.
	GcpAggregationAlignMax GcpAggregationAligner = "ALIGN_MAX"
	// GcpAggregationAlignMean - Arithmetic mean of values in the window.
	GcpAggregationAlignMean GcpAggregationAligner = "ALIGN_MEAN"
	// GcpAggregationAlignCount - Count of data points in the window.
	GcpAggregationAlignCount GcpAggregationAligner = "ALIGN_COUNT"
	// GcpAggregationAlignSum - Sum of values in the window.
	GcpAggregationAlignSum GcpAggregationAligner = "ALIGN_SUM"
	// GcpAggregationAlignStdDev - Standard deviation of values in the window.
	GcpAggregationAlignStdDev GcpAggregationAligner = "ALIGN_STDDEV"
	// GcpAggregationAlignCountTrue - Count of true BOOL values in the window.
	GcpAggregationAlignCountTrue GcpAggregationAligner = "ALIGN_COUNT_TRUE"
	// GcpAggregationAlignCountFalse - Count of false BOOL values in the window.
	GcpAggregationAlignCountFalse GcpAggregationAligner = "ALIGN_COUNT_FALSE"
	// GcpAggregationAlignFractionTrue - Fraction of true BOOL values in the window.
	GcpAggregationAlignFractionTrue GcpAggregationAligner = "ALIGN_FRACTION_TRUE"
	// GcpAggregationAlignPercentile99 - 99th percentile of values in the window.
	GcpAggregationAlignPercentile99 GcpAggregationAligner = "ALIGN_PERCENTILE_99"
	// GcpAggregationAlignPercentile95 - 95th percentile of values in the window.
	GcpAggregationAlignPercentile95 GcpAggregationAligner = "ALIGN_PERCENTILE_95"
	// GcpAggregationAlignPercentile50 - 50th percentile (median) of values in the window.
	GcpAggregationAlignPercentile50 GcpAggregationAligner = "ALIGN_PERCENTILE_50"
	// GcpAggregationAlignPercentile05 - 5th percentile of values in the window.
	GcpAggregationAlignPercentile05 GcpAggregationAligner = "ALIGN_PERCENTILE_05"
	// GcpAggregationAlignPercentChange - Percent change between first and last point in the window.
	GcpAggregationAlignPercentChange GcpAggregationAligner = "ALIGN_PERCENT_CHANGE"
)

// GcpAggregationReducer describes how to combine time series.
type GcpAggregationReducer string

const (
	// GcpAggregationReduceNone - No cross-series reduction.
	GcpAggregationReduceNone GcpAggregationReducer = "REDUCE_NONE"
	// GcpAggregationReduceMean - Mean value across all series.
	GcpAggregationReduceMean GcpAggregationReducer = "REDUCE_MEAN"
	// GcpAggregationReduceMin - Minimum value across all series.
	GcpAggregationReduceMin GcpAggregationReducer = "REDUCE_MIN"
	// GcpAggregationReduceMax - Maximum value across all series.
	GcpAggregationReduceMax GcpAggregationReducer = "REDUCE_MAX"
	// GcpAggregationReduceSum - Sum across all series.
	GcpAggregationReduceSum GcpAggregationReducer = "REDUCE_SUM"
	// GcpAggregationReduceStdDev - Standard deviation across all series.
	GcpAggregationReduceStdDev GcpAggregationReducer = "REDUCE_STDDEV"
	// GcpAggregationReduceCount - Count of series.
	GcpAggregationReduceCount GcpAggregationReducer = "REDUCE_COUNT"
	// GcpAggregationReduceCountTrue - Count of true BOOL values across all series.
	GcpAggregationReduceCountTrue GcpAggregationReducer = "REDUCE_COUNT_TRUE"
	// GcpAggregationReduceCountFalse - Count of false BOOL values across all series.
	GcpAggregationReduceCountFalse GcpAggregationReducer = "REDUCE_COUNT_FALSE"
	// GcpAggregationReduceFractionTrue - Fraction of true BOOL values across all series.
	GcpAggregationReduceFractionTrue GcpAggregationReducer = "REDUCE_FRACTION_TRUE"
	// GcpAggregationReducePercentile99 - 99th percentile across all series.
	GcpAggregationReducePercentile99 GcpAggregationReducer = "REDUCE_PERCENTILE_99"
	// GcpAggregationReducePercentile95 - 95th percentile across all series.
	GcpAggregationReducePercentile95 GcpAggregationReducer = "REDUCE_PERCENTILE_95"
	// GcpAggregationReducePercentile50 - 50th percentile (median) across all series.
	GcpAggregationReducePercentile50 GcpAggregationReducer = "REDUCE_PERCENTILE_50"
	// GcpAggregationReducePercentile05 - 5th percentile across all series.
	GcpAggregationReducePercentile05 GcpAggregationReducer = "REDUCE_PERCENTILE_05"
)
