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
	"testing"

	"github.com/stretchr/testify/assert"
)

// buildHostEvalReq 构造 bkcmdb_host 节点的 EvalBlock 测试请求。
func buildHostEvalReq() *EvalReq {
	return &EvalReq{
		Config: map[string]any{
			"operator": "user",
			"tags": []any{
				map[string]any{"key": "subnet_id", "value": nil},
			},
		},
		PendingExprs: map[string]string{
			"bk_host_innerip": "module.app.tencentcloud_instance.web.private_ip",
			"bk_biz_id":       "tostring(data.bkcmdb_biz.target.bk_biz_id)",
			"tags.0.value":    "module.network.tencentcloud_subnet.app.id",
		},
		DependencyStates: map[string]map[string]any{
			"data.bkcmdb_biz.target":                 {"bk_biz_id": float64(213)},
			"module.network.tencentcloud_subnet.app": {"id": "subnet-demo-01"},
			"module.app.tencentcloud_instance.web":   {"id": "ins-001", "private_ip": "127.0.0.1"},
		},
	}
}

// buildMultiInstanceEvalReq 构造带 count/for_each 实例下标引用和嵌套 slice/map 路径回填的 EvalReq。
func buildMultiInstanceEvalReq() *EvalReq {
	return &EvalReq{
		Config: map[string]any{
			"fields": []map[string]any{
				{"key": "vpc_gz", "value": nil},
			},
			"arr_leaf": []any{"placeholder"},
		},
		PendingExprs: map[string]string{
			"subnets":        "[tencentcloud_subnet.sub[0].id, tencentcloud_subnet.sub[1].id]",
			"fields.0.value": `module.vpc["gz"].tencentcloud_vpc.main.id`,
			"biz_num":        "data.bkcmdb_biz.target.bk_biz_id",
			"arr_leaf.0":     "tencentcloud_subnet.sub[0].id",
		},
		DependencyStates: map[string]map[string]any{
			"tencentcloud_subnet.sub[0]":             {"id": "sub-0"},
			"tencentcloud_subnet.sub[1]":             {"id": "sub-1"},
			`module.vpc["gz"].tencentcloud_vpc.main`: {"id": "vpc-gz-001"},
			"data.bkcmdb_biz.target":                 {"bk_biz_id": 213},
		},
	}
}

// TestParser_EvalBlock 测试阶段二根据上游输出计算 PendingExprs、按路径回填 Config，以及不修改入参 req.Config。
func TestParser_EvalBlock(t *testing.T) {
	testCases := []struct {
		name       string
		req        *EvalReq
		wantConfig map[string]any
		verifyFn   func(t *testing.T, req *EvalReq, out map[string]any)
	}{
		{
			name: "eval pending expressions with builtin tostring, dot-path backfill, and deep-copy isolation",
			req:  buildHostEvalReq(),
			wantConfig: map[string]any{
				"bk_host_innerip": "127.0.0.1",
				"bk_biz_id":       "213",
				"operator":        "user",
				"tags": []any{
					map[string]any{"key": "subnet_id", "value": "subnet-demo-01"},
				},
			},
			verifyFn: func(t *testing.T, req *EvalReq, out map[string]any) {
				// 确认 EvalBlock 不会改掉调用方传进来的 req.Config
				origTags := req.Config["tags"].([]any)
				origItem := origTags[0].(map[string]any)
				assert.Nil(t, origItem["value"])
				_, existsInOrig := req.Config["bk_host_innerip"]
				assert.False(t, existsInOrig)
			},
		},
		{
			name: "eval cross-module reference with modules and resources named module",
			req: &EvalReq{
				Config:       map[string]any{"name": "app-subnet"},
				PendingExprs: map[string]string{"vpc_id": "module.network.module.module.tencentcloud_vpc.module.id"},
				DependencyStates: map[string]map[string]any{
					"module.network.module.module.tencentcloud_vpc.module": {"id": "vpc-mod-001"},
				},
			},
			wantConfig: map[string]any{"name": "app-subnet", "vpc_id": "vpc-mod-001"},
		},
		{
			name: "eval multi-instance count and for_each subscripts with slice map backfill",
			req:  buildMultiInstanceEvalReq(),
			wantConfig: map[string]any{
				"subnets":  []any{"sub-0", "sub-1"},
				"biz_num":  int64(213),
				"arr_leaf": []any{"sub-0"},
				"fields": []map[string]any{
					{"key": "vpc_gz", "value": "vpc-gz-001"},
				},
			},
		},
		{
			name:       "empty PendingExprs returns deep-copied Config immediately",
			req:        &EvalReq{Config: map[string]any{"a": "b"}, PendingExprs: nil},
			wantConfig: map[string]any{"a": "b"},
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := parser.EvalBlock(tc.req)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantConfig, out)
			if tc.verifyFn != nil {
				tc.verifyFn(t, tc.req, out)
			}
		})
	}
}

