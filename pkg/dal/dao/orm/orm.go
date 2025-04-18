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
	"regexp"
	"strings"
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
	Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error
	Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error)
	Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error)
	Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error)
	// Exec 通过modifySQLDo的Exec方法执行sql时，不支持多租户ID的替换
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
	// TableSharding at least one TableSharding option
	TableSharding(opts ...TableShardingOpt) Interface
	// ModifySQLOpts modify sql options
	ModifySQLOpts(opts ...ModifySQLOpt) Interface
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

// TableSharding ...
func (o *runtimeOrm) TableSharding(opts ...TableShardingOpt) Interface {
	return &tableShardingOrm{
		orm:               o,
		tableShardingOpts: opts,
	}

}

// ModifySQLOpts ...
func (o *runtimeOrm) ModifySQLOpts(opts ...ModifySQLOpt) Interface {
	return &modifySQLOrm{
		orm:           o,
		modifySQLOpts: opts,
	}
}

// tableShardingOrm orm for table sharding, it replaces table name in sql query
type tableShardingOrm struct {
	orm               Interface
	tableShardingOpts []TableShardingOpt
}

type tableShardingDo struct {
	do                DoOrm
	tableShardingOpts []TableShardingOpt
}

type tableShardingDoTxn struct {
	doTxn             DoOrmWithTransaction
	tableShardingOpts []TableShardingOpt
}

// Count ...
func (dt *tableShardingDoTxn) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	replaced := replaceFromJoinTableName(dt.tableShardingOpts, expr)
	return dt.doTxn.Count(ctx, replaced, arg)
}

// Select ...
func (dt *tableShardingDoTxn) Select(ctx context.Context, dest interface{}, expr string,
	arg map[string]interface{}) error {

	replaced := replaceFromJoinTableName(dt.tableShardingOpts, expr)
	return dt.doTxn.Select(ctx, dest, replaced, arg)
}

// Delete ...
func (dt *tableShardingDoTxn) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceFromJoinTableName(dt.tableShardingOpts, expr)
	return dt.doTxn.Delete(ctx, replaced, arg)
}

// Update ...
func (dt *tableShardingDoTxn) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceUpdateTableName(dt.tableShardingOpts, expr)
	return dt.doTxn.Update(ctx, replaced, arg)
}

// Insert ...
func (dt *tableShardingDoTxn) Insert(ctx context.Context, expr string, data interface{}) error {
	replaced := replaceInsertTableName(dt.tableShardingOpts, expr)
	return dt.doTxn.BulkInsert(ctx, replaced, data)
}

// BulkInsert ...
func (dt *tableShardingDoTxn) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	replaced := replaceInsertTableName(dt.tableShardingOpts, expr)
	return dt.doTxn.BulkInsert(ctx, replaced, args)
}

// Count ...
func (ds *tableShardingDo) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)
	return ds.do.Count(ctx, replaced, arg)
}

// Delete ...
func (ds *tableShardingDo) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)
	return ds.do.Delete(ctx, replaced, arg)
}

// Update ...
func (ds *tableShardingDo) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceUpdateTableName(ds.tableShardingOpts, expr)
	return ds.do.Update(ctx, replaced, arg)
}

// Exec ...
func (ds *tableShardingDo) Exec(ctx context.Context, expr string) (int64, error) {
	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)

	replaced = replaceUpdateTableName(ds.tableShardingOpts, replaced)

	replaced = replaceInsertTableName(ds.tableShardingOpts, replaced)

	return ds.do.Exec(ctx, replaced)
}

// Insert ...
func (ds *tableShardingDo) Insert(ctx context.Context, expr string, data interface{}) error {
	replaced := replaceInsertTableName(ds.tableShardingOpts, expr)

	return ds.do.BulkInsert(ctx, replaced, data)
}

// BulkInsert ...
func (ds *tableShardingDo) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	replaced := replaceInsertTableName(ds.tableShardingOpts, expr)
	return ds.do.BulkInsert(ctx, replaced, args)
}

// Select ...
func (ds *tableShardingDo) Select(ctx context.Context, dest interface{}, expr string,
	arg map[string]interface{}) error {

	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)

	return ds.do.Select(ctx, dest, replaced, arg)
}

// Do ...
func (t tableShardingOrm) Do() DoOrm {
	return &tableShardingDo{
		do:                t.orm.Do(),
		tableShardingOpts: t.tableShardingOpts,
	}
}

