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
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
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

// AccountBizRel 云账号与业务关联关系
type AccountBizRel struct {
	ID        uint64
	BkBizID   int
	AccountID uint64
	Creator   string
	Reviser   string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// Create one account instance.
func (a *AccountBizRel) Create(kt *kit.Kit) (uint64, error) {
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

// Update ...
func (a *AccountBizRel) Update(kt *kit.Kit, expr *filter.Expression, updateFields []string) error {
	td, err := a.toUpdateTableData(kt.User, updateFields)
	if err != nil {
		return nil
	}

	sql, err := td.SQLForUpdate(expr)
	if err != nil {
		return err
	}

	toUpdate := td.FieldKVForUpdate()

	daoClient := dao.DaoClient

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

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List ...
func (a *AccountBizRel) List(kt *kit.Kit, opt *types.ListOption) ([]*AccountBizRel, error) {
	tp := new(tablecloud.AccountBizRelTable)
	listSQL, err := tp.SQLForList(opt, &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
	})
	if err != nil {
		return nil, err
	}

	td := make([]*tablecloud.AccountBizRelTable, 0)
	err = dao.DaoClient.Orm.Do().Select(kt.Ctx, &td, listSQL)
	if err != nil {
		return nil, err
	}

	data := make([]*AccountBizRel, 0)

	for _, d := range td {
		data = append(
			data,
			&AccountBizRel{
				ID:        d.ID,
				BkBizID:   d.BkBizID,
				AccountID: d.AccountID,
				Creator:   d.Creator,
				Reviser:   d.Reviser,
				CreatedAt: d.CreatedAt,
				UpdatedAt: d.UpdatedAt,
			},
		)
	}

	return data, nil
}

// Delete ...
func (a *AccountBizRel) Delete(kt *kit.Kit, expr *filter.Expression) error {
	tp := new(tablecloud.AccountBizRelTable)
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

func (a *AccountBizRel) toCreateTableData(creator string) (*tablecloud.AccountBizRelTable, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}

	return &tablecloud.AccountBizRelTable{
		BkBizID:      a.BkBizID,
		AccountID:    a.AccountID,
		Creator:      creator,
		Reviser:      creator,
		TableManager: &table.TableManager{},
	}, nil
}

func (a *AccountBizRel) toUpdateTableData(
	reviser string,
	updateFields []string,
) (*tablecloud.AccountBizRelTable, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}

	return &tablecloud.AccountBizRelTable{
		BkBizID:      a.BkBizID,
		Reviser:      reviser,
		TableManager: &table.TableManager{UpdateFields: updateFields},
	}, nil
}

func (a *AccountBizRel) validate() error {
	// TODO 校验处理
	return nil
}
