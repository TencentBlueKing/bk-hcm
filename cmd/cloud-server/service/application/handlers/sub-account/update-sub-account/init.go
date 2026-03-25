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

// Package updatesubaccount provides the handler for updating sub accounts.
package updatesubaccount

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	subaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

var _ handlers.ApplicationHandler = (*ApplicationOfUpdateSubAccount)(nil)

func init() {
	subaccount.RegisterActionHandler(enumor.SubAccountActionUpdate, newHandlerFromContent)
}

func newHandlerFromContent(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent, content string,
) (handlers.ApplicationHandler, error) {

	ct := new(updateSubAccountContent)
	if err := json.UnmarshalFromString(content, ct); err != nil {
		return nil, fmt.Errorf("unmarshal update sub account content failed, err: %w", err)
	}

	return NewApplicationOfUpdateSubAccount(opt, base, ct.SubAccountName, converter.ValToPtr(ct.Req)), nil
}

// ApplicationOfUpdateSubAccount handler for updating sub account.
type ApplicationOfUpdateSubAccount struct {
	subaccount.ApplicationBaseSubAccount

	req            *proto.SubAccountUpdateReq
	subAccountName string
}

// NewApplicationOfUpdateSubAccount create a new handler for updating sub account.
func NewApplicationOfUpdateSubAccount(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent,
	subAccountName string, req *proto.SubAccountUpdateReq) *ApplicationOfUpdateSubAccount {

	return &ApplicationOfUpdateSubAccount{
		ApplicationBaseSubAccount: subaccount.NewApplicationBaseSubAccount(opt, base),
		req:                       req,
		subAccountName:            subAccountName,
	}
}
