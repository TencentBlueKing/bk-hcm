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
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"

	"github.com/zclconf/go-cty/cty"
)

// goToCty 把外部传入的 Go 值转成 HCL 求值需要的 cty.Value。
//
// 举例：
//   - nil -> cty.NullVal(cty.DynamicPseudoType)
//   - "ap-guangzhou" -> cty.StringVal("ap-guangzhou")
//   - 213 或 float64(213) -> cty.NumberIntVal(213)
//   - []any{"a", 1} -> cty.TupleVal
//   - map[string]any{"k": "v"} -> cty.ObjectVal
func goToCty(val any) (cty.Value, error) {
	if val == nil {
		return cty.NullVal(cty.DynamicPseudoType), nil
	}
	if cv, ok := val.(cty.Value); ok {
		return cv, nil
	}
	switch v := val.(type) {
	case bool:
		return cty.BoolVal(v), nil
	case *bool:
		if v == nil {
			return cty.NullVal(cty.Bool), nil
		}
		return cty.BoolVal(cvt.PtrToVal(v)), nil
	case string:
		return cty.StringVal(v), nil
	case *string:
		if v == nil {
			return cty.NullVal(cty.String), nil
		}
		return cty.StringVal(cvt.PtrToVal(v)), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number, *int64:
		return goNumericToCty(v)
	default:
		return goCollectionToCty(v)
	}
}

// goNumericToCty 把 Go 整数、浮点数和 json.Number 转成 cty.Number。
func goNumericToCty(val any) (cty.Value, error) {
	switch v := val.(type) {
	case int:
		return cty.NumberIntVal(int64(v)), nil
	case int8:
		return cty.NumberIntVal(int64(v)), nil
	case int16:
		return cty.NumberIntVal(int64(v)), nil
	case int32:
		return cty.NumberIntVal(int64(v)), nil
	case int64:
		return cty.NumberIntVal(v), nil
	case *int64:
		if v == nil {
			return cty.NullVal(cty.Number), nil
		}
		return cty.NumberIntVal(cvt.PtrToVal(v)), nil
	case uint:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint8:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint16:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint32:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint64:
		return cty.NumberUIntVal(v), nil
	case float32:
		return goFloat64ToCty(float64(v)), nil
	case float64:
		return goFloat64ToCty(v), nil
	case json.Number:
		if i64, err := v.Int64(); err == nil {
			return cty.NumberIntVal(i64), nil
		}
		return cty.ParseNumberVal(v.String())
	default:
		return cty.NilVal, fmt.Errorf("unsupported numeric type %T", val)
	}
}

// goFloat64ToCty 把 float64 转成 cty.Number。
// JSON 反序列化到 any 时数字默认是 float64（比如业务 ID 213 会变成 float64(213)），
// 这里如果发现没有小数部分且在 int64 范围内，就转回整数，避免拼字符串时带出小数点。
// 注意 float64(math.MaxInt64) 会进位成 2^63 导致 int64 溢出，所以上界要用小于号。
func goFloat64ToCty(v float64) cty.Value {
	if !math.IsNaN(v) && !math.IsInf(v, 0) && math.Trunc(v) == v &&
		v >= float64(math.MinInt64) && v < float64(math.MaxInt64) {
		return cty.NumberIntVal(int64(v))
	}
	return cty.NumberFloatVal(v)
}

// goCollectionToCty 把 Go 切片或 map 转成 cty.Tuple 或 cty.Object。
func goCollectionToCty(val any) (cty.Value, error) {
	switch v := val.(type) {
	case []any:
		return sliceAnyToCtyTuple(v)
	case []string:
		elems := make([]cty.Value, len(v))
		for i, s := range v {
			elems[i] = cty.StringVal(s)
		}
		return cty.TupleVal(elems), nil
	case []map[string]any:
		elems := make([]cty.Value, len(v))
		for i, m := range v {
			cv, err := mapAnyToCtyObject(m)
			if err != nil {
				return cty.NilVal, err
			}
			elems[i] = cv
		}
		return cty.TupleVal(elems), nil
	case map[string]any:
		return mapAnyToCtyObject(v)
	case map[string]string:
		obj := make(map[string]cty.Value, len(v))
		for k, s := range v {
			obj[k] = cty.StringVal(s)
		}
		return cty.ObjectVal(obj), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported Go type %T for cty conversion", val)
	}
}

