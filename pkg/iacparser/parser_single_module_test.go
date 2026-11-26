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
	"testing"

	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

// buildSingleModule8BlockTypesReq 构造包含 variable、locals、provider、import、data、resource、output 的单模块测试请求。
func buildSingleModule8BlockTypesReq() *ParseReq {
	varsTF := `
variable "subnet_cidr" { type = string }
variable "zone" {
  type    = string
  default = "ap-guangzhou-3"
}
locals { subnet_name = upper("app-subnet") }
`
	mainTF := `
provider "tencentcloud" { plugin_instance_id = "00000009" }
provider "bkcmdb"       { plugin_instance_id = "00000002" }
import {
  to = tencentcloud_vpc.existing
  id = "vpc-existing-001"
}
data "bkcmdb_biz" "target" { bk_biz_name = "hcm" }
resource "tencentcloud_subnet" "app" {
  vpc_id            = tencentcloud_vpc.existing.id
  name              = local.subnet_name
  cidr_block        = var.subnet_cidr
  availability_zone = var.zone
}
output "out_subnet_id" { value = tencentcloud_subnet.app.id }
`
	return &ParseReq{
		Vars: map[string]any{"subnet_cidr": "127.0.0.1/24"},
		Modules: map[string]map[string][]byte{
			".": {"variables.tf": []byte(varsTF), "main.tf": []byte(mainTF)},
		},
	}
}

// TestParser_Parse_SingleModuleBlocks 测试单模块下各类顶层块的解析，以及同时定义 import 和同名 resource 时自动建立依赖。
func TestParser_Parse_SingleModuleBlocks(t *testing.T) {
	importAndResourceTF := `
provider "tencentcloud" { plugin_instance_id = "00000001" }
import {
  to = tencentcloud_vpc.legacy
  id = "vpc-legacy-001"
}
resource "tencentcloud_vpc" "legacy" { name = "corp-vpc" }
`
	testCases := []struct {
		name      string
		req       *ParseReq
		wantCount int
		checkFn   func(t *testing.T, res *ParseResult)
	}{
		{
			name:      "single module with variable, locals, provider, output, import, data, and resource",
			req:       buildSingleModule8BlockTypesReq(),
			wantCount: 3,
			checkFn:   assertSingleModuleThreeBlocks,
		},
		{
			name: "import and resource both defined for same target automatically links dependency",
			req: &ParseReq{
				Modules: map[string]map[string][]byte{".": {"main.tf": []byte(importAndResourceTF)}},
			},
			wantCount: 2,
			checkFn: func(t *testing.T, res *ParseResult) {
				// 同时定义了 import 和同名 resource 时，import 排在前面，且 resource 自动依赖该 import
				assert.Equal(t, "import.tencentcloud_vpc.legacy", res.Blocks[0].BlockID)
				assert.Equal(t, "tencentcloud_vpc.legacy", res.Blocks[1].BlockID)
				assert.Equal(t, []string{"import.tencentcloud_vpc.legacy"}, res.Blocks[1].DependsOn)
				assert.Nil(t, res.Blocks[1].Lifecycle)
				assert.Nil(t, res.Blocks[1].PendingExprs)
			},
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := parser.Parse(tc.req)
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Len(t, res.Blocks, tc.wantCount)
			tc.checkFn(t, res)
		})
	}
}

