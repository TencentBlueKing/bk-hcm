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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// threeModuleRootTF 三模块测试工程的根模块 HCL 配置。
const threeModuleRootTF = `
variable "subnet_cidr" { type = string }
provider "tencentcloud" { plugin_instance_id = "00000001" }
provider "tencentcloud" {
  alias              = "prod"
  plugin_instance_id = "00000009"
}
provider "bkcmdb" { plugin_instance_id = "00000002" }
data "bkcmdb_biz" "target" { bk_biz_name = "hcm" }
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
`

// threeModuleNetworkTF 三模块测试工程的 network 子模块 HCL 配置。
const threeModuleNetworkTF = `
variable "subnet_cidr" { type = string }
import {
  to = tencentcloud_vpc.existing
  id = "vpc-existing-001"
}
resource "tencentcloud_subnet" "app" {
  vpc_id            = tencentcloud_vpc.existing.id
  name              = "app-subnet"
  cidr_block        = var.subnet_cidr
  availability_zone = "ap-guangzhou-3"
}
output "out_subnet_id" { value = tencentcloud_subnet.app.id }
`

// threeModuleAppTF 三模块测试工程的 app 子模块 HCL 配置。
const threeModuleAppTF = `
variable "subnet_id" { type = string }
variable "biz_id"    { type = number }
resource "tencentcloud_instance" "web" {
  instance_name     = "prod-web-01"
  availability_zone = "ap-guangzhou-3"
  image_id          = "img-demo-01"
  instance_type     = "S5.MEDIUM4"
  subnet_id         = var.subnet_id
}
resource "bkcmdb_host" "web" {
  bk_host_innerip = tencentcloud_instance.web.private_ip
  bk_biz_id       = var.biz_id
  operator        = "user"
  tags {
    key   = "subnet_id"
    value = var.subnet_id
  }
  lifecycle {
    create_before_destroy = true
    ignore_changes        = [operator]
    replace_triggered_by  = [tencentcloud_instance.web.image_id]
    recompute             = ["bk_module_id"]
  }
}
`

// buildThreeModuleProjectReq 构造包含根模块、network 子模块和 app 子模块的 ParseReq。
func buildThreeModuleProjectReq() *ParseReq {
	return &ParseReq{
		Vars: map[string]any{"subnet_cidr": "127.0.0.1/24"},
		Modules: map[string]map[string][]byte{
			".":                 {"main.tf": []byte(threeModuleRootTF)},
			"./modules/network": {"main.tf": []byte(threeModuleNetworkTF)},
			"./modules/app":     {"main.tf": []byte(threeModuleAppTF)},
		},
	}
}

// TestParser_Parse_EndToEndAndDeterminism 测试三模块工程的跨模块解析结果，以及连续跑 100 次 Parse 的输出稳定性。
func TestParser_Parse_EndToEndAndDeterminism(t *testing.T) {
	parser := NewParser()
	req := buildThreeModuleProjectReq()
	res, err := parser.Parse(req)
	assert.NoError(t, err)
	assert.Len(t, res.Blocks, 5)

	// 拓扑排序后，没有前置依赖的 import 和 data 排在最前面
	assert.Equal(t, "module.network.import.tencentcloud_vpc.existing", res.Blocks[0].BlockID)
	assert.Equal(t, "00000009", res.Blocks[0].PluginInstanceID)
	assert.Equal(t, map[string]any{"import_id": "vpc-existing-001"}, res.Blocks[0].Config)

	assert.Equal(t, "data.bkcmdb_biz.target", res.Blocks[1].BlockID)
	assert.Equal(t, "00000002", res.Blocks[1].PluginInstanceID)

	// 跨模块的 output 和 variable 引用穿透还原成底层真实的 resource BlockID
	assert.Equal(t, "module.network.tencentcloud_subnet.app", res.Blocks[2].BlockID)
	assert.Equal(t, []string{"module.network.import.tencentcloud_vpc.existing"}, res.Blocks[2].DependsOn)
	assert.Equal(t, "module.network.import.tencentcloud_vpc.existing.id", res.Blocks[2].PendingExprs["vpc_id"])

	assert.Equal(t, "module.app.tencentcloud_instance.web", res.Blocks[3].BlockID)
	assert.Equal(t, []string{"module.network.tencentcloud_subnet.app"}, res.Blocks[3].DependsOn)
	assert.Equal(t, "module.network.tencentcloud_subnet.app.id", res.Blocks[3].PendingExprs["subnet_id"])

	// 子模块没在 providers 里显式指定 bkcmdb，自动继承根模块的默认 bkcmdb 配置
	host := res.Blocks[4]
	assert.Equal(t, "module.app.bkcmdb_host.web", host.BlockID)
	assert.Equal(t, "00000002", host.PluginInstanceID)
	assert.Equal(t, []string{
		"data.bkcmdb_biz.target",
		"module.app.tencentcloud_instance.web",
		"module.network.tencentcloud_subnet.app",
	}, host.DependsOn)
	assert.Equal(t, map[string]string{
		"bk_host_innerip": "module.app.tencentcloud_instance.web.private_ip",
		"bk_biz_id":       "data.bkcmdb_biz.target.bk_biz_id",
		"tags.0.value":    "module.network.tencentcloud_subnet.app.id",
	}, host.PendingExprs)
	assert.Equal(t, []string{"module.app.tencentcloud_instance.web.image_id"}, host.Lifecycle.ReplaceTriggeredBy)

	// 连续跑 100 次 Parse，确保每次输出的 JSON 完全一致（不受 map 遍历顺序影响）
	baselineJSON, err := json.Marshal(res)
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		runRes, runErr := parser.Parse(req)
		assert.NoError(t, runErr)
		runJSON, mErr := json.Marshal(runRes)
		assert.NoError(t, mErr)
		assert.Equal(t, string(baselineJSON), string(runJSON))
	}
}

