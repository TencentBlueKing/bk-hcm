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
	"fmt"

	"hcm/pkg/kit"
	"hcm/pkg/thirdparty"
)

// ApproveReq define approve req.
type ApproveReq struct {
	Sn       string `json:"sn"`
	StateID  int    `json:"state_id"`
	Approver string `json:"approver"`
	Action   string `json:"action"`
	Remark   string `json:"remark"`
}

// Approve 快捷审批接口。
func (i *itsm) Approve(kt *kit.Kit, req *ApproveReq) error {

	resp := new(thirdparty.BaseResponse)
	err := i.client.Post().
		SubResourcef("/approve/").
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}

	if !resp.Result || resp.Code != 0 {
		return fmt.Errorf("approve failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return nil
}
