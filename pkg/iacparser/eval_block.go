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
	"fmt"
	"sort"
	"strconv"
	"strings"

	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// stateTreeNode 用于把平铺的 DependencyStates 拼成多层嵌套树，供 HCL 求值器按点号路径逐层查找。
//
// 为什么需要这棵树：
// 调用方传入的 DependencyStates 是平铺的字典（key 是带点号的字符串，如 "module.network.tencentcloud_subnet.app"），
// 但 HCL 求值器在计算 module.network.tencentcloud_subnet.app.id 时，必须先从顶层变量 "module" 开始，
// 像剥洋葱一样按 "module" -> "network" -> "tencentcloud_subnet" -> "app" -> "id" 一层层往下找。
// 因此需要先用 stateTreeNode 把平铺的 key 按点号拆开，拼成一棵多层嵌套树：
//   - children（map）：存点号属性名（如 .network）或字符串下标（如 vpc["gz"] 里的 "gz"）对应的下一层节点，最终转成 HCL 的对象字典类型 cty.Object；
//   - elemList（slice）：存数字下标（如 sub[0]、sub[1]）对应的下一层节点列表，最终转成 HCL 的数组元组类型 cty.Tuple；
//   - leafVal：走到路径最后一段（如 app）时，存放该节点的运行结果 state（如 {"id": "subnet-999"}）。
//
// 为什么 children 和 elemList 必须分开存：
// 因为 HCL 底层的 cty.Object 只接受字符串 key，不允许用数字 [0] 取下标；只有 cty.Tuple / cty.List 才支持用数字 [0] 取下标。
type stateTreeNode struct {
	leafVal  *cty.Value
	children map[string]*stateTreeNode
	elemList []*stateTreeNode
}

// newStateTreeNode 创建空的 stateTreeNode。
func newStateTreeNode() *stateTreeNode {
	return &stateTreeNode{
		children: make(map[string]*stateTreeNode),
	}
}

// buildDepStateEvalContext 把平铺的 DependencyStates 还原成 HCL 求值需要的嵌套变量树（hcl.EvalContext）。
//
// 转换示意：
//
//	平铺输入 DependencyStates：
//	  "module.network.tencentcloud_subnet.sub[0]" -> {"id": "subnet-0"}
//	  "module.network.tencentcloud_subnet.sub[1]" -> {"id": "subnet-1"}
//	  "data.bkcmdb_biz.target"                    -> {"bk_biz_id": 213}
//
//	还原后的 EvalContext.Variables 嵌套树：
//	  Variables["module"]
//	    └── "network" (cty.Object)
//	          └── "tencentcloud_subnet" (cty.Object)
//	                └── "sub" (cty.Tuple)
//	                      ├── [0]: {"id": "subnet-0"}
//	                      └── [1]: {"id": "subnet-1"}
//	  Variables["data"]
//	    └── "bkcmdb_biz" (cty.Object)
//	          └── "target": {"bk_biz_id": 213}
func buildDepStateEvalContext(depStates map[string]map[string]any) (*hcl.EvalContext, error) {
	root := newStateTreeNode()
	blockIDs := maps.Keys(depStates)
	sort.Strings(blockIDs)

	for _, blockID := range blockIDs {
		// 1. 按点号切分 blockID（如拆成 [module, network, tencentcloud_subnet, sub[0]]），并校验它是纯节点地址（不能多带属性段）
		segs, err := parseAddressString(blockID)
		if err != nil {
			return nil, fmt.Errorf("invalid block_id %q in dependency_states: %w", blockID, err)
		}
		parsed, err := consumePairwiseAddress(segs)
		if err != nil || len(parsed.AttrSegments) != 0 {
			return nil, fmt.Errorf("invalid block_id %q in dependency_states", blockID)
		}
		// 2. 把该节点的运行结果 state map（如 {"id": "subnet-0"}）转成 HCL 认识的 cty.Value
		stateCty, err := goToCty(depStates[blockID])
		if err != nil {
			return nil, fmt.Errorf("convert state for %q: %w", blockID, err)
		}
		// 3. 顺着切出来的地址段 segs 一层层挂进 root 树中
		root = insertIntoStateTree(root, segs, stateCty)
	}

	// 4. 把 root 树第一层的每个分支（如 "module"、"data"、"tencentcloud_vpc"）转成嵌套的 cty.Object / cty.Tuple 放入 Variables
	vars := make(map[string]cty.Value, len(root.children))
	for k, child := range root.children {
		vars[k] = stateTreeToCtyValue(child)
	}
	return &hcl.EvalContext{
		Variables: vars,
		Functions: safeFunctions,
	}, nil
}

