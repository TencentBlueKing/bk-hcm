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
	"database/sql"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"unicode/utf8"

	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	// import mysql driver, used to create conn.
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Local MySQL tests use the same server as pkg/dal/dao/id-generator
// (127.0.0.1:3306) but never open the application database.
// This Homebrew server accepts root with an empty password; root/admin is only a fallback.
// Each test creates hcm_migrate_engine_test_<pid>_<n> and drops only that name.

const (
	localMySQLHost = "127.0.0.1"
	localMySQLPort = "3306"
)

// localMySQLCredentials tries the empty root password first. The id-generator
// fixture uses root/admin, which this server rejects.
var localMySQLCredentials = []struct {
	user string
	pass string
}{
	{user: "root", pass: ""},
	{user: "root", pass: "admin"},
}

// testDatabaseRe is the only schema this file is allowed to create or drop.
var testDatabaseRe = regexp.MustCompile(`^hcm_migrate_engine_test_[0-9]+_[0-9]+$`)

var testDBSeq uint64

func TestValidateTestDatabase(t *testing.T) {
	testCases := []struct {
		name    string
		dbName  string
		wantErr bool
	}{
		{name: "generated name", dbName: "hcm_migrate_engine_test_1_1"},
		{name: "app database", dbName: "hcm", wantErr: true},
		{name: "app database upper", dbName: "HCM", wantErr: true},
		{name: "mysql system", dbName: "mysql", wantErr: true},
		{name: "information_schema", dbName: "information_schema", wantErr: true},
		{name: "performance_schema", dbName: "performance_schema", wantErr: true},
		{name: "sys", dbName: "sys", wantErr: true},
		{name: "empty", dbName: "", wantErr: true},
		{name: "prefix only", dbName: "hcm_migrate_engine_test_", wantErr: true},
		{name: "util package prefix", dbName: "hcm_migrate_util_test_1_1", wantErr: true},
		{name: "injection", dbName: "hcm_migrate_engine_test_1_1`; DROP DATABASE hcm; --", wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTestDatabase(tc.dbName)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestLocalMySQLTargetIsLoopback(t *testing.T) {
	addr, err := localMySQLAddr()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:3306", addr)
}

func TestLocalMySQLEngine(t *testing.T) {
	kt := kit.New()
	db, o, dbName, _, _ := openIsolatedMySQL(t)
	_ = dbName

	t.Run("fresh init creates both tables", func(t *testing.T) {
		result, err := schema.InitTables(kt, o, nil, nil)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.True(t, tableExists(t, db, constant.MigrationAuditTable))
		assert.True(t, tableExists(t, db, constant.MigrationRecordTable))
		assert.Equal(t, 0, countRecords(t, db))

		records, err := schema.NewRecordStore(o).Load(kt)
		require.NoError(t, err)
		assert.Empty(t, records)
		current, ok, err := schema.CurrentVersion(records)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, register.Version{}, current)
	})

	t.Run("second init is a no-op even with a baseline", func(t *testing.T) {
		baseline := mustVersion(t, "v1.9.10")
		migrations := []register.Migration{
			mustMigration(t, migA, "v1.9.3", "20260101120000", "would be adopted"),
		}
		result, err := schema.InitTables(kt, o, &baseline, migrations)
		require.NoError(t, err)
		assert.False(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Equal(t, 0, countRecords(t, db))
	})

	baseline := mustVersion(t, "v1.9.3")
	// migA v1.9.3 is listed before the same ID at v1.9.2, so the kept row is v1.9.3.
	migrations := []register.Migration{
		mustMigration(t, migA, "v1.9.3", "20260201120000", "first"),
		mustMigration(t, migA, "v1.9.2", "20260101120000", "duplicate"),
		mustMigration(t, migB, "v1.9.3.0", "20260101120000", "folded"),
		mustMigration(t, migC, "v1.9.3.1", "20260101120000", "fourth"),
		mustMigration(t, migD, constant.MigrationPendingVersion, "20260101120000", "pending"),
		mustMigration(t, migE, "v1.9.4", "20260101120000", "later"),
	}

	t.Run("init adopt", func(t *testing.T) {
		dropRecordTable(t, db)
		result, err := schema.InitTables(kt, o, &baseline, migrations)
		require.NoError(t, err)
		assert.False(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		adopted := make([]string, len(result.Adopted))
		for i, m := range result.Adopted {
			adopted[i] = m.ID
		}
		require.Equal(t, []string{migA, migB}, adopted)

		rows := loadRows(t, db)
		require.Len(t, rows, 2)
		assert.Equal(t, "v1.9.3", rows[migA].Version)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migA].Status)
		assert.Equal(t, "", rows[migA].Message)
		assert.Equal(t, result.Adopted[0].Pkg, rows[migA].AppliedPkg)
		assert.Equal(t, "v1.9.3.0", rows[migB].Version)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migB].Status)
		assert.Equal(t, result.Adopted[1].Pkg, rows[migB].AppliedPkg)
		_, hasFourth := rows[migC]
		assert.False(t, hasFourth)
		_, hasPending := rows[migD]
		assert.False(t, hasPending)
		_, hasLater := rows[migE]
		assert.False(t, hasLater)

		records, err := schema.NewRecordStore(o).Load(kt)
		require.NoError(t, err)
		current, ok, err := schema.CurrentVersion(records)
		require.NoError(t, err)
		require.True(t, ok)
		// v1.9.3 and v1.9.3.0 compare equal; the smaller raw text wins.
		assert.Equal(t, "v1.9.3", current.Raw)
	})

	t.Run("second adopt does not raise the baseline", func(t *testing.T) {
		higher := mustVersion(t, "v1.9.10")
		before := countRecords(t, db)
		result, err := schema.InitTables(kt, o, &higher, migrations)
		require.NoError(t, err)
		assert.False(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Equal(t, before, countRecords(t, db))
		rows := loadRows(t, db)
		_, hasLater := rows[migE]
		assert.False(t, hasLater)
	})

	t.Run("drop audit only then rerun recreates audit leaves records", func(t *testing.T) {
		beforeRows := loadRows(t, db)
		require.NotEmpty(t, beforeRows)
		dropAuditTable(t, db)
		assert.False(t, tableExists(t, db, constant.MigrationAuditTable))
		assert.True(t, tableExists(t, db, constant.MigrationRecordTable))

		result, err := schema.InitTables(kt, o, &baseline, migrations)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.True(t, tableExists(t, db, constant.MigrationAuditTable))
		assert.Equal(t, beforeRows, loadRows(t, db))
	})

	t.Run("mark running failed and success", func(t *testing.T) {
		store := schema.NewRecordStore(o)
		running := register.Migration{
			ID: migF, Version: "v1.9.3", Pkg: "main/pending/20260101120000_mig_f",
		}
		require.NoError(t, store.MarkRunning(kt, running))
		rows := loadRows(t, db)
		require.Contains(t, rows, migF)
		assert.Equal(t, enumor.MigrationStatusRunning, rows[migF].Status)
		assert.Equal(t, "v1.9.3", rows[migF].Version)
		assert.Equal(t, running.Pkg, rows[migF].AppliedPkg)
		assert.Equal(t, "", rows[migF].Message)
		assert.Equal(t, 1, countByID(t, db, migF))

		raw := strings.Repeat("迁移失败，磁盘已满。", 80)
		require.Greater(t, len(raw), constant.MigrationRecordMessageMaxBytes)
		require.NoError(t, store.MarkFailed(kt, running, fmt.Errorf("%s", raw)))
		rows = loadRows(t, db)
		msg := rows[migF].Message
		assert.Equal(t, enumor.MigrationStatusFailed, rows[migF].Status)
		assert.Equal(t, running.Pkg, rows[migF].AppliedPkg)
		assert.LessOrEqual(t, len(msg), constant.MigrationRecordMessageMaxBytes)
		assert.True(t, utf8.ValidString(msg))
		assert.True(t, strings.HasPrefix(raw, msg))
		assert.NotEqual(t, raw, msg)
		assert.Equal(t, migrate.TruncateUTF8(raw, constant.MigrationRecordMessageMaxBytes), msg)

		refreshed := register.Migration{
			ID: migF, Version: "v1.9.4", Pkg: "main/pending/20260101120000_mig_f_v2",
		}
		require.NoError(t, store.MarkRunning(kt, refreshed))
		rows = loadRows(t, db)
		assert.Equal(t, enumor.MigrationStatusRunning, rows[migF].Status)
		assert.Equal(t, "", rows[migF].Message)
		assert.Equal(t, refreshed.Pkg, rows[migF].AppliedPkg)
		// ON DUPLICATE KEY UPDATE clears the message and leaves the version.
		assert.Equal(t, "v1.9.3", rows[migF].Version)
		assert.Equal(t, 1, countByID(t, db, migF))

		require.NoError(t, store.MarkSuccess(kt, refreshed))
		rows = loadRows(t, db)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migF].Status)
		assert.Equal(t, "v1.9.4", rows[migF].Version)
		assert.Equal(t, refreshed.Pkg, rows[migF].AppliedPkg)
		assert.Equal(t, "", rows[migF].Message)

		require.NoError(t, store.MarkRunning(kt, register.Migration{
			ID: migC, Version: "v1.9.9", Pkg: "main/pending/20260101120000_mig_c",
		}))
		require.NoError(t, store.MarkSuccess(kt, register.Migration{
			ID: migC, Version: "v1.9.9", Pkg: "main/pending/20260101120000_mig_c",
		}))
		require.NoError(t, store.MarkRunning(kt, register.Migration{
			ID: migD, Version: "v1.9.10", Pkg: "main/pending/20260101120000_mig_d",
		}))
		require.NoError(t, store.MarkSuccess(kt, register.Migration{
			ID: migD, Version: "v1.9.10", Pkg: "main/pending/20260101120000_mig_d",
		}))
		require.NoError(t, store.MarkRunning(kt, register.Migration{
			ID: migE, Version: "v9.0.0", Pkg: "main/pending/20260101120000_mig_e",
		}))

		records, err := store.Load(kt)
		require.NoError(t, err)
		current, ok, err := schema.CurrentVersion(records)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "v1.9.10", current.Raw)
	})

	t.Run("rerun with different pkg overwrites applied_pkg", func(t *testing.T) {
		store := schema.NewRecordStore(o)
		first := register.Migration{
			ID: "20260101-1200-Z-0001", Version: "v1.9.3",
			Pkg: "main/pending/20260101120000_first_pkg",
		}
		second := register.Migration{
			ID: first.ID, Version: "v1.9.3",
			Pkg: "main/pending/20260101120000_second_pkg",
		}
		require.NoError(t, store.MarkRunning(kt, first))
		require.NoError(t, store.MarkSuccess(kt, first))
		assert.Equal(t, first.Pkg, loadRows(t, db)[first.ID].AppliedPkg)

		require.NoError(t, store.MarkRunning(kt, second))
		assert.Equal(t, second.Pkg, loadRows(t, db)[first.ID].AppliedPkg)
		require.NoError(t, store.MarkSuccess(kt, second))
		assert.Equal(t, second.Pkg, loadRows(t, db)[first.ID].AppliedPkg)
	})

	t.Run("255-char applied_pkg stored whole", func(t *testing.T) {
		store := schema.NewRecordStore(o)
		pkg := "main/g/" + "20260101120000_" + strings.Repeat("a", 233)
		require.Equal(t, constant.MigrationPkgMaxLen, len(pkg))
		m := register.Migration{
			ID: "20260101-1200-Y-0001", Version: "v1.9.3", Pkg: pkg,
		}
		require.NoError(t, store.MarkRunning(kt, m))
		require.NoError(t, store.MarkSuccess(kt, m))
		assert.Equal(t, pkg, loadRows(t, db)[m.ID].AppliedPkg)
	})

	t.Run("load rejects a raw running row with an unparsable version", func(t *testing.T) {
		dropRecordTable(t, db)
		result, err := schema.InitTables(kt, o, nil, nil)
		require.NoError(t, err)
		assert.False(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)

		_, err = db.Exec("INSERT INTO `hcm_migration_record` (`migration_id`, `version`, `status`) VALUES (?, ?, ?)",
			migA, "garbage", string(enumor.MigrationStatusRunning))
		require.NoError(t, err)

		records, err := schema.NewRecordStore(o).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), migA)
		assert.Contains(t, err.Error(), "garbage")
		assert.Contains(t, err.Error(), "unparsable version")
	})
}

