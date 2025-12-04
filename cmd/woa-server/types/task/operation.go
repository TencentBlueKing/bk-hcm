/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package task define task operation
package task

import (
	"fmt"
	"time"

	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/querybuilder"
)

// GetApplyStatReq get resource apply operation statistics request
type GetApplyStatReq struct {
	Start     string                    `json:"start" bson:"start"`
	End       string                    `json:"end" bson:"end"`
	Dimension TimeDimension             `json:"dimension" bson:"dimension"`
	Filter    *querybuilder.QueryFilter `json:"filter" bson:"filter"`
}

// Validate whether GetApplyStatReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (req *GetApplyStatReq) Validate() (errKey string, err error) {
	start, err := time.Parse(dateLayout, req.Start)
	if err != nil {
		return "start", fmt.Errorf("date format should be like %s", dateLayout)
	}

	end, err := time.Parse(dateLayout, req.End)
	if err != nil {
		return "end", fmt.Errorf("date format should be like %s", dateLayout)
	}

	if len(req.Dimension) > 0 {
		if err := req.Dimension.Validate(); err != nil {
			return "dimension", err
		}
	}

	switch req.Dimension {
	// time range limit is different for different dimension
	case DimensionDay:
		if end.After(start.AddDate(0, 0, 90)) {
			return "start,end", fmt.Errorf("time range exeeds limit 90 days")
		}
	case DimensionMonth:
		if end.After(start.AddDate(0, 24, 0)) {
			return "start,end", fmt.Errorf("time range exeeds limit 24 months")
		}
	case DimensionYear:
		if end.After(start.AddDate(3, 0, 0)) {
			return "start,end", fmt.Errorf("time range exeeds limit 3 years")
		}
	}

	if req.Filter != nil {
		if key, err := req.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return fmt.Sprintf("filter.%s", key), err
		}
		if req.Filter.GetDeep() > querybuilder.MaxDeep {
			return "filter.rules", fmt.Errorf("exceed max query condition deepth: %d",
				querybuilder.MaxDeep)
		}
	}

	return "", nil
}

// GetFilter get mgo filter
func (req *GetApplyStatReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})

	if req.Filter != nil {
		mgoFilter, key, err := req.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		filter = mgoFilter
	}

	timeCond := make(map[string]interface{})
	if len(req.Start) != 0 {
		startTime, err := time.Parse(dateLayout, req.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}
	if len(req.End) != 0 {
		endTime, err := time.Parse(dateLayout, req.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}
	if len(timeCond) != 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// TimeDimension statistics time dimension
type TimeDimension string

// Validate whether TimeDimension is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param TimeDimension) Validate() (err error) {
	switch param {
	case DimensionDay, DimensionMonth, DimensionYear:
	default:
		return fmt.Errorf("unkown %s dimension type", param)
	}

	return nil
}

// TimeDimension statistics time dimension
const (
	DimensionDay   TimeDimension = "DAY"
	DimensionMonth TimeDimension = "MONTH"
	DimensionYear  TimeDimension = "YEAR"
)

// GetApplyStatRst get resource apply operation statistics result
type GetApplyStatRst struct {
	Info []*ApplyStat `json:"info" bson:"info"`
}

// ApplyStat resource apply operation statistics
type ApplyStat struct {
	Date            string  `json:"date" bson:"date"`
	OrderTotal      uint    `json:"order_total" bson:"order_total"`
	OrderSucc       uint    `json:"order_succ" bson:"order_succ"`
	OrderSuccRate   float64 `json:"order_succ_rate" bson:"order_succ_rate"`
	OrderManual     uint    `json:"order_manual" bson:"order_manual"`
	OrderManualRate float64 `json:"order_manual_rate" bson:"order_manual_rate"`
	OsTotal         uint    `json:"os_total" bson:"os_total"`
	OsSucc          uint    `json:"os_succ" bson:"os_succ"`
	OsSuccRate      float64 `json:"os_succ_rate" bson:"os_succ_rate"`
}

// AverageTimeConsumptionReq request for average time consumption overview
type AverageTimeConsumptionReq struct {
	// Date format: YYYY-MM-DD, e.g., 2025-01-01
	StartTime string `json:"start_time" bson:"start_time"`
	// Date format: YYYY-MM-DD, e.g., 2025-01-31
	EndTime string `json:"end_time" bson:"end_time"`
}

// Validate whether AverageTimeConsumptionReq is valid
func (req *AverageTimeConsumptionReq) Validate() (errKey string, err error) {
	if len(req.StartTime) == 0 {
		return "start_time", fmt.Errorf("start_time is not set")
	}
	if _, err := time.Parse(constant.DateLayout, req.StartTime); err != nil {
		return "start_time", fmt.Errorf("invalid start_time format, must be YYYY-MM-DD, e.g., 2025-01-01")
	}

	if len(req.EndTime) == 0 {
		return "end_time", fmt.Errorf("end_time is not set")
	}
	if _, err := time.Parse(constant.DateLayout, req.EndTime); err != nil {
		return "end_time", fmt.Errorf("invalid end_time format, must be YYYY-MM-DD, e.g., 2025-01-31")
	}
	return "", nil
}

// GetStartTime converts start_time string to time.Time
func (req *AverageTimeConsumptionReq) GetStartTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
	}
	// Return the date at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// GetEndTime converts end_time string to time.Time
