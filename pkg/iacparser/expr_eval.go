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

	"hcm/pkg/tools/maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

// evalValue 表示阶段一对一个 HCL 表达式求值后的结果，有两种情况：
//  1. 静态值（isPending == false）：阶段一就能直接算出具体值，结果存在 ctyVal 里。
//     比如 "127.0.0.1/16" 或 var.bk_biz_id（值为 213）。
//  2. 动态引用（isPending == true）：表达式直接或间接引用了其他资源的输出属性，需要留到阶段二再算。
//     比如在子模块 module.app 里引用 var.subnet_id（由父模块传入 module.network.subnet_id，底层指向 tencentcloud_subnet.app.id），
//     此时 exprStr 为 "module.network.tencentcloud_subnet.app.id"，deps 为 ["module.network.tencentcloud_subnet.app"]。
type evalValue struct {
	isPending bool
	ctyVal    cty.Value
	exprStr   string
	deps      []string
}

// evalExpr 在阶段一计算 HCL 表达式：
// 如果表达式只涉及字面量、静态 var/local、count.index、each.key 或内置函数，就直接算出 cty.Value；
// 如果表达式直接或通过 var/local/output 间接引用了其他资源的属性，就把其中已知的静态部分替换成字面量，
// 把引用路径改写成源头节点的完整 BlockID，并收集依赖的节点 ID。
//
// 举例：
//  1. 纯静态：upper(var.env)（当 var.env = "prod"） -> 直接算出 "PROD"
//  2. 跨模块引用：var.subnet_id（最终指向 module.network 下的 tencentcloud_subnet.app.id）
//     -> 改写为 "module.network.tencentcloud_subnet.app.id"，依赖为 ["module.network.tencentcloud_subnet.app"]
//  3. 静态与动态混合："${var.env}-${tencentcloud_vpc.main.id}"（当 var.env = "prod"）
//     -> 改写为 "\"prod-${tencentcloud_vpc.main.id}\""，依赖为 ["tencentcloud_vpc.main"]
func evalExpr(expr hclsyntax.Expression, ctx *moduleEvalCtx) (*evalValue, error) {
	// 先看是不是变量或资源引用（如 var.env、local.x、tencentcloud_vpc.main.id），不是再走下面的分支
	chainSegs, ok, err := tryParseRefChain(expr, ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		return resolveRefChain(chainSegs, ctx)
	}

	switch e := expr.(type) {
	// 常量，如 "127.0.0.1/16"、213、true
	case *hclsyntax.LiteralValueExpr:
		return &evalValue{ctyVal: e.Val}, nil
	// 单个插值 "${var.env}"，直接算里面的 var.env
	case *hclsyntax.TemplateWrapExpr:
		return evalExpr(e.Wrapped, ctx)
	// 圆括号 (a + b)
	case *hclsyntax.ParenthesesExpr:
		inner, err := evalExpr(e.Expression, ctx)
		if err != nil || !inner.isPending {
			return inner, err
		}
		return &evalValue{isPending: true, exprStr: "(" + inner.exprStr + ")", deps: inner.deps}, nil
	// 字符串拼接，如 "${var.env}-${tencentcloud_vpc.main.id}"
	case *hclsyntax.TemplateExpr:
		return evalTemplateExpr(e, ctx)
	// 函数调用，如 join("-", ["a", "b"])、upper(var.env)
	case *hclsyntax.FunctionCallExpr:
		return evalFunctionCallExpr(e, ctx)
	// 列表，如 ["a", "b"]
	case *hclsyntax.TupleConsExpr:
		return evalTupleConsExpr(e, ctx)
	// 对象，如 { env = "prod", vpc_id = tencentcloud_vpc.main.id }
	case *hclsyntax.ObjectConsExpr:
		return evalObjectConsExpr(e, ctx)
	// [*] 表达式，如 tencentcloud_subnet.sub[*].id
	case *hclsyntax.SplatExpr:
		return evalSplatExpr(e, ctx)
	// 二元运算，如 count.index + 1、a == b
	case *hclsyntax.BinaryOpExpr:
		return evalBinaryOpExpr(e, ctx)
	// 一元运算，如 -5、!var.disabled
	case *hclsyntax.UnaryOpExpr:
		return evalUnaryOpExpr(e, ctx)
	// 三元运算，如 var.enable ? "yes" : "no"
	case *hclsyntax.ConditionalExpr:
		return evalConditionalExpr(e, ctx)
	// 普通下标访问，如 ["sh", "gz"][0]
	case *hclsyntax.IndexExpr:
		return evalGeneralIndexExpr(e, ctx)
	default:
		return nil, fmt.Errorf("unsupported HCL expression type %T", expr)
	}
}