func TestLocalMySQLPendingBackfill(t *testing.T) {
	kt := kit.New()
	db, o, _, _, _ := openIsolatedMySQL(t)

	result, err := schema.InitTables(kt, o, nil, nil)
	require.NoError(t, err)
	assert.True(t, result.AuditCreated)
	assert.True(t, result.RecordCreated)

	store := schema.NewRecordStore(o)
	pending := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "pending_backfill")
	released := mustMigration(t, migB, "v1.9.2", "20260101120000", "released")

	require.NoError(t, store.MarkRunning(kt, pending))
	require.NoError(t, store.MarkSuccess(kt, pending))
	rows := loadRows(t, db)
	require.Contains(t, rows, migA)
	assert.Equal(t, constant.MigrationPendingVersion, rows[migA].Version)
	assert.Equal(t, enumor.MigrationStatusSuccess, rows[migA].Status)
	assert.Equal(t, pending.Pkg, rows[migA].AppliedPkg)

	require.NoError(t, store.MarkRunning(kt, released))
	require.NoError(t, store.MarkSuccess(kt, released))

	records, err := store.Load(kt)
	require.NoError(t, err)
	require.Contains(t, records, migA)
	assert.Equal(t, constant.MigrationPendingVersion, records[migA].Version)
	current, ok, err := schema.CurrentVersion(records)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "v1.9.2", current.Raw)

	backfill := register.Migration{ID: migA, Version: "v1.9.3", Pkg: pending.Pkg}
	require.NoError(t, store.BackfillPendingVersion(kt, backfill))
	rows = loadRows(t, db)
	assert.Equal(t, "v1.9.3", rows[migA].Version)
	assert.Equal(t, pending.Pkg, rows[migA].AppliedPkg)
	assert.Equal(t, enumor.MigrationStatusSuccess, rows[migA].Status)
	assert.Equal(t, 1, countByID(t, db, migA))

	// Conditional update is idempotent: a second call (even with a newer
	// registered version) leaves the already-backfilled row alone.
	again := register.Migration{ID: migA, Version: "v1.9.4", Pkg: pending.Pkg}
	require.NoError(t, store.BackfillPendingVersion(kt, again))
	assert.Equal(t, "v1.9.3", loadRows(t, db)[migA].Version)

	failedPending := mustMigration(t, migC, constant.MigrationPendingVersion, "20260102120000", "failed_pending")
	require.NoError(t, store.MarkRunning(kt, failedPending))
	require.NoError(t, store.MarkFailed(kt, failedPending, fmt.Errorf("boom")))
	assert.Equal(t, constant.MigrationPendingVersion, loadRows(t, db)[migC].Version)
	assert.Equal(t, enumor.MigrationStatusFailed, loadRows(t, db)[migC].Status)

	require.NoError(t, store.BackfillPendingVersion(kt, register.Migration{
		ID: migC, Version: "v1.9.3", Pkg: failedPending.Pkg,
	}))
	assert.Equal(t, constant.MigrationPendingVersion, loadRows(t, db)[migC].Version)
	assert.Equal(t, enumor.MigrationStatusFailed, loadRows(t, db)[migC].Status)
}

