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

package daily

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetBillDays(t *testing.T) {
	testCases := []struct {
		billYear  int
		billMonth int
		delay     int
		nowStr    string
		result    []int
	}{
		{
			billYear:  2024,
			billMonth: 5,
			delay:     2,
			nowStr:    "2024-05-02 13:33:37",
			result:    nil,
		},
		{
			billYear:  2024,
			billMonth: 4,
			delay:     20,
			nowStr:    "2024-05-01 13:33:37",
			result:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		{
			billYear:  2024,
			billMonth: 5,
			delay:     0,
			nowStr:    "2024-05-01 13:33:37",
			result:    []int{1},
		},
	}
	for _, test := range testCases {
		now, err := time.Parse("2006-01-02 15:04:05", test.nowStr)
		if err != nil {
			t.Error(err)
			break
		}
		result := getBillDays(test.billYear, test.billMonth, test.delay, now)
		assert.Equal(t, result, test.result, "result should be equal")
	}
}

func Test_getBillDays(t *testing.T) {
	type args struct {
		billYear  int
		billMonth int
		billDelay int
		now       time.Time
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "test_2024-04",
			args: args{
				billYear:  2024,
				billMonth: 4,
				billDelay: 1,
				now:       time.Now(),
			},
			want: genIntSlice(1, 30),
		},
		{
			name: "test_202402",
			args: args{
				billYear:  2024,
				billMonth: 2,
				billDelay: 1,
				now:       time.Now(),
			},
			want: genIntSlice(1, 29),
		},
		{
			name: "test_2023-02",
			args: args{
				billYear:  2023,
				billMonth: 2,
				billDelay: 1,
				now:       time.Now(),
			},
			want: genIntSlice(1, 28),
		},

		{
			name: "test_2023-06",
			args: args{
				billYear:  2023,
				billMonth: 6,
				billDelay: 1,
				now:       time.Now(),
			},
			want: genIntSlice(1, 30),
		},
		{
			name: "test_2024-06 at day 1",
			args: args{
				billYear:  2024,
				billMonth: 6,
				billDelay: 1,
				now:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			want: genIntSlice(1, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getBillDays(tt.args.billYear, tt.args.billMonth, tt.args.billDelay, tt.args.now),
				"getBillDays(%v, %v, %v, %v)", tt.args.billYear, tt.args.billMonth, tt.args.billDelay, tt.args.now)
		})
	}
}
func genIntSlice(start, end int) []int {
	ret := make([]int, 0)
	for i := start; i <= end; i++ {
		ret = append(ret, i)
	}
	return ret
}
