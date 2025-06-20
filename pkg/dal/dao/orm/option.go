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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/table"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

type options struct {
	// ingressLimiter db request limiter.
	ingressLimiter *rate.Limiter
	// logLimiter write db request log limiter.
	logLimiter *rate.Limiter
	// mc db request metrics.
	mc *metric
	// slowRequestMS db slow request time, beyond this time, the db request will be logged. unit: millisecond
	slowRequestMS time.Duration
}

// Option orm option func defines.
type Option func(opt *options)

// IngressLimiter set db request limiter related params.
func IngressLimiter(qps, burst uint) Option {
	return func(opt *options) {
		opt.ingressLimiter = rate.NewLimiter(rate.Limit(qps), int(burst))
	}
}

// LogLimiter write db request log limiter related params.
func LogLimiter(qps, burst uint) Option {
	return func(opt *options) {
		opt.logLimiter = rate.NewLimiter(rate.Limit(qps), int(burst))
	}
}

// MetricsRegisterer set metrics registerer.
func MetricsRegisterer(register prometheus.Registerer) Option {
	return func(opt *options) {
		opt.mc = initMetric(register)
	}
}

// SlowRequestMS set db slow request time.
func SlowRequestMS(ms uint) Option {
	return func(opt *options) {
		opt.slowRequestMS = time.Duration(ms) * time.Millisecond
	}
}

// TableShardingOpt defines table name generation options.
type TableShardingOpt interface {
	// Match check if table name match this sharding option
	Match(name string) bool
	// ReplaceTableName try to replace table name, only matched table name should be replaced
	ReplaceTableName(old string) string
}

// TableSuffixShardingOpt  append suffix to table name
type TableSuffixShardingOpt struct {
	tableName string
	suffixes  []string
}

// NewTableSuffixShardingOpt ...
func NewTableSuffixShardingOpt(tableName string, suffixes []string) *TableSuffixShardingOpt {
	return &TableSuffixShardingOpt{tableName: tableName, suffixes: suffixes}
}

// Match given table name
func (r *TableSuffixShardingOpt) Match(name string) bool {
	return strings.Compare(r.tableName, name) == 0
}

// ReplaceTableName append suffix to original table name
func (r *TableSuffixShardingOpt) ReplaceTableName(old string) string {
	replaced := old
	for _, s := range r.suffixes {
		replaced += "_" + s
	}
	return replaced
}

// String ...
func (r *TableSuffixShardingOpt) String() string {
	return fmt.Sprintf("{tableName:%s, suffixes: %v}", r.tableName, r.suffixes)
}

// ModifySQLOpt defines modify sql options.
type ModifySQLOpt interface {
	// InjectInsertSQL inject insert sql.
	InjectInsertSQL(expr string, args interface{}) (string, any)
	// InjectUpdateSQL inject update sql.
	InjectUpdateSQL(expr string, arg map[string]interface{}) (string, map[string]interface{})
	// InjectDeleteSQL inject delete sql.
	InjectDeleteSQL(expr string, arg map[string]interface{}) (string, map[string]interface{})
	// InjectJoinSQL inject join sql.
	InjectJoinSQL(expr string, arg map[string]interface{}) (string, map[string]interface{})
}

// InjectTenantIDOpt  inject tenant id option
type InjectTenantIDOpt struct {
	enabledTenant bool
	tenantID      string
}

