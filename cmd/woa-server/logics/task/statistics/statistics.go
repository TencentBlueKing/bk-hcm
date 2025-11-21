/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package statistics package
package statistics

import (
	"fmt"
	"strings"
	"time"

	taskModel "hcm/cmd/woa-server/model/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	tableapplystat "hcm/pkg/dal/table/cvm-apply-order-statistics-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// Interface statistics interface
type Interface interface {
	// ListExcludedSubOrderIDs 根据查询时间范围，将配置表内容统一转换为需要排除的主机申请子订单号列表
	ListExcludedSubOrderIDs(kt *kit.Kit, queryStart, queryEnd time.Time) ([]string, error)
}

// New 创建统计能力实例
func New(clientSet *client.ClientSet) Interface {
	return &statistics{
		client: clientSet,
	}
}

type statistics struct {
	client *client.ClientSet
}

// ListExcludedSubOrderIDs 根据查询时间范围，将配置表内容统一转换为需要排除的主机申请子订单号列表
func (s *statistics) ListExcludedSubOrderIDs(kt *kit.Kit, queryStart, queryEnd time.Time) ([]string, error) {
	if err := validateQueryRange(queryStart, queryEnd); err != nil {
		return nil, err
	}
	// 从配置表中加载所有相关月份的配置
	monthSet, monthSlice := buildQueryMonths(queryStart, queryEnd)

	configs, err := s.loadConfigs(kt, monthSlice)
	if err != nil {
		return nil, fmt.Errorf("list statistics config failed: %w", err)
	}

	explicitIDs, timeRanges := s.extractConfigInfo(kt, configs, monthSet, queryStart, queryEnd)

	result := make([]string, 0, len(explicitIDs))
	result = append(result, explicitIDs...)

	if len(timeRanges) > 0 {
		// 从申请单中反查所有符合时间范围的子订单号
		queryIDs, err := s.fetchSubOrderIDsFromOrders(kt, timeRanges)
		if err != nil {
			return nil, err
		}
		result = append(result, queryIDs...)
	}

	return slice.Unique(result), nil
}

type timeRangeConfig struct {
	start time.Time
	end   time.Time
}

// fetchSubOrderIDsFromOrders 根据时间段配置，到申请单集合按时间反查子单号
func (s *statistics) fetchSubOrderIDsFromOrders(kt *kit.Kit, ranges []timeRangeConfig) ([]string, error) {
	if len(ranges) == 0 {
		return nil, nil
	}

	orConditions := make([]map[string]interface{}, 0, len(ranges))
	for _, tr := range ranges {
		if tr.start.After(tr.end) {
			continue
		}
		cond := map[string]interface{}{
			"create_at": map[string]interface{}{
				pkg.BKDBGTE: tr.start,
				pkg.BKDBLTE: tr.end,
			},
		}
		orConditions = append(orConditions, cond)
	}

	if len(orConditions) == 0 {
		return nil, fmt.Errorf("all time ranges are invalid")
	}

	filters := map[string]interface{}{
		pkg.BKDBOR: orConditions,
	}

	// 先查询总数
	count, err := taskModel.Operation().ApplyOrder().CountApplyOrder(kt.Ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("count apply order failed: %w", err)
	}

	if count == 0 {
		return nil, nil
	}

	// 使用分页循环查询，参考 MySQL 标准分页模式
	allIDs := make([]string, 0)
	for offset := uint64(0); offset < count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		page := metadata.BasePage{
			Start: int(offset),
			Limit: int(core.DefaultMaxPageLimit),
		}

		orders, err := taskModel.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filters)
		if err != nil {
			return nil, fmt.Errorf("find apply order failed: %w", err)
		}

		for _, order := range orders {
			if order == nil {
				continue
			}
			if order.SubOrderId == "" {
				continue
			}
			allIDs = append(allIDs, order.SubOrderId)
		}
	}

	return allIDs, nil
}

// splitSubOrderIDs 从逗号分隔的子单号字符串中提取有效子单号
func splitSubOrderIDs(ids string) []string {
	parts := strings.Split(ids, ",")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		results = append(results, part)
	}
	return slice.Unique(results)
}

// parseConfigTime 解析配置中的时间字符串
func parseConfigTime(value string) (time.Time, bool, error) {
	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{constant.DateTimeLayout, false},
		{constant.DateLayout, true},
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout.layout, value, time.Local)
		if err == nil {
			return t, layout.dateOnly, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("unsupported time format: %s", value)
}

