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
	"math/big"
	"sort"
	"strconv"
	"strings"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/tools/maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// providerKey 用插件名和可选的 alias 标识模块里的一个 provider 配置。
// 比如默认 provider 是 {plugin: "tencentcloud", alias: ""}，带别名的是 {plugin: "tencentcloud", alias: "prod"}。
type providerKey struct {
	plugin string
	alias  string
}

// instanceDesc 描述 module / resource / data 块按 count 或 for_each 展开后的单个实例信息。
//
// 举例：
//   - 普通单实例：instanceDesc{suffix: ""}
//   - count = 2 的第 0 个实例：instanceDesc{suffix: "[0]", isCount: true, countIdx: 0}
//   - for_each = {gz = "127.0.0.1/24"} 的实例：
//     instanceDesc{suffix: "[\"gz\"]", isForEach: true, eachKey: "gz", eachVal: cty.StringVal("127.0.0.1/24")}
type instanceDesc struct {
	suffix    string
	isCount   bool
	countIdx  int
	isForEach bool
	eachKey   string
	eachVal   cty.Value
}

// moduleEvalCtx 保存单个模块实例在阶段一展开和求值期间的变量表和上下文信息。
type moduleEvalCtx struct {
	// modulePrefix 当前模块实例的地址前缀，根模块为 ""，子模块如 "module.network" 或 "module.vpc[\"gz\"].module.sub[0]"
	modulePrefix string
	rawMod       *rawModule
	vars         map[string]*evalValue
	locals       map[string]*evalValue
	// visitingLocals 记录正在求值的 local 名字，用来检测 local.a -> local.b -> local.a 这种循环引用
	visitingLocals map[string]bool
	// visitingNodes 记录正在计算 count / for_each 的节点名，用来检测实例数量互相引用的死循环
	visitingNodes map[string]bool
	providers     map[providerKey]string
	// nodeInstances 记录当前模块里每个资源/数据源/导入节点展开后的实例列表，
	// 比如 "tencentcloud_subnet.sub": [{suffix: "[0]", ...}, {suffix: "[1]", ...}]
	nodeInstances map[string][]instanceDesc
	// childOutputs 子模块实例的 output 求值结果，格式为 childOutputs["network"]["subnet_id"]
	childOutputs map[string]map[string]*evalValue
	hasCount     bool
	countIndex   int
	hasEach      bool
	eachKey      string
	eachVal      cty.Value
}

// withInstanceVar 浅拷贝当前上下文，并填入单个 count 或 for_each 实例的 count.index / each.key / each.value。
func (ctx *moduleEvalCtx) withInstanceVar(inst instanceDesc) *moduleEvalCtx {
	cp := *ctx
	cp.hasCount = inst.isCount
	cp.countIndex = inst.countIdx
	cp.hasEach = inst.isForEach
	cp.eachKey = inst.eachKey
	cp.eachVal = inst.eachVal
	return &cp
}

// expander 负责从根模块 "." 开始递归展开所有子模块和多实例节点，并缓存每个模块目录的语法树避免重复解析。
type expander struct {
	modules     map[string]map[string][]byte
	parsedCache map[string]*rawModule
}

// newExpander 创建模块展开器。
func newExpander(modules map[string]map[string][]byte) *expander {
	return &expander{
		modules:     modules,
		parsedCache: make(map[string]*rawModule, len(modules)),
	}
}

// expandRoot 把外部传入的根模块变量转成 cty.Value，然后从根模块 "." 开始递归展开整个模块树。
func (e *expander) expandRoot(rootVars map[string]any) ([]*Block, error) {
	inVars := make(map[string]*evalValue, len(rootVars))
	for k, v := range rootVars {
		cv, err := goToCty(v)
		if err != nil {
			return nil, fmt.Errorf("invalid root variable %q: %w", k, err)
		}
		inVars[k] = &evalValue{ctyVal: cv}
	}
	blocks, _, err := e.expandModule(
		constant.RootModuleKey,
		"",
		inVars,
		nil,
		0,
		[]string{constant.RootModuleKey},
	)
	return blocks, err
}

