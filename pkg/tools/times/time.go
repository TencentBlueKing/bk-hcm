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
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
)

// ConvStdTimeFormat 转为HCM标准时间格式
func ConvStdTimeFormat(t time.Time) string {
	return t.In(time.Local).Format(constant.TimeStdFormat)
}

// ConvStdTimeNow 转为HCM标准时间格式的当前时间
func ConvStdTimeNow() time.Time {
	return time.Now().In(time.Local)
}

// ParseToStdTime parse layout format time to std time.
func ParseToStdTime(layout, t string) (string, error) {
	tm, err := time.Parse(layout, t)
	if err != nil {
		return "", err
	}

	return tm.In(time.Local).Format(constant.TimeStdFormat), nil
}

// Day 24 hours
const Day = time.Hour * 24

// DaysInMonth 返回给定年份和月份的天数
func DaysInMonth(year int, month time.Month) int {
	// 获取下个月的第一天
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	// 获取本月的最后一天
	lastOfThisMonth := firstOfNextMonth.AddDate(0, 0, -1)

	return lastOfThisMonth.Day()
}

// GetMonthDays 获取指定年月的天数列表
func GetMonthDays(year int, month time.Month) []int {
	lastDay := DaysInMonth(year, month)
	// 创建日期列表
	days := make([]int, lastDay)
	for day := 1; day <= int(lastDay); day++ {
		days[day-1] = day
	}
	return days
}

// ParseDateTime parse date time from string.
func ParseDateTime(layout, t string) (time.Time, error) {
	if len(t) == 0 {
		return time.Time{}, errf.New(errf.InvalidParameter, "empty date time")
	}

	pdTime, err := time.Parse(layout, t)
	if err != nil {
		return time.Time{}, errf.Newf(errf.InvalidParameter, "invalid date time format, should be like %s, err: %v",
			layout, err)
	}

	return pdTime, nil
}
