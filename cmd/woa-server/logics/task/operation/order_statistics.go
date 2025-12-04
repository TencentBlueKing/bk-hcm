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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

const millisecondsPerHour = int64(time.Hour / time.Millisecond)

var (
	bizOrderMap = map[string]interface{}{pkg.BKDBSort: map[string]interface{}{
		"bk_biz_id":  pkg.BKDBAsc,
		"year_month": pkg.BKDBAsc,
	}}
)

// GetOrderTimeCostOverview aggregates order time cost by month within a range
func (op *operation) GetOrderTimeCostOverview(kt *kit.Kit, param *types.OrderTimeCostReq) ([]types.OrderTimeCostItem, error) {
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

	match := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: start,
			pkg.BKDBLTE: end,
		},
		"stage":  types.TicketStageDone,
		"status": types.ApplyStatusDone,
		"source": map[string]interface{}{
			pkg.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool,
		},
	}

	// Exclude suborder IDs if any
	if len(excludeSuborderIDs) > 0 {
		match["suborder_id"] = map[string]interface{}{
			pkg.BKDBNIN: excludeSuborderIDs,
		}
	}

	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: match},
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at",
				},
			},
			"duration_hours": map[string]interface{}{
				pkg.BKDBDivide: []interface{}{
					map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$update_at", "$create_at"}},
					millisecondsPerHour,
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

	rst := make([]types.OrderTimeCostItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate order time cost overview failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}

// GetOrderTimeCostCompare implements comparison aggregation per biz across two months
func (op *operation) GetOrderTimeCostCompare(kt *kit.Kit, param *types.OrderTimeCostCompareReq) (*types.OrderTimeCostCompareRst, error) {
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
	current, err := op.aggregateOrderTimeCostByRange(kt, currentStart, currentEnd)
	if err != nil {
		logs.Errorf("aggregate current range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	compare, err := op.aggregateOrderTimeCostByRange(kt, compareStart, compareEnd)
	if err != nil {
		logs.Errorf("aggregate compare range failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.OrderTimeCostCompareRst{Current: current, Compare: compare}, nil
}

// aggregateOrderTimeCostByRange runs the aggregation for a given time range
func (op *operation) aggregateOrderTimeCostByRange(kt *kit.Kit, start time.Time, end time.Time) ([]types.OrderTimeCostCompareItem, error) {
	// Get exclude suborder IDs
	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, start, end)
	if err != nil {
		logs.Errorf("get exclude suborder IDs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	match := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: start,
			pkg.BKDBLTE: end,
		},
		// only completed orders
		"stage":  types.TicketStageDone,
		"status": types.ApplyStatusDone,
		"source": map[string]interface{}{
			pkg.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool,
		},
	}

	// Exclude suborder IDs if any
	if len(excludeSuborderIDs) > 0 {
		match["suborder_id"] = map[string]interface{}{
			pkg.BKDBNIN: excludeSuborderIDs,
		}
	}

	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: match},
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at",
				},
			},
			"duration_hours": map[string]interface{}{
				pkg.BKDBDivide: []interface{}{
					map[string]interface{}{pkg.BKDBSubtract: []interface{}{"$update_at", "$create_at"}},
					millisecondsPerHour,
				},
			},
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"bk_biz_id":  "$bk_biz_id",
				"year_month": "$year_month",
			},
			"done_orders":        map[string]interface{}{pkg.BKDBSum: 1},
			"avg_duration_hours": map[string]interface{}{pkg.BKDBAvg: "$duration_hours"},
		}},
		{pkg.BKDBProject: map[string]interface{}{
			"_id":                0,
			"bk_biz_id":          "$_id.bk_biz_id",
			"year_month":         "$_id.year_month",
			"done_orders":        1,
			"avg_duration_hours": map[string]interface{}{pkg.BKDBRound: []interface{}{"$avg_duration_hours", 2}},
		}},
		bizOrderMap,
	}

	rst := make([]types.OrderTimeCostCompareItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate order time cost compare failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}
