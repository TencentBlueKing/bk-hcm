/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package table

import (
	"testing"
)

// TestTableField test table field
type TestTableField struct {
	Field1 string  `json:"field1" header:"字段1"`
	Field2 string  `json:"field2" header:"字段2;field2"`
	Field3 string  `json:"field3" header:"字段3"`
	Field4 *string `json:"field4" header:"字段4;field4"`
}

// GetHeaders 获取表头列
func (t TestTableField) GetHeaders() ([][]string, error) {
	return GetHeaders(t)
}

// GetValuesByHeader 根据表头字段顺序解析数据
func (t TestTableField) GetValuesByHeader() ([]string, error) {
	return GetValuesByHeader(t)
}

// TestTable ...
func TestTable(t *testing.T) {
	inst1Field4 := "inst1-4"
	inst2Field4 := "inst2-4"
	insts := []TestTableField{
		{
			Field1: "inst1-1",
			Field2: "inst1-2",
			Field3: "inst1-3",
			Field4: &inst1Field4,
		},
		{
			Field1: "inst2-1",
			Field2: "inst2-2",
			Field3: "inst2-3",
			Field4: &inst2Field4,
		},
	}

	// test header
	expectHeader := [][]string{
		{"字段1", "字段2", "字段3", "字段4"},
		{"", "field2", "", "field4"},
	}
	headers, err := insts[0].GetHeaders()
	if err != nil {
		t.Errorf("get headers err: %v", err)
		return
	}
	t.Logf("headers: %v", headers)

	if len(headers) != len(expectHeader) {
		t.Errorf("headers len not equal, expect: %d, actual: %d", len(expectHeader), len(headers))
		return
	}
	for i := 0; i < len(headers); i++ {
		if len(headers[i]) != len(expectHeader[i]) {
			t.Errorf("headers len not equal, row: %d, expect: %d, actual: %d", i, len(expectHeader[i]),
				len(headers[i]))
			return
		}
		for j := 0; j < len(headers[i]); j++ {
			if headers[i][j] != expectHeader[i][j] {
				t.Errorf("headers not equal, row: %d, col: %d, expect: %s, actual: %s", i, j,
					expectHeader[i][j], headers[i][j])
				return
			}
		}
	}

	// test values
	inst1Vals, err := insts[0].GetValuesByHeader()
	if err != nil {
		t.Errorf("get values failed, err: %v", err)
		return
	}
	t.Logf("values: %v", inst1Vals)
	expectInst1Vals := []string{"inst1-1", "inst1-2", "inst1-3", "inst1-4"}
	if len(inst1Vals) != len(expectInst1Vals) {
		t.Errorf("values len not equal, expect: %d, actual: %d", len(expectInst1Vals), len(inst1Vals))
		return
	}
	for i := 0; i < len(inst1Vals); i++ {
		if inst1Vals[i] != expectInst1Vals[i] {
			t.Errorf("values not equal, row: %d, expect: %s, actual: %s", i, expectInst1Vals[i], inst1Vals[i])
			return
		}
	}

	inst2Vals, err := insts[1].GetValuesByHeader()
	if err != nil {
		t.Errorf("get values failed, err: %v", err)
		return
	}
	t.Logf("values: %v", inst2Vals)
	expectInst2Vals := []string{"inst2-1", "inst2-2", "inst2-3", "inst2-4"}
	if len(inst2Vals) != len(expectInst2Vals) {
		t.Errorf("values len not equal, expect: %d, actual: %d", len(expectInst2Vals), len(inst2Vals))
		return
	}
	for i := 0; i < len(inst2Vals); i++ {
		if inst2Vals[i] != expectInst2Vals[i] {
			t.Errorf("values not equal, row: %d, expect: %s, actual: %s", i, expectInst2Vals[i], inst2Vals[i])
			return
		}
	}
}
