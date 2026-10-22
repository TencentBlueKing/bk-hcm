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

// Package instance 提供 hc-service 侧插件实例核心业务逻辑与属性拆分合并能力。
package instance

import (
	"fmt"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/tools/assert"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

// splitAndValidateProperties 基于 Provider SchemaBlock 校验 properties，并按属性敏感标记同构拆分为 attrs 与 secretAttrs。
func splitAndValidateProperties(block *tfjson.SchemaBlock, properties map[string]any) (
	map[string]any, map[string]any, error) {

	if block == nil {
		return nil, nil, errf.New(errf.InvalidParameter, "provider schema block is nil")
	}
	return splitBlockProperties(block, properties, "")
}

// splitBlockProperties 递归校验当前 SchemaBlock 层级的属性与子块，并按敏感标记同构拆分为 attrs 与 secretAttrs。
func splitBlockProperties(block *tfjson.SchemaBlock, props map[string]any, prefix string) (map[string]any,
	map[string]any, error) {

	attrs := make(map[string]any)
	secretAttrs := make(map[string]any)

	// 1. 校验当前 Block 层级声明的必填属性与最小数量大于 0 的嵌套子块是否缺失
	if err := validateRequiredBlockFields(block, props, prefix); err != nil {
		return nil, nil, err
	}

	// 2. 遍历调用方传入的属性键值，按 Attributes 或 NestedBlocks 递归校验并拆分
	for key, val := range props {
		if val == nil {
			continue
		}
		fieldPath := joinFieldPath(prefix, key)

		// 命中当前层的 Attributes（含标量、集合、cty.Object 以及 Plugin Framework 的 NestedAttribute）
		if attr, ok := block.Attributes[key]; ok && attr != nil {
			subAttrs, subSecrets, hasAttrs, hasSecrets, err := splitAndValidateAttribute(attr, val, fieldPath)
			if err != nil {
				return nil, nil, err
			}
			if hasAttrs {
				attrs[key] = subAttrs
			}
			if hasSecrets {
				secretAttrs[key] = subSecrets
			}
			continue
		}

		// 命中当前层的 NestedBlocks（传统 SDKv2 block_types 嵌套块）
		if blockType, ok := block.NestedBlocks[key]; ok && blockType != nil {
			subAttrs, subSecrets, hasAttrs, hasSecrets, err := splitNestedBlockValue(blockType, val, fieldPath)
			if err != nil {
				return nil, nil, err
			}
			if hasAttrs {
				attrs[key] = subAttrs
			}
			if hasSecrets {
				secretAttrs[key] = subSecrets
			}
			continue
		}

		return nil, nil, errf.Newf(errf.InvalidParameter,
			"unknown property %q is not defined in provider schema", fieldPath)
	}

	return attrs, secretAttrs, nil
}

// validateRequiredBlockFields 校验当前 SchemaBlock 层级下的必填属性（Required=true）与必传子块（MinItems>0）；
// 子块内部的必填项会在后续递归调用 splitBlockProperties 时逐层触发本函数完成校验。
func validateRequiredBlockFields(block *tfjson.SchemaBlock, props map[string]any, prefix string) error {
	for key, attr := range block.Attributes {
		if attr == nil || !attr.Required {
			continue
		}
		fieldPath := joinFieldPath(prefix, key)
		val, exists := props[key]
		if !exists || val == nil {
			return errf.Newf(errf.InvalidParameter, "required property %q is missing", fieldPath)
		}
	}

	for key, blockType := range block.NestedBlocks {
		if blockType == nil || blockType.MinItems == 0 {
			continue
		}
		fieldPath := joinFieldPath(prefix, key)
		val, exists := props[key]
		if !exists || val == nil {
			return errf.Newf(errf.InvalidParameter,
				"nested block %q requires at least %d items", fieldPath, blockType.MinItems)
		}
	}

	return nil
}

// splitAndValidateAttribute 校验单个 SchemaAttribute 并按 Sensitive 标记拆分返回值：
// - 若为 Plugin Framework 的 AttributeNestedType（nested_type），将其适配为虚拟 SchemaBlockType 复用嵌套块递归校验与拆分；
// - 若为常规 AttributeType（cty.Type），校验标量/集合/对象/元组类型后按外层 Sensitive 归入 attrs 或 secretAttrs。
func splitAndValidateAttribute(attr *tfjson.SchemaAttribute, val any, fieldPath string) (any, any, bool, bool, error) {
	if attr.Computed && !attr.Optional && !attr.Required {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"property %q is computed read-only and cannot be set", fieldPath)
	}

	// Terraform 存在两套嵌套结构定义方式，对应 terraform-json 中两个不同结构体：
	// 1) SDKv2 的 NestedBlocks（block_types）：类型为 *tfjson.SchemaBlockType，子字段挂在 Block.Attributes 下；
	// 2) Plugin Framework 的 AttributeNestedType（nested_type）：类型为 *tfjson.SchemaNestedAttributeType，
	//    直接平铺持有 NestingMode / MinItems / MaxItems / Attributes，没有内层 Block 包装。
	// 两者的校验与同构拆分语义完全一致（按 NestingMode 分发、校验数量上下限、递归拆分内部 Attributes），
	// 但 Go 结构体不同无法直接传参，因此将 SchemaNestedAttributeType 包装为等价的 SchemaBlockType 以复用同一套逻辑。
	if attr.AttributeNestedType != nil {
		syntheticBlockType := &tfjson.SchemaBlockType{
			NestingMode: attr.AttributeNestedType.NestingMode,
			MinItems:    attr.AttributeNestedType.MinItems,
			MaxItems:    attr.AttributeNestedType.MaxItems,
			Block: &tfjson.SchemaBlock{
				Attributes: attr.AttributeNestedType.Attributes,
			},
		}
		subAttrs, subSecrets, hasAttrs, hasSecrets, err := splitNestedBlockValue(syntheticBlockType, val, fieldPath)
		if err != nil {
			return nil, nil, false, false, err
		}
		// 若父属性自身标记了 Sensitive=true，则整棵子树统一归入 secretAttrs
		if attr.Sensitive {
			return nil, val, false, true, nil
		}
		return subAttrs, subSecrets, hasAttrs, hasSecrets, nil
	}

	// 常规属性（非 nested_type）：通过 cty.Type 声明值类型（标量 string/number/bool，或复合类型 list/set/map/object/tuple）。
	// 由于 cty.Type 仅描述数据结构与元素类型，不包含子字段级的 Sensitive 元数据，
	// 因此完成递归类型校验后，直接按当前属性外层的 Sensitive 标记将整个属性值归入 attrs 或 secretAttrs。
	if attr.AttributeType != cty.NilType {
		if err := validateCtyType(attr.AttributeType, val, fieldPath); err != nil {
			return nil, nil, false, false, err
		}
	}

	if attr.Sensitive {
		return nil, val, false, true, nil
	}
	return val, nil, true, false, nil
}