// Txn ...
func (t tableShardingOrm) Txn(tx *sqlx.Tx) DoOrmWithTransaction {
	return &tableShardingDoTxn{
		doTxn:             t.orm.Txn(tx),
		tableShardingOpts: t.tableShardingOpts,
	}
}

// AutoTxn ...
func (t tableShardingOrm) AutoTxn(kt *kit.Kit, run TxnFunc) (interface{}, error) {
	return t.orm.AutoTxn(kt, run)
}

// TableSharding ...
func (t tableShardingOrm) TableSharding(opts ...TableShardingOpt) Interface {
	return &tableShardingOrm{
		orm:               t,
		tableShardingOpts: opts,
	}
}

// ModifySQLOpts ...
func (t tableShardingOrm) ModifySQLOpts(opts ...ModifySQLOpt) Interface {
	return &modifySQLOrm{
		orm:           t,
		modifySQLOpts: opts,
	}
}

func replaceUpdateTableName(shardingOpts []TableShardingOpt, origin string) (replaced string) {
	updateTables := updateTableNameRe.FindAllString(origin, -1)
	if len(updateTables) == 0 {
		return origin
	}
	replaced = origin
	// extract any table name after `UPDATE`
	for _, updateTable := range updateTables {
		// trim `update` prefix with trailing spaces
		updateTable = updateTable[6:]
		replaced = replaceTableName(shardingOpts, replaced, updateTable)
	}
	return replaced
}

func replaceInsertTableName(shardingOpts []TableShardingOpt, origin string) (replaced string) {
	insertNames := insertTableNameRe.FindAllString(origin, -1)
	if len(insertNames) == 0 {
		return origin
	}
	replaced = origin
	// extract any table name after `INSERT INTO`
	for _, insertName := range insertNames {
		// trim `insert` prefix with trailing spaces
		insertName = insertName[6:]
		insertName = strings.TrimSpace(insertName)

		// trim `into` prefix with trailing spaces
		insertName = insertName[4:]

		replaced = replaceTableName(shardingOpts, replaced, insertName)
	}
	return replaced
}

func replaceFromJoinTableName(shardingOpts []TableShardingOpt, origin string) (replaced string) {
	fromNames := fromTableNameRe.FindAllString(origin, -1)
	if len(fromNames) == 0 {
		return origin
	}
	replaced = origin
	// extract any table name after `FROM`
	for _, fromName := range fromNames {
		// trim ` from` prefix
		fromName = fromName[5:]
		replaced = replaceTableName(shardingOpts, replaced, fromName)
	}

	// extract table name after `JOIN`
	joinNames := joinTableNameRe.FindAllString(replaced, -1)
	if len(joinNames) == 0 {
		return replaced
	}
	for _, joinName := range joinNames {
		// trim ` join` prefix
		joinName = joinName[5:]
		// replace all table names
		replaced = replaceTableName(shardingOpts, replaced, joinName)
	}
	return replaced
}

// replaceTableName replace table name in sql by sharding options
func replaceTableName(shardingOpts []TableShardingOpt, origin string, tableName string) string {
	tableName = extractTableName(tableName)
	for _, shardingOpt := range shardingOpts {
		// only match one option
		if shardingOpt.Match(tableName) {
			replaced := strings.Replace(origin, tableName, shardingOpt.ReplaceTableName(tableName), -1)
			return replaced
		}
	}
	return origin
}

// unquote and remove database name
func extractTableName(name string) string {

	replaced := strings.TrimSpace(name)

	if len(replaced) < 2 {
		return replaced
	}
	// split database name
	if idxDot := strings.IndexByte(replaced, '.'); idxDot >= 0 {
		replaced = replaced[idxDot+1:]
	}

	// remove backquote
	if len(replaced) > 0 && replaced[0] == '`' && replaced[len(replaced)-1] == '`' {
		replaced = replaced[1 : len(replaced)-1]
	}

	return replaced
}

const (
	// fromJoinMatchRe FROM/JOIN 匹配组
	fromJoinMatchRe = `(?i)(?:FROM|JOIN)\s+`
	// tableNameMatchRe 表名匹配组
	tableNameMatchRe = "((?:`?\\w+`?\\.)?`?\\w+`?)"
	// aliasNameMatchRe 别名匹配组
	aliasNameMatchRe = `(?:\s+(?:AS\s+)?("` + `[^"` + `]+"` + `|` + "`" + `[^` + "`" + `]+` + "`" + `|\w+))?`
)