func (req *AverageTimeConsumptionReq) GetEndTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
	}
	// Return the next day at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC), nil
}

// AverageTimeConsumptionItem one month aggregated metrics for average time consumption
type AverageTimeConsumptionItem struct {
	YearMonth        string  `json:"year_month" bson:"year_month"`
	AvgDurationHours float64 `json:"avg_duration_hours" bson:"avg_duration_hours"`
}

// AverageTimeConsumptionOverviewResp wraps overview list under details
type AverageTimeConsumptionOverviewResp struct {
	Details []AverageTimeConsumptionItem `json:"details"`
}

// AverageTimeConsumptionCompareReq request for average time consumption compare
type AverageTimeConsumptionCompareReq struct {
	// YearMonth format: YYYY-MM, e.g., 2025-10 for October 2025
	CurrentDate string `json:"current_date" bson:"current_date"`
	// YearMonth format: YYYY-MM, e.g., 2025-11 for November 2025
	CompareDate string `json:"compare_date" bson:"compare_date"`
}

// Validate whether AverageTimeConsumptionCompareReq is valid
func (req *AverageTimeConsumptionCompareReq) Validate() (errKey string, err error) {
	if _, err := time.Parse(constant.YearMonthLayout, req.CurrentDate); err != nil {
		return "current_date", fmt.Errorf("invalid current_date format, must be YYYY-MM (7 characters), e.g., 2025-10")
	}

	if _, err := time.Parse(constant.YearMonthLayout, req.CompareDate); err != nil {
		return "compare_date", fmt.Errorf("invalid compare_date format, must be YYYY-MM (7 characters), e.g., 2025-11")
	}
	return "", nil
}

// GetCurrentRange converts current_date to time range (start and end of month)
func (req *AverageTimeConsumptionCompareReq) GetCurrentRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CurrentDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid current_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// GetCompareRange converts compare_date to time range (start and end of month)
func (req *AverageTimeConsumptionCompareReq) GetCompareRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CompareDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid compare_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// AverageTimeConsumptionCompareItem one month aggregated metrics by biz for average time consumption compare
type AverageTimeConsumptionCompareItem struct {
	BkBizID          int64   `json:"bk_biz_id" bson:"bk_biz_id"`
	YearMonth        string  `json:"year_month" bson:"year_month"`
	DoneOrders       int64   `json:"done_orders" bson:"done_orders"`
	AvgDurationHours float64 `json:"avg_duration_hours" bson:"avg_duration_hours"`
}

// AverageTimeConsumptionCompareRst wraps compare result with current and compare arrays
type AverageTimeConsumptionCompareRst struct {
	Current []AverageTimeConsumptionCompareItem `json:"current"`
	Compare []AverageTimeConsumptionCompareItem `json:"compare"`
}

// OrderTimeCostReq request for order time cost overview
type OrderTimeCostReq struct {
	// Date format: YYYY-MM-DD, e.g., 2025-01-01
	StartTime string `json:"start_time" bson:"start_time"`
	// Date format: YYYY-MM-DD, e.g., 2025-01-31
	EndTime string `json:"end_time" bson:"end_time"`
}

