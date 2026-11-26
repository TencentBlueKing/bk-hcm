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

package iacparser

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// build50BlocksProjectReq 构造包含 50 个节点（1 个 import + 1 个 data + 48 个 resource）的三模块测试请求。
func build50BlocksProjectReq() *ParseReq {
	return &ParseReq{
		Vars: map[string]any{"subnet_cidr": "127.0.0.1/24"},
		Modules: map[string]map[string][]byte{
			".": {
				"main.tf": []byte(`
variable "subnet_cidr" { type = string }
provider "tencentcloud" { plugin_instance_id = "00000001" }
provider "tencentcloud" {
  alias              = "prod"
  plugin_instance_id = "00000009"
}
provider "bkcmdb" { plugin_instance_id = "00000002" }

data "bkcmdb_biz" "target" {
  bk_biz_name = "hcm"
}

module "network" {
  source      = "./modules/network"
  providers   = { tencentcloud = tencentcloud.prod }
  subnet_cidr = var.subnet_cidr
}

module "app" {
  source    = "./modules/app"
  providers = { tencentcloud = tencentcloud.prod }
  subnet_id = module.network.out_subnet_id
  biz_id    = data.bkcmdb_biz.target.bk_biz_id
}
`),
			},
			"./modules/network": {
				"main.tf": []byte(`
variable "subnet_cidr" { type = string }
import {
  to = tencentcloud_vpc.existing
  id = "vpc-existing-001"
}
resource "tencentcloud_subnet" "app" {
  count             = 24
  vpc_id            = tencentcloud_vpc.existing.id
  name              = "subnet-${count.index}"
  cidr_block        = var.subnet_cidr
  availability_zone = "ap-guangzhou-3"
}
output "out_subnet_id" {
  value = tencentcloud_subnet.app[0].id
}
`),
			},
			"./modules/app": {
				"main.tf": []byte(`
variable "subnet_id" { type = string }
variable "biz_id"    { type = number }
resource "tencentcloud_instance" "web" {
  count         = 24
  instance_name = "web-${count.index}"
  subnet_id     = var.subnet_id
  biz_id        = var.biz_id
}
`),
			},
		},
	}
}

// buildBenchmarkEvalReq 构造带 5 条 PendingExprs 和 5 个上游节点输出的 EvalReq。
func buildBenchmarkEvalReq() *EvalReq {
	return &EvalReq{
		Config: map[string]any{
			"operator": "user",
			"tags": []any{
				map[string]any{"key": "subnet_id", "value": nil},
				map[string]any{"key": "vpc_id", "value": nil},
			},
		},
		PendingExprs: map[string]string{
			"bk_host_innerip": "module.app.tencentcloud_instance.web.private_ip",
			"bk_biz_id":       "tostring(data.bkcmdb_biz.target.bk_biz_id)",
			"cluster_label":   `upper(join("-", [module.cluster.tencentcloud_tke.main.name, module.network.import.tencentcloud_vpc.existing.id]))`,
			"tags.0.value":    "module.network.tencentcloud_subnet.app.id",
			"tags.1.value":    "module.network.import.tencentcloud_vpc.existing.id",
		},
		DependencyStates: map[string]map[string]any{
			"data.bkcmdb_biz.target":                          {"bk_biz_id": 213},
			"module.network.import.tencentcloud_vpc.existing": {"id": "vpc-existing-001"},
			"module.network.tencentcloud_subnet.app":          {"id": "subnet-demo-01"},
			"module.app.tencentcloud_instance.web":            {"id": "ins-001", "private_ip": "127.0.0.1"},
			"module.cluster.tencentcloud_tke.main":            {"name": "prod-tke"},
		},
	}
}

// TestParser_ConcurrentSafety 测试 50 个 goroutine 共用同一个 Parser 并发调用 Parse 和 EvalBlock 是否存在数据竞争。
func TestParser_ConcurrentSafety(t *testing.T) {
	parser := NewParser()
	parseReq := buildThreeModuleProjectReq()
	evalReq := buildBenchmarkEvalReq()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			// 并发调用 Parse 和 EvalBlock，配合 -race 检查并发安全
			pRes, pErr := parser.Parse(parseReq)
			assert.NoError(t, pErr)
			assert.NotNil(t, pRes)
			assert.Len(t, pRes.Blocks, 5)

			eRes, eErr := parser.EvalBlock(evalReq)
			assert.NoError(t, eErr)
			assert.Equal(t, "127.0.0.1", eRes["bk_host_innerip"])
			assert.Equal(t, "213", eRes["bk_biz_id"])
			assert.Equal(t, "PROD-TKE-VPC-EXISTING-001", eRes["cluster_label"])
		}()
	}
	wg.Wait()
}

// TestParser_PerformanceAssertions 检查 50 个节点的 Parse 耗时 < 50ms，单次 EvalBlock 耗时 < 5ms。
func TestParser_PerformanceAssertions(t *testing.T) {
	parser := NewParser()
	req50 := build50BlocksProjectReq()

	// 50 个节点的三模块工程单次 Parse 耗时应小于 50ms
	startParse := time.Now()
	res, err := parser.Parse(req50)
	parseElapsed := time.Since(startParse)
	assert.NoError(t, err)
	assert.Len(t, res.Blocks, 50)
	assert.Less(t, parseElapsed, 50*time.Millisecond)

	// 含 5 条 PendingExprs 的单次 EvalBlock 耗时应小于 5ms
	evalReq := buildBenchmarkEvalReq()
	startEval := time.Now()
	out, err := parser.EvalBlock(evalReq)
	evalElapsed := time.Since(startEval)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Less(t, evalElapsed, 5*time.Millisecond)
}

// BenchmarkParse_50Blocks 测试 50 个节点的 Parse 性能。
func BenchmarkParse_50Blocks(b *testing.B) {
	parser := NewParser()
	req := build50BlocksProjectReq()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := parser.Parse(req)
		if err != nil || len(res.Blocks) != 50 {
			b.Fatalf("unexpected parse failure: %v", err)
		}
	}
}

// BenchmarkEvalBlock 测试单次 EvalBlock 的性能。
func BenchmarkEvalBlock(b *testing.B) {
	parser := NewParser()
	req := buildBenchmarkEvalReq()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := parser.EvalBlock(req)
		if err != nil || out["bk_biz_id"] != "213" {
			b.Fatalf("unexpected eval failure: %v", err)
		}
	}
}
