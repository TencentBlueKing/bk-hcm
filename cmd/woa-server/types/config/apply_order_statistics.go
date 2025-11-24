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
	"errors"
	"fmt"
	"time"

	"hcm/pkg/criteria/constant"
)

// CreateApplyOrderStatisticsConfigParam 创建申请单统计配置请求参数
type CreateApplyOrderStatisticsConfigParam struct {
	StatMonth string                           `json:"stat_month" validate:"required"`
	Configs   []ApplyOrderStatisticsConfigItem `json:"configs" validate:"required"`
}

// ApplyOrderStatisticsConfigItem 申请单统计配置项
type ApplyOrderStatisticsConfigItem struct {
	BkBizID     int64    `json:"bk_biz_id" validate:"required"`
	SubOrderIDs []string `json:"sub_order_ids"`
	StartAt     string   `json:"start_at"`
	EndAt       string   `json:"end_at"`
	Memo        string   `json:"memo" validate:"required,max=255"`
}

// Validate 验证创建申请单统计配置请求参数
func (c *CreateApplyOrderStatisticsConfigParam) Validate() error {
	// 验证年月格式 YYYY-MM
	_, err := time.Parse(constant.YearMonthLayout, c.StatMonth)
	if err != nil {
		return fmt.Errorf("stat_month format must be YYYY-MM, invalid: %w", err)
	}

	if len(c.Configs) == 0 {
		return errors.New("configs is required and must not be empty")
	}

	for idx := range c.Configs {
		if err := c.Configs[idx].Validate(); err != nil {
			return fmt.Errorf("configs[%d] validation failed: %w", idx, err)
		}
	}

	return nil
}

// CreateApplyOrderStatisticsConfigResult 创建申请单统计配置响应结果
type CreateApplyOrderStatisticsConfigResult struct {
	IDs []string `json:"ids"`
}

// UpdateApplyOrderStatisticsConfigParam 更新申请单统计配置请求参数
type UpdateApplyOrderStatisticsConfigParam struct {
	StatMonth string                           `json:"stat_month" validate:"required"`
	Configs   []ApplyOrderStatisticsConfigItem `json:"configs" validate:"required"`
}

// Validate 验证更新申请单统计配置请求参数
func (u *UpdateApplyOrderStatisticsConfigParam) Validate() error {
	// 验证年月格式 YYYY-MM
	_, err := time.Parse(constant.YearMonthLayout, u.StatMonth)
	if err != nil {
		return fmt.Errorf("stat_month format must be YYYY-MM, invalid: %w", err)
	}

	// configs 可以为空数组（表示删除该月份下的所有配置）
	// 如果不为空，需要验证每个配置项
	for i, cfg := range u.Configs {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configs[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

// validateBkBizID 验证业务ID
func (u *ApplyOrderStatisticsConfigItem) validateBkBizID() error {
	if u.BkBizID <= 0 {
		return errors.New("bk_biz_id is required and must be greater than 0")
	}
	return nil
}

// validateSubOrderIDs 验证子单号
func (u *ApplyOrderStatisticsConfigItem) validateSubOrderIDs() error {
	hasSubOrderIDs := len(u.SubOrderIDs) > 0
	if hasSubOrderIDs && len(u.SubOrderIDs) > 100 {
		return errors.New("sub_order_ids length must not exceed 100")
	}
	return nil
}

// validateTimeRangeFormat 验证时间范围格式
func (u *ApplyOrderStatisticsConfigItem) validateTimeRangeFormat() (time.Time, time.Time, error) {
	startTime, err := time.Parse(constant.DateTimeLayout, u.StartAt)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("start_at format invalid: %w", err)
	}

	endTime, err := time.Parse(constant.DateTimeLayout, u.EndAt)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("end_at format invalid: %w", err)
	}

	return startTime, endTime, nil

}

// validateTimeRangeLogic 验证时间范围逻辑
func (u *ApplyOrderStatisticsConfigItem) validateTimeRangeLogic() error {

	hasStartAt := len(u.StartAt) > 0
	hasEndAt := len(u.EndAt) > 0

	// 如果设置了时间范围的一部分，start_at 和 end_at 必须同时提供
	if hasStartAt && !hasEndAt {
		return errors.New("end_at is required when start_at is set")
	}

	if !hasStartAt && hasEndAt {
		return errors.New("start_at is required when end_at is set")
	}

	// 如果两个都为空，直接返回
	if !hasStartAt && !hasEndAt {
		return nil
	}

	// 如果提供了完整的时间范围，验证格式和时间逻辑
	startTime, endTime, err := u.validateTimeRangeFormat()
	if err != nil {
		return err
	}

	// 验证结束时间不能早于开始时间

	if endTime.Before(startTime) {
		return errors.New("end_at must not be earlier than start_at")
	}

	return nil
}

// validateSubOrderIDsAndTimeRange 验证子单号和时间范围的关系
func (u *ApplyOrderStatisticsConfigItem) validateSubOrderIDsAndTimeRange() error {
	hasSubOrderIDs := len(u.SubOrderIDs) > 0
	hasStartAt := len(u.StartAt) > 0
	hasEndAt := len(u.EndAt) > 0
	hasTimeRange := hasStartAt && hasEndAt

	// 子单号和时间范围不能都为空
	if !hasSubOrderIDs && !hasTimeRange {
		return errors.New("sub_order_ids and time range cannot be empty, at least one must be provided")
	}

	return nil
}

// Validate 验证配置项
func (u *ApplyOrderStatisticsConfigItem) Validate() error {
	if err := u.validateBkBizID(); err != nil {
		return err
	}

	if err := u.validateSubOrderIDs(); err != nil {
		return err
	}

	if err := u.validateTimeRangeLogic(); err != nil {
		return err
	}

	if err := u.validateSubOrderIDsAndTimeRange(); err != nil {
		return err
	}

	return nil
}

// ListApplyOrderStatisticsConfigParam 查询申请单统计配置请求参数
type ListApplyOrderStatisticsConfigParam struct {
	StatMonth string `json:"stat_month" validate:"required"`
}

// Validate 验证查询申请单统计配置请求参数
func (l *ListApplyOrderStatisticsConfigParam) Validate() error {
	// 验证年月格式 YYYY-MM
	_, err := time.Parse(constant.YearMonthLayout, l.StatMonth)
	if err != nil {
		return fmt.Errorf("stat_month format must be YYYY-MM, invalid: %w", err)
	}

	return nil
}

// ListApplyOrderStatisticsConfigResult 查询申请单统计配置响应结果
type ListApplyOrderStatisticsConfigResult struct {
	Details []ApplyOrderStatisticsConfigDetail `json:"details"`
}

// ApplyOrderStatisticsConfigDetail 申请单统计配置详情
type ApplyOrderStatisticsConfigDetail struct {
	ID          string   `json:"id"`
	StatMonth   string   `json:"stat_month"`
	BkBizID     int64    `json:"bk_biz_id"`
	SubOrderIDs []string `json:"sub_order_ids"`
	StartAt     string   `json:"start_at"`
	EndAt       string   `json:"end_at"`
	Memo        string   `json:"memo"`
}

// ListApplyOrderStatisticsYearMonthsResult 查询申请单统计配置月份列表响应结果
type ListApplyOrderStatisticsYearMonthsResult struct {
	Count   int64                           `json:"count"`
	Details []ApplyOrderStatisticsYearMonth `json:"details"`
}

// ApplyOrderStatisticsYearMonth 申请单统计配置月份
type ApplyOrderStatisticsYearMonth struct {
	StatMonth string `json:"stat_month"`
}
