/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package assert

import (
	"testing"

	"hcm/pkg/tools/converter"
)

func TestIsStringSliceEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "两者都为空",
			a:    []string{},
			b:    []string{},
			want: true,
		},
		{
			name: "a 为空，b 不为空",
			a:    []string{},
			b:    []string{"a"},
			want: false,
		},
		{
			name: "a 不为空，b 为空",
			a:    []string{"a"},
			b:    []string{},
			want: false,
		},
		{
			name: "长度不同",
			a:    []string{"a", "b"},
			b:    []string{"a", "b", "c"},
			want: false,
		},
		{
			name: "a 是 b 的子集（bug 场景）",
			a:    []string{"a", "b"},
			b:    []string{"a", "b", "c"},
			want: false,
		},
		{
			name: "b 是 a 的子集（bug 场景）",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b"},
			want: false,
		},
		{
			name: "元素完全相同，顺序相同",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: true,
		},
		{
			name: "元素完全相同，顺序不同",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "a", "b"},
			want: true,
		},
		{
			name: "元素不同",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "d"},
			want: false,
		},
		{
			name: "有重复元素，a 有重复",
			a:    []string{"a", "b", "b"},
			b:    []string{"a", "b"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStringSliceEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("IsStringSliceEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPtrStringSliceEqual(t *testing.T) {
	strA := converter.ValToPtr("a")
	strB := converter.ValToPtr("b")
	strC := converter.ValToPtr("c")

	tests := []struct {
		name string
		a    []*string
		b    []*string
		want bool
	}{
		{
			name: "两者都为空",
			a:    []*string{},
			b:    []*string{},
			want: true,
		},
		{
			name: "a 为空，b 不为空",
			a:    []*string{},
			b:    []*string{strA},
			want: false,
		},
		{
			name: "a 不为空，b 为空",
			a:    []*string{strA},
			b:    []*string{},
			want: false,
		},
		{
			name: "长度不同",
			a:    []*string{strA, strB},
			b:    []*string{strA, strB, strC},
			want: false,
		},
		{
			name: "a 是 b 的子集（bug 场景）",
			a:    []*string{strA, strB},
			b:    []*string{strA, strB, strC},
			want: false,
		},
		{
			name: "b 是 a 的子集（bug 场景）",
			a:    []*string{strA, strB, strC},
			b:    []*string{strA, strB},
			want: false,
		},
		{
			name: "元素完全相同，顺序相同",
			a:    []*string{strA, strB, strC},
			b:    []*string{strA, strB, strC},
			want: true,
		},
		{
			name: "元素完全相同，顺序不同",
			a:    []*string{strA, strB, strC},
			b:    []*string{strC, strA, strB},
			want: true,
		},
		{
			name: "元素不同",
			a:    []*string{strA, strB, strC},
			b:    []*string{strA, strB, strC},
			want: true,
		},
		{
			name: "指针值相同但指向内容不同",
			a:    []*string{converter.ValToPtr("a"), converter.ValToPtr("b")},
			b:    []*string{converter.ValToPtr("a"), converter.ValToPtr("c")},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPtrStringSliceEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("IsPtrStringSliceEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
