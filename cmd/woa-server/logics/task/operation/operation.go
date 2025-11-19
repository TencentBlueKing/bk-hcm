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
	"sort"
	"time"

	model "hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/language"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/util"
)

// Interface operation interface
type Interface interface {
	// GetApplyStatistics get resource apply operation statistics
	GetApplyStatistics(kit *kit.Kit, param *types.GetApplyStatReq) (*types.GetApplyStatRst, error)
	// GetAverageTimeConsumptionOverview get average time consumption overview
	GetAverageTimeConsumptionOverview(kit *kit.Kit, param *types.AverageTimeConsumptionReq) ([]types.AverageTimeConsumptionItem, error)
	// GetAverageTimeConsumptionCompare get average time consumption compare
	GetAverageTimeConsumptionCompare(kit *kit.Kit, param *types.AverageTimeConsumptionCompareReq) (*types.AverageTimeConsumptionCompareRst, error)
	// GetOrderTimeCostOverview get order time cost overview
	GetOrderTimeCostOverview(kit *kit.Kit, param *types.OrderTimeCostReq) ([]types.OrderTimeCostItem, error)
	// GetOrderTimeCostCompare get order time cost compare
	GetOrderTimeCostCompare(kit *kit.Kit, param *types.OrderTimeCostCompareReq) (*types.OrderTimeCostCompareRst, error)
	// GetProductionStageTimeCostOverview get production stage time cost overview
	GetProductionStageTimeCostOverview(kit *kit.Kit, param *types.ProductionStageTimeCostReq) ([]types.ProductionStageTimeCostItem, error)
	// GetProductionStageTimeCostCompare get production stage time cost compare
	GetProductionStageTimeCostCompare(kit *kit.Kit, param *types.ProductionStageTimeCostCompareReq) (*types.ProductionStageTimeCostCompareRst, error)
	// GetPercentileTimeConsumptionOverview get percentile time consumption overview
	GetPercentileTimeConsumptionOverview(kit *kit.Kit, param *types.PercentileTimeConsumptionReq) ([]types.PercentileTimeConsumptionItem, error)
	// GetPercentileTimeConsumptionCompare get percentile time consumption compare
	GetPercentileTimeConsumptionCompare(kit *kit.Kit, param *types.PercentileTimeConsumptionCompareReq) (*types.PercentileTimeConsumptionCompareRst, error)
	// GetDeliveryRateStatistics get delivery rate statistics
	GetDeliveryRateStatistics(kit *kit.Kit, param *types.DeliveryRateStatisticsReq) ([]types.DeliveryRateStatisticsItem, error)
	// GetDeliveryRateDetail get delivery rate detail
	GetDeliveryRateDetail(kit *kit.Kit, param *types.DeliveryRateDetailReq) (*types.DeliveryRateDetailResp, error)
}

// operation provides operation statistics service
type operation struct {
	lang language.CCLanguageIf
}

// New create a operation instance
func New(_ context.Context) (*operation, error) {

	operation := &operation{
		lang: language.NewFromCtx(language.EmptyLanguageSetting),
	}

	return operation, nil
}

// GetApplyStatistics get resource apply operation statistics
func (op *operation) GetApplyStatistics(kit *kit.Kit, param *types.GetApplyStatReq) (*types.GetApplyStatRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	orderTotalStats, err := op.getOrderStats(filter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply total order statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	succOrderFilter := util.CopyMap(filter, nil, nil)
	succOrderFilter["status"] = types.ApplyStatusDone
	orderSuccStats, err := op.getOrderStats(succOrderFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply success order statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	osTotalStats, err := op.getDeviceStats(filter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply total os statistics, err: %v, rid: %s", err, kit.Rid)
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
		logs.Errorf("failed to get resource apply manual order statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	succOsFilter := util.CopyMap(filter, nil, nil)
	succOsFilter["is_delivered"] = true
	succOsFilter["deliverer"] = "icr"
	osSuccStats, err := op.getDeviceStats(succOsFilter, param.Dimension)
	if err != nil {
		logs.Errorf("failed to get resource apply success os statistics, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

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

	return rst, nil
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

func (op *operation) getExcludeSuborderIDs(kt *kit.Kit, startDate, endDate time.Time) ([]string, error) {

	// TODO 从配置表中，按指定的查询日期，读取需要排除的主机申请单号列表

	// excludeList, err := op.getExcludeListFromConfig(ctx)
	// return excludeList, err

	return []string{}, nil
}
