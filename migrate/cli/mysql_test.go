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

package cli

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

	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"

	// import mysql driver, used to create conn.
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Local MySQL tests use 127.0.0.1:3306 and never open the application database.
// Each test creates hcm_migrate_cli_test_<pid>_<n> and drops only that name.

const (
	cliMySQLHost = "127.0.0.1"
	cliMySQLPort = "3306"
)

var cliMySQLCredentials = []struct {
	user string
	pass string
}{
	{user: "root", pass: ""},
	{user: "root", pass: "admin"},
}

var cliTestDatabaseRe = regexp.MustCompile(`^hcm_migrate_cli_test_[0-9]+_[0-9]+$`)

var cliTestDBSeq uint64

func TestValidateCLITestDatabase(t *testing.T) {
	testCases := []struct {
		name    string
		dbName  string
		wantErr bool
	}{
		{name: "generated name", dbName: "hcm_migrate_cli_test_1_1"},
		{name: "app database", dbName: "hcm", wantErr: true},
		{name: "engine prefix", dbName: "hcm_migrate_engine_test_1_1", wantErr: true},
		{name: "empty", dbName: "", wantErr: true},
		{name: "injection", dbName: "hcm_migrate_cli_test_1_1`; DROP DATABASE hcm; --", wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCLITestDatabase(tc.dbName)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestLocalMySQLCLI(t *testing.T) {
	skipIfNoCLIMySQL(t)

	mainDB, mainOrm, _, _, _ := openCLIIsolatedMySQL(t)
	auxDB, auxOrm, _, _, _ := openCLIIsolatedMySQL(t)

	var mainUpCalls, auxUpCalls, auxFailCalls atomic.Int32
	mainA := mustMigration(t, "main", migA, "v1.9.3", "20260101120000", "main_a")
	mainB := mustMigration(t, "main", migB, "v1.9.4", "20260101120000", "main_b")
	mainA.Up = func(_ context.Context, _ orm.Interface) error {
		mainUpCalls.Add(1)
		return nil
	}
	mainB.Up = func(_ context.Context, _ orm.Interface) error {
		mainUpCalls.Add(1)
		return nil
	}
	auxOK := mustMigration(t, "aux", migC, "v1.9.3", "20260101120000", "aux_ok")
	auxFail := mustMigration(t, "aux", migD, "v1.9.4", "20260101120000", "aux_fail")
	auxOK.Up = func(_ context.Context, _ orm.Interface) error {
		auxUpCalls.Add(1)
		return nil
	}
	auxFail.Up = func(_ context.Context, _ orm.Interface) error {
		auxFailCalls.Add(1)
		return fmt.Errorf("aux boom")
	}

	mainReg := mustRegistry(t, "main", []register.Migration{mainA, mainB})
	auxReg := mustRegistry(t, "aux", []register.Migration{auxOK, auxFail})
	regs := []*register.Registry{mainReg, auxReg}
	source := &mapSource{orms: map[string]orm.Interface{
		"main": mainOrm,
		"aux":  auxOrm,
	}}

	runCLI := func(t *testing.T, args []string) (int, string, string) {
		t.Helper()
		r, stdout, stderr := newTestRunner(t, args, source, regs)
		code := r.run(args)
		return code, stdout.String(), stderr.String()
	}

	t.Run("up before init → exit 3 and audit not written", func(t *testing.T) {
		code, _, _ := runCLI(t, []string{"up", "-c", "fake.yaml", "-d", "main"})
		assert.Equal(t, constant.MigrationExitPrecondition, code)
		assert.False(t, cliTableExists(t, mainDB, constant.MigrationAuditTable))
		assert.False(t, cliTableExists(t, mainDB, constant.MigrationRecordTable))
		assert.Equal(t, 0, cliCountAudits(t, mainDB))
	})

	t.Run("init --mode=empty creates both tables and one audit row", func(t *testing.T) {
		code, out, _ := runCLI(t, []string{"init", "-c", "fake.yaml", "--mode=empty", "-d", "main"})
		assert.Equal(t, constant.MigrationExitSuccess, code)
		assert.Contains(t, out, "created")
		assert.True(t, cliTableExists(t, mainDB, constant.MigrationAuditTable))
		assert.True(t, cliTableExists(t, mainDB, constant.MigrationRecordTable))
		assert.Equal(t, 0, cliCountRecords(t, mainDB))
		rows := cliLoadAudits(t, mainDB)
		require.Len(t, rows, 1)
		assert.Equal(t, "init", rows[0].Command)
		assert.Equal(t, string(enumor.MigrationStatusSuccess), rows[0].Status)
		require.True(t, rows[0].ExitCode.Valid)
		assert.Equal(t, int64(0), rows[0].ExitCode.Int64)
	})

	t.Run("init again → no change and another audit row", func(t *testing.T) {
		before := cliCountAudits(t, mainDB)
		code, out, _ := runCLI(t, []string{"init", "-c", "fake.yaml", "--mode=empty", "-d", "main"})
		assert.Equal(t, constant.MigrationExitSuccess, code)
		assert.Contains(t, out, "already initialized")
		assert.Equal(t, 0, cliCountRecords(t, mainDB))
		assert.Equal(t, before+1, cliCountAudits(t, mainDB))
	})

	t.Run("init --mode=adopt records baseline without calling Up", func(t *testing.T) {
		freshDB, freshOrm, _, _, _ := openCLIIsolatedMySQL(t)
		freshA := mustMigration(t, "main", migA, "v1.9.3", "20260101120000", "adopt_a")
		freshB := mustMigration(t, "main", migB, "v1.9.4", "20260101120000", "adopt_b")
		var calls atomic.Int32
		freshA.Up = func(_ context.Context, _ orm.Interface) error {
			calls.Add(1)
			return nil
		}
		freshB.Up = func(_ context.Context, _ orm.Interface) error {
			calls.Add(1)
			return nil
		}
		freshReg := mustRegistry(t, "main", []register.Migration{freshA, freshB})
		args := []string{"init", "-c", "fake.yaml", "--mode=adopt", "--baseline", "main=v1.9.3", "-d", "main"}
		r, stdout, stderr := newTestRunner(t, args, &mapSource{orms: map[string]orm.Interface{
			"main": freshOrm,
		}}, []*register.Registry{freshReg})
		code := r.run(args)
		assertExit(t, code, constant.MigrationExitSuccess, stdout, stderr)
		assert.Equal(t, int32(0), calls.Load())
		rows := cliLoadRecords(t, freshDB)
		require.Len(t, rows, 1)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migA].Status)
		assert.Equal(t, freshA.Pkg, rows[migA].AppliedPkg)
		_, hasB := rows[migB]
		assert.False(t, hasB)

		audits := cliLoadAudits(t, freshDB)
		require.Len(t, audits, 1)
		assert.Equal(t, "init", audits[0].Command)
		assert.Equal(t, string(enumor.MigrationStatusSuccess), audits[0].Status)
		assert.Contains(t, audits[0].Skipped.String, string(enumor.MigrationSkipBaseline))
		assert.Equal(t, "", audits[0].VersionAfter)
	})

	t.Run("init --plan writes nothing", func(t *testing.T) {
		freshDB, freshOrm, _, _, _ := openCLIIsolatedMySQL(t)
		freshReg := mustRegistry(t, "main", []register.Migration{
			mustMigration(t, "main", migA, "v1.9.3", "20260101120000", "plan"),
		})
		args := []string{"init", "-c", "fake.yaml", "--mode=empty", "--plan", "-d", "main"}
		r, stdout, stderr := newTestRunner(t, args, &mapSource{orms: map[string]orm.Interface{
			"main": freshOrm,
		}}, []*register.Registry{freshReg})
		code := r.run(args)
		assertExit(t, code, constant.MigrationExitSuccess, stdout, stderr)
		assert.Contains(t, stdout.String(), "would create")
		assert.False(t, cliTableExists(t, freshDB, constant.MigrationAuditTable))
		assert.False(t, cliTableExists(t, freshDB, constant.MigrationRecordTable))
	})

	t.Run("up --plan writes no record and no audit row", func(t *testing.T) {
		beforeAudits := cliCountAudits(t, mainDB)
		beforeRecords := cliCountRecords(t, mainDB)
		code, out, _ := runCLI(t, []string{"up", "-c", "fake.yaml", "--plan", "-d", "main"})
		assert.Equal(t, constant.MigrationExitSuccess, code)
		assert.Contains(t, out, "EXECUTE")
		assert.Equal(t, beforeAudits, cliCountAudits(t, mainDB))
		assert.Equal(t, beforeRecords, cliCountRecords(t, mainDB))
		assert.Equal(t, int32(0), mainUpCalls.Load())
	})

	t.Run("up executes and records success with applied_pkg", func(t *testing.T) {
		mainUpCalls.Store(0)
		beforeAudits := cliCountAudits(t, mainDB)
		code, out, _ := runCLI(t, []string{"up", "-c", "fake.yaml", "-d", "main"})
		assert.Equal(t, constant.MigrationExitSuccess, code)
		assert.Contains(t, out, "summary:")
		assert.Contains(t, out, "success")
		assert.Equal(t, int32(2), mainUpCalls.Load())
		rows := cliLoadRecords(t, mainDB)
		require.Contains(t, rows, migA)
		require.Contains(t, rows, migB)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migA].Status)
		assert.Equal(t, mainA.Pkg, rows[migA].AppliedPkg)
		assert.Equal(t, enumor.MigrationStatusSuccess, rows[migB].Status)
		assert.Equal(t, mainB.Pkg, rows[migB].AppliedPkg)

		audits := cliLoadAudits(t, mainDB)
		require.Equal(t, beforeAudits+1, len(audits))
		last := audits[len(audits)-1]
		assert.Equal(t, "up", last.Command)
		assert.Equal(t, string(enumor.MigrationStatusSuccess), last.Status)
		require.True(t, last.ExitCode.Valid)
		assert.Equal(t, int64(0), last.ExitCode.Int64)
		assert.Equal(t, "", last.VersionBefore)
		assert.Equal(t, "v1.9.4", last.VersionAfter)
	})

	t.Run("multi-db shared run_id and failing aux", func(t *testing.T) {
		// Reset aux DB; main already migrated. Init aux only.
		code, _, _ := runCLI(t, []string{"init", "-c", "fake.yaml", "--mode=empty", "-d", "aux"})
		require.Equal(t, constant.MigrationExitSuccess, code)

		mainUpCalls.Store(0)
		auxUpCalls.Store(0)
		auxFailCalls.Store(0)
		// main has nothing left to run; aux runs ok then fails.
		code, out, _ := runCLI(t, []string{"up", "-c", "fake.yaml"})
		assert.Equal(t, constant.MigrationExitFailure, code)
		assert.Contains(t, out, "summary:")
		assert.Equal(t, int32(0), mainUpCalls.Load())
		assert.Equal(t, int32(1), auxUpCalls.Load())
		assert.Equal(t, int32(1), auxFailCalls.Load())

		auxRows := cliLoadRecords(t, auxDB)
		require.Contains(t, auxRows, migC)
		assert.Equal(t, enumor.MigrationStatusSuccess, auxRows[migC].Status)
		require.Contains(t, auxRows, migD)
		assert.Equal(t, enumor.MigrationStatusFailed, auxRows[migD].Status)

		mainAudits := cliLoadAudits(t, mainDB)
		auxAudits := cliLoadAudits(t, auxDB)
		mainLast := mainAudits[len(mainAudits)-1]
		auxLast := auxAudits[len(auxAudits)-1]
		assert.Equal(t, "up", mainLast.Command)
		assert.Equal(t, "up", auxLast.Command)
		assert.Equal(t, mainLast.RunID, auxLast.RunID, "both databases share the run kit rid")
		require.True(t, mainLast.ExitCode.Valid)
		assert.Equal(t, int64(1), mainLast.ExitCode.Int64)
		assert.Equal(t, string(enumor.MigrationStatusFailed), mainLast.Status)
		require.True(t, auxLast.ExitCode.Valid)
		assert.Equal(t, int64(1), auxLast.ExitCode.Int64)
	})

	t.Run("missed migration default mode → exit 4 then catch-up", func(t *testing.T) {
		freshDB, freshOrm, _, _, _ := openCLIIsolatedMySQL(t)
		high := mustMigration(t, "main", migB, "v1.9.4", "20260101120000", "high")
		low := mustMigration(t, "main", migA, "v1.9.3", "20260101120000", "low")
		var highCalls, lowCalls atomic.Int32
		high.Up = func(_ context.Context, _ orm.Interface) error {
			highCalls.Add(1)
			return nil
		}
		low.Up = func(_ context.Context, _ orm.Interface) error {
			lowCalls.Add(1)
			return nil
		}

		// First: only high in registry, run it.
		regHigh := mustRegistry(t, "main", []register.Migration{high})
		src := &mapSource{orms: map[string]orm.Interface{"main": freshOrm}}
		args := []string{"init", "-c", "fake.yaml", "--mode=empty", "-d", "main"}
		r, _, _ := newTestRunner(t, args, src, []*register.Registry{regHigh})
		require.Equal(t, constant.MigrationExitSuccess, r.run(args))
		args = []string{"up", "-c", "fake.yaml", "-d", "main"}
		r, _, _ = newTestRunner(t, args, src, []*register.Registry{regHigh})
		require.Equal(t, constant.MigrationExitSuccess, r.run(args))
		require.Equal(t, int32(1), highCalls.Load())

		// Second: registry has missed low below current; default up fails.
		regBoth := mustRegistry(t, "main", []register.Migration{low, high})
		auxFreshDB, auxFreshOrm, _, _, _ := openCLIIsolatedMySQL(t)
		auxM := mustMigration(t, "aux", migC, "v1.9.5", "20260101120000", "aux_later")
		var auxCalls atomic.Int32
		auxM.Up = func(_ context.Context, _ orm.Interface) error {
			auxCalls.Add(1)
			return nil
		}
		auxFreshReg := mustRegistry(t, "aux", []register.Migration{auxM})
		multiSrc := &mapSource{orms: map[string]orm.Interface{
			"main": freshOrm,
			"aux":  auxFreshOrm,
		}}
		initAux := []string{"init", "-c", "fake.yaml", "--mode=empty", "-d", "aux"}
		r, _, _ = newTestRunner(t, initAux, multiSrc, []*register.Registry{regBoth, auxFreshReg})
		require.Equal(t, constant.MigrationExitSuccess, r.run(initAux))

		args = []string{"up", "-c", "fake.yaml"}
		r, stdout, stderr := newTestRunner(t, args, multiSrc, []*register.Registry{regBoth, auxFreshReg})
		code := r.run(args)
		assertExit(t, code, constant.MigrationExitMissed, stdout, stderr)
		assert.Equal(t, int32(0), lowCalls.Load())
		assert.Equal(t, int32(0), auxCalls.Load(), "nothing executed in any db")
		assert.Equal(t, 1, cliCountRecords(t, freshDB), "main still only high")
		assert.Equal(t, 0, cliCountRecords(t, auxFreshDB))

		mainAudits := cliLoadAudits(t, freshDB)
		auxAudits := cliLoadAudits(t, auxFreshDB)
		mainLast := mainAudits[len(mainAudits)-1]
		auxLast := auxAudits[len(auxAudits)-1]
		assert.Equal(t, mainLast.RunID, auxLast.RunID)
		assert.Equal(t, int64(4), mainLast.ExitCode.Int64)
		assert.Equal(t, int64(4), auxLast.ExitCode.Int64)
		assert.Equal(t, string(enumor.MigrationStatusFailed), mainLast.Status)
		assert.Contains(t, mainLast.Message.String, "missed migration")

		// status shows <= current version mark
		args = []string{"status", "-c", "fake.yaml", "-d", "main"}
		r, stdout, stderr = newTestRunner(t, args, multiSrc, []*register.Registry{regBoth, auxFreshReg})
		code = r.run(args)
		assertExit(t, code, constant.MigrationExitSuccess, stdout, stderr)
		assert.Contains(t, stdout.String(), "<= current version")
		assert.Contains(t, stdout.String(), migA)

		// catch-up runs the missed one
		args = []string{"up", "-c", "fake.yaml", "--catch-up", "-d", "main"}
		r, stdout, stderr = newTestRunner(t, args, multiSrc, []*register.Registry{regBoth})
		code = r.run(args)
		assertExit(t, code, constant.MigrationExitSuccess, stdout, stderr)
		assert.Equal(t, int32(1), lowCalls.Load())
		assert.Contains(t, cliLoadRecords(t, freshDB), migA)
	})

	t.Run("status after runs shows counts", func(t *testing.T) {
		code, out, _ := runCLI(t, []string{"status", "-c", "fake.yaml", "-d", "main"})
		assert.Equal(t, constant.MigrationExitSuccess, code)
		assert.Contains(t, out, "current version: v1.9.4")
		assert.Contains(t, out, "records: success")
		assert.Contains(t, out, "not executed:")
		assert.NotContains(t, out, "not executed: none")
		assert.NotContains(t, out, "RECORD")
	})

	t.Run("status on uninitialized db → exit 3", func(t *testing.T) {
		freshDB, freshOrm, _, _, _ := openCLIIsolatedMySQL(t)
		_ = freshDB
		freshReg := mustRegistry(t, "main", []register.Migration{
			mustMigration(t, "main", migA, "v1.9.3", "20260101120000", "uninit"),
		})
		args := []string{"status", "-c", "fake.yaml", "-d", "main"}
		r, stdout, stderr := newTestRunner(t, args, &mapSource{orms: map[string]orm.Interface{
			"main": freshOrm,
		}}, []*register.Registry{freshReg})
		code := r.run(args)
		assertExit(t, code, constant.MigrationExitPrecondition, stdout, stderr)
		assert.Contains(t, stdout.String(), "do not exist")
	})
}

