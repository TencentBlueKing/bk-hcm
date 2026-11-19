/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package codec

import (
	"errors"

	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func schemaObject(schema *tfplugin5.Schema) (tftypes.Type, error) {
	if schema == nil || schema.GetBlock() == nil {
		return nil, errors.New("schema is required")
	}
	return blockObject(schema.GetBlock(), "")
}

func blockObject(block *tfplugin5.Schema_Block, path string) (tftypes.Type, error) {
	attrs := map[string]tftypes.Type{}
	if block == nil {
		return tftypes.Object{AttributeTypes: attrs}, nil
	}
	for _, attr := range block.GetAttributes() {
		if attr == nil {
			continue
		}
		name := attr.GetName()
		attrPath := joinPath(path, name)
		if name == "" {
			return nil, pathError(attrPath, "attribute name is empty")
		}
		if _, exists := attrs[name]; exists {
			return nil, pathError(attrPath, "duplicate attribute")
		}
		typ, err := tftypes.ParseJSONType(attr.GetType())
		if err != nil {
			return nil, pathError(attrPath, "invalid attribute type")
		}
		if err = rejectDynamic(typ, attrPath); err != nil {
			return nil, err
		}
		attrs[name] = typ
	}
	for _, nested := range block.GetBlockTypes() {
		if nested == nil {
			continue
		}
		name := nested.GetTypeName()
		blockPath := joinPath(path, name)
		if name == "" {
			return nil, pathError(blockPath, "nested block name is empty")
		}
		if _, exists := attrs[name]; exists {
			return nil, pathError(blockPath, "attribute conflicts with nested block")
		}
		inner, err := blockObject(nested.GetBlock(), blockPath)
		if err != nil {
			return nil, err
		}
		switch nested.GetNesting() {
		case tfplugin5.Schema_NestedBlock_SINGLE, tfplugin5.Schema_NestedBlock_GROUP:
			attrs[name] = inner
		case tfplugin5.Schema_NestedBlock_LIST:
			attrs[name] = tftypes.List{ElementType: inner}
		case tfplugin5.Schema_NestedBlock_SET:
			attrs[name] = tftypes.Set{ElementType: inner}
		case tfplugin5.Schema_NestedBlock_MAP:
			attrs[name] = tftypes.Map{ElementType: inner}
		default:
			return nil, pathError(blockPath, "invalid nesting mode")
		}
	}
	return tftypes.Object{AttributeTypes: attrs}, nil
}

func rejectDynamic(typ tftypes.Type, path string) error {
	if typ == nil || typ.Is(tftypes.DynamicPseudoType) {
		return pathError(path, "dynamic type is not supported")
	}
	switch typed := typ.(type) {
	case tftypes.List:
		return rejectDynamic(typed.ElementType, path)
	case tftypes.Set:
		return rejectDynamic(typed.ElementType, path)
	case tftypes.Map:
		return rejectDynamic(typed.ElementType, path)
	case tftypes.Object:
		for name, attr := range typed.AttributeTypes {
			if err := rejectDynamic(attr, joinPath(path, name)); err != nil {
				return err
			}
		}
	case tftypes.Tuple:
		for i, elem := range typed.ElementTypes {
			if err := rejectDynamic(elem, indexPath(path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}
