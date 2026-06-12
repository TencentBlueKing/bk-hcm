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

// Package util ...
package util

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGetIntByInterface(t *testing.T) {
	type args struct {
		a interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			args: args{
				a: int(1),
			},
			want: 1,
		},
		{
			args: args{
				a: int32(1),
			},
			want: 1,
		},
		{
			args: args{
				a: int64(1),
			},
			want: 1,
		},
		{
			args: args{
				a: float32(1.01),
			},
			want: 1,
		},
		{
			args: args{
				a: float64(1.01),
			},
			want: 1,
		},
		{
			args: args{
				a: "1",
			},
			want: 1,
		},
		{
			args: args{
				a: json.Number("1"),
			},
			want: 1,
		},
		{
			args: args{
				a: "a",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetIntByInterface(tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIntByInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetIntByInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInt64ByInterface(t *testing.T) {
	type args struct {
		a interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			args: args{
				a: int(1),
			},
			want: 1,
		},
		{
			args: args{
				a: int32(1),
			},
			want: 1,
		},
		{
			args: args{
				a: int64(1),
			},
			want: 1,
		},
		{
			args: args{
				a: float32(1.01),
			},
			want: 1,
		},
		{
			args: args{
				a: float64(1.01),
			},
			want: 1,
		},
		{
			args: args{
				a: "1",
			},
			want: 1,
		},
		{
			args: args{
				a: json.Number("1"),
			},
			want: 1,
		},
		{
			args: args{
				a: "a",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetInt64ByInterface(tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt64ByInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInt64ByInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMapInterfaceByInerface(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []interface{}
		wantErr bool
	}{
		{
			args: args{
				[]int{1, 2, 3},
			},
			want: []interface{}{1, 2, 3},
		},
		{
			args: args{
				[]int64{1, 2, 3},
			},
			want: []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			args: args{
				[]int32{1, 2, 3},
			},
			want: []interface{}{int32(1), int32(2), int32(3)},
		},
		{
			args: args{
				[]string{"1", "2", "3"},
			},
			want: []interface{}{"1", "2", "3"},
		},
		{
			args: args{
				"123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMapInterfaceByInerface(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMapInterfaceByInerface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMapInterfaceByInerface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStrByInterface(t *testing.T) {
	type args struct {
		a interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"string"}, "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStrByInterface(tt.args.a); got != tt.want {
				t.Errorf("GetStrByInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliceStrToInt(t *testing.T) {
	type args struct {
		sliceStr []string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{"", args{[]string{"1"}}, []int{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SliceStrToInt(tt.args.sliceStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceStrToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceStrToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliceStrToInt64(t *testing.T) {
	type args struct {
		sliceStr []string
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{"", args{[]string{"1"}}, []int64{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SliceStrToInt64(tt.args.sliceStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceStrToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceStrToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStrValsFromArrMapInterfaceByKey(t *testing.T) {
	type args struct {
		arrI []interface{}
		key  string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"", args{[]interface{}{map[string]interface{}{"key": "string"}}, "key"}, []string{"string"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStrValsFromArrMapInterfaceByKey(tt.args.arrI, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStrValsFromArrMapInterfaceByKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStrSliceByInterface(t *testing.T) {
	type CustomString *[]string
	cs := CustomString(&[]string{"xxx"})
	nilPtr := CustomString(nil)

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// 预期解析失败的测试用例
		{"customStruct", args{data: struct{}{}}, []string{}, true},
		{"interface", args{data: []interface{}{}}, []string{}, true},
		{"string", args{data: "xxx"}, []string{}, true},
		{"nil", args{data: nil}, []string{}, true},
		// 预期解析成功的测试用例
		{"empty", args{data: []string{}}, []string{}, false},
		{"array", args{data: [1]string{"x"}}, []string{"x"}, false},
		{"nilPtr", args{data: nilPtr}, []string{}, false},
		{"pass", args{data: []string{"x", "y", "z"}}, []string{"x", "y", "z"}, false},
		{"customType", args{data: CustomString(&[]string{"xxx"})}, []string{"xxx"}, false},
		{"nestedPtr", args{data: &cs}, []string{"xxx"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStrSliceByInterface(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStrSliceByInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Logf("GetStrSliceByInterface() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStrSliceByInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}