type cliAuditRow struct {
	RunID         string         `db:"run_id"`
	Command       string         `db:"command"`
	Status        string         `db:"status"`
	ExitCode      sql.NullInt64  `db:"exit_code"`
	VersionBefore string         `db:"version_before"`
	VersionAfter  string         `db:"version_after"`
	Skipped       sql.NullString `db:"skipped"`
	Message       sql.NullString `db:"message"`
}

type cliRecordRow struct {
	MigrationID string                 `db:"migration_id"`
	Version     string                 `db:"version"`
	AppliedPkg  string                 `db:"applied_pkg"`
	Status      enumor.MigrationStatus `db:"status"`
	Message     string                 `db:"message"`
}

func cliTableExists(t *testing.T, db *sqlx.DB, name string) bool {
	t.Helper()
	var n int
	err := db.Get(&n, "SELECT COUNT(*) FROM information_schema.tables "+
		"WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' AND table_name = ?", name)
	require.NoError(t, err)
	return n > 0
}

func cliCountRecords(t *testing.T, db *sqlx.DB) int {
	t.Helper()
	if !cliTableExists(t, db, constant.MigrationRecordTable) {
		return 0
	}
	var n int
	require.NoError(t, db.Get(&n, "SELECT COUNT(*) FROM `hcm_migration_record`"))
	return n
}

