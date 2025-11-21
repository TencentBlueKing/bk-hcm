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

package config

import (
	"fmt"
	"sort"
	"strings"
	"time"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	"hcm/pkg/dal/dao"
	daoorm "hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	tableapplystat "hcm/pkg/dal/table/cvm-apply-order-statistics-config"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/times"

	"github.com/jmoiron/sqlx"
)

// ApplyOrderStatisticsIf provides management interface for operations of apply order statistics config
type ApplyOrderStatisticsIf interface {
	// CreateConfig creates apply order statistics config
	CreateConfig(kt *kit.Kit,
		input *types.CreateApplyOrderStatisticsConfigParam) (*types.CreateApplyOrderStatisticsConfigResult, error)
	// UpdateConfig updates apply order statistics config with full coverage
	UpdateConfig(kt *kit.Kit, input *types.UpdateApplyOrderStatisticsConfigParam) error
	// ListConfig lists apply order statistics config by stat_month
	ListConfig(kt *kit.Kit,
		input *types.ListApplyOrderStatisticsConfigParam) (*types.ListApplyOrderStatisticsConfigResult, error)
	// ListYearMonths lists all stat_months from config table
	ListYearMonths(kt *kit.Kit) (*types.ListApplyOrderStatisticsYearMonthsResult, error)
}

// NewApplyOrderStatisticsOp creates an apply order statistics interface
func NewApplyOrderStatisticsOp(daoSet dao.Set) ApplyOrderStatisticsIf {
	return &applyOrderStatistics{
		dao: daoSet,
	}
}

type applyOrderStatistics struct {
	dao dao.Set
}

// CreateConfig creates apply order statistics config
func (a *applyOrderStatistics) CreateConfig(kt *kit.Kit,
	input *types.CreateApplyOrderStatisticsConfigParam) (*types.CreateApplyOrderStatisticsConfigResult, error) {

	// 验证参数
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("validate input failed: %w", err)
	}

	now := time.Now()
	models := make([]*tableapplystat.CvmApplyOrderStatisticsConfigTable, 0, len(input.Configs))
	for _, cfg := range input.Configs {
		model, err := a.composeModel(composeModelOption{
			User:        kt.User,
			StatMonth:   input.StatMonth,
			BkBizID:     cfg.BkBizID,
			Memo:        cfg.Memo,
			SubOrderIDs: cfg.SubOrderIDs,
			StartAt:     cfg.StartAt,
			EndAt:       cfg.EndAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			return nil, err
		}

		modelCopy := model
		models = append(models, &modelCopy)
	}

	var ids []string
	result, err := a.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, _ *daoorm.TxnOption) (interface{}, error) {
		createdIDs, err := a.dao.CvmApplyOrderStatisticsConfig().CreateWithTx(kt, txn, models)
		if err != nil {
			return nil, err
		}
		return createdIDs, nil
	})
	if err != nil {
		logs.Errorf("create apply order statistics config failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create apply order statistics config failed: %w", err)
	}

	ids, ok := result.([]string)
	if !ok {
		logs.Errorf("create apply order statistics config, invalid result type, rid: %s", kt.Rid)
		return nil, fmt.Errorf("create apply order statistics config, invalid result type")
	}

	if len(ids) != len(models) {
		logs.Errorf("create apply order statistics config, id count mismatch, expect: %d, actual: %d, rid: %s",
			len(models), len(ids), kt.Rid)
		return nil, fmt.Errorf("create apply order statistics config, id count mismatch")
	}

	return &types.CreateApplyOrderStatisticsConfigResult{
		IDs: ids,
	}, nil
}

