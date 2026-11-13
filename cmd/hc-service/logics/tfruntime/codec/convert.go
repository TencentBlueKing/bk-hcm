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
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func convert(typ tftypes.Type, value any, path string) (tftypes.Value, error) {
	value = converter.NormalizeNil(value)
	if value == nil || value == Unknown {
		return typedValue(typ, value, path)
	}
	switch typed := typ.(type) {
	case tftypes.Object:
		return convertObject(typed, value, path)
	case tftypes.Map:
		return convertMap(typed.ElementType, value, path)
	case tftypes.List:
		return convertCollection(tftypes.List{ElementType: typed.ElementType}, typed.ElementType, value, path)
	case tftypes.Set:
		return convertCollection(tftypes.Set{ElementType: typed.ElementType}, typed.ElementType, value, path)
	case tftypes.Tuple:
		return convertTuple(typed, value, path)
	default:
		if !scalarOK(typ, value) {
			return tftypes.Value{}, pathError(path, "type mismatch")
		}
		return typedValue(typ, value, path)
	}
}

func convertObject(typ tftypes.Object, value any, path string) (tftypes.Value, error) {
	if value == nil || value == Unknown {
		return typedValue(typ, value, path)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return tftypes.Value{}, pathError(path, "type mismatch")
	}
	for name := range object {
		if _, exists := typ.AttributeTypes[name]; !exists {
			return tftypes.Value{}, pathError(joinPath(path, name), "unsupported attribute")
		}
	}
	attrs := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attrType := range typ.AttributeTypes {
		child, exists := object[name]
		if !exists {
			child = nil
		}
		converted, err := convert(attrType, child, joinPath(path, name))
		if err != nil {
			return tftypes.Value{}, err
		}
		attrs[name] = converted
	}
	return typedValue(typ, attrs, path)
}

func convertMap(elem tftypes.Type, value any, path string) (tftypes.Value, error) {
	typ := tftypes.Map{ElementType: elem}
	if value == nil || value == Unknown {
		return typedValue(typ, value, path)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return tftypes.Value{}, pathError(path, "type mismatch")
	}
	attrs := make(map[string]tftypes.Value, len(object))
	for name, child := range object {
		converted, err := convert(elem, child, joinPath(path, name))
		if err != nil {
			return tftypes.Value{}, err
		}
		attrs[name] = converted
	}
	return typedValue(typ, attrs, path)
}

func convertCollection(typ tftypes.Type, elem tftypes.Type, value any, path string) (tftypes.Value, error) {
	if value == nil || value == Unknown {
		return typedValue(typ, value, path)
	}
	items, ok := value.([]any)
	if !ok {
		return tftypes.Value{}, pathError(path, "type mismatch")
	}
	converted := make([]tftypes.Value, len(items))
	for i, item := range items {
		got, err := convert(elem, item, indexPath(path, i))
		if err != nil {
			return tftypes.Value{}, err
		}
		converted[i] = got
	}
	return typedValue(typ, converted, path)
}

func convertTuple(typ tftypes.Tuple, value any, path string) (tftypes.Value, error) {
	if value == nil || value == Unknown {
		return typedValue(typ, value, path)
	}
	items, ok := value.([]any)
	if !ok || len(items) != len(typ.ElementTypes) {
		return tftypes.Value{}, pathError(path, "type mismatch")
	}
	converted := make([]tftypes.Value, len(items))
	for i, item := range items {
		got, err := convert(typ.ElementTypes[i], item, indexPath(path, i))
		if err != nil {
			return tftypes.Value{}, err
		}
		converted[i] = got
	}
	return typedValue(typ, converted, path)
}

func typedValue(typ tftypes.Type, value any, path string) (tftypes.Value, error) {
	native := value
	if value == Unknown {
		native = tftypes.UnknownValue
	}
	if err := tftypes.ValidateValue(typ, native); err != nil {
		return tftypes.Value{}, pathError(path, "type mismatch")
	}
	return tftypes.NewValue(typ, native), nil
}

func scalarOK(typ tftypes.Type, value any) bool {
	switch {
	case typ.Is(tftypes.String):
		_, ok := value.(string)
		return ok
	case typ.Is(tftypes.Bool):
		_, ok := value.(bool)
		return ok
	case typ.Is(tftypes.Number):
		return assert.IsNumeric(value)
	default:
		return false
	}
}
