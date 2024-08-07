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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetLastMonth(t *testing.T) {
	testCases := []struct {
		year        int
		month       int
		resultYear  int
		resultMonth int
		hasErr      bool
	}{
		{
			year:        2024,
			month:       5,
			resultYear:  2024,
			resultMonth: 4,
			hasErr:      false,
		},
		{
			year:        2024,
			month:       1,
			resultYear:  2023,
			resultMonth: 12,
			hasErr:      false,
		},
		{
			year:   2024,
			month:  110,
			hasErr: true,
		},
	}
	for _, testCase := range testCases {
		tmpY, tmpM, err := GetLastMonth(testCase.year, testCase.month)
		if testCase.hasErr {
			assert.Error(t, err)
			continue
		}
		assert.Equal(t, testCase.resultYear, tmpY, "should be equal")
		assert.Equal(t, testCase.resultMonth, tmpM, "should be equal")
		if !testCase.hasErr {
			assert.NoError(t, err, "no error")
		}
	}
}

func TestGetRelativeMonth(t *testing.T) {
	type args struct {
		base   time.Time
		mRange int
	}
	tests := []struct {
		name       string
		args       args
		startYear  int
		startMonth int
	}{
		{
			name: "2024/07-0",
			args: args{
				base:   getDate(2024, 7),
				mRange: 0,
			},
			startYear:  2024,
			startMonth: 7,
		},
		{
			name: "2024/07-1",
			args: args{
				base:   getDate(2024, 7),
				mRange: 1,
			},
			startYear:  2024,
			startMonth: 6,
		},
		{
			name: "2024/07-7",
			args: args{
				base:   getDate(2024, 7),
				mRange: 7,
			},
			startYear:  2023,
			startMonth: 12,
		},
		{
			name: "2024/07-30",
			args: args{
				base:   getDate(2024, 7),
				mRange: 30,
			},
			startYear:  2022,
			startMonth: 1,
		},
		{
			name: "2024/07-40",
			args: args{
				base:   getDate(2024, 7),
				mRange: 40,
			},
			startYear:  2021,
			startMonth: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correctYear := tt.startYear
			correctMonth := tt.startMonth

			for offset := -tt.args.mRange; offset <= tt.args.mRange; offset++ {
				gotYear, gotMonth := GetRelativeMonth(tt.args.base, offset)
				assert.Equalf(t, correctYear, gotYear, "CalRelativeMonth(%v, %v, %v)",
					tt.args.base.Year(), tt.args.base.Month(), offset)
				assert.Equalf(t, correctMonth, gotMonth, "CalRelativeMonth(%v, %v, %v)",
					tt.args.base.Year(), tt.args.base.Month(), offset)
				correctMonth += 1
				if correctMonth > 12 {
					correctMonth = 1
					correctYear += 1
				}
			}

		})
	}
}

func getDate(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}
