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

// Package updatesecretstatus provides the handler for updating sub account secret key status.
package updatesecretstatus

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	subaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

var _ handlers.ApplicationHandler = (*ApplicationOfUpdateSecretKeyStatus)(nil)

func init() {
	subaccount.RegisterActionHandler(enumor.SubAccountActionUpdateSecretKeyStatus, newHandlerFromContent)
}

func newHandlerFromContent(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent, content string,
) (handlers.ApplicationHandler, error) {
	ct := new(updateSecretKeyStatusContent)
	if err := json.UnmarshalFromString(content, ct); err != nil {
		return nil, fmt.Errorf("unmarshal update secret key status content failed, err: %w", err)
	}

	h := NewApplicationOfUpdateSecretKeyStatus(opt, base,
		&proto.SubAccountSecretStatusUpdateReq{ID: ct.SecretID, Status: ct.TargetStatus},
	)
	h.SetAccountID(ct.AccountID)
	return h, nil
}

// ApplicationOfUpdateSecretKeyStatus handler for updating sub account secret key status.
type ApplicationOfUpdateSecretKeyStatus struct {
	subaccount.ApplicationBaseSubAccount

	req *proto.SubAccountSecretStatusUpdateReq
}

// NewApplicationOfUpdateSecretKeyStatus create a new handler.
func NewApplicationOfUpdateSecretKeyStatus(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent,
	req *proto.SubAccountSecretStatusUpdateReq) *ApplicationOfUpdateSecretKeyStatus {

	return &ApplicationOfUpdateSecretKeyStatus{
		ApplicationBaseSubAccount: subaccount.NewApplicationBaseSubAccount(opt, base),
		req:                       req,
	}
}

// GetItsmApprover 获取itsm审批人
func (a *ApplicationOfUpdateSecretKeyStatus) GetItsmApprover(kt *kit.Kit, managers []string) (
	[]itsm.VariableApprover, error) {

	return a.GetAccountApprover(kt, a.AccountID())
}