func cliCountAudits(t *testing.T, db *sqlx.DB) int {
	t.Helper()
	if !cliTableExists(t, db, constant.MigrationAuditTable) {
		return 0
	}
	var n int
	require.NoError(t, db.Get(&n, "SELECT COUNT(*) FROM `hcm_migration_audit`"))
	return n
}

func cliLoadAudits(t *testing.T, db *sqlx.DB) []cliAuditRow {
	t.Helper()
	var rows []cliAuditRow
	err := db.Select(&rows, "SELECT `run_id`, `command`, `status`, `exit_code`, `version_before`, "+
		"`version_after`, `skipped`, `message` FROM `hcm_migration_audit` ORDER BY `id`")
	require.NoError(t, err)
	return rows
}

func cliLoadRecords(t *testing.T, db *sqlx.DB) map[string]cliRecordRow {
	t.Helper()
	var rows []cliRecordRow
	err := db.Select(&rows, "SELECT `migration_id`, `version`, `applied_pkg`, `status`, `message` "+
		"FROM `hcm_migration_record`")
	require.NoError(t, err)
	out := make(map[string]cliRecordRow, len(rows))
	for _, row := range rows {
		out[row.MigrationID] = row
	}
	return out
}

func validateCLITestDatabase(name string) error {
	lower := strings.ToLower(name)
	switch lower {
	case "hcm", "mysql", "information_schema", "performance_schema", "sys":
		return fmt.Errorf("refusing application or system database %q", name)
	}
	if !cliTestDatabaseRe.MatchString(name) {
		return fmt.Errorf("refusing database %q, only hcm_migrate_cli_test_<pid>_<n> is allowed", name)
	}
	return nil
}

