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

package billmonthtask

import (
	dataproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// ListBillMonthPullTask list bill month task with options
func (svc *service) ListBillMonthPullTask(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.BillMonthTaskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.AccountBillMonthPullTask().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.BillMonthTaskResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoPullerResult(&d)
	}

	return &dataproto.BillMonthTaskListResult{Details: details, Count: &data.Count}, nil
}

func toProtoPullerResult(m *tablebill.AccountBillMonthTask) *dataproto.BillMonthTaskResult {
	return &dataproto.BillMonthTaskResult{
		ID:                 m.ID,
		RootAccountID:      m.RootAccountID,
		RootAccountCloudID: m.RootAccountCloudID,
		Vendor:             m.Vendor,
		BillYear:           m.BillYear,
		BillMonth:          m.BillMonth,
		VersionID:          m.VersionID,
		State:              m.State,
		Count:              m.Count,
		Currency:           m.Currency,
		Cost:               cvt.PtrToVal(m.Cost).Decimal,
		PullIndex:          m.PullIndex,
		PullFlowID:         m.PullFlowID,
		SplitIndex:         m.SplitIndex,
		SplitFlowID:        m.SplitFlowID,
		SummaryFlowID:      m.SummaryFlowID,
		SummaryDetail:      m.SummaryDetail,
		Creator:            m.Creator,
		Reviser:            m.Reviser,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}
