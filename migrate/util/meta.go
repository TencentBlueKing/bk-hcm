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
)

// MetaOrm probes MySQL structure via information_schema. It is a tool used
// internally by this package's idempotent Add*/Drop*/CreateTableIfNotExists
// helpers, not a public API: migration files must call those helpers instead
// of calling MetaOrm directly.
type MetaOrm interface {
	// HasTable reports whether table exists in the current database.
	HasTable(ctx context.Context, table string) (bool, error)
	// HasView reports whether view exists in the current database.
	HasView(ctx context.Context, view string) (bool, error)
	// HasColumn reports whether column exists on table.
	HasColumn(ctx context.Context, table, column string) (bool, error)
	// HasIndex reports whether index exists on table.
	HasIndex(ctx context.Context, table, index string) (bool, error)
	// HasConstraint reports whether constraint (e.g. foreign key) exists on table.
	HasConstraint(ctx context.Context, table, constraint string) (bool, error)
	// ListTables returns all base table names in the current database.
	ListTables(ctx context.Context) ([]string, error)
}

// NewMetaOrm creates a MetaOrm backed by the given orm.Interface. Callers must
// pass in a bare orm (e.g. dal.GetOrm().Do() root, without ModifySQLOpts), so
// structure probing is never affected by tenant SQL rewriting.
func NewMetaOrm(o orm.Interface) MetaOrm {
	return &metaOrm{orm: o}
}

type metaOrm struct {
	orm orm.Interface
}

// countExists runs a "select count(*) from information_schema.xxx where ..."
// style expr and reports whether the count is greater than zero.
func (m *metaOrm) countExists(ctx context.Context, expr string, arg map[string]interface{}) (bool, error) {
	count, err := m.orm.Do().Count(ctx, expr, arg)
	if err != nil {
		return false, fmt.Errorf("query information_schema failed, err: %v", err)
	}
	return count > 0, nil
}

const hasTableExpr = "SELECT COUNT(*) FROM information_schema.tables " +
	"WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' AND table_name = :table"

// HasTable reports whether table exists in the current database.
func (m *metaOrm) HasTable(ctx context.Context, table string) (bool, error) {
	exist, err := m.countExists(ctx, hasTableExpr, map[string]interface{}{"table": table})
	if err != nil {
		return false, fmt.Errorf("has table %s failed, err: %v", table, err)
	}
	return exist, nil
}

const hasViewExpr = "SELECT COUNT(*) FROM information_schema.views " +
	"WHERE table_schema = DATABASE() AND table_name = :view"

// HasView reports whether view exists in the current database.
func (m *metaOrm) HasView(ctx context.Context, view string) (bool, error) {
	exist, err := m.countExists(ctx, hasViewExpr, map[string]interface{}{"view": view})
	if err != nil {
		return false, fmt.Errorf("has view %s failed, err: %v", view, err)
	}
	return exist, nil
}

const hasColumnExpr = "SELECT COUNT(*) FROM information_schema.columns " +
	"WHERE table_schema = DATABASE() AND table_name = :table AND column_name = :column"

// HasColumn reports whether column exists on table.
func (m *metaOrm) HasColumn(ctx context.Context, table, column string) (bool, error) {
	arg := map[string]interface{}{"table": table, "column": column}
	exist, err := m.countExists(ctx, hasColumnExpr, arg)
	if err != nil {
		return false, fmt.Errorf("has column %s.%s failed, err: %v", table, column, err)
	}
	return exist, nil
}

const hasIndexExpr = "SELECT COUNT(*) FROM information_schema.statistics " +
	"WHERE table_schema = DATABASE() AND table_name = :table AND index_name = :index"

// HasIndex reports whether index exists on table.
func (m *metaOrm) HasIndex(ctx context.Context, table, index string) (bool, error) {
	arg := map[string]interface{}{"table": table, "index": index}
	exist, err := m.countExists(ctx, hasIndexExpr, arg)
	if err != nil {
		return false, fmt.Errorf("has index %s.%s failed, err: %v", table, index, err)
	}
	return exist, nil
}

const hasConstraintExpr = "SELECT COUNT(*) FROM information_schema.table_constraints " +
	"WHERE table_schema = DATABASE() AND table_name = :table AND constraint_name = :constraint"

// HasConstraint reports whether constraint exists on table.
func (m *metaOrm) HasConstraint(ctx context.Context, table, constraint string) (bool, error) {
	arg := map[string]interface{}{"table": table, "constraint": constraint}
	exist, err := m.countExists(ctx, hasConstraintExpr, arg)
	if err != nil {
		return false, fmt.Errorf("has constraint %s.%s failed, err: %v", table, constraint, err)
	}
	return exist, nil
}

const listTablesExpr = "SELECT table_name FROM information_schema.tables " +
	"WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'"

// ListTables returns all base table names in the current database.
func (m *metaOrm) ListTables(ctx context.Context) ([]string, error) {
	var rows []struct {
		Name string `db:"table_name"`
	}
	if err := m.orm.Do().Select(ctx, &rows, listTablesExpr, map[string]interface{}{}); err != nil {
		return nil, fmt.Errorf("list tables failed, err: %v", err)
	}

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names, nil
}