// validateCtyType 递归校验 JSON 值是否符合声明的 cty.Type（支持标量、List/Set、Map、Object、Tuple）。
func validateCtyType(ty cty.Type, val any, fieldPath string) error {
	if ty == cty.NilType || ty == cty.DynamicPseudoType {
		return nil
	}

	switch {
	case ty == cty.String:
		if _, ok := val.(string); !ok {
			return errf.Newf(errf.InvalidParameter, "property %q expects string value", fieldPath)
		}
	case ty == cty.Bool:
		if _, ok := val.(bool); !ok {
			return errf.Newf(errf.InvalidParameter, "property %q expects bool value", fieldPath)
		}
	case ty == cty.Number:
		if !assert.IsNumeric(val) {
			return errf.Newf(errf.InvalidParameter, "property %q expects number value", fieldPath)
		}
	case ty.IsListType() || ty.IsSetType():
		return validateCtySlice(ty.ElementType(), val, fieldPath)
	case ty.IsMapType():
		return validateCtyMap(ty.ElementType(), val, fieldPath)
	case ty.IsObjectType():
		return validateCtyObject(ty, val, fieldPath)
	case ty.IsTupleType():
		return validateCtyTuple(ty, val, fieldPath)
	default:
		return nil
	}

	return nil
}

// validateCtyObject 校验 cty.Object 结构体的非可选字段必填性、未知字段拦截及各子字段类型。
func validateCtyObject(ty cty.Type, val any, fieldPath string) error {
	objMap, ok := val.(map[string]any)
	if !ok {
		return errf.Newf(errf.InvalidParameter, "property %q expects object value", fieldPath)
	}
	attrTypes := ty.AttributeTypes()
	for k := range attrTypes {
		if ty.AttributeOptional(k) {
			continue
		}
		itemVal, exists := objMap[k]
		if !exists || itemVal == nil {
			return errf.Newf(errf.InvalidParameter,
				"required property %q is missing", joinFieldPath(fieldPath, k))
		}
	}
	for k, itemVal := range objMap {
		itemPath := joinFieldPath(fieldPath, k)
		attrTy, exists := attrTypes[k]
		if !exists {
			return errf.Newf(errf.InvalidParameter,
				"unknown property %q is not defined in provider schema", itemPath)
		}
		if itemVal == nil {
			continue
		}
		if err := validateCtyType(attrTy, itemVal, itemPath); err != nil {
			return err
		}
	}
	return nil
}

