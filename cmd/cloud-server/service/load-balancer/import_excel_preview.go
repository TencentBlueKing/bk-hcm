/*
 *
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

package loadbalancer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"

	"github.com/xuri/excelize/v2"
)

// ImportPreview 上传excel导入的文件, 解析&预校验
func (svc *lbSvc) ImportPreview(cts *rest.Contexts) (interface{}, error) {

	operationType := cts.PathParameter("operation_type").String()
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	file, _, err := cts.Request.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	regionIDsStr := cts.Request.Request.FormValue("region_ids")
	var regionIDs []string
	if err = json.Unmarshal([]byte(regionIDsStr), &regionIDs); err != nil {
		logs.Errorf("unmarshal region_ids failed, str: %s, err: %v, rid: %s", regionIDsStr, err, cts.Kit.Rid)
		return nil, err
	}
	accountID := cts.Request.Request.FormValue("account_id")

	handlerOpt := &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  &types.CloudResourceBasicInfo{AccountID: accountID},
	}
	if err = handler.BizOperateAuth(cts, handlerOpt); err != nil {
		return nil, err
	}

	curVendor, rawData, err := parseExcelStr(file)
	if err != nil {
		logs.Errorf("parse excel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("parse excel failed, err: %v", err)
	}
	if len(rawData) > constant.ExcelImportRowLimit {
		return nil, fmt.Errorf("rows count should less than %d", constant.ExcelImportRowLimit)
	}
	if vendor != curVendor {
		return nil, errors.New("excel file vendor not match")
	}

	executor, err := lblogic.NewImportPreviewExecutor(lblogic.OperationType(operationType), svc.client.DataService(),
		vendor, bizID, accountID, regionIDs)
	if err != nil {
		logs.Errorf("new import preview executor failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	result, err := executor.Execute(cts.Kit, rawData)
	if err != nil {
		logs.Errorf("execute import preview failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return cslb.UploadExcelFileBaseResp{
		Details: result,
	}, nil
}

func parseExcelStr(reader io.Reader) (enumor.Vendor, [][]string, error) {
	excel, err := excelize.OpenReader(reader)
	if err != nil {
		return "", nil, err
	}
	defer excel.Close()

	rows, err := excel.Rows(excel.GetSheetName(0))
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()
	rows.Next()
	columns, err := rows.Columns()
	if err != nil {
		return "", nil, err
	}
	vendor, err := parseVendor(columns)
	if err != nil {
		return "", nil, err
	}

	// 跳过表头
	for i := 0; i < 2; i++ {
		rows.Next()
	}
	result := make([][]string, 0)
	for rows.Next() {
		columns, err = rows.Columns()
		if err != nil {
			return "", nil, err
		}
		if len(columns) == 0 {
			continue
		}

		result = append(result, columns)
	}

	return vendor, result, nil
}

var supportVendorMap = map[string]enumor.Vendor{
	"tencent_cloud_public(腾讯云-公有云)": enumor.TCloud,
}

func parseVendor(columns []string) (enumor.Vendor, error) {
	if len(columns) < 2 {
		return "", errors.New("excel file format error, line 1: not enough columns")
	}
	vendor, ok := supportVendorMap[columns[1]]
	if !ok {
		return "", fmt.Errorf("unsupported vendor: %s", columns[1])
	}
	return vendor, nil
}
