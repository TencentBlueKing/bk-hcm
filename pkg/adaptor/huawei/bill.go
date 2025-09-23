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

package huawei

import (
	"fmt"

	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/gogo/protobuf/proto"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
)

// GetBillList get bill list.
// reference: https://support.huaweicloud.com/api-oce/mbc_00003.html
func (h *HuaWei) GetBillList(kt *kit.Kit, opt *typesBill.HuaWeiBillListOption) (
	*model.ListCustomerselfResourceRecordDetailsResponse, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, fmt.Errorf("new bill client failed, err: %v", err)
	}

	req := new(model.ListCustomerselfResourceRecordDetailsRequest)
	req.Body = &model.QueryResRecordsDetailReq{
		Cycle:         opt.Month,
		StatisticType: proto.Int(2), // 统计类型。默认值为1。 1：按账期2：按天
	}
	if opt.Page != nil {
		req.Body.Offset = opt.Page.Offset
		req.Body.Limit = opt.Page.Limit
	}

	resp, err := client.ListCustomerselfResourceRecordDetails(req)
	if err != nil {
		logs.Errorf("huawei bill list request adaptor failed, err: %+v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return resp, nil
}

// GetFeeRecordList get fee record list.
// doc: https://console-intl.huaweicloud.com/apiexplorer/#/openapi/BSSINTL/debug?api=ListCustomerselfResourceRecords
func (h *HuaWei) GetFeeRecordList(kt *kit.Kit, opt *typesBill.HuaWeiFeeRecordListOption) (
	*model.ListCustomerselfResourceRecordsResponse, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, fmt.Errorf("new fee record failed, err: %v", err)
	}

	method := "sub_customer"
	req := new(model.ListCustomerselfResourceRecordsRequest)
	req.SubCustomerId = &opt.SubAccountID
	req.Method = &method
	req.Cycle = opt.Month
	req.BillDateBegin = &opt.BillDateBegin
	req.BillDateEnd = &opt.BillDateEnd

	// 统计类型。默认值为3。 1：按账期 3：按明细；
	// 当前版本sdk没有StatisticType字段，暂时先不填，默认为3
	// req.StatisticType = proto.Int(3)

	if opt.Page != nil {
		req.Offset = opt.Page.Offset
		req.Limit = opt.Page.Limit
	}

	resp, err := client.ListCustomerselfResourceRecords(req)
	if err != nil {
		logs.Errorf("huawei fee record list request adaptor failed, err: %+v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return resp, nil
}