// Validate whether OrderTimeCostReq is valid
func (req *OrderTimeCostReq) Validate() (errKey string, err error) {
	if _, err := time.Parse(constant.DateLayout, req.StartTime); err != nil {
		return "start_time", fmt.Errorf("invalid start_time format, must be YYYY-MM-DD, e.g., 2025-01-01")
	}

	if _, err := time.Parse(constant.DateLayout, req.EndTime); err != nil {
		return "end_time", fmt.Errorf("invalid end_time format, must be YYYY-MM-DD, e.g., 2025-01-31")
	}
	return "", nil
}

// GetStartTime converts start_time string to time.Time
func (req *OrderTimeCostReq) GetStartTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
	}
	// Return the date at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// GetEndTime converts end_time string to time.Time
func (req *OrderTimeCostReq) GetEndTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
	}
	// Return the next day at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC), nil
}

// OrderTimeCostItem one month aggregated metrics for order time cost
type OrderTimeCostItem struct {
	YearMonth        string  `json:"year_month" bson:"year_month"`
	AvgDurationHours float64 `json:"avg_duration_hours" bson:"avg_duration_hours"`
}

// OrderTimeCostOverviewResp wraps overview list under details
type OrderTimeCostOverviewResp struct {
	Details []OrderTimeCostItem `json:"details"`
}

// OrderTimeCostCompareReq request for order time cost compare
type OrderTimeCostCompareReq struct {
	// YearMonth format: YYYY-MM, e.g., 2025-10 for October 2025
	CurrentDate string `json:"current_date" bson:"current_date"`
	// YearMonth format: YYYY-MM, e.g., 2025-11 for November 2025
	CompareDate string `json:"compare_date" bson:"compare_date"`
}

// Validate whether OrderTimeCostCompareReq is valid
func (req *OrderTimeCostCompareReq) Validate() (errKey string, err error) {

	if _, err := time.Parse(constant.YearMonthLayout, req.CurrentDate); err != nil {
		return "current_date", fmt.Errorf("invalid current_date format, must be YYYY-MM (7 characters), e.g., 2025-10")
	}

	if _, err := time.Parse(constant.YearMonthLayout, req.CompareDate); err != nil {
		return "compare_date", fmt.Errorf("invalid compare_date format, must be YYYY-MM (7 characters), e.g., 2025-11")
	}
	return "", nil
}

// GetCurrentRange converts current_date to time range (start and end of month)
func (req *OrderTimeCostCompareReq) GetCurrentRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CurrentDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid current_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// GetCompareRange converts compare_date to time range (start and end of month)
func (req *OrderTimeCostCompareReq) GetCompareRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CompareDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid compare_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// OrderTimeCostCompareItem one month aggregated metrics by biz for order time cost compare
type OrderTimeCostCompareItem struct {
	BkBizID          int64   `json:"bk_biz_id" bson:"bk_biz_id"`
	YearMonth        string  `json:"year_month" bson:"year_month"`
	DoneOrders       int64   `json:"done_orders" bson:"done_orders"`
	AvgDurationHours float64 `json:"avg_duration_hours" bson:"avg_duration_hours"`
}

// OrderTimeCostCompareRst wraps compare result with current and compare arrays
type OrderTimeCostCompareRst struct {
	Current []OrderTimeCostCompareItem `json:"current"`
	Compare []OrderTimeCostCompareItem `json:"compare"`
}

// ProductionStageTimeCostReq request for production stage time cost overview
type ProductionStageTimeCostReq struct {
	// Date format: YYYY-MM-DD, e.g., 2025-01-01
	StartTime string `json:"start_time" bson:"start_time"`
	// Date format: YYYY-MM-DD, e.g., 2025-01-31
	EndTime string `json:"end_time" bson:"end_time"`
}

// Validate whether ProductionStageTimeCostReq is valid
func (req *ProductionStageTimeCostReq) Validate() (errKey string, err error) {

	if _, err := time.Parse(constant.DateLayout, req.StartTime); err != nil {
		return "start_time", fmt.Errorf("invalid start_time format, must be YYYY-MM-DD, e.g., 2025-01-01")
	}

	if _, err := time.Parse(constant.DateLayout, req.EndTime); err != nil {
		return "end_time", fmt.Errorf("invalid end_time format, must be YYYY-MM-DD, e.g., 2025-01-31")
	}
	return "", nil
}

