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

	"github.com/jmoiron/sqlx"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// Account supplies all the cloud account related operations.
type AccountBizRel interface {
	Create(kt *kit.Kit, rel *tablecloud.AccountBizRelModel) (uint64, error)
	Update(kt *kit.Kit, expr *filter.Expression, rel *tablecloud.AccountBizRelModel) error
	List(kt *kit.Kit, opt *types.ListOption) ([]*tablecloud.AccountBizRelModel, error)
	Delete(kt *kit.Kit, expr *filter.Expression, rel *tablecloud.AccountBizRelModel) error
}

var _ AccountBizRel = new(AccountBizRelDao)

type AccountBizRelDao struct {
	orm      orm.Interface
	auditDao audit.AuditDao
}

func NewAccountBizRelDao(orm orm.Interface, auditDao audit.AuditDao) *AccountBizRelDao {
	return &AccountBizRelDao{orm, auditDao}
}

// Create one account instance.
func (ad *AccountBizRelDao) Create(kt *kit.Kit, rel *tablecloud.AccountBizRelModel) (uint64, error) {
	sql := rel.GenerateInsertSQL()

	result, err := ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		id, err := ad.orm.Txn(txn).Insert(kt.Ctx, sql, rel)
		if err != nil {
			return 0, fmt.Errorf("insert account_biz_rel failed, err: %v", err)
		}

		rel.ID = id
		return id, nil
	})
	if err != nil {
		logs.Errorf("create account_biz_rel, but do auto txn failed, err: %v, rid: %s", err, kt.Rid)
		return 0, fmt.Errorf("create account_biz_rel, but auto run txn failed, err: %v", err)
	}

	id, ok := result.(uint64)
	if !ok {
		logs.Errorf("insert account_biz_rel return id type not is uint64, id type: %v, rid: %s",
			reflect.TypeOf(result).String(), kt.Rid)
	}

	return id, nil
}

func (ad *AccountBizRelDao) Update(kt *kit.Kit, expr *filter.Expression, rel *tablecloud.AccountBizRelModel) error {
	sql, err := rel.GenerateUpdateSQL(expr)
	if err != nil {
		return err
	}

	toUpdate := rel.GenerateUpdateFieldKV()

	_, err = ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := ad.orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update account_biz_rel: %d failed, err: %v, rid: %v", rel.ID, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account_biz_rel, but record not found, filter: %v, rid: %v", expr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ad *AccountBizRelDao) List(kt *kit.Kit, opt *types.ListOption) ([]*tablecloud.AccountBizRelModel, error) {
	rel := new(tablecloud.AccountBizRelModel)
	listSQL, err := rel.GenerateListSQL(opt)
	if err != nil {
		return nil, err
	}

	rels := make([]*tablecloud.AccountBizRelModel, 0)
	err = ad.orm.Do().Select(kt.Ctx, &rels, listSQL)
	if err != nil {
		return nil, err
	}

	return rels, nil
}

func (ad *AccountBizRelDao) Delete(kt *kit.Kit, expr *filter.Expression, rel *tablecloud.AccountBizRelModel) error {
	sql, err := rel.GenerateDeleteSQL(expr)
	if err != nil {
		return err
	}
	_, err = ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, option *orm.TxnOption) (interface{}, error) {
		// delete the account at first.
		err := ad.orm.Txn(txn).Delete(kt.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.ErrorJson("delete account_biz_rel failed, filter: %v, err: %v, rid: %v", expr, err, kt.Rid)
		return fmt.Errorf("delete account_biz_rel, but run txn failed, err: %v", err)
	}

	return nil
}
