package billitem

import (
	"testing"
)

func Test_getMonthDays(t *testing.T) {
	type args struct {
		year  int
		month int
	}
	tests := []struct {
		name       string
		args       args
		wantLength int
	}{
		{
			name: "test1",
			args: args{
				year:  2024,
				month: 6,
			},
			wantLength: 30,
		},
		{
			name: "test1",
			args: args{
				year:  2000,
				month: 2,
			},
			wantLength: 29,
		},
		{
			name: "test1",
			args: args{
				year:  2001,
				month: 2,
			},
			wantLength: 28,
		},
		{
			name: "test1",
			args: args{
				year:  2000,
				month: 12,
			},
			wantLength: 31,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMonthDays(tt.args.year, tt.args.month); len(got) != tt.wantLength {
				t.Errorf("getMonthDays() = %v, want %v", len(got), tt.wantLength)
			}
		})
	}
}
