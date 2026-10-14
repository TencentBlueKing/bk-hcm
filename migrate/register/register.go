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

// Package register holds the in-process migration registries and the version
// grammar. Migration files must not import hcm/migrate/engine.
package register

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"slices"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/migrate"
)

// UpFunc applies one migration. The orm has no tenant SQL rewriting.
// Migration bodies should build statements with hcm/migrate/util.
type UpFunc func(ctx context.Context, o orm.Interface) error

// Migrator is the interface a migration package passes to Regist. Regist
// reads the import path of the type implementing it.
type Migrator interface {
	Up(ctx context.Context, o orm.Interface) error
}

// Migration is one registered migration file.
type Migration struct {
	// ID is the migration ID. It is immutable and is the key used to decide
	// whether a migration is already applied.
	ID string
	// Version is the raw version string, either a valid version or the
	// constant.MigrationPendingVersion placeholder.
	Version string
	// Timestamp is the 14-digit creation stamp, used to order migrations
	// that share a version.
	Timestamp string
	// Pkg is the import path of the migration package without
	// constant.MigrationPkgPrefix, e.g. main/v1.9.3/v1.9.3_20260905160000_add_bk_asset_id.
	// It takes no part in ordering.
	Pkg string
	// Up applies the migration.
	Up UpFunc

	// parsed is Version after parsing, zero when Version is the placeholder.
	parsed Version
}

// IsPending reports whether the migration is still registered under the
// undecided-version placeholder.
func (m Migration) IsPending() bool {
	return IsPending(m.Version)
}

// ParsedVersion returns the parsed version. The second result is false when
// Version is the constant.MigrationPendingVersion placeholder.
func (m Migration) ParsedVersion() (Version, bool) {
	if m.IsPending() {
		return Version{}, false
	}
	return m.parsed, true
}

// Registry holds every migration registered for one database.
type Registry struct {
	database string
	list     []Migration
}

// Main is the registry of the main database.
var Main = &Registry{database: constant.MigrationDatabaseMain}

// Database returns the database this registry belongs to. It is the name
// --database selects registries by.
func (r *Registry) Database() string {
	return r.database
}

// Regist registers one migration. The package path of m must match this
// registry's database. Invalid arguments panic. Duplicate IDs are allowed;
// see All.
func (r *Registry) Regist(id, version, timestamp string, m Migrator) {
	pkgPath, up, err := migratorOf(m)
	if err != nil {
		panic(fmt.Sprintf("regist migration into %s registry failed, err: %v, id: %q", r.database, err, id))
	}

	r.regist(pkgPath, id, version, timestamp, up)
}

// regist registers a migration whose package path is already known.
func (r *Registry) regist(pkgPath, id, version, timestamp string, up UpFunc) {
	m, err := NewMigration(r.database, pkgPath, id, version, timestamp, up)
	if err != nil {
		panic(fmt.Sprintf("regist migration into %s registry failed, err: %v", r.database, err))
	}

	r.list = append(r.list, m)
}

// migratorOf returns the import path of m's type and its Up method. A nil m
// or a nil pointer is rejected.
func migratorOf(m Migrator) (pkgPath string, up UpFunc, err error) {
	if m == nil {
		return "", nil, fmt.Errorf("migrator is nil")
	}

	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Pointer {
		if reflect.ValueOf(m).IsNil() {
			return "", nil, fmt.Errorf("migrator is a nil %s", t)
		}
		t = t.Elem()
	}

	return t.PkgPath(), m.Up, nil
}

// NewMigration validates the arguments and builds a Migration without
// registering it. Errors name every given field.
func NewMigration(database, pkgPath, id, version, timestamp string, up UpFunc) (Migration, error) {
	where := fmt.Sprintf("id: %q, version: %q, timestamp: %q, package: %q", id, version, timestamp, pkgPath)

	if err := migrate.ValidateMigrationID(id); err != nil {
		return Migration{}, fmt.Errorf("%v, %s", err, where)
	}

	if err := migrate.ValidateTimestamp(timestamp); err != nil {
		return Migration{}, fmt.Errorf("%v, %s", err, where)
	}

	if up == nil {
		return Migration{}, fmt.Errorf("up function is nil, %s", where)
	}

	if len(version) > constant.MigrationVersionMaxLen {
		return Migration{}, fmt.Errorf("invalid version, longer than %d characters, %s",
			constant.MigrationVersionMaxLen, where)
	}

	pkg, err := migrate.ParsePkgPath(database, pkgPath, timestamp)
	if err != nil {
		return Migration{}, fmt.Errorf("%v, %s", err, where)
	}

	m := Migration{
		ID:        id,
		Version:   version,
		Timestamp: timestamp,
		Pkg:       pkg,
		Up:        up,
	}

	// The placeholder has no parsed version.
	if IsPending(version) {
		return m, nil
	}

	parsed, err := Parse(version)
	if err != nil {
		return Migration{}, fmt.Errorf("%v, %s", err, where)
	}
	m.parsed = parsed

	return m, nil
}

// All returns the registered migrations in execution order: version, then
// timestamp, then migration ID. The package path takes no part. Migrations
// on the constant.MigrationPendingVersion placeholder sort last. Duplicate
// IDs are all included. The returned slice is a fresh copy.
func (r *Registry) All() []Migration {
	out := slices.Clone(r.list)

	slices.SortStableFunc(out, compareMigration)

	return out
}

// compareMigration returns -1, 0 or +1 ordering two migrations by version,
// then timestamp, then ID.
func compareMigration(a, b Migration) int {
	av, aOK := a.ParsedVersion()
	bv, bOK := b.ParsedVersion()

	switch {
	case !aOK && !bOK:
		// Both undecided, fall through to timestamp and ID.
	case !aOK:
		return 1
	case !bOK:
		return -1
	default:
		if c := Compare(av, bv); c != 0 {
			return c
		}
	}

	return cmp.Or(cmp.Compare(a.Timestamp, b.Timestamp), cmp.Compare(a.ID, b.ID))
}