// regular expression for extracting table name
var (
	fromTableNameRe   = regexp.MustCompile("(?i)\\sfrom\\s+(`?\\w+`?\\.)?`?\\w+`?")
	joinTableNameRe   = regexp.MustCompile("(?i)\\sjoin\\s+(`?\\w+`?\\.)?`?\\w+`?")
	updateTableNameRe = regexp.MustCompile("(?i)update\\s+(`?\\w+`?\\.)?`?\\w+`?")
	insertTableNameRe = regexp.MustCompile("(?i)insert\\s+into\\s+(`?\\w+`?\\.)?`?\\w+`?")

	fromJoinTableNameRe = regexp.MustCompile(fromJoinMatchRe + // FROM/JOIN 匹配组
		tableNameMatchRe + // 表名匹配组
		aliasNameMatchRe, // 别名匹配组
	)
)

// modifySQLOrm orm for modify sql.
type modifySQLOrm struct {
	orm           Interface
	modifySQLOpts []ModifySQLOpt
}

type modifySQLDo struct {
	do            DoOrm
	modifySQLOpts []ModifySQLOpt
}

type modifySQLDoTxn struct {
	doTxn         DoOrmWithTransaction
	modifySQLOpts []ModifySQLOpt
}

// TableSharding ...
func (t modifySQLOrm) TableSharding(opts ...TableShardingOpt) Interface {
	return &tableShardingOrm{
		orm:               t,
		tableShardingOpts: opts,
	}
}

// ModifySQLOpts ...
func (t modifySQLOrm) ModifySQLOpts(opts ...ModifySQLOpt) Interface {
	return &modifySQLOrm{
		orm:           t,
		modifySQLOpts: opts,
	}
}

// Count ...
func (dt *modifySQLDoTxn) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	for _, opt := range dt.modifySQLOpts {
		expr, arg = opt.InjectJoinSQL(expr, arg)
	}

	return dt.doTxn.Count(ctx, expr, arg)
}

// Select ...
func (dt *modifySQLDoTxn) Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error {
	for _, opt := range dt.modifySQLOpts {
		expr, arg = opt.InjectJoinSQL(expr, arg)
	}

	return dt.doTxn.Select(ctx, dest, expr, arg)
}

// Delete ...
func (dt *modifySQLDoTxn) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	for _, opt := range dt.modifySQLOpts {
		expr, arg = opt.InjectDeleteSQL(expr, arg)
	}

	return dt.doTxn.Delete(ctx, expr, arg)
}

// Update ...
func (dt *modifySQLDoTxn) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	for _, opt := range dt.modifySQLOpts {
		expr, arg = opt.InjectUpdateSQL(expr, arg)
	}

	return dt.doTxn.Update(ctx, expr, arg)
}

// Insert ...
func (dt *modifySQLDoTxn) Insert(ctx context.Context, expr string, data interface{}) error {
	for _, opt := range dt.modifySQLOpts {
		expr, data = opt.InjectInsertSQL(expr, data)
	}

	return dt.doTxn.BulkInsert(ctx, expr, data)
}

// BulkInsert ...
func (dt *modifySQLDoTxn) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	for _, opt := range dt.modifySQLOpts {
		expr, args = opt.InjectInsertSQL(expr, args)
	}

	return dt.doTxn.BulkInsert(ctx, expr, args)
}

// Count ...
func (ds *modifySQLDo) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	for _, opt := range ds.modifySQLOpts {
		expr, arg = opt.InjectJoinSQL(expr, arg)
	}

	return ds.do.Count(ctx, expr, arg)
}

// Delete ...
func (ds *modifySQLDo) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	for _, opt := range ds.modifySQLOpts {
		expr, arg = opt.InjectDeleteSQL(expr, arg)
	}

	return ds.do.Delete(ctx, expr, arg)
}

// Update ...
func (ds *modifySQLDo) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	for _, opt := range ds.modifySQLOpts {
		expr, arg = opt.InjectUpdateSQL(expr, arg)
	}

	return ds.do.Update(ctx, expr, arg)
}

// Exec 通过modifySQLDo的Exec方法执行sql时，不支持多租户ID的替换
func (ds *modifySQLDo) Exec(ctx context.Context, expr string) (int64, error) {
	return ds.do.Exec(ctx, expr)
}

// Insert ...
func (ds *modifySQLDo) Insert(ctx context.Context, expr string, data interface{}) error {
	for _, opt := range ds.modifySQLOpts {
		expr, data = opt.InjectInsertSQL(expr, data)
	}

	return ds.do.BulkInsert(ctx, expr, data)
}

