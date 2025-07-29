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
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/uuid"
	zipOp "hcm/pkg/zip"
)

// Operator ...
type Operator struct {
	sync.Mutex
	path           string
	tempDir        string
	zipName        string
	excelWriterMap map[string]*excelWriter
}

// NewOperator ...
func NewOperator(path string, name string) (zipOp.OperatorI, error) {
	tempDir := filepath.Join(path, "temp"+uuid.UUID())
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return nil, err
	}

	return &Operator{
		path:           path,
		zipName:        name,
		tempDir:        tempDir,
		excelWriterMap: make(map[string]*excelWriter),
	}, nil
}

func (o *Operator) getZipPath() string {
	return filepath.Join(o.path, o.zipName)
}

func (o *Operator) getExcelPath(fileName string) string {
	return filepath.Join(o.tempDir, fileName)
}

// Write data to excel
func (o *Operator) Write(name string, data [][]string) error {
	o.Lock()
	defer o.Unlock()

	fileName, sheet, err := splitToFileNameAndSheet(name)
	if err != nil {
		return err
	}

	if _, ok := o.excelWriterMap[fileName]; !ok {
		path := o.getExcelPath(fileName)
		writer, err := newExcelWriter(path)
		if err != nil {
			return err
		}
		o.excelWriterMap[fileName] = writer
	}

	formattedData := make([][]interface{}, len(data))
	for i, row := range data {
		formattedData[i] = make([]interface{}, len(row))
		for j, cell := range row {
			formattedData[i][j] = cell
		}
	}

	if err = o.excelWriterMap[fileName].write(sheet, formattedData); err != nil {
		return err
	}

	return nil
}

// Flush ...
func (o *Operator) Flush() error {
	return nil
}

// Save ...
func (o *Operator) Save() error {
	o.Lock()
	defer o.Unlock()

	for _, writer := range o.excelWriterMap {
		if err := writer.save(); err != nil {
			return err
		}

	}

	zipPath := o.getZipPath()
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, excelName := range maps.Keys(o.excelWriterMap) {
		filePath := o.getExcelPath(excelName)
		if err := addToZip(zipWriter, filePath); err != nil {
			return err
		}
	}

	return nil
}

// addToZip 添加文件到zip
func addToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath) // 仅保留文件名
	header.Method = zip.Deflate           // 启用压缩

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// 使用缓冲区流式复制（避免大文件内存溢出）
	buf := make([]byte, 4*1024*1024) // 4MB缓冲区
	_, err = io.CopyBuffer(writer, file, buf)
	return err
}

// Close ...
func (o *Operator) Close() error {
	o.Lock()
	defer o.Unlock()

	errs := make([]error, 0)
	for _, writer := range o.excelWriterMap {
		if err := writer.close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 清除临时目录
	if err := os.RemoveAll(o.tempDir); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}
