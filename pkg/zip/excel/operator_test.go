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
	"testing"

	"hcm/pkg/zip"
)

func TestExcelZip(t *testing.T) {
	excelData := map[string][][]string{
		`excel1.xlsx`: {
			{"excel1第一列", "excel1第二列", "excel1第三列"},
			{"excel1-1-1", "excel1-1-2", "excel1-1-3"},
			{"excel1-2-1", "excel1-2-2", "excel1-2-3"},
		},
		"excel2.xlsx": {
			{"excel2第一列", "excel2第二列", "excel2第三列"},
			{"excel2-1-1", "excel2-1-2", "excel2-1-3"},
			{"excel2-2-1", "excel2-2-2", "excel2-2-3"},
		},
		"excel3.xlsx": {
			{"excel3第一列", "excel3第二列", "excel3第三列"},
			{"excel3-1-1", "excel3-1-2", "excel3-1-3"},
			{"excel3-2-1", "excel3-2-2", "excel3-2-3"},
		},
	}

	op, err := NewOperator("./", zip.GenFileName("testExcelZip"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	lastRow := [][]string{{"最后第1列", "最后第2列", "最后第3列"}}
	for fileName, data := range excelData {
		name := CombineFileNameAndSheet(fileName, "sheet_test")
		if err = op.Write(name, data); err != nil {
			t.Error(err.Error())
			return
		}

		// add last Row
		if err = op.Write(name, lastRow); err != nil {
			t.Error(err.Error())
			return
		}
	}
	if err = op.Save(); err != nil {
		t.Error(err.Error())
		return
	}
	if err = op.Close(); err != nil {
		t.Error(err.Error())
		return
	}
}