// buildPairwiseModuleNamedModuleReq 构造子模块名和资源名都叫 "module" 的嵌套模块测试请求。
func buildPairwiseModuleNamedModuleReq() *ParseReq {
	return &ParseReq{
		Modules: map[string]map[string][]byte{
			".": {
				"main.tf": []byte(`
provider "tencentcloud" { plugin_instance_id = "00000001" }
module "network" { source = "./modules/network" }
resource "tencentcloud_subnet" "app" { vpc_id = module.network.vpc_id }
`),
			},
			"./modules/network": {
				"main.tf": []byte(`
module "module" { source = "./modules/inner" }
output "vpc_id" { value = module.module.inner_vpc_id }
`),
			},
			"./modules/inner": {
				"main.tf": []byte(`
resource "tencentcloud_vpc" "module" { name = "vpc-named-module" }
output "inner_vpc_id" { value = tencentcloud_vpc.module.id }
`),
			},
		},
	}
}

// buildProvidersMappingReq 构造带 providers alias 显式映射和默认继承的子模块测试请求。
func buildProvidersMappingReq() *ParseReq {
	return &ParseReq{
		Modules: map[string]map[string][]byte{
			".": {
				"main.tf": []byte(`
provider "tencentcloud" { plugin_instance_id = "00000001" }
provider "tencentcloud" {
  alias              = "prod"
  plugin_instance_id = "00000009"
}
module "network" {
  source    = "./modules/child"
  providers = { tencentcloud = tencentcloud.prod }
}
module "dev" { source = "./modules/child" }
`),
			},
			"./modules/child": {
				"main.tf": []byte(`resource "tencentcloud_vpc" "main" { name = "vpc" }`),
			},
		},
	}
}

// TestParser_Parse_CrossModuleAndMultiInstance 测试同名 module 嵌套、module for_each、resource count、[*] 引用和跨模块 provider 传递。
func TestParser_Parse_CrossModuleAndMultiInstance(t *testing.T) {
	testCases := []struct {
		name    string
		req     *ParseReq
		checkFn func(t *testing.T, res *ParseResult)
	}{
		{
			name: "nested modules and resources named module",
			req:  buildPairwiseModuleNamedModuleReq(),
			checkFn: func(t *testing.T, res *ParseResult) {
				// 即使子模块名和资源名都叫 "module"，也能正确拆出 BlockID
				assert.Len(t, res.Blocks, 2)
				assert.Equal(t, "module.network.module.module.tencentcloud_vpc.module", res.Blocks[0].BlockID)
				assert.Equal(t, "tencentcloud_subnet.app", res.Blocks[1].BlockID)
				assert.Equal(t, []string{"module.network.module.module.tencentcloud_vpc.module"}, res.Blocks[1].DependsOn)
				assert.Equal(t, "module.network.module.module.tencentcloud_vpc.module.id", res.Blocks[1].PendingExprs["vpc_id"])
			},
		},
		{
			name:    "module for_each, resource count, index reference, and splat reference",
			req:     buildMultiInstanceAndSplatReq(),
			checkFn: assertMultiInstanceAndSplatBlocks,
		},
		{
			name: "cross-module providers mapping and default provider inheritance",
			req:  buildProvidersMappingReq(),
			checkFn: func(t *testing.T, res *ParseResult) {
				assert.Len(t, res.Blocks, 2)
				assert.Equal(t, "module.dev.tencentcloud_vpc.main", res.Blocks[0].BlockID)
				assert.Equal(t, "00000001", res.Blocks[0].PluginInstanceID)
				assert.Equal(t, "module.network.tencentcloud_vpc.main", res.Blocks[1].BlockID)
				assert.Equal(t, "00000009", res.Blocks[1].PluginInstanceID)
			},
		},
		{
			name:    "resource count referencing length of another resource",
			req:     buildChainedCountReq(),
			checkFn: assertChainedCountBlocks,
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := parser.Parse(tc.req)
			assert.NoError(t, err)
			tc.checkFn(t, res)
		})
	}
}

