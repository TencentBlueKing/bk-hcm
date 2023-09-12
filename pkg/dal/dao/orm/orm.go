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

// Package orm ...
package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
	prm "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

// DoOrm defines all the orm method.
type DoOrm interface {
	Get(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error
	Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error
	Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error)
	Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error)
	Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error)
	Exec(ctx context.Context, expr string) (int64, error)

	Insert(ctx context.Context, expr string, data interface{}) error
	BulkInsert(ctx context.Context, expr string, args interface{}) error
}

// DoOrmWithTransaction defines all the orm method with transaction.
type DoOrmWithTransaction interface {
	Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error)
	Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error
	Delete(ctx context.Context, expr string, args map[string]interface{}) (int64, error)
	Update(ctx context.Context, expr string, args map[string]interface{}) (int64, error)

	Insert(ctx context.Context, expr string, args interface{}) error
	BulkInsert(ctx context.Context, expr string, args interface{}) error
}

// Interface defines all the orm related operations.
type Interface interface {
	Do() DoOrm
	Txn(tx *sqlx.Tx) DoOrmWithTransaction
	AutoTxn(kt *kit.Kit, run TxnFunc) (interface{}, error)
}

// InitOrm return orm operations.
func InitOrm(db *sqlx.DB, opts ...Option) Interface {
	ormOpts := new(options)
	for _, opt := range opts {
		opt(ormOpts)
	}

	if ormOpts.ingressLimiter == nil {
		ormOpts.ingressLimiter = rate.NewLimiter(rate.Limit(500), 500)
	}

	if ormOpts.logLimiter == nil {
		ormOpts.logLimiter = rate.NewLimiter(rate.Limit(50), 25)
	}

	if ormOpts.mc == nil {
		ormOpts.mc = initMetric(prm.DefaultRegisterer)
	}

	if ormOpts.slowRequestMS == 0 {
		ormOpts.slowRequestMS = 50 * time.Millisecond
	}

	return &runtimeOrm{
		db:             db,
		mc:             ormOpts.mc,
		ingressLimiter: ormOpts.ingressLimiter,
		logLimiter:     ormOpts.logLimiter,
		slowRequestMS:  ormOpts.slowRequestMS,
	}
}

type runtimeOrm struct {
	db             *sqlx.DB
	ingressLimiter *rate.Limiter
	logLimiter     *rate.Limiter
	mc             *metric
	slowRequestMS  time.Duration
}

func (o *runtimeOrm) logSlowCmd(ctx context.Context, sql string, latency time.Duration) {
	if latency < o.slowRequestMS {
		return
	}

	if !o.logLimiter.Allow() {
		// if the log rate have already exceeded the limit, then skip the log.
		// we do this to avoid write lots of log to file and slow down the request.
		return
	}

	rid := ctx.Value(constant.RidKey)
	logs.InfoDepthf(2, "[orm slow log], sql: %s, latency: %d ms, rid: %v", sql, latency.Milliseconds(), rid)
}

// tryAccept is used to test if the incoming orm request can be accepted.
// TODO: test the accept for each sharding, but not for all the sharding with one limiter.
func (o *runtimeOrm) tryAccept() error {
	if o.ingressLimiter.Allow() {
		return nil
	}

	o.mc.errCounter.With(prm.Labels{"cmd": "limiter"}).Inc()

	// have already oversize the limit
	return errf.New(errf.TooManyRequest, "orm too many requests")
}

// Do create a new orm do instance.
func (o *runtimeOrm) Do() DoOrm {
	return &do{
		db: o.db,
		ro: o,
	}
}

// Txn create a new transaction orm instance.
func (o *runtimeOrm) Txn(tx *sqlx.Tx) DoOrmWithTransaction {
	return &doTxn{
		tx: tx,
		ro: o,
	}
}

// TxnFunc is a callback function to process logic tasks between a transaction.
type TxnFunc func(txn *sqlx.Tx, opt *TxnOption) (interface{}, error)

// TxnOption defines all the options to do distributed
// transaction in the AutoTxn processes.
type TxnOption struct{}

// ErrRetryTransaction defines errors that need to retry transaction, like deadlock error in upsert scenario
var ErrRetryTransaction = errors.New("RETRY TRANSACTION ERROR")

// AutoTxn is a wrapper to do all the transaction operations as follows:
// 1. auto launch the transaction
// 2. process the logics, which is a callback run function
// 3. rollback the transaction if 'run' hit an error automatically.
// 4. commit the transaction if no error happens.
func (o *runtimeOrm) AutoTxn(kit *kit.Kit, run TxnFunc) (interface{}, error) {
	if run == nil {
		return nil, errors.New("transaction function is nil")
	}

	retry, result, err := o.autoTxn(kit, run)
	if err == nil {
		return result, nil
	}

	if !retry && err != nil {
		return nil, err
	}

	// if the operation need to retry, retry for at most 3 times, each wait for 50~500ms
	for retryCount := 1; retryCount <= 3; retryCount++ {
		logs.Warnf("retry transaction, retry count: %d, rid: %s", retryCount, kit.Rid)
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(450)+50))

		retry, result, err = o.autoTxn(kit, run)
		if err == nil {
			return result, nil
		}

		if !retry && err != nil {
			return nil, err
		}

		// do next retry
	}

	logs.Warnf("retry transaction exceeds maximum count, **skip**, rid: %s", kit.Rid)
	return nil, err
}

func (o *runtimeOrm) autoTxn(kit *kit.Kit, run TxnFunc) (bool, interface{}, error) {
	if run == nil {
		return false, nil, errors.New("transaction function is nil")
	}

	txn, err := o.db.BeginTxx(kit.Ctx, new(sql.TxOptions))
	if err != nil {
		return false, nil, fmt.Errorf("auto txn, but begin txn failed, err: %v", err)
	}

	result, err := run(txn, new(TxnOption))
	if err != nil {
		if rollErr := txn.Rollback(); rollErr != nil {
			logs.ErrorDepthf(1, "run sharding one transaction rollback failed, err: %v, rid: %v", rollErr, kit.Rid)
			// do not return error. the transaction will be aborted automatically after timeout.
			// mysql transaction's default timeout is 50s.
		}

		if err == ErrRetryTransaction {
			return true, nil, err
		}

		return false, nil, err
	}

	if err := txn.Commit(); err != nil {
		return false, nil, fmt.Errorf("commit sharding transaction failed, err: %v", err)
	}

	return false, result, nil
}
