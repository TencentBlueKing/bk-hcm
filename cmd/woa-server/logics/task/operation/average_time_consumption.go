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

// GetAverageTimeConsumptionOverview aggregates average time consumption by month within a range
func (op *operation) GetAverageTimeConsumptionOverview(kt *kit.Kit, param *types.AverageTimeConsumptionReq) (
	[]types.AverageTimeConsumptionItem, error) {

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

	pipeline := op.buildAverageTimeConsumptionOverviewPipeline(start, end, excludeSuborderIDs)

	rst := make([]types.AverageTimeConsumptionItem, 0)
	if err := model.Operation().ApplyTicket().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate average time consumption overview failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}

// buildAverageTimeConsumptionOverviewPipeline builds the aggregation pipeline for average time consumption overview
func (op *operation) buildAverageTimeConsumptionOverviewPipeline(start, end time.Time, excludeSuborderIDs []string) []map[string]interface{} {
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
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at",
				},
			},
			"completed_suborders": map[string]interface{}{
				"$filter": map[string]interface{}{
					"input": "$suborders",
					"as":    "suborder",
					"cond":  map[string]interface{}{pkg.BKDBEQ: []interface{}{"$$suborder.stage", types.TicketStageDone}},
				},
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"last_suborder_end_time": map[string]interface{}{
				"$max": "$completed_suborders.update_at",
			},
		}},
		{pkg.BKDBMatch: map[string]interface{}{
			"last_suborder_end_time": map[string]interface{}{pkg.BKDBNE: nil},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"duration_hours": map[string]interface{}{
				pkg.BKDBDivide: []interface{}{
					map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$last_suborder_end_time", "$create_at"}},
					3600000,
				},
			},
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":                "$year_month",
			"avg_duration_hours": map[string]interface{}{pkg.BKDBAvg: "$duration_hours"},
		}},
		{pkg.BKDBProject: map[string]interface{}{
			"_id":                0,
			"year_month":         "$_id",
			"avg_duration_hours": map[string]interface{}{pkg.BKDBRound: []interface{}{"$avg_duration_hours", 2}},
		}},
		{pkg.BKDBSort: map[string]interface{}{"year_month": 1}},
	}
}

// GetAverageTimeConsumptionCompare aggregates average time consumption compare by biz and month
func (op *operation) GetAverageTimeConsumptionCompare(kt *kit.Kit, param *types.AverageTimeConsumptionCompareReq) (*types.AverageTimeConsumptionCompareRst, error) {
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
	current, err := op.aggregateAverageTimeConsumptionByRange(kt, currentStart, currentEnd)
	if err != nil {
		logs.Errorf("aggregate current range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	compare, err := op.aggregateAverageTimeConsumptionByRange(kt, compareStart, compareEnd)
	if err != nil {
		logs.Errorf("aggregate compare range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.AverageTimeConsumptionCompareRst{Current: current, Compare: compare}, nil
}

// aggregateAverageTimeConsumptionByRange runs the aggregation for a given time range
func (op *operation) aggregateAverageTimeConsumptionByRange(kt *kit.Kit, start time.Time, end time.Time) ([]types.AverageTimeConsumptionCompareItem, error) {
	// Get exclude suborder IDs
	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, start, end)
	if err != nil {
		logs.Errorf("get exclude suborder IDs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pipeline := op.buildAverageTimeConsumptionComparePipeline(start, end, excludeSuborderIDs)

	rst := make([]types.AverageTimeConsumptionCompareItem, 0)
	if err := model.Operation().ApplyTicket().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate average time consumption compare failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}

// buildAverageTimeConsumptionComparePipeline builds the aggregation pipeline for average time consumption comparison
func (op *operation) buildAverageTimeConsumptionComparePipeline(start, end time.Time,
	excludeSuborderIDs []string) []map[string]interface{} {

	match := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: start,
			pkg.BKDBLTE: end,
		},
	}
	return []map[string]interface{}{
		{pkg.BKDBMatch: match},
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at",
				},
			},
		}},
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
			"suborders.0": map[string]interface{}{pkg.BKDBExists: true},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"suborder_update_times": "$suborders.update_at",
			"completed_suborders": map[string]interface{}{
				"$filter": map[string]interface{}{
					"input": "$suborders",
					"as":    "suborder",
					"cond":  map[string]interface{}{pkg.BKDBEQ: []interface{}{"$$suborder.stage", types.TicketStageDone}},
				},
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"done_orders": map[string]interface{}{
				"$size": "$completed_suborders",
			},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"last_suborder_end_time": map[string]interface{}{
				"$max": "$completed_suborders.update_at",
			},
		}},
		{pkg.BKDBMatch: map[string]interface{}{
			"last_suborder_end_time": map[string]interface{}{pkg.BKDBNE: nil},
		}},
		{pkg.BKDBAddFields: map[string]interface{}{
			"duration_hours": map[string]interface{}{
				pkg.BKDBDivide: []interface{}{
					map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$last_suborder_end_time", "$create_at"}},
					3600000,
				},
			},
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"bk_biz_id":  "$bk_biz_id",
				"year_month": "$year_month",
			},
			"done_orders":        map[string]interface{}{pkg.BKDBSum: "$done_orders"},
			"avg_duration_hours": map[string]interface{}{pkg.BKDBAvg: "$duration_hours"},
		}},
		{pkg.BKDBProject: map[string]interface{}{
			"_id":                0,
			"bk_biz_id":          "$_id.bk_biz_id",
			"year_month":         "$_id.year_month",
			"done_orders":        1,
			"avg_duration_hours": map[string]interface{}{pkg.BKDBRound: []interface{}{"$avg_duration_hours", 2}},
		}},
		{pkg.BKDBSort: map[string]interface{}{
			"bk_biz_id":  1,
			"year_month": 1,
		}},
	}
}

// buildFilterExcludedSubordersStage builds a $addFields stage to filter out excluded suborders
func buildFilterExcludedSubordersStage(excludeSuborderIDs []string) map[string]interface{} {
	return map[string]interface{}{
		pkg.BKDBAddFields: map[string]interface{}{
			"suborders": map[string]interface{}{
				"$filter": map[string]interface{}{
					"input": "$suborders",
					"as":    "suborder",
					"cond": map[string]interface{}{
						pkg.BKDBNot: []interface{}{
							map[string]interface{}{
								pkg.BKDBIN: []interface{}{"$$suborder.suborder_id", excludeSuborderIDs},
							},
						},
					},
				},
			},
		},
	}
}
