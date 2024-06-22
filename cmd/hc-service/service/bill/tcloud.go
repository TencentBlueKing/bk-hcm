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

// Package bill defines bill service.
package bill

import (
	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/adaptor/types/core"
	hcbillservice "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudGetBillList get tcloud bill list.
func (b bill) TCloudGetBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbillservice.TCloudBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page == nil {
		req.Page = &core.TCloudPage{Offset: 0, Limit: core.TCloudQueryLimit}
	}

	cli, err := b.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("tcloud request adaptor client err, err: %+v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.TCloudBillListOption{
		AccountID: req.AccountID,
		Month:     req.Month,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
		Page: &core.TCloudPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		},
	}
	if req.Context != nil {
		opt.Context = req.Context
	}
	resp, err := cli.GetBillList(cts.Kit, opt)
	if err != nil {
		logs.Errorf("tcloud request adaptor list bill failed, req: %v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return &hcbillservice.TCloudBillListResult{
		Count:     resp.Total,
		Details:   resp.DetailSet,
		Context:   resp.Context,
		RequestId: resp.RequestId,
	}, nil
}