// NewInjectTenantIDOpt ...
// InjectTenantIDOpt.InjectInsertSQL
// 支持的case示例：
//  1. 支持：普通的INSERT INTO ... VALUES形式
//  2. 支持：INSERT INTO ... SELECT（
//
// SELECT的table需要支持多租户，否则不会自动添加租户ID的字段及过滤条件；不支持查询指定租户ID，插入当前租户的需求）
//
// InjectTenantIDOpt.InjectUpdateSQL
// 支持的case示例：
//  1. 支持：普通的UPDATE ... SET ... WHERE ...形式（table需要支持多租户，否则不会自动添加租户ID的过滤条件）
//
// InjectTenantIDOpt.InjectDeleteSQL
// 支持的case示例：
//  1. 支持：DELETE FROM ... WHERE ...形式（table需要支持多租户，否则不会自动添加租户ID的过滤条件）
//
// InjectTenantIDOpt.InjectJoinSQL
// 支持的case示例：（table需要支持多租户，否则不会自动添加租户ID的过滤条件）
//  1. 简单SELECT查询-表名带引号
//     INPUT:  SELECT * FROM `cvm`
//     OUTPUT: SELECT * FROM `cvm` WHERE cvm.tenant_id = :tenant_id
//  2. 简单SELECT查询-DB+表名带引号
//     INPUT:  SELECT * FROM db.`cvm`
//     OUTPUT: SELECT * FROM db.`cvm` WHERE cvm.tenant_id = :tenant_id
//  3. 带JOIN的SELECT查询-带AS
//     INPUT:  SELECT * FROM network_interface AS ni JOIN network_interface_cvm_rel AS rel
//     ON ni.id = rel.network_interface_id WHERE rel.id=1
//     OUTPUT: SELECT * FROM network_interface AS ni JOIN network_interface_cvm_rel AS rel
//     ON ni.id = rel.network_interface_id WHERE ni.tenant_id = :tenant_id AND rel.id=1
//  4. 带子查询的SELECT-带WHERE
//     INPUT： SELECT * FROM (SELECT * FROM cvm WHERE id > :id) AS sub WHERE id=1 and name="test"
//     OUTPUT：SELECT * FROM (SELECT * FROM cvm WHERE cvm.tenant_id = :tenant_id AND id > :id) AS sub
//     WHERE id=1 AND name="test"
//
// 不支持的case示例：
//  1. 只要是支持多租户的表，就会增加租户ID的过滤条件，不支持只给指定表添加租户ID条件的需求
func NewInjectTenantIDOpt(tenantID string) *InjectTenantIDOpt {
	return &InjectTenantIDOpt{tenantID: tenantID, enabledTenant: cc.TenantEnable()}
}

// enabled check is enabled.
func (ito *InjectTenantIDOpt) enabled() bool {
	return ito.enabledTenant && ito.tenantID != "" && ito.tenantID != constant.DefaultTenantID
}

// InjectInsertSQL Insert使用
// 支持的case：
//  1. 普通的INSERT INTO ... VALUES形式
//     INPUT： INSERT INTO db1.`cvm` (name) VALUES (:name)
//     OUTPUT：INSERT INTO db1.`cvm` (name, tenant_id) VALUES (:name, :tenant_id)
//  2. INSERT INTO ... SELECT（SELECT的table需要支持多租户，否则无法注入tenant_id）
//     INPUT： INSERT INTO db1.`cvm` (id,name) SELECT id,name FROM disk WHERE id=1
//     OUTPUT：INSERT INTO db1.`cvm` (id,name, tenant_id) SELECT id,name,tenant_id FROM disk
//     WHERE tenant_id = :tenant_id AND id=1
func (ito *InjectTenantIDOpt) InjectInsertSQL(expr string, args interface{}) (string, any) {
	// 没开启多租户，直接返回
	if !ito.enabled() {
		return expr, args
	}

	// 如果不是INSERT语句，直接返回
	if !insertTableNameRe.MatchString(expr) {
		return expr, args
	}

	// 使用正则表达式匹配表名
	tableMatch := insertTableNameRe.FindString(expr)
	if tableMatch == "" {
		return expr, args
	}

	// 去掉 "INSERT INTO " 前缀，"INSERT INTO " 长度为 12
	tableMatch = tableMatch[12:]
	// 提取表名
	tableName := extractTableName(tableMatch)

	// 该表不支持多租户
	tableConfig, ok := table.TableMap[table.Name(tableName)]
	logs.V(4).Infof("injectInsertTenantID:start, tableName: %s, tableConfig: %+v, ok: %v", tableName, tableConfig, ok)
	if !ok || !tableConfig.EnableTenant {
		return expr, args
	}

	// 处理INSERT INTO ... SELECT形式
	if strings.Contains(strings.ToUpper(expr), "SELECT") {
		newExpr, selectTblName, isMatched := assemblyInsertSelectSQL(expr, constant.TenantIDField)
		if !isMatched {
			return expr, args
		}

		// 检查SELECT表是否支持多租户
		selectTableName := extractTableName(selectTblName)
		selectTableConfig, ok := table.TableMap[table.Name(selectTableName)]
		if !ok || !selectTableConfig.EnableTenant {
			return expr, args
		}
		expr = newExpr
	} else {
		// 处理普通INSERT INTO ... VALUES形式
		newExpr, isMatched := assemblyInsertSQL(expr, constant.TenantIDField)
		if !isMatched {
			return expr, args
		}
		expr = newExpr
	}

	// 处理参数
	args = ito.parseArgsAndSetTenant(args)

	if logs.V(4) {
		argsJson, err := json.Marshal(args)
		if err != nil {
			logs.Warnf("injectInsertTenantID:end, tenantID: %s, json marshal(args) failed, err: %v, args: %v",
				ito.tenantID, err, args)
		}
		logs.Infof("injectInsertTenantID:end, tenantID: %s, newExpr: %s, argsJson: %s", ito.tenantID, expr, argsJson)
	}

	return expr, args
}