// validateCtyTuple 校验 cty.Tuple 定长数组的长度与各位置元素类型。
func validateCtyTuple(ty cty.Type, val any, fieldPath string) error {
	sliceVal, ok := val.([]any)
	if !ok {
		return errf.Newf(errf.InvalidParameter, "property %q expects tuple/array value", fieldPath)
	}
	elemTypes := ty.TupleElementTypes()
	if len(sliceVal) != len(elemTypes) {
		return errf.Newf(errf.InvalidParameter,
			"property %q expects %d tuple elements, got %d", fieldPath, len(elemTypes), len(sliceVal))
	}
	for i, elemTy := range elemTypes {
		if sliceVal[i] == nil {
			continue
		}
		itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		if err := validateCtyType(elemTy, sliceVal[i], itemPath); err != nil {
			return err
		}
	}
	return nil
}

func validateCtySlice(elemType cty.Type, val any, fieldPath string) error {
	sliceVal, ok := val.([]any)
	if !ok {
		return errf.Newf(errf.InvalidParameter, "property %q expects list/array value", fieldPath)
	}
	for i, item := range sliceVal {
		itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		if err := validateCtyType(elemType, item, itemPath); err != nil {
			return err
		}
	}
	return nil
}

func validateCtyMap(elemType cty.Type, val any, fieldPath string) error {
	mapVal, ok := val.(map[string]any)
	if !ok {
		return errf.Newf(errf.InvalidParameter, "property %q expects map/object value", fieldPath)
	}
	for k, item := range mapVal {
		itemPath := joinFieldPath(fieldPath, k)
		if err := validateCtyType(elemType, item, itemPath); err != nil {
			return err
		}
	}
	return nil
}

// splitNestedBlockValue 按嵌套模式（single/group、map、list/set）分发嵌套块或嵌套属性的递归校验与拆分。
func splitNestedBlockValue(blockType *tfjson.SchemaBlockType, val any, fieldPath string) (any, any, bool, bool, error) {
	if blockType.Block == nil {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q schema block definition is nil", fieldPath)
	}

	switch blockType.NestingMode {
	case tfjson.SchemaNestingModeSingle, tfjson.SchemaNestingModeGroup:
		subMap, ok := val.(map[string]any)
		if !ok {
			return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
				"nested block %q expects object value", fieldPath)
		}
		subAttrs, subSecrets, err := splitBlockProperties(blockType.Block, subMap, fieldPath)
		if err != nil {
			return nil, nil, false, false, err
		}
		return subAttrs, subSecrets, len(subAttrs) > 0, len(subSecrets) > 0, nil

	case tfjson.SchemaNestingModeMap:
		return splitMapNestedBlock(blockType, val, fieldPath)

	default:
		return splitSliceNestedBlock(blockType, val, fieldPath)
	}
}

