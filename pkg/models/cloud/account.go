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

package cloud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

type Account struct {
	ID           uint64      `json:"id"`
	Name         string      `json:"name"`
	Vendor       string      `json:"vendor"`
	Managers     []string    `json:"managers"`
	DepartmentID int         `json:"department_id"`
	Type         string      `json:"type"`
	SyncStatus   string      `json:"sync_status"`
	Price        string      `json:"price"`
	PriceUnit    string      `json:"price_unit"`
	Extension    interface{} `json:"extension"`
	Creator      string      `json:"creator"`
	Reviser      string      `json:"reviser"`
	CreatedAt    *time.Time  `json:"created_at"`
	UpdatedAt    *time.Time  `json:"updated_at"`
	Memo         string      `json:"memo"`
}

// Create one account instance.
func (a *Account) Create(kt *kit.Kit) (uint64, error) {
	td, err := a.toCreateTableData(kt.User)
	if err != nil {
		return 0, nil
	}

	sql := td.GenerateInsertSQL()

	daoClient := dao.DaoClient
	result, err := daoClient.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		id, err := daoClient.Orm.Txn(txn).Insert(kt.Ctx, sql, td)
		if err != nil {
			return 0, fmt.Errorf("insert account failed, err: %v", err)
		}

		td.ID = id

		return id, nil
	})
	if err != nil {
		logs.Errorf("create account, but do auto txn failed, err: %v, rid: %s", err, kt.Rid)
		return 0, fmt.Errorf("create account, but auto run txn failed, err: %v", err)
	}

	id, ok := result.(uint64)
	if !ok {
		logs.Errorf("insert account return id type not is uint64, id type: %v, rid: %s",
			reflect.TypeOf(result).String(), kt.Rid)
	}

	return id, nil
}

func (a *Account) Update(kt *kit.Kit, expr *filter.Expression, updateFields []string) error {
	td, err := a.toUpdateTableData(kt.User, updateFields)
	if err != nil {
		return nil
	}

	sql, err := td.GenerateUpdateSQL(expr)
	if err != nil {
		return err
	}

	whereExpr, _ := table.GenerateWhereExpr(expr)
	toUpdate := td.GenerateUpdateFieldKV()

	daoClient := dao.DaoClient

	ab := daoClient.AuditDao.Decorator(kt, enumor.Account).PrepareUpdate(whereExpr, toUpdate)

	_, err = daoClient.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := daoClient.Orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update account: %d failed, err: %v, rid: %v", td.ID, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account, but record not found, filter: %v, rid: %v", expr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		if err := ab.Do(txn); err != nil {
			return nil, fmt.Errorf("do account update audit failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Account) List(kt *kit.Kit, opt *types.ListOption) ([]*Account, error) {
	tp := new(tablecloud.AccountTable)
	listSQL, err := tp.GenerateListSQL(opt)
	if err != nil {
		return nil, err
	}

	td := make([]*tablecloud.AccountTable, 0)
	err = dao.DaoClient.Orm.Do().Select(kt.Ctx, &td, listSQL)
	if err != nil {
		return nil, err
	}

	data := make([]*Account, 0)
	var managers []string
	var ext interface{}

	for _, d := range td {
		err = json.Unmarshal([]byte(d.Managers), &managers)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(d.Extension), &ext)
		if err != nil {
			return nil, err
		}

		data = append(
			data,
			&Account{
				ID:           d.ID,
				Name:         d.Name,
				Vendor:       d.Vendor,
				Managers:     managers,
				DepartmentID: d.DepartmentID,
				Type:         d.Type,
				SyncStatus:   d.SyncStatus,
				Price:        d.Price,
				PriceUnit:    d.PriceUnit,
				Extension:    ext,
				Creator:      d.Creator,
				Reviser:      d.Reviser,
				CreatedAt:    d.CreatedAt,
				UpdatedAt:    d.UpdatedAt,
				Memo:         d.Memo,
			},
		)
	}

	return data, nil
}

func (a *Account) Delete(kt *kit.Kit, expr *filter.Expression) error {
	tp := new(tablecloud.AccountTable)
	sql, err := tp.GenerateDeleteSQL(expr)
	if err != nil {
		return err
	}

	daoClient := dao.DaoClient

	_, err = daoClient.Orm.AutoTxn(kt, func(txn *sqlx.Tx, option *orm.TxnOption) (interface{}, error) {
		// delete the account at first.
		err := daoClient.Orm.Txn(txn).Delete(kt.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.ErrorJson("delete account failed, filter: %v, err: %v, rid: %v", expr, err, kt.Rid)
		return fmt.Errorf("delete account, but run txn failed, err: %v", err)
	}

	return nil
}

func (a *Account) toCreateTableData(creator string) (*tablecloud.AccountTable, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}

	managers, err := json.Marshal(a.Managers)
	if err != nil {
		return nil, err
	}

	ext, err := json.Marshal(a.Extension)
	if err != nil {
		return nil, err
	}

	return &tablecloud.AccountTable{
		Name:         a.Name,
		Vendor:       a.Vendor,
		DepartmentID: a.DepartmentID,
		Managers:     table.JsonField(managers),
		Extension:    table.JsonField(ext),
		Creator:      creator,
		Reviser:      creator,
		SyncStatus:   "", // 账号初始状态设置
		TableManager: &table.TableManager{},
	}, nil
}

func (a *Account) toUpdateTableData(reviser string, updateFields []string) (*tablecloud.AccountTable, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}

	managers, err := json.Marshal(a.Managers)
	if err != nil {
		return nil, err
	}

	ext, err := json.Marshal(a.Extension)
	if err != nil {
		return nil, err
	}

	return &tablecloud.AccountTable{
		Name:         a.Name,
		Managers:     table.JsonField(managers),
		Price:        a.Price,
		PriceUnit:    a.PriceUnit,
		DepartmentID: a.DepartmentID,
		Extension:    table.JsonField(ext),
		Memo:         a.Memo,
		Reviser:      reviser,
		TableManager: &table.TableManager{UpdateFields: updateFields},
	}, nil
}

func (a *Account) validate() error {
	// TODO 校验处理
	return nil
}
