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

package updatesecretstatus

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/dal/dao/tools"
)

// CheckReq validate the request, resolve and populate the secret's parent account info.
func (a *ApplicationOfUpdateSecretKeyStatus) CheckReq() error {
	if err := a.req.Validate(); err != nil {
		return err
	}

	subAccountID, err := a.getSubAccountIDBySecret()
	if err != nil {
		return err
	}

	accountID, err := a.getAccountIDBySubAccount(subAccountID)
	if err != nil {
		return err
	}
	a.SetAccountID(accountID)

	account, err := a.GetAccount(a.AccountID())
	if err != nil {
		return fmt.Errorf("found parent account(%s) failed, err: %w", a.AccountID(), err)
	}

	if account.BkBizID != a.BkBizID() {
		return fmt.Errorf("account(%s)'s biz_id is %d,biz_id(%d) no perssion to operate subaccounts secret of it",
			a.AccountID(), account.BkBizID, a.BkBizID())
	}

	return nil
}

func (a *ApplicationOfUpdateSecretKeyStatus) getSubAccountIDBySecret() (string, error) {
	result, err := a.Client.DataService().Global.SubAccountSecret.ListSubAccountSecret(a.Cts.Kit,
		&protocloud.SubAccountSecretListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.req.ID)),
			Page:   &core.BasePage{Start: 0, Limit: 1},
		},
	)
	if err != nil {
		return "", fmt.Errorf("query sub account secret failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return "", fmt.Errorf("sub account secret(id=%s) not found", a.req.ID)
	}

	return result.Details[0].SubAccountID, nil
}

func (a *ApplicationOfUpdateSecretKeyStatus) getAccountIDBySubAccount(subAccountID string) (string, error) {
	result, err := a.Client.DataService().Global.SubAccount.List(a.Cts.Kit,
		&core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("id", subAccountID)),
			Page:   &core.BasePage{Start: 0, Limit: 1},
		},
	)
	if err != nil {
		return "", fmt.Errorf("query sub account failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return "", fmt.Errorf("sub account(id=%s) not found", subAccountID)
	}

	return result.Details[0].AccountID, nil
}

func (a *ApplicationOfUpdateSecretKeyStatus) getSubAccountNameForDisplay() (string, error) {
	subAccountID, err := a.getSubAccountIDBySecret()
	if err != nil {
		return "", err
	}

	result, err := a.Client.DataService().Global.SubAccount.List(a.Cts.Kit,
		&core.ListReq{Filter: tools.ExpressionAnd(tools.RuleEqual("id", subAccountID)),
			Page: &core.BasePage{Start: 0, Limit: 1},
		},
	)
	if err != nil {
		return "", fmt.Errorf("query sub account failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return "", fmt.Errorf("sub account(id=%s) not found", subAccountID)
	}

	return result.Details[0].Name, nil
}
