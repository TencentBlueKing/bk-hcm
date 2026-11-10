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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// rawModule 汇总一个模块目录下所有 .tf 文件解析出来的顶层块。
// 其中 variables、locals、providers、modules、outputs 只在解析阶段参与传参和计算，不会生成最终节点；
// 只有 resources、datas、imports 会生成最终的 Block 节点。
type rawModule struct {
	modPath   string
	variables map[string]*hclsyntax.Block
	locals    map[string]hclsyntax.Expression
	outputs   map[string]hclsyntax.Expression
	providers []*hclsyntax.Block
	modules   []*hclsyntax.Block
	resources []*hclsyntax.Block
	datas     []*hclsyntax.Block
	imports   []*hclsyntax.Block
}

// parseModuleFiles 按文件名顺序解析模块目录下的所有 .tf 文件，并把顶层块归类到 rawModule 中。
func parseModuleFiles(modPath string, files map[string][]byte) (*rawModule, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("module %q has no .tf files", modPath)
	}
	fileNames := maps.Keys(files)
	sort.Strings(fileNames)

	rawMod := &rawModule{
		modPath:   modPath,
		variables: make(map[string]*hclsyntax.Block),
		locals:    make(map[string]hclsyntax.Expression),
		outputs:   make(map[string]hclsyntax.Expression),
	}
	for _, fn := range fileNames {
		fullPath := modPath + "/" + fn
		hclFile, diags := hclsyntax.ParseConfig(files[fn], fullPath, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, fmt.Errorf("hcl syntax error in %s: %s", fullPath, diags.Error())
		}
		// *hclsyntax.Body 对应单个 .tf 文件解析出来的顶层 HCL 块与属性容器
		body, ok := hclFile.Body.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("unexpected HCL body type in %s", fullPath)
		}
		var err error
		rawMod, err = collectTopLevelBlocks(rawMod, body, fullPath)
		if err != nil {
			return nil, err
		}
	}
	return rawMod, nil
}

// collectTopLevelBlocks 检查文件顶层有没有散落的属性，并把顶层块逐个分类放进 rawModule。
func collectTopLevelBlocks(rawMod *rawModule, body *hclsyntax.Body, filePath string) (*rawModule, error) {
	// 文件最外层只能写 resource / locals 等块，不能脱离 {} 直接写 key = value 属性
	if len(body.Attributes) > 0 {
		names := maps.Keys(body.Attributes)
		sort.Strings(names)
		return nil, fmt.Errorf("unexpected top-level attribute %q in %s", names[0], filePath)
	}
	var err error
	for _, blk := range body.Blocks {
		rawMod, err = classifyTopLevelBlock(rawMod, blk, filePath)
		if err != nil {
			return nil, err
		}
	}
	return rawMod, nil
}

// classifyTopLevelBlock 按顶层块类型分发处理，遇到不支持的块类型（比如 terraform {}）直接报错。
func classifyTopLevelBlock(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	switch blk.Type {
	case "variable":
		return addRawVariable(rawMod, blk, filePath)
	case "locals":
		return addRawLocals(rawMod, blk, filePath)
	case "provider":
		return addRawProvider(rawMod, blk, filePath)
	case "module":
		return addRawModuleCall(rawMod, blk, filePath)
	case "output":
		return addRawOutput(rawMod, blk, filePath)
	case "resource":
		return addRawResourceOrData(rawMod, blk, enumor.BlockKindResource, filePath)
	case "data":
		return addRawResourceOrData(rawMod, blk, enumor.BlockKindData, filePath)
	case "import":
		return addRawImport(rawMod, blk, filePath)
	default:
		return nil, fmt.Errorf("unsupported top-level block type %q in %s", blk.Type, filePath)
	}
}

// addRawVariable 校验并记录 variable "<name>" {} 块。
func addRawVariable(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	if len(blk.Labels) != 1 || blk.Labels[0] == "" {
		return nil, fmt.Errorf("variable block in %s must have exactly 1 label", filePath)
	}
	rawMod.variables[blk.Labels[0]] = blk
	return rawMod, nil
}

