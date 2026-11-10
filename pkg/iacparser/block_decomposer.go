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

	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// metaArgNames 是 resource / data 块顶层的 HCL 元参数，只用于控制实例展开、选择 provider 和声明依赖，不放进业务 Config。
var metaArgNames = map[string]bool{
	"count":      true,
	"for_each":   true,
	"provider":   true,
	"depends_on": true,
}

// resolveNorthboundTriplet 确定节点调用哪个插件（plugin）和哪个插件实例（plugin_instance_id）。
// 默认从 typeName 的下划线前缀取 plugin（比如 "tencentcloud_subnet" 取 "tencentcloud"）并查找默认 provider；
// 如果块里写了 provider = tencentcloud.prod，就按指定的插件和别名去查对应的 plugin_instance_id。
func resolveNorthboundTriplet(typeName string, providerExpr hclsyntax.Expression, providers map[providerKey]string) (
	string, string, error) {

	plugin, err := extractPluginFromTypeName(typeName)
	if err != nil {
		return "", "", err
	}
	alias := ""
	if providerExpr != nil {
		provPlugin, provAlias, err := parseProviderRefExpr(providerExpr)
		if err != nil {
			return "", "", err
		}
		plugin = provPlugin
		alias = provAlias
	}
	instanceID, ok := providers[providerKey{plugin: plugin, alias: alias}]
	if !ok || instanceID == "" {
		return "", "", fmt.Errorf("missing provider configuration for plugin %q (alias %q)", plugin, alias)
	}
	return plugin, instanceID, nil
}

// parseProviderRefExpr 解析块里的 provider = <plugin> 或 provider = <plugin>.<alias> 引用。
// 比如 tencentcloud.prod 解析成 ("tencentcloud", "prod", nil)。
func parseProviderRefExpr(expr hclsyntax.Expression) (string, string, error) {
	// ScopeTraversalExpr 对应 provider = tencentcloud 或 provider = tencentcloud.prod 这种点分引用写法
	travExpr, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok || len(travExpr.Traversal) == 0 || len(travExpr.Traversal) > 2 {
		return "", "", fmt.Errorf("invalid provider reference expression")
	}
	plugin := travExpr.Traversal.RootName()
	if len(travExpr.Traversal) == 1 {
		return plugin, "", nil
	}
	// TraverseAttr 对应点号后面的属性段（即 tencentcloud.prod 里的 .prod）
	attr, ok := travExpr.Traversal[1].(hcl.TraverseAttr)
	if !ok {
		return "", "", fmt.Errorf("invalid provider alias reference")
	}
	return plugin, attr.Name, nil
}

// parseImportTarget 解析 import 块里的 to = <type_name>.<name> 目标地址，
// 比如 tencentcloud_vpc.main 拆成 ("tencentcloud_vpc", "main")。
func parseImportTarget(toExpr hclsyntax.Expression) (string, string, error) {
	segs, ok, err := tryParseRefChain(toExpr, &moduleEvalCtx{})
	if err != nil {
		return "", "", fmt.Errorf("invalid import 'to' address: %w", err)
	}
	if !ok || len(segs) != 2 {
		return "", "", fmt.Errorf("import 'to' must be <type_name>.<name>")
	}
	if _, err := extractPluginFromTypeName(segs[0].Name); err != nil {
		return "", "", fmt.Errorf("invalid import 'to' type_name: %w", err)
	}
	return segs[0].Name, segs[1].String(), nil
}