// getOrParseModule 按 sourcePath 查找并解析模块文件，如果之前解析过就直接返回缓存结果。
func (e *expander) getOrParseModule(sourcePath string) (*rawModule, string, error) {
	resolvedKey, files, ok := e.findModuleFiles(sourcePath)
	if !ok {
		return nil, "", fmt.Errorf("module source %q not found in ParseReq.Modules", sourcePath)
	}
	if cached, hit := e.parsedCache[resolvedKey]; hit {
		return cached, resolvedKey, nil
	}
	rawMod, err := parseModuleFiles(resolvedKey, files)
	if err != nil {
		return nil, "", err
	}
	e.parsedCache[resolvedKey] = rawMod
	return rawMod, resolvedKey, nil
}

// findModuleFiles 查找模块文件，兼容带 "./" 和不带 "./" 的路径写法（比如 "./modules/vpc" 和 "modules/vpc" 都能匹配上）。
func (e *expander) findModuleFiles(sourcePath string) (string, map[string][]byte, bool) {
	if files, ok := e.modules[sourcePath]; ok {
		return sourcePath, files, true
	}
	trimmed := strings.TrimPrefix(sourcePath, "./")
	if files, ok := e.modules[trimmed]; ok {
		return trimmed, files, true
	}
	withDot := "./" + trimmed
	if files, ok := e.modules[withDot]; ok {
		return withDot, files, true
	}
	return "", nil, false
}

// expandModule 递归展开单个模块实例（根模块或子模块实例）：
//  1. 检查嵌套层数是否超过 MaxModuleDepth
//  2. 绑定当前模块的 variable 和 provider
//  3. 展开当前模块内 import / data / resource 的实例，并把实例下标登记好（方便后面校验引用和展开 [*]）
//  4. 按依赖顺序递归展开子模块 module 块，拿到子模块的 Block 列表和 output 结果
//  5. 计算当前模块的 locals，把 import / data / resource 拆解成 Block，最后算出当前模块的 output 返回给父模块
func (e *expander) expandModule(sourcePath string, modulePrefix string, inputVars map[string]*evalValue,
	inheritedProviders map[providerKey]string, depth int, callStack []string) ([]*Block, map[string]*evalValue, error) {

	if depth > constant.MaxModuleDepth {
		return nil, nil, fmt.Errorf("module recursion depth %d exceeds maximum limit of %d", depth,
			constant.MaxModuleDepth)
	}
	rawMod, _, err := e.getOrParseModule(sourcePath)
	if err != nil {
		return nil, nil, err
	}

	ctx := &moduleEvalCtx{
		modulePrefix:   modulePrefix,
		rawMod:         rawMod,
		locals:         make(map[string]*evalValue),
		visitingLocals: make(map[string]bool),
		visitingNodes:  make(map[string]bool),
		nodeInstances:  make(map[string][]instanceDesc),
		childOutputs:   make(map[string]map[string]*evalValue),
	}

	ctx, err = resolveModuleVariables(inputVars, ctx)
	if err != nil {
		return nil, nil, err
	}

	ctx, err = resolveModuleProviders(inheritedProviders, ctx)
	if err != nil {
		return nil, nil, err
	}

	ctx, err = expandModuleNodeInstances(ctx)
	if err != nil {
		return nil, nil, err
	}

	childBlocks, ctx, err := e.expandChildModules(ctx, depth, callStack)
	if err != nil {
		return nil, nil, err
	}

	ctx, err = evaluateAllLocals(ctx)
	if err != nil {
		return nil, nil, err
	}

	ownBlocks, err := decomposeModuleBlocks(ctx)
	if err != nil {
		return nil, nil, err
	}

	outputs, err := evaluateModuleOutputs(ctx)
	if err != nil {
		return nil, nil, err
	}

	allBlocks := append(childBlocks, ownBlocks...)
	return allBlocks, outputs, nil
}

