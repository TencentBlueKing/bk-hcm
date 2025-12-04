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

// Package operation define the operation interface
package operation

import (
	"context"
	"fmt"
	"sort"
	"time"

	"hcm/cmd/woa-server/logics/task/statistics"
	model "hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/client"
	"hcm/pkg/condition"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/language"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/util"

	"go.mongodb.org/mongo-driver/bson"
)

// Interface operation interface
type Interface interface {
	// GetApplyStatistics get resource apply operation statistics
	GetApplyStatistics(kt *kit.Kit, param *types.GetApplyStatReq) (*types.GetApplyStatRst, error)
	// GetAverageTimeConsumptionOverview get average time consumption overview
	GetAverageTimeConsumptionOverview(kt *kit.Kit, param *types.AverageTimeConsumptionReq) (
		[]types.AverageTimeConsumptionItem, error)
	// GetAverageTimeConsumptionCompare get average time consumption compare
	GetAverageTimeConsumptionCompare(kt *kit.Kit, param *types.AverageTimeConsumptionCompareReq) (
		*types.AverageTimeConsumptionCompareRst, error)
	// GetOrderTimeCostOverview get order time cost overview
	GetOrderTimeCostOverview(kt *kit.Kit, param *types.OrderTimeCostReq) ([]types.OrderTimeCostItem, error)
	// GetOrderTimeCostCompare get order time cost compare
	GetOrderTimeCostCompare(kt *kit.Kit, param *types.OrderTimeCostCompareReq) (*types.OrderTimeCostCompareRst, error)
	// GetProductionStageTimeCostOverview get production stage time cost overview
	GetProductionStageTimeCostOverview(kt *kit.Kit, param *types.ProductionStageTimeCostReq) (
		[]types.ProductionStageTimeCostItem, error)
	// GetProductionStageTimeCostCompare get production stage time cost compare
	GetProductionStageTimeCostCompare(kt *kit.Kit, param *types.ProductionStageTimeCostCompareReq) (
		*types.ProductionStageTimeCostCompareRst, error)
	// GetPercentileTimeConsumptionOverview get percentile time consumption overview
	GetPercentileTimeConsumptionOverview(kt *kit.Kit, param *types.PercentileTimeConsumptionReq) (
		[]types.PercentileTimeConsumptionItem, error)
	// GetPercentileTimeConsumptionCompare get percentile time consumption compare
	GetPercentileTimeConsumptionCompare(kt *kit.Kit, param *types.PercentileTimeConsumptionCompareReq) (
		*types.PercentileTimeConsumptionCompareRst, error)
	// GetDeliveryRateStatistics get delivery rate statistics
	GetDeliveryRateStatistics(kt *kit.Kit, param *types.DeliveryRateStatisticsReq) (
		[]types.DeliveryRateStatisticsItem, error)
	// GetDeliveryRateDetail get delivery rate detail
	GetDeliveryRateDetail(kt *kit.Kit, param *types.DeliveryRateDetailReq) (*types.DeliveryRateDetailResp, error)
	// GetCompletionRateStatistics get completion rate statistics
	GetCompletionRateStatistics(kt *kit.Kit,
		param *types.GetCompletionRateStatReq) (*types.GetCompletionRateStatRst, error)
	// GetCompletionRateDetail 获取结单率详情统计
	GetCompletionRateDetail(kt *kit.Kit,
		param *types.GetCompletionRateDetailReq) (*types.GetCompletionRateDetailRst, error)
	// GetApplyBizHostsStatistics get apply biz hosts statistics
	GetApplyBizHostsStatistics(kt *kit.Kit, startDate, endDate time.Time) (*types.ApplyBizHostsStatisticsResult, error)
	// GetApplyBizCpuCoresStatistics get apply biz cpu cores statistics
	GetApplyBizCpuCoresStatistics(kt *kit.Kit, startDate, endDate time.Time) (
		*types.ApplyBizCpuCoresStatisticsResult, error)
}

// operation provides operation statistics service
type operation struct {
	lang       language.CCLanguageIf
	statistics statistics.Interface
}

