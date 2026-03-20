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

// Package excel 提供基于Schema的Excel文件校验和解析能力
package excel

import (
	"encoding/json"
)

// Schema 定义Excel模版的完整结构，与tpl_schema JSON对应
type Schema struct {
	Sheets     []Sheet           `json:"sheets"`
	sheetIndex map[string]*Sheet // built after unmarshal
}

// UnmarshalJSON implements json.Unmarshaler. It decodes JSON into a temporary
// struct to avoid infinite recursion, then builds sheetIndex for O(1) lookup.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var raw struct {
		Sheets []Sheet `json:"sheets"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Sheets = raw.Sheets
	s.sheetIndex = make(map[string]*Sheet, len(s.Sheets))
	for i := range s.Sheets {
		s.sheetIndex[s.Sheets[i].Name] = &s.Sheets[i]
	}

	return nil
}

// Sheet 定义单个sheet的结构
type Sheet struct {
	// Name sheet名称
	Name string `json:"name"`
	// HeadRow 表头所在行号（1-based）
	HeadRow int `json:"head_row"`
	// RowStart 数据起始行号（Excel中数据从第几行开始，1-based，不含表头）
	RowStart int `json:"row_start"`
	// FixedHeaders 固定列定义列表，用于排序和聚合等固定字段
	FixedHeaders []Header `json:"fixed_headers"`
	// Headers 动态列定义列表，用于业务扩展字段
	Headers []Header `json:"headers"`
}

// Header 定义单个列的结构
type Header struct {
	// Name 列展示名称，对应Excel表头
	Name string `json:"name"`
	// Type 列数据类型，枚举：string、int、float、enum
	Type string `json:"type"`
	// Field 列在Excel中的列号，如A、B、C；"-"表示无对应Excel列
	Field string `json:"field"`
	// DBField DB字段名，标识该列用于数据排序或聚合的key
	DBField string `json:"db_field,omitempty"`
	// Value 当Type为enum时，定义可选枚举值列表
	Value []interface{} `json:"value,omitempty"`
	// Hidden 是否在前端隐藏该列
	Hidden bool `json:"hidden,omitempty"`
	// Required 是否必填
	Required bool `json:"required,omitempty"`
	// Formula Excel公式，当该列由公式自动计算时提供
	Formula string `json:"formula,omitempty"`
	// Readonly 是否只读，为true时该列由公式自动计算，用户不可编辑
	Readonly bool `json:"readonly,omitempty"`
	// GT 严格大于约束（开区间下限）：string类型表示最小字符数，int/float类型表示数值下限，nil表示不校验
	GT *float64 `json:"gt,omitempty"`
	// GTE 大于等于约束（闭区间下限）：string类型表示最小字符数，int/float类型表示数值下限，nil表示不校验
	GTE *float64 `json:"gte,omitempty"`
	// LT 严格小于约束（开区间上限）：string类型表示最大字符数，int/float类型表示数值上限，nil表示不校验
	LT *float64 `json:"lt,omitempty"`
	// LTE 小于等于约束（闭区间上限）：string类型表示最大字符数，int/float类型表示数值上限，nil表示不校验
	LTE *float64 `json:"lte,omitempty"`
}

// FindSheet looks up a Sheet by name. Returns nil if not found.
// Uses the pre-built index for O(1) lookup when the schema was decoded via JSON;
// falls back to a linear scan for schemas constructed manually.
func (s *Schema) FindSheet(name string) *Sheet {
	if s.sheetIndex != nil {
		return s.sheetIndex[name]
	}

	for i := range s.Sheets {
		if s.Sheets[i].Name == name {
			return &s.Sheets[i]
		}
	}

	return nil
}

// SheetNames 返回Schema中所有sheet的名称列表
func (s *Schema) SheetNames() []string {
	names := make([]string, 0, len(s.Sheets))
	for _, sheet := range s.Sheets {
		names = append(names, sheet.Name)
	}

	return names
}

// AllExcelHeaders 返回FixedHeaders和Headers中所有有对应Excel列（field != "-"）的Header列表
func (s *Sheet) AllExcelHeaders() []Header {
	headers := make([]Header, 0, len(s.FixedHeaders)+len(s.Headers))
	for _, h := range s.FixedHeaders {
		if h.Field != "-" {
			headers = append(headers, h)
		}
	}
	for _, h := range s.Headers {
		if h.Field != "-" {
			headers = append(headers, h)
		}
	}

	return headers
}

// AllVisibleHeaders 返回FixedHeaders和Headers中field != "-"且hidden != true的Header列表
func (s *Sheet) AllVisibleHeaders() []Header {
	headers := make([]Header, 0, len(s.FixedHeaders)+len(s.Headers))
	for _, h := range s.FixedHeaders {
		if h.Field != "-" && !h.Hidden {
			headers = append(headers, h)
		}
	}
	for _, h := range s.Headers {
		if h.Field != "-" && !h.Hidden {
			headers = append(headers, h)
		}
	}

	return headers
}
