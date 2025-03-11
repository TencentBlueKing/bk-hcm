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
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"golang.org/x/time/rate"

	"github.com/jmoiron/sqlx"
	prm "github.com/prometheus/client_golang/prometheus"
)

// DoOrm defines all the orm method.
type DoOrm interface {
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
	// TableSharding at least one TableSharding option
	TableSharding(opts ...TableShardingOpt) Interface
	// InjectTenant inject tenant
	InjectTenant(kt *kit.Kit) Interface
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
	kt             *kit.Kit
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

// InjectTenant ...
func (o *runtimeOrm) InjectTenant(kt *kit.Kit) Interface {
	return &tableShardingOrm{
		orm: o,
		kt:  kt,
	}
}

// tableShardingOrm orm for table sharding, it replaces table name in sql query
type tableShardingOrm struct {
	orm               *runtimeOrm
	tableShardingOpts []TableShardingOpt
	kt                *kit.Kit
}

type tableShardingDo struct {
	do                *do
	tableShardingOpts []TableShardingOpt
	kt                *kit.Kit
}

type tableShardingDoTxn struct {
	doTxn             DoOrmWithTransaction
	tableShardingOpts []TableShardingOpt
	kt                *kit.Kit
}

// Count ...
func (dt *tableShardingDoTxn) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	replaced := replaceFromJoinTableName(dt.tableShardingOpts, expr)

	// 租户开关开启
	if dt.kt.TenantEnable && dt.kt.TenantID != "" && dt.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectJoinTenantID(replaced, arg, dt.kt.TenantID)
	}

	return dt.doTxn.Count(ctx, replaced, arg)
}

// Select ...
func (dt *tableShardingDoTxn) Select(ctx context.Context, dest interface{}, expr string,
	arg map[string]interface{}) error {

	replaced := replaceFromJoinTableName(dt.tableShardingOpts, expr)

	// 租户开关开启
	if dt.kt.TenantEnable && dt.kt.TenantID != "" && dt.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectJoinTenantID(replaced, arg, dt.kt.TenantID)
	}

	return dt.doTxn.Select(ctx, dest, replaced, arg)
}

// Delete ...
func (dt *tableShardingDoTxn) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceFromJoinTableName(dt.tableShardingOpts, expr)

	// 租户开关开启
	if dt.kt.TenantEnable && dt.kt.TenantID != "" && dt.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectJoinTenantID(replaced, arg, dt.kt.TenantID)
	}

	return dt.doTxn.Delete(ctx, replaced, arg)
}

// Update ...
func (dt *tableShardingDoTxn) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceUpdateTableName(dt.tableShardingOpts, expr)

	// 租户开关开启
	if dt.kt.TenantEnable && dt.kt.TenantID != "" && dt.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectUpdateTenantID(replaced, arg, dt.kt.TenantID)
	}

	return dt.doTxn.Update(ctx, replaced, arg)
}

// Insert ...
func (dt *tableShardingDoTxn) Insert(ctx context.Context, expr string, data interface{}) error {
	replaced := replaceInsertTableName(dt.tableShardingOpts, expr)

	// 租户开关开启
	if dt.kt.TenantEnable && dt.kt.TenantID != "" && dt.kt.TenantID != constant.DefaultTenantID {
		replaced, data = injectInsertTenantID(replaced, data, dt.kt.TenantID)
	}

	return dt.doTxn.BulkInsert(ctx, replaced, data)
}

// BulkInsert ...
func (dt *tableShardingDoTxn) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	replaced := replaceInsertTableName(dt.tableShardingOpts, expr)

	// 租户开关开启
	if dt.kt.TenantEnable && dt.kt.TenantID != "" && dt.kt.TenantID != constant.DefaultTenantID {
		replaced, args = injectInsertTenantID(replaced, args, dt.kt.TenantID)
	}

	return dt.doTxn.BulkInsert(ctx, replaced, args)
}

// Count ...
func (ds *tableShardingDo) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)

	// 租户开关开启
	if ds.kt.TenantEnable && ds.kt.TenantID != "" && ds.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectJoinTenantID(replaced, arg, ds.kt.TenantID)
	}

	return ds.do.Count(ctx, replaced, arg)
}

// Delete ...
func (ds *tableShardingDo) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)

	// 租户开关开启
	if ds.kt.TenantEnable && ds.kt.TenantID != "" && ds.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectJoinTenantID(replaced, arg, ds.kt.TenantID)
	}

	return ds.do.Delete(ctx, replaced, arg)
}

