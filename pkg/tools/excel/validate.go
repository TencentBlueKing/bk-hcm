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

package excel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// valueValidator 定义统一的类型校验器接口，同时支持string（Excel单元格）
// 和interface{}（JSON反序列化）两种值来源，便于扩展新类型
type valueValidator interface {
	Validate(val interface{}, header Header) []string
}

// validatorRegistry 按header type映射对应的校验器
var validatorRegistry = map[string]valueValidator{
	"int":    intValidator{},
	"float":  floatValidator{},
	"enum":   enumValidator{},
	"string": stringValidator{},
}

// ValidateCellValue 根据Header定义校验Excel单元格字符串值，返回面向前端的中文校验描述列表（非Go error）。
// 校验顺序：required → 类型校验 → 范围/长度校验。
// 空值（TrimSpace后）在非required场景下跳过所有校验。
func ValidateCellValue(val string, header Header) []string {
	trimmed := strings.TrimSpace(val)

	if trimmed == "" {
		if header.Required {
			return []string{fmt.Sprintf(constant.ExcelValidateRequiredEmpty, header.Name)}
		}
		return nil
	}

	v, ok := validatorRegistry[header.Type]
	if !ok {
		return nil
	}

	return v.Validate(trimmed, header)
}

// ValidateTypedValue 根据Header定义校验JSON反序列化后的interface{}值，返回面向前端的中文校验描述列表。
// 与ValidateCellValue共享同一套校验器，区别仅在于空值判定逻辑（nil/空字符串）。
// 入参先经 normalizeJSONNumber 归一化，确保 json.Number（UseNumber模式）统一转为 float64，
// 后续所有校验器无需感知 json.Number 类型。
func ValidateTypedValue(val interface{}, header Header) []string {
	val = normalizeJSONNumber(val)

	if isEmptyValue(val) {
		if header.Required {
			return []string{fmt.Sprintf(constant.ExcelValidateRequiredEmpty, header.Name)}
		}
		return nil
	}

	v, ok := validatorRegistry[header.Type]
	if !ok {
		logs.Warnf("validatorRegistry[%s] not found, header: %+v", header.Type, header)
		return nil
	}

	return v.Validate(val, header)
}

// normalizeJSONNumber 将 json.Number（jsoniter UseNumber 模式下的反序列化产物）转为 float64，
// 使下游校验器只需处理 string 和 float64 两种数值表示，无需感知 JSON 库的具体实现。
// 若 json.Number 无法解析为 float64，则退化为其字符串表示交由后续校验器处理。
// 其他类型原样返回。
func normalizeJSONNumber(v interface{}) interface{} {
	n, ok := v.(json.Number)
	if !ok {
		return v
	}

	fv, err := n.Float64()
	if err != nil {
		return n.String()
	}

	return fv
}

// ValidateExtension 校验extension数组中的每个值，按索引与headers一一对应。
// 用于CreateGpuDemandOrder等场景，extension为JSON反序列化后的[]interface{}。
func ValidateExtension(values []interface{}, headers []Header) []string {
	var errs []string

	for i, h := range headers {
		var val interface{}
		if i < len(values) {
			val = values[i]
		}

		if vErrs := ValidateTypedValue(val, h); len(vErrs) > 0 {
			errs = append(errs, vErrs...)
		}
	}

	return errs
}

// ValidateFixedFields 校验固定字段值，按headers中每个Header的DBField从values map中查找对应值，
// 调用ValidateTypedValue进行类型和范围校验。DBField为空的Header将被跳过。
func ValidateFixedFields(values map[string]interface{}, headers []Header) []string {
	var errs []string

	for _, h := range headers {
		if h.DBField == "" {
			continue
		}

		if vErrs := ValidateTypedValue(values[h.DBField], h); len(vErrs) > 0 {
			errs = append(errs, vErrs...)
		}
	}

	return errs
}

func isEmptyValue(val interface{}) bool {
	if val == nil {
		return true
	}
	s, ok := val.(string)

	return ok && strings.TrimSpace(s) == ""
}

// --- int validator ---

type intValidator struct{}

// Validate 校验值是否为合法整数，并检查是否满足GT/GTE/LT/LTE约束
func (intValidator) Validate(val interface{}, header Header) []string {
	switch v := val.(type) {
	case string:
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return []string{fmt.Sprintf(constant.ExcelValidateMustBeInt, header.Name)}
		}
		return checkIntConstraints(iv, header)
	case float64:
		iv := int64(v)
		if v != float64(iv) {
			return []string{fmt.Sprintf(constant.ExcelValidateMustBeInt, header.Name)}
		}
		return checkIntConstraints(iv, header)
	default:
		return []string{fmt.Sprintf(constant.ExcelValidateMustBeInt, header.Name)}
	}
}

// checkIntConstraints 校验整数值是否满足 GT/GTE/LT/LTE 约束，返回所有未通过的校验描述
func checkIntConstraints(v int64, h Header) []string {
	fv := float64(v)
	var errs []string

	if h.GT != nil && fv <= converter.PtrToVal(h.GT) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateIntGT, h.Name, v, formatFloat(converter.PtrToVal(h.GT))))
	}
	if h.GTE != nil && fv < converter.PtrToVal(h.GTE) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateIntGTE, h.Name, v, formatFloat(converter.PtrToVal(h.GTE))))
	}
	if h.LT != nil && fv >= converter.PtrToVal(h.LT) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateIntLT, h.Name, v, formatFloat(converter.PtrToVal(h.LT))))
	}
	if h.LTE != nil && fv > converter.PtrToVal(h.LTE) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateIntLTE, h.Name, v, formatFloat(converter.PtrToVal(h.LTE))))
	}

	return errs
}

