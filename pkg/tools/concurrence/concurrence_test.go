package concurrence

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBaseExecWithResult(t *testing.T) {
	type args[T any, R any] struct {
		concurrenceLimit int
		params           []T
		execFunc         func(param T) (R, error)
	}
	type testCase[T any, R any] struct {
		name    string
		args    args[T, R]
		want    []R
		wantErr bool
	}
	tests := []testCase[string, string]{
		{
			name: "happy path",
			args: args[string, string]{
				concurrenceLimit: 3,
				params:           []string{"1", "2", "3"},
				execFunc: func(param string) (string, error) {
					return param, nil
				},
			},
			want:    []string{"1", "2", "3"},
			wantErr: false,
		},
		{
			name: "error case",
			args: args[string, string]{
				concurrenceLimit: 3,
				params:           []string{"1", "2", "3"},
				execFunc: func(param string) (string, error) {
					return "", fmt.Errorf("this is an error")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BaseExecWithResult(tt.args.concurrenceLimit, tt.args.params, tt.args.execFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseExecWithResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseExecWithResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}
