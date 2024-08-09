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

package times

import (
	"fmt"
	"time"
)

// GetLastMonth get last month year and month
func GetLastMonth(year, month int) (int, int, error) {
	if year < 1 {
		return 0, 0, fmt.Errorf("invalid year for related: %d", year)
	}
	if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("invalid month: %d", month)
	}
	y, m := getRelativeMonth(year, month, -1)
	return y, m, nil
}

// IsLastDayOfMonth 判断给定的天是否是该月的最后一天
func IsLastDayOfMonth(month, day int) (bool, error) {
	if month < 1 || month > 12 {
		return false, fmt.Errorf("invalid month: %d", month)
	}

	// 获取当前年份
	year := time.Now().Year()

	// 创建当月的下个月的第一天
	var nextMonth time.Month
	if month == 12 {
		nextMonth = 1
		year++
	} else {
		nextMonth = time.Month(month + 1)
	}
	firstDayOfNextMonth := time.Date(year, nextMonth, 1, 0, 0, 0, 0, time.UTC)

	// 获取当月的最后一天
	lastDayOfCurrentMonth := firstDayOfNextMonth.AddDate(0, 0, -1).Day()

	// 比较提供的天和当月的最后一天
	if day == lastDayOfCurrentMonth {
		return true, nil
	}
	return false, nil
}

// AddDaysToDate 计算给定日期在间隔 n 天之后的日期
func AddDaysToDate(year, month, day, n int) (int, time.Month, int, error) {
	// 创建日期对象
	startDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	// 检查日期是否合法
	if startDate.Year() != year || startDate.Month() != time.Month(month) || startDate.Day() != day {
		return 0, 0, 0, fmt.Errorf("invalid date: %d-%d-%d", year, month, day)
	}

	// 增加间隔天数
	resultDate := startDate.AddDate(0, 0, n)

	// 返回新的日期
	return resultDate.Year(), resultDate.Month(), resultDate.Day(), nil
}

// GetFirstDayOfMonth 获取指定年月的第一天
func GetFirstDayOfMonth(year int, month int) (int, error) {
	if month < 1 || month > 12 {
		return 0, fmt.Errorf("invalid month: %d", month)
	}
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return firstDay.Day(), nil
}

// GetLastDayOfMonth 获取指定年月的最后一天
func GetLastDayOfMonth(year int, month int) (int, error) {
	if month < 1 || month > 12 {
		return 0, fmt.Errorf("invalid month: %d", month)
	}

	// 获取下个月的第一天
	var nextMonth time.Month
	nextYear := year
	if month == 12 {
		nextMonth = 1
		nextYear++
	} else {
		nextMonth = time.Month(month + 1)
	}

	firstDayOfNextMonth := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, time.UTC)
	// 获取当前月的最后一天
	lastDay := firstDayOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Day(), nil
}

// GetRelativeMonth 计算相对月份和年份
func GetRelativeMonth(base time.Time, offset int) (int, int) {
	return getRelativeMonth(base.Year(), int(base.Month()), offset)
}

// getRelativeMonth 计算相对月份和年份
func getRelativeMonth(year, month, offset int) (int, int) {
	cur := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	wanted := cur.AddDate(0, offset, 0)
	return wanted.Year(), int(wanted.Month())
}

// GetCurrentMonthUTC current month in utc
func GetCurrentMonthUTC() (year int, month int) {
	return GetRelativeMonthUTC(0)

}

// GetLastMonthUTC last month in utc
func GetLastMonthUTC() (year int, month int) {
	return GetRelativeMonthUTC(-1)
}

// GetRelativeMonthUTC relative month in UTC
func GetRelativeMonthUTC(offset int) (int, int) {
	now := time.Now().UTC()
	return getRelativeMonth(now.Year(), int(now.Month()), offset)
}
