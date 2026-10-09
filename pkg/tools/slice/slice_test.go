/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package slice

import (
	"cmp"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopK(t *testing.T) {

	type testCase[T any, SL []T] struct {
		name string
		args SL
	}
	tests := []testCase[int, []int]{
		{
			name: "fixed-1",
			args: []int{9, 8, 1, 5, 1, 8, 0, 6, 0, 2},
		},
		{
			name: "fixed-2",
			args: []int{1, 3, 48, 62, 567, 22, 5, 13, 6, 7, 45, 32},
		},
		{
			name: "fixed-3-same",
			args: []int{1, 1, 1},
		},
		{
			name: "fixed-6-same",
			args: []int{1, 1, 1, 1, 1, 1},
		},
		{
			name: "fixed-9-same",
			args: []int{1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name: "fixed-zero-length",
			args: []int{},
		},
	}
	// generate some random test cases
	for length := 0; length < 15; length++ {
		for i := 1; i <= 5; i++ {
			tests = append(tests, testCase[int, []int]{
				name: fmt.Sprintf("random-%d-length-%d", i, length),
				args: randomSlice(length, -128, 127),
			})
		}
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.name), func(t *testing.T) {
			n := len(tt.args)
			sorted := append([]int{}, tt.args...)
			sort.Ints(sorted)
			t.Logf("origin: %v", tt.args)
			t.Logf("sorted: %v", sorted)
			for i := -1; i <= n+1; i++ {
				t.Run(fmt.Sprintf("top%d", i), func(t *testing.T) {
					data := append([]int{}, tt.args...)
					t.Logf("top: %d of %v", i, data)
					TopKSort(i, data, cmp.Less)
					k := min(n, max(0, i))
					t.Logf("%v %v", data[:n-k], data[n-k:])
					assert.ElementsMatch(t, sorted[n-k:], data[n-k:])
				})
			}
		})
	}
}

func randomSlice(length int, min int, max int) []int {
	data := make([]int, length)
	for i := 0; i < length; i++ {
		data[i] = rand.Intn(max-min) + min
	}
	return data
}