// New create a operation instance
func New(_ context.Context, clientSet *client.ClientSet) (*operation, error) {
	op := &operation{
		lang: language.NewFromCtx(language.EmptyLanguageSetting),
	}

	if clientSet != nil {
		op.statistics = statistics.New(clientSet)
	}

	return op, nil
}

// GetApplyStatistics get resource apply operation statistics
func (op *operation) GetApplyStatistics(kt *kit.Kit, param *types.GetApplyStatReq) (*types.GetApplyStatRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	orderTotalStats, err := op.getOrderStats(filter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply total order statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	succOrderFilter := util.CopyMap(filter, nil, nil)
	succOrderFilter["status"] = types.ApplyStatusDone
	orderSuccStats, err := op.getOrderStats(succOrderFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply success order statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	osTotalStats, err := op.getDeviceStats(filter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply total os statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	manualOrderList, err := op.getManualOrderList(filter)
	if err != nil {
		logs.Errorf("failed to get manual order list, err: %v", err)
		return nil, err
	}
	manualOrderFilter := util.CopyMap(filter, nil, nil)
	manualOrderFilter["suborder_id"] = mapstr.MapStr{
		pkg.BKDBIN: manualOrderList,
	}
	orderManualStats, err := op.getOrderStats(manualOrderFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply manual order statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	succOsFilter := util.CopyMap(filter, nil, nil)
	succOsFilter["is_delivered"] = true
	succOsFilter["deliverer"] = "icr"
	osSuccStats, err := op.getDeviceStats(succOsFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply success os statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := getApplyStatRst(orderTotalStats, orderSuccStats, orderManualStats, osTotalStats, osSuccStats)
	return rst, nil
}

func getApplyStatRst(orderTotalStats map[string]metadata.StringIDCount,
	orderSuccStats map[string]metadata.StringIDCount, orderManualStats map[string]metadata.StringIDCount,
	osTotalStats map[string]metadata.StringIDCount,
	osSuccStats map[string]metadata.StringIDCount) *types.GetApplyStatRst {

	// sort date keys
	dateKeys := make([]string, 0)
	for k := range orderTotalStats {
		dateKeys = append(dateKeys, k)
	}
	sort.Strings(dateKeys)

	rst := new(types.GetApplyStatRst)
	for _, date := range dateKeys {
		orderTotalCnt := uint(0)
		if orderTotalStat, ok := orderTotalStats[date]; ok {
			orderTotalCnt = uint(orderTotalStat.Count)
		}

		orderSuccCnt := uint(0)
		if orderSuccStat, ok := orderSuccStats[date]; ok {
			orderSuccCnt = uint(orderSuccStat.Count)
		}

		orderSuccRate := float64(0)
		if orderTotalCnt > 0 {
			orderSuccRate = float64(orderSuccCnt) / float64(orderTotalCnt)
		}

		orderManualCnt := uint(0)
		if orderManualStat, ok := orderManualStats[date]; ok {
			orderManualCnt = uint(orderManualStat.Count)
		}

		orderManualRate := float64(0)
		if orderManualCnt > 0 {
			orderManualRate = float64(orderManualCnt) / float64(orderTotalCnt)
		}

		osTotalCnt := uint(0)
		if osTotalStat, ok := osTotalStats[date]; ok {
			osTotalCnt = uint(osTotalStat.Count)
		}

		osSuccCnt := uint(0)
		if osSuccStat, ok := osSuccStats[date]; ok {
			osSuccCnt = uint(osSuccStat.Count)
		}

		osSuccRate := float64(0)
		if osTotalCnt > 0 {
			osSuccRate = float64(osSuccCnt) / float64(osTotalCnt)
		}

		applyStat := &types.ApplyStat{
			Date:            date,
			OrderTotal:      orderTotalCnt,
			OrderSucc:       orderSuccCnt,
			OrderSuccRate:   orderSuccRate,
			OrderManual:     orderManualCnt,
			OrderManualRate: orderManualRate,
			OsTotal:         osTotalCnt,
			OsSucc:          osSuccCnt,
			OsSuccRate:      osSuccRate,
		}
		rst.Info = append(rst.Info, applyStat)
	}
	return rst
}

// getOrderStats get resource apply order operation statistics
func (op *operation) getOrderStats(filter map[string]interface{}, dimension types.TimeDimension) (
	map[string]metadata.StringIDCount, error) {

	format := op.getDateFormat(dimension)
	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": format,
					"date":   "$create_at"}},
			"count": map[string]interface{}{pkg.BKDBSum: 1}},
		},
		{pkg.BKDBSort: map[string]interface{}{"_id": 1}},
	}

	aggRst := make([]metadata.StringIDCount, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(context.Background(), pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get resource apply order operation statistics, err: %v", err)
		return nil, err
	}

	mapDateStat := make(map[string]metadata.StringIDCount)
	for _, stat := range aggRst {
		mapDateStat[stat.ID] = stat
	}

	return mapDateStat, nil
}

// getDeviceStats get resource apply delivered device operation statistics
func (op *operation) getDeviceStats(filter map[string]interface{}, dimension types.TimeDimension) (
	map[string]metadata.StringIDCount, error) {

	format := op.getDateFormat(dimension)
	pipeline := []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": format,
					"date":   "$create_at"}},
			"count": map[string]interface{}{pkg.BKDBSum: 1}},
		},
		{pkg.BKDBSort: map[string]interface{}{"_id": 1}},
	}

	aggRst := make([]metadata.StringIDCount, 0)
	if err := model.Operation().DeviceInfo().AggregateAll(context.Background(), pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get resource apply delivered device operation statistics, err: %v", err)
		return nil, err
	}

	mapDateStat := make(map[string]metadata.StringIDCount)
	for _, stat := range aggRst {
		mapDateStat[stat.ID] = stat
	}

	return mapDateStat, nil
}