// decomposeImportBlock 把 import 块解析成 BlockKindImport 节点。
// 比如 import { to = tencentcloud_vpc.main, id = "vpc-12345" } 会生成 BlockID 为 "import.tencentcloud_vpc.main"、
// Config 为 {"import_id": "vpc-12345"} 的节点。
func decomposeImportBlock(imp *hclsyntax.Block, ctx *moduleEvalCtx) (*Block, error) {
	typeName, resName, err := parseImportTarget(imp.Body.Attributes["to"].Expr)
	if err != nil {
		return nil, err
	}
	var provExpr hclsyntax.Expression
	if provAttr, ok := imp.Body.Attributes["provider"]; ok {
		provExpr = provAttr.Expr
	}
	plugin, instanceID, err := resolveNorthboundTriplet(typeName, provExpr, ctx.providers)
	if err != nil {
		return nil, err
	}
	idVal, err := evalExpr(imp.Body.Attributes["id"].Expr, ctx)
	if err != nil {
		return nil, fmt.Errorf("evaluate import id for %s.%s: %w", typeName, resName, err)
	}
	if idVal.isPending || idVal.ctyVal.Type() != cty.String {
		return nil, fmt.Errorf("import id for %s.%s must evaluate to a static string", typeName, resName)
	}
	var deps []string
	if depAttr, ok := imp.Body.Attributes["depends_on"]; ok {
		explicitDeps, err := evalExplicitDependsOn(depAttr.Expr, ctx)
		if err != nil {
			return nil, err
		}
		deps = append(deps, explicitDeps...)
	}
	localID := fmt.Sprintf("import.%s.%s", typeName, resName)
	return &Block{
		BlockID:          joinModulePrefix(ctx.modulePrefix, localID),
		Kind:             enumor.BlockKindImport,
		Plugin:           plugin,
		PluginInstanceID: instanceID,
		TypeName:         typeName,
		DependsOn:        dedupAndSortStrings(deps),
		Config:           map[string]any{"import_id": idVal.ctyVal.AsString()},
	}, nil
}

// decomposeNodeBlockInstance 把单个 resource / data 块解析成最终的 Block 结构。
// 块里的属性会按是否依赖其他资源分开处理：
//  1. 纯静态属性（字面量、var、local）：直接算出值放进 Config
//  2. 顶层动态属性（引用了其他资源的输出）：把引用路径改写成源头节点的 BlockID，放进 PendingExprs，留到阶段二求值
//  3. 嵌套块/列表里的动态字段：在 Config 里先把外层结构建好、动态位置填 nil 占位，再用 "tags.0.value" 这种点号路径记到 PendingExprs
//  4. lifecycle 和 depends_on：提取到 Block 对应字段，不放进业务 Config
func decomposeNodeBlockInstance(node *hclsyntax.Block, inst instanceDesc, ctx *moduleEvalCtx) (*Block, error) {
	instCtx := ctx.withInstanceVar(inst)
	kind := enumor.BlockKind(node.Type)
	typeName, name := node.Labels[0], node.Labels[1]
	var provExpr hclsyntax.Expression
	if provAttr, ok := node.Body.Attributes["provider"]; ok {
		provExpr = provAttr.Expr
	}
	plugin, instanceID, err := resolveNorthboundTriplet(typeName, provExpr, instCtx.providers)
	if err != nil {
		return nil, err
	}
	localID := buildModuleNodeID(kind, typeName, name, inst.suffix)
	fullBlockID := joinModulePrefix(instCtx.modulePrefix, localID)

	config := make(map[string]any)
	pendingExprs := make(map[string]string)
	var deps []string

	// 如果同模块里有指向当前 resource 的 import 块，自动让 resource 依赖对应的 import 节点
	impLocalID := fmt.Sprintf("import.%s.%s%s", typeName, name, inst.suffix)
	if _, hasImport := instCtx.nodeInstances[impLocalID]; kind == enumor.BlockKindResource && hasImport {
		impID := joinModulePrefix(instCtx.modulePrefix, impLocalID)
		deps = append(deps, impID)
	}

	var attrDeps []string
	config, pendingExprs, attrDeps, err = decomposeBodyAttributes(node.Body, "", config, pendingExprs, instCtx)
	if err != nil {
		return nil, fmt.Errorf("block %s: %w", fullBlockID, err)
	}
	deps = append(deps, attrDeps...)

	var nestedDeps []string
	var lc *Lifecycle
	config, pendingExprs, nestedDeps, lc, err = decomposeNestedBlocks(node.Body, "", config, pendingExprs, instCtx)
	if err != nil {
		return nil, fmt.Errorf("block %s: %w", fullBlockID, err)
	}
	deps = append(deps, nestedDeps...)

	if depAttr, ok := node.Body.Attributes["depends_on"]; ok {
		explicitDeps, err := evalExplicitDependsOn(depAttr.Expr, instCtx)
		if err != nil {
			return nil, fmt.Errorf("block %s depends_on: %w", fullBlockID, err)
		}
		deps = append(deps, explicitDeps...)
	}

	var pendingOut map[string]string
	if len(pendingExprs) > 0 {
		pendingOut = pendingExprs
	}
	return &Block{
		BlockID:          fullBlockID,
		Kind:             kind,
		Plugin:           plugin,
		PluginInstanceID: instanceID,
		TypeName:         typeName,
		DependsOn:        dedupAndSortStrings(deps),
		Config:           config,
		PendingExprs:     pendingOut,
		Lifecycle:        lc,
	}, nil
}