// assertSingleModuleThreeBlocks 检查单模块解析出的 import、data 和 resource 三个节点。
func assertSingleModuleThreeBlocks(t *testing.T, res *ParseResult) {
	// 检查 import 节点的字段和 import_id 配置
	assert.Equal(t, "import.tencentcloud_vpc.existing", res.Blocks[0].BlockID)
	assert.Equal(t, enumor.BlockKindImport, res.Blocks[0].Kind)
	assert.NoError(t, res.Blocks[0].Kind.Validate())
	assert.Error(t, enumor.BlockKind("invalid").Validate())
	assert.Equal(t, "tencentcloud", res.Blocks[0].Plugin)
	assert.Equal(t, "00000009", res.Blocks[0].PluginInstanceID)
	assert.Equal(t, "tencentcloud_vpc", res.Blocks[0].TypeName)
	assert.Equal(t, []string{}, res.Blocks[0].DependsOn)
	assert.Equal(t, map[string]any{"import_id": "vpc-existing-001"}, res.Blocks[0].Config)

	// 检查 data 节点的 provider 绑定和查询参数
	assert.Equal(t, "data.bkcmdb_biz.target", res.Blocks[1].BlockID)
	assert.Equal(t, enumor.BlockKindData, res.Blocks[1].Kind)
	assert.Equal(t, "bkcmdb", res.Blocks[1].Plugin)
	assert.Equal(t, "00000002", res.Blocks[1].PluginInstanceID)
	assert.Equal(t, map[string]any{"bk_biz_name": "hcm"}, res.Blocks[1].Config)

	// 模块里只写了 import 没写同名 resource 时，引用 tencentcloud_vpc.existing.id 会自动指向 import 节点
	assert.Equal(t, "tencentcloud_subnet.app", res.Blocks[2].BlockID)
	assert.Equal(t, enumor.BlockKindResource, res.Blocks[2].Kind)
	assert.Equal(t, []string{"import.tencentcloud_vpc.existing"}, res.Blocks[2].DependsOn)
	assert.Equal(t, map[string]any{
		"name":              "APP-SUBNET",
		"cidr_block":        "127.0.0.1/24",
		"availability_zone": "ap-guangzhou-3",
	}, res.Blocks[2].Config)
	assert.Equal(t, map[string]string{
		"vpc_id": "import.tencentcloud_vpc.existing.id",
	}, res.Blocks[2].PendingExprs)
}

// TestParser_Parse_AttrDecompositionAndLifecycle 测试静态属性与运行时引用的拆分、"${var.env}-..." 混合插值处理，以及 lifecycle 提取。
func TestParser_Parse_AttrDecompositionAndLifecycle(t *testing.T) {
	decompTF := `
variable "env" { type = string }
provider "tencentcloud" { plugin_instance_id = "00000009" }
provider "bkcmdb"       { plugin_instance_id = "00000002" }
data "bkcmdb_biz" "target" { bk_biz_name = "hcm" }
resource "tencentcloud_subnet" "app" { name = "sub" }
resource "tencentcloud_instance" "web" { image_id = "img-01" }
resource "bkcmdb_host" "web" {
  operator        = "user"
  host_tag        = "${var.env}-${tencentcloud_instance.web.id}"
  bk_host_innerip = tencentcloud_instance.web.private_ip
  bk_biz_id       = data.bkcmdb_biz.target.bk_biz_id
  tags {
    key   = "subnet_id"
    value = tencentcloud_subnet.app.id
  }
  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = [operator, tags[0].value]
    replace_triggered_by  = [tencentcloud_instance.web.image_id]
    recompute             = ["bk_module_id"]
  }
}
`
	parser := NewParser()
	res, err := parser.Parse(&ParseReq{
		Vars:    map[string]any{"env": "prod"},
		Modules: map[string]map[string][]byte{".": {"main.tf": []byte(decompTF)}},
	})
	assert.NoError(t, err)
	assertDecomposedHostBlock(t, res)
}