// TestParser_EvalBlock_ValidationAndPathErrors 测试 DependencyStates 缺少依赖节点或属性、回填路径 slice 越界或类型不匹配的报错。
func TestParser_EvalBlock_ValidationAndPathErrors(t *testing.T) {
	testCases := []struct {
		name        string
		req         *EvalReq
		errContains string
	}{
		{name: "nil EvalReq rejected", req: nil, errContains: "nil"},
		{
			name: "missing dependency block_id in DependencyStates",
			req: &EvalReq{
				Config:           map[string]any{"name": "app"},
				PendingExprs:     map[string]string{"vpc_id": "module.network.import.tencentcloud_vpc.existing.id"},
				DependencyStates: map[string]map[string]any{},
			},
			errContains: "module.network.import.tencentcloud_vpc.existing",
		},
		{
			name: "missing attribute in dependency state",
			req: &EvalReq{
				Config:       map[string]any{"name": "app"},
				PendingExprs: map[string]string{"vpc_id": "module.network.import.tencentcloud_vpc.existing.id"},
				DependencyStates: map[string]map[string]any{
					"module.network.import.tencentcloud_vpc.existing": {"name": "only-name-no-id"},
				},
			},
			errContains: "id",
		},
		{
			name: "dot-path slice index out of bounds",
			req: &EvalReq{
				Config:       map[string]any{"tags": []any{map[string]any{"key": "subnet_id", "value": nil}}},
				PendingExprs: map[string]string{"tags.5.value": "module.network.tencentcloud_subnet.app.id"},
				DependencyStates: map[string]map[string]any{
					"module.network.tencentcloud_subnet.app": {"id": "subnet-1"},
				},
			},
			errContains: "out of bounds",
		},
		{
			name: "dot-path container type mismatch",
			req: &EvalReq{
				Config:       map[string]any{"tags": "not-a-slice"},
				PendingExprs: map[string]string{"tags.0.value": "module.network.tencentcloud_subnet.app.id"},
				DependencyStates: map[string]map[string]any{
					"module.network.tencentcloud_subnet.app": {"id": "subnet-1"},
				},
			},
			errContains: "expected slice",
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := parser.EvalBlock(tc.req)
			assert.Error(t, err)
			assert.Nil(t, out)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}

// TestParser_EvalBlock_ForbiddenFunctions 测试 EvalBlock 阶段拦截 file、templatefile 等禁用函数。
func TestParser_EvalBlock_ForbiddenFunctions(t *testing.T) {
	testCases := []struct {
		name        string
		req         *EvalReq
		errContains string
	}{
		{
			name: "forbidden file function call rejected",
			req: &EvalReq{
				Config:           map[string]any{},
				PendingExprs:     map[string]string{"content": `file("config.txt")`},
				DependencyStates: map[string]map[string]any{},
			},
			errContains: "file",
		},
		{
			name: "forbidden templatefile function call rejected",
			req: &EvalReq{
				Config:           map[string]any{},
				PendingExprs:     map[string]string{"content": `templatefile("config.tpl", {})`},
				DependencyStates: map[string]map[string]any{},
			},
			errContains: "templatefile",
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := parser.EvalBlock(tc.req)
			assert.Error(t, err)
			assert.Nil(t, out)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}
