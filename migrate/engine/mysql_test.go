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
	"hcm/pkg/cc"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"

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
		{name: "obs database", dbName: "hcm_obs", wantErr: true},
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

	t.Run("init empty", func(t *testing.T) {
		created, adopted, err := InitRecordTable(kt, o, nil, nil)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Nil(t, adopted)
		assert.Equal(t, 0, countRecords(t, db))

		records, err := NewRecordStore(o).Load(kt)
		require.NoError(t, err)
		assert.Empty(t, records)
		current, ok, err := CurrentVersion(records)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, register.Version{}, current)
	})

	t.Run("second init is a no-op even with a baseline", func(t *testing.T) {
		baseline := mustVersion(t, "v1.9.10")
		migrations := []register.Migration{
			mustMigration(t, migA, "v1.9.3", "20260101120000", "would be adopted"),
		}
		created, adopted, err := InitRecordTable(kt, o, &baseline, migrations)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Equal(t, 0, countRecords(t, db))
	})

	baseline := mustVersion(t, "v1.9.3")
	// migA v1.9.3 is listed before the same ID at v1.9.2, so the kept row is v1.9.3.
	migrations := []register.Migration{
		mustMigration(t, migA, "v1.9.3", "20260201120000", "first"),
		mustMigration(t, migA, "v1.9.2", "20260101120000", "duplicate"),
		mustMigration(t, migB, "v1.9.3.0", "20260101120000", "folded"),
		mustMigration(t, migC, "v1.9.3.1", "20260101120000", "fourth"),
		mustMigration(t, migD, register.PendingVersion, "20260101120000", "pending"),
		mustMigration(t, migE, "v1.9.4", "20260101120000", "later"),
	}

	t.Run("init adopt", func(t *testing.T) {
		dropRecordTable(t, db)
		created, adopted, err := InitRecordTable(kt, o, &baseline, migrations)
		require.NoError(t, err)
		assert.True(t, created)
		require.Equal(t, []string{migA, migB}, migrationIDs(adopted))

		rows := loadRows(t, db)
		require.Len(t, rows, 2)
		assert.Equal(t, "v1.9.3", rows[migA].Version)
		assert.Equal(t, StatusSuccess, rows[migA].Status)
		assert.Equal(t, "", rows[migA].Message)
		assert.Equal(t, "v1.9.3.0", rows[migB].Version)
		assert.Equal(t, StatusSuccess, rows[migB].Status)
		_, hasFourth := rows[migC]
		assert.False(t, hasFourth)
		_, hasPending := rows[migD]
		assert.False(t, hasPending)
		_, hasLater := rows[migE]
		assert.False(t, hasLater)

		records, err := NewRecordStore(o).Load(kt)
		require.NoError(t, err)
		current, ok, err := CurrentVersion(records)
		require.NoError(t, err)
		require.True(t, ok)
		// v1.9.3 and v1.9.3.0 compare equal; the smaller raw text wins.
		assert.Equal(t, "v1.9.3", current.Raw)
	})

	t.Run("second adopt does not raise the baseline", func(t *testing.T) {
		higher := mustVersion(t, "v1.9.10")
		before := countRecords(t, db)
		created, adopted, err := InitRecordTable(kt, o, &higher, migrations)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Equal(t, before, countRecords(t, db))
		rows := loadRows(t, db)
		_, hasLater := rows[migE]
		assert.False(t, hasLater)
	})

	t.Run("mark running failed and success", func(t *testing.T) {
		store := NewRecordStore(o)
		running := register.Migration{ID: migF, Version: "v1.9.3"}
		require.NoError(t, store.MarkRunning(kt, running))
		rows := loadRows(t, db)
		require.Contains(t, rows, migF)
		assert.Equal(t, StatusRunning, rows[migF].Status)
		assert.Equal(t, "v1.9.3", rows[migF].Version)
		assert.Equal(t, "", rows[migF].Message)
		assert.Equal(t, 1, countByID(t, db, migF))

		raw := strings.Repeat("迁移失败，磁盘已满。", 80)
		require.Greater(t, len(raw), maxMessageBytes)
		require.NoError(t, store.MarkFailed(kt, running, fmt.Errorf("%s", raw)))
		rows = loadRows(t, db)
		msg := rows[migF].Message
		assert.Equal(t, StatusFailed, rows[migF].Status)
		assert.LessOrEqual(t, len(msg), maxMessageBytes)
		assert.True(t, utf8.ValidString(msg))
		assert.True(t, strings.HasPrefix(raw, msg))
		assert.NotEqual(t, raw, msg)
		assert.Equal(t, truncateMessage(raw, maxMessageBytes), msg)

		refreshed := register.Migration{ID: migF, Version: "v1.9.4"}
		require.NoError(t, store.MarkRunning(kt, refreshed))
		rows = loadRows(t, db)
		assert.Equal(t, StatusRunning, rows[migF].Status)
		assert.Equal(t, "", rows[migF].Message)
		// ON DUPLICATE KEY UPDATE clears the message and leaves the version.
		assert.Equal(t, "v1.9.3", rows[migF].Version)
		assert.Equal(t, 1, countByID(t, db, migF))

		require.NoError(t, store.MarkSuccess(kt, refreshed))
		rows = loadRows(t, db)
		assert.Equal(t, StatusSuccess, rows[migF].Status)
		assert.Equal(t, "v1.9.4", rows[migF].Version)
		assert.Equal(t, "", rows[migF].Message)

		require.NoError(t, store.MarkRunning(kt, register.Migration{ID: migC, Version: "v1.9.9"}))
		require.NoError(t, store.MarkSuccess(kt, register.Migration{ID: migC, Version: "v1.9.9"}))
		require.NoError(t, store.MarkRunning(kt, register.Migration{ID: migD, Version: "v1.9.10"}))
		require.NoError(t, store.MarkSuccess(kt, register.Migration{ID: migD, Version: "v1.9.10"}))
		require.NoError(t, store.MarkRunning(kt, register.Migration{ID: migE, Version: "v9.0.0"}))

		records, err := store.Load(kt)
		require.NoError(t, err)
		current, ok, err := CurrentVersion(records)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "v1.9.10", current.Raw)
	})

	t.Run("load rejects a raw running row with an unparsable version", func(t *testing.T) {
		dropRecordTable(t, db)
		created, adopted, err := InitRecordTable(kt, o, nil, nil)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Nil(t, adopted)

		_, err = db.Exec("INSERT INTO `hcm_migration_record` (`migration_id`, `version`, `status`) VALUES (?, ?, ?)",
			migA, "garbage", string(StatusRunning))
		require.NoError(t, err)

		records, err := NewRecordStore(o).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assert.ErrorIs(t, err, ErrPrecondition)
		assert.Contains(t, err.Error(), migA)
		assert.Contains(t, err.Error(), "garbage")
		assert.Contains(t, err.Error(), "unparsable version")
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
	assert.NotEqual(t, "hcm_obs", strings.ToLower(rows[0].Name))
}

func dropRecordTable(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("DROP TABLE IF EXISTS `hcm_migration_record`")
	require.NoError(t, err)
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

func loadRows(t *testing.T, db *sqlx.DB) map[string]Record {
	t.Helper()
	var rows []Record
	err := db.Select(&rows, "SELECT `id`, `migration_id`, `version`, `status`, `message`, `created_at`, `updated_at` "+
		"FROM `hcm_migration_record`")
	require.NoError(t, err)
	out := make(map[string]Record, len(rows))
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
	case "hcm", "hcm_obs", "mysql", "information_schema", "performance_schema", "sys":
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