// assertMultiInstanceAndSplatBlocks 检查 for_each/count 展开后的实例 BlockID，以及按索引引用和 [*] 引用的 DependsOn。
func assertMultiInstanceAndSplatBlocks(t *testing.T, res *ParseResult) {
	blockMap := make(map[string]*Block, len(res.Blocks))
	for _, b := range res.Blocks {
		blockMap[b.BlockID] = b
	}
	assert.Contains(t, blockMap, `module.vpc["gz"].tencentcloud_vpc.main`)
	assert.Contains(t, blockMap, `module.vpc["sh"].tencentcloud_vpc.main`)
	assert.Contains(t, blockMap, "tencentcloud_subnet.sub[0]")
	assert.Contains(t, blockMap, "tencentcloud_subnet.sub[1]")
	// sub[count.index] 只依赖对应下标的实例，sub[*] 则依赖所有展开的实例
	assert.Equal(t, []string{"tencentcloud_subnet.sub[0]"}, blockMap["tencentcloud_instance.app[0]"].DependsOn)
	assert.Equal(t, []string{"tencentcloud_subnet.sub[1]"}, blockMap["tencentcloud_instance.app[1]"].DependsOn)
	assert.Equal(t, []string{
		"tencentcloud_subnet.sub[0]",
		"tencentcloud_subnet.sub[1]",
	}, blockMap["tencentcloud_clb.lb"].DependsOn)
	assert.Equal(t, "[tencentcloud_subnet.sub[0].id, tencentcloud_subnet.sub[1].id]", blockMap["tencentcloud_clb.lb"].PendingExprs["subnets"])
}

// buildMultiInstanceAndSplatReq 构造带 module for_each、resource count 和 [*] 通配引用的测试请求。
func buildMultiInstanceAndSplatReq() *ParseReq {
	rootTF := `
provider "tencentcloud" { plugin_instance_id = "00000001" }
module "vpc" {
  source   = "./modules/vpc"
  for_each = toset(["sh", "gz"])
  region   = each.key
}
resource "tencentcloud_subnet" "sub" {
  count = 2
  name  = "sub-${count.index}"
}
resource "tencentcloud_subnet" "zero" {
  count = 0
  name  = "none"
}
resource "tencentcloud_subnet" "map_each" {
  for_each = { a = "zone-a" }
  zone     = each.value
}
resource "tencentcloud_instance" "app" {
  count     = 2
  subnet_id = tencentcloud_subnet.sub[count.index].id
}
resource "tencentcloud_clb" "lb" {
  subnets = tencentcloud_subnet.sub[*].id
}
`
	vpcTF := `
variable "region" { type = string }
resource "tencentcloud_vpc" "main" { name = "vpc-${var.region}" }
`
	return &ParseReq{
		Modules: map[string]map[string][]byte{
			".":             {"main.tf": []byte(rootTF)},
			"./modules/vpc": {"main.tf": []byte(vpcTF)},
		},
	}
}

// buildChainedCountReq 构造 count = length(aws_instance.a) 的测试请求。
func buildChainedCountReq() *ParseReq {
	tf := `
provider "aws" { plugin_instance_id = "00000001" }
resource "aws_instance" "a" {
  count = 3
  name  = "inst-${count.index}"
}
resource "aws_eip" "b" {
  count    = length(aws_instance.a)
  instance = aws_instance.a[count.index].id
}
`
	return &ParseReq{
		Modules: map[string]map[string][]byte{
			".": {"main.tf": []byte(tf)},
		},
	}
}

// assertChainedCountBlocks 验证通过 count = length(aws_instance.a) 展开的节点列表与依赖。
func assertChainedCountBlocks(t *testing.T, res *ParseResult) {
	assert.Len(t, res.Blocks, 6)
	blockMap := make(map[string]*Block, len(res.Blocks))
	for _, b := range res.Blocks {
		blockMap[b.BlockID] = b
	}
	for i := 0; i < 3; i++ {
		instID := fmt.Sprintf("aws_instance.a[%d]", i)
		eipID := fmt.Sprintf("aws_eip.b[%d]", i)
		assert.Contains(t, blockMap, instID)
		assert.Contains(t, blockMap, eipID)
		assert.Equal(t, []string{instID}, blockMap[eipID].DependsOn)
		assert.Equal(t, instID+".id", blockMap[eipID].PendingExprs["instance"])
	}
}