// buildModuleNodeID 拼接模块内的节点名（data 节点带 "data." 前缀，resource 直接是 "<type>.<name><suffix>"）。
func buildModuleNodeID(kind enumor.BlockKind, typeName, name, suffix string) string {
	if kind == enumor.BlockKindData {
		return fmt.Sprintf("data.%s.%s%s", typeName, name, suffix)
	}
	return fmt.Sprintf("%s.%s%s", typeName, name, suffix)
}

// decomposeBodyAttributes 遍历块内的属性：
// 静态属性直接写进 targetMap；动态属性如果是顶层字段（dotPrefix == ""）只记进 pendingExprs，
// 如果是嵌套子块里的字段（比如 dotPrefix == "tags.0"），会在 targetMap 里填 nil 占位并记进 pendingExprs["tags.0.value"]。
func decomposeBodyAttributes(body *hclsyntax.Body, dotPrefix string, targetMap map[string]any,
	pendingExprs map[string]string, ctx *moduleEvalCtx) (map[string]any, map[string]string, []string, error) {

	attrNames := maps.Keys(body.Attributes)
	sort.Strings(attrNames)
	var deps []string
	for _, name := range attrNames {
		if dotPrefix == "" && metaArgNames[name] {
			continue
		}
		attr := body.Attributes[name]
		path := joinDotPath(dotPrefix, name)
		var subDeps []string
		var handled bool
		var err error
		// 先尝试按内联对象 {...} 或对象列表 [{...}] 拆解
		targetMap, pendingExprs, subDeps, handled, err = tryDecomposeStructuredAttr(
			attr.Expr, path, name, targetMap, pendingExprs, ctx,
		)
		if err != nil {
			return nil, nil, nil, err
		}
		// 若 handled == true 说明已按字段拆解完毕，收集依赖后跳过后续处理
		if handled {
			deps = append(deps, subDeps...)
			continue
		}
		// 普通属性（非内联对象或对象列表）直接求值
		ev, err := evalExpr(attr.Expr, ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		if ev.isPending {
			if dotPrefix != "" {
				targetMap[name] = nil
			}
			pendingExprs[path] = ev.exprStr
			deps = append(deps, ev.deps...)
			continue
		}
		goVal, err := ctyToGo(ev.ctyVal)
		if err != nil {
			return nil, nil, nil, err
		}
		targetMap[name] = goVal
	}
	return targetMap, pendingExprs, deps, nil
}

// tryDecomposeStructuredAttr 处理属性值是对象（{...}）或对象列表（[{...}, {...}]）的情况：
// 直接逐字段拆解，静态字段填值进 Config（targetMap），动态字段填 nil 占位并记入 PendingExprs。
//
// 举例：
//
//	属性：network_info = { zone = "ap-guangzhou-3", subnet_id = tencentcloud_subnet.app.id }
//	结果：targetMap["network_info"] = {"zone": "ap-guangzhou-3", "subnet_id": nil}
//	     pendingExprs["network_info.subnet_id"] = "tencentcloud_subnet.app.id"
func tryDecomposeStructuredAttr(expr hclsyntax.Expression, path string, keyName string, targetMap map[string]any,
	pendingExprs map[string]string, ctx *moduleEvalCtx) (map[string]any, map[string]string, []string, bool, error) {

	// 情况 1：属性值是花括号对象 { k = v, ... }（ObjectConsExpr）
	if objExpr, ok := expr.(*hclsyntax.ObjectConsExpr); ok {
		var subMap map[string]any
		var deps []string
		var err error
		subMap, pendingExprs, deps, err = decomposeObjectConsSkeleton(objExpr, path, pendingExprs, ctx)
		if err != nil {
			return nil, nil, nil, false, err
		}
		targetMap[keyName] = subMap
		return targetMap, pendingExprs, deps, true, nil
	}
	// 情况 2：属性值是方括号列表 [...]（TupleConsExpr），且列表里每一项都是花括号对象 [{...}, {...}]
	// （如 tags = [{ key = "k", value = tencentcloud_subnet.app.id }]）
	tupExpr, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok || len(tupExpr.Exprs) == 0 {
		return targetMap, pendingExprs, nil, false, nil
	}
	for _, elem := range tupExpr.Exprs {
		// 检查列表元素是否都是花括号对象 {...}（ObjectConsExpr）
		if _, isObj := elem.(*hclsyntax.ObjectConsExpr); !isObj {
			return targetMap, pendingExprs, nil, false, nil
		}
	}
	sliceVal := make([]any, 0, len(tupExpr.Exprs))
	var deps []string
	for i, elem := range tupExpr.Exprs {
		elemPath := fmt.Sprintf("%s.%d", path, i)
		var subMap map[string]any
		var subDeps []string
		var err error
		subMap, pendingExprs, subDeps, err = decomposeObjectConsSkeleton(
			elem.(*hclsyntax.ObjectConsExpr), elemPath, pendingExprs, ctx,
		)
		if err != nil {
			return nil, nil, nil, false, err
		}
		sliceVal = append(sliceVal, subMap)
		deps = append(deps, subDeps...)
	}
	targetMap[keyName] = sliceVal
	return targetMap, pendingExprs, deps, true, nil
}

// decomposeObjectConsSkeleton 拆解单个内联对象表达式：静态字段直接放进 map，动态字段填 nil 并把点号路径记进 pendingExprs。
func decomposeObjectConsSkeleton(objExpr *hclsyntax.ObjectConsExpr, prefix string, pendingExprs map[string]string,
	ctx *moduleEvalCtx) (map[string]any, map[string]string, []string, error) {

	subMap := make(map[string]any, len(objExpr.Items))
	var deps []string
	for _, item := range objExpr.Items {
		k, err := evalObjectItemKey(item.KeyExpr, ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		leafPath := joinDotPath(prefix, k)
		ev, err := evalExpr(item.ValueExpr, ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		if ev.isPending {
			subMap[k] = nil
			pendingExprs[leafPath] = ev.exprStr
			deps = append(deps, ev.deps...)
			continue
		}
		gv, err := ctyToGo(ev.ctyVal)
		if err != nil {
			return nil, nil, nil, err
		}
		subMap[k] = gv
	}
	return subMap, pendingExprs, deps, nil
}

// decomposeNestedBlocks 解析块内的嵌套子块：
// 顶层 lifecycle {} 单独提取到 Block.Lifecycle；其余业务子块（比如多个同名 tags {}）按块名合并成 []any 切片放进 targetMap，
// 子块内部路径按 "<块名>.<下标>"（如 "tags.0"、"tags.1"）递归处理。
//
// 举例：
//
//	HCL 写法：
//	  tags { key = "subnet_id", value = tencentcloud_subnet.app.id }
//	  tags { key = "env",       value = "prod" }
//	  lifecycle { prevent_destroy = true }
//	转换结果：
//	  targetMap["tags"] = [{"key": "subnet_id", "value": nil}, {"key": "env", "value": "prod"}]
//	  pendingExprs["tags.0.value"] = "tencentcloud_subnet.app.id"
//	  lc = &Lifecycle{PreventDestroy: true}（不放进 targetMap）
func decomposeNestedBlocks(body *hclsyntax.Body, dotPrefix string, targetMap map[string]any,
	pendingExprs map[string]string, ctx *moduleEvalCtx) (map[string]any, map[string]string,
	[]string, *Lifecycle, error) {

	var deps []string
	var lc *Lifecycle
	var blockOrder []string
	grouped := make(map[string][]*hclsyntax.Block)

	// 1. 遍历所有子块：把顶层 lifecycle {} 单独拎出来解析；
	// 其余业务子块按块名（如 "tags"、"disk"）归组到 grouped 中，并用 blockOrder 记录首次出现的顺序（避免直接遍历 map 顺序随机）。
	for _, blk := range body.Blocks {
		if dotPrefix == "" && blk.Type == "lifecycle" {
			parsedLC, lcDeps, err := parseLifecycleBlock(blk, ctx)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			lc = parsedLC
			deps = append(deps, lcDeps...)
			continue
		}
		if _, exists := grouped[blk.Type]; !exists {
			blockOrder = append(blockOrder, blk.Type)
		}
		grouped[blk.Type] = append(grouped[blk.Type], blk)
	}

	// 2. 按组遍历同名子块：将第 idx 个子块的路径前缀拼成 "<块名>.<idx>"（如 "tags.0"、"tags.1"），
	// 依次解析子块内的属性和更深层嵌套子块，最后组装成 []any 切片写入 targetMap[blkType]。
	for _, blkType := range blockOrder {
		blks := grouped[blkType]
		sliceOut := make([]any, 0, len(blks))
		for idx, subBlk := range blks {
			subPrefix := fmt.Sprintf("%s.%d", joinDotPath(dotPrefix, blkType), idx)
			subMap := make(map[string]any)
			var aDeps, nDeps []string
			var err error
			subMap, pendingExprs, aDeps, err = decomposeBodyAttributes(subBlk.Body, subPrefix, subMap, pendingExprs,
				ctx)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			deps = append(deps, aDeps...)
			subMap, pendingExprs, nDeps, _, err = decomposeNestedBlocks(subBlk.Body, subPrefix, subMap, pendingExprs,
				ctx)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			deps = append(deps, nDeps...)
			sliceOut = append(sliceOut, subMap)
		}
		targetMap[blkType] = sliceOut
	}
	return targetMap, pendingExprs, deps, lc, nil
}

// parseLifecycleBlock 解析 lifecycle {} 子块。
func parseLifecycleBlock(blk *hclsyntax.Block, ctx *moduleEvalCtx) (*Lifecycle, []string, error) {
	if len(blk.Labels) != 0 || len(blk.Body.Blocks) != 0 {
		return nil, nil, fmt.Errorf("lifecycle block must not contain labels or nested blocks")
	}
	lc := cvt.ValToPtr(Lifecycle{})
	var deps []string
	attrNames := maps.Keys(blk.Body.Attributes)
	sort.Strings(attrNames)
	var err error
	for _, name := range attrNames {
		attr := blk.Body.Attributes[name]
		lc, deps, err = applyLifecycleAttr(name, attr.Expr, lc, deps, ctx)
		if err != nil {
			return nil, nil, err
		}
	}
	return lc, deps, nil
}

// applyLifecycleAttr 解析单个 lifecycle 字段并填进 Lifecycle 结构体。
func applyLifecycleAttr(name string, expr hclsyntax.Expression, lc *Lifecycle, deps []string, ctx *moduleEvalCtx) (
	*Lifecycle, []string, error) {

	switch name {
	case "create_before_destroy":
		b, err := evalStaticBool(expr, ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("lifecycle.create_before_destroy: %w", err)
		}
		lc.CreateBeforeDestroy = b
		return lc, deps, nil
	case "prevent_destroy":
		b, err := evalStaticBool(expr, ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("lifecycle.prevent_destroy: %w", err)
		}
		lc.PreventDestroy = b
		return lc, deps, nil
	case "ignore_changes":
		list, err := parseIgnoreChanges(expr, ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("lifecycle.ignore_changes: %w", err)
		}
		lc.IgnoreChanges = list
		return lc, deps, nil
	case "replace_triggered_by":
		triggers, trigDeps, err := parseReplaceTriggeredBy(expr, ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("lifecycle.replace_triggered_by: %w", err)
		}
		lc.ReplaceTriggeredBy = triggers
		deps = append(deps, trigDeps...)
		return lc, deps, nil
	case "recompute":
		list, err := evalStaticStringSlice(expr, ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("lifecycle.recompute: %w", err)
		}
		lc.Recompute = list
		return lc, deps, nil
	default:
		return nil, nil, fmt.Errorf("unsupported lifecycle attribute %q", name)
	}
}

// evalStaticBool 计算并检查表达式是否为静态布尔值。
func evalStaticBool(expr hclsyntax.Expression, ctx *moduleEvalCtx) (bool, error) {
	ev, err := evalExpr(expr, ctx)
	if err != nil || ev.isPending || ev.ctyVal.Type() != cty.Bool {
		return false, fmt.Errorf("must evaluate to a static boolean")
	}
	return ev.ctyVal.True(), nil
}

// evalStaticStringSlice 计算并检查表达式是否为静态字符串列表（如 ["ipv6_cidr"]）。
func evalStaticStringSlice(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]string, error) {
	ev, err := evalExpr(expr, ctx)
	if err != nil || ev.isPending {
		return nil, fmt.Errorf("must evaluate to a static list of strings")
	}
	if !ev.ctyVal.Type().IsTupleType() && !ev.ctyVal.Type().IsListType() {
		return nil, fmt.Errorf("expected list or tuple of strings")
	}
	var out []string
	it := ev.ctyVal.ElementIterator()
	for it.Next() {
		_, elem := it.Element()
		if elem.Type() != cty.String {
			return nil, fmt.Errorf("all elements must be strings")
		}
		out = append(out, elem.AsString())
	}
	return out, nil
}

// parseIgnoreChanges 解析 lifecycle.ignore_changes，支持写 all 或属性列表（如 [tags["env"], "memo"] -> ["tags.env", "memo"]）。
func parseIgnoreChanges(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]string, error) {
	// ScopeTraversalExpr 对应不加引号的裸关键字写法 ignore_changes = all
	st, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if ok && len(st.Traversal) == 1 && st.Traversal.RootName() == "all" {
		return []string{"all"}, nil
	}
	// TupleConsExpr 对应方括号列表写法 ignore_changes = [...]
	tup, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("ignore_changes must be 'all' or a list of attribute paths")
	}
	out := make([]string, 0, len(tup.Exprs))
	for _, item := range tup.Exprs {
		norm, err := normalizeIgnoreChangeItem(item, ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, norm)
	}
	return out, nil
}

