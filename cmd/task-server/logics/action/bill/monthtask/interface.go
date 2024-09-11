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

package monthtask

import (
	rawjson "encoding/json"
	"fmt"

	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// GetRunner return month task vendor runner
func GetRunner(vendor enumor.Vendor, taskType enumor.MonthTaskType) (MonthTaskRunner, error) {
	switch vendor {
	case enumor.Gcp:
		return newGcpRunner(taskType)
	case enumor.Aws:
		return newAwsRunner(taskType)
	default:
		return nil, fmt.Errorf("vendor %s not support now", vendor)
	}
}

// MonthTaskRunner ...
type MonthTaskRunner interface {
	GetBatchSize(kt *kit.Kit) uint64
	Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
		itemList []bill.RawBillItem, isFinished bool, err error)
	Split(kt *kit.Kit, opt *MonthTaskActionOption, rawItemList []*bill.RawBillItem) (
		[]bill.BillItemCreateReq[rawjson.RawMessage], error)

	// GetHcProductCodes return all hc product codes for this month task type
	GetHcProductCodes() []string
}