// TestParser_Parse_BuiltinFunctionsAndOperators 测试内置函数、算术/逻辑/三元运算符，以及内联 map/slice 里的运行时引用拆分。
func TestParser_Parse_BuiltinFunctionsAndOperators(t *testing.T) {
	fnTF := `
provider "tencentcloud" { plugin_instance_id = "00000001" }
locals {
  az_map = {
    "gz" = ["ap-guangzhou-3", "ap-guangzhou-4"]
  }
  tag_key = "env"
}
resource "tencentcloud_vpc" "base" { name = "vpc" }
resource "tencentcloud_subnet" "sub" {
  name       = join("-", [lower("PROD"), tostring(10), lookup({a = "x"}, "a", "y")])
  az         = local.az_map["gz"][1]
  matrix_ref = tencentcloud_vpc.base.matrix[0][1]
  splat_dyn  = tencentcloud_vpc.base[*].tags[local.tag_key].value
  tags_set   = tolist(toset(["b", "a", "a"]))
  mapped     = tomap({k = tobool("true")})
  num_val    = tonumber("42") + (10 - 2) * 1
  neg_val    = -5
  not_val    = !false
  cond_val   = true ? "yes" : "no"
  cond_pend  = (tencentcloud_vpc.base.enable ? tencentcloud_vpc.base.id : "none")
  bin_pend   = tencentcloud_vpc.base.count + 1
  un_pend    = !tencentcloud_vpc.base.disabled
  inline_obj = { static_k = "v", dyn_k = tencentcloud_vpc.base.id }
  inline_arr = [{ k = "s", v = tencentcloud_vpc.base.id }]
  try_val    = try("ok")
  can_val    = can("ok")
  idx_val    = index(["x", "y"], "y")
  b64_val    = base64decode(base64encode("hello"))
  hash_len   = length(md5("a")) + length(sha1("a")) + length(sha256("a")) + length(urlencode("a b"))
  depends_on = [tencentcloud_vpc.base]
  lifecycle { ignore_changes = all }
}
`
	res, err := NewParser().Parse(&ParseReq{
		Modules: map[string]map[string][]byte{".": {"main.tf": []byte(fnTF)}},
	})
	assert.NoError(t, err)
	sub := res.Blocks[1]
	// 静态表达式和内置函数在 Parse 阶段直接算出结果放入 Config
	assert.Equal(t, "prod-10-x", sub.Config["name"])
	assert.Equal(t, "ap-guangzhou-4", sub.Config["az"])
	assert.Equal(t, "tencentcloud_vpc.base.matrix[0][1]", sub.PendingExprs["matrix_ref"])
	assert.Equal(t, "[tencentcloud_vpc.base.tags[\"env\"].value]", sub.PendingExprs["splat_dyn"])
	assert.Equal(t, []any{"a", "b"}, sub.Config["tags_set"])
	assert.Equal(t, map[string]any{"k": true}, sub.Config["mapped"])
	assert.Equal(t, int64(50), sub.Config["num_val"])
	assert.Equal(t, int64(-5), sub.Config["neg_val"])
	assert.Equal(t, true, sub.Config["not_val"])
	assert.Equal(t, "yes", sub.Config["cond_val"])
	assert.Equal(t, "hello", sub.Config["b64_val"])
	// 内联 map 和 slice 里的动态字段在 Config 中填 nil 占位，表达式放到 PendingExprs
	assert.Equal(t, map[string]any{"static_k": "v", "dyn_k": nil}, sub.Config["inline_obj"])
	assert.Equal(t, "tencentcloud_vpc.base.id", sub.PendingExprs["inline_obj.dyn_k"])
	assert.Equal(t, "tencentcloud_vpc.base.id", sub.PendingExprs["inline_arr.0.v"])
	assert.Equal(t, []string{"tencentcloud_vpc.base"}, sub.DependsOn)
	assert.Equal(t, []string{"all"}, sub.Lifecycle.IgnoreChanges)
}