// BulkInsert ...
func (ds *modifySQLDo) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	for _, opt := range ds.modifySQLOpts {
		expr, args = opt.InjectInsertSQL(expr, args)
	}

	return ds.do.BulkInsert(ctx, expr, args)
}

// Select ...
func (ds *modifySQLDo) Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error {
	for _, opt := range ds.modifySQLOpts {
		expr, arg = opt.InjectJoinSQL(expr, arg)
	}

	return ds.do.Select(ctx, dest, expr, arg)
}

// Do ...
func (t modifySQLOrm) Do() DoOrm {
	return &modifySQLDo{
		do:            t.orm.Do(),
		modifySQLOpts: t.modifySQLOpts,
	}
}

// Txn ...
func (t modifySQLOrm) Txn(tx *sqlx.Tx) DoOrmWithTransaction {
	return &modifySQLDoTxn{
		doTxn:         t.orm.Txn(tx),
		modifySQLOpts: t.modifySQLOpts,
	}
}

// AutoTxn ...
func (t modifySQLOrm) AutoTxn(kt *kit.Kit, run TxnFunc) (interface{}, error) {
	return t.orm.AutoTxn(kt, run)
}

// assemblyInsertSQL 处理"INSERT INTO ... VALUES"格式的SQL，并把指定字段的占位符，加到SQL语句中
func assemblyInsertSQL(expr, fieldName string) (newExpr string, isMatched bool) {
	// 找到字段列表的开始和结束位置
	fieldsStart := strings.Index(expr, "(")
	if fieldsStart == -1 {
		return expr, false
	}

	fieldsEnd := strings.Index(expr[fieldsStart:], ")") + fieldsStart
	if fieldsEnd == -1 {
		return expr, false
	}

	// 找到VALUES关键字
	valuesIndex := strings.Index(strings.ToUpper(expr), "VALUES")
	if valuesIndex == -1 {
		return expr, false
	}

	// 检查是否已包含fieldName字段
	fields := expr[fieldsStart+1 : fieldsEnd]
	if strings.Contains(strings.ToLower(fields), strings.ToLower(fieldName)) {
		return expr, false
	}

	// 添加fieldName到字段列表
	newFields := fields
	if len(fields) > 0 {
		newFields += ", "
	}
	newFields += fieldName

	// 处理VALUES部分
	valuesContent := strings.TrimSpace(expr[valuesIndex+6:])
	// 分割多个VALUES组
	valuesGroups := strings.Split(valuesContent, "),(")

	// 为每个VALUES组添加fieldName参数
	for i := range valuesGroups {
		if i == 0 {
			valuesGroups[i] = strings.TrimPrefix(valuesGroups[i], "(")
		}
		if i == len(valuesGroups)-1 {
			valuesGroups[i] = strings.TrimSuffix(valuesGroups[i], ")")
		}
		// 添加fieldName参数
		valuesGroups[i] = valuesGroups[i] + fmt.Sprintf(", :%s", fieldName)
	}

	// 重新组装VALUES部分
	newValues := "(" + strings.Join(valuesGroups, "),(") + ")"

	// 构建新的SQL语句
	newExpr = expr[:fieldsStart+1] + newFields + expr[fieldsEnd:valuesIndex+6] + newValues
	return newExpr, true
}