// resolveModuleVariables 计算模块内所有 variable 的实际值：
// 调用方传了值就用传入的值（可以是静态值，也可以是父模块传进来的资源引用），没传就用 default 默认值；如果既没传值又没有 default 就报错。
func resolveModuleVariables(inputVars map[string]*evalValue, ctx *moduleEvalCtx) (*moduleEvalCtx, error) {
	rawMod := ctx.rawMod
	resolved := make(map[string]*evalValue, len(rawMod.variables))
	varNames := maps.Keys(rawMod.variables)
	sort.Strings(varNames)
	for _, name := range varNames {
		vBlk := rawMod.variables[name]
		if inVal, ok := inputVars[name]; ok {
			resolved[name] = inVal
			continue
		}
		defAttr, hasDefault := vBlk.Body.Attributes["default"]
		if !hasDefault {
			return nil, fmt.Errorf("missing required variable %q in module %q", name, rawMod.modPath)
		}
		defVal, err := evalExpr(defAttr.Expr, ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluate default for variable %q: %w", name, err)
		}
		if defVal.isPending {
			return nil, fmt.Errorf("variable %q default cannot depend on runtime attributes", name)
		}
		resolved[name] = defVal
	}
	ctx.vars = resolved
	return ctx, nil
}

// resolveModuleProviders 合并从父模块继承过来的 provider 和当前模块自己声明的 provider 块，
// 生成 (plugin, alias) 到 plugin_instance_id 的映射表。
func resolveModuleProviders(inherited map[providerKey]string, ctx *moduleEvalCtx) (*moduleEvalCtx, error) {
	providers := maps.Clone(inherited)
	if providers == nil {
		providers = make(map[providerKey]string)
	}
	for _, p := range ctx.rawMod.providers {
		plugin := p.Labels[0]
		instVal, err := evalExpr(p.Body.Attributes["plugin_instance_id"].Expr, ctx)
		if err != nil || instVal.isPending || instVal.ctyVal.Type() != cty.String {
			return nil, fmt.Errorf("provider %q plugin_instance_id must be a static string", plugin)
		}
		instanceID := instVal.ctyVal.AsString()
		if instanceID == "" {
			return nil, fmt.Errorf("provider %q plugin_instance_id cannot be empty", plugin)
		}
		alias := ""
		if aliasAttr, hasAlias := p.Body.Attributes["alias"]; hasAlias {
			aliasVal, err := evalExpr(aliasAttr.Expr, ctx)
			if err != nil || aliasVal.isPending || aliasVal.ctyVal.Type() != cty.String {
				return nil, fmt.Errorf("provider %q alias must be a static string", plugin)
			}
			alias = aliasVal.ctyVal.AsString()
		}
		providers[providerKey{plugin: plugin, alias: alias}] = instanceID
	}
	ctx.providers = providers
	return ctx, nil
}

