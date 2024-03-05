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

package tcloud

import (
	"fmt"

	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/golang/protobuf/proto"
	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
)

// GetBillList get bill list.
// ref: https://console.cloud.tencent.com/api/explorer?Product=billing&Version=2018-07-09&Action=DescribeBillDetail
func (t *TCloudImpl) GetBillList(kt *kit.Kit, opt *typesBill.TCloudBillListOption) (
	*billing.DescribeBillDetailResponseParams, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	billClient, err := t.clientSet.BillClient()
	if err != nil {
		return nil, fmt.Errorf("new bill client failed, err: %v, rid: %s", err, kt.Rid)
	}

	req := billing.NewDescribeBillDetailRequest()
	req.Offset = proto.Uint64(opt.Page.Offset)
	req.Limit = proto.Uint64(opt.Page.Limit)
	if opt.Month != "" {
		req.Month = proto.String(opt.Month)
	}
	if opt.BeginDate != "" {
		req.BeginTime = proto.String(opt.BeginDate)
	}
	if opt.EndDate != "" {
		req.EndTime = proto.String(opt.EndDate)
	}
	// 是否需要访问列表的总记录数，用于前端分页(1-表示需要 0-表示不需要)
	req.NeedRecordNum = proto.Int64(1)

	resp, err := billClient.DescribeBillDetail(req)
	if err != nil {
		logs.Errorf("get tencent cloud bill list failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}
	if resp.Response == nil {
		return nil, errf.New(errf.RecordNotFound, "tcloud bill list is not found")
	}

	return resp.Response, nil
}
