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

package billitem

import (
	"testing"

	dsbill "hcm/pkg/api/data-service/bill"
)

func Test_generateRemainingPullTask(t *testing.T) {
	type args struct {
		existBillDays []int
		summary       *dsbill.BillSummaryMain
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "Test_generateRemainingPullTask_1",
			args: args{
				existBillDays: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				summary: &dsbill.BillSummaryMain{
					BillYear:  2024,
					BillMonth: 7,
				},
			},
			wantLen: 21,
		},
		{
			name: "Test_generateRemainingPullTask_2",
			args: args{
				existBillDays: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				summary: &dsbill.BillSummaryMain{
					BillYear:  2024,
					BillMonth: 6,
				},
			},
			wantLen: 20,
		},
		{
			name: "Test_generateRemainingPullTask_3",
			args: args{
				existBillDays: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				summary: &dsbill.BillSummaryMain{
					BillYear:  2024,
					BillMonth: 2,
				},
			},
			wantLen: 19,
		},
		{
			name: "Test_generateRemainingPullTask_4",
			args: args{
				existBillDays: []int{},
				summary: &dsbill.BillSummaryMain{
					BillYear:  2024,
					BillMonth: 7,
				},
			},
			wantLen: 31,
		},
		{
			name: "Test_generateRemainingPullTask_5",
			args: args{
				existBillDays: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
					24, 25, 26, 27, 28, 29, 30, 31},
				summary: &dsbill.BillSummaryMain{
					BillYear:  2024,
					BillMonth: 7,
				},
			},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateRemainingPullTask(tt.args.existBillDays, tt.args.summary); len(got) != tt.wantLen {
				t.Errorf("generateRemainingPullTask() = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