func (op *operation) getManualOrderList(filter map[string]interface{}) ([]interface{}, error) {
	manualFilter := util.CopyMap(filter, nil, nil)
	manualFilter["deliverer"] = mapstr.MapStr{
		pkg.BKDBNE: "icr",
	}

	orderList, err := model.Operation().DeviceInfo().Distinct(context.Background(), "suborder_id", manualFilter)
	if err != nil {
		return nil, err
	}

	return orderList, nil
}

func (op *operation) getDateFormat(dimension types.TimeDimension) string {
	format := ""
	switch dimension {
	case types.DimensionDay:
		format = "%Y-%m-%d"
	case types.DimensionMonth:
		format = "%Y-%m"
	case types.DimensionYear:
		format = "%Y"
	default:
		// treat dimension as day by default
		format = "%Y-%m-%d"
	}

	return format
}

// parseTimeRange 解析时间范围
func parseTimeRange(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	startTime, err := time.Parse(constant.DateLayout, startTimeStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse start_time: %w", err)
	}

	endTime, err := time.Parse(constant.DateLayout, endTimeStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse end_time: %w", err)
	}

	return startTime, endTime, nil
}

// getExcludeSuborderIDs 获取排除的子单号列表
func (op *operation) getExcludeSuborderIDs(kt *kit.Kit, startTime, endTime time.Time) ([]string, error) {
	if op.statistics == nil {
		return nil, nil
	}

	excludeSuborderIDs, err := op.statistics.ListExcludedSubOrderIDs(kt, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get exclude suborder ids: %w", err)
	}

	return excludeSuborderIDs, nil
}

// addExcludeSuborderFilter 添加排除子单号过滤条件
func addExcludeSuborderFilter(filter map[string]interface{}, excludeSuborderIDs []string) {
	if len(excludeSuborderIDs) == 0 {
		return
	}

	suborderFilter, ok := filter["suborder_id"].(map[string]interface{})
	if !ok || suborderFilter == nil {
		suborderFilter = make(map[string]interface{})
	}
	suborderFilter[pkg.BKDBNIN] = excludeSuborderIDs
	filter["suborder_id"] = suborderFilter
}

