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

package lblogic

import (
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/table"
	"hcm/pkg/tools/slice"
	"hcm/pkg/zip"
	"hcm/pkg/zip/excel"
)

// Exporter ...
type Exporter interface {
	PreCheck(kt *kit.Kit) error
	Export(kt *kit.Kit) (string, error)
}

func write[T table.Table](kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI, infos map[string][]T,
	headers [][]string, filePrefix string, sheetName string) error {

	firstRow, err := getFirstRow(vendor)
	if err != nil {
		logs.Errorf("get first row failed, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
		return err
	}

	for clbFlag, info := range infos {
		for i, batch := range slice.Split(info, constant.ExportClbOneFileRowLimit) {
			data := make([][]string, 0)
			data = append(data, firstRow)
			data = append(data, headers...)

			for _, one := range batch {
				row, err := one.GetValuesByHeader()
				if err != nil {
					logs.Errorf("get values by header failed, err: %v, data: %v, rid: %s", err, one, kt.Rid)
					return err
				}
				data = append(data, row)
			}

			fileName := fmt.Sprintf("%s-%s-%d.xlsx", filePrefix, clbFlag, i+1)
			name := excel.CombineFileNameAndSheet(fileName, sheetName)
			if err = zipOperator.Write(name, data); err != nil {
				logs.Errorf("write excel failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}
	return nil
}

func writeLayer4Listeners(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	clbListenerMap map[string][]Layer4ListenerDetail) error {

	return write[Layer4ListenerDetail](kt, vendor, zipOperator, clbListenerMap, layer4ListenerHeaders,
		constant.Layer4ListenerFilePrefix, constant.Layer4ListenerSheetName)
}

func writeLayer7Listeners(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	clbListenerMap map[string][]Layer7ListenerDetail) error {

	return write[Layer7ListenerDetail](kt, vendor, zipOperator, clbListenerMap, layer7ListenerHeaders,
		constant.Layer7ListenerFilePrefix, constant.Layer7ListenerSheetName)
}

func writeRules(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	clbRuleMap map[string][]RuleDetail) error {

	return write[RuleDetail](kt, vendor, zipOperator, clbRuleMap, ruleHeaders, constant.RuleFilePrefix,
		constant.RuleSheetName)
}

func writeLayer4Rs(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	clbRsMap map[string][]Layer4RsDetail) error {

	return write[Layer4RsDetail](kt, vendor, zipOperator, clbRsMap, layer4RsHeaders, constant.Layer4RsFilePrefix,
		constant.Layer4RsSheetName)
}

func writeLayer7Rs(kt *kit.Kit, vendor enumor.Vendor, zipOperator zip.OperatorI,
	clbRsMap map[string][]Layer7RsDetail) error {

	return write[Layer7RsDetail](kt, vendor, zipOperator, clbRsMap, layer7RsHeaders, constant.Layer7RsFilePrefix,
		constant.Layer7RsSheetName)
}