// expandModuleNodeInstances 算出当前模块里 import、data、resource 各自有哪些实例存入 ctx.nodeInstances。
func expandModuleNodeInstances(ctx *moduleEvalCtx) (*moduleEvalCtx, error) {
	rawMod := ctx.rawMod
	for _, imp := range rawMod.imports {
		typeName, resName, err := parseImportTarget(imp.Body.Attributes["to"].Expr)
		if err != nil {
			return nil, err
		}
		ctx.nodeInstances["import."+typeName+"."+resName] = []instanceDesc{{suffix: ""}}
	}
	for _, d := range rawMod.datas {
		// Labels[0] 是数据源类型、Labels[1] 是名称，比如 data "bkcmdb_biz" "target" 拼出 "data.bkcmdb_biz.target"
		if _, _, err := getOrExpandNodeInstances("data."+d.Labels[0]+"."+d.Labels[1], ctx); err != nil {
			return nil, err
		}
	}
	for _, r := range rawMod.resources {
		// Labels[0] 是资源类型、Labels[1] 是名称，比如 resource "aws_instance" "a" 拼出 "aws_instance.a"
		if _, _, err := getOrExpandNodeInstances(r.Labels[0]+"."+r.Labels[1], ctx); err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

// getOrExpandNodeInstances 按需展开并缓存节点的实例列表，同时用 visitingNodes 检查 count / for_each 之间有没有循环引用。
// 比如 aws_eip.b 写了 count = length(aws_instance.a)，哪怕 b 写在 a 前面，算 b 的时候也会先去把 a 展开。
func getOrExpandNodeInstances(baseKey string, ctx *moduleEvalCtx) ([]instanceDesc, bool, error) {
	if insts, ok := ctx.nodeInstances[baseKey]; ok {
		return insts, true, nil
	}
	blk := findRawNodeBlock(ctx.rawMod, baseKey)
	if blk == nil {
		return nil, false, nil
	}
	if ctx.visitingNodes[baseKey] {
		return nil, false, fmt.Errorf("cycle detected in count/for_each involving %q", baseKey)
	}
	ctx.visitingNodes[baseKey] = true
	insts, err := expandInstances(blk.Body, ctx)
	delete(ctx.visitingNodes, baseKey)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", baseKey, err)
	}
	ctx.nodeInstances[baseKey] = insts
	return insts, true, nil
}

// findRawNodeBlock 按节点名（如 "data.bkcmdb_biz.target" 或 "aws_instance.a"）在当前模块里找到对应的语法块。
func findRawNodeBlock(rawMod *rawModule, baseKey string) *hclsyntax.Block {
	if rawMod == nil {
		return nil
	}
	for _, d := range rawMod.datas {
		if "data."+d.Labels[0]+"."+d.Labels[1] == baseKey {
			return d
		}
	}
	for _, r := range rawMod.resources {
		if r.Labels[0]+"."+r.Labels[1] == baseKey {
			return r
		}
	}
	return nil
}

// expandChildModules 先把同级 module 块按依赖顺序排好，再依次展开每个子模块。
// 比如 module "app" 的入参引用了 module.network.subnet_id，就必须先展开 module "network" 拿到它的 output，再展开 module "app"。
func (e *expander) expandChildModules(ctx *moduleEvalCtx, depth int, callStack []string) (
	[]*Block, *moduleEvalCtx, error) {

	orderedCalls, err := sortModuleCallsTopologically(ctx.rawMod.modules)
	if err != nil {
		return nil, nil, err
	}
	var allChildBlocks []*Block
	for _, mc := range orderedCalls {
		var blocks []*Block
		blocks, ctx, err = e.expandOneModuleCall(mc, ctx, depth, callStack)
		if err != nil {
			return nil, nil, err
		}
		allChildBlocks = append(allChildBlocks, blocks...)
	}
	return allChildBlocks, ctx, nil
}

// expandOneModuleCall 展开单个 module "<name>" {} 调用块（支持在 module 块上配置 count / for_each 和 providers）。
func (e *expander) expandOneModuleCall(mc *hclsyntax.Block, ctx *moduleEvalCtx, depth int, callStack []string) (
	[]*Block, *moduleEvalCtx, error) {

	modName := mc.Labels[0]
	srcVal, err := evalExpr(mc.Body.Attributes["source"].Expr, ctx)
	if err != nil || srcVal.isPending || srcVal.ctyVal.Type() != cty.String {
		return nil, nil, fmt.Errorf("module %q source must be a static string", modName)
	}
	srcPath := srcVal.ctyVal.AsString()
	_, resolvedKey, err := e.getOrParseModule(srcPath)
	if err != nil {
		return nil, nil, err
	}
	for _, prev := range callStack {
		if prev == resolvedKey {
			return nil, nil, fmt.Errorf("module cycle detected: %s -> %s", strings.Join(callStack, " -> "), resolvedKey)
		}
	}

	childProviders, err := buildChildModuleProviders(mc.Body, ctx.providers)
	if err != nil {
		return nil, nil, fmt.Errorf("module %q providers: %w", modName, err)
	}
	insts, err := expandInstances(mc.Body, ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("module %q: %w", modName, err)
	}

	var blocks []*Block
	for _, inst := range insts {
		instCtx := ctx.withInstanceVar(inst)
		childInVars, err := evalChildModuleInputVars(mc.Body, instCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("module %q%s: %w", modName, inst.suffix, err)
		}
		modKey := modName + inst.suffix
		childPrefix := joinModulePrefix(ctx.modulePrefix, "module."+modKey)
		nextStack := append(append([]string{}, callStack...), resolvedKey)
		cBlocks, cOutputs, err := e.expandModule(
			srcPath,
			childPrefix,
			childInVars,
			childProviders,
			depth+1,
			nextStack,
		)
		if err != nil {
			return nil, nil, err
		}
		ctx.childOutputs[modKey] = cOutputs
		blocks = append(blocks, cBlocks...)
	}
	return blocks, ctx, nil
}

// buildChildModuleProviders 构造传给子模块的 provider 映射表：
// 默认自动继承父模块里所有不带 alias 的默认 provider；如果 module 块里写了 providers = { tencentcloud = tencentcloud.prod }，
// 就把父模块的 tencentcloud.prod 映射成子模块对应的 provider。
func buildChildModuleProviders(body *hclsyntax.Body, parentProviders map[providerKey]string) (
	map[providerKey]string, error) {

	childProvs := make(map[providerKey]string)
	for pk, instID := range parentProviders {
		if pk.alias == "" {
			childProvs[pk] = instID
		}
	}
	provAttr, ok := body.Attributes["providers"]
	if !ok {
		return childProvs, nil
	}
	// ObjectConsExpr 对应 providers = { tencentcloud = tencentcloud.prod } 花括号字典写法
	objExpr, ok := provAttr.Expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, fmt.Errorf("providers attribute must be an object map")
	}
	for _, item := range objExpr.Items {
		cPlugin, cAlias, err := parseModuleProviderKeyExpr(item.KeyExpr)
		if err != nil {
			return nil, err
		}
		pPlugin, pAlias, err := parseProviderRefExpr(item.ValueExpr)
		if err != nil {
			return nil, err
		}
		instID, exists := parentProviders[providerKey{plugin: pPlugin, alias: pAlias}]
		if !exists || instID == "" {
			return nil, fmt.Errorf("referenced provider \"%s.%s\" does not exist in parent module", pPlugin, pAlias)
		}
		childProvs[providerKey{plugin: cPlugin, alias: cAlias}] = instID
	}
	return childProvs, nil
}