// tryParseRefChain 尝试把链式属性访问表达式（含动态下标如 [count.index] 或 [each.key]）展平为 []addrSegment。
//
// 举例：
//
//	输入表达式：tencentcloud_subnet.sub[count.index].id（当前实例 count.index == 1）
//	先算出下标 count.index 为 1，再拼成 [tencentcloud_subnet, sub[1], id]。
func tryParseRefChain(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]addrSegment, bool, error) {
	switch e := expr.(type) {
	// tencentcloud_subnet.sub[*].id 里代表 [*] 本身的节点
	case *hclsyntax.AnonSymbolExpr:
		return nil, true, nil
	// 不带变量下标的普通路径（如 tencentcloud_subnet.sub[0].id）；
	// 或带变量下标时最左边的部分（如 tencentcloud_subnet.sub[count.index].id 里的 tencentcloud_subnet.sub）
	case *hclsyntax.ScopeTraversalExpr:
		segs, err := traversalToSegments(e.Traversal)
		return segs, true, err
	// 方括号里是变量（如 tencentcloud_subnet.sub[count.index]）：
	// e.Collection 是方括号左边的 tencentcloud_subnet.sub，e.Key 是方括号里的 count.index
	case *hclsyntax.IndexExpr:
		baseSegs, ok, err := tryParseRefChain(e.Collection, ctx)
		if err != nil || !ok || len(baseSegs) == 0 {
			return nil, false, err
		}
		keyVal, err := evalExpr(e.Key, ctx)
		if err != nil {
			return nil, false, err
		}
		if keyVal.isPending {
			return nil, false, fmt.Errorf("index expression cannot depend on runtime pending attributes")
		}
		lastIdx := len(baseSegs) - 1
		indexedSeg, err := attachCtyIndexToSegment(baseSegs[lastIdx], keyVal.ctyVal)
		if err != nil {
			return nil, false, err
		}
		baseSegs[lastIdx] = indexedSeg
		return baseSegs, true, nil
	// 变量方括号后面又点了属性（如 tencentcloud_subnet.sub[count.index].id）：
	// e.Source 是点号左边的 tencentcloud_subnet.sub[count.index]，e.Traversal 是点号右边的 .id
	case *hclsyntax.RelativeTraversalExpr:
		baseSegs, ok, err := tryParseRefChain(e.Source, ctx)
		if err != nil || !ok {
			return nil, false, err
		}
		relSegs, err := traversalToSegments(e.Traversal)
		if err != nil {
			return nil, false, err
		}
		return append(baseSegs, relSegs...), true, nil
	default:
		return nil, false, nil
	}
}

// resolveRefChain 根据引用链的第一段名字（var、local、count、each、module、data、import 或资源类型名）分发解析。
func resolveRefChain(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) == 0 {
		return nil, fmt.Errorf("empty reference chain")
	}
	root := segs[0].Name
	switch root {
	case "var":
		return resolveVarRef(segs, ctx)
	case "local":
		return resolveLocalRef(segs, ctx)
	case "count":
		return resolveCountRef(segs, ctx)
	case "each":
		return resolveEachRef(segs, ctx)
	case "module":
		return resolveModuleOutputRef(segs, ctx)
	case "data", "import":
		return resolveDataOrImportRef(segs, ctx)
	default:
		if strings.Contains(root, "_") {
			return resolveResourceRef(segs, ctx)
		}
		return nil, fmt.Errorf("undefined reference %q", formatSegments(segs))
	}
}

// resolveVarRef 解析 var.<变量名> 引用，支持继续访问变量内部的子属性或下标（比如 var.config.port）。
func resolveVarRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) < 2 || segs[0].HasIndex() {
		return nil, fmt.Errorf("invalid var reference %q", formatSegments(segs))
	}
	varName := segs[1].Name
	val, ok := ctx.vars[varName]
	if !ok {
		return nil, fmt.Errorf("undefined variable \"var.%s\"", varName)
	}
	return applyRemainingSegments(val, segs[1], segs[2:])
}

// resolveLocalRef 解析 local.<名称> 引用。
func resolveLocalRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) < 2 || segs[0].HasIndex() {
		return nil, fmt.Errorf("invalid local reference %q", formatSegments(segs))
	}
	localName := segs[1].Name
	var val *evalValue
	var err error
	val, ctx, err = getOrEvalLocal(localName, ctx)
	if err != nil {
		return nil, err
	}
	return applyRemainingSegments(val, segs[1], segs[2:])
}