// Update ...
func (ds *tableShardingDo) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	replaced := replaceUpdateTableName(ds.tableShardingOpts, expr)

	// 租户开关开启
	if ds.kt.TenantEnable && ds.kt.TenantID != "" && ds.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectUpdateTenantID(replaced, arg, ds.kt.TenantID)
	}

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

	// 租户开关开启
	if ds.kt.TenantEnable && ds.kt.TenantID != "" && ds.kt.TenantID != constant.DefaultTenantID {
		replaced, data = injectInsertTenantID(replaced, data, ds.kt.TenantID)
	}

	return ds.do.BulkInsert(ctx, replaced, data)
}

// BulkInsert ...
func (ds *tableShardingDo) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	replaced := replaceInsertTableName(ds.tableShardingOpts, expr)

	// 租户开关开启
	if ds.kt.TenantEnable && ds.kt.TenantID != "" && ds.kt.TenantID != constant.DefaultTenantID {
		replaced, args = injectInsertTenantID(replaced, args, ds.kt.TenantID)
	}

	return ds.do.BulkInsert(ctx, replaced, args)
}

// Select ...
func (ds *tableShardingDo) Select(ctx context.Context, dest interface{}, expr string,
	arg map[string]interface{}) error {

	replaced := replaceFromJoinTableName(ds.tableShardingOpts, expr)

	// 租户开关开启
	if ds.kt.TenantEnable && ds.kt.TenantID != "" && ds.kt.TenantID != constant.DefaultTenantID {
		replaced, arg = injectJoinTenantID(replaced, arg, ds.kt.TenantID)
	}

	return ds.do.Select(ctx, dest, replaced, arg)
}

// Do ...
func (t tableShardingOrm) Do() DoOrm {

	return &tableShardingDo{
		do: &do{
			db: t.orm.db,
			ro: t.orm,
		},
		tableShardingOpts: t.tableShardingOpts,
		kt:                t.kt,
	}
}

// Txn ...
func (t tableShardingOrm) Txn(tx *sqlx.Tx) DoOrmWithTransaction {
	return &tableShardingDoTxn{
		doTxn:             t.orm.Txn(tx),
		tableShardingOpts: t.tableShardingOpts,
		kt:                t.kt,
	}
}

// AutoTxn ...
func (t tableShardingOrm) AutoTxn(kt *kit.Kit, run TxnFunc) (interface{}, error) {
	return t.orm.AutoTxn(kt, run)
}

// TableSharding ...
func (t tableShardingOrm) TableSharding(opts ...TableShardingOpt) Interface {

	t.tableShardingOpts = append(t.tableShardingOpts, opts...)

	return t
}

// InjectTenant ...
func (t tableShardingOrm) InjectTenant(kt *kit.Kit) Interface {
	t.kt = kt
	return t
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
	if replaced[0] == '`' && replaced[len(replaced)-1] == '`' {
		replaced = replaced[1 : len(replaced)-1]
	}

	return replaced
}

// regular expression for extracting table name
var (
	fromTableNameRe   = regexp.MustCompile("(?i)\\sfrom\\s+(`?\\w+`?\\.)?`?\\w+`?")
	joinTableNameRe   = regexp.MustCompile("(?i)\\sjoin\\s+(`?\\w+`?\\.)?`?\\w+`?")
	updateTableNameRe = regexp.MustCompile("(?i)update\\s+(`?\\w+`?\\.)?`?\\w+`?")
	insertTableNameRe = regexp.MustCompile("(?i)insert\\s+into\\s+(`?\\w+`?\\.)?`?\\w+`?")
)