// parseModuleProviderKeyExpr 解析 providers = { tencentcloud = ... } 或 { tencentcloud.backup = ... } 左边的 key。
func parseModuleProviderKeyExpr(expr hclsyntax.Expression) (string, string, error) {
	// ObjectConsKeyExpr 是花括号字典左边 key 的外层包装节点，解包后拿到里面的 tencentcloud 或 tencentcloud.backup
	if okExpr, ok := expr.(*hclsyntax.ObjectConsKeyExpr); ok {
		expr = okExpr.Wrapped
	}
	return parseProviderRefExpr(expr)
}

// evalChildModuleInputVars 计算 module 块里除了 source/providers/version/count/for_each/depends_on 之外的所有输入参数。
func evalChildModuleInputVars(body *hclsyntax.Body, ctx *moduleEvalCtx) (map[string]*evalValue, error) {
	inVars := make(map[string]*evalValue)
	attrNames := maps.Keys(body.Attributes)
	sort.Strings(attrNames)
	for _, name := range attrNames {
		if name == "source" || name == "providers" || name == "version" || metaArgNames[name] {
			continue
		}
		ev, err := evalExpr(body.Attributes[name].Expr, ctx)
		if err != nil {
			return nil, fmt.Errorf("input %q: %w", name, err)
		}
		inVars[name] = ev
	}
	return inVars, nil
}