func TestLocalMySQLAudit(t *testing.T) {
	kt := kit.New()
	db, o, _, _, _ := openIsolatedMySQL(t)

	result, err := schema.InitTables(kt, o, nil, nil)
	require.NoError(t, err)
	assert.True(t, result.AuditCreated)
	assert.True(t, result.RecordCreated)

	t.Run("full cycle begin and end", func(t *testing.T) {
		runKT := kit.New()
		runKT.Rid = "audit-run-1"
		a := schema.NewAudit(runKT, o).Begin(runKT, schema.AuditMeta{Command: "up", Args: []string{"--db=main"}})
		require.NotNil(t, a)

		var before struct {
			Status   string         `db:"status"`
			ExitCode sql.NullInt64  `db:"exit_code"`
			EndAt    sql.NullTime   `db:"end_at"`
			Skipped  sql.NullString `db:"skipped"`
		}
		require.NoError(t, db.Get(&before,
			"SELECT `status`, `exit_code`, `end_at`, `skipped` FROM `hcm_migration_audit` WHERE `run_id` = ?",
			runKT.Rid))
		assert.Equal(t, string(enumor.MigrationStatusRunning), before.Status)
		assert.False(t, before.ExitCode.Valid)
		assert.False(t, before.EndAt.Valid)

		a.End(runKT, schema.AuditResult{
			ExitCode: 0,
			Skipped: []schema.SkippedMigration{{
				ID: migA, Version: "v1.9.3", Pkg: "main/pending/20260101120000_a",
				Reason: enumor.MigrationSkipBaseline,
			}},
			Warnings: []string{"note"},
			Messages: []string{},
		})

		var after struct {
			Status   string        `db:"status"`
			ExitCode sql.NullInt64 `db:"exit_code"`
			EndAt    sql.NullTime  `db:"end_at"`
			Skipped  string        `db:"skipped"`
			Warnings string        `db:"warnings"`
			Message  string        `db:"message"`
		}
		require.NoError(t, db.Get(&after,
			"SELECT `status`, `exit_code`, `end_at`, `skipped`, `warnings`, `message` "+
				"FROM `hcm_migration_audit` WHERE `run_id` = ?", runKT.Rid))
		assert.Equal(t, string(enumor.MigrationStatusSuccess), after.Status)
		require.True(t, after.ExitCode.Valid)
		assert.Equal(t, int64(0), after.ExitCode.Int64)
		assert.True(t, after.EndAt.Valid)

		var skippedOK, warningsOK, messageOK int
		require.NoError(t, db.Get(&skippedOK, "SELECT JSON_VALID(?)", after.Skipped))
		require.NoError(t, db.Get(&warningsOK, "SELECT JSON_VALID(?)", after.Warnings))
		require.NoError(t, db.Get(&messageOK, "SELECT JSON_VALID(?)", after.Message))
		assert.Equal(t, 1, skippedOK)
		assert.Equal(t, 1, warningsOK)
		assert.Equal(t, 1, messageOK)
	})

	t.Run("second Begin with same rid returns nil", func(t *testing.T) {
		runKT := kit.New()
		runKT.Rid = "audit-run-dup"
		a1 := schema.NewAudit(runKT, o).Begin(runKT, schema.AuditMeta{Command: "up"})
		require.NotNil(t, a1)
		a2 := schema.NewAudit(runKT, o).Begin(runKT, schema.AuditMeta{Command: "up"})
		assert.Nil(t, a2)
	})

	t.Run("unended run appears in next warnings", func(t *testing.T) {
		first := kit.New()
		first.Rid = "audit-run-orphan"
		a1 := schema.NewAudit(first, o).Begin(first, schema.AuditMeta{Command: "up"})
		require.NotNil(t, a1)

		second := kit.New()
		second.Rid = "audit-run-next"
		a2 := schema.NewAudit(second, o).Begin(second, schema.AuditMeta{Command: "up"})
		require.NotNil(t, a2)
		a2.End(second, schema.AuditResult{ExitCode: 0})

		var warnings string
		require.NoError(t, db.Get(&warnings,
			"SELECT `warnings` FROM `hcm_migration_audit` WHERE `run_id` = ?", second.Rid))
		assert.Contains(t, warnings, first.Rid)
	})
}