// GetStartTime converts start_time string to time.Time
func (req *ProductionStageTimeCostReq) GetStartTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
	}
	// Return the date at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// GetEndTime converts end_time string to time.Time
func (req *ProductionStageTimeCostReq) GetEndTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
	}
	// Return the next day at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC), nil
}

// ProductionStageTimeCostItem one month aggregated metrics for production stage time cost
type ProductionStageTimeCostItem struct {
	YearMonth        string  `json:"year_month" bson:"year_month"`
	AvgDurationHours float64 `json:"avg_duration_hours" bson:"avg_duration_hours"`
}

// ProductionStageTimeCostOverviewResp wraps overview list under details
type ProductionStageTimeCostOverviewResp struct {
	Details []ProductionStageTimeCostItem `json:"details"`
}

// ProductionStageTimeCostCompareReq request for production stage time cost compare
type ProductionStageTimeCostCompareReq struct {
	// YearMonth format: YYYY-MM, e.g., 2025-10 for October 2025
	CurrentDate string `json:"current_date" bson:"current_date"`
	// YearMonth format: YYYY-MM, e.g., 2025-11 for November 2025
	CompareDate string `json:"compare_date" bson:"compare_date"`
}

// Validate whether ProductionStageTimeCostCompareReq is valid
func (req *ProductionStageTimeCostCompareReq) Validate() (errKey string, err error) {

	if _, err := time.Parse(constant.YearMonthLayout, req.CurrentDate); err != nil {
		return "current_date", fmt.Errorf("invalid current_date format, must be YYYY-MM (7 characters), e.g., 2025-10")
	}

	if _, err := time.Parse(constant.YearMonthLayout, req.CompareDate); err != nil {
		return "compare_date", fmt.Errorf("invalid compare_date format, must be YYYY-MM (7 characters), e.g., 2025-11")
	}
	return "", nil
}

// GetCurrentRange converts current_date to time range (start and end of month)
func (req *ProductionStageTimeCostCompareReq) GetCurrentRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CurrentDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid current_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// GetCompareRange converts compare_date to time range (start and end of month)
func (req *ProductionStageTimeCostCompareReq) GetCompareRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CompareDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid compare_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// ProductionStageTimeCostBizItem one month aggregated metrics by biz for production stage time cost compare
type ProductionStageTimeCostBizItem struct {
	BkBizID          int64   `json:"bk_biz_id" bson:"bk_biz_id"`
	YearMonth        string  `json:"year_month" bson:"year_month"`
	DoneOrders       int64   `json:"done_orders" bson:"done_orders"`
	AvgDurationHours float64 `json:"avg_duration_hours" bson:"avg_duration_hours"`
}

// ProductionStageTimeCostCompareRst wraps compare result with current and compare arrays
type ProductionStageTimeCostCompareRst struct {
	Current []ProductionStageTimeCostBizItem `json:"current"`
	Compare []ProductionStageTimeCostBizItem `json:"compare"`
}

// PercentileTimeConsumptionReq request for percentile time consumption overview
type PercentileTimeConsumptionReq struct {
	// Date format: YYYY-MM-DD, e.g., 2025-01-01
	StartTime string `json:"start_time" bson:"start_time"`
	// Date format: YYYY-MM-DD, e.g., 2025-01-31
	EndTime string `json:"end_time" bson:"end_time"`
}

// Validate whether PercentileTimeConsumptionReq is valid
func (req *PercentileTimeConsumptionReq) Validate() (errKey string, err error) {

	if _, err := time.Parse(constant.DateLayout, req.StartTime); err != nil {
		return "start_time", fmt.Errorf("invalid start_time format, must be YYYY-MM-DD, e.g., 2025-01-01")
	}

	if _, err := time.Parse(constant.DateLayout, req.EndTime); err != nil {
		return "end_time", fmt.Errorf("invalid end_time format, must be YYYY-MM-DD, e.g., 2025-01-31")
	}
	return "", nil
}

// GetStartTime converts start_time string to time.Time
func (req *PercentileTimeConsumptionReq) GetStartTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
	}
	// Return the date at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// GetEndTime converts end_time string to time.Time