func cliMySQLAddr() (string, error) {
	addr := net.JoinHostPort(cliMySQLHost, cliMySQLPort)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return "", fmt.Errorf("migrate cli mysql tests refuse non-loopback host %s", host)
	}
	return addr, nil
}

func connectCLIMySQL(addr string) (*sqlx.DB, string, string, error) {
	var lastErr error
	for _, cred := range cliMySQLCredentials {
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true&charset=utf8mb4", cred.user, cred.pass, addr)
		db, err := sqlx.Connect("mysql", dsn)
		if err == nil {
			return db, cred.user, cred.pass, nil
		}
		lastErr = err
	}
	return nil, "", "", lastErr
}

func skipIfNoCLIMySQL(t *testing.T) {
	t.Helper()
	addr, err := cliMySQLAddr()
	require.NoError(t, err)
	admin, _, _, err := connectCLIMySQL(addr)
	if err != nil {
		t.Skipf("local MySQL unavailable at %s: %v", addr, err)
	}
	_ = admin.Close()
}

func nextCLITestDatabase() (string, error) {
	name := fmt.Sprintf("hcm_migrate_cli_test_%d_%d", os.Getpid(), atomic.AddUint64(&cliTestDBSeq, 1))
	if err := validateCLITestDatabase(name); err != nil {
		return "", err
	}
	return name, nil
}

func quoteCLITestDatabase(name string) (string, error) {
	if err := validateCLITestDatabase(name); err != nil {
		return "", err
	}
	return "`" + name + "`", nil
}

func openCLIIsolatedMySQL(t *testing.T) (db *sqlx.DB, o orm.Interface, dbName, user, pass string) {
	t.Helper()
	addr, err := cliMySQLAddr()
	require.NoError(t, err)
	dbName, err = nextCLITestDatabase()
	require.NoError(t, err)

	admin, user, pass, err := connectCLIMySQL(addr)
	require.NoError(t, err, "connect 127.0.0.1:3306 failed; this test does not open database hcm")

	t.Cleanup(func() {
		quoted, dropErr := quoteCLITestDatabase(dbName)
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

	quoted, err := quoteCLITestDatabase(dbName)
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
