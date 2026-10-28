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

// Package iacplugin 提供本地 tf_provider 与三方 http 插件共用的 Schema 校验与目标状态合成能力。
package iacplugin

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/errf"

	tfjson "github.com/hashicorp/terraform-json"
)

// ValidateConfigAndRecompute validates user config and recompute attributes against the resource SchemaBlock.
func ValidateConfigAndRecompute(block *tfjson.SchemaBlock, config map[string]any, recompute []string) error {
	if block == nil {
		return errf.New(errf.InvalidParameter, "resource schema block is nil")
	}
	if _, err := buildRecomputeSet(block, recompute); err != nil {
		return err
	}
	if config == nil {
		return nil
	}
	return validateBlockConfig(block, config, "")
}

// ComposeProposedState validates config and recompute against block and synthesizes proposed_state.
// When config is nil (Delete), proposed_state is nil.
//
// 示例（更新场景，含顶层属性与 List 嵌套块 data_disks）：
//   - Schema:
//     instance_id (computed), instance_name (required),
//     data_disks (NestedBlock, list):
//     disk_size (required), disk_id (computed), kms_key_id (optional+computed)
//   - priorState:
//     {"instance_id": "ins-1", "instance_name": "old", "data_disks": [{"disk_id": "d-1", "disk_size": 50, "kms_key_id": "k-1"}]}
//   - config:
//     {"instance_name": "new", "data_disks": [{"disk_size": 100}, {"disk_size": 200}]}
//   - recompute:
//     ["data_disks[0].kms_key_id"]
//   - 合成结果 proposed_state:
//     {
//     "instance_id":   "ins-1",   // 纯 computed：沿用 priorState
//     "instance_name": "new",     // required：取 config
//     "data_disks": [
//     {"disk_id": "d-1", "disk_size": 100, "kms_key_id": nil}, // [0] 对齐旧项：disk_id 沿用，kms_key_id 命中 recompute 置 nil
//     {"disk_id": nil,   "disk_size": 200, "kms_key_id": nil}, // [1] 新增元素：priorItem=nil，按创建规则置 nil
//     ]
//     }
func ComposeProposedState(block *tfjson.SchemaBlock, priorState, config map[string]any,
	recompute []string) (map[string]any, error) {

	if block == nil {
		return nil, errf.New(errf.InvalidParameter, "resource schema block is nil")
	}

	recomputeSet, err := buildRecomputeSet(block, recompute)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	if err := validateBlockConfig(block, config, ""); err != nil {
		return nil, err
	}

	return composeBlockState(block, priorState, config, recomputeSet, ""), nil
}

// buildRecomputeSet validates each path in recompute against the SchemaBlock tree and returns a lookup set.
//
// 支持的 recompute 路径格式示例：
//   - 顶层属性：         "bk_module_id"
//   - Single 嵌套块属性： "system_disk.disk_type" 或 AttributeNestedType "network.private_ip"
//   - List 嵌套块属性：   "data_disks[0].kms_key_id"（指定下标）或 "data_disks.kms_key_id"（通配所有元素）
//   - Map 嵌套块属性：    "labels_map.env.tag_id"（其中 "env" 不是 Schema 中的字段名，而是用户 config 中动态传入的 map 字典 key）
//     或 "labels_map.tag_id"（省略动态 map key，通配 labels_map 下所有字典项的 tag_id）
//   - 多层嵌套块属性：   "forward_balancer_ids[0].target_attribute[1].weight"
func buildRecomputeSet(block *tfjson.SchemaBlock, recompute []string) (map[string]struct{}, error) {
	recomputeSet := make(map[string]struct{}, len(recompute))
	for _, field := range recompute {
		attr, ok := findAttributeByPath(block, field)
		if !ok || attr == nil {
			return nil, errf.Newf(errf.InvalidParameter, "unknown attribute %q in recompute", field)
		}
		if !(attr.Optional && attr.Computed) {
			return nil, errf.Newf(errf.InvalidParameter,
				"recompute attribute %q must be an optional+computed attribute in schema", field)
		}
		recomputeSet[field] = struct{}{}
	}
	return recomputeSet, nil
}

func findAttributeByPath(block *tfjson.SchemaBlock, path string) (*tfjson.SchemaAttribute, bool) {
	return findAttributeFromParts(block, strings.Split(path, "."))
}