func (req *PercentileTimeConsumptionReq) GetEndTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
	}
	// Return the next day at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC), nil
}

// PercentileTimeConsumptionItem one month aggregated metrics for percentile time consumption
type PercentileTimeConsumptionItem struct {
	YearMonth string  `json:"year_month" bson:"year_month"`
	P90Hours  float64 `json:"p90_hours" bson:"p90_hours"`
	P95Hours  float64 `json:"p95_hours" bson:"p95_hours"`
	P99Hours  float64 `json:"p99_hours" bson:"p99_hours"`
}

// PercentileTimeConsumptionOverviewResp wraps overview list under details
type PercentileTimeConsumptionOverviewResp struct {
	Details []PercentileTimeConsumptionItem `json:"details"`
}

// PercentileTimeConsumptionCompareReq request for percentile time consumption compare
type PercentileTimeConsumptionCompareReq struct {
	// YearMonth format: YYYY-MM, e.g., 2025-10 for October 2025
	CurrentDate string `json:"current_date" bson:"current_date"`
	// YearMonth format: YYYY-MM, e.g., 2025-11 for November 2025
	CompareDate string `json:"compare_date" bson:"compare_date"`
}

// Validate whether PercentileTimeConsumptionCompareReq is valid
func (req *PercentileTimeConsumptionCompareReq) Validate() (errKey string, err error) {

	if _, err := time.Parse(constant.YearMonthLayout, req.CurrentDate); err != nil {
		return "current_date", fmt.Errorf("invalid current_date format, must be YYYY-MM (7 characters), e.g., 2025-10")
	}

	if _, err := time.Parse(constant.YearMonthLayout, req.CompareDate); err != nil {
		return "compare_date", fmt.Errorf("invalid compare_date format, must be YYYY-MM (7 characters), e.g., 2025-11")
	}
	return "", nil
}

// GetCurrentRange converts current_date to time range (start and end of month)
func (req *PercentileTimeConsumptionCompareReq) GetCurrentRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CurrentDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid current_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// GetCompareRange converts compare_date to time range (start and end of month)
func (req *PercentileTimeConsumptionCompareReq) GetCompareRange() (start time.Time, end time.Time, err error) {
	t, err := time.Parse(constant.YearMonthLayout, req.CompareDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid compare_date format: %w", err)
	}
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	end = nextMonth.Add(-time.Nanosecond)
	return start, end, nil
}

// PercentileTimeConsumptionCompareItem one month aggregated metrics by biz for percentile time consumption compare
type PercentileTimeConsumptionCompareItem struct {
	BkBizID    int64   `json:"bk_biz_id" bson:"bk_biz_id"`
	YearMonth  string  `json:"year_month" bson:"year_month"`
	DoneOrders int64   `json:"done_orders" bson:"done_orders"`
	P90Hours   float64 `json:"p90_hours" bson:"p90_hours"`
	P95Hours   float64 `json:"p95_hours" bson:"p95_hours"`
	P99Hours   float64 `json:"p99_hours" bson:"p99_hours"`
}

// PercentileTimeConsumptionCompareRst wraps compare result with current and compare arrays
type PercentileTimeConsumptionCompareRst struct {
	Current []PercentileTimeConsumptionCompareItem `json:"current"`
	Compare []PercentileTimeConsumptionCompareItem `json:"compare"`
}

// DeliveryRateStatisticsReq request for delivery rate statistics
type DeliveryRateStatisticsReq struct {
	// Date format: YYYY-MM-DD, e.g., 2025-01-01
	StartTime string `json:"start_time" bson:"start_time"`
	// Date format: YYYY-MM-DD, e.g., 2025-06-30
	EndTime string `json:"end_time" bson:"end_time"`
}

// Validate whether DeliveryRateStatisticsReq is valid
func (req *DeliveryRateStatisticsReq) Validate() (errKey string, err error) {

	if _, err := time.Parse(constant.DateLayout, req.StartTime); err != nil {
		return "start_time", fmt.Errorf("invalid start_time format, must be YYYY-MM-DD, e.g., 2025-01-01")
	}

	if _, err := time.Parse(constant.DateLayout, req.EndTime); err != nil {
		return "end_time", fmt.Errorf("invalid end_time format, must be YYYY-MM-DD, e.g., 2025-06-30")
	}
	return "", nil
}