// normalizeIgnoreChangeItem 把 ignore_changes 里的每一项统一转成点号分隔路径，比如 tags["env"] 转成 "tags.env"。
func normalizeIgnoreChangeItem(item hclsyntax.Expression, ctx *moduleEvalCtx) (string, error) {
	segs, ok, err := tryParseRefChain(item, ctx)
	if err == nil && ok && len(segs) > 0 {
		var parts []string
		for _, s := range segs {
			parts = append(parts, s.Name)
			for _, idx := range s.Indices {
				if idx.IsInt {
					parts = append(parts, strconv.Itoa(idx.IntVal))
					continue
				}
				parts = append(parts, idx.StrVal)
			}
		}
		return strings.Join(parts, "."), nil
	}
	ev, err := evalExpr(item, ctx)
	if err != nil || ev.isPending || ev.ctyVal.Type() != cty.String {
		return "", fmt.Errorf("invalid ignore_changes item")
	}
	return ev.ctyVal.AsString(), nil
}

// parseReplaceTriggeredBy 解析 lifecycle.replace_triggered_by，把里面的引用改写成完整节点路径，同时把依赖的节点 ID 收集起来。
func parseReplaceTriggeredBy(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]string, []string, error) {
	// TupleConsExpr 对应方括号列表写法 replace_triggered_by = [...]
	tup, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, nil, fmt.Errorf("replace_triggered_by must be a list of resource references")
	}
	var triggers []string
	var deps []string
	for _, item := range tup.Exprs {
		ev, err := evalExpr(item, ctx)
		if err != nil {
			return nil, nil, err
		}
		if !ev.isPending || len(ev.deps) == 0 {
			return nil, nil, fmt.Errorf("replace_triggered_by item must reference a resource or resource attribute")
		}
		triggers = append(triggers, ev.exprStr)
		deps = append(deps, ev.deps...)
	}
	return triggers, deps, nil
}

