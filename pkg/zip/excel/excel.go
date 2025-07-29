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
	"fmt"
	"os"
	"sync"

	"github.com/xuri/excelize/v2"
)

type excelWriter struct {
	sync.Mutex
	sheetWriters map[string]*excelSheetWriter
	file         *excelize.File
}

func newExcelWriter(filePath string) (*excelWriter, error) {
	if !isFileExist(filePath) {
		if err := excelize.NewFile().SaveAs(filePath); err != nil {
			return nil, err
		}
	}

	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	return &excelWriter{file: file, sheetWriters: make(map[string]*excelSheetWriter)}, nil
}

func isFileExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func (e *excelWriter) write(sheet string, data [][]interface{}) error {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.sheetWriters[sheet]; !ok {
		if _, err := e.file.NewSheet(sheet); err != nil {
			return err
		}
		streamWriter, err := e.file.NewStreamWriter(sheet)
		if err != nil {
			return err
		}
		writer, err := newExcelSheetWriter(streamWriter)
		if err != nil {
			return err
		}
		e.sheetWriters[sheet] = writer
	}

	if err := e.sheetWriters[sheet].write(data); err != nil {
		return err
	}

	return nil
}

const defaultSheet = "Sheet1"

func (e *excelWriter) save() error {
	e.Lock()
	defer e.Unlock()

	existDefaultSheet := false
	for sheet, writer := range e.sheetWriters {
		if err := writer.flush(); err != nil {
			return err
		}
		if sheet == defaultSheet {
			existDefaultSheet = true
		}
	}
	if !existDefaultSheet {
		if err := e.file.DeleteSheet(defaultSheet); err != nil {
			return fmt.Errorf("delete default sheet failed: %v", err)
		}
	}

	return e.file.Save()
}

func (e *excelWriter) close() error {
	if err := e.save(); err != nil {
		return err
	}

	return e.file.Close()
}

type excelSheetWriter struct {
	sync.Mutex
	rowIdx int
	writer *excelize.StreamWriter
}

func newExcelSheetWriter(writer *excelize.StreamWriter) (*excelSheetWriter, error) {
	return &excelSheetWriter{writer: writer}, nil
}

func (e *excelSheetWriter) write(data [][]interface{}) error {
	e.Lock()
	defer e.Unlock()

	for _, row := range data {
		e.rowIdx++
		cell, _ := excelize.CoordinatesToCellName(1, e.rowIdx)
		if err := e.writer.SetRow(cell, row); err != nil {
			return err
		}
	}

	return nil
}

func (e *excelSheetWriter) flush() error {
	return e.writer.Flush()
}
