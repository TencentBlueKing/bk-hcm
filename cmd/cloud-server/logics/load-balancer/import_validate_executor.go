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
	"encoding/json"
	"fmt"

	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// ImportValidator clb导入 参数校验器
type ImportValidator interface {
	Validate(*kit.Kit, json.RawMessage) (interface{}, error)
}

// NewImportValidator ...
func NewImportValidator(operationType OperationType, service *dataservice.Client,
	vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) (ImportValidator, error) {

	switch operationType {
	case CreateLayer4Listener:
		return newCreateLayer4ListenerValidator(service, vendor, bkBizID, accountID, regionIDs), nil
	case CreateLayer7Listener:
		return newCreateLayer7ListenerValidator(service, vendor, bkBizID, accountID, regionIDs), nil
	case CreateUrlRule:
		return newCreateUrlRuleValidator(service, vendor, bkBizID, accountID, regionIDs), nil
	case Layer4ListenerBindRs:
		return newLayer4ListenerBindRSValidator(service, vendor, bkBizID, accountID, regionIDs), nil
	case Layer7ListenerBindRs:
		return newLayer7ListenerBindRSValidator(service, vendor, bkBizID, accountID, regionIDs), nil
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operationType)
	}
}