// evaluateAllLocals 按名称字典序计算当前模块所有的 locals，并检查是否有循环引用。
func evaluateAllLocals(ctx *moduleEvalCtx) (*moduleEvalCtx, error) {
	localNames := maps.Keys(ctx.rawMod.locals)
	sort.Strings(localNames)
	var err error
	for _, name := range localNames {
		_, ctx, err = getOrEvalLocal(name, ctx)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

// getOrEvalLocal 按需计算并缓存 local.<名称> 的值，同时用 visitingLocals 检查 local 之间有没有循环引用。
// 比如 locals { vpc_id = tencentcloud_vpc.main.id, subnet_vpc = local.vpc_id }，
// 求 local.subnet_vpc 时会先算出 local.vpc_id，最终得到 "tencentcloud_vpc.main.id"。
func getOrEvalLocal(name string, ctx *moduleEvalCtx) (*evalValue, *moduleEvalCtx, error) {
	if val, ok := ctx.locals[name]; ok {
		return val, ctx, nil
	}
	if ctx.rawMod == nil {
		return nil, nil, fmt.Errorf("undefined local value \"local.%s\"", name)
	}
	expr, ok := ctx.rawMod.locals[name]
	if !ok {
		return nil, nil, fmt.Errorf("undefined local value \"local.%s\"", name)
	}
	if ctx.visitingLocals[name] {
		return nil, nil, fmt.Errorf("cycle detected in locals involving \"local.%s\"", name)
	}
	ctx.visitingLocals[name] = true
	val, err := evalExpr(expr, ctx)
	delete(ctx.visitingLocals, name)
	if err != nil {
		return nil, nil, err
	}
	ctx.locals[name] = val
	return val, ctx, nil
}

// decomposeModuleBlocks 把当前模块里所有的 import、data、resource 实例拆解成 Block 节点。
func decomposeModuleBlocks(ctx *moduleEvalCtx) ([]*Block, error) {
	rawMod := ctx.rawMod
	var blocks []*Block
	for _, imp := range rawMod.imports {
		blk, err := decomposeImportBlock(imp, ctx)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, blk)
	}
	for _, d := range rawMod.datas {
		key := "data." + d.Labels[0] + "." + d.Labels[1]
		for _, inst := range ctx.nodeInstances[key] {
			blk, err := decomposeNodeBlockInstance(d, inst, ctx)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, blk)
		}
	}
	for _, r := range rawMod.resources {
		key := r.Labels[0] + "." + r.Labels[1]
		for _, inst := range ctx.nodeInstances[key] {
			blk, err := decomposeNodeBlockInstance(r, inst, ctx)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, blk)
		}
	}
	return blocks, nil
}

// evaluateModuleOutputs 计算当前模块所有的 output 值，供父模块引用。
func evaluateModuleOutputs(ctx *moduleEvalCtx) (map[string]*evalValue, error) {
	rawMod := ctx.rawMod
	outputs := make(map[string]*evalValue, len(rawMod.outputs))
	outNames := maps.Keys(rawMod.outputs)
	sort.Strings(outNames)
	for _, name := range outNames {
		ev, err := evalExpr(rawMod.outputs[name], ctx)
		if err != nil {
			return nil, fmt.Errorf("output %q: %w", name, err)
		}
		outputs[name] = ev
	}
	return outputs, nil
}

// expandInstances 解析块上的 count 或 for_each，生成实例列表：
// 同一个块不能同时写 count 和 for_each；都没写时返回单个默认实例；
// 写了 count = N 返回 [0] .. [N-1]；写了 for_each 按 key 字典序返回 ["key1"] .. ["keyK"]。
func expandInstances(body *hclsyntax.Body, ctx *moduleEvalCtx) ([]instanceDesc, error) {
	countAttr, hasCount := body.Attributes["count"]
	forEachAttr, hasForEach := body.Attributes["for_each"]
	if hasCount && hasForEach {
		return nil, fmt.Errorf("cannot specify both count and for_each on the same block")
	}
	if !hasCount && !hasForEach {
		return []instanceDesc{{suffix: ""}}, nil
	}
	if hasCount {
		return expandCountInstances(countAttr.Expr, ctx)
	}
	return expandForEachInstances(forEachAttr.Expr, ctx)
}

// expandCountInstances 计算 count 表达式并生成 0 .. N-1 的实例列表（count = 0 时返回空切片）。
func expandCountInstances(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]instanceDesc, error) {
	ev, err := evalExpr(expr, ctx)
	if err != nil {
		return nil, fmt.Errorf("evaluate count: %w", err)
	}
	if ev.isPending {
		return nil, fmt.Errorf("count must be statically evaluable in stage 1 and cannot depend on runtime attributes")
	}
	if ev.ctyVal.Type() != cty.Number {
		return nil, fmt.Errorf("count must be a number")
	}
	bf := ev.ctyVal.AsBigFloat()
	i64, acc := bf.Int64()
	if acc != big.Exact || i64 < 0 {
		return nil, fmt.Errorf("count must be a non-negative integer, got %s", bf.Text('f', -1))
	}
	n := int(i64)
	insts := make([]instanceDesc, n)
	for i := 0; i < n; i++ {
		insts[i] = instanceDesc{
			suffix:   fmt.Sprintf("[%d]", i),
			isCount:  true,
			countIdx: i,
		}
	}
	return insts, nil
}