// injectInsertTenantID Insert使用
func injectInsertTenantID(expr string, args interface{}, tenantID string) (string, interface{}) {
	// 如果不是INSERT语句，直接返回
	if !insertTableNameRe.MatchString(expr) {
		return expr, args
	}

	// 使用正则表达式匹配表名
	tableMatch := insertTableNameRe.FindString(expr)
	if tableMatch != "" {
		// 去掉 "INSERT INTO " 前缀
		tableMatch = tableMatch[12:] // "INSERT INTO " 长度为 12
		// 去掉可能的反引号
		tableName := extractTableName(tableMatch)
		// 该表不支持多租户
		isTenant, ok := table.TableMap[table.Name(tableName)]
		logs.Infof("injectInsertTenantID: 表名: %s, isTenant: %v, ok: %v", tableName, isTenant, ok)
		if !ok || !isTenant {
			return expr, args
		}
	}

	// 找到第一个左括号（字段列表开始）
	firstLeftParen := strings.Index(expr, "(")
	if firstLeftParen == -1 {
		return expr, args
	}

	// 找到字段列表的右括号
	fieldsStr := expr[firstLeftParen+1:]
	fieldsEnd := strings.Index(fieldsStr, ")")
	if fieldsEnd == -1 {
		return expr, args
	}

	// 提取字段列表
	fields := fieldsStr[:fieldsEnd]

	// 找到VALUES关键字
	remainder := fieldsStr[fieldsEnd+1:]
	valuesIdx := strings.Index(strings.ToUpper(remainder), "VALUES")
	if valuesIdx == -1 {
		return expr, args
	}

	// 找到VALUES后的左括号(6 是 "VALUES" 的长度)
	valuesStr := remainder[valuesIdx+6:]
	valuesStart := strings.Index(valuesStr, "(")
	if valuesStart == -1 {
		return expr, args
	}

	// 找到匹配的右括号（考虑嵌套括号）
	valueContent := valuesStr[valuesStart+1:]
	valuesEnd := findMatchingParenthesis(valueContent)
	if valuesEnd == -1 {
		return expr, args
	}

	// 构建新的字段列表
	newFields := fields
	if !strings.Contains(strings.ToLower(fields), constant.TenantIDField) {
		if len(fields) > 0 {
			newFields += ", "
		}
		newFields += constant.TenantIDField
	}

	// 构建新的值列表
	values := valueContent[:valuesEnd]
	if !strings.Contains(strings.ToLower(values), constant.TenantIDField) {
		if len(values) > 0 {
			values += ", "
		}
		values += fmt.Sprintf(":%s", constant.TenantIDField)
	}

	// 组装最终的SQL
	newExpr := fmt.Sprintf("%s(%s) VALUES(%s)",
		expr[:firstLeftParen],
		newFields,
		values,
	)

	logs.Infof("injectInsertTenantID:expr, tenantID: %s, expr: %s", tenantID, expr)
	logs.Infof("injectInsertTenantID:newExpr, tenantID: %s, newFields: %s, newExpr: %s", tenantID, newFields, newExpr)

	// 处理参数
	args, skip := parseArgsAndSetTenant(args, tenantID)
	if skip {
		return newExpr, args
	}

	argsJson, _ := json.Marshal(args)
	logs.Infof("injectInsertTenantID:end:669, tenantID: %s, argsJson: %s", tenantID, argsJson)

	return newExpr, args
}

func findMatchingParenthesis(s string) int {
	count := 1
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			count++
		case ')':
			count--
			if count == 0 {
				return i
			}
		}
	}
	return -1
}

// toCamelCase 蛇形命名转驼峰
func toCamelCase(s string) string {
	slice := strings.Split(s, "_")
	for i := 0; i < len(slice); i++ {
		if i == len(slice)-1 {
			slice[i] = strings.ToUpper(slice[i])
		} else {
			slice[i] = strings.Title(slice[i])
		}
	}
	return strings.Join(slice, "")
}

// parseTableName 从 SQL 片段中提取表名或别名
func parseTableName(sqlPart string) string {
	parts := strings.Fields(sqlPart)
	var tableName string

	// 查找最后一个单词，它可能是表名或别名
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		// 跳过 AS 关键字
		if strings.ToUpper(part) == "AS" {
			continue
		}
		// 移除可能的反引号
		part = strings.Trim(part, "`")
		// 如果包含点号，取最后一部分
		if strings.Contains(part, ".") {
			parts = strings.Split(part, ".")
			tableName = parts[len(parts)-1]
		} else {
			tableName = part
		}
		break
	}

	return tableName
}

func parseArgsAndSetTenant(args interface{}, tenantID string) (interface{}, bool) {
	rv := reflect.ValueOf(args)
	if rv.Kind() != reflect.Slice {
		return args, true
	}

	// 处理切片类型
	skip := false
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() != reflect.Ptr || elem.IsNil() {
			skip = true
			continue
		}
		// 获取指针指向的结构体
		structValue := elem.Elem()
		if structValue.Kind() != reflect.Struct {
			skip = true
			continue
		}
		// 查找并设置 TenantID 字段
		fieldTenantID := toCamelCase(constant.TenantIDField)
		field := structValue.FieldByName(fieldTenantID)
		if field.IsValid() && field.CanSet() {
			field.SetString(tenantID)
		}
	}
	return args, skip
}