// resolveCountRef 解析 count.index，返回当前实例的整数下标。
func resolveCountRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) != 2 || segs[1].Name != "index" || !ctx.hasCount {
		return nil, fmt.Errorf("count.index is not available in this context")
	}
	return &evalValue{ctyVal: cty.NumberIntVal(int64(ctx.countIndex))}, nil
}

// resolveEachRef 解析 for_each 实例里的 each.key 或 each.value（支持继续访问子属性如 each.value.cidr）。
func resolveEachRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) < 2 || !ctx.hasEach {
		return nil, fmt.Errorf("each.%s is not available in this context", formatSegments(segs[1:]))
	}
	if segs[1].Name == "key" && len(segs) == 2 && !segs[1].HasIndex() {
		return &evalValue{ctyVal: cty.StringVal(ctx.eachKey)}, nil
	}
	if segs[1].Name == "value" {
		base := &evalValue{ctyVal: ctx.eachVal}
		return applyRemainingSegments(base, segs[1], segs[2:])
	}
	return nil, fmt.Errorf("invalid each reference %q: must be each.key or each.value", formatSegments(segs))
}

// resolveModuleOutputRef 解析对子模块 output 的引用 module.<模块名>.<output名>。
// 比如子模块 module.network 里定义了 output "subnet_id" { value = tencentcloud_subnet.app.id }，
// 父模块引用 module.network.subnet_id 时，会直接拿到子模块里算好的 "module.network.tencentcloud_subnet.app.id"。
func resolveModuleOutputRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) < 3 || segs[0].HasIndex() {
		return nil, fmt.Errorf("invalid module output reference %q", formatSegments(segs))
	}
	modKey := segs[1].String()
	outName := segs[2].Name
	outs, ok := ctx.childOutputs[modKey]
	if !ok {
		return nil, fmt.Errorf("undefined module reference \"module.%s\"", modKey)
	}
	outVal, ok := outs[outName]
	if !ok {
		return nil, fmt.Errorf("undefined output %q on module %q", outName, modKey)
	}
	return applyRemainingSegments(outVal, segs[2], segs[3:])
}

// resolveDataOrImportRef 解析当前模块内对 data.<type>.<name>.<attr> 或 import.<type>.<name>.<attr> 的引用，
// 并加上当前模块的 modulePrefix 拼成完整 BlockID。
func resolveDataOrImportRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) < 3 {
		return nil, fmt.Errorf("incomplete %s reference %q", segs[0].Name, formatSegments(segs))
	}
	baseKey := segs[0].Name + "." + segs[1].Name + "." + segs[2].Name
	resolvedLocalID, err := lookupModuleNodeInstance(baseKey, segs[2], ctx)
	if err != nil {
		// 按单实例没找到，但节点本身存在且只写到 data.<type>.<name>（没带下标也没点属性），按整个多实例列表展开
		if _, ok := ctx.nodeInstances[baseKey]; ok && len(segs) == 3 && !segs[2].HasIndex() {
			return evalMultiInstanceSlice(baseKey, "", ctx)
		}
		return nil, err
	}
	fullBlockID := joinModulePrefix(ctx.modulePrefix, resolvedLocalID)
	exprStr := fullBlockID + formatAttrSuffix(segs[3:])
	return &evalValue{isPending: true, exprStr: exprStr, deps: []string{fullBlockID}}, nil
}

// resolveResourceRef 解析当前模块内对普通资源 <type>.<name>.<attr> 的引用。
// 比如在子模块 module.network 里写 tencentcloud_vpc.main.id，
// 会改写成 "module.network.tencentcloud_vpc.main.id"，并记录依赖 "module.network.tencentcloud_vpc.main"。
func resolveResourceRef(segs []addrSegment, ctx *moduleEvalCtx) (*evalValue, error) {
	if len(segs) < 2 {
		return nil, fmt.Errorf("incomplete resource reference %q", formatSegments(segs))
	}
	baseKey := segs[0].Name + "." + segs[1].Name
	resolvedLocalID, err := lookupModuleNodeInstance(baseKey, segs[1], ctx)
	if err != nil {
		// 按单实例没找到，但资源本身存在且只写到 <type>.<name>（如 length(aws_instance.a) 里的 aws_instance.a，没带下标也没点属性），按整个多实例列表展开
		if _, ok := ctx.nodeInstances[baseKey]; ok && len(segs) == 2 && !segs[1].HasIndex() {
			return evalMultiInstanceSlice(baseKey, "", ctx)
		}
		return nil, err
	}
	fullBlockID := joinModulePrefix(ctx.modulePrefix, resolvedLocalID)
	exprStr := fullBlockID + formatAttrSuffix(segs[2:])
	return &evalValue{isPending: true, exprStr: exprStr, deps: []string{fullBlockID}}, nil
}