// expandForEachInstances 计算 for_each 表达式，并按 key 字典序生成实例列表。
// 比如 for_each = toset(["sh", "gz"]) 会按顺序生成 ["gz"] 和 ["sh"] 两个实例。
func expandForEachInstances(expr hclsyntax.Expression, ctx *moduleEvalCtx) ([]instanceDesc, error) {
	ev, err := evalExpr(expr, ctx)
	if err != nil {
		return nil, fmt.Errorf("evaluate for_each: %w", err)
	}
	if ev.isPending {
		return nil, fmt.Errorf("for_each must be statically evaluable in stage 1 and cannot depend on runtime attributes")
	}
	kvMap, err := extractForEachEntries(ev.ctyVal)
	if err != nil {
		return nil, err
	}
	keys := maps.Keys(kvMap)
	sort.Strings(keys)
	insts := make([]instanceDesc, 0, len(keys))
	for _, k := range keys {
		insts = append(insts, instanceDesc{
			suffix:    fmt.Sprintf("[%s]", strconv.Quote(k)),
			isForEach: true,
			eachKey:   k,
			eachVal:   kvMap[k],
		})
	}
	return insts, nil
}

// extractForEachEntries 从 for_each 的求值结果（支持字符串集合/列表，或任意类型的 Map/Object）里提取 key -> value 映射。
func extractForEachEntries(val cty.Value) (map[string]cty.Value, error) {
	if !val.IsKnown() || val.IsNull() {
		return nil, fmt.Errorf("for_each value cannot be null or unknown")
	}
	ty := val.Type()
	res := make(map[string]cty.Value)
	if ty.IsSetType() || ty.IsTupleType() || ty.IsListType() {
		it := val.ElementIterator()
		for it.Next() {
			_, elem := it.Element()
			if elem.Type() != cty.String {
				return nil, fmt.Errorf("for_each set/list elements must be strings")
			}
			s := elem.AsString()
			res[s] = elem
		}
		return res, nil
	}
	if ty.IsObjectType() || ty.IsMapType() {
		it := val.ElementIterator()
		for it.Next() {
			k, v := it.Element()
			res[k.AsString()] = v
		}
		return res, nil
	}
	return nil, fmt.Errorf("for_each requires a map, object, or set of strings")
}

// sortModuleCallsTopologically 对同一个模块里的多个 module 调用块做拓扑排序，
// 保证被依赖的子模块（比如 module.network）先展开，引用它 output 的子模块（比如 module.app）后展开。
func sortModuleCallsTopologically(calls []*hclsyntax.Block) ([]*hclsyntax.Block, error) {
	if len(calls) <= 1 {
		return calls, nil
	}
	callMap := make(map[string]*hclsyntax.Block, len(calls))
	depMap := make(map[string][]string, len(calls))
	for _, mc := range calls {
		modName := mc.Labels[0]
		callMap[modName] = mc
		depMap[modName] = extractSiblingModuleRefs(mc.Body)
	}
	// 0: 未访问, 1: 访问中（用来检测循环依赖）, 2: 已完成
	color := make(map[string]int, len(calls))
	var ordered []*hclsyntax.Block
	var visit func(name string) error
	visit = func(name string) error {
		if color[name] == 2 {
			return nil
		}
		if color[name] == 1 {
			return fmt.Errorf("cycle detected between module calls involving \"module.%s\"", name)
		}
		color[name] = 1
		for _, dep := range depMap[name] {
			if _, exists := callMap[dep]; exists {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		color[name] = 2
		ordered = append(ordered, callMap[name])
		return nil
	}
	for _, mc := range calls {
		if err := visit(mc.Labels[0]); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// extractSiblingModuleRefs 扫描 module 块属性里引用的其他同级模块名（即 module.<模块名>.*）。
func extractSiblingModuleRefs(body *hclsyntax.Body) []string {
	var refs []string
	for _, attr := range body.Attributes {
		for _, trav := range attr.Expr.Variables() {
			if len(trav) >= 2 && trav.RootName() == "module" {
				// TraverseAttr 对应 module.<模块名> 里的第二段 .<模块名>
				if step, ok := trav[1].(hcl.TraverseAttr); ok {
					refs = append(refs, step.Name)
				}
			}
		}
	}
	return dedupAndSortStrings(refs)
}