// --- float validator ---

type floatValidator struct{}

// Validate 校验值是否为合法浮点数，并检查是否满足GT/GTE/LT/LTE约束
func (floatValidator) Validate(val interface{}, header Header) []string {
	switch v := val.(type) {
	case string:
		fv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return []string{fmt.Sprintf(constant.ExcelValidateMustBeNumber, header.Name)}
		}
		return checkFloatConstraints(fv, header)
	case float64:
		return checkFloatConstraints(v, header)
	default:
		return []string{fmt.Sprintf(constant.ExcelValidateMustBeNumber, header.Name)}
	}
}

// checkFloatConstraints 校验浮点值是否满足 GT/GTE/LT/LTE 约束，返回所有未通过的校验描述
func checkFloatConstraints(v float64, h Header) []string {
	sv := formatFloat(v)
	var errs []string

	if h.GT != nil && v <= converter.PtrToVal(h.GT) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateFloatGT, h.Name, sv, formatFloat(converter.PtrToVal(h.GT))))
	}
	if h.GTE != nil && v < converter.PtrToVal(h.GTE) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateFloatGTE, h.Name, sv, formatFloat(converter.PtrToVal(h.GTE))))
	}
	if h.LT != nil && v >= converter.PtrToVal(h.LT) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateFloatLT, h.Name, sv, formatFloat(converter.PtrToVal(h.LT))))
	}
	if h.LTE != nil && v > converter.PtrToVal(h.LTE) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateFloatLTE, h.Name, sv, formatFloat(converter.PtrToVal(h.LTE))))
	}

	return errs
}

// --- enum validator ---

type enumValidator struct{}

// Validate 校验值是否在Header.Value定义的枚举列表内，支持数字和字符串枚举
func (enumValidator) Validate(val interface{}, header Header) []string {
	if len(header.Value) == 0 {
		return nil
	}

	switch normalizeJSONNumber(header.Value[0]).(type) {
	case float64:
		return validateNumericEnum(val, header)
	default:
		return validateStringEnum(val, header)
	}
}

// validateNumericEnum 将 val 统一转为 float64 后与枚举表比对。
// 支持 string（Excel 单元格）和 float64（JSON 反序列化后经 normalizeJSONNumber 归一化）两种来源。
func validateNumericEnum(val interface{}, header Header) []string {
	var fv float64
	var display string
	switch v := val.(type) {
	case string:
		var err error
		fv, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return []string{fmt.Sprintf(constant.ExcelValidateEnumTypeMismatch, header.Name, val)}
		}
		display = v
	case float64:
		fv = v
		display = formatFloat(v)
	default:
		return []string{fmt.Sprintf(constant.ExcelValidateEnumTypeMismatch, header.Name, val)}
	}

	for _, ev := range header.Value {
		if enumFloat, ok := normalizeJSONNumber(ev).(float64); ok && fv == enumFloat {
			return nil
		}
	}

	return []string{
		fmt.Sprintf(constant.ExcelValidateEnumNotInRange, header.Name, display, formatEnumValues(header.Value)),
	}
}

func validateStringEnum(val interface{}, header Header) []string {
	sv, ok := val.(string)
	if !ok {
		return []string{fmt.Sprintf("%s: expected string value, got %T", header.Name, val)}
	}

	for _, ev := range header.Value {
		if fmt.Sprint(ev) == sv {
			return nil
		}
	}

	return []string{
		fmt.Sprintf(constant.ExcelValidateEnumNotInRange,
			header.Name, sv, formatEnumValues(header.Value)),
	}
}

func formatEnumValues(values []interface{}) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		switch tv := normalizeJSONNumber(v).(type) {
		case float64:
			parts = append(parts, formatFloat(tv))
		default:
			parts = append(parts, fmt.Sprint(tv))
		}
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// --- string validator ---

type stringValidator struct{}

// Validate 校验值是否为字符串，并检查字符长度是否满足GT/GTE/LT/LTE约束
func (stringValidator) Validate(val interface{}, header Header) []string {
	sv, ok := val.(string)
	if !ok {
		return []string{fmt.Sprintf(constant.ExcelValidateMustBeString, header.Name)}
	}

	return checkLengthConstraints(utf8.RuneCountInString(sv), header)
}

// checkLengthConstraints 校验字符串字符数是否满足 GT/GTE/LT/LTE 约束，返回所有未通过的校验描述
func checkLengthConstraints(length int, h Header) []string {
	fl := float64(length)
	var errs []string

	if h.GT != nil && fl <= converter.PtrToVal(h.GT) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateStrLenGT, h.Name, length, formatFloat(converter.PtrToVal(h.GT))))
	}
	if h.GTE != nil && fl < converter.PtrToVal(h.GTE) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateStrLenGTE, h.Name, length, formatFloat(converter.PtrToVal(h.GTE))))
	}
	if h.LT != nil && fl >= converter.PtrToVal(h.LT) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateStrLenLT, h.Name, length, formatFloat(converter.PtrToVal(h.LT))))
	}
	if h.LTE != nil && fl > converter.PtrToVal(h.LTE) {
		errs = append(errs,
			fmt.Sprintf(constant.ExcelValidateStrLenLTE, h.Name, length, formatFloat(converter.PtrToVal(h.LTE))))
	}

	return errs
}

// --- helpers ---

func formatFloat(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}

	return strconv.FormatFloat(v, 'f', -1, 64)
}