// buildQueryMonths 构建查询范围内的月份列表，格式为"YYYY-MM"
func buildQueryMonths(start, end time.Time) (map[string]bool, []string) {
	set := make(map[string]bool)
	months := make([]string, 0)

	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	last := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	for !current.After(last) {
		key := current.Format(constant.YearMonthLayout)
		set[key] = true
		months = append(months, key)
		current = current.AddDate(0, 1, 0)
	}
	return set, months
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// validateQueryRange 验证查询时间范围是否有效
func validateQueryRange(queryStart, queryEnd time.Time) error {
	if queryStart.IsZero() || queryEnd.IsZero() {
		return fmt.Errorf("query start/end time can not be empty")
	}
	if queryStart.After(queryEnd) {
		return fmt.Errorf("query start time can not be after end time")
	}
	return nil
}

// loadConfigs 从数据库加载所有符合月份条件的配置
func (s *statistics) loadConfigs(kt *kit.Kit, monthSlice []string) ([]*tableapplystat.CvmApplyOrderStatisticsConfigTable, error) {
	if s.client == nil || s.client.DataService() == nil {
		return nil, fmt.Errorf("data service client is not initialized")
	}

	if len(monthSlice) == 0 {
		return nil, fmt.Errorf("query month slice can not be empty")
	}

	filterExpr := tools.ContainersExpression("stat_month", monthSlice)
	page := &core.BasePage{
		Start: 0,
		Limit: core.DefaultMaxPageLimit,
	}

	configs := make([]*tableapplystat.CvmApplyOrderStatisticsConfigTable, 0)

	listReq := &core.ListReq{
		Filter: filterExpr,
		Page:   page,
	}

	for {
		result, err := s.client.DataService().Global.ApplyOrderStatisticsConfig.List(kt, listReq)
		if err != nil {
			return nil, fmt.Errorf("list apply order statistics config from data service failed: %w", err)
		}

		if result == nil || len(result.Details) == 0 {
			break
		}

		for i := range result.Details {
			configs = append(configs, &result.Details[i])
		}

		if uint(len(result.Details)) < core.DefaultMaxPageLimit {
			break
		}

		page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return configs, nil
}

// extractConfigInfo 从配置中提取显式子单号和时间范围配置
func (s *statistics) extractConfigInfo(kt *kit.Kit, configs []*tableapplystat.CvmApplyOrderStatisticsConfigTable,
	monthSet map[string]bool, queryStart, queryEnd time.Time) ([]string, []timeRangeConfig) {

	explicitIDs := make([]string, 0)
	explicitSet := make(map[string]struct{})
	timeRanges := make([]timeRangeConfig, 0)

	for _, cfg := range configs {
		// 检查配置的月份是否在查询范围内
		if cfg == nil || !monthSet[cfg.StatMonth] {
			continue
		}
		if cfg.SubOrderIDs != nil {
			for _, id := range splitSubOrderIDs(*cfg.SubOrderIDs) {
				if _, ok := explicitSet[id]; ok {
					continue
				}
				explicitSet[id] = struct{}{}
				explicitIDs = append(explicitIDs, id)
			}
		}

		rangeCfg, ok := s.buildTimeRange(kt, cfg, queryStart, queryEnd)
		if ok {
			timeRanges = append(timeRanges, rangeCfg)
		}
	}

	return explicitIDs, timeRanges
}

// buildTimeRange 构建配置的时间范围，确保不超出查询范围
func (s *statistics) buildTimeRange(kt *kit.Kit, cfg *tableapplystat.CvmApplyOrderStatisticsConfigTable,
	queryStart, queryEnd time.Time) (timeRangeConfig, bool) {

	if cfg == nil {
		return timeRangeConfig{}, false
	}
	if cfg.StartAt == nil || cfg.EndAt == nil {
		return timeRangeConfig{}, false
	}
	if strings.TrimSpace(*cfg.StartAt) == "" || strings.TrimSpace(*cfg.EndAt) == "" {
		return timeRangeConfig{}, false
	}

	start, startDateOnly, err := parseConfigTime(*cfg.StartAt)
	if err != nil {
		logs.Warnf(
			"parse start_at failed, cfg_id: %s, start_at: %s, err: %v, rid: %s",
			cfg.ID, *cfg.StartAt, err, kt.Rid)
		return timeRangeConfig{}, false
	}
	end, endDateOnly, err := parseConfigTime(*cfg.EndAt)
	if err != nil {
		logs.Warnf(
			"parse end_at failed, cfg_id: %s, end_at: %s, err: %v, rid: %s",
			cfg.ID, *cfg.EndAt, err, kt.Rid)
		return timeRangeConfig{}, false
	}

	if endDateOnly {
		end = end.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}
	if startDateOnly {
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	}

	if end.Before(start) {
		logs.Warnf("invalid config time range, cfg_id: %s, start_at: %s, end_at: %s, rid: %s",
			cfg.ID, *cfg.StartAt, *cfg.EndAt, kt.Rid)
		return timeRangeConfig{}, false
	}

	if end.Before(queryStart) || start.After(queryEnd) {
		return timeRangeConfig{}, false
	}

	return timeRangeConfig{
		start: maxTime(start, queryStart),
		end:   minTime(end, queryEnd),
	}, true
}
