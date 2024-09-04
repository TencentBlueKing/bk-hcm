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
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

var (
	bomHeader = []byte{0xEF, 0xBB, 0xBF}
)

// NewCsvWriter ...
func NewCsvWriter(kt *kit.Kit, writer io.Writer) (*csv.Writer, error) {
	// 写入BOM头, 兼容windows excel打开csv文件时中文乱码
	_, err := writer.Write(bomHeader)
	if err != nil {
		logs.Errorf("write BOM failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return csv.NewWriter(writer), nil
}

// CreateWriterByFileName ...
func CreateWriterByFileName(kt *kit.Kit, filename string) (
	finalFilename, filepath string, writer *csv.Writer, closeFunc func() error, err error) {

	if err := os.MkdirAll(cc.AccountServer().TmpFileDir, 0600); err != nil {
		logs.Errorf("mkdir failed: %v, rid: %s", err, kt.Rid)
		return "", "", nil, nil, err
	}

	finalFilename = fmt.Sprintf("%s.zip", filename)
	filepath = fmt.Sprintf("%s/%s", cc.AccountServer().TmpFileDir, finalFilename)
	file, err := os.Create(filepath)
	if err != nil {
		logs.Errorf("create file failed: %v, filepath: %s,rid: %s", err, filepath, kt.Rid)
		return "", "", nil, file.Close, err
	}

	zipWriter := zip.NewWriter(file)
	zipFile, err := zipWriter.Create(filename)
	if err != nil {
		return "", "", nil, zipWriter.Close, err
	}

	writer, err = NewCsvWriter(kt, zipFile)
	if err != nil {
		logs.Errorf("new csv writer failed: %v, rid: %s", err, kt.Rid)
		return "", "", nil, zipWriter.Close, err
	}

	return finalFilename, filepath, writer, zipWriter.Close, nil
}
