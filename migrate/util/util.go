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

// Package util provides idempotent DDL/DML helpers for hcm-migrate. Migration
// files must only import this package and hcm/migrate/register, never
// hcm/migrate/engine or MetaOrm directly.
package util

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/table"
	"hcm/pkg/logs"
)

// plainIdentifierRe matches a table/column/index/constraint name safe to
// embed into DDL after backtick quoting.
var plainIdentifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// validateIdent checks that name is a safe, plain identifier before it is
// concatenated into a DDL statement built by this package.
func validateIdent(kind, name string) error {
	if !plainIdentifierRe.MatchString(name) {
		return fmt.Errorf("invalid %s name %q, must match %s", kind, name, plainIdentifierRe.String())
	}
	return nil
}

// quoteIdent backtick-quotes an already-validated identifier.
func quoteIdent(name string) string {
	return "`" + name + "`"
}

// escapeLiteral escapes a string so it can be embedded as a single-quoted
// MySQL string literal, e.g. in a DEFAULT or COMMENT clause.
func escapeLiteral(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return replacer.Replace(s)
}

// ColumnDefault is the DEFAULT clause emitted by AddColumn. Pick the
// constructor from the SQL text. Do not collapse both forms into one string.
//
//	SQL has no DEFAULT                 leave Default nil; do not pass an empty string
//	DEFAULT ''                         StringDefault("")
//	DEFAULT 'itsm'                     StringDefault("itsm")
//	DEFAULT NULL                       ExprDefault("NULL")
//	DEFAULT 0 / DEFAULT -1             ExprDefault("0") / ExprDefault("-1")
//	DEFAULT false                      ExprDefault("false")
//	DEFAULT CURRENT_TIMESTAMP          ExprDefault("CURRENT_TIMESTAMP")
//
// StringDefault("NULL") emits the string literal DEFAULT 'NULL', not the NULL keyword.
// Do not add quotes inside ExprDefault: an expr that carries its own single quotes, such as
// an empty quoted string or 'itsm', is wrong.
// Construct values only with StringDefault or ExprDefault, never a ColumnDefault literal.
type ColumnDefault struct {
	quoted bool
	value  string
}

// StringDefault returns a quoted string DEFAULT. v is the text inside the quotes, not a full SQL fragment.
//
//	DEFAULT ''        -> StringDefault("")
//	DEFAULT 'itsm'    -> StringDefault("itsm")
//	DEFAULT 'NULL'    -> StringDefault("NULL") // four characters; rare
//
// Use ExprDefault when the SQL says DEFAULT NULL, a number, false, or CURRENT_TIMESTAMP.
func StringDefault(v string) *ColumnDefault {
	return &ColumnDefault{quoted: true, value: v}
}

// ExprDefault returns an unquoted DEFAULT. expr must match the text after
// DEFAULT in the SQL and must be a constant in the migration.
//
//	DEFAULT NULL                  -> ExprDefault("NULL")
//	DEFAULT 0                     -> ExprDefault("0")
//	DEFAULT -1                    -> ExprDefault("-1")
//	DEFAULT false                 -> ExprDefault("false")
//	DEFAULT CURRENT_TIMESTAMP     -> ExprDefault("CURRENT_TIMESTAMP")
//
// An empty expr is rejected. Use StringDefault for string literals, including an empty-string DEFAULT.
func ExprDefault(expr string) *ColumnDefault {
	return &ColumnDefault{value: expr}
}

// formatDefault renders d as a leading-space DEFAULT clause. A nil d omits
// the clause. A blank unquoted expression is rejected.
func formatDefault(d *ColumnDefault) (string, error) {
	if d == nil {
		return "", nil
	}
	if d.quoted {
		return fmt.Sprintf(" DEFAULT '%s'", escapeLiteral(d.value)), nil
	}
	if strings.TrimSpace(d.value) == "" {
		return "", errors.New("add column: default expression is required")
	}
	return " DEFAULT " + d.value, nil
}

// exec runs a DDL statement built by this package on the caller-provided
// bare orm.Interface. Migration DDL must not go through ModifySQLOpts.
func exec(ctx context.Context, o orm.Interface, ddl string) error {
	if _, err := o.Do().Exec(ctx, ddl); err != nil {
		return fmt.Errorf("exec %q failed, err: %v", ddl, err)
	}
	return nil
}

