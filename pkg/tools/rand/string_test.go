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

package rand

import (
	"testing"

	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/uuid"
)

func BenchmarkStringN(b *testing.B) {

	b.Run("rand_string", func(b *testing.B) {
		result := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			result[i] = String(32)
		}
	})

	b.Run("uuid", func(b *testing.B) {
		result := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			result[i] = uuid.UUID()
		}
	})
}

func TestStringN(t *testing.T) {
	N := 10000
	result := make([]string, N)
	for i := 0; i < N; i++ {
		result[i] = String(4)
	}
	uniqued := slice.Unique(result)
	t.Logf("uniqued: %d, total: %d, rate:%.3f%%\n", len(uniqued), N, float64(len(uniqued))/float64(N)*100)

}

func BenchmarkPrefix(b *testing.B) {
	result := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		result[i] = Prefix("prefix-", 10)
	}
}

func BenchmarkPrefixAddString(b *testing.B) {
	result := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		result[i] = "prefix-" + String(10)
	}
}

func TestPrefix(t *testing.T) {
	type args struct {
		prefix string
		n      int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{"all-0",
			args{
				prefix: "",
				n:      0,
			},
			0,
		},
		{"prefix-0",
			args{
				prefix: "",
				n:      1,
			},
			1,
		},
		{"rand-0",
			args{
				prefix: "xx",
				n:      0,
			},
			2,
		},
		{"normal",
			args{
				prefix: "xx-",
				n:      2,
			},
			5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Prefix(tt.args.prefix, tt.args.n); len(got) != tt.wantLen {
				t.Errorf("Prefix() = %v, wantLen %v", got, tt.wantLen)
			}
		})
	}
}
