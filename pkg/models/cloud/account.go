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

// 云账号
type Account struct {
	ID           uint64
	Name         string
	Vendor       string
	Managers     []string
	DepartmentID int
	Type         string
	SyncStatus   string
	Price        string
	PriceUnit    string
	Extension    interface{}
	Creator      string
	Reviser      string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Memo         string
}

// Create one account instance.
func (a *Account) Create(kt *kit.Kit) (uint64, error) {
	td, err := a.toCreateTableData(kt.User)
	if err != nil {
		return 0, nil
	}

	sql := td.SQLForInsert()

	daoClient := dao.DaoClient
	result, err := daoClient.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		id, err := daoClient.Orm.Txn(txn).Insert(kt.Ctx, sql, td)
		if err != nil {
			return 0, fmt.Errorf("insert %s failed, err: %v", td.TableName(), err)
		}

		td.ID = id

		return id, nil
	})
	if err != nil {
		logs.Errorf("create %s, but do auto txn failed, err: %v, rid: %s", td.TableName(), err, kt.Rid)
		return 0, fmt.Errorf("create %s, but auto run txn failed, err: %v", td.TableName(), err)
	}

	id, ok := result.(uint64)
	if !ok {
		logs.Errorf("insert %s return id type not is uint64, id type: %v, rid: %s", td.TableName(),
			reflect.TypeOf(result).String(), kt.Rid)
	}

	return id, nil
}

//  Update ...
func (a *Account) Update(kt *kit.Kit, expr *filter.Expression, updateFields []string) error {
	td, err := a.toUpdateTableData(kt.User, updateFields)
	if err != nil {
		return nil
	}

	sql, err := td.SQLForUpdate(expr)
	if err != nil {
		return err
	}

	whereExpr, _ := table.SQLWhereExpr(expr, nil)
	toUpdate := td.FieldKVForUpdate()

	daoClient := dao.DaoClient

	ab := daoClient.AuditDao.Decorator(kt, enumor.Account).PrepareUpdate(whereExpr, toUpdate)

	_, err = daoClient.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := daoClient.Orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update %s: %d failed, err: %v, rid: %v", td.TableName(), td.ID, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update %s, but record not found, filter: %v, rid: %v", td.TableName(), expr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		if err := ab.Do(txn); err != nil {
			return nil, fmt.Errorf("do %s update audit failed, err: %v", td.TableName(), err)
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List ...
func (a *Account) List(kt *kit.Kit, opt *types.ListOption) ([]*Account, error) {
	tp := new(tablecloud.AccountTable)
	listSQL, err := tp.SQLForList(opt, &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
	})
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

// Delete ...
func (a *Account) Delete(kt *kit.Kit, expr *filter.Expression) error {
	tp := new(tablecloud.AccountTable)
	sql, err := tp.SQLForDelete(expr)
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
		logs.ErrorJson("delete %s failed, filter: %v, err: %v, rid: %v", tp.TableName(), expr, err, kt.Rid)
		return fmt.Errorf("delete %s, but run txn failed, err: %v", tp.TableName(), err)
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