// CreateTableIfNotExists creates table by running ddl only if table does not
// exist yet. ddl is the caller's full CREATE TABLE statement. created is true
// only when this call ran ddl.
func CreateTableIfNotExists(ctx context.Context, o orm.Interface, table, ddl string) (created bool, err error) {
	if err := validateIdent("table", table); err != nil {
		return false, err
	}
	if strings.TrimSpace(ddl) == "" {
		return false, errors.New("create table: ddl is required")
	}

	exist, err := NewMetaOrm(o).HasTable(ctx, table)
	if err != nil {
		return false, err
	}
	if exist {
		logs.V(1).Infof("table %s already exists, skip create", table)
		return false, nil
	}

	if err := exec(ctx, o, ddl); err != nil {
		return false, err
	}
	return true, nil
}

// DropTable drops table if it exists.
func DropTable(ctx context.Context, o orm.Interface, table string) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}

	exist, err := NewMetaOrm(o).HasTable(ctx, table)
	if err != nil {
		return err
	}
	if !exist {
		logs.V(1).Infof("table %s does not exist, skip drop", table)
		return nil
	}

	ddl := fmt.Sprintf("DROP TABLE %s", quoteIdent(table))
	return exec(ctx, o, ddl)
}

// AddColumnOpt describes the column to add via AddColumn.
type AddColumnOpt struct {
	// Table is the table to alter.
	Table string
	// Column is the column name to add.
	Column string
	// Type is the raw column type, e.g. "int(11)", "varchar(64)". Must be a
	// static literal, never user input.
	Type string
	// NotNull adds a `NOT NULL` constraint when true.
	NotNull bool
	// Default omits the DEFAULT clause when nil. When the SQL has a default, follow ColumnDefault:
	// quoted text uses StringDefault (DEFAULT '' is StringDefault("")),
	// NULL / numbers / false / CURRENT_TIMESTAMP use ExprDefault (DEFAULT NULL is ExprDefault("NULL")).
	Default *ColumnDefault
	// Comment sets a `COMMENT` clause when non-empty.
	Comment string
	// After places the new column after this existing column when non-empty.
	// Leave empty to append the column at the end of the table.
	After string
}

// AddColumn adds the column when it does not exist yet. The statement is built from opt; do not pass a full ALTER.
//
// Fill Default from the SQL text; see ColumnDefault. Leave it nil when the SQL has no DEFAULT.
// varchar with an empty-string DEFAULT uses StringDefault(""). json DEFAULT NULL uses ExprDefault("NULL").
// bigint NOT NULL DEFAULT 0 uses NotNull: true and ExprDefault("0").
func AddColumn(ctx context.Context, o orm.Interface, opt AddColumnOpt) error {
	if err := validateIdent("table", opt.Table); err != nil {
		return err
	}
	if err := validateIdent("column", opt.Column); err != nil {
		return err
	}
	if strings.TrimSpace(opt.Type) == "" {
		return errors.New("add column: type is required")
	}
	if opt.After != "" {
		if err := validateIdent("column", opt.After); err != nil {
			return err
		}
	}
	defaultClause, err := formatDefault(opt.Default)
	if err != nil {
		return err
	}

	exist, err := NewMetaOrm(o).HasColumn(ctx, opt.Table, opt.Column)
	if err != nil {
		return err
	}
	if exist {
		logs.V(1).Infof("column %s.%s already exists, skip add", opt.Table, opt.Column)
		return nil
	}

	ddl := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", quoteIdent(opt.Table), quoteIdent(opt.Column), opt.Type)
	if opt.NotNull {
		ddl += " NOT NULL"
	}
	ddl += defaultClause
	if opt.Comment != "" {
		ddl += fmt.Sprintf(" COMMENT '%s'", escapeLiteral(opt.Comment))
	}
	if opt.After != "" {
		ddl += fmt.Sprintf(" AFTER %s", quoteIdent(opt.After))
	}

	return exec(ctx, o, ddl)
}

// DropColumn drops column from table if it exists.
func DropColumn(ctx context.Context, o orm.Interface, table, column string) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}
	if err := validateIdent("column", column); err != nil {
		return err
	}

	exist, err := NewMetaOrm(o).HasColumn(ctx, table, column)
	if err != nil {
		return err
	}
	if !exist {
		logs.V(1).Infof("column %s.%s does not exist, skip drop", table, column)
		return nil
	}

	ddl := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", quoteIdent(table), quoteIdent(column))
	return exec(ctx, o, ddl)
}

// RenameColumn renames oldColumn to newColumn on table. Skips if newColumn
// already exists. Fails if neither column exists.
func RenameColumn(ctx context.Context, o orm.Interface, table, oldColumn, newColumn string) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}
	if err := validateIdent("column", oldColumn); err != nil {
		return err
	}
	if err := validateIdent("column", newColumn); err != nil {
		return err
	}

	meta := NewMetaOrm(o)

	newExist, err := meta.HasColumn(ctx, table, newColumn)
	if err != nil {
		return err
	}
	if newExist {
		logs.V(1).Infof("column %s.%s already exists, skip rename from %s", table, newColumn, oldColumn)
		return nil
	}

	oldExist, err := meta.HasColumn(ctx, table, oldColumn)
	if err != nil {
		return err
	}
	if !oldExist {
		return fmt.Errorf("rename column %s.%s -> %s failed: neither column exists", table, oldColumn, newColumn)
	}

	ddl := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", quoteIdent(table), quoteIdent(oldColumn),
		quoteIdent(newColumn))
	return exec(ctx, o, ddl)
}

