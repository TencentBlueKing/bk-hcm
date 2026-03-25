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

package createsubaccount

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	subaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/tools/json"
)

var _ handlers.ApplicationHandler = (*ApplicationOfCreateSubAccount)(nil)

func init() {
	subaccount.RegisterActionHandler(enumor.SubAccountActionCreate, newHandlerFromContent)
}

func newHandlerFromContent(opt *handlers.HandlerOption, base *subaccount.BaseSubAccountContent, content string,
) (handlers.ApplicationHandler, error) {

	req := new(proto.SubAccountAddReq)
	if err := json.UnmarshalFromString(content, req); err != nil {
		return nil, fmt.Errorf("unmarshal create sub account content failed, err: %w", err)
	}
	return NewApplicationOfCreateSubAccount(opt, base.Vendor, base.BkBizID, req), nil
}

// ApplicationOfCreateSubAccount handler for creating sub account.
type ApplicationOfCreateSubAccount struct {
	subaccount.ApplicationBaseSubAccount

	req *proto.SubAccountAddReq
}

// NewApplicationOfCreateSubAccount create a new handler for creating sub account.
func NewApplicationOfCreateSubAccount(opt *handlers.HandlerOption, vendor enumor.Vendor, bkBizID int64,
	req *proto.SubAccountAddReq,
) *ApplicationOfCreateSubAccount {

	return &ApplicationOfCreateSubAccount{
		ApplicationBaseSubAccount: subaccount.NewApplicationBaseSubAccount(
			opt, enumor.SubAccountActionCreate, vendor, bkBizID, req.AccountID,
		),
		req: req,
	}
}