// addRawLocals 校验并提取 locals { k1 = expr1, k2 = expr2 } 块里的属性表达式。
func addRawLocals(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	if len(blk.Labels) != 0 {
		return nil, fmt.Errorf("locals block in %s must not have labels", filePath)
	}
	for name, attr := range blk.Body.Attributes {
		rawMod.locals[name] = attr.Expr
	}
	return rawMod, nil
}

// addRawProvider 校验并记录 provider "<plugin>" { plugin_instance_id = ..., alias = ... } 块。
func addRawProvider(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	if len(blk.Labels) != 1 || blk.Labels[0] == "" {
		return nil, fmt.Errorf("provider block in %s must have 1 plugin label", filePath)
	}
	if _, ok := blk.Body.Attributes["plugin_instance_id"]; !ok {
		return nil, fmt.Errorf("provider %q in %s missing required plugin_instance_id", blk.Labels[0], filePath)
	}
	rawMod.providers = append(rawMod.providers, blk)
	return rawMod, nil
}

// addRawModuleCall 校验并记录 module "<name>" { source = ... } 块。
func addRawModuleCall(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	if len(blk.Labels) != 1 || blk.Labels[0] == "" {
		return nil, fmt.Errorf("module block in %s must have 1 name label", filePath)
	}
	if _, ok := blk.Body.Attributes["source"]; !ok {
		return nil, fmt.Errorf("module %q in %s missing required source attribute", blk.Labels[0], filePath)
	}
	rawMod.modules = append(rawMod.modules, blk)
	return rawMod, nil
}

// addRawOutput 校验并记录 output "<name>" { value = ... } 块的 value 表达式。
func addRawOutput(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	if len(blk.Labels) != 1 || blk.Labels[0] == "" {
		return nil, fmt.Errorf("output block in %s must have 1 name label", filePath)
	}
	valAttr, ok := blk.Body.Attributes["value"]
	if !ok {
		return nil, fmt.Errorf("output %q in %s missing required value attribute", blk.Labels[0], filePath)
	}
	rawMod.outputs[blk.Labels[0]] = valAttr.Expr
	return rawMod, nil
}

// addRawResourceOrData 校验并记录 resource "<type_name>" "<name>" 或 data "<type_name>" "<name>" 块，
// 同时检查 type_name 是否满足 <plugin>_<resource> 格式。
func addRawResourceOrData(rawMod *rawModule, blk *hclsyntax.Block, kind enumor.BlockKind, filePath string) (
	*rawModule, error) {

	if len(blk.Labels) != 2 || blk.Labels[0] == "" || blk.Labels[1] == "" {
		return nil, fmt.Errorf("%s block in %s must have 2 labels (<type_name> <name>)", kind, filePath)
	}
	if _, err := extractPluginFromTypeName(blk.Labels[0]); err != nil {
		return nil, fmt.Errorf("invalid %s block in %s: %w", kind, filePath, err)
	}
	if kind == enumor.BlockKindData {
		rawMod.datas = append(rawMod.datas, blk)
		return rawMod, nil
	}
	rawMod.resources = append(rawMod.resources, blk)
	return rawMod, nil
}

// addRawImport 校验并记录 import { to = <type>.<name>, id = "..." } 块。
func addRawImport(rawMod *rawModule, blk *hclsyntax.Block, filePath string) (*rawModule, error) {
	if len(blk.Labels) != 0 {
		return nil, fmt.Errorf("import block in %s must not have labels", filePath)
	}
	_, okTo := blk.Body.Attributes["to"]
	_, okID := blk.Body.Attributes["id"]
	if !okTo || !okID {
		return nil, fmt.Errorf("import block in %s requires both 'to' and 'id' attributes", filePath)
	}
	rawMod.imports = append(rawMod.imports, blk)
	return rawMod, nil
}