// findAttributeFromParts 沿点分路径分段在 SchemaBlock 树中递归查找目标 SchemaAttribute。
//
// 寻址示例：
//  1. 路径 "forward_balancer_ids[0].target_attribute[0].weight"
//     parts = ["forward_balancer_ids[0]", "target_attribute[0]", "weight"]
//     - 第 1 层：解析出 key="forward_balancer_ids", hasIdx=true -> 命中 NestedBlocks["forward_balancer_ids"] (list 模式) -> 进入子 Block；
//     - 第 2 层：解析出 key="target_attribute", hasIdx=true -> 命中子 Block.NestedBlocks["target_attribute"] (list 模式) -> 进入孙 Block；
//     - 第 3 层（末段）：len(parts)==1, key="weight", hasIdx=false -> 返回孙 Block.Attributes["weight"]。
//  2. Map 模式路径 "labels_map.env.tag_id"
//     parts = ["labels_map", "env", "tag_id"]
//     - 第 1 层：命中 NestedBlocks["labels_map"] (map 模式)；
//     - 先尝试在 labels_map.Block 下直接找 ["env", "tag_id"]（未命中，因为 "env" 是运行时动态 map key 而非 Schema 字段名）；
//     - 触发 Map 模式分支：跳过动态 key parts[1]（"env"），拿剩余的 parts[2:]（["tag_id"]）在 labels_map.Block 下查找，命中 Attributes["tag_id"]。
func findAttributeFromParts(curr *tfjson.SchemaBlock, parts []string) (*tfjson.SchemaAttribute, bool) {
	if curr == nil || len(parts) == 0 {
		return nil, false
	}
	key, hasIdx, valid := parsePathSegment(parts[0])
	if !valid {
		return nil, false
	}
	if len(parts) == 1 {
		if hasIdx {
			return nil, false
		}
		attr, ok := curr.Attributes[key]
		return attr, ok && attr != nil
	}

	bt := lookupChildBlockType(curr, key)
	if bt == nil || bt.Block == nil {
		return nil, false
	}
	if hasIdx && !isSliceNestingMode(bt.NestingMode) {
		return nil, false
	}
	if attr, ok := findAttributeFromParts(bt.Block, parts[1:]); ok {
		return attr, true
	}
	if bt.NestingMode == tfjson.SchemaNestingModeMap && len(parts) >= 3 {
		if _, mapKeyHasIdx, mapKeyValid := parsePathSegment(parts[1]); mapKeyValid && !mapKeyHasIdx {
			return findAttributeFromParts(bt.Block, parts[2:])
		}
	}
	return nil, false
}

// lookupChildBlockType 统一从当前 SchemaBlock 的 NestedBlocks（SDKv2 block_types）
// 或 Attributes[key].AttributeNestedType（Plugin Framework nested_type）中获取子块定义。
func lookupChildBlockType(curr *tfjson.SchemaBlock, key string) *tfjson.SchemaBlockType {
	if bt, ok := curr.NestedBlocks[key]; ok && bt != nil {
		return bt
	}
	if attr, ok := curr.Attributes[key]; ok && attr != nil && attr.AttributeNestedType != nil {
		return toSyntheticBlockType(attr.AttributeNestedType)
	}
	return nil
}