// GetStartTime converts start_time string to time.Time
func (req *DeliveryRateStatisticsReq) GetStartTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
	}
	// Return the date at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// GetEndTime converts end_time string to time.Time
func (req *DeliveryRateStatisticsReq) GetEndTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
	}
	// Return the next day at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC), nil
}

// DeliveryRateStatisticsItem one month aggregated metrics for delivery rate statistics
type DeliveryRateStatisticsItem struct {
	YearMonth    string  `json:"year_month" bson:"year_month"`
	DeliveryRate float64 `json:"delivery_rate" bson:"delivery_rate"`
}

// DeliveryRateStatisticsResp wraps statistics list under details
type DeliveryRateStatisticsResp struct {
	Details []DeliveryRateStatisticsItem `json:"details"`
}

// DeliveryRateDetailReq request for delivery rate detail
type DeliveryRateDetailReq struct {
	// Date format: YYYY-MM-DD, e.g., 2025-10-01
	StartTime string `json:"start_time" bson:"start_time"`
	// Date format: YYYY-MM-DD, e.g., 2025-10-31
	EndTime string `json:"end_time" bson:"end_time"`
}

// Validate whether DeliveryRateDetailReq is valid
func (req *DeliveryRateDetailReq) Validate() (errKey string, err error) {
	if _, err := time.Parse(constant.DateLayout, req.StartTime); err != nil {
		return "start_time", fmt.Errorf("invalid start_time format, must be YYYY-MM-DD, e.g., 2025-10-01")
	}

	if _, err := time.Parse(constant.DateLayout, req.EndTime); err != nil {
		return "end_time", fmt.Errorf("invalid end_time format, must be YYYY-MM-DD, e.g., 2025-10-31")
	}
	return "", nil
}

// GetStartTime converts start_time string to time.Time
func (req *DeliveryRateDetailReq) GetStartTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start_time format: %w", err)
	}
	// Return the date at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// GetEndTime converts end_time string to time.Time
func (req *DeliveryRateDetailReq) GetEndTime() (time.Time, error) {
	t, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time format: %w", err)
	}
	// Return the next day at 00:00:00 UTC
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC), nil
}

// DeliveryRateDetailItem one month aggregated metrics by biz for delivery rate detail
type DeliveryRateDetailItem struct {
	BkBizID          int64   `json:"bk_biz_id" bson:"bk_biz_id"`
	YearMonth        string  `json:"year_month" bson:"year_month"`
	TotalOrders      int64   `json:"total_orders" bson:"total_orders"`
	DoneOrders       int64   `json:"done_orders" bson:"done_orders"`
	TotalNumSum      int64   `json:"total_num_sum" bson:"total_num_sum"`
	SuccessNumSum    int64   `json:"success_num_sum" bson:"success_num_sum"`
	HostDeliveryRate float64 `json:"host_delivery_rate" bson:"host_delivery_rate"`
}

// DeliveryRateDetailResp wraps detail result with details array
type DeliveryRateDetailResp struct {
	Details []DeliveryRateDetailItem `json:"details"`
}

// GetCompletionRateStatReq get completion rate statistics request
type GetCompletionRateStatReq struct {
	StartTime string `json:"start_time" bson:"start_time"`
	EndTime   string `json:"end_time" bson:"end_time"`
}

// Validate whether GetCompletionRateStatReq is valid
func (req *GetCompletionRateStatReq) Validate() error {
	startTime, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start_time, expected format %s", constant.DateLayout)
	}

	endTime, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return fmt.Errorf("invalid end_time, expected format %s", constant.DateLayout)
	}

	if endTime.Before(startTime) {
		return fmt.Errorf("end_time must be after start_time")

	}

	return nil
}

// GetFilter get mgo filter
func (req *GetCompletionRateStatReq) GetFilter() (map[string]interface{}, error) {
	startTime, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time, expected format %s", constant.DateLayout)
	}

	endTime, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time, expected format %s", constant.DateLayout)
	}

	filter := make(map[string]interface{})
	timeCond := map[string]interface{}{
		pkg.BKDBGTE: startTime,
		// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
		pkg.BKDBLT: endTime.AddDate(0, 0, 1),
	}
	filter["create_at"] = timeCond

	filter["source"] = map[string]interface{}{
		pkg.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool,
	}

	return filter, nil
}

