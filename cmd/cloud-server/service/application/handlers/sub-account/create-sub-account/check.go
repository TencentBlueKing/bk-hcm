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

	svccommon "hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/runtime/filter"
)

// CheckReq validate the request and check business rules.
func (a *ApplicationOfCreateSubAccount) CheckReq() error {
	if err := a.req.Validate(); err != nil {
		return err
	}

	if err := a.validateExtension(); err != nil {
		return err
	}

	account, err := a.GetAccount(a.req.AccountID)
	if err != nil {
		return fmt.Errorf("found parent account(%s) failed, err: %w", a.req.AccountID, err)
	}

	if a.BkBizID() != account.BkBizID {
		return fmt.Errorf("account(%s)'s biz_id is %d,biz_id(%d) no perssion to operate subaccount of it",
			a.req.AccountID, account.BkBizID, a.BkBizID())
	}

	if err := a.checkDuplicateName(); err != nil {
		return err
	}

	return nil
}

func (a *ApplicationOfCreateSubAccount) validateExtension() error {
	switch a.Vendor() {
	case enumor.TCloud:
		ext, err := decodeTCloudExtension(a)
		if err != nil {
			return err
		}
		return ext.Validate()
	default:
		return fmt.Errorf("vendor %s is not supported", a.Vendor())
	}
}

// decodeTCloudExtension decodes the request extension into TCloudSubAccountAddExtension.
// It is shared between CheckReq and Deliver to avoid duplicating decode logic.
func decodeTCloudExtension(a *ApplicationOfCreateSubAccount) (*proto.TCloudSubAccountAddExtension, error) {
	ext := new(proto.TCloudSubAccountAddExtension)
	if err := svccommon.DecodeExtension(a.Cts.Kit, a.req.Extension, ext); err != nil {
		return nil, fmt.Errorf("decode tcloud sub account extension failed, err: %w", err)
	}

	return ext, nil
}

func (a *ApplicationOfCreateSubAccount) checkDuplicateName() error {
	result, err := a.Client.DataService().Global.SubAccount.List(
		a.Cts.Kit, &core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: a.req.AccountID},
					filter.AtomRule{Field: "name", Op: filter.Equal.Factory(), Value: a.req.Name},
				},
			},
			Page: &core.BasePage{Count: true},
		},
	)
	if err != nil {
		return fmt.Errorf("check sub account name duplicate failed, err: %w", err)
	}

	if result.Count > 0 {
		return fmt.Errorf("sub account name [%s] already exists under account [%s]", a.req.Name, a.req.AccountID)
	}

	return nil
}