// InjectUpdateSQL Update使用
func (ito *InjectTenantIDOpt) InjectUpdateSQL(expr string, arg map[string]interface{}) (
	string, map[string]interface{}) {

	// 没开启多租户，直接返回
	if !ito.enabled() {
		return expr, arg
	}

	if arg == nil {
		arg = make(map[string]interface{})
	}

	// 使用正则表达式匹配表名
	tableMatch := updateTableNameRe.FindStringSubmatch(expr)
	if len(tableMatch) == 0 {
		return expr, arg
	}

	// 提取表名
	tableName := extractTableName(tableMatch[0][6:])
	// 该表不支持多租户
	tableConfig, ok := table.TableMap[table.Name(tableName)]
	logs.V(4).Infof("injectUpdateTenantID, tableName: %s, tableConfig: %+v, ok: %v", tableName, tableConfig, ok)
	if !ok || !tableConfig.EnableTenant {
		return expr, arg
	}

	// 构建 tenant_id 条件
	conditions := make([]string, 0)
	tenantCondition := fmt.Sprintf("%s = :%s", constant.TenantIDField, constant.TenantIDField)
	if !strings.Contains(expr, tenantCondition) {
		conditions = append(conditions, tenantCondition)
	}
	expr = appendConditionToExpr(expr, conditions)

	// 添加 tenant_id 参数
	arg[constant.TenantIDField] = ito.tenantID

	logs.V(4).Infof("injectUpdateTenantID:end, tenantID: %s, tableName: %s, expr: %s, arg: %+v",
		ito.tenantID, tableName, expr, arg)

	return expr, arg
}

// InjectDeleteSQL Delete使用
func (ito *InjectTenantIDOpt) InjectDeleteSQL(expr string, arg map[string]interface{}) (
	string, map[string]interface{}) {

	// 没开启多租户，直接返回
	if !ito.enabled() {
		return expr, arg
	}

	if arg == nil {
		arg = make(map[string]interface{})
	}

	// 获取表名
	fromMatches := fromTableNameRe.FindAllString(expr, -1)
	if len(fromMatches) == 0 {
		return expr, arg
	}

	// 提取表名
	tableName := extractTableName(fromMatches[0][5:])

	// 该表不支持多租户
	tableConfig, ok := table.TableMap[table.Name(tableName)]
	logs.V(4).Infof("injectDeleteTenantID:tableName: %s, tableConfig: %+v, ok: %v", tableName, tableConfig, ok)
	if !ok || !tableConfig.EnableTenant {
		return expr, arg
	}

	// 构建 tenant_id 条件
	conditions := make([]string, 0)
	tenantCondition := fmt.Sprintf("%s = :%s", constant.TenantIDField, constant.TenantIDField)
	if !strings.Contains(expr, tenantCondition) {
		conditions = append(conditions, tenantCondition)
	}
	expr = appendConditionToExpr(expr, conditions)

	// 添加参数
	arg[constant.TenantIDField] = ito.tenantID

	logs.V(4).Infof("injectDeleteTenantID:end, tenantID: %s, conditions: %v, expr: %s, arg: %+v",
		ito.tenantID, conditions, expr, arg)

	return expr, arg
}

