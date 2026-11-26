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

// Package iacparser 实现纯内存的 HCL 两阶段解析与单节点求值：
//   - 阶段一（Parse）：解析根模块和子模块的 .tf 文件。其中 variable、locals、provider、module、output
//     只在解析阶段参与传参和计算，不会生成最终节点；只有 resource、data、import 会展开并生成最终的 Block 节点列表。
//   - 阶段二（EvalBlock）：在节点实际执行前，拿到上游依赖节点的运行结果（DependencyStates），
//     把阶段一留下的待求值表达式（PendingExprs）算出来，填回 Config 对应位置。
package iacparser

import (
	"sort"

	"hcm/pkg/tools/maps"
)

// Parser 定义 HCL 两阶段解析器接口。
type Parser interface {
	// Parse 阶段一：把多模块 .tf 文件和外部输入变量解析成按依赖排好序的节点列表。
	Parse(req *ParseReq) (*ParseResult, error)
	// EvalBlock 阶段二：结合上游节点的运行结果，计算当前节点的 PendingExprs 并填回 Config 副本。
	EvalBlock(req *EvalReq) (map[string]any, error)
}

// parserImpl 是 Parser 的无状态实现，可并发安全调用。
type parserImpl struct{}

// NewParser 创建 HCL 解析器实例。
func NewParser() Parser {
	return &parserImpl{}
}

// Parse 执行阶段一解析：从根模块 "." 开始展开所有子模块和多实例，完成引用改写、环路检查和拓扑排序。
func (p *parserImpl) Parse(req *ParseReq) (*ParseResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	exp := newExpander(req.Modules)
	blocks, err := exp.expandRoot(req.Vars)
	if err != nil {
		return nil, err
	}
	sortedBlocks, err := validateAndSortBlocks(blocks)
	if err != nil {
		return nil, err
	}
	return &ParseResult{Blocks: sortedBlocks}, nil
}

// EvalBlock 执行阶段二单节点求值：
// 先把 req.Config 深拷贝一份避免改动调用方入参，然后把上游节点的运行结果（req.DependencyStates）组装成 HCL 变量树，
// 依次算出 req.PendingExprs 里每条表达式的值，最后按顶层字段名或点号路径（如 "tags.0.value"）填回 Config 副本。
//
// 举例：
//
//	输入：
//	  Config: {"instance_name": "web-0", "tags": [{"key": "subnet", "value": nil}]}
//	  PendingExprs: {
//	      "subnet_id":    "module.network.tencentcloud_subnet.app.id",
//	      "tags.0.value": "upper(module.network.tencentcloud_subnet.app.id)",
//	  }
//	  DependencyStates: {"module.network.tencentcloud_subnet.app": {"id": "subnet-abc"}}
//	输出：
//	  {"instance_name": "web-0", "subnet_id": "subnet-abc", "tags": [{"key": "subnet", "value": "SUBNET-ABC"}]}
func (p *parserImpl) EvalBlock(req *EvalReq) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	configCopy := deepCopyMap(req.Config)
	if len(req.PendingExprs) == 0 {
		return configCopy, nil
	}

	evalCtx, err := buildDepStateEvalContext(req.DependencyStates)
	if err != nil {
		return nil, err
	}

	fieldPaths := maps.Keys(req.PendingExprs)
	sort.Strings(fieldPaths)
	for _, fieldPath := range fieldPaths {
		exprStr := req.PendingExprs[fieldPath]
		val, err := evalPendingExpr(fieldPath, exprStr, evalCtx)
		if err != nil {
			return nil, err
		}
		configCopy, err = backfillConfigField(configCopy, fieldPath, val)
		if err != nil {
			return nil, err
		}
	}
	return configCopy, nil
}