// sliceAnyToCtyTuple 把 []any 递归转成 cty.Tuple，支持切片里混放不同类型元素。
func sliceAnyToCtyTuple(s []any) (cty.Value, error) {
	if len(s) == 0 {
		return cty.EmptyTupleVal, nil
	}
	elems := make([]cty.Value, len(s))
	for i, item := range s {
		cv, err := goToCty(item)
		if err != nil {
			return cty.NilVal, err
		}
		elems[i] = cv
	}
	return cty.TupleVal(elems), nil
}

// mapAnyToCtyObject 把 map[string]any 递归转成 cty.Object。
func mapAnyToCtyObject(m map[string]any) (cty.Value, error) {
	if len(m) == 0 {
		return cty.EmptyObjectVal, nil
	}
	obj := make(map[string]cty.Value, len(m))
	for k, item := range m {
		cv, err := goToCty(item)
		if err != nil {
			return cty.NilVal, err
		}
		obj[k] = cv
	}
	return cty.ObjectVal(obj), nil
}

// ctyToGo 把求值后的 cty.Value 转回普通的 Go 原生类型，方便直接做 JSON 序列化。
//
// 举例：
//   - cty.NullVal -> nil
//   - cty.StringVal("vpc-123") -> "vpc-123"
//   - cty.NumberIntVal(213) -> int64(213)
//   - cty.NumberFloatVal(3.14) -> float64(3.14)
//   - Tuple / List / Set -> []any
//   - Object / Map -> map[string]any
func ctyToGo(val cty.Value) (any, error) {
	if !val.IsKnown() {
		return nil, fmt.Errorf("cannot convert unknown cty.Value to Go value")
	}
	if val.IsNull() {
		return nil, nil
	}
	ty := val.Type()
	switch {
	case ty == cty.String:
		return val.AsString(), nil
	case ty == cty.Bool:
		return val.True(), nil
	case ty == cty.Number:
		return ctyNumberToGo(val), nil
	case ty.IsTupleType() || ty.IsListType() || ty.IsSetType():
		return ctySequenceToGoSlice(val)
	case ty.IsObjectType() || ty.IsMapType():
		return ctyMappingToGoMap(val)
	default:
		return nil, fmt.Errorf("unsupported cty.Type %s", ty.FriendlyName())
	}
}

// ctyNumberToGo 把 cty.Number 转成 Go 数字：能精确转成 int64 的就返回 int64，否则返回 float64。
func ctyNumberToGo(val cty.Value) any {
	bf := val.AsBigFloat()
	if bf.IsInt() {
		if i64, acc := bf.Int64(); acc == big.Exact {
			return i64
		}
	}
	f64, _ := bf.Float64()
	return f64
}

// ctySequenceToGoSlice 把 cty 的列表、元组或集合转成 []any。
func ctySequenceToGoSlice(val cty.Value) ([]any, error) {
	length := val.LengthInt()
	if length == 0 {
		return []any{}, nil
	}
	res := make([]any, 0, length)
	it := val.ElementIterator()
	for it.Next() {
		_, elemVal := it.Element()
		gv, err := ctyToGo(elemVal)
		if err != nil {
			return nil, err
		}
		res = append(res, gv)
	}
	return res, nil
}

// ctyMappingToGoMap 把 cty 的 Object 或 Map 转成 map[string]any。
func ctyMappingToGoMap(val cty.Value) (map[string]any, error) {
	length := val.LengthInt()
	if length == 0 {
		return map[string]any{}, nil
	}
	res := make(map[string]any, length)
	it := val.ElementIterator()
	for it.Next() {
		kVal, vVal := it.Element()
		gv, err := ctyToGo(vVal)
		if err != nil {
			return nil, err
		}
		res[kVal.AsString()] = gv
	}
	return res, nil
}