// assertDecomposedHostBlock 检查 bkcmdb_host.web 节点的 Config、PendingExprs、DependsOn 和 Lifecycle。
func assertDecomposedHostBlock(t *testing.T, res *ParseResult) {
	var host *Block
	for _, b := range res.Blocks {
		if b.BlockID == "bkcmdb_host.web" {
			host = b
		}
	}
	assert.NotNil(t, host)
	// lifecycle 单独提取到 Block.Lifecycle，不放进 Config
	_, hasLifecycleInConfig := host.Config["lifecycle"]
	assert.False(t, hasLifecycleInConfig)
	// 嵌套块里的动态字段在 Config 中留 nil 占位
	assert.Equal(t, map[string]any{
		"operator": "user",
		"tags": []any{
			map[string]any{"key": "subnet_id", "value": nil},
		},
	}, host.Config)
	// "${var.env}-${tencentcloud_instance.web.id}" 里的 var.env 在阶段一直接替换成 "prod"
	assert.Equal(t, map[string]string{
		"bk_biz_id":       "data.bkcmdb_biz.target.bk_biz_id",
		"bk_host_innerip": "tencentcloud_instance.web.private_ip",
		"host_tag":        `"prod-${tencentcloud_instance.web.id}"`,
		"tags.0.value":    "tencentcloud_subnet.app.id",
	}, host.PendingExprs)
	assert.Equal(t, []string{
		"data.bkcmdb_biz.target",
		"tencentcloud_instance.web",
		"tencentcloud_subnet.app",
	}, host.DependsOn)
	assert.Equal(t, &Lifecycle{
		CreateBeforeDestroy: true,
		PreventDestroy:      true,
		IgnoreChanges:       []string{"operator", "tags.0.value"},
		ReplaceTriggeredBy:  []string{"tencentcloud_instance.web.image_id"},
		Recompute:           []string{"bk_module_id"},
	}, host.Lifecycle)
}