func TestLocalMySQLExecutor(t *testing.T) {
	addr, err := localMySQLAddr()
	require.NoError(t, err)
	admin, _, _, err := connectLocalMySQL(addr)
	if err != nil {
		t.Skipf("local MySQL unavailable at %s: %v", addr, err)
	}
	_ = admin.Close()

	kt := kit.New()
	db, o, _, _, _ := openIsolatedMySQL(t)

	t.Run("up without init returns migrate.ErrPrecondition", func(t *testing.T) {
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "need_init")
		reg := mustRegistry(t, "main", []register.Migration{m})
		p, err := Prepare(kt, o, reg, Options{})
		require.Error(t, err)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), "do not exist")
	})

	t.Run("prepare and execute records success with applied_pkg", func(t *testing.T) {
		result, err := schema.InitTables(kt, o, nil, nil)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)

		first := mustMigration(t, migA, "v1.9.3", "20260101120000", "first")
		second := mustMigration(t, migB, "v1.9.4", "20260101120000", "second")
		ran := 0
		first.Up = func(_ context.Context, _ orm.Interface) error {
			ran++
			return nil
		}
		second.Up = func(_ context.Context, _ orm.Interface) error {
			ran++
			return nil
		}
		reg := mustRegistry(t, "main", []register.Migration{first, second})
		p, err := Prepare(kt, o, reg, Options{})
		require.NoError(t, err)
		require.True(t, p.Passed())
		r := Execute(kt, o, p)
		require.NoError(t, r.Err)
		assert.Equal(t, 2, ran)
		assert.Equal(t, []string{migA, migB}, r.Executed)
		assert.Equal(t, "v1.9.4", r.VersionAfter)

		rows := loadRows(t, db)
		require.Contains(t, rows, migA)
		require.Contains(t, rows, migB)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migA].Status)
		assert.Equal(t, first.Pkg, rows[migA].AppliedPkg)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migB].Status)
		assert.Equal(t, second.Pkg, rows[migB].AppliedPkg)
	})

	t.Run("failing up leaves failed record and rerun resumes", func(t *testing.T) {
		fail := mustMigration(t, migC, "v1.9.5", "20260101120000", "fail_then_ok")
		attempts := 0
		fail.Up = func(_ context.Context, _ orm.Interface) error {
			attempts++
			if attempts == 1 {
				return fmt.Errorf("first attempt boom")
			}
			return nil
		}
		reg := mustRegistry(t, "main", []register.Migration{fail})
		p, err := Prepare(kt, o, reg, Options{})
		require.NoError(t, err)
		r := Execute(kt, o, p)
		require.Error(t, r.Err)
		assert.Contains(t, r.Err.Error(), "first attempt boom")
		assert.Empty(t, r.Executed)

		rows := loadRows(t, db)
		require.Contains(t, rows, migC)
		assert.Equal(t, enumor.MigrationStatusFailed, rows[migC].Status)
		assert.Equal(t, fail.Pkg, rows[migC].AppliedPkg)
		assert.Contains(t, rows[migC].Message, "first attempt boom")

		// Failed records do not seed owners, so the next Prepare plans EXECUTE again.
		p2, err := Prepare(kt, o, reg, Options{})
		require.NoError(t, err)
		require.Len(t, p2.Items, 1)
		assert.Equal(t, enumor.MigrationActionExecute, p2.Items[0].Action)
		r2 := Execute(kt, o, p2)
		require.NoError(t, r2.Err)
		assert.Equal(t, []string{migC}, r2.Executed)
		assert.Equal(t, 2, attempts)
		assert.Equal(t, enumor.MigrationStatusSuccess, loadRows(t, db)[migC].Status)
	})

	t.Run("drop audit table prepare returns migrate.ErrPrecondition", func(t *testing.T) {
		dropAuditTable(t, db)
		assert.False(t, tableExists(t, db, constant.MigrationAuditTable))
		assert.True(t, tableExists(t, db, constant.MigrationRecordTable))

		m := mustMigration(t, migD, "v1.9.6", "20260101120000", "after_drop")
		reg := mustRegistry(t, "main", []register.Migration{m})
		p, err := Prepare(kt, o, reg, Options{})
		require.Error(t, err)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), constant.MigrationAuditTable)
	})
}

