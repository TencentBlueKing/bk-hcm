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

package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"hcm/pkg/tools/runtimes"
	"hcm/pkg/traces"

	"github.com/jmoiron/sqlx"
	prm "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/codes"
)

var (
	// ErrRecordNotFound returns a "record not found error".
	// Occurs only when attempting to query the database with a struct,
	// querying with a slice won't return this error.
	ErrRecordNotFound = sql.ErrNoRows
)

const (
	// ormRuntimeSkip is the runtimes skip value, for getting the caller of ORM functions.
	ormRuntimeSkip = 1
)

var _ DoOrm = new(do)

type do struct {
	db *sqlx.DB
	ro *runtimeOrm
}

// Select a collection of data, and decode into dest *[]struct{}.
func (do *do) Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) (err error) {
	if err = do.ro.tryAccept(); err != nil {
		return err
	}

	// tracing
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(1), runtimes.FuncName(1)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	rows, err := do.db.QueryContext(ctx, do.db.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	if err = sqlx.StructScan(rows, dest); err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "select"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Count the number of the filtered resource.
func (do *do) Count(ctx context.Context, expr string, arg map[string]interface{}) (count uint64, err error) {
	if err = do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))

	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	rows, err := do.db.QueryContext(ctx, do.db.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	count = uint64(0)
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
			return 0, err
		}
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "count"}).Observe(float64(time.Since(start).Milliseconds()))

	return count, nil
}

// Delete a collection of data.
func (do *do) Delete(ctx context.Context, expr string, arg map[string]interface{}) (rowsAffected int64, err error) {
	if err := do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	result, err := do.db.ExecContext(ctx, do.db.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "delete"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}

// Update a collection of data
func (do *do) Update(ctx context.Context, expr string, arg map[string]interface{}) (rowsAffected int64, err error) {
	if arg == nil {
		return 0, errors.New("update args is required")
	}

	if err = do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	result, err := do.db.ExecContext(ctx, do.db.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "update"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}

// Insert a row data to db
func (do *do) Insert(ctx context.Context, expr string, data interface{}) (err error) {
	if err = do.ro.tryAccept(); err != nil {
		return err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	_, err = do.db.NamedExecContext(ctx, expr, data)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Exec a command
func (do *do) Exec(ctx context.Context, expr string) (effected int64, err error) {
	if err = do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	result, err := do.db.ExecContext(ctx, expr)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "exec"}).Inc()
		return 0, err
	}

	effected, err = result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "exec"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "exec"}).Observe(float64(time.Since(start).Milliseconds()))

	return effected, nil
}

// BulkInsert insert multiple data at one time, the order in which ids is returned
// is the same as the order in which data is inserted.
func (do *do) BulkInsert(ctx context.Context, expr string, args interface{}) (err error) {
	if err = do.ro.tryAccept(); err != nil {
		return err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	_, err = do.db.NamedExecContext(ctx, expr, args)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "bulk-insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "bulk-insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

var _ DoOrmWithTransaction = new(doTxn)

type doTxn struct {
	tx *sqlx.Tx
	ro *runtimeOrm
}

// Count the number of the filtered resource.
func (do *doTxn) Count(ctx context.Context, expr string, arg map[string]interface{}) (count uint64, err error) {
	if err = do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	rows, err := do.tx.QueryContext(ctx, do.tx.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	count = uint64(0)
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
			return 0, err
		}
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "count"}).Observe(float64(time.Since(start).Milliseconds()))

	return count, nil
}

// Select a collection of data, and decode into dest *[]struct{}.
func (do *doTxn) Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) (err error) {
	if err = do.ro.tryAccept(); err != nil {
		return err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	rows, err := do.tx.QueryContext(ctx, do.tx.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	if err = sqlx.StructScan(rows, dest); err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "select"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Delete a collection of data with transaction.
func (do *doTxn) Delete(ctx context.Context, expr string, arg map[string]interface{}) (rowsAffected int64, err error) {
	if err = do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	result, err := do.tx.ExecContext(ctx, do.tx.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "delete"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}

// Insert data with transaction
func (do *doTxn) Insert(ctx context.Context, expr string, args interface{}) (err error) {
	if args == nil {
		return errors.New("insert args is required")
	}

	if err = do.ro.tryAccept(); err != nil {
		return err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	_, err = do.tx.NamedExecContext(ctx, expr, args)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// BulkInsert insert data batch with transaction, the order in which ids is
// returned is the same as the order in which data is inserted.
func (do *doTxn) BulkInsert(ctx context.Context, expr string, args interface{}) (err error) {
	if err = do.ro.tryAccept(); err != nil {
		return err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	_, err = do.tx.NamedExecContext(ctx, expr, args)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "bulk-insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "bulk-insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Update with transaction
func (do *doTxn) Update(ctx context.Context, expr string, arg map[string]interface{}) (rowsAffected int64, err error) {
	if arg == nil {
		return 0, errors.New("update args is required")
	}

	if err = do.ro.tryAccept(); err != nil {
		return 0, err
	}

	// tracing, if the function stacks are changed, remember to change the "skip" value
	_, span := traces.StartCtx(ctx, fmt.Sprintf("%s.%s", runtimes.StructName(ormRuntimeSkip),
		runtimes.FuncName(ormRuntimeSkip)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	start := time.Now()

	query, args, err := sqlx.Named(expr, arg)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	result, err := do.tx.ExecContext(ctx, do.tx.Rebind(query), args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "update"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}