// TestParser_Parse_CycleAndDepthErrors 测试模块循环调用、兄弟模块互相引用 output、resource 循环依赖以及模块嵌套超过 16 层的报错。
func TestParser_Parse_CycleAndDepthErrors(t *testing.T) {
	testCases := []struct {
		name        string
		req         *ParseReq
		errContains string
	}{
		{
			name: "circular module call rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".":   {"main.tf": []byte(`module "a" { source = "./a" }`)},
				"./a": {"main.tf": []byte(`module "b" { source = "./b" }`)},
				"./b": {"main.tf": []byte(`module "a" { source = "./a" }`)},
			}},
			errContains: "cycle",
		},
		{
			name: "circular sibling module output dependency rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".":     {"main.tf": []byte("module \"a\" {\n  source = \"./sub\"\n  in_v = module.b.out_v\n}\nmodule \"b\" {\n  source = \"./sub\"\n  in_v = module.a.out_v\n}")},
				"./sub": {"main.tf": []byte("variable \"in_v\" {}\noutput \"out_v\" { value = var.in_v }")},
			}},
			errContains: "cycle",
		},
		{
			name: "circular resource dependency rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte("provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_vpc\" \"a\" { name = tencentcloud_vpc.b.id }\nresource \"tencentcloud_vpc\" \"b\" { name = tencentcloud_vpc.a.id }")},
			}},
			errContains: "cycle",
		},
		{
			name: "circular count dependency between resources rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte("provider \"aws\" { plugin_instance_id = \"001\" }\nresource \"aws_instance\" \"a\" { count = length(aws_eip.b) }\nresource \"aws_eip\" \"b\" { count = length(aws_instance.a) }")},
			}},
			errContains: "cycle",
		},
		{
			name:        "module recursion exceeding 16 levels rejected",
			req:         buildDeepModuleChainReq(18),
			errContains: "recursion depth",
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := parser.Parse(tc.req)
			assert.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// TestParser_Parse_CountAndProviderErrors 测试 count 引用运行时 data 属性、count 为负数以及 providers 映射引用不存在的 alias 报错。
func TestParser_Parse_CountAndProviderErrors(t *testing.T) {
	testCases := []struct {
		name        string
		req         *ParseReq
		errContains string
	}{
		{
			name: "count depending on runtime data attribute rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte("provider \"bkcmdb\" { plugin_instance_id = \"002\" }\nprovider \"tencentcloud\" { plugin_instance_id = \"001\" }\ndata \"bkcmdb_biz\" \"target\" { bk_biz_name = \"hcm\" }\nresource \"tencentcloud_instance\" \"web\" { count = data.bkcmdb_biz.target.host_count }")},
			}},
			errContains: "count",
		},
		{
			name: "negative count rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte("provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_instance\" \"web\" { count = -1 }")},
			}},
			errContains: "non-negative",
		},
		{
			name: "non-existent provider alias in module providers mapping rejected",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".":     {"main.tf": []byte("provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nmodule \"net\" {\n  source = \"./net\"\n  providers = { tencentcloud = tencentcloud.non_existent }\n}")},
				"./net": {"main.tf": []byte("resource \"tencentcloud_vpc\" \"v\" {}")},
			}},
			errContains: "tencentcloud.non_existent",
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := parser.Parse(tc.req)
			assert.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// buildDeepModuleChainReq 构造链式嵌套指定层数的 module 测试请求。
func buildDeepModuleChainReq(levels int) *ParseReq {
	mods := make(map[string]map[string][]byte, levels+1)
	mods["."] = map[string][]byte{
		"main.tf": []byte(`module "m0" { source = "./m0" }`),
	}
	for i := 0; i < levels; i++ {
		cur := fmt.Sprintf("./m%d", i)
		next := fmt.Sprintf("./m%d", i+1)
		mods[cur] = map[string][]byte{
			"main.tf": []byte(fmt.Sprintf(`module "m%d" { source = %q }`, i+1, next)),
		}
	}
	last := fmt.Sprintf("./m%d", levels)
	mods[last] = map[string][]byte{
		"main.tf": []byte(`locals { x = 1 }`),
	}
	return &ParseReq{Modules: mods}
}