// lookupModuleNodeInstance 检查当前模块里是否存在被引用的节点实例。
// 如果模块里只写了 import { to = tencentcloud_vpc.legacy } 而没有写对应的 resource 块，
// 引用 tencentcloud_vpc.legacy.id 时会自动指向 "import.tencentcloud_vpc.legacy"。
func lookupModuleNodeInstance(baseKey string, nameSeg addrSegment, ctx *moduleEvalCtx) (string, error) {
	wantSuffix := ""
	if nameSeg.HasIndex() {
		wantSuffix = strings.TrimPrefix(nameSeg.String(), nameSeg.Name)
	}
	insts, ok, err := getOrExpandNodeInstances(baseKey, ctx)
	if err != nil {
		return "", err
	}
	if ok {
		for _, inst := range insts {
			if inst.suffix == wantSuffix {
				return baseKey + wantSuffix, nil
			}
		}
		return "", fmt.Errorf("instance %q not found for %q", baseKey+wantSuffix, baseKey)
	}
	if _, hasImport := ctx.nodeInstances["import."+baseKey]; hasImport && wantSuffix == "" {
		return "import." + baseKey, nil
	}
	return "", fmt.Errorf("undefined resource or data reference %q", baseKey+wantSuffix)
}

// applyRemainingSegments 拿到 var、local、each.value 或子模块 output 的值 base 后，接着处理后面剩下的下标和属性。
// 比如算 var.subnets[0].cidr 时，前面只用 "subnets" 查出了列表 base，这里接着处理 [0] 和 .cidr。
func applyRemainingSegments(base *evalValue, headSeg addrSegment, tail []addrSegment) (*evalValue, error) {
	if !headSeg.HasIndex() && len(tail) == 0 {
		return base, nil
	}

	// base 是其他资源的引用（如 "tencentcloud_vpc.main.subnet_list"），阶段一还没真实数据，直接把 [0].cidr 拼到字符串后面
	if base.isPending {
		indexSuffix := strings.TrimPrefix(headSeg.String(), headSeg.Name)
		exprStr := base.exprStr + indexSuffix + formatAttrSuffix(tail)
		return &evalValue{isPending: true, exprStr: exprStr, deps: base.deps}, nil
	}

	// base 是已知的具体值（如 [{"cidr": "127.0.0.1/24"}]），直接按 [0] 和 .cidr 一层层把值取出来
	cur, err := stepCtyIndices(base.ctyVal, headSeg.Name, headSeg.Indices)
	if err != nil {
		return nil, err
	}
	for _, seg := range tail {
		cur, err = stepCtyAttrAndIndices(cur, seg)
		if err != nil {
			return nil, err
		}
	}
	return &evalValue{ctyVal: cur}, nil
}

// stepCtyAttrAndIndices 从 val 里先取字段 seg.Name，再取它后面的下标 seg.Indices。
func stepCtyAttrAndIndices(val cty.Value, seg addrSegment) (cty.Value, error) {
	// TraverseAttr 对应点号属性访问 .<字段名>
	next, diags := (hcl.TraverseAttr{Name: seg.Name}).TraversalStep(val)
	if diags.HasErrors() {
		return cty.NilVal, fmt.Errorf("access attribute %q: %s", seg.Name, diags.Error())
	}
	return stepCtyIndices(next, seg.Name, seg.Indices)
}

// stepCtyIndices 从 val 里依次按方括号下标（如 [0] 或 ["gz"]）取值。
func stepCtyIndices(val cty.Value, name string, indices []segmentIndex) (cty.Value, error) {
	cur := val
	for _, idx := range indices {
		key := cty.StringVal(idx.StrVal)
		if idx.IsInt {
			key = cty.NumberIntVal(int64(idx.IntVal))
		}
		// TraverseIndex 对应方括号下标访问 [0] 或 ["gz"]
		next, diags := (hcl.TraverseIndex{Key: key}).TraversalStep(cur)
		if diags.HasErrors() {
			return cty.NilVal, fmt.Errorf("access index on %q: %s", name, diags.Error())
		}
		cur = next
	}
	return cur, nil
}