// GetCompletionRateStatRst get completion rate statistics result
type GetCompletionRateStatRst struct {
	Details []*CompletionRateStat `json:"details" bson:"details"`
}

// CompletionRateStat completion rate statistics
type CompletionRateStat struct {
	YearMonth      string  `json:"year_month" bson:"year_month"`
	CompletionRate float64 `json:"completion_rate" bson:"completion_rate"`
}

// GetCompletionRateDetailReq 获取结单率详情统计请求
type GetCompletionRateDetailReq struct {
	StartTime string `json:"start_time" bson:"start_time"` // 开始时间，格式：YYYY-MM-DD
	EndTime   string `json:"end_time" bson:"end_time"`     // 结束时间，格式：YYYY-MM-DD
}

// Validate 验证请求参数
func (req *GetCompletionRateDetailReq) Validate() error {
	startTime, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start_time, expected format %s", constant.DateLayout)
	}

	endTime, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return fmt.Errorf("invalid end_time, expected format %s", constant.DateLayout)
	}

	if endTime.Before(startTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	return nil
}

// GetCompletionRateDetailRst 获取结单率详情统计响应
type GetCompletionRateDetailRst struct {
	Details []*CompletionRateDetailItem `json:"details" bson:"details"`
}

// CompletionRateDetailItem 结单率详情统计项
type CompletionRateDetailItem struct {
	BkBizID        int64   `json:"bk_biz_id" bson:"bk_biz_id"`             // 业务ID
	YearMonth      string  `json:"year_month" bson:"year_month"`           // 年月，格式：YYYY-MM
	TotalOrders    int     `json:"total_orders" bson:"total_orders"`       // 总单据数
	DoneOrders     int     `json:"done_orders" bson:"done_orders"`         // 已完成单据数（stage=DONE）
	CompletionRate float64 `json:"completion_rate" bson:"completion_rate"` // 结单率（百分比），保留2位小数
}

// GetApplyBizTopStatReq get apply biz top statistics request
type GetApplyBizTopStatReq struct {
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
}

// ParseAndValidate validate GetApplyBizTopStatReq
func (req *GetApplyBizTopStatReq) ParseAndValidate() (time.Time, time.Time, error) {
	if err := validator.Validate.Struct(req); err != nil {
		return time.Time{}, time.Time{}, err
	}

	start, err := time.Parse(constant.DateLayout, req.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("date format should be like %s", constant.DateLayout)
	}

	if len(req.EndTime) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("end_time is not set")
	}

	end, err := time.Parse(constant.DateLayout, req.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("date format should be like %s", constant.DateLayout)
	}

	// 开始时间、结束时间最长1年
	if end.After(start.AddDate(1, 0, 0)) {
		return time.Time{}, time.Time{}, fmt.Errorf("time range exceeds limit 1 year")
	}

	return start, end, nil
}

// ApplyBizHostsStatisticsItem 申请主机数-业务统计单项
type ApplyBizHostsStatisticsItem struct {
	BkBizID    int64 `json:"bk_biz_id" bson:"bk_biz_id"`
	HostCount  uint  `json:"host_count" bson:"host_count"`
	OrderCount int   `json:"order_count" bson:"order_count"`
}

// ApplyBizHostsStatisticsResult 申请主机数TOP10的业务统计结果
type ApplyBizHostsStatisticsResult = core.ListResultT[ApplyBizHostsStatisticsItem]

// ApplyBizCpuCoresStatisticsItem 申请核心数-业务统计单项
type ApplyBizCpuCoresStatisticsItem struct {
	BkBizID            int64 `json:"bk_biz_id" bson:"bk_biz_id"`
	DeliveredCoreCount uint  `json:"delivered_core_count" bson:"delivered_core_count"`
	OrderCount         int   `json:"order_count" bson:"order_count"`
}

// ApplyBizCpuCoresStatisticsResult 申请核心数TOP10的业务统计结果
type ApplyBizCpuCoresStatisticsResult = core.ListResultT[ApplyBizCpuCoresStatisticsItem]