// validateBlockConfig 递归校验当前 SchemaBlock 层级下的用户 config：
//  1. 拦截既不在 Attributes 也不在 NestedBlocks 中的未知字段；
//  2. 拦截非 nil 的纯 computed 属性（computed=true, required=false, optional=false）；
//  3. 若属性为 AttributeNestedType 或属于 NestedBlocks，继续向下递归校验子块内容。
//
// 报错路径示例：
//   - 顶层未知字段：        unknown attribute "foo" in config
//   - 二层 List 嵌套块内未知：unknown attribute "forward_balancer_ids[0].target_attribute[1].bad_field" in config
//   - 嵌套块内设置纯 computed：computed-only attribute "system_disk.disk_id" cannot be set in config
func validateBlockConfig(block *tfjson.SchemaBlock, config map[string]any, prefix string) error {
	for k, v := range config {
		fieldPath := joinFieldPath(prefix, k)
		attr, isAttr := block.Attributes[k]
		bt, isBlock := block.NestedBlocks[k]
		if !isAttr && !isBlock {
			return errf.Newf(errf.InvalidParameter, "unknown attribute %q in config", fieldPath)
		}
		if isAttr && attr != nil {
			if isComputedOnly(attr) && v != nil {
				return errf.Newf(errf.InvalidParameter, "computed-only attribute %q cannot be set in config", fieldPath)
			}
			if attr.AttributeNestedType != nil && v != nil {
				bt = toSyntheticBlockType(attr.AttributeNestedType)
				isBlock = true
			}
		}
		if isBlock && bt != nil && v != nil {
			if err := validateNestedBlockConfig(bt, v, fieldPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateNestedBlockConfig 按嵌套块的 NestingMode 校验 config 值的容器类型并剥壳递归：
//   - Single / Group 模式：要求 val 为 map[string]any，例如 {"system_disk": {"disk_size": 100}}
//   - Map 模式：           要求 val 为 map[string]map[string]any，例如 {"labels_map": {"env": {"value": "prod"}}}
//   - List / Set 模式：    要求 val 为 []map[string]any，例如 {"data_disks": [{"disk_size": 100}]}
func validateNestedBlockConfig(bt *tfjson.SchemaBlockType, val any, fieldPath string) error {
	if bt.Block == nil {
		return nil
	}
	switch bt.NestingMode {
	case tfjson.SchemaNestingModeSingle, tfjson.SchemaNestingModeGroup:
		subMap, ok := val.(map[string]any)
		if !ok {
			return errf.Newf(errf.InvalidParameter, "nested block %q expects object value", fieldPath)
		}
		return validateBlockConfig(bt.Block, subMap, fieldPath)
	case tfjson.SchemaNestingModeMap:
		return validateMapNestedBlock(bt.Block, val, fieldPath)
	default:
		return validateSliceNestedBlock(bt.Block, val, fieldPath)
	}
}

func validateMapNestedBlock(block *tfjson.SchemaBlock, val any, fieldPath string) error {
	outerMap, ok := val.(map[string]any)
	if !ok {
		return errf.Newf(errf.InvalidParameter, "nested block %q expects map of objects", fieldPath)
	}
	for k, item := range outerMap {
		itemPath := joinFieldPath(fieldPath, k)
		itemMap, ok := item.(map[string]any)
		if !ok {
			return errf.Newf(errf.InvalidParameter, "nested block element %q expects object value", itemPath)
		}
		if err := validateBlockConfig(block, itemMap, itemPath); err != nil {
			return err
		}
	}
	return nil
}

func validateSliceNestedBlock(block *tfjson.SchemaBlock, val any, fieldPath string) error {
	sliceVal, ok := val.([]any)
	if !ok {
		return errf.Newf(errf.InvalidParameter, "nested block %q expects array of objects", fieldPath)
	}
	for i, item := range sliceVal {
		itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		itemMap, ok := item.(map[string]any)
		if !ok {
			return errf.Newf(errf.InvalidParameter, "nested block element %q expects object value", itemPath)
		}
		if err := validateBlockConfig(block, itemMap, itemPath); err != nil {
			return err
		}
	}
	return nil
}

// composeBlockState 递归合成单个 SchemaBlock 层级的 proposed_state：
//  1. 遍历 Attributes：普通属性按 4 类属性规则合成；AttributeNestedType 递归合成子属性；
//  2. 遍历 NestedBlocks：当 config 中存在该子块时，取出 priorState 同名子块对齐并递归合成。
func composeBlockState(block *tfjson.SchemaBlock, priorState, config map[string]any, recomputeSet map[string]struct{},
	prefix string) map[string]any {

	proposed := make(map[string]any, len(block.Attributes)+len(block.NestedBlocks))
	for name, attr := range block.Attributes {
		if attr == nil {
			continue
		}
		fieldPath := joinFieldPath(prefix, name)
		if attr.AttributeNestedType != nil {
			proposed[name] = composeNestedAttrValue(name, fieldPath, attr, priorState, config, recomputeSet)
			continue
		}
		proposed[name] = composeAttributeValue(name, fieldPath, attr, priorState, config, recomputeSet)
	}

	for name, bt := range block.NestedBlocks {
		val, ok := config[name]
		if !ok {
			continue
		}
		if val == nil || bt == nil || bt.Block == nil {
			proposed[name] = val
			continue
		}
		var rawPrior any
		if priorState != nil {
			rawPrior = priorState[name]
		}
		proposed[name] = composeNestedBlockValue(bt, rawPrior, val, recomputeSet, joinFieldPath(prefix, name))
	}

	return proposed
}

// composeNestedAttrValue 处理 Plugin Framework 风格的嵌套属性（AttributeNestedType）：
//   - 若用户在 config 中未传该外层属性（或传 nil）：退化为按外层 attr 自身的 Optional/Computed 规则处理
//     （例如 network 为 optional+computed 且 config 未传时，直接沿用 priorState["network"] 整棵子树）；
//   - 若用户在 config 中传入了非 nil 的子对象（如 {"network": {"subnet_id": "sub-1"}}）：
//     将其与 priorState["network"]（如 {"subnet_id": "sub-1", "private_ip": "10.0.0.1", "mac_address": "aa:bb"}）
//     对齐递归，合成出保留了 private_ip 和 mac_address 的完整子对象。
func composeNestedAttrValue(name, fieldPath string, attr *tfjson.SchemaAttribute, priorState, config map[string]any,
	recomputeSet map[string]struct{}) any {

	rawCfg, hasCfg := config[name]
	if !hasCfg || rawCfg == nil {
		return composeAttributeValue(name, fieldPath, attr, priorState, config, recomputeSet)
	}
	var rawPrior any
	if priorState != nil {
		rawPrior = priorState[name]
	}
	bt := toSyntheticBlockType(attr.AttributeNestedType)
	return composeNestedBlockValue(bt, rawPrior, rawCfg, recomputeSet, fieldPath)
}

// composeNestedBlockValue 按 NestingMode 将嵌套块的旧状态 rawPrior 与新配置 rawCfg 对齐并递归合成：
//   - Single / Group：priorMap 与 cfgMap 一对一对齐，路径保持 fieldPath（如 "system_disk"）；
//   - Map：           按字典 key 对齐 priorOuter[k] 与 cfgOuter[k]，路径为 "labels_map.env"；
//   - List / Set：    按下标 i 对齐 priorSlice[i] 与 cfgSlice[i]，路径为 "data_disks[i]"：
//     若 i < len(priorSlice)（更新已有元素），传入 priorSlice[i] 从而保留其内部 computed/optional+computed 字段；
//     若 i >= len(priorSlice)（列表末尾新增元素），priorItem 为 nil，子块内部按创建规则将未指定字段置为 nil。
func composeNestedBlockValue(bt *tfjson.SchemaBlockType, rawPrior, rawCfg any,
	recomputeSet map[string]struct{}, fieldPath string) any {

	switch bt.NestingMode {
	case tfjson.SchemaNestingModeSingle, tfjson.SchemaNestingModeGroup:
		cfgMap, _ := rawCfg.(map[string]any)
		priorMap, _ := rawPrior.(map[string]any)
		return composeBlockState(bt.Block, priorMap, cfgMap, recomputeSet, fieldPath)
	case tfjson.SchemaNestingModeMap:
		cfgOuter, _ := rawCfg.(map[string]any)
		priorOuter, _ := rawPrior.(map[string]any)
		out := make(map[string]any, len(cfgOuter))
		for k, item := range cfgOuter {
			cfgItem, _ := item.(map[string]any)
			var priorItem map[string]any
			if priorOuter != nil {
				priorItem, _ = priorOuter[k].(map[string]any)
			}
			// k 是用户在 config 中传入的运行时动态 map key（如 "env"），而非 Schema 字段名；
			// 若 recompute 传了通配路径 "labels_map.tag_id"，将子块内部相对路径透传进 itemRecompute 以便匹配。
			itemPath := joinFieldPath(fieldPath, k)
			itemRecompute := deriveMapItemRecomputeSet(recomputeSet, fieldPath, itemPath)
			out[k] = composeBlockState(bt.Block, priorItem, cfgItem, itemRecompute, itemPath)
		}
		return out
	default:
		cfgSlice, _ := rawCfg.([]any)
		priorSlice, _ := rawPrior.([]any)
		out := make([]any, len(cfgSlice))
		for i, item := range cfgSlice {
			cfgItem, _ := item.(map[string]any)
			var priorItem map[string]any
			if i < len(priorSlice) {
				priorItem, _ = priorSlice[i].(map[string]any)
			}
			itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			out[i] = composeBlockState(bt.Block, priorItem, cfgItem, recomputeSet, itemPath)
		}
		return out
	}
}

func isComputedOnly(attr *tfjson.SchemaAttribute) bool {
	return attr.Computed && !attr.Required && !attr.Optional
}

// composeAttributeValue 按 Schema 属性的 4 类规则合成单个叶子属性的值：
//  1. 纯 computed（computed=true, required=false, optional=false，如 bk_host_id）：
//     创建时（priorState==nil）置 nil；更新时沿用 priorState[name]。
//  2. optional + computed（optional=true, computed=true，如 bk_module_id）：
//     - config 显式传了非 nil 值 -> 取 config[name]；
//     - config 未传或为 nil：创建时置 nil；更新时若命中 recomputeSet 则显式置 nil 触发重算，否则沿用 priorState[name]。
//  3. required 或纯 optional（computed=false，如 bk_biz_id、operator）：
//     取 config[name]（未传则为 nil）。
func composeAttributeValue(name, fieldPath string, attr *tfjson.SchemaAttribute, priorState, config map[string]any,
	recomputeSet map[string]struct{}) any {

	if isComputedOnly(attr) {
		if priorState == nil {
			return nil
		}
		return priorState[name]
	}

	if attr.Optional && attr.Computed {
		if val, ok := config[name]; ok && val != nil {
			return val
		}
		if priorState == nil {
			return nil
		}
		if shouldRecomputeField(fieldPath, recomputeSet) {
			return nil
		}
		return priorState[name]
	}

	if val, ok := config[name]; ok {
		return val
	}
	return nil
}

// shouldRecomputeField 判断当前属性路径是否命中 recompute 集合（同时支持精确下标路径与无下标通配路径）。
//
// 示例：当正在合成第 0 个数据盘的子字段，fieldPath = "data_disks[0].kms_key_id" 时：
//   - 若 recompute 包含精确路径 "data_disks[0].kms_key_id" -> 命中（仅重算第 0 块盘）；
//   - 若 recompute 包含通配路径 "data_disks.kms_key_id"    -> stripAllSliceIndices 后命中（重算所有数据盘）。
func shouldRecomputeField(fieldPath string, recomputeSet map[string]struct{}) bool {
	if _, ok := recomputeSet[fieldPath]; ok {
		return true
	}
	wildcardPath := stripAllSliceIndices(fieldPath)
	_, ok := recomputeSet[wildcardPath]
	return ok
}

// deriveMapItemRecomputeSet 当 recompute 中包含省略了动态 map key 的通配路径（如 "labels_map.tag_id"）时，
// 将其展开为当前动态 key 对应的具体路径（如 "labels_map.env.tag_id"），使子块内部能统一按 itemPath 匹配。
func deriveMapItemRecomputeSet(recomputeSet map[string]struct{}, mapBlockPath, itemPath string) map[string]struct{} {
	if len(recomputeSet) == 0 {
		return recomputeSet
	}
	prefix := mapBlockPath + "."
	var expanded map[string]struct{}
	for p := range recomputeSet {
		if !strings.HasPrefix(p, prefix) {
			continue
		}
		if expanded == nil {
			expanded = make(map[string]struct{}, len(recomputeSet)+1)
			for orig := range recomputeSet {
				expanded[orig] = struct{}{}
			}
		}
		expanded[itemPath+"."+strings.TrimPrefix(p, prefix)] = struct{}{}
	}
	if expanded != nil {
		return expanded
	}
	return recomputeSet
}

func toSyntheticBlockType(nested *tfjson.SchemaNestedAttributeType) *tfjson.SchemaBlockType {
	return &tfjson.SchemaBlockType{
		NestingMode: nested.NestingMode,
		MinItems:    nested.MinItems,
		MaxItems:    nested.MaxItems,
		Block:       &tfjson.SchemaBlock{Attributes: nested.Attributes},
	}
}

func isSliceNestingMode(mode tfjson.SchemaNestingMode) bool {
	return mode != tfjson.SchemaNestingModeSingle &&
		mode != tfjson.SchemaNestingModeGroup &&
		mode != tfjson.SchemaNestingModeMap
}

// parsePathSegment 解析单段路径，返回（字段名 key, 是否带数组下标 hasIdx, 语法是否合法 valid）。
//
// 示例：
//   - "system_disk"              -> ("system_disk", false, true)
//   - "forward_balancer_ids[0]"  -> ("forward_balancer_ids", true, true)
//   - "forward_balancer_ids[]"   -> ("", false, false)
//   - "forward_balancer_ids[ab]" -> ("", false, false)
func parsePathSegment(seg string) (string, bool, bool) {
	openIdx := strings.IndexByte(seg, '[')
	if openIdx < 0 {
		if strings.ContainsRune(seg, ']') || seg == "" {
			return "", false, false
		}
		return seg, false, true
	}
	if openIdx == 0 || !strings.HasSuffix(seg, "]") {
		return "", false, false
	}
	idxStr := seg[openIdx+1 : len(seg)-1]
	if idxStr == "" {
		return "", false, false
	}
	for i := 0; i < len(idxStr); i++ {
		if idxStr[i] < '0' || idxStr[i] > '9' {
			return "", false, false
		}
	}
	return seg[:openIdx], true, true
}

// stripAllSliceIndices 剥离路径中所有段的数组下标，用于通配匹配。
// 示例："forward_balancer_ids[0].target_attribute[1].weight" -> "forward_balancer_ids.target_attribute.weight"
func stripAllSliceIndices(path string) string {
	if !strings.Contains(path, "[") {
		return path
	}
	parts := strings.Split(path, ".")
	for i, part := range parts {
		if key, _, valid := parsePathSegment(part); valid {
			parts[i] = key
		}
	}
	return strings.Join(parts, ".")
}

func joinFieldPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
