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

package adaptor

import (
	"encoding/base64"
	"sort"
	"sync"
	"testing"
	"time"

	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"
)

var (
	benchP99Once sync.Once
	benchP99Ms   float64
)

// BenchmarkTFPluginAdaptorExecute 基准测试在 Schema 命中与内存 Mock ProviderIF 条件下，
// 单次 Execute 全链路（静态校验 + ValidateResourceTypeConfig + ReadResource + ComposeProposedState +
// PlanResourceChange + ApplyResourceChange + nullifyUnknowns + Close）进程内编排耗时，校验 1,000 次 P99 < 5ms。
func BenchmarkTFPluginAdaptorExecute(b *testing.B) {
	mp := &mockProvider{}
	schema := buildTestPluginSchema()
	kt := kit.New()
	privB64 := base64.StdEncoding.EncodeToString([]byte("bench-private-bytes"))

	req := &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{
			Plugin:           "tencentcloud",
			PluginInstanceID: "inst-bench-001",
			TypeName:         "tencentcloud_instance",
		},
		PriorState: map[string]any{
			"id":            "ins-bench-01",
			"instance_name": "vm-bench-old",
			"image_id":      "img-bench-old",
			"clouds": map[string]any{
				"names": []any{
					map[string]any{"region": "ap-guangzhou", "name": "auto-gz"},
				},
			},
		},
		Config: map[string]any{
			"instance_name": "vm-bench-new",
			"clouds": map[string]any{
				"names": []any{
					map[string]any{"region": "ap-shanghai"},
				},
			},
		},
		Recompute:    []string{"clouds.names[0].name"},
		PriorPrivate: cvt.ValToPtr(privB64),
	}

	benchP99Once.Do(func() {
		const sampleCount = 1000
		durations := make([]time.Duration, 0, sampleCount)
		for i := 0; i < sampleCount; i++ {
			start := time.Now()
			adp := &TFPluginAdaptor{
				pluginName: "tencentcloud",
				instanceID: "inst-bench-001",
				schema:     schema,
				lease:      mp,
			}
			_, err := adp.Execute(kt, req)
			adp.Close()
			if err != nil {
				b.Fatalf("execute failed at sample %d: %v", i, err)
			}
			durations = append(durations, time.Since(start))
		}

		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		p99 := durations[sampleCount*99/100]
		benchP99Ms = float64(p99.Microseconds()) / 1000.0
		if p99 >= 5*time.Millisecond {
			b.Fatalf("expected P99 < 5ms over 1,000 iterations, got %v", p99)
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adp := &TFPluginAdaptor{
			pluginName: "tencentcloud",
			instanceID: "inst-bench-001",
			schema:     schema,
			lease:      mp,
		}
		if _, err := adp.Execute(kt, req); err != nil {
			b.Fatalf("execute failed: %v", err)
		}
		adp.Close()
	}
	b.ReportMetric(benchP99Ms, "p99_ms")
}