// evalExplicitDependsOn 解析 depends_on = [...] 列表，转成依赖的目标节点 BlockID 列表。
// 如果写的是不带下标的多实例资源名（比如 tencentcloud_subnet.sub），会自动展开成依赖它的所有实例（如 [0]、[1]）。
func evalExplicitDependsOn(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]string, error) {
	// TupleConsExpr 对应方括号列表写法 depends_on = [...]
	tup, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("depends_on must be a list of references")
	}
	var deps []string
	for _, item := range tup.Exprs {
		segs, ok, err := tryParseRefChain(item, ctx)
		if err != nil || !ok || len(segs) == 0 {
			return nil, fmt.Errorf("invalid depends_on reference")
		}
		itemDeps, err := resolveDependsOnTarget(segs, ctx)
		if err != nil {
			return nil, err
		}
		deps = append(deps, itemDeps...)
	}
	return deps, nil
}

// resolveDependsOnTarget 解析单个 depends_on 项指向的源头节点 BlockID 列表。
func resolveDependsOnTarget(segs []addrSegment, ctx *moduleEvalCtx) ([]string, error) {
	baseKey := formatSegments(segs)
	// 1. 当 depends_on 写的是不带下标的资源或数据源名（如 tencentcloud_subnet.sub）时，
	// 直接从 nodeInstances 展开返回它的全部实例 BlockID（如 tencentcloud_subnet.sub[0]、tencentcloud_subnet.sub[1]）。
	if insts, ok := ctx.nodeInstances[baseKey]; ok {
		var out []string
		for _, inst := range insts {
			out = append(out, joinModulePrefix(ctx.modulePrefix, baseKey+inst.suffix))
		}
		return out, nil
	}
	// 2. 处理指定了具体下标（如 tencentcloud_subnet.sub[0]）、跨模块 output 引用（如 module.network.subnet_id）
	// 或纯 import 目标（如 import.tencentcloud_vpc.main）的场景。
	// 这里返回 ev.deps 而不是 ev.exprStr，是因为 ev.exprStr 可能带有 .id 等属性后缀或完整表达式，
	// 而 ev.deps 存放的正是剥离了属性名之后的纯源头节点 BlockID 列表。
	ev, err := resolveRefChain(segs, ctx)
	if err != nil {
		return nil, err
	}
	if len(ev.deps) == 0 {
		return nil, fmt.Errorf("depends_on target %q does not reference any resource or data block", baseKey)
	}
	return ev.deps, nil
}

// joinDotPath 用点号拼接路径前缀和字段名，比如 joinDotPath("tags.0", "value") 返回 "tags.0.value"。
func joinDotPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
