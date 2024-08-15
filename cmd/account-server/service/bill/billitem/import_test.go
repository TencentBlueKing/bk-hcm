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
