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
