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

	"github.com/stretchr/testify/assert"
	"hcm/pkg/tools/converter"
)

func TestIsPtrEqual(t *testing.T) {
	// 测试 string 类型指针
	t.Run("string", func(t *testing.T) {
		val1 := "hello"
		val2 := "world"

		// 两个 nil 指针
		assert.True(t, IsPtrEqual[string](nil, nil))

		// 两个非 nil 指针，值相同
		assert.True(t, IsPtrEqual(&val1, &val1))
		sameVal := "hello"
		assert.True(t, IsPtrEqual(&val1, &sameVal))

		// 两个非 nil 指针，值不同
		assert.False(t, IsPtrEqual(&val1, &val2))

		// 一个 nil，一个非 nil
		assert.False(t, IsPtrEqual(nil, &val1))
		assert.False(t, IsPtrEqual(&val1, nil))
	})

	// 测试 bool 类型指针
	t.Run("bool", func(t *testing.T) {
		trueVal := true
		falseVal := false

		// 两个 nil 指针
		assert.True(t, IsPtrEqual[bool](nil, nil))

		// 两个非 nil 指针，值相同
		assert.True(t, IsPtrEqual(&trueVal, &trueVal))
		anotherTrue := true
		assert.True(t, IsPtrEqual(&trueVal, &anotherTrue))

		// 两个非 nil 指针，值不同
		assert.False(t, IsPtrEqual(&trueVal, &falseVal))

		// 一个 nil，一个非 nil
		assert.False(t, IsPtrEqual(nil, &trueVal))
		assert.False(t, IsPtrEqual(&trueVal, nil))
	})

	// 测试 int64 类型指针
	t.Run("int64", func(t *testing.T) {
		val1 := int64(100)
		val2 := int64(200)

		assert.True(t, IsPtrEqual[int64](nil, nil))
		assert.True(t, IsPtrEqual(&val1, &val1))
		sameVal := int64(100)
		assert.True(t, IsPtrEqual(&val1, &sameVal))
		assert.False(t, IsPtrEqual(&val1, &val2))
		assert.False(t, IsPtrEqual(nil, &val1))
		assert.False(t, IsPtrEqual(&val1, nil))
	})

	// 测试 float64 类型指针
	t.Run("float64", func(t *testing.T) {
		val1 := 3.14
		val2 := 2.71

		assert.True(t, IsPtrEqual[float64](nil, nil))
		assert.True(t, IsPtrEqual(&val1, &val1))
		sameVal := 3.14
		assert.True(t, IsPtrEqual(&val1, &sameVal))
		assert.False(t, IsPtrEqual(&val1, &val2))
		assert.False(t, IsPtrEqual(nil, &val1))
		assert.False(t, IsPtrEqual(&val1, nil))
	})

	// 测试 int32 类型指针
	t.Run("int32", func(t *testing.T) {
		val1 := int32(42)
		val2 := int32(99)

		assert.True(t, IsPtrEqual[int32](nil, nil))
		assert.True(t, IsPtrEqual(&val1, &val1))
		sameVal := int32(42)
		assert.True(t, IsPtrEqual(&val1, &sameVal))
		assert.False(t, IsPtrEqual(&val1, &val2))
		assert.False(t, IsPtrEqual(nil, &val1))
		assert.False(t, IsPtrEqual(&val1, nil))
	})

	// 测试零值情况
	t.Run("zero_value", func(t *testing.T) {
		var zeroStr string
		var zeroInt int64

		// 零值非 nil 指针与 nil 指针不等
		assert.False(t, IsPtrEqual(&zeroStr, nil))
		assert.False(t, IsPtrEqual(nil, &zeroStr))
		assert.False(t, IsPtrEqual(&zeroInt, nil))

		// 两个零值非 nil 指针相等
		anotherZeroStr := ""
		assert.True(t, IsPtrEqual(&zeroStr, &anotherZeroStr))
		anotherZeroInt := int64(0)
		assert.True(t, IsPtrEqual(&zeroInt, &anotherZeroInt))
	})
}

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
