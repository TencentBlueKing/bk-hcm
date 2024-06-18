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

package bill

import (
	"fmt"

	"hcm/pkg/api/core"
	apicoreaccount "hcm/pkg/api/core/account-set"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
)

// Account account info
type Account struct {
	*apicoreaccount.BaseMainAccount
}

// Key account key
func (a *Account) Key() string {
	if a.BaseMainAccount == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s",
		a.BaseMainAccount.Vendor, a.BaseMainAccount.ParentAccountID, a.BaseMainAccount.CloudID)
}

// AccountLister account lister
type AccountLister interface {
	ListAllAccount(kt *kit.Kit) ([]*Account, error)
}

// MainAccountLister lister for main account
type MainAccountLister struct {
	Client *client.ClientSet
}

// ListAccount list main account
func (t *MainAccountLister) ListAllAccount(kt *kit.Kit) ([]*Account, error) {
	result, err := t.Client.DataService().Global.MainAccount.List(kt, &core.ListWithoutFieldReq{
		Filter: tools.AllExpression(),
		Page: &core.BasePage{
			Count: true,
		},
	})
	if err != nil {
		return nil, err
	}
	var retList []*Account
	for offset := uint64(0); offset < result.Count; offset = offset + defaultAccountListLimit {
		accountResult, err := t.Client.DataService().Global.MainAccount.List(kt, &core.ListWithoutFieldReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: uint(defaultAccountListLimit),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list account failed, err %s", err.Error())
		}
		for _, item := range accountResult.Details {
			retList = append(retList, &Account{
				BaseMainAccount: item,
			})
		}
	}
	return retList, nil
}