// TestParser_Parse_ValidationErrors 测试空请求、HCL 语法错误、缺少必填 variable、非法资源类型名和未配置 provider 等报错场景。
func TestParser_Parse_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name        string
		req         *ParseReq
		errContains string
	}{
		{name: "nil ParseReq rejected", req: nil, errContains: "nil"},
		{name: "empty Modules rejected", req: &ParseReq{Modules: map[string]map[string][]byte{}}, errContains: "empty"},
		{
			name:        "missing root module rejected",
			req:         &ParseReq{Modules: map[string]map[string][]byte{"./sub": {"main.tf": []byte("")}}},
			errContains: "root module",
		},
		{
			name: "unclosed brace syntax error",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte(`resource "tencentcloud_vpc" "main" { name = "x"`)},
			}},
			errContains: "syntax error",
		},
		{
			name: "missing required variable subnet_cidr",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte(`variable "subnet_cidr" { type = string }`)},
			}},
			errContains: "subnet_cidr",
		},
		{
			name: "invalid resource type_name without underscore",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte(`resource "invalidtype" "main" {}`)},
			}},
			errContains: "underscore",
		},
		{
			name: "unbound provider for resource plugin",
			req: &ParseReq{Modules: map[string]map[string][]byte{
				".": {"main.tf": []byte(`resource "aws_instance" "web" { name = "a" }`)},
			}},
			errContains: "provider",
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

// TestParser_Parse_ReferenceErrors 测试引用不存在的 resource/local、非法 lifecycle 字段、locals 循环依赖和无效 depends_on 的报错。
func TestParser_Parse_ReferenceErrors(t *testing.T) {
	testCases := []struct {
		name        string
		tf          string
		errContains string
	}{
		{
			name:        "undefined resource reference",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_subnet\" \"app\" { vpc_id = tencentcloud_vpc.not_exist.id }",
			errContains: "tencentcloud_vpc.not_exist",
		},
		{
			name:        "undefined local reference",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_subnet\" \"app\" { name = local.unknown_val }",
			errContains: "local.unknown_val",
		},
		{
			name:        "unsupported lifecycle attribute",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_subnet\" \"app\" {\n  lifecycle {\n    unknown_opt = true\n  }\n}",
			errContains: "unknown_opt",
		},
		{
			name:        "replace_triggered_by references non-existent resource",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_subnet\" \"app\" {\n  lifecycle {\n    replace_triggered_by = [tencentcloud_instance.none.id]\n  }\n}",
			errContains: "tencentcloud_instance.none",
		},
		{
			name:        "cycle in locals rejected",
			tf:          "locals {\n  a = local.b\n  b = local.a\n}",
			errContains: "cycle",
		},
		{
			name:        "depends_on references static variable without resource dependencies",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nvariable \"region\" { default = \"ap-guangzhou\" }\nresource \"tencentcloud_subnet\" \"app\" {\n  depends_on = [var.region]\n}",
			errContains: "does not reference any resource or data block",
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ParseReq{Modules: map[string]map[string][]byte{".": {"main.tf": []byte(tc.tf)}}}
			res, err := parser.Parse(req)
			assert.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// TestParser_Parse_ForbiddenFunctions 测试 Parse 阶段拦截 file、templatefile 等禁用的文件读取函数。
func TestParser_Parse_ForbiddenFunctions(t *testing.T) {
	testCases := []struct {
		name        string
		tf          string
		errContains string
	}{
		{
			name:        "forbidden file function call in static expression",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_subnet\" \"app\" { name = file(\"config.txt\") }",
			errContains: "file",
		},
		{
			name:        "forbidden templatefile function call in pending expression",
			tf:          "provider \"tencentcloud\" { plugin_instance_id = \"001\" }\nresource \"tencentcloud_vpc\" \"main\" { name = \"v\" }\nresource \"tencentcloud_subnet\" \"app\" { name = templatefile(\"config.tpl\", { id = tencentcloud_vpc.main.id }) }",
			errContains: "templatefile",
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ParseReq{Modules: map[string]map[string][]byte{".": {"main.tf": []byte(tc.tf)}}}
			res, err := parser.Parse(req)
			assert.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// TestGoToCtyAndCtyToGo 测试 Go 基础类型指针、json.Number、嵌套 map/slice 与 cty.Value 之间的互转。
func TestGoToCtyAndCtyToGo(t *testing.T) {
	bVal := cvt.ValToPtr(true)
	sVal := cvt.ValToPtr("str")
	iVal := cvt.ValToPtr(int64(99))

	// 测试 bool/string/int64 指针转成 cty.Value
	cvBool, err := goToCty(bVal)
	assert.NoError(t, err)
	assert.True(t, cvBool.True())

	cvStr, err := goToCty(sVal)
	assert.NoError(t, err)
	assert.Equal(t, "str", cvStr.AsString())

	cvInt, err := goToCty(iVal)
	assert.NoError(t, err)
	gvInt, err := ctyToGo(cvInt)
	assert.NoError(t, err)
	assert.Equal(t, int64(99), gvInt)

	// 整数格式的 json.Number 转回 Go 类型时还原成 int64
	cvJNum, err := goToCty(json.Number("213"))
	assert.NoError(t, err)
	gvJNum, err := ctyToGo(cvJNum)
	assert.NoError(t, err)
	assert.Equal(t, int64(213), gvJNum)

	// 测试嵌套 map/slice 的深拷贝和转成 HCL 字面量字符串
	complexGo := map[string]any{
		"arr":     []string{"x", "y"},
		"map_arr": []map[string]any{{"k": 1.5}},
		"str_map": map[string]string{"a": "b"},
		"uints":   []any{int8(1), int16(2), int32(3), uint(4), uint8(5), uint16(6), uint32(7), uint64(8), float32(2.5)},
	}
	cpMap := deepCopyMap(complexGo)
	cvComplex, err := goToCty(cpMap)
	assert.NoError(t, err)
	hclLit, err := ctyValueToHCLString(cvComplex)
	assert.NoError(t, err)
	assert.Contains(t, hclLit, `"arr" = ["x", "y"]`)

	nullLit, err := ctyValueToHCLString(cty.NullVal(cty.String))
	assert.NoError(t, err)
	assert.Equal(t, "null", nullLit)
}
