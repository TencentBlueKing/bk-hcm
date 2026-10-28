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

// Package adaptor 实现本地 tf_provider 插件北向运行态适配与协议数据转换。
package adaptor

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"

	"hcm/cmd/hc-service/logics/tfruntime/codec"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/plugin/tfplugin5"
	cvt "hcm/pkg/tools/converter"

	tfjson "github.com/hashicorp/terraform-json"
)

// decodePrivate 将北向请求中的标准 Base64 编码字符串解码为原始 Private 字节切片；
// 当指针为 nil 或空字符串时直接返回 nil, nil。
func decodePrivate(b64 *string) ([]byte, error) {
	if b64 == nil {
		return nil, nil
	}
	rawStr := cvt.PtrToVal(b64)
	if len(rawStr) == 0 {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(rawStr)
	if err != nil {
		return nil, errf.Newf(errf.InvalidParameter, "invalid base64 private: %v", err)
	}
	return decoded, nil
}

// encodePrivate 将底层插件返回的原始 Private 字节切片编码为标准 Base64 字符串指针；
// 空切片直接返回 nil，避免向北向响应输出无意义的空字符串。
func encodePrivate(raw []byte) *string {
	if len(raw) == 0 {
		return nil
	}
	return cvt.ValToPtr(base64.StdEncoding.EncodeToString(raw))
}

// convertDiagnostics 将底层 tfplugin5 gRPC 返回的 protobuf 诊断列表转换为北向统一的 coreplugin.Diagnostic 列表。
func convertDiagnostics(executorDiags []*tfplugin5.Diagnostic) []coreplugin.Diagnostic {
	if len(executorDiags) == 0 {
		return nil
	}
	out := make([]coreplugin.Diagnostic, 0, len(executorDiags))
	for _, d := range executorDiags {
		if d == nil {
			continue
		}
		severity := tfjson.DiagnosticSeverityError
		switch d.GetSeverity() {
		case tfplugin5.Diagnostic_WARNING:
			severity = tfjson.DiagnosticSeverityWarning
		case tfplugin5.Diagnostic_ERROR:
			severity = tfjson.DiagnosticSeverityError
		default:
			severity = tfjson.DiagnosticSeverityError
		}

		summary := d.GetSummary()
		detail := d.GetDetail()
		msg := summary
		if summary != "" && detail != "" {
			msg = summary + ": " + detail
		} else if summary == "" {
			msg = detail
		}

		out = append(out, coreplugin.Diagnostic{
			Diagnostic: tfjson.Diagnostic{
				Severity: severity,
				Summary:  summary,
				Detail:   detail,
			},
			Path:    convertAttributePathSteps(d.GetAttribute()),
			Message: msg,
		})
	}
	return out
}

// convertAttributePathSteps 将 protobuf AttributePath 的 Steps 映射为北向三选一的 PathStep 切片。
func convertAttributePathSteps(attrPath *tfplugin5.AttributePath) []coreplugin.PathStep {
	steps := attrPath.GetSteps()
	if len(steps) == 0 {
		return []coreplugin.PathStep{}
	}
	out := make([]coreplugin.PathStep, 0, len(steps))
	for _, step := range steps {
		if step == nil {
			continue
		}
		switch sel := step.GetSelector().(type) {
		case *tfplugin5.AttributePath_Step_AttributeName:
			out = append(out, coreplugin.NewAttributeStep(sel.AttributeName))
		case *tfplugin5.AttributePath_Step_ElementKeyString:
			out = append(out, coreplugin.NewKeyStep(sel.ElementKeyString))
		case *tfplugin5.AttributePath_Step_ElementKeyInt:
			out = append(out, coreplugin.NewIndexStep(int(sel.ElementKeyInt)))
		default:
		}
	}
	return out
}

// extractErrorSummary 提取转换后诊断列表中所有 Error 级别诊断的 Message，以 "; " 拼接作为顶层失败摘要。
func extractErrorSummary(diags []coreplugin.Diagnostic) string {
	var parts []string
	for _, d := range diags {
		if d.EffectiveSeverity() != tfjson.DiagnosticSeverityError {
			continue
		}
		text := d.Message
		if text == "" {
			text = d.Summary
		}
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "; ")
}

// newErrorDiagnostic 构造一条宿主侧主动封装的 Error 级别北向 Diagnostic。
func newErrorDiagnostic(summary, detail string) coreplugin.Diagnostic {
	msg := summary
	if summary != "" && detail != "" {
		msg = summary + ": " + detail
	} else if summary == "" {
		msg = detail
	}
	return coreplugin.Diagnostic{
		Diagnostic: tfjson.Diagnostic{
			Severity: tfjson.DiagnosticSeverityError,
			Summary:  summary,
			Detail:   detail,
		},
		Path:    []coreplugin.PathStep{},
		Message: msg,
	}
}

// convertReplacePaths 将 PlanResourceChange 返回的 RequiresReplace 属性路径转换为点分与下标格式的字符串切片。
// 例如 [AttributeName("disks"), ElementKeyInt(0), AttributeName("size")] 转换为 "disks[0].size"。
func convertReplacePaths(paths []*tfplugin5.AttributePath) []string {
	if len(paths) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		steps := p.GetSteps()
		if len(steps) == 0 {
			continue
		}
		var sb strings.Builder
		for _, step := range steps {
			if step == nil {
				continue
			}
			switch sel := step.GetSelector().(type) {
			case *tfplugin5.AttributePath_Step_AttributeName:
				if sb.Len() > 0 {
					sb.WriteByte('.')
				}
				sb.WriteString(sel.AttributeName)
			case *tfplugin5.AttributePath_Step_ElementKeyString:
				if sb.Len() > 0 {
					sb.WriteByte('.')
				}
				sb.WriteString(sel.ElementKeyString)
			case *tfplugin5.AttributePath_Step_ElementKeyInt:
				sb.WriteString(fmt.Sprintf("[%d]", sel.ElementKeyInt))
			default:
			}
		}
		if sb.Len() > 0 {
			out = append(out, sb.String())
		}
	}
	return out
}

