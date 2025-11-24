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

package operation

import (
	"time"

	model "hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetPercentileTimeConsumptionOverview aggregates percentile time consumption by month within a range
func (op *operation) GetPercentileTimeConsumptionOverview(kt *kit.Kit, param *types.PercentileTimeConsumptionReq) ([]types.PercentileTimeConsumptionItem, error) {
	start, err := param.GetStartTime()
	if err != nil {
		logs.Errorf("parse start time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	end, err := param.GetEndTime()
	if err != nil {
		logs.Errorf("parse end time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// Get exclude suborder IDs
	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, start, end)
	if err != nil {
		logs.Errorf("get exclude suborder IDs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pipeline := op.buildPercentileTimeConsumptionOverviewPipeline(start, end, excludeSuborderIDs)

	rst := make([]types.PercentileTimeConsumptionItem, 0)
	if err := model.Operation().ApplyTicket().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate percentile time consumption overview failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}

// buildPercentileTimeConsumptionOverviewPipeline builds the aggregation pipeline for
// percentile time consumption overview
func (op *operation) buildPercentileTimeConsumptionOverviewPipeline(start, end time.Time,
	excludeSuborderIDs []string) []map[string]interface{} {

	match := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: start,
			pkg.BKDBLTE: end,
		},
	}

	return []map[string]interface{}{
		{pkg.BKDBMatch: match},
		{pkg.BKDBLookup: map[string]interface{}{
			"from":         pkg.BKTableNameApplyOrder,
			"localField":   "order_id",
			"foreignField": "order_id",
			"as":           "suborders",
		}},
		// filter out excluded suborders
		buildFilterExcludedSubordersStage(excludeSuborderIDs),
		// ensure still has suborders after filtering
		{pkg.BKDBMatch: map[string]interface{}{
			"suborders": map[string]interface{}{pkg.BKDBNE: []interface{}{}},
		}},
		// 过滤出已完成子订单
		{pkg.BKDBAddFields: map[string]interface{}{
			"completed_suborders": map[string]interface{}{
				"$filter": map[string]interface{}{
					"input": "$suborders",
					"as":    "suborder",
					"cond":  map[string]interface{}{pkg.BKDBEQ: []interface{}{"$$suborder.stage", types.TicketStageDone}},
				},
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at",
				},
			},
			"last_suborder_time": map[string]interface{}{
				"$max": map[string]interface{}{
					"$max": "$completed_suborders.update_at",
				},
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"duration_hours": map[string]interface{}{
				pkg.BKDBDivide: []interface{}{
					map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$last_suborder_time", "$create_at"}},
					3600000,
				},
			},
		}},
		{pkg.BKDBMatch: map[string]interface{}{
			"duration_hours": map[string]interface{}{
				pkg.BKDBGT: 0,
			},
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": "$year_month",
			"durations": map[string]interface{}{
				"$push": "$duration_hours",
			},
			"count": map[string]interface{}{pkg.BKDBSum: 1},
		}},
		{"$unwind": "$durations"},
		{pkg.BKDBSort: map[string]interface{}{
			"_id":       pkg.BKDBAsc,
			"durations": pkg.BKDBAsc,
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":              "$_id",
			"count":            map[string]interface{}{"$first": "$count"},
			"sorted_durations": map[string]interface{}{"$push": "$durations"},
		}},
		buildPercentileIndicesStage(),
		buildPercentileOverviewProjectStage(),
		{pkg.BKDBSort: map[string]interface{}{"year_month": pkg.BKDBAsc}},
	}
}

// GetPercentileTimeConsumptionCompare implements comparison aggregation per biz across two months
func (op *operation) GetPercentileTimeConsumptionCompare(kt *kit.Kit, param *types.PercentileTimeConsumptionCompareReq) (*types.PercentileTimeConsumptionCompareRst, error) {
	currentStart, currentEnd, err := param.GetCurrentRange()
	if err != nil {
		logs.Errorf("parse current range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	compareStart, compareEnd, err := param.GetCompareRange()
	if err != nil {
		logs.Errorf("parse compare range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// build and run pipelines for current and compare ranges
	current, err := op.aggregatePercentileTimeConsumptionByRange(kt, currentStart, currentEnd)
	if err != nil {
		logs.Errorf("aggregate current range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	compare, err := op.aggregatePercentileTimeConsumptionByRange(kt, compareStart, compareEnd)
	if err != nil {
		logs.Errorf("aggregate compare range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.PercentileTimeConsumptionCompareRst{Current: current, Compare: compare}, nil
}

// aggregatePercentileTimeConsumptionByRange runs the aggregation for a given time range
func (op *operation) aggregatePercentileTimeConsumptionByRange(kt *kit.Kit, start time.Time, end time.Time) ([]types.PercentileTimeConsumptionCompareItem, error) {
	// Get exclude suborder IDs
	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, start, end)
	if err != nil {
		logs.Errorf("get exclude suborder IDs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pipeline := op.buildPercentileTimeConsumptionComparePipeline(start, end, excludeSuborderIDs)

	rst := make([]types.PercentileTimeConsumptionCompareItem, 0)
	if err := model.Operation().ApplyTicket().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate percentile time consumption compare failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}

// buildPercentileTimeConsumptionComparePipeline builds the aggregation pipeline for
// percentile time consumption comparison
func (op *operation) buildPercentileTimeConsumptionComparePipeline(start, end time.Time,
	excludeSuborderIDs []string) []map[string]interface{} {

	match := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: start,
			pkg.BKDBLTE: end,
		},
	}

	return []map[string]interface{}{
		{pkg.BKDBMatch: match},
		{pkg.BKDBLookup: map[string]interface{}{
			"from":         pkg.BKTableNameApplyOrder,
			"localField":   "order_id",
			"foreignField": "order_id",
			"as":           "suborders",
		}},
		// filter out excluded suborders
		buildFilterExcludedSubordersStage(excludeSuborderIDs),
		// ensure still has suborders after filtering
		{pkg.BKDBMatch: map[string]interface{}{"suborders": map[string]interface{}{pkg.BKDBNE: []interface{}{}}}},
		// 过滤出已完成子订单
		{pkg.BKDBAddFields: map[string]interface{}{
			"completed_suborders": map[string]interface{}{
				"$filter": map[string]interface{}{
					"input": "$suborders",
					"as":    "suborder",
					"cond":  map[string]interface{}{pkg.BKDBEQ: []interface{}{"$$suborder.stage", types.TicketStageDone}},
				},
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{"format": "%Y-%m", "date": "$create_at"}},
			"last_suborder_end_time": map[string]interface{}{
				"$max": "$completed_suborders.update_at",
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"duration_hours": map[string]interface{}{
				pkg.BKDBDivide: []interface{}{
					map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$last_suborder_end_time", "$create_at"}},
					3600000,
				},
			},
		}},
		{pkg.BKDBMatch: map[string]interface{}{"duration_hours": map[string]interface{}{pkg.BKDBGT: 0}}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"bk_biz_id":  "$bk_biz_id",
				"year_month": "$year_month",
			},
			"durations": map[string]interface{}{
				"$push": "$duration_hours",
			},
			"count": map[string]interface{}{pkg.BKDBSum: 1},
		}},
		{"$unwind": "$durations"},
		{pkg.BKDBSort: map[string]interface{}{
			"_id.bk_biz_id":  1,
			"_id.year_month": 1,
			"durations":      1,
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":              "$_id",
			"count":            map[string]interface{}{"$first": "$count"},
			"sorted_durations": map[string]interface{}{"$push": "$durations"},
		}},
		buildPercentileIndicesStage(),
		buildPercentileCompareProjectStage(),
		bizOrderMap,
	}
}

// buildPercentileIndicesStage builds a $addFields stage to calculate p90, p95, p99 indices
func buildPercentileIndicesStage() map[string]interface{} {
	return map[string]interface{}{
		pkg.BKDBAddFields: map[string]interface{}{
			"p90_index": map[string]interface{}{
				"$floor": map[string]interface{}{
					"$multiply": []interface{}{
						map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$count", 1}},
						0.90,
					},
				},
			},
			"p95_index": map[string]interface{}{
				"$floor": map[string]interface{}{
					"$multiply": []interface{}{
						map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$count", 1}},
						0.95,
					},
				},
			},
			"p99_index": map[string]interface{}{
				"$floor": map[string]interface{}{
					"$multiply": []interface{}{
						map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$count", 1}},
						0.99,
					},
				},
			},
		},
	}
}

// buildPercentileHoursFields builds the p90/p95/p99 hours fields from sorted durations
func buildPercentileHoursFields() map[string]interface{} {
	return map[string]interface{}{
		"p90_hours": map[string]interface{}{
			pkg.BKDBRound: []interface{}{
				map[string]interface{}{"$arrayElemAt": []interface{}{"$sorted_durations", "$p90_index"}},
				2,
			},
		},
		"p95_hours": map[string]interface{}{
			pkg.BKDBRound: []interface{}{
				map[string]interface{}{"$arrayElemAt": []interface{}{"$sorted_durations", "$p95_index"}},
				2,
			},
		},
		"p99_hours": map[string]interface{}{
			pkg.BKDBRound: []interface{}{
				map[string]interface{}{"$arrayElemAt": []interface{}{"$sorted_durations", "$p99_index"}},
				2,
			},
		},
	}
}

// buildPercentileOverviewProjectStage builds a $project stage for percentile overview
func buildPercentileOverviewProjectStage() map[string]interface{} {
	projectFields := map[string]interface{}{
		"_id":        0,
		"year_month": "$_id",
	}

	// Add percentile hours fields
	for k, v := range buildPercentileHoursFields() {
		projectFields[k] = v
	}

	return map[string]interface{}{
		pkg.BKDBProject: projectFields,
	}
}

// buildPercentileCompareProjectStage builds a $project stage for percentile comparison
func buildPercentileCompareProjectStage() map[string]interface{} {
	projectFields := map[string]interface{}{
		"_id":         0,
		"bk_biz_id":   "$_id.bk_biz_id",
		"year_month":  "$_id.year_month",
		"done_orders": "$count",
	}

	// Add percentile hours fields
	for k, v := range buildPercentileHoursFields() {
		projectFields[k] = v
	}

	return map[string]interface{}{
		pkg.BKDBProject: projectFields,
	}
}
