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

// MainAccount main account info
type MainAccount struct {
	*apicoreaccount.BaseMainAccount
}

// Key main account key
func (a *MainAccount) Key() string {
	if a.BaseMainAccount == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s",
		a.BaseMainAccount.Vendor, a.BaseMainAccount.ParentAccountID, a.BaseMainAccount.ID)
}

// RootAccount root account info
type RootAccount struct {
	*apicoreaccount.BaseRootAccount
}

// Key root account key
func (r *RootAccount) Key() string {
	if r.BaseRootAccount == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s",
		r.BaseRootAccount.Vendor, r.BaseRootAccount.ID)
}

// AccountLister account lister
type AccountLister interface {
	ListAllMainAccount(kt *kit.Kit) ([]*MainAccount, error)
	ListAllRootAccount(kt *kit.Kit) ([]*RootAccount, error)
}

// HcmAccountLister lister for main account and root account
type HcmAccountLister struct {
	Client *client.ClientSet
}

// ListAllMainAccount list main account
func (t *HcmAccountLister) ListAllMainAccount(kt *kit.Kit) ([]*MainAccount, error) {
	result, err := t.Client.DataService().Global.MainAccount.List(kt, &core.ListReq{
		Filter: tools.AllExpression(),
		Page: &core.BasePage{
			Count: true,
		},
	})
	if err != nil {
		return nil, err
	}
	var retList []*MainAccount
	for offset := uint64(0); offset < result.Count; offset = offset + defaultAccountListLimit {
		accountResult, err := t.Client.DataService().Global.MainAccount.List(kt, &core.ListReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: uint(defaultAccountListLimit),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("list main account failed, err %s", err.Error())
		}
		for _, item := range accountResult.Details {
			retList = append(retList, &MainAccount{
				BaseMainAccount: item,
			})
		}
	}
	return retList, nil
}

// ListAllRootAccount list root account
func (t *HcmAccountLister) ListAllRootAccount(kt *kit.Kit) ([]*RootAccount, error) {
	listReq := &core.ListReq{Filter: tools.AllExpression(), Page: core.NewCountPage()}
	result, err := t.Client.DataService().Global.RootAccount.List(kt, listReq)
	if err != nil {
		return nil, err
	}
	var retList []*RootAccount
	for offset := uint64(0); offset < result.Count; offset = offset + defaultAccountListLimit {
		rootAccountReq := &core.ListReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: uint(defaultAccountListLimit),
			},
		}
		accountResult, err := t.Client.DataService().Global.RootAccount.List(kt, rootAccountReq)
		if err != nil {
			return nil, fmt.Errorf("list root account failed, err %s", err.Error())
		}
		for _, item := range accountResult.Details {
			retList = append(retList, &RootAccount{
				BaseRootAccount: item,
			})
		}
	}
	return retList, nil
}