// nullifyUnknowns 递归遍历状态树（map[string]any 与 []any），将底层解码出的 codec.Unknown 哨兵值置为 nil，
// 同时将 *big.Float 归一化为标准 JSON 数值类型（int64、uint64 或 float64），并返回原状态树中是否包含 codec.Unknown。
func nullifyUnknowns(state map[string]any) (map[string]any, bool) {
	if state == nil {
		return nil, false
	}
	cleaned, hadUnknown := nullifyUnknownValue(state)
	cleanedMap, _ := cleaned.(map[string]any)
	return cleanedMap, hadUnknown
}

func nullifyUnknownValue(val any) (any, bool) {
	if val == codec.Unknown {
		return nil, true
	}
	switch typed := val.(type) {
	case *big.Float:
		return normalizeBigFloat(typed), false
	case map[string]any:
		if typed == nil {
			return nil, false
		}
		out := make(map[string]any, len(typed))
		anyUnknown := false
		for k, v := range typed {
			cleanedChild, childUnknown := nullifyUnknownValue(v)
			if childUnknown {
				anyUnknown = true
			}
			out[k] = cleanedChild
		}
		return out, anyUnknown
	case []any:
		if typed == nil {
			return nil, false
		}
		out := make([]any, len(typed))
		anyUnknown := false
		for i, elem := range typed {
			cleanedElem, elemUnknown := nullifyUnknownValue(elem)
			if elemUnknown {
				anyUnknown = true
			}
			out[i] = cleanedElem
		}
		return out, anyUnknown
	default:
		return val, false
	}
}

// normalizeBigFloat 将底层解码出的 *big.Float 转换为标准 JSON 数值类型（int64、uint64 或 float64），
// 避免 encoding/json 因 *big.Float 实现 encoding.TextMarshaler 而将其序列化为带引号的字符串。
func normalizeBigFloat(f *big.Float) any {
	if f == nil {
		return nil
	}
	if f.IsInt() {
		if i64, acc := f.Int64(); acc == big.Exact {
			return i64
		}
		if u64, acc := f.Uint64(); acc == big.Exact {
			return normalizeUint64(u64)
		}
	}
	f64, _ := f.Float64()
	return f64
}

func normalizeUint64(u uint64) any {
	if u <= math.MaxInt64 {
		return int64(u)
	}
	return u
}

// stateEqual 递归比较两份状态树在结构与数值上是否相等，抹平 JSON 反序列化的 float64 与插件解码的 *big.Float 类型差异。
func stateEqual(a, b any) bool {
	if na, ok := normalizeNumeric(a); ok {
		nb, okB := normalizeNumeric(b)
		return okB && na == nb
	}
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok || (va == nil) != (vb == nil) || len(va) != len(vb) {
			return false
		}
		for k, elemA := range va {
			elemB, exists := vb[k]
			if !exists || !stateEqual(elemA, elemB) {
				return false
			}
		}
		return true
	case []any:
		vb, ok := b.([]any)
		if !ok || (va == nil) != (vb == nil) || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !stateEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}

func normalizeNumeric(v any) (any, bool) {
	switch n := v.(type) {
	case *big.Float:
		if n == nil {
			return nil, false
		}
		return normalizeBigFloat(n), true
	case float64:
		if !math.IsNaN(n) && !math.IsInf(n, 0) && math.Trunc(n) == n &&
			n >= float64(math.MinInt64) && n < float64(1<<63) {
			return int64(n), true
		}
		return n, true
	case float32:
		return normalizeNumeric(float64(n))
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return normalizeUint64(uint64(n)), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return normalizeUint64(n), true
	default:
		return nil, false
	}
}