// evalSplatExpr 把 tencentcloud_subnet.sub[*].id 这种通配写法展开成所有实例的列表，并把每个实例都加进依赖列表。
//
// 举例：
//
//	假设子模块 module.network 里的 tencentcloud_subnet.sub 配置了 count = 2，
//	引用 tencentcloud_subnet.sub[*].id 会展开成：
//	exprStr: "[module.network.tencentcloud_subnet.sub[0].id, module.network.tencentcloud_subnet.sub[1].id]"
//	deps:    ["module.network.tencentcloud_subnet.sub[0]", "module.network.tencentcloud_subnet.sub[1]"]
func evalSplatExpr(expr *hclsyntax.SplatExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	// expr.Source 是 [*] 左边的部分，如 tencentcloud_subnet.sub[*].id 里的 tencentcloud_subnet.sub
	srcSegs, ok, err := tryParseRefChain(expr.Source, ctx)
	if err != nil {
		return nil, err
	}
	if !ok || len(srcSegs) < 2 {
		return nil, fmt.Errorf("invalid splat source expression")
	}
	// expr.Each 是 [*] 右边的部分，如 tencentcloud_subnet.sub[*].id 里的 .id
	eachSegs, ok, err := tryParseRefChain(expr.Each, ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unsupported splat each expression %T", expr.Each)
	}
	return evalMultiInstanceSlice(formatSegments(srcSegs), formatAttrSuffix(eachSegs), ctx)
}

// evalMultiInstanceSlice 把多实例节点（如 aws_instance.a 或 sub[*].id）展开成实例列表，
// 同时在 ctyVal 里放一个等长的 Tuple（元素填未知值 cty.DynamicVal），这样 length(aws_instance.a) 在阶段一就能直接算出实例个数。
func evalMultiInstanceSlice(baseKey, attrSuffix string, ctx *moduleEvalCtx) (*evalValue, error) {
	insts, exists, err := getOrExpandNodeInstances(baseKey, ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("undefined multi-instance resource or data %q in splat expression", baseKey)
	}
	if len(insts) == 0 {
		return &evalValue{ctyVal: cty.EmptyTupleVal}, nil
	}
	elems := make([]cty.Value, len(insts))
	itemExprs := make([]string, len(insts))
	deps := make([]string, len(insts))
	for i, inst := range insts {
		fullBlockID := joinModulePrefix(ctx.modulePrefix, baseKey+inst.suffix)
		elems[i] = cty.DynamicVal
		deps[i] = fullBlockID
		itemExprs[i] = fullBlockID + attrSuffix
	}
	return &evalValue{
		isPending: true,
		ctyVal:    cty.TupleVal(elems),
		exprStr:   "[" + strings.Join(itemExprs, ", ") + "]",
		deps:      deps,
	}, nil
}

// evalTemplateExpr 处理字符串插值模板，比如 "${var.env}-${tencentcloud_vpc.main.id}"。
// 如果模板里全是静态值，直接拼成最终字符串；如果包含动态引用，就把已知的静态部分先替换进去，生成 "\"prod-${tencentcloud_vpc.main.id}\""。
func evalTemplateExpr(expr *hclsyntax.TemplateExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	parts := make([]*evalValue, 0, len(expr.Parts))
	anyPending := false
	for _, p := range expr.Parts {
		ev, err := evalExpr(p, ctx)
		if err != nil {
			return nil, err
		}
		if ev.isPending {
			anyPending = true
		}
		parts = append(parts, ev)
	}
	if !anyPending {
		var sb strings.Builder
		for _, p := range parts {
			strVal, err := convert.Convert(p.ctyVal, cty.String)
			if err != nil {
				return nil, err
			}
			sb.WriteString(strVal.AsString())
		}
		return &evalValue{ctyVal: cty.StringVal(sb.String())}, nil
	}

	var raw strings.Builder
	var deps []string
	for _, p := range parts {
		if p.isPending {
			raw.WriteString("${")
			raw.WriteString(p.exprStr)
			raw.WriteString("}")
			deps = append(deps, p.deps...)
			continue
		}
		strVal, err := convert.Convert(p.ctyVal, cty.String)
		if err != nil {
			return nil, err
		}
		escaped := strings.TrimSuffix(strings.TrimPrefix(strconv.Quote(strVal.AsString()), "\""), "\"")
		raw.WriteString(escaped)
	}
	return &evalValue{isPending: true, exprStr: "\"" + raw.String() + "\"", deps: deps}, nil
}

