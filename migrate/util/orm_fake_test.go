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

package util

import (
	"context"
	"fmt"

	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"

	"github.com/jmoiron/sqlx"
)

// fakeDo is an in-memory stand-in for orm.DoOrm used by this package's unit
// tests, so MetaOrm/idempotent helpers can be tested without a real MySQL.
type fakeDo struct {
	// countFn answers Count calls, keyed by the caller-supplied expr/arg.
	countFn func(expr string, arg map[string]interface{}) (uint64, error)
	// selectFn answers Select calls.
	selectFn func(dest interface{}, expr string, arg map[string]interface{}) error
	// execErr, when non-nil, makes every Exec call fail.
	execErr error
	// insertErr, when non-nil, makes every Insert call fail.
	insertErr error

	// execCalls records every ddl string passed to Exec, in call order.
	execCalls []string
	// insertCalls records every (expr, data) pair passed to Insert, in call order.
	insertCalls []fakeInsertCall
}

type fakeInsertCall struct {
	expr string
	data interface{}
}

func (f *fakeDo) Select(_ context.Context, dest interface{}, expr string, arg map[string]interface{}) error {
	if f.selectFn == nil {
		return fmt.Errorf("fakeDo: unexpected Select call, expr: %s", expr)
	}
	return f.selectFn(dest, expr, arg)
}

func (f *fakeDo) Count(_ context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	if f.countFn == nil {
		return 0, fmt.Errorf("fakeDo: unexpected Count call, expr: %s", expr)
	}
	return f.countFn(expr, arg)
}

func (f *fakeDo) Delete(_ context.Context, _ string, _ map[string]interface{}) (int64, error) {
	panic("fakeDo: Delete not implemented")
}

func (f *fakeDo) Update(_ context.Context, _ string, _ map[string]interface{}) (int64, error) {
	panic("fakeDo: Update not implemented")
}

func (f *fakeDo) Exec(_ context.Context, expr string) (int64, error) {
	f.execCalls = append(f.execCalls, expr)
	if f.execErr != nil {
		return 0, f.execErr
	}
	return 1, nil
}

func (f *fakeDo) Insert(_ context.Context, expr string, data interface{}) error {
	f.insertCalls = append(f.insertCalls, fakeInsertCall{expr: expr, data: data})
	return f.insertErr
}

func (f *fakeDo) BulkInsert(_ context.Context, _ string, _ interface{}) error {
	panic("fakeDo: BulkInsert not implemented")
}

// fakeOrm is a minimal orm.Interface backed by a fakeDo, only Do() is wired
// up: this package's helpers only ever call orm.Interface.Do().
type fakeOrm struct {
	do *fakeDo
}

func newFakeOrm(do *fakeDo) orm.Interface {
	return &fakeOrm{do: do}
}

func (f *fakeOrm) Do() orm.DoOrm {
	return f.do
}

func (f *fakeOrm) Txn(_ *sqlx.Tx) orm.DoOrmWithTransaction {
	panic("fakeOrm: Txn not implemented")
}

func (f *fakeOrm) AutoTxn(_ *kit.Kit, _ orm.TxnFunc) (interface{}, error) {
	panic("fakeOrm: AutoTxn not implemented")
}

func (f *fakeOrm) TableSharding(_ ...orm.TableShardingOpt) orm.Interface {
	panic("fakeOrm: TableSharding not implemented")
}

func (f *fakeOrm) ModifySQLOpts(_ ...orm.ModifySQLOpt) orm.Interface {
	panic("fakeOrm: ModifySQLOpts not implemented")
}
