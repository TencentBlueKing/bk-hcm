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

package itsm

import (
	"context"
	"fmt"

	"hcm/pkg/thirdparty/esb/types"
)

type TicketResult struct {
	SN            string `json:"sn"`
	TicketURL     string `json:"ticket_url"`
	CurrentStatus string `json:"current_status"`
	ApproveResult bool   `json:"approve_result"`
}

type queryTicketResp struct {
	types.BaseResponse `json:",inline"`
	Data               []TicketResult `json:"data"`
}

func (i *itsm) batchQueryTicketResult(ctx context.Context, sns []string) (results []TicketResult, err error) {
	req := map[string]interface{}{"sn": sns}
	resp := new(queryTicketResp)
	header := types.GetCommonHeader(i.config)
	err = i.client.Post().
		SubResourcef("/itsm/ticket_approval_result/").
		WithContext(ctx).
		WithHeaders(*header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return results, err
	}
	if !resp.Result || resp.Code != 0 {
		return results, fmt.Errorf(
			"query ticket result failed, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, resp.Rid,
		)
	}

	return resp.Data, nil
}

func (i *itsm) GetTicketResult(ctx context.Context, sn string) (result TicketResult, err error) {
	results, err := i.batchQueryTicketResult(ctx, []string{sn})
	if err != nil {
		return result, err
	}

	return results[0], nil
}