// buildCompletionRateStatisticsPipeline 构建结单率统计聚合管道
func buildCompletionRateStatisticsPipeline(filter map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{
		{pkg.BKDBMatch: filter},
		{"$addFields": map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at"}},
			"is_done": map[string]interface{}{
				"$cond": []interface{}{
					map[string]interface{}{"$eq": []interface{}{"$stage", "DONE"}},
					1,
					0,
				},
			},
		}},
		{pkg.BKDBGroup: map[string]interface{}{
			"_id":         "$year_month",
			"total_count": map[string]interface{}{pkg.BKDBSum: 1},
			"done_count":  map[string]interface{}{pkg.BKDBSum: "$is_done"},
		}},
		{pkg.BKDBProject: map[string]interface{}{
			"year_month": "$_id",
			"completion_rate": map[string]interface{}{
				"$round": []interface{}{
					map[string]interface{}{
						"$multiply": []interface{}{
							map[string]interface{}{
								"$divide": []interface{}{
									"$done_count",
									map[string]interface{}{
										"$cond": []interface{}{
											map[string]interface{}{"$eq": []interface{}{"$total_count", 0}},
											1,
											"$total_count",
										},
									},
								},
							},
							100,
						},
					},
					2,
				},
			},
		}},
		{pkg.BKDBSort: map[string]interface{}{"year_month": 1}},
	}
}

// convertCompletionRateStatisticsResult 转换结单率统计结果
func convertCompletionRateStatisticsResult(aggRst []struct {
	YearMonth      string  `bson:"year_month"`
	CompletionRate float64 `bson:"completion_rate"`
}) *types.GetCompletionRateStatRst {
	rst := &types.GetCompletionRateStatRst{
		Details: make([]*types.CompletionRateStat, 0, len(aggRst)),
	}

	for _, stat := range aggRst {
		rst.Details = append(rst.Details, &types.CompletionRateStat{
			YearMonth:      stat.YearMonth,
			CompletionRate: stat.CompletionRate,
		})
	}

	return rst
}