// InjectJoinSQL List、Count使用
// 支持的case示例：
//  1. 简单SELECT查询-表名带引号
//     INPUT:  SELECT * FROM `cvm`
//     OUTPUT: SELECT * FROM `cvm` WHERE cvm.tenant_id = :tenant_id
//  2. 简单SELECT查询-DB+表名带引号
//     INPUT:  SELECT * FROM db.`cvm`
//     OUTPUT: SELECT * FROM db.`cvm` WHERE cvm.tenant_id = :tenant_id
//  3. 带JOIN的SELECT查询-带AS
//     INPUT:  SELECT * FROM network_interface AS ni JOIN network_interface_cvm_rel AS rel
//     ON ni.id = rel.network_interface_id WHERE rel.id=1
//     OUTPUT: SELECT * FROM network_interface AS ni JOIN network_interface_cvm_rel AS rel
//     ON ni.id = rel.network_interface_id WHERE ni.tenant_id = :tenant_id AND rel.id=1
//  4. 带子查询的SELECT-带WHERE
//     INPUT： SELECT * FROM (SELECT * FROM cvm WHERE id > :id) AS sub WHERE id=1 and name="test"
//     OUTPUT：SELECT * FROM (SELECT * FROM cvm WHERE cvm.tenant_id = :tenant_id AND id > :id) AS sub
//     WHERE id=1 AND name="test"
//
// 不支持的case示例：
//  1. 只要是支持多租户的表，就会增加租户ID的过滤条件，不支持只给指定表添加租户ID条件的需求
func (ito *InjectTenantIDOpt) InjectJoinSQL(expr string, arg map[string]interface{}) (string, map[string]interface{}) {
	// 没开启多租户，直接返回
	if !ito.enabled() {
		return expr, arg
	}

	if arg == nil {
		arg = make(map[string]interface{})
	}

	// 提取表名及别名
	matchTables := parseTableAliases(expr)
	if len(matchTables) == 0 {
		return expr, arg
	}

	// 构建 tenant_id 条件
	conditions := make([]string, 0)
	// 该表是否支持多租户
	tableConfig, ok := table.TableMap[table.Name(matchTables[0].TableName)]
	logs.V(4).Infof("injectJoinTenantID:mainTable, tableConfig: %+v, ok: %v, mainTable: %+v",
		tableConfig, ok, matchTables[0])
	if ok && tableConfig.EnableTenant {
		tenantCondition := fmt.Sprintf("%s.%s = :%s", matchTables[0].Alias, constant.TenantIDField,
			constant.TenantIDField)
		if !strings.Contains(expr, tenantCondition) {
			conditions = append(conditions, tenantCondition)
		}
	}

	// 处理 JOIN 表条件
	if len(matchTables) > 1 {
		for i := 1; i < len(matchTables); i++ {
			joinTable := matchTables[i]
			// 该表不支持多租户
			tableConfig, ok = table.TableMap[table.Name(joinTable.TableName)]
			logs.V(4).Infof("injectJoinTenantID:joinTable, tableConfig: %+v, ok: %v, joinTable: %+v",
				tableConfig, ok, joinTable)
			if !ok || !tableConfig.EnableTenant {
				continue
			}
			tenantJoinCondition := fmt.Sprintf("%s.%s = :%s", joinTable.Alias, constant.TenantIDField,
				constant.TenantIDField)
			if !strings.Contains(expr, tenantJoinCondition) {
				conditions = append(conditions, tenantJoinCondition)
			}
		}
	}

	// 注入条件到 WHERE 子句
	expr = appendConditionToExpr(expr, conditions)

	// 添加参数
	arg[constant.TenantIDField] = ito.tenantID

	logs.V(4).Infof("injectJoinTenantID:end, tenantID: %s, conditions: %v, expr: %s, arg: %+v",
		ito.tenantID, conditions, expr, arg)

	return expr, arg
}

// setTenantID 通过反射设置结构体的tenantID字段
func (ito *InjectTenantIDOpt) setTenantID(structValue reflect.Value) {
	// 查找并设置 TenantID 字段
	field := structValue.FieldByName(constant.TenantIDTableField)
	if field.IsValid() && field.CanSet() {
		field.SetString(ito.tenantID)
	}
}

func (ito *InjectTenantIDOpt) parseArgsAndSetTenant(args interface{}) interface{} {
	rv := reflect.ValueOf(args)
	switch rv.Kind() {
	case reflect.Ptr:
		// 获取指针指向的值
		elem := rv.Elem()
		// 如果是结构体，设置tenantID
		if elem.Kind() == reflect.Struct {
			ito.setTenantID(elem)
		}
		return args
	case reflect.Map:
		if m, ok := args.(map[string]interface{}); ok {
			m[constant.TenantIDField] = ito.tenantID
		}
		return args
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			// 如果元素是指针且不为nil
			if elem.Kind() == reflect.Ptr && !elem.IsNil() {
				// 获取指针指向的结构体
				structValue := elem.Elem()
				if structValue.Kind() == reflect.Struct {
					ito.setTenantID(structValue)
				}
			} else if elem.Kind() == reflect.Struct {
				// 直接处理结构体类型
				ito.setTenantID(elem)
			}
		}
		return args
	default:
		return args
	}
}