// evalFunctionCallExpr 处理内置函数调用。
// 先检查函数名是否在黑名单或未注册；如果所有参数都是静态值就当场算出结果；
// 如果参数里有动态引用（比如 join("-", [var.env, tencentcloud_vpc.main.id])），
// 就把静态参数转成字面量，拼成 "join(\"-\", [\"prod\", tencentcloud_vpc.main.id])" 留到阶段二再算。
func evalFunctionCallExpr(expr *hclsyntax.FunctionCallExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	if forbiddenFunctions[expr.Name] {
		return nil, fmt.Errorf("forbidden host/template function %q is not allowed", expr.Name)
	}
	fn, ok := safeFunctions[expr.Name]
	if !ok {
		return nil, fmt.Errorf("unknown or unsupported function %q", expr.Name)
	}
	evalArgs := make([]*evalValue, 0, len(expr.Args))
	anyPending := false
	for _, argExpr := range expr.Args {
		ev, err := evalExpr(argExpr, ctx)
		if err != nil {
			return nil, err
		}
		if ev.isPending {
			anyPending = true
		}
		evalArgs = append(evalArgs, ev)
	}
	if !anyPending {
		return callStaticFunction(expr, fn, evalArgs)
	}
	// 像 length(aws_instance.a) 这种参数带了外层占位 Tuple 的情况，先试算一次；
	// IsWhollyKnown 为 true 说明算出的结果（包括内部所有子元素）已经完全确定、不含未知的 DynamicVal，阶段一直接返回
	if res, err := callStaticFunction(expr, fn, evalArgs); err == nil && res.ctyVal.IsWhollyKnown() {
		return res, nil
	}
	argStrs := make([]string, 0, len(evalArgs))
	var deps []string
	for i, ev := range evalArgs {
		suffix := ""
		// 最后一个参数如果带了 ...（如 max(list...)），拼字符串时补回 "..."
		if expr.ExpandFinal && i == len(evalArgs)-1 {
			suffix = "..."
		}
		if ev.isPending {
			argStrs = append(argStrs, ev.exprStr+suffix)
			deps = append(deps, ev.deps...)
			continue
		}
		lit, err := ctyValueToHCLString(ev.ctyVal)
		if err != nil {
			return nil, err
		}
		argStrs = append(argStrs, lit+suffix)
	}
	return &evalValue{
		isPending: true,
		exprStr:   fmt.Sprintf("%s(%s)", expr.Name, strings.Join(argStrs, ", ")),
		deps:      deps,
	}, nil
}

// callStaticFunction 在所有参数都是静态值时，处理可变参数展开（...）和类型转换并直接调用函数。
func callStaticFunction(expr *hclsyntax.FunctionCallExpr, fn function.Function, evalArgs []*evalValue) (
	*evalValue, error) {

	var rawArgs []cty.Value
	for i, ev := range evalArgs {
		if ev.ctyVal == cty.NilVal {
			return nil, fmt.Errorf("argument %d has no static cty value", i)
		}
		if expr.ExpandFinal && i == len(evalArgs)-1 {
			it := ev.ctyVal.ElementIterator()
			for it.Next() {
				_, elem := it.Element()
				rawArgs = append(rawArgs, elem)
			}
			continue
		}
		rawArgs = append(rawArgs, ev.ctyVal)
	}
	params := fn.Params()
	varParam := fn.VarParam()
	ctyArgs := make([]cty.Value, len(rawArgs))
	for i, arg := range rawArgs {
		ctyArgs[i] = arg
		var targetType cty.Type
		if i < len(params) {
			targetType = params[i].Type
		}
		if i >= len(params) && varParam != nil {
			targetType = varParam.Type
		}
		if targetType != cty.NilType && targetType != cty.DynamicPseudoType && !arg.IsNull() {
			if convArg, err := convert.Convert(arg, targetType); err == nil {
				ctyArgs[i] = convArg
			}
		}
	}
	res, err := fn.Call(ctyArgs)
	if err != nil {
		return nil, fmt.Errorf("function %q evaluation failed: %w", expr.Name, err)
	}
	return &evalValue{ctyVal: res}, nil
}

// evalTupleConsExpr 处理列表构造表达式 [elem1, elem2, ...]。
// 比如 [var.env, tencentcloud_subnet.app.id] 会改写成 "[\"prod\", tencentcloud_subnet.app.id]"。
func evalTupleConsExpr(expr *hclsyntax.TupleConsExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	items := make([]*evalValue, 0, len(expr.Exprs))
	anyPending := false
	for _, sub := range expr.Exprs {
		ev, err := evalExpr(sub, ctx)
		if err != nil {
			return nil, err
		}
		if ev.isPending {
			anyPending = true
		}
		items = append(items, ev)
	}
	if !anyPending {
		vals := make([]cty.Value, len(items))
		for i, it := range items {
			vals[i] = it.ctyVal
		}
		return &evalValue{ctyVal: cty.TupleVal(vals)}, nil
	}
	parts := make([]string, 0, len(items))
	var deps []string
	for _, it := range items {
		if it.isPending {
			parts = append(parts, it.exprStr)
			deps = append(deps, it.deps...)
			continue
		}
		lit, err := ctyValueToHCLString(it.ctyVal)
		if err != nil {
			return nil, err
		}
		parts = append(parts, lit)
	}
	return &evalValue{isPending: true, exprStr: "[" + strings.Join(parts, ", ") + "]", deps: deps}, nil
}

