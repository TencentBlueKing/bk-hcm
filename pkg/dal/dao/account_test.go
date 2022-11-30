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

package dao

import (
	"fmt"
	"testing"
	"time"

	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

func TestCreateAccount(t *testing.T) {
	daoSet, err := testDaoSet()
	checkErr(t, err)

	kt := kit.New()
	kt.User = "Jim"

	memo := "create account test"
	account := &table.Account{
		Spec: &table.AccountSpec{
			Name: "create-account-test",
			Memo: &memo,
		},
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	_, err = daoSet.Account().Create(kt, account)
	checkErr(t, err)
}

func TestUpdateAccount(t *testing.T) {
	daoSet, err := testDaoSet()
	checkErr(t, err)

	kt := kit.New()
	kt.User = "Jim"

	memo := "update account test"
	create := &table.Account{
		Spec: &table.AccountSpec{
			Name: "update-account-test",
			Memo: &memo,
		},
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	id, err := daoSet.Account().Create(kt, create)
	checkErr(t, err)

	time.Sleep(1 * time.Second)

	kt.User = "Tom"
	memo = ""
	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.Equal.Factory(),
				Value: id,
			},
		},
	}

	update := &table.Account{
		Spec: &table.AccountSpec{
			Name: "updated-account-test",
			Memo: &memo,
		},
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}

	err = daoSet.Account().Update(kt, expr, update)
	checkErr(t, err)
}

func TestListAccount(t *testing.T) {
	daoSet, err := testDaoSet()
	checkErr(t, err)

	kt := kit.New()
	kt.User = "Jim"

	opt := &types.ListAccountsOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "creator",
					Op:    "eq",
					Value: "Jim",
				},
			},
		},
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: 100,
			Sort:  "id",
			Order: types.Ascending,
		},
	}

	list, err := daoSet.Account().List(kt, opt)
	checkErr(t, err)

	if len(list.Details) == 0 {
		t.Errorf("list account return data is empty")
		return
	}
}

func TestCountAccount(t *testing.T) {
	daoSet, err := testDaoSet()
	checkErr(t, err)

	kt := kit.New()
	kt.User = "Jim"

	opt := &types.ListAccountsOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "creator",
					Op:    "eq",
					Value: "Jim",
				},
			},
		},
		Page: &types.BasePage{
			Count: true,
		},
	}

	list, err := daoSet.Account().List(kt, opt)
	checkErr(t, err)

	if list.Count == 0 {
		t.Errorf("count account return data is 0")
		return
	}
}

func TestDeleteAccount(t *testing.T) {
	daoSet, err := testDaoSet()
	checkErr(t, err)

	kt := kit.New()
	kt.User = "Jim"

	opt := &types.ListAccountsOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "creator",
					Op:    "eq",
					Value: "Jim",
				},
			},
		},
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: 100,
			Sort:  "id",
			Order: types.Ascending,
		},
	}

	list, err := daoSet.Account().List(kt, opt)
	checkErr(t, err)

	if len(list.Details) == 0 {
		t.Errorf("list account return data is empty")
		return
	}

	ids := make([]uint64, 0)
	for _, one := range list.Details {
		ids = append(ids, one.ID)
	}

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: ids,
			},
		},
	}

	err = daoSet.Account().Delete(kt, expr)
	checkErr(t, err)
}

func TestBatchCreate(t *testing.T) {

	// TODO: 这里仅展示批量创建如何使用，账号应该没有批量创建接口，之后删掉

	daoSet, err := testDaoSet()
	checkErr(t, err)

	kt := kit.New()
	kt.User = "Jim"

	as := make([]table.Account, 0)
	for i := 0; i < 5; i++ {
		memo := "create account test"
		account := table.Account{
			Spec: &table.AccountSpec{
				Name: "create-account-test",
				Memo: &memo,
			},
			Revision: &table.Revision{
				Creator: kt.User,
				Reviser: kt.User,
			},
		}

		as = append(as, account)
	}

	ids, err := daoSet.Account().BatchCreate(kt, as)
	checkErr(t, err)

	fmt.Println("[ ---- ids ---- ]: ", ids)
}