// insertIntoStateTree 按地址段把单个节点的 state 插入前缀树中，并返回根节点。
func insertIntoStateTree(root *stateTreeNode, segs []addrSegment, leaf cty.Value) *stateTreeNode {
	cur := root
	for _, seg := range segs {
		next, ok := cur.children[seg.Name]
		if !ok {
			next = newStateTreeNode()
			cur.children[seg.Name] = next
		}
		cur = next
		for _, idx := range seg.Indices {
			if idx.IsInt {
				for len(cur.elemList) <= idx.IntVal {
					cur.elemList = append(cur.elemList, newStateTreeNode())
				}
				cur = cur.elemList[idx.IntVal]
				continue
			}
			strChild, exists := cur.children[idx.StrVal]
			if !exists {
				strChild = newStateTreeNode()
				cur.children[idx.StrVal] = strChild
			}
			cur = strChild
		}
	}
	cur.leafVal = cvt.ValToPtr(leaf)
	return root
}

// stateTreeToCtyValue 把前缀树递归转成 cty.Object 或 cty.Tuple。
func stateTreeToCtyValue(node *stateTreeNode) cty.Value {
	if node.leafVal != nil && len(node.children) == 0 && len(node.elemList) == 0 {
		return cvt.PtrToVal(node.leafVal)
	}
	if len(node.elemList) > 0 {
		elems := make([]cty.Value, len(node.elemList))
		for i, el := range node.elemList {
			elems[i] = stateTreeToCtyValue(el)
		}
		return cty.TupleVal(elems)
	}
	if len(node.children) > 0 {
		obj := make(map[string]cty.Value, len(node.children))
		for k, ch := range node.children {
			obj[k] = stateTreeToCtyValue(ch)
		}
		return cty.ObjectVal(obj)
	}
	return cty.EmptyObjectVal
}

// evalPendingExpr 解析单条待求值表达式，检查函数安全性后在 evalCtx 上算出结果并转回 Go 类型。
func evalPendingExpr(fieldPath string, exprStr string, evalCtx *hcl.EvalContext) (any, error) {
	expr, diags := hclsyntax.ParseExpression([]byte(exprStr), fieldPath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse pending expression %q (%s): %s", fieldPath, exprStr, diags.Error())
	}
	if err := validatePendingExprAST(expr); err != nil {
		return nil, fmt.Errorf("pending expression %q: %w", fieldPath, err)
	}
	ctyVal, evalDiags := expr.Value(evalCtx)
	if evalDiags.HasErrors() {
		return nil, fmt.Errorf("evaluate pending expression %q (%s): %s", fieldPath, exprStr, evalDiags.Error())
	}
	return ctyToGo(ctyVal)
}