func TestOpenOrm(t *testing.T) {
	_, _, dbName, user, pass := openIsolatedMySQL(t)

	// Connect to a name this file is allowed to use, but do not create it, so
	// the single connection probe fails.
	missing := "hcm_migrate_engine_test_0_0"
	require.NoError(t, validateTestDatabase(missing))
	require.NotEqual(t, dbName, missing)
	_, err := openOrm(testDBOpt(user, pass, missing))
	require.Error(t, err)
	if pass != "" {
		assert.NotContains(t, err.Error(), pass)
	}

	got, err := openOrm(testDBOpt(user, pass, dbName))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "*orm.runtimeOrm", fmt.Sprintf("%T", got))
	assert.NotContains(t, fmt.Sprintf("%T", got), "modifySQL")

	var rows []struct {
		Name string `db:"name"`
	}
	err = got.Do().Select(context.Background(), &rows, "SELECT DATABASE() AS name", map[string]interface{}{})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, dbName, rows[0].Name)
	assert.NotEqual(t, "hcm", strings.ToLower(rows[0].Name))
}

func dropRecordTable(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("DROP TABLE IF EXISTS `hcm_migration_record`")
	require.NoError(t, err)
}

func dropAuditTable(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("DROP TABLE IF EXISTS `hcm_migration_audit`")
	require.NoError(t, err)
}

