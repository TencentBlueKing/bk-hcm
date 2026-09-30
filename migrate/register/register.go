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

package register

import (
	"cmp"
	"context"
	"fmt"
	"regexp"
	"slices"
	"time"

	"hcm/pkg/dal/dao/orm"
)

// UpFunc applies one migration. It receives the orm of the database the
// migration was registered into, with no tenant SQL rewriting attached, so
// statements hit exactly the table they name. Migration bodies should build
// their statements with hcm/migrate/util so that re-running them is safe.
type UpFunc func(ctx context.Context, o orm.Interface) error

// migrationIDRe matches a canonical lowercase UUID. Uppercase and the
// urn/brace spellings are rejected on purpose: the executor skips an already
// applied migration by comparing this string to the one in the record table,
// so one migration must have exactly one spelling. MySQL compares the stored
// value case-insensitively while Go does not, and that gap is what turns a
// re-spelled ID into either a duplicate-key failure or a silent re-run.
var migrationIDRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// timestampRe matches the fixed 14-digit creation stamp (yyyyMMddHHmmss).
// Fixed width is what lets the executor order same-version migrations by
// plain string comparison.
var timestampRe = regexp.MustCompile(`^\d{14}$`)

// MaxVersionLen is the most characters a version string may have. It is also
// the width of the record table's version column: a longer value would be
// rejected or silently truncated on insert, and the truncated text would
// then fail to parse. Versions are ASCII, so this byte length is the
// character length MySQL counts.
const MaxVersionLen = 64

// Migration is one registered migration file.
type Migration struct {
	// ID is the canonical lowercase UUID minted when the file was created.
	// It never changes, not even when the file is released or archived, and
	// it is the only key the executor uses to decide "already applied".
	ID string
	// Version is the raw version string, either a valid version or the
	// PendingVersion placeholder.
	Version string
	// Timestamp is the 14-digit creation stamp, used to order migrations
	// that share a version.
	Timestamp string
	// Description is the semantic name of the change, for logs and list
	// output only. It takes no part in ordering.
	Description string
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

// ParsedVersion returns the parsed version. The second result is false for a
// migration still on the PendingVersion placeholder, which has no version to
// compare; callers that order or filter by version must handle that case
// rather than treating the zero Version as a real one.
func (m Migration) ParsedVersion() (Version, bool) {
	if m.IsPending() {
		return Version{}, false
	}
	return m.parsed, true
}

// Registry holds every migration registered for one database. Registration
// happens from the init() of the migration files, which Go runs one at a
// time, so no locking is needed here.
type Registry struct {
	database string
	list     []Migration
}

// Main is the registry of the main database, fed by migrations/main.
// Obs is the registry of the OBS database, fed by migrations/obs.
var (
	Main = &Registry{database: "main"}
	Obs  = &Registry{database: "obs"}
)

// Database returns the database this registry belongs to. It is the name
// --database selects registries by.
func (r *Registry) Database() string {
	return r.database
}

// Regist registers one migration file. Which database a migration belongs to
// is decided here, by the registry it is registered into, never by parsing
// its path at runtime.
//
// Every argument is validated now, at init() time, and anything invalid
// panics before the process reaches main: a migration that cannot be ordered
// or recorded must not be allowed to run at all. Registering the same ID
// twice is explicitly allowed, see the package documentation of All.
func (r *Registry) Regist(id, version, timestamp, description string, up UpFunc) {
	m, err := newMigration(id, version, timestamp, description, up)
	if err != nil {
		panic(fmt.Sprintf("regist migration into %s registry failed, err: %v", r.database, err))
	}

	r.list = append(r.list, m)
}

// newMigration validates the arguments and builds the migration. Errors name
// every field that was given, because the panic message is all the reader
// gets to find the offending file with.
func newMigration(id, version, timestamp, description string, up UpFunc) (Migration, error) {
	where := fmt.Sprintf("id: %q, version: %q, timestamp: %q, description: %q",
		id, version, timestamp, description)

	if !migrationIDRe.MatchString(id) {
		return Migration{}, fmt.Errorf("invalid migration id, want a lowercase uuid, %s", where)
	}

	if !timestampRe.MatchString(timestamp) {
		return Migration{}, fmt.Errorf("invalid timestamp, want 14 digits, %s", where)
	}
	// 14 digits is not enough: 20261399999999 matches the pattern and would
	// still sort, but it is not a time anyone created a file at.
	if _, err := time.Parse("20060102150405", timestamp); err != nil {
		return Migration{}, fmt.Errorf("invalid timestamp, want yyyyMMddHHmmss, %s", where)
	}

	if up == nil {
		return Migration{}, fmt.Errorf("up function is nil, %s", where)
	}

	if len(version) > MaxVersionLen {
		return Migration{}, fmt.Errorf("invalid version, longer than %d characters, %s", MaxVersionLen, where)
	}

	m := Migration{
		ID:          id,
		Version:     version,
		Timestamp:   timestamp,
		Description: description,
		Up:          up,
	}

	// The placeholder is not parsable. It is accepted here and rejected
	// later by the pre-execution scan of init and up, so that an undecided
	// migration can still be listed but never executed.
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
// timestamp, then migration ID. Path and description take no part in it.
// Migrations still on the PendingVersion placeholder have no version to
// compare and are put last, after every versioned one.
//
// Two migrations may share an ID. That is the normal result of merging the
// same change from another line, where one copy carries a three-segment
// version and the other a fourth segment. Both stay registered and both
// appear here; the executor runs whichever comes first and skips the other
// once the ID is recorded as applied.
//
// The returned slice is a fresh copy, so the caller may reorder or trim it
// without touching the registry.
func (r *Registry) All() []Migration {
	out := slices.Clone(r.list)
	// Stable, so that migrations identical on all three sort keys keep
	// their registration order and repeated calls return the same sequence.
	slices.SortStableFunc(out, compareMigration)

	return out
}

// compareMigration returns -1, 0 or +1 ordering two migrations. It extends
// Compare with the two keys that live on the migration rather than on the
// version: the timestamp and then the ID, both fixed-format strings.
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