// AddIndex adds a (optionally unique) index on table over columns, using
// index as the index name, if that index does not exist yet.
func AddIndex(ctx context.Context, o orm.Interface, table, index string, columns []string, unique bool) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}
	if err := validateIdent("index", index); err != nil {
		return err
	}
	if len(columns) == 0 {
		return errors.New("add index: columns is required")
	}

	quotedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		if err := validateIdent("column", column); err != nil {
			return err
		}
		quotedColumns = append(quotedColumns, quoteIdent(column))
	}

	exist, err := NewMetaOrm(o).HasIndex(ctx, table, index)
	if err != nil {
		return err
	}
	if exist {
		logs.V(1).Infof("index %s.%s already exists, skip add", table, index)
		return nil
	}

	keyword := "INDEX"
	if unique {
		keyword = "UNIQUE INDEX"
	}
	ddl := fmt.Sprintf("ALTER TABLE %s ADD %s %s (%s)", quoteIdent(table), keyword, quoteIdent(index),
		strings.Join(quotedColumns, ", "))
	return exec(ctx, o, ddl)
}

// DropIndex drops index from table if it exists.
func DropIndex(ctx context.Context, o orm.Interface, table, index string) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}
	if err := validateIdent("index", index); err != nil {
		return err
	}

	exist, err := NewMetaOrm(o).HasIndex(ctx, table, index)
	if err != nil {
		return err
	}
	if !exist {
		logs.V(1).Infof("index %s.%s does not exist, skip drop", table, index)
		return nil
	}

	ddl := fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", quoteIdent(table), quoteIdent(index))
	return exec(ctx, o, ddl)
}

// AddConstraint runs ddl to add constraint on table only if it does not
// exist yet. ddl is the caller's full ALTER TABLE ... ADD CONSTRAINT ...
// statement.
func AddConstraint(ctx context.Context, o orm.Interface, table, constraint, ddl string) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}
	if err := validateIdent("constraint", constraint); err != nil {
		return err
	}
	if strings.TrimSpace(ddl) == "" {
		return errors.New("add constraint: ddl is required")
	}

	exist, err := NewMetaOrm(o).HasConstraint(ctx, table, constraint)
	if err != nil {
		return err
	}
	if exist {
		logs.V(1).Infof("constraint %s.%s already exists, skip add", table, constraint)
		return nil
	}

	return exec(ctx, o, ddl)
}

// DropConstraint runs ddl to drop constraint from table only if it still
// exists. ddl is the caller's full ALTER TABLE ... DROP ... statement.
func DropConstraint(ctx context.Context, o orm.Interface, table, constraint, ddl string) error {
	if err := validateIdent("table", table); err != nil {
		return err
	}
	if err := validateIdent("constraint", constraint); err != nil {
		return err
	}
	if strings.TrimSpace(ddl) == "" {
		return errors.New("drop constraint: ddl is required")
	}

	exist, err := NewMetaOrm(o).HasConstraint(ctx, table, constraint)
	if err != nil {
		return err
	}
	if !exist {
		logs.V(1).Infof("constraint %s.%s does not exist, skip drop", table, constraint)
		return nil
	}

	return exec(ctx, o, ddl)
}

// idGeneratorInsertExpr inserts a resource row, swallowing a duplicate
// `resource` primary key with a no-op update.
const idGeneratorInsertExpr = "INSERT INTO `" + string(table.IDGenerator) + "` (`resource`, `max_id`) " +
	"VALUES (:resource, :max_id) ON DUPLICATE KEY UPDATE `resource` = `resource`"

// InsertIDGenerator inserts one id_generator seed row (resource, maxID). A
// duplicate resource key is a no-op.
func InsertIDGenerator(ctx context.Context, o orm.Interface, resource, maxID string) error {
	if resource == "" || maxID == "" {
		return errors.New("insert id_generator: resource and max_id are required")
	}

	arg := map[string]interface{}{"resource": resource, "max_id": maxID}
	if err := o.Do().Insert(ctx, idGeneratorInsertExpr, arg); err != nil {
		return fmt.Errorf("insert id_generator for resource %s failed, err: %v", resource, err)
	}
	return nil
}
