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
	"testing"
)

func TestSchema_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantSheets []string
	}{
		{
			name: "valid schema with two sheets",
			input: `{
				"sheets": [
					{"name": "SheetA", "head_row": 1, "row_start": 2, "fixed_headers": [], "headers": []},
					{"name": "SheetB", "head_row": 1, "row_start": 2, "fixed_headers": [], "headers": []}
				]
			}`,
			wantSheets: []string{"SheetA", "SheetB"},
		},
		{
			name:       "empty sheets array",
			input:      `{"sheets": []}`,
			wantSheets: []string{},
		},
		{
			name:    "invalid json",
			input:   `{not valid json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(Schema)
			err := json.Unmarshal([]byte(tt.input), s)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(s.Sheets) != len(tt.wantSheets) {
				t.Fatalf("got %d sheets, want %d", len(s.Sheets), len(tt.wantSheets))
			}

			for _, name := range tt.wantSheets {
				if s.FindSheet(name) == nil {
					t.Errorf("FindSheet(%q) = nil, want non-nil", name)
				}
			}
		})
	}
}

func TestSchema_FindSheet(t *testing.T) {
	const raw = `{
		"sheets": [
			{
				"name": "GPU-A",
				"head_row": 1,
				"row_start": 2,
				"fixed_headers": [{"name": "年份", "type": "int", "field": "A"}],
				"headers": []
			},
			{
				"name": "GPU-B",
				"head_row": 1,
				"row_start": 2,
				"fixed_headers": [],
				"headers": [{"name": "备注", "type": "string", "field": "B"}]
			}
		]
	}`

	s := new(Schema)
	if err := json.Unmarshal([]byte(raw), s); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	t.Run("existing sheet returns correct pointer", func(t *testing.T) {
		sheet := s.FindSheet("GPU-A")
		if sheet == nil {
			t.Fatal("FindSheet(GPU-A) = nil, want non-nil")
		}
		if sheet.Name != "GPU-A" {
			t.Errorf("got name %q, want %q", sheet.Name, "GPU-A")
		}
		if len(sheet.FixedHeaders) != 1 || sheet.FixedHeaders[0].Name != "年份" {
			t.Errorf("unexpected FixedHeaders: %+v", sheet.FixedHeaders)
		}
	})

	t.Run("non-existing sheet returns nil", func(t *testing.T) {
		if got := s.FindSheet("GPU-C"); got != nil {
			t.Errorf("FindSheet(GPU-C) = %+v, want nil", got)
		}
	})

	t.Run("returned pointer points into Sheets slice", func(t *testing.T) {
		sheet := s.FindSheet("GPU-B")
		if sheet != &s.Sheets[1] {
			t.Error("FindSheet returned a copy instead of a pointer into Sheets slice")
		}
	})

	t.Run("fallback linear scan when sheetIndex is nil", func(t *testing.T) {
		// manually constructed Schema has no sheetIndex
		manual := &Schema{
			Sheets: []Sheet{
				{Name: "Manual-X"},
			},
		}
		sheet := manual.FindSheet("Manual-X")
		if sheet == nil {
			t.Error("FindSheet(Manual-X) = nil, want non-nil via linear scan")
		}
		if manual.FindSheet("not-exist") != nil {
			t.Error("FindSheet(not-exist) should return nil via linear scan")
		}
	})
}