// UpdateConfig updates apply order statistics config with full coverage
func (a *applyOrderStatistics) UpdateConfig(kt *kit.Kit,
	input *types.UpdateApplyOrderStatisticsConfigParam) error {

	// 验证参数
	if err := input.Validate(); err != nil {
		return fmt.Errorf("validate input failed: %w", err)
	}

	_, err := a.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, _ *daoorm.TxnOption) (interface{}, error) {
		deleteExpr := tools.EqualExpression("stat_month", input.StatMonth)
		if err := a.dao.CvmApplyOrderStatisticsConfig().DeleteWithTx(kt, txn, deleteExpr); err != nil {
			logs.Errorf("delete apply order statistics config failed, stat_month: %s, err: %v, rid: %s",
				input.StatMonth, err, kt.Rid)
			return nil, fmt.Errorf("delete config failed: %w", err)
		}

		if len(input.Configs) == 0 {
			return nil, nil
		}

		models := make([]*tableapplystat.CvmApplyOrderStatisticsConfigTable, 0, len(input.Configs))
		now := time.Now()
		for _, cfg := range input.Configs {
			model, err := a.composeModel(composeModelOption{
				User:        kt.User,
				StatMonth:   input.StatMonth,
				BkBizID:     cfg.BkBizID,
				Memo:        cfg.Memo,
				SubOrderIDs: cfg.SubOrderIDs,
				StartAt:     cfg.StartAt,
				EndAt:       cfg.EndAt,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			if err != nil {
				return nil, err
			}
			modelCopy := model
			models = append(models, &modelCopy)
		}

		if _, err := a.dao.CvmApplyOrderStatisticsConfig().CreateWithTx(kt, txn, models); err != nil {
			logs.Errorf("create apply order statistics config failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("create config failed: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		logs.Errorf("update apply order statistics config failed, stat_month: %s, err: %v, rid: %s",
			input.StatMonth, err, kt.Rid)
		return fmt.Errorf("update apply order statistics config failed: %w", err)
	}

	return nil
}

// ListConfig lists apply order statistics config by stat_month
func (a *applyOrderStatistics) ListConfig(kt *kit.Kit,
	input *types.ListApplyOrderStatisticsConfigParam) (*types.ListApplyOrderStatisticsConfigResult, error) {

	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("validate input failed: %w", err)
	}

	filterExpr := tools.EqualExpression("stat_month", input.StatMonth)

	page := core.NewDefaultBasePage()
	opt := &daotypes.ListOption{
		Filter: filterExpr,
		Page:   page,
	}

	result, err := a.dao.CvmApplyOrderStatisticsConfig().List(kt, opt)
	if err != nil {
		logs.Errorf("list apply order statistics config failed, stat_month: %s, err: %v, rid: %s",
			input.StatMonth, err, kt.Rid)
		return nil, fmt.Errorf("list apply order statistics config failed: %w", err)
	}

	return &types.ListApplyOrderStatisticsConfigResult{
		Details: convertDetails(result),
	}, nil
}

// ListYearMonths lists all stat_months from config table
func (a *applyOrderStatistics) ListYearMonths(kt *kit.Kit) (*types.ListApplyOrderStatisticsYearMonthsResult, error) {

	page := core.NewDefaultBasePage()
	opt := &daotypes.ListOption{
		Filter: tools.AllExpression(),
		Page:   page,

		Fields: []string{"stat_month"},
	}

	result, err := a.dao.CvmApplyOrderStatisticsConfig().List(kt, opt)
	if err != nil {
		logs.Errorf("list apply order statistics config failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list apply order statistics config failed: %w", err)
	}

	yearMonths := uniqueYearMonths(result)
	sort.Slice(yearMonths, func(i, j int) bool {
		return yearMonths[i] > yearMonths[j]
	})

	details := make([]types.ApplyOrderStatisticsYearMonth, 0, len(yearMonths))
	for _, month := range yearMonths {
		details = append(details, types.ApplyOrderStatisticsYearMonth{StatMonth: month})
	}

	return &types.ListApplyOrderStatisticsYearMonthsResult{
		Count:   int64(len(details)),
		Details: details,
	}, nil
}

type composeModelOption struct {
	User        string
	StatMonth   string
	BkBizID     int64
	Memo        string
	SubOrderIDs []string
	StartAt     string
	EndAt       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// composeModel composes apply order statistics config model
func (a *applyOrderStatistics) composeModel(opt composeModelOption) (tableapplystat.CvmApplyOrderStatisticsConfigTable, error) {
	subOrderIDs := normalizeSubOrderIDs(opt.SubOrderIDs)
	subOrderIDsStr := strings.Join(subOrderIDs, ",")

	startAt := strings.TrimSpace(opt.StartAt)
	endAt := strings.TrimSpace(opt.EndAt)

	// extension 字段用于存储扩展数据，目前没有扩展需求，存储空对象
	ext, err := buildExtension(nil)
	if err != nil {
		return tableapplystat.CvmApplyOrderStatisticsConfigTable{}, fmt.Errorf("build extension failed: %w", err)
	}

	return tableapplystat.CvmApplyOrderStatisticsConfigTable{
		StatMonth:   opt.StatMonth,
		BkBizID:     opt.BkBizID,
		SubOrderIDs: &subOrderIDsStr,
		StartAt:     &startAt,
		EndAt:       &endAt,
		Memo:        opt.Memo,
		Extension:   &ext,
		Creator:     opt.User,
		Reviser:     opt.User,
		CreatedAt:   tabletypes.Time(times.ConvStdTimeFormat(opt.CreatedAt)),
		UpdatedAt:   tabletypes.Time(times.ConvStdTimeFormat(opt.UpdatedAt)),
	}, nil
}

// convertDetails 将返回的配置表记录转换为对外返回的配置明细
func convertDetails(result *daotypes.ListResult[tableapplystat.CvmApplyOrderStatisticsConfigTable],
) []types.ApplyOrderStatisticsConfigDetail {
	if result == nil || len(result.Details) == 0 {
		return []types.ApplyOrderStatisticsConfigDetail{}
	}

	details := make([]types.ApplyOrderStatisticsConfigDetail, 0, len(result.Details))
	for _, cfg := range result.Details {
		var subOrderIDs []string
		if cfg.SubOrderIDs != nil {
			subOrderIDs = normalizeSubOrderIDs(strings.Split(*cfg.SubOrderIDs, ","))
		}
		detail := types.ApplyOrderStatisticsConfigDetail{
			ID:          cfg.ID,
			StatMonth:   cfg.StatMonth,
			BkBizID:     cfg.BkBizID,
			SubOrderIDs: subOrderIDs,
			Memo:        cfg.Memo,
		}

		if cfg.StartAt != nil && *cfg.StartAt != "" {
			detail.StartAt = *cfg.StartAt
		}
		if cfg.EndAt != nil && *cfg.EndAt != "" {
			detail.EndAt = *cfg.EndAt
		}

		details = append(details, detail)
	}

	return details
}

// uniqueYearMonths 从查询结果中提取唯一的列表，用于去重和后续展示
func uniqueYearMonths(result *daotypes.ListResult[tableapplystat.CvmApplyOrderStatisticsConfigTable]) []string {
	if result == nil || len(result.Details) == 0 {
		return []string{}
	}

	months := make(map[string]struct{})
	for _, cfg := range result.Details {
		month := strings.TrimSpace(cfg.StatMonth)
		if month == "" {
			continue
		}
		months[month] = struct{}{}
	}

	list := make([]string, 0, len(months))
	for month := range months {
		list = append(list, month)
	}
	return list
}

// normalizeSubOrderIDs 清洗传入的子单号数组，去掉空白和空字符串
func normalizeSubOrderIDs(ids []string) []string {
	if len(ids) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		result = append(result, id)
	}
	return result
}

// buildExtension 构建 extension 扩展字段
func buildExtension(extraData map[string]interface{}) (tabletypes.JsonField, error) {
	if extraData == nil || len(extraData) == 0 {

		// 目前没有扩展数据，返回空对象以满足 JSON NOT NULL 约束
		return tabletypes.NewJsonField(map[string]interface{}{})
	}

	return tabletypes.NewJsonField(extraData)
}
