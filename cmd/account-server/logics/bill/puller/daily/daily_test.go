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