// injectJoinTenantID Delete、List、Count使用
func injectJoinTenantID(expr string, arg map[string]interface{}, tenantID string) (string, map[string]interface{}) {
	if arg == nil {
		arg = make(map[string]interface{})
	}

	// 获取主表名
	fromMatches := fromTableNameRe.FindAllString(expr, -1)
	// 获取 JOIN 表名
	joinMatches := joinTableNameRe.FindAllString(expr, -1)

	// 构建 tenant_id 条件
	conditions := make([]string, 0)

	// 处理主表条件
	if len(fromMatches) > 0 {
		// 提取主表名或别名
		mainTable := parseTableName(fromMatches[0])
		// 该表不支持多租户
		isTenant, ok := table.TableMap[table.Name(mainTable)]
		logs.Infof("injectJoinTenantID: 表名: %s, isTenant: %v, ok: %v, expr: %s", mainTable, isTenant, ok, expr)
		if !ok || !isTenant {
			return expr, arg
		}
		conditions = append(conditions, fmt.Sprintf("%s.%s = :%s", mainTable,
			constant.TenantIDField, constant.TenantIDField))
	}

	// 处理 JOIN 表条件
	for _, joinMatch := range joinMatches {
		joinTable := parseTableName(joinMatch)
		conditions = append(conditions, fmt.Sprintf("%s.%s = :%s", joinTable,
			constant.TenantIDField, constant.TenantIDField))
	}

	// 注入条件到 WHERE 子句
	whereIndex := strings.Index(strings.ToUpper(expr), "WHERE")
	if whereIndex == -1 {
		// 查找其他可能的子句
		orderIndex := strings.Index(strings.ToUpper(expr), "ORDER BY")
		limitIndex := strings.Index(strings.ToUpper(expr), "LIMIT")
		groupIndex := strings.Index(strings.ToUpper(expr), "GROUP BY")

		insertPos := len(expr)
		if orderIndex != -1 {
			insertPos = orderIndex
		} else if groupIndex != -1 {
			insertPos = groupIndex
		} else if limitIndex != -1 {
			insertPos = limitIndex
		}

		// 在合适的位置插入 WHERE 子句
		tenantCond := " WHERE " + strings.Join(conditions, " AND ")
		expr = expr[:insertPos] + tenantCond + expr[insertPos:]
	} else {
		// 在现有 WHERE 子句后添加条件
		tenantCond := strings.Join(conditions, " AND ") + " AND "
		expr = expr[:whereIndex+6] + tenantCond + expr[whereIndex+6:]
	}

	// 添加参数
	arg[constant.TenantIDField] = tenantID

	logs.Infof("injectJoinTenantID:end, tenantID: %s, conditions: %v, ====expr: %s, ====arg: %+v",
		tenantID, conditions, expr, arg)

	return expr, arg
}

// injectUpdateTenantID Update使用
func injectUpdateTenantID(expr string, arg map[string]interface{}, tenantID string) (string, map[string]interface{}) {
	if arg == nil {
		arg = make(map[string]interface{})
	}

	// 使用正则表达式匹配表名
	tableMatch := updateTableNameRe.FindStringSubmatch(expr)
	if len(tableMatch) == 0 {
		return expr, arg
	}

	// 提取表名或别名
	tableName := extractTableName(tableMatch[0][6:]) // 去掉 "UPDATE " 前缀

	// 该表不支持多租户
	isTenant, ok := table.TableMap[table.Name(tableName)]
	logs.Infof("injectUpdateTenantID: 表名: %s, isTenant: %v, ok: %v", tableName, isTenant, ok)
	if !ok || !isTenant {
		return expr, arg
	}

	// 查找 WHERE 子句位置
	whereIndex := strings.Index(strings.ToUpper(expr), "WHERE")
	if whereIndex == -1 {
		// 如果没有 WHERE 子句，添加一个
		expr += " WHERE " + fmt.Sprintf("%s.%s = :%s", tableName, constant.TenantIDField, constant.TenantIDField)
	} else {
		// 在现有 WHERE 子句后添加 AND
		expr = expr[:whereIndex+6] + fmt.Sprintf("%s.%s = :%s AND ", tableName,
			constant.TenantIDField, constant.TenantIDField) + expr[whereIndex+6:]
	}

	// 添加 tenant_id 参数
	arg[constant.TenantIDField] = tenantID

	logs.Infof("injectUpdateTenantID:end, tenantID: %s, tableName: %s, ====expr: %s, ====arg: %+v",
		tenantID, tableName, expr, arg)

	return expr, arg
}