// splitSliceNestedBlock 处理 list/set 模式的嵌套块数组，校验 MinItems/MaxItems，
// 并使用空对象 {} 占位保持 attrs 与 secretAttrs 数组下标对齐。
func splitSliceNestedBlock(blockType *tfjson.SchemaBlockType, val any, fieldPath string) (any, any, bool, bool, error) {
	sliceVal, ok := val.([]any)
	if !ok {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q expects array of objects", fieldPath)
	}

	if blockType.MinItems > 0 && uint64(len(sliceVal)) < blockType.MinItems {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q requires at least %d items", fieldPath, blockType.MinItems)
	}
	if blockType.MaxItems > 0 && uint64(len(sliceVal)) > blockType.MaxItems {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q allows at most %d items", fieldPath, blockType.MaxItems)
	}

	attrsSlice := make([]any, len(sliceVal))
	secretSlice := make([]any, len(sliceVal))
	hasAnyAttrs := false
	hasAnySecrets := false

	for i, item := range sliceVal {
		itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
				"nested block element %q expects object value", itemPath)
		}

		subAttrs, subSecrets, err := splitBlockProperties(blockType.Block, itemMap, itemPath)
		if err != nil {
			return nil, nil, false, false, err
		}

		if len(subAttrs) > 0 {
			hasAnyAttrs = true
			attrsSlice[i] = subAttrs
		} else {
			attrsSlice[i] = map[string]any{}
		}

		if len(subSecrets) > 0 {
			hasAnySecrets = true
			secretSlice[i] = subSecrets
		} else {
			secretSlice[i] = map[string]any{}
		}
	}

	return attrsSlice, secretSlice, hasAnyAttrs, hasAnySecrets, nil
}

// splitMapNestedBlock 处理 map 模式的嵌套块字典，校验 MinItems/MaxItems 并按 key 递归校验拆分。
func splitMapNestedBlock(blockType *tfjson.SchemaBlockType, val any, fieldPath string) (any, any, bool, bool, error) {
	outerMap, ok := val.(map[string]any)
	if !ok {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q expects map of objects", fieldPath)
	}

	if blockType.MinItems > 0 && uint64(len(outerMap)) < blockType.MinItems {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q requires at least %d items", fieldPath, blockType.MinItems)
	}
	if blockType.MaxItems > 0 && uint64(len(outerMap)) > blockType.MaxItems {
		return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
			"nested block %q allows at most %d items", fieldPath, blockType.MaxItems)
	}

	attrsMap := make(map[string]any)
	secretMap := make(map[string]any)

	for k, item := range outerMap {
		itemPath := joinFieldPath(fieldPath, k)
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, nil, false, false, errf.Newf(errf.InvalidParameter,
				"nested block element %q expects object value", itemPath)
		}

		subAttrs, subSecrets, err := splitBlockProperties(blockType.Block, itemMap, itemPath)
		if err != nil {
			return nil, nil, false, false, err
		}
		if len(subAttrs) > 0 {
			attrsMap[k] = subAttrs
		}
		if len(subSecrets) > 0 {
			secretMap[k] = subSecrets
		}
	}

	return attrsMap, secretMap, len(attrsMap) > 0, len(secretMap) > 0, nil
}

// mergeInstanceProperties 将同构拆分的 attrs 与解密后的 secretAttrs 深度合并为完整 properties。
func mergeInstanceProperties(attrs, secretAttrs map[string]any) map[string]any {
	merged := make(map[string]any)

	for k, v := range attrs {
		merged[k] = v
	}

	for k, secretVal := range secretAttrs {
		existing, exists := merged[k]
		if !exists {
			merged[k] = secretVal
			continue
		}
		merged[k] = mergeJSONNode(existing, secretVal)
	}

	return merged
}

func mergeJSONNode(attrNode, secretNode any) any {
	attrMap, attrIsMap := attrNode.(map[string]any)
	secretMap, secretIsMap := secretNode.(map[string]any)
	if attrIsMap && secretIsMap {
		return mergeInstanceProperties(attrMap, secretMap)
	}

	attrSlice, attrIsSlice := attrNode.([]any)
	secretSlice, secretIsSlice := secretNode.([]any)
	if attrIsSlice && secretIsSlice {
		maxLen := len(attrSlice)
		if len(secretSlice) > maxLen {
			maxLen = len(secretSlice)
		}
		res := make([]any, maxLen)
		for i := 0; i < maxLen; i++ {
			switch {
			case i < len(attrSlice) && i < len(secretSlice):
				res[i] = mergeJSONNode(attrSlice[i], secretSlice[i])
			case i < len(attrSlice):
				res[i] = attrSlice[i]
			default:
				res[i] = secretSlice[i]
			}
		}
		return res
	}

	return secretNode
}

func joinFieldPath(prefix, key string) string {
	if len(prefix) == 0 {
		return key
	}
	return prefix + "." + key
}