func tableExists(t *testing.T, db *sqlx.DB, name string) bool {
	t.Helper()
	var n int
	err := db.Get(&n, "SELECT COUNT(*) FROM information_schema.tables "+
		"WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' AND table_name = ?", name)
	require.NoError(t, err)
	return n > 0
}

func countRecords(t *testing.T, db *sqlx.DB) int {
	t.Helper()
	var n int
	err := db.Get(&n, "SELECT COUNT(*) FROM `hcm_migration_record`")
	require.NoError(t, err)
	return n
}

func countByID(t *testing.T, db *sqlx.DB, id string) int {
	t.Helper()
	var n int
	err := db.Get(&n, "SELECT COUNT(*) FROM `hcm_migration_record` WHERE `migration_id` = ?", id)
	require.NoError(t, err)
	return n
}

func loadRows(t *testing.T, db *sqlx.DB) map[string]schema.Record {
	t.Helper()
	var rows []schema.Record
	err := db.Select(&rows, "SELECT `id`, `migration_id`, `version`, `applied_pkg`, `status`, `message`, "+
		"`created_at`, `updated_at` FROM `hcm_migration_record`")
	require.NoError(t, err)
	out := make(map[string]schema.Record, len(rows))
	for _, row := range rows {
		out[row.MigrationID] = row
	}
	return out
}

