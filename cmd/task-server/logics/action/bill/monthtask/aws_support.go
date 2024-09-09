/*
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

package monthtask

import (
	"encoding/json"

	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/kit"
)

type AwsSupportMonthTask struct {
}

func (a AwsSupportMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	// TODO implement me
	panic("implement me")
}

func (a AwsSupportMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (itemList []bill.RawBillItem,
	isFinished bool, err error) {
	// TODO implement me
	panic("implement me")
}

func (a AwsSupportMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {
	// TODO implement me
	panic("implement me")
}
