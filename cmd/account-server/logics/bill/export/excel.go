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

package export

import (
	"bytes"

	"hcm/pkg/logs"

	"github.com/xuri/excelize/v2"
)

const (
	defaultSheetName = "Sheet1"
)

// GenerateExcel 生成 Excel 文件
func GenerateExcel(data [][]interface{}) (*bytes.Buffer, error) {
	// 创建一个新的 Excel 文档
	f := excelize.NewFile()

	// 设置工作表的名称为 "Sheet1"
	_, err := f.NewSheet(defaultSheetName)
	if err != nil {
		return nil, err
	}

	// 写入数据到 Excel
	for row, rowData := range data {
		for col, value := range rowData {
			cell, err := excelize.CoordinatesToCellName(col+1, row+1)
			if err != nil {
				return nil, err
			}
			if err = f.SetCellValue(defaultSheetName, cell, value); err != nil {
				logs.Errorf("write value (%v) to cell[%d,%d] error: %v", value, col+1, row+1, err.Error())
				return nil, err
			}
		}
	}

	return f.WriteToBuffer()
}