// assemblyInsertSelectSQL 处理"INSERT INTO ... SELECT"格式的SQL，并把指定字段的占位符，加到SQL语句中
func assemblyInsertSelectSQL(expr, fieldName string) (newExpr string, selectTableName string, isMatched bool) {
	// 提取字段列表
	leftParen := strings.Index(expr, "(")
	if leftParen == -1 {
		return expr, "", false
	}

	rightParen := strings.Index(expr[leftParen:], ")")
	if rightParen == -1 {
		return expr, "", false
	}

	rightParen += leftParen

	// 检查是否已包含fieldName字段
	fields := expr[leftParen+1 : rightParen]
	if strings.Contains(fields, fieldName) {
		return expr, "", false
	}

	// 添加fieldName到字段列表
	newFields := fields
	if len(fields) > 0 {
		newFields += ", "
	}
	newFields += fieldName

	// 在SELECT部分添加fieldName字段
	selectPos := strings.Index(strings.ToUpper(expr), "SELECT")
	fromPos := strings.Index(strings.ToUpper(expr[selectPos:]), "FROM")
	fromPos += selectPos

	selectFields := expr[selectPos+6 : fromPos]
	newSelectFields := strings.TrimRight(selectFields, " ")
	if len(selectFields) > 0 {
		newSelectFields += ","
	}
	newSelectFields += fmt.Sprintf("%s ", fieldName)

	// 构建新的SQL
	expr = expr[:leftParen+1] + newFields + expr[rightParen:selectPos+6] +
		newSelectFields + expr[fromPos:]

	// 识别SELECT部分的表名
	selectTableMatches := fromTableNameRe.FindAllString(expr[fromPos:], -1)
	if len(selectTableMatches) <= 0 {
		return expr, "", false
	}

	// 在WHERE条件中添加fieldName过滤
	wherePos := strings.Index(strings.ToUpper(expr), "WHERE")
	if wherePos == -1 {
		// 如果没有WHERE条件，添加一个
		expr += fmt.Sprintf(" WHERE %s = :%s", fieldName, fieldName)
	} else {
		// 在现有WHERE条件前添加
		expr = fmt.Sprintf("%sWHERE %s = :%s AND %s", expr[:wherePos], fieldName, fieldName, expr[wherePos+6:])
	}
	return expr, selectTableMatches[0][5:], true
}

// TableAlias 存储表名和别名信息
type TableAlias struct {
	TableName string // 表名
	Alias     string // 别名
}

// parseTableAliases 从 SQL 语句中解析表名和别名
func parseTableAliases(sql string) []TableAlias {
	// 正则表达式匹配 FROM 和 JOIN 后的表名和别名
	matches := fromJoinTableNameRe.FindAllStringSubmatch(sql, -1)

	var result []TableAlias
	for _, match := range matches {
		tableName := match[1]
		alias := match[2]

		tableName = extractTableName(tableName)
		tmpAlias := TableAlias{
			TableName: tableName,
			Alias:     alias,
		}
		// 如果别名为空，使用表名作为别名
		if alias == "" || isSQLKeyword(alias) {
			tmpAlias.Alias = tableName
			result = append(result, tmpAlias)
			continue
		}

		// 去掉别名引号
		if len(alias) >= 2 && alias[0] == '`' && alias[len(alias)-1] == '`' {
			alias = alias[1 : len(alias)-1]
		}

		tmpAlias.Alias = alias
		result = append(result, tmpAlias)
	}
	return result
}

// isSQLKeyword 检查字符串是否是SQL关键字
func isSQLKeyword(s string) bool {
	keywords := []string{"WHERE", "GROUP", "HAVING", "ORDER", "LIMIT", "JOIN", "ON", "AND", "OR"}
	for _, kw := range keywords {
		if strings.EqualFold(s, kw) {
			return true
		}
	}
	return false
}

func appendConditionToExpr(expr string, conditions []string) string {
	if len(conditions) == 0 {
		return expr
	}

	// 处理JOIN查询中的条件
	joinStart := strings.Index(strings.ToUpper(expr), "JOIN")
	if joinStart != -1 && len(conditions) > 1 {
		modifiedJoinQuery := appendConditionToExpr(expr[joinStart+4:], conditions[1:])
		expr = expr[:joinStart+4] + modifiedJoinQuery
		conditions = conditions[:1]
	}

	// 处理主查询的条件
	whereIndex := strings.Index(strings.ToUpper(expr), "WHERE")
	if whereIndex == -1 {
		// 查找其他可能的子句
		groupIndex := strings.Index(strings.ToUpper(expr), "GROUP BY")
		orderIndex := strings.Index(strings.ToUpper(expr), "ORDER BY")
		limitIndex := strings.Index(strings.ToUpper(expr), "LIMIT")

		insertPos := len(expr)
		if groupIndex != -1 {
			insertPos = groupIndex
		} else if orderIndex != -1 {
			insertPos = orderIndex
		} else if limitIndex != -1 {
			insertPos = limitIndex
		}

		// 在合适的位置插入 WHERE 子句
		appendCond := "WHERE " + strings.Join(conditions, " AND ")
		expr = strings.TrimSpace(fmt.Sprintf("%s %s %s", expr[:insertPos], appendCond, expr[insertPos:]))
		return expr
	}

	// 在现有 WHERE 子句后添加条件
	appendCond := strings.Join(conditions, " AND ") + " AND "
	expr = expr[:whereIndex+6] + appendCond + expr[whereIndex+6:]
	return expr
}