func testDBOpt(user, pass, name string) cc.DataBase {
	return cc.DataBase{
		Resource: cc.ResourceDB{
			Endpoints:         []string{net.JoinHostPort(localMySQLHost, localMySQLPort)},
			Database:          name,
			User:              user,
			Password:          pass,
			DialTimeoutSec:    5,
			ReadTimeoutSec:    30,
			WriteTimeoutSec:   30,
			MaxOpenConn:       2,
			MaxIdleConn:       1,
			MaxIdleTimeoutMin: 1,
			TimeZone:          "UTC",
		},
		MaxSlowLogLatencyMS: 100,
		Limiter:             &cc.Limiter{QPS: 50, Burst: 50},
	}
}

func validateTestDatabase(name string) error {
	lower := strings.ToLower(name)
	switch lower {
	case "hcm", "mysql", "information_schema", "performance_schema", "sys":
		return fmt.Errorf("refusing application or system database %q", name)
	}
	if !testDatabaseRe.MatchString(name) {
		return fmt.Errorf("refusing database %q, only hcm_migrate_engine_test_<pid>_<n> is allowed", name)
	}
	return nil
}

func localMySQLAddr() (string, error) {
	addr := net.JoinHostPort(localMySQLHost, localMySQLPort)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return "", fmt.Errorf("migrate engine mysql tests refuse non-loopback host %s", host)
	}
	return addr, nil
}

func connectLocalMySQL(addr string) (*sqlx.DB, string, string, error) {
	var lastErr error
	for _, cred := range localMySQLCredentials {
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true&charset=utf8mb4", cred.user, cred.pass, addr)
		db, err := sqlx.Connect("mysql", dsn)
		if err == nil {
			return db, cred.user, cred.pass, nil
		}
		lastErr = err
	}
	return nil, "", "", lastErr
}

func nextTestDatabase() (string, error) {
	name := fmt.Sprintf("hcm_migrate_engine_test_%d_%d", os.Getpid(), atomic.AddUint64(&testDBSeq, 1))
	if err := validateTestDatabase(name); err != nil {
		return "", err
	}
	return name, nil
}

func openIsolatedMySQL(t *testing.T) (db *sqlx.DB, o orm.Interface, dbName, user, pass string) {
	t.Helper()
	addr, err := localMySQLAddr()
	require.NoError(t, err)
	dbName, err = nextTestDatabase()
	require.NoError(t, err)

	admin, user, pass, err := connectLocalMySQL(addr)
	require.NoError(t, err, "connect 127.0.0.1:3306 failed; this test does not open database hcm")

	t.Cleanup(func() {
		quoted, dropErr := quoteTestDatabase(dbName)
		if dropErr != nil {
			t.Errorf("skip drop, unsafe database name: %v", dropErr)
			_ = admin.Close()
			return
		}
		if _, dropErr = admin.Exec("DROP DATABASE IF EXISTS " + quoted); dropErr != nil {
			t.Errorf("drop test database %s failed, err: %v", dbName, dropErr)
		}
		_ = admin.Close()
	})

	quoted, err := quoteTestDatabase(dbName)
	require.NoError(t, err)
	_, err = admin.Exec("CREATE DATABASE " + quoted + " DEFAULT CHARSET utf8mb4")
	require.NoError(t, err)

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4", user, pass, addr, dbName)
	db, err = sqlx.Connect("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	var current sql.NullString
	require.NoError(t, db.Get(&current, "SELECT DATABASE()"))
	require.True(t, current.Valid)
	require.Equal(t, dbName, current.String)
	require.NotEqual(t, "hcm", strings.ToLower(current.String))

	o = orm.InitOrm(db, orm.MetricsRegisterer(prometheus.NewRegistry()))
	return db, o, dbName, user, pass
}

func quoteTestDatabase(name string) (string, error) {
	if err := validateTestDatabase(name); err != nil {
		return "", err
	}
	return "`" + name + "`", nil
}