// evalObjectConsExpr 处理对象构造表达式 { k1 = v1, k2 = v2 }。
func evalObjectConsExpr(expr *hclsyntax.ObjectConsExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	valMap := make(map[string]*evalValue, len(expr.Items))
	anyPending := false
	for _, item := range expr.Items {
		keyStr, err := evalObjectItemKey(item.KeyExpr, ctx)
		if err != nil {
			return nil, err
		}
		ev, err := evalExpr(item.ValueExpr, ctx)
		if err != nil {
			return nil, err
		}
		if ev.isPending {
			anyPending = true
		}
		valMap[keyStr] = ev
	}
	if !anyPending {
		ctyMap := make(map[string]cty.Value, len(valMap))
		for k, v := range valMap {
			ctyMap[k] = v.ctyVal
		}
		return &evalValue{ctyVal: cty.ObjectVal(ctyMap)}, nil
	}
	keys := maps.Keys(valMap)
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	var deps []string
	for _, k := range keys {
		ev := valMap[k]
		if ev.isPending {
			parts = append(parts, fmt.Sprintf("%s = %s", strconv.Quote(k), ev.exprStr))
			deps = append(deps, ev.deps...)
			continue
		}
		lit, err := ctyValueToHCLString(ev.ctyVal)
		if err != nil {
			return nil, err
		}
		parts = append(parts, fmt.Sprintf("%s = %s", strconv.Quote(k), lit))
	}
	return &evalValue{isPending: true, exprStr: "{" + strings.Join(parts, ", ") + "}", deps: deps}, nil
}

// evalObjectItemKey 解析对象字面量里的 key（支持裸标识符 { env = "prod" } 和括号动态 key { (var.k) = "prod" }）。
func evalObjectItemKey(keyExpr hclsyntax.Expression, ctx *moduleEvalCtx) (string, error) {
	// ObjectConsKeyExpr 是花括号对象 { k = v } 左边 key 的外层包装节点
	if okExpr, ok := keyExpr.(*hclsyntax.ObjectConsKeyExpr); ok {
		// key 没套圆括号时（如 { env = "prod" }），ForceNonLiteral 为 false，
		// 里面的 ScopeTraversalExpr 就是裸写的标识符 env，直接取 "env" 当 key；
		// 套了圆括号（如 { (var.k) = "prod" }）再往下算变量的值
		if !okExpr.ForceNonLiteral {
			if st, isTrav := okExpr.Wrapped.(*hclsyntax.ScopeTraversalExpr); isTrav && len(st.Traversal) == 1 {
				return st.Traversal.RootName(), nil
			}
		}
		keyExpr = okExpr.Wrapped
	}
	kv, err := evalExpr(keyExpr, ctx)
	if err != nil || kv.isPending || kv.ctyVal.Type() != cty.String {
		return "", fmt.Errorf("object key must evaluate to a static string")
	}
	return kv.ctyVal.AsString(), nil
}

// evalBinaryOpExpr 处理二元运算表达式（如加减乘除、比较、逻辑与或），比如 count.index + 1。
func evalBinaryOpExpr(expr *hclsyntax.BinaryOpExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	lhs, err := evalExpr(expr.LHS, ctx)
	if err != nil {
		return nil, err
	}
	rhs, err := evalExpr(expr.RHS, ctx)
	if err != nil {
		return nil, err
	}
	if !lhs.isPending && !rhs.isPending {
		res, err := expr.Op.Impl.Call([]cty.Value{lhs.ctyVal, rhs.ctyVal})
		if err != nil {
			return nil, err
		}
		return &evalValue{ctyVal: res}, nil
	}
	lStr, err := evalValToExprStr(lhs)
	if err != nil {
		return nil, err
	}
	rStr, err := evalValToExprStr(rhs)
	if err != nil {
		return nil, err
	}
	opSym := binaryOpSymbol(expr.Op)
	deps := append(append([]string{}, lhs.deps...), rhs.deps...)
	return &evalValue{isPending: true, exprStr: fmt.Sprintf("%s %s %s", lStr, opSym, rStr), deps: deps}, nil
}

