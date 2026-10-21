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

package engine

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"hcm/migrate/schema"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"

	"github.com/jmoiron/sqlx"
)

// fakeCall is one recorded call to the fake orm.
type fakeCall struct {
	op   string
	expr string
	arg  interface{}
}

// countReply is one scripted answer for Count, consumed in order.
// util.MetaOrm.HasTable issues
//
//	SELECT COUNT(*) FROM information_schema.tables
//	WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' AND table_name = :table
//
// with arg {"table": name}. A count greater than zero means the table exists.
type countReply struct {
	n   uint64
	err error
}

// fakeDo is an in-memory stand-in for orm.DoOrm. Select fills *[]schema.Record, or
// when selectNonRecord is set, any other pointer-to-slice dest whose element
// type matches. Update returns updateAffected unless updateErr is set. Count
// answers from countQueue and then falls back to countErr or countDefault.
type fakeDo struct {
	calls []fakeCall

	countQueue   []countReply
	countErr     error
	countDefault uint64

	execErr      error
	execErrMatch string

	selectRows      []schema.Record
	selectNonRecord interface{}
	selectErr       error

	insertErr      error
	insertErrs     []error
	updateAffected int64
	updateErr      error

	bulkErr error
}

func newFakeDo() *fakeDo {
	return &fakeDo{updateAffected: 1}
}

func (f *fakeDo) record(op, expr string, arg interface{}) {
	f.calls = append(f.calls, fakeCall{op: op, expr: expr, arg: cloneArg(arg)})
}

func (f *fakeDo) callsOf(op string) []fakeCall {
	out := make([]fakeCall, 0)
	for _, c := range f.calls {
		if c.op == op {
			out = append(out, c)
		}
	}
	return out
}

func cloneArg(arg interface{}) interface{} {
	m, ok := arg.(map[string]interface{})
	if !ok {
		return arg
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (f *fakeDo) Select(_ context.Context, dest interface{}, expr string, arg map[string]interface{}) error {
	f.record("select", expr, arg)
	if f.selectErr != nil {
		return f.selectErr
	}
	if rows, ok := dest.(*[]schema.Record); ok {
		copied := make([]schema.Record, len(f.selectRows))
		copy(copied, f.selectRows)
		*rows = copied
		return nil
	}
	if f.selectNonRecord != nil {
		dv := reflect.ValueOf(dest)
		if dv.Kind() != reflect.Pointer || dv.Elem().Kind() != reflect.Slice {
			return fmt.Errorf("fakeDo: Select dest %T not a pointer to slice", dest)
		}
		src := reflect.ValueOf(f.selectNonRecord)
		if src.Kind() != reflect.Slice {
			return fmt.Errorf("fakeDo: selectNonRecord %T is not a slice", f.selectNonRecord)
		}
		if !src.Type().AssignableTo(dv.Elem().Type()) {
			return fmt.Errorf("fakeDo: Select dest %T, selectNonRecord is %T", dest, f.selectNonRecord)
		}
		dv.Elem().Set(src)
		return nil
	}
	return fmt.Errorf("fakeDo: Select dest %T, want *[]schema.Record", dest)
}

func (f *fakeDo) Count(_ context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	f.record("count", expr, arg)
	if len(f.countQueue) > 0 {
		reply := f.countQueue[0]
		f.countQueue = f.countQueue[1:]
		return reply.n, reply.err
	}
	if f.countErr != nil {
		return 0, f.countErr
	}
	return f.countDefault, nil
}

func (f *fakeDo) Delete(_ context.Context, expr string, arg map[string]interface{}) (int64, error) {
	f.record("delete", expr, arg)
	return 0, fmt.Errorf("fakeDo: Delete not implemented")
}

func (f *fakeDo) Update(_ context.Context, expr string, arg map[string]interface{}) (int64, error) {
	f.record("update", expr, arg)
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	return f.updateAffected, nil
}

func (f *fakeDo) Exec(_ context.Context, expr string) (int64, error) {
	f.record("exec", expr, nil)
	if f.execErr != nil && (f.execErrMatch == "" || strings.Contains(expr, f.execErrMatch)) {
		return 0, f.execErr
	}
	return 1, nil
}

func (f *fakeDo) Insert(_ context.Context, expr string, data interface{}) error {
	f.record("insert", expr, data)
	if len(f.insertErrs) > 0 {
		err := f.insertErrs[0]
		f.insertErrs = f.insertErrs[1:]
		return err
	}
	return f.insertErr
}

func (f *fakeDo) BulkInsert(_ context.Context, expr string, args interface{}) error {
	f.record("bulk-insert", expr, args)
	return f.bulkErr
}

// fakeOrm is a minimal orm.Interface. This package only calls Do().
type fakeOrm struct {
	do *fakeDo
}

func newFakeOrm(do *fakeDo) orm.Interface {
	if do == nil {
		do = newFakeDo()
	}
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
