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
	"strings"

	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"
)

// OperationType ...
type OperationType string

const (
	// CreateLayer4Listener 创建四层监听器
	CreateLayer4Listener = "create_layer4_listener"
	// CreateLayer7Listener 创建七层监听器
	CreateLayer7Listener = "create_layer7_listener"
	// CreateUrlRule 创建URL规则
	CreateUrlRule = "create_url_rule"
	// Layer4ListenerBindRs 四层监听器绑定RS
	Layer4ListenerBindRs = "layer4_listener_bind_rs"
	// Layer7ListenerBindRs 七层监听器绑定RS
	Layer7ListenerBindRs = "layer7_listener_bind_rs"
)

// ImportPreviewExecutor 导入预览执行器
type ImportPreviewExecutor interface {
	Execute(*kit.Kit, [][]string) (interface{}, error)

	convertDataToPreview([][]string) error
	validate(*kit.Kit) error
}

// NewImportPreviewExecutor ...
func NewImportPreviewExecutor(operationType OperationType, service *dataservice.Client,
	vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) (ImportPreviewExecutor, error) {

	switch operationType {
	case CreateLayer4Listener:
		return newCreateLayer4ListenerPreviewExecutor(service, vendor, bkBizID, accountID, regionIDs), nil
	case CreateLayer7Listener:
		return newCreateLayer7ListenerPreviewExecutor(service, vendor, bkBizID, accountID, regionIDs), nil
	case CreateUrlRule:
		return newCreateUrlRulePreviewExecutor(service, vendor, bkBizID, accountID, regionIDs), nil
	case Layer4ListenerBindRs:
		return newLayer4ListenerBindRSPreviewExecutor(service, vendor, bkBizID, accountID, regionIDs), nil
	case Layer7ListenerBindRs:
		return newLayer7ListenerBindRSPreviewExecutor(service, vendor, bkBizID, accountID, regionIDs), nil
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operationType)
	}
}

type basePreviewExecutor struct {
	vendor         enumor.Vendor
	accountID      string
	bkBizID        int64
	regionIDMap    map[string]struct{}
	dataServiceCli *dataservice.Client
}

func newBasePreviewExecutor(cli *dataservice.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) *basePreviewExecutor {

	return &basePreviewExecutor{
		dataServiceCli: cli,
		accountID:      accountID,
		bkBizID:        bkBizID,
		vendor:         vendor,
		regionIDMap:    converter.StringSliceToMap(regionIDs),
	}
}

// ImportStatus excel导入的数据状态
type ImportStatus string

const (
	// Executable ...
	Executable ImportStatus = "executable"
	// NotExecutable ...
	NotExecutable ImportStatus = "not_executable"
	// Existing ...
	Existing ImportStatus = "existing"
)

func trimSpaceForSlice(strs []string) []string {
	for i, str := range strs {
		strs[i] = strings.TrimSpace(str)
	}
	return strs
}