// validatePendingExprAST 在求值待决表达式前检查语法树：
// 拦截黑名单函数（如 file、templatefile）和未注册函数，防止调用方绕过阶段一直接传危险表达式。
func validatePendingExprAST(expr hclsyntax.Expression) error {
	var walkErr error
	_ = hclsyntax.VisitAll(expr, func(node hclsyntax.Node) hcl.Diagnostics {
		if walkErr != nil {
			return nil
		}
		// FunctionCallExpr 对应函数调用语法 fn(...)，如 file(...)、join(...)
		if fnCall, ok := node.(*hclsyntax.FunctionCallExpr); ok {
			if forbiddenFunctions[fnCall.Name] {
				walkErr = fmt.Errorf("forbidden host/template function %q is not allowed", fnCall.Name)
				return nil
			}
			if _, allowed := safeFunctions[fnCall.Name]; !allowed {
				walkErr = fmt.Errorf("unknown or unsupported function %q", fnCall.Name)
				return nil
			}
		}
		return nil
	})
	return walkErr
}

// backfillConfigField 把阶段二算出来的值填回 config 对应字段位置并返回更新后的 config：
// 如果 fieldPath 不含点号（如 "subnet_id"），直接写进顶层 config["subnet_id"]；
// 如果含点号（如 "tags.0.value"），按点号逐层往下找（数字当切片下标，字符串当 map key），把叶子位置的 nil 替换成算出来的值。
func backfillConfigField(config map[string]any, fieldPath string, val any) (map[string]any, error) {
	if !strings.Contains(fieldPath, ".") {
		config[fieldPath] = val
		return config, nil
	}
	parts := strings.Split(fieldPath, ".")
	var cur any = config
	for i := 0; i < len(parts)-1; i++ {
		next, err := stepBackfillContainer(cur, parts[i], fieldPath)
		if err != nil {
			return nil, err
		}
		cur = next
	}
	lastPart := parts[len(parts)-1]
	if _, err := setBackfillLeaf(cur, lastPart, val, fieldPath); err != nil {
		return nil, err
	}
	return config, nil
}

// stepBackfillContainer 按点号路径往下走一层容器（数字段按切片下标取，非数字段按 map key 取）。
func stepBackfillContainer(cur any, seg string, fieldPath string) (any, error) {
	if idx, err := strconv.Atoi(seg); err == nil {
		return stepBackfillSlice(cur, idx, fieldPath)
	}
	m, ok := cur.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("dot-path %q expected map at segment %q, got %T", fieldPath, seg, cur)
	}
	next, exists := m[seg]
	if !exists || next == nil {
		return nil, fmt.Errorf("dot-path %q missing container key %q", fieldPath, seg)
	}
	return next, nil
}

// stepBackfillSlice 按整数下标访问切片容器，并检查是否越界。
func stepBackfillSlice(cur any, idx int, fieldPath string) (any, error) {
	switch s := cur.(type) {
	case []any:
		if idx < 0 || idx >= len(s) {
			return nil, fmt.Errorf("dot-path %q index %d out of bounds (len %d)", fieldPath, idx, len(s))
		}
		return s[idx], nil
	case []map[string]any:
		if idx < 0 || idx >= len(s) {
			return nil, fmt.Errorf("dot-path %q index %d out of bounds (len %d)", fieldPath, idx, len(s))
		}
		return s[idx], nil
	default:
		return nil, fmt.Errorf("dot-path %q expected slice at index %d, got %T", fieldPath, idx, cur)
	}
}

// setBackfillLeaf 把算出来的值写进点号路径的最后一段（支持写切片元素或 map 键值），并返回更新后的容器。
func setBackfillLeaf(cur any, lastSeg string, val any, fieldPath string) (any, error) {
	if idx, err := strconv.Atoi(lastSeg); err == nil {
		s, ok := cur.([]any)
		if !ok {
			return nil, fmt.Errorf("dot-path %q expected []any for leaf index %d, got %T", fieldPath, idx, cur)
		}
		if idx < 0 || idx >= len(s) {
			return nil, fmt.Errorf("dot-path %q leaf index %d out of bounds (len %d)", fieldPath, idx, len(s))
		}
		s[idx] = val
		return s, nil
	}
	m, ok := cur.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("dot-path %q expected map[string]any for leaf key %q, got %T", fieldPath, lastSeg, cur)
	}
	m[lastSeg] = val
	return m, nil
}