// GetCompletionRateStatistics get completion rate statistics
func (op *operation) GetCompletionRateStatistics(kt *kit.Kit,
	param *types.GetCompletionRateStatReq) (*types.GetCompletionRateStatRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get completion rate statistics, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	startTime, endTime, err := parseTimeRange(param.StartTime, param.EndTime)
	if err != nil {
		logs.Errorf("failed to parse time range, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, startTime, endTime)
	if err != nil {
		logs.Errorf("failed to get exclude suborder ids for completion rate statistics, err: %v, rid: %s",
			err, kt.Rid)
		return nil, err
	}

	addExcludeSuborderFilter(filter, excludeSuborderIDs)

	pipeline := buildCompletionRateStatisticsPipeline(filter)

	aggRst := make([]struct {
		YearMonth      string  `bson:"year_month"`
		CompletionRate float64 `bson:"completion_rate"`
	}, 0)

	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get completion rate statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return convertCompletionRateStatisticsResult(aggRst), nil
}

// buildCompletionRateDetailBaseFilter 构建结单率详情基础过滤条件
func buildCompletionRateDetailBaseFilter(startTime, endTime time.Time, excludeSuborderIDs []string) map[string]interface{} {
	baseFilter := map[string]interface{}{
		"create_at": map[string]interface{}{
			pkg.BKDBGTE: startTime,
			pkg.BKDBLT:  endTime,
		},
		"source": map[string]interface{}{
			pkg.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool,
		},
	}

	if len(excludeSuborderIDs) > 0 {
		baseFilter["suborder_id"] = map[string]interface{}{
			pkg.BKDBNIN: excludeSuborderIDs,
		}
	}

	return baseFilter
}

// buildCompletionRateDetailPipeline 构建结单率详情聚合管道
func buildCompletionRateDetailPipeline(baseFilter map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{
		{pkg.BKDBMatch: baseFilter},
		{
			"$addFields": map[string]interface{}{
				"year_month": map[string]interface{}{
					"$dateToString": map[string]interface{}{
						"format": "%Y-%m",
						"date":   "$create_at",
					},
				},
			},
		},
		{
			"$addFields": map[string]interface{}{
				"is_done": map[string]interface{}{
					"$cond": []interface{}{
						map[string]interface{}{
							"$and": []interface{}{
								map[string]interface{}{"$eq": []interface{}{"$stage", types.TicketStageDone}},
								map[string]interface{}{"$eq": []interface{}{"$status", types.ApplyStatusDone}},
							},
						},
						1,
						0,
					},
				},
			},
		},
		{
			pkg.BKDBGroup: map[string]interface{}{
				"_id": map[string]interface{}{
					"bk_biz_id":  "$bk_biz_id",
					"year_month": "$year_month",
				},
				"total_orders": map[string]interface{}{pkg.BKDBSum: 1},
				"done_orders":  map[string]interface{}{pkg.BKDBSum: "$is_done"},
			},
		},
		{
			"$addFields": map[string]interface{}{
				"completion_rate": map[string]interface{}{
					"$cond": []interface{}{
						map[string]interface{}{"$eq": []interface{}{"$total_orders", 0}},
						0.0,
						map[string]interface{}{
							"$multiply": []interface{}{
								map[string]interface{}{
									"$divide": []interface{}{"$done_orders", "$total_orders"},
								},
								100,
							},
						},
					},
				},
			},
		},
		{
			pkg.BKDBProject: map[string]interface{}{
				"_id":          0,
				"bk_biz_id":    "$_id.bk_biz_id",
				"year_month":   "$_id.year_month",
				"total_orders": "$total_orders",
				"done_orders":  "$done_orders",
				"completion_rate": map[string]interface{}{
					"$round": []interface{}{"$completion_rate", 2},
				},
			},
		},
		{
			pkg.BKDBSort: map[string]interface{}{
				"completion_rate": -1,
				"bk_biz_id":       1,
				"year_month":      1,
			},
		},
	}
}

// GetCompletionRateDetail 获取结单率详情统计
func (op *operation) GetCompletionRateDetail(kt *kit.Kit,
	param *types.GetCompletionRateDetailReq) (*types.GetCompletionRateDetailRst, error) {
	startTime, endTime, err := parseTimeRange(param.StartTime, param.EndTime)
	if err != nil {
		logs.Errorf("failed to parse time range, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 结束时间需要加1天，因为查询条件是 $lt（小于）不包含当天
	endTime = endTime.AddDate(0, 0, 1)

	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, startTime, endTime)
	if err != nil {
		logs.Errorf("failed to get exclude suborder ids for completion rate detail, err: %v, rid: %s",
			err, kt.Rid)
		return nil, err
	}

	baseFilter := buildCompletionRateDetailBaseFilter(startTime, endTime, excludeSuborderIDs)
	pipeline := buildCompletionRateDetailPipeline(baseFilter)

	aggRst := make([]*types.CompletionRateDetailItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &aggRst); err != nil {
		logs.Errorf("failed to get completion rate detail statistics, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.GetCompletionRateDetailRst{
		Details: aggRst,
	}, nil
}

// GetApplyBizHostsStatistics 按日期范围统计业务维度的申请主机数据
// startDate: 开始日期字符串，格式：2025-11-01
// endDate: 结束日期字符串，格式：2025-11-30
func (op *operation) GetApplyBizHostsStatistics(kt *kit.Kit, startDate, endDate time.Time) (
	*types.ApplyBizHostsStatisticsResult, error) {

	// 构建基础过滤条件
	baseFilter := bson.M{
		"create_at": bson.M{condition.BKDBGTE: startDate, condition.BKDBLTE: endDate},
		"stage":     types.TicketStageDone,
		"status":    types.ApplyStatusDone,
		"source":    bson.M{condition.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool},
	}

	// 获取需要排除的suborder_id列表（例如：手动处理的订单）
	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, startDate, endDate)
	if err != nil {
		logs.Errorf("failed to get exclude suborder ids, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 如果有需要排除的订单，添加到过滤条件中
	if len(excludeSuborderIDs) > 0 {
		logs.Infof("query biz host statistics exclude [%d] suborder_ids, rid: %s", len(excludeSuborderIDs), kt.Rid)
		baseFilter["suborder_id"] = bson.M{condition.BKDBNIN: excludeSuborderIDs}
	}

	// 构建聚合管道
	pipeline := bson.A{
		// 第一步：过滤时间范围 + 排除特定订单
		bson.M{pkg.BKDBMatch: baseFilter},
		// 第二步：按业务ID分组，统计每个业务的申请成功的主机总数
		bson.M{pkg.BKDBGroup: bson.M{
			"_id":         "$bk_biz_id",
			"order_count": bson.M{"$sum": 1},              // 申请单数量
			"host_count":  bson.M{"$sum": "$success_num"}, // 成功交付的主机数
		},
		},
		// 第三步：按成功交付主机总数降序排序
		bson.M{pkg.BKDBSort: bson.M{"host_count": -1}},
		// 第四步：格式化输出字段
		bson.M{pkg.BKDBProject: bson.M{"_id": 0, "bk_biz_id": "$_id", "order_count": 1, "host_count": 1}},
	}

	var result []types.ApplyBizHostsStatisticsItem
	err = model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &result)
	if err != nil {
		logs.Errorf("failed to aggregate apply biz hosts statistics by date range, startDate: %s, endDate: %s, "+
			"err: %v, rid: %s", startDate, endDate, err, kt.Rid)
		return nil, err
	}

	return &types.ApplyBizHostsStatisticsResult{Details: result}, nil
}

// GetApplyBizCpuCoresStatistics 按日期范围统计业务维度的申请核心数数据
// startDate: 开始日期字符串，格式：2025-11-01
// endDate: 结束日期字符串，格式：2025-11-30
func (op *operation) GetApplyBizCpuCoresStatistics(kt *kit.Kit, startDate, endDate time.Time) (
	*types.ApplyBizCpuCoresStatisticsResult, error) {

	// 构建基础过滤条件
	baseFilter := bson.M{
		"create_at": bson.M{condition.BKDBGTE: startDate, condition.BKDBLTE: endDate},
		"stage":     types.TicketStageDone,
		"status":    types.ApplyStatusDone,
		"source":    bson.M{condition.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool},
	}

	// 获取需要排除的suborder_id列表（例如：手动处理的订单）
	excludeSuborderIDs, err := op.getExcludeSuborderIDs(kt, startDate, endDate)
	if err != nil {
		logs.Errorf("failed to get exclude suborder ids, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 如果有需要排除的订单，添加到过滤条件中
	if len(excludeSuborderIDs) > 0 {
		logs.Infof("query biz cpu cores statistics exclude [%d] suborder_ids, rid: %s", len(excludeSuborderIDs), kt.Rid)
		baseFilter["suborder_id"] = bson.M{condition.BKDBNIN: excludeSuborderIDs}
	}

	// 构建聚合管道
	pipeline := bson.A{
		// 第一步：过滤时间范围 + 排除特定订单
		bson.M{pkg.BKDBMatch: baseFilter},
		// 第二步：按业务ID分组，统计每个业务的申请成功的主机总数
		bson.M{pkg.BKDBGroup: bson.M{
			"_id":                  "$bk_biz_id",
			"order_count":          bson.M{"$sum": 1},                 // 申请单数量
			"delivered_core_count": bson.M{"$sum": "$delivered_core"}, // 成功交付的主机数
		},
		},
		// 第三步：按成功交付主机总数降序排序
		bson.M{pkg.BKDBSort: bson.M{"delivered_core_count": -1}},
		// 第四步：格式化输出字段
		bson.M{pkg.BKDBProject: bson.M{"_id": 0, "bk_biz_id": "$_id", "order_count": 1, "delivered_core_count": 1}},
	}

	var result []types.ApplyBizCpuCoresStatisticsItem
	err = model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &result)
	if err != nil {
		logs.Errorf("failed to aggregate apply biz cpu core statistics by date range, startDate: %s, endDate: %s, "+
			"err: %v, rid: %s", startDate, endDate, err, kt.Rid)
		return nil, err
	}

	return &types.ApplyBizCpuCoresStatisticsResult{Details: result}, nil
}
