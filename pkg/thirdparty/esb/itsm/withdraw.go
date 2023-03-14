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

// WithdrawTicket 撤销单据
func (i *itsm) WithdrawTicket(ctx context.Context, sn string, operator string) error {
	req := map[string]interface{}{
		"sn":             sn,
		"operator":       operator,
		"action_type":    "WITHDRAW",
		"action_message": "applicant withdraw ticket",
	}
	resp := new(types.BaseResponse)
	//  Note: 由于某些版本的ESB bkApiAuthorization里的bk_username 会与接口参数里的operator冲突，导致接口里的operator会被bk_username覆盖
	//   所以这里传入operator避免被默认esb调用bk_username替代
	header := types.GetCommonHeaderByUser(i.config, operator)
	err := i.client.Post().
		SubResourcef("/itsm/operate_ticket/").
		WithContext(ctx).
		WithHeaders(*header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return err
	}
	if !resp.Result || resp.Code != 0 {
		return fmt.Errorf("withdraw ticket failed, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, resp.Rid)
	}

	return nil
}