// evalUnaryOpExpr 处理一元运算表达式（如取负 -x 或取反 !flag）。
func evalUnaryOpExpr(expr *hclsyntax.UnaryOpExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	val, err := evalExpr(expr.Val, ctx)
	if err != nil {
		return nil, err
	}
	if !val.isPending {
		res, err := expr.Op.Impl.Call([]cty.Value{val.ctyVal})
		if err != nil {
			return nil, err
		}
		return &evalValue{ctyVal: res}, nil
	}
	// HCL 的一元运算符只有取负（-）和取反（!）两种，把符号拼回表达式前面留给阶段二计算
	sym := "-"
	if expr.Op == hclsyntax.OpLogicalNot {
		sym = "!"
	}
	return &evalValue{isPending: true, exprStr: sym + val.exprStr, deps: val.deps}, nil
}

// evalConditionalExpr 处理三元表达式 cond ? trueExpr : falseExpr。
// 如果 cond 在阶段一就能算出 true/false，只计算命中的那个分支（没命中的分支不会产生多余依赖）；
// 如果 cond 本身依赖其他资源属性，就把三个分支都改写并合并依赖。
func evalConditionalExpr(expr *hclsyntax.ConditionalExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	cond, err := evalExpr(expr.Condition, ctx)
	if err != nil {
		return nil, err
	}
	if !cond.isPending && cond.ctyVal.Type() == cty.Bool {
		if cond.ctyVal.True() {
			return evalExpr(expr.TrueResult, ctx)
		}
		return evalExpr(expr.FalseResult, ctx)
	}
	tVal, err := evalExpr(expr.TrueResult, ctx)
	if err != nil {
		return nil, err
	}
	fVal, err := evalExpr(expr.FalseResult, ctx)
	if err != nil {
		return nil, err
	}
	cStr, err := evalValToExprStr(cond)
	if err != nil {
		return nil, err
	}
	tStr, err := evalValToExprStr(tVal)
	if err != nil {
		return nil, err
	}
	fStr, err := evalValToExprStr(fVal)
	if err != nil {
		return nil, err
	}
	var deps []string
	deps = append(deps, cond.deps...)
	deps = append(deps, tVal.deps...)
	deps = append(deps, fVal.deps...)
	return &evalValue{isPending: true, exprStr: fmt.Sprintf("%s ? %s : %s", cStr, tStr, fStr), deps: deps}, nil
}

// evalGeneralIndexExpr 处理非引用链形式的下标访问，比如 ["a", "b"][0]。
func evalGeneralIndexExpr(expr *hclsyntax.IndexExpr, ctx *moduleEvalCtx) (*evalValue, error) {
	coll, err := evalExpr(expr.Collection, ctx)
	if err != nil {
		return nil, err
	}
	key, err := evalExpr(expr.Key, ctx)
	if err != nil {
		return nil, err
	}
	if !coll.isPending && !key.isPending {
		if !coll.ctyVal.HasIndex(key.ctyVal).True() {
			return nil, fmt.Errorf("index not found in collection")
		}
		return &evalValue{ctyVal: coll.ctyVal.Index(key.ctyVal)}, nil
	}
	cStr, err := evalValToExprStr(coll)
	if err != nil {
		return nil, err
	}
	kStr, err := evalValToExprStr(key)
	if err != nil {
		return nil, err
	}
	deps := append(append([]string{}, coll.deps...), key.deps...)
	return &evalValue{isPending: true, exprStr: fmt.Sprintf("%s[%s]", cStr, kStr), deps: deps}, nil
}

// evalValToExprStr 把 evalValue 转成 HCL 表达式字符串（动态引用直接取 exprStr，静态值转成字面量）。
func evalValToExprStr(ev *evalValue) (string, error) {
	if ev.isPending {
		return ev.exprStr, nil
	}
	return ctyValueToHCLString(ev.ctyVal)
}

// binaryOpSymbol 返回二元操作符对应的运算符字符串。
func binaryOpSymbol(op *hclsyntax.Operation) string {
	switch op {
	case hclsyntax.OpAdd:
		return "+"
	case hclsyntax.OpSubtract:
		return "-"
	case hclsyntax.OpMultiply:
		return "*"
	case hclsyntax.OpDivide:
		return "/"
	case hclsyntax.OpModulo:
		return "%"
	case hclsyntax.OpEqual:
		return "=="
	case hclsyntax.OpNotEqual:
		return "!="
	case hclsyntax.OpGreaterThan:
		return ">"
	case hclsyntax.OpGreaterThanOrEqual:
		return ">="
	case hclsyntax.OpLessThan:
		return "<"
	case hclsyntax.OpLessThanOrEqual:
		return "<="
	case hclsyntax.OpLogicalAnd:
		return "&&"
	default:
		return "||"
	}
}