// ctyValueToHCLString 把阶段一已经算出来的静态 cty.Value 格式化成 HCL 字面量字符串。
// 当一个表达式里既有静态值又有动态引用时（比如 format("%s-%s", var.env, tencentcloud_vpc.main.id)），
// 需要把静态的 var.env 先替换成 "prod" 这种字面量写进 PendingExprs。
//
// 举例：
//   - cty.StringVal("prod") -> "\"prod\""
//   - cty.NumberIntVal(213) -> "213"
//   - cty.TupleVal(["a", "b"]) -> "[\"a\", \"b\"]"
//   - cty.ObjectVal({"env": "prod"}) -> "{\"env\" = \"prod\"}"
func ctyValueToHCLString(val cty.Value) (string, error) {
	if !val.IsKnown() {
		return "", fmt.Errorf("cannot format unknown cty.Value as HCL literal")
	}
	if val.IsNull() {
		return "null", nil
	}
	ty := val.Type()
	switch {
	case ty == cty.String:
		return strconv.Quote(val.AsString()), nil
	case ty == cty.Bool:
		if val.True() {
			return "true", nil
		}
		return "false", nil
	case ty == cty.Number:
		gv := ctyNumberToGo(val)
		if i64, ok := gv.(int64); ok {
			return strconv.FormatInt(i64, 10), nil
		}
		return strconv.FormatFloat(gv.(float64), 'f', -1, 64), nil
	case ty.IsTupleType() || ty.IsListType() || ty.IsSetType():
		return ctySeqToHCLString(val)
	case ty.IsObjectType() || ty.IsMapType():
		return ctyMapToHCLString(val)
	default:
		return "", fmt.Errorf("unsupported cty.Type %s for HCL literal", ty.FriendlyName())
	}
}

// ctySeqToHCLString 把列表或元组格式化成 [elem1, elem2] 形式的 HCL 字符串。
func ctySeqToHCLString(val cty.Value) (string, error) {
	parts := make([]string, 0, val.LengthInt())
	it := val.ElementIterator()
	for it.Next() {
		_, elem := it.Element()
		s, err := ctyValueToHCLString(elem)
		if err != nil {
			return "", err
		}
		parts = append(parts, s)
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

// ctyMapToHCLString 把对象或 map 按 key 字典序格式化成 {"k1" = v1, "k2" = v2} 形式的 HCL 字符串。
func ctyMapToHCLString(val cty.Value) (string, error) {
	elemMap := val.AsValueMap()
	if len(elemMap) == 0 {
		return "{}", nil
	}
	keys := maps.Keys(elemMap)
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		s, err := ctyValueToHCLString(elemMap[k])
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s = %s", strconv.Quote(k), s))
	}
	return "{" + strings.Join(parts, ", ") + "}", nil
}

// deepCopyMap 对 map[string]any 做递归深拷贝，防止阶段二 EvalBlock 修改调用方传入的原始 Config。
func deepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	cloned := maps.Clone(src)
	for k, v := range cloned {
		cloned[k] = deepCopyAny(v)
	}
	return cloned
}

// deepCopyAny 递归拷贝嵌套的 map 或切片。
func deepCopyAny(val any) any {
	switch v := val.(type) {
	case map[string]any:
		return deepCopyMap(v)
	case []any:
		cp := make([]any, len(v))
		for i, item := range v {
			cp[i] = deepCopyAny(item)
		}
		return cp
	case []map[string]any:
		cp := make([]map[string]any, len(v))
		for i, m := range v {
			cp[i] = deepCopyMap(m)
		}
		return cp
	case []string:
		cp := make([]string, len(v))
		copy(cp, v)
		return cp
	default:
		return val
	}
}
