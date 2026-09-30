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
	"database/sql"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

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
// Each test creates hcm_migrate_util_test_<pid>_<n> and drops only that name.

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
var testDatabaseRe = regexp.MustCompile(`^hcm_migrate_util_test_[0-9]+_[0-9]+$`)

var testDBSeq uint64

func TestValidateTestDatabase(t *testing.T) {
	testCases := []struct {
		name    string
		dbName  string
		wantErr bool
	}{
		{name: "generated name", dbName: "hcm_migrate_util_test_1_1"},
		{name: "app database", dbName: "hcm", wantErr: true},
		{name: "app database upper", dbName: "HCM", wantErr: true},
		{name: "mysql system", dbName: "mysql", wantErr: true},
		{name: "information_schema", dbName: "information_schema", wantErr: true},
		{name: "empty", dbName: "", wantErr: true},
		{name: "prefix only", dbName: "hcm_migrate_util_test_", wantErr: true},
		{name: "injection", dbName: "hcm_migrate_util_test_1_1`; DROP DATABASE hcm; --", wantErr: true},
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

func TestLocalMySQLIdempotent(t *testing.T) {
	ctx := context.Background()
	db, o, counter := openIsolatedMySQL(t)

	t.Run("created flag", func(t *testing.T) {
		const ddl = "CREATE TABLE `created_flag` (`id` bigint NOT NULL, PRIMARY KEY (`id`))"
		created, err := CreateTableIfNotExists(ctx, o, "created_flag", ddl)
		require.NoError(t, err)
		assert.True(t, created)

		created, err = CreateTableIfNotExists(ctx, o, "created_flag", ddl)
		require.NoError(t, err)
		assert.False(t, created)
	})

	runTwiceNoExtraExec(t, counter, func() error {
		_, err := CreateTableIfNotExists(ctx, o, "probe",
			"CREATE TABLE `probe` (`id` bigint NOT NULL, `cloud_id` varchar(64) NOT NULL DEFAULT '', "+
				"PRIMARY KEY (`id`))")
		return err
	})

	has, err := NewMetaOrm(o).HasTable(ctx, "probe")
	require.NoError(t, err)
	assert.True(t, has)

	for _, tc := range append(mysqlTextColumnCases(), mysqlExprColumnCases()...) {
		t.Run(tc.name, func(t *testing.T) {
			counter.execs = 0
			require.NoError(t, AddColumn(ctx, o, tc.opt))
			assert.Equal(t, 1, counter.execs)
			assertColumn(t, db, tc.opt.Table, tc.opt.Column, tc.want)

			require.NoError(t, AddColumn(ctx, o, tc.opt))
			assert.Equal(t, 1, counter.execs, "second AddColumn must not emit DDL")
			assertColumn(t, db, tc.opt.Table, tc.opt.Column, tc.want)
		})
	}

}

func TestLocalMySQLIdempotentSchema(t *testing.T) {
	ctx := context.Background()
	db, o, counter := openIsolatedMySQL(t)
	_, err := CreateTableIfNotExists(ctx, o, "probe",
		"CREATE TABLE `probe` (`id` bigint NOT NULL, `cloud_id` varchar(64) NOT NULL DEFAULT '', PRIMARY KEY (`id`))")
	require.NoError(t, err)

	t.Run("add and drop index", func(t *testing.T) {
		runTwiceNoExtraExec(t, counter, func() error {
			return AddIndex(ctx, o, "probe", "idx_state_updated_at", []string{"id", "cloud_id"}, false)
		})
		hasIndex, err := NewMetaOrm(o).HasIndex(ctx, "probe", "idx_state_updated_at")
		require.NoError(t, err)
		assert.True(t, hasIndex)

		runTwiceNoExtraExec(t, counter, func() error {
			return DropIndex(ctx, o, "probe", "idx_state_updated_at")
		})
		hasIndex, err = NewMetaOrm(o).HasIndex(ctx, "probe", "idx_state_updated_at")
		require.NoError(t, err)
		assert.False(t, hasIndex)
	})

	t.Run("rename column", func(t *testing.T) {
		require.NoError(t, AddColumn(ctx, o, AddColumnOpt{Table: "probe", Column: "old_name", Type: "varchar(16)"}))
		runTwiceNoExtraExec(t, counter, func() error {
			return RenameColumn(ctx, o, "probe", "old_name", "new_name")
		})
		meta := NewMetaOrm(o)
		oldExists, err := meta.HasColumn(ctx, "probe", "old_name")
		require.NoError(t, err)
		newExists, err := meta.HasColumn(ctx, "probe", "new_name")
		require.NoError(t, err)
		assert.False(t, oldExists)
		assert.True(t, newExists)
	})

	t.Run("drop column", func(t *testing.T) {
		require.NoError(t, AddColumn(ctx, o, AddColumnOpt{Table: "probe", Column: "instance_id", Type: "varchar(64)"}))
		runTwiceNoExtraExec(t, counter, func() error {
			return DropColumn(ctx, o, "probe", "instance_id")
		})
		exists, err := NewMetaOrm(o).HasColumn(ctx, "probe", "instance_id")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("foreign key", func(t *testing.T) {
		_, err := CreateTableIfNotExists(ctx, o, "parent_row",
			"CREATE TABLE `parent_row` (`id` bigint NOT NULL, PRIMARY KEY (`id`))")
		require.NoError(t, err)
		_, err = CreateTableIfNotExists(ctx, o, "child_row",
			"CREATE TABLE `child_row` (`id` bigint NOT NULL, `parent_id` bigint NULL, PRIMARY KEY (`id`))")
		require.NoError(t, err)
		add := "ALTER TABLE `child_row` ADD CONSTRAINT `fk_parent` FOREIGN KEY (`parent_id`) " +
			"REFERENCES `parent_row` (`id`)"
		runTwiceNoExtraExec(t, counter, func() error {
			return AddConstraint(ctx, o, "child_row", "fk_parent", add)
		})
		hasFK, err := NewMetaOrm(o).HasConstraint(ctx, "child_row", "fk_parent")
		require.NoError(t, err)
		assert.True(t, hasFK)

		drop := "ALTER TABLE `child_row` DROP FOREIGN KEY `fk_parent`"
		runTwiceNoExtraExec(t, counter, func() error {
			return DropConstraint(ctx, o, "child_row", "fk_parent", drop)
		})
		hasFK, err = NewMetaOrm(o).HasConstraint(ctx, "child_row", "fk_parent")
		require.NoError(t, err)
		assert.False(t, hasFK)
	})

	t.Run("id generator insert", func(t *testing.T) {
		_, err := CreateTableIfNotExists(ctx, o, "id_generator",
			"CREATE TABLE `id_generator` (`resource` varchar(64) NOT NULL, `max_id` varchar(64) NOT NULL, "+
				"PRIMARY KEY (`resource`))")
		require.NoError(t, err)
		require.NoError(t, InsertIDGenerator(ctx, o, "account", "0"))
		require.NoError(t, InsertIDGenerator(ctx, o, "account", "999"))
		var row struct {
			Count int    `db:"cnt"`
			MaxID string `db:"max_id"`
		}
		err = db.Get(&row, "SELECT COUNT(*) AS cnt, MAX(`max_id`) AS max_id FROM `id_generator` "+
			"WHERE `resource` = ?", "account")
		require.NoError(t, err)
		assert.Equal(t, 1, row.Count)
		assert.Equal(t, "0", row.MaxID)
	})

	t.Run("drop table", func(t *testing.T) {
		_, err := CreateTableIfNotExists(ctx, o, "gone",
			"CREATE TABLE `gone` (`id` bigint NOT NULL, PRIMARY KEY (`id`))")
		require.NoError(t, err)
		runTwiceNoExtraExec(t, counter, func() error {
			return DropTable(ctx, o, "gone")
		})
		exists, err := NewMetaOrm(o).HasTable(ctx, "gone")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

type mysqlColumnCase struct {
	name string
	opt  AddColumnOpt
	want columnExpect
}

func mysqlTextColumnCases() []mysqlColumnCase {
	return []mysqlColumnCase{
		{
			name: "varchar not null no default",
			opt:  AddColumnOpt{Table: "probe", Column: "managers", Type: "varchar(255)", NotNull: true},
			want: columnExpect{typ: "varchar(255)", nullable: "NO", defNull: true},
		},
		{
			name: "string default and after",
			opt: AddColumnOpt{Table: "probe", Column: "source", Type: "varchar(64)",
				Default: StringDefault("itsm"), After: "id"},
			want: columnExpect{typ: "varchar(64)", nullable: "YES", def: "itsm"},
		},
		{
			name: "empty string default",
			opt: AddColumnOpt{Table: "probe", Column: "bk_asset_id", Type: "varchar(64)",
				Default: StringDefault(""), Comment: "固资号"},
			want: columnExpect{typ: "varchar(64)", nullable: "YES", def: "", comment: "固资号"},
		},
		{
			name: "quoted null stays a string",
			opt: AddColumnOpt{Table: "probe", Column: "memo", Type: "varchar(255)",
				Default: StringDefault("NULL")},
			want: columnExpect{typ: "varchar(255)", nullable: "YES", def: "NULL"},
		},
		{
			name: "comment with quote",
			opt: AddColumnOpt{Table: "probe", Column: "note", Type: "varchar(64)",
				Default: StringDefault("it's\\x"), Comment: "it's\\x"},
			want: columnExpect{typ: "varchar(64)", nullable: "YES", def: "it's\\x", comment: "it's\\x"},
		},
	}
}

func mysqlExprColumnCases() []mysqlColumnCase {
	return []mysqlColumnCase{
		{
			name: "null keyword",
			opt: AddColumnOpt{Table: "probe", Column: "security_managers", Type: "json",
				Default: ExprDefault("NULL"), Comment: "安全负责人"},
			want: columnExpect{typ: "json", nullable: "YES", defNull: true, comment: "安全负责人"},
		},
		{
			name: "negative number",
			opt: AddColumnOpt{Table: "probe", Column: "bk_host_id", Type: "bigint",
				Default: ExprDefault("-1"), Comment: "主机ID"},
			want: columnExpect{typ: "bigint", nullable: "YES", def: "-1", comment: "主机ID"},
		},
		{
			name: "not null zero",
			opt: AddColumnOpt{Table: "probe", Column: "bk_biz_id", Type: "bigint",
				NotNull: true, Default: ExprDefault("0")},
			want: columnExpect{typ: "bigint", nullable: "NO", def: "0"},
		},
		{
			name: "boolean false",
			opt: AddColumnOpt{Table: "probe", Column: "not_notice", Type: "boolean",
				Default: ExprDefault("false")},
			want: columnExpect{typ: "tinyint", nullable: "YES", def: "0"},
		},
		{
			name: "current timestamp",
			opt: AddColumnOpt{Table: "probe", Column: "created_at", Type: "timestamp",
				NotNull: true, Default: ExprDefault("CURRENT_TIMESTAMP")},
			want: columnExpect{typ: "timestamp", nullable: "NO", def: "CURRENT_TIMESTAMP"},
		},
		{
			name: "unsigned zero",
			opt: AddColumnOpt{Table: "probe", Column: "exempted_returned_core", Type: "bigint unsigned",
				NotNull: true, Default: ExprDefault("0")},
			want: columnExpect{typ: "bigint unsigned", nullable: "NO", def: "0"},
		},
		{
			name: "decimal zero",
			opt: AddColumnOpt{Table: "probe", Column: "tech_class_res_amt", Type: "decimal(10,2)",
				NotNull: true, Default: ExprDefault("0")},
			want: columnExpect{typ: "decimal(10,2)", nullable: "NO", def: "0.00"},
		},
	}
}

type columnExpect struct {
	typ      string
	nullable string
	def      string
	defNull  bool
	comment  string
}

func assertColumn(t *testing.T, db *sqlx.DB, table, column string, want columnExpect) {
	t.Helper()
	var got struct {
		ColumnType    string         `db:"COLUMN_TYPE"`
		IsNullable    string         `db:"IS_NULLABLE"`
		ColumnDefault sql.NullString `db:"COLUMN_DEFAULT"`
		ColumnComment string         `db:"COLUMN_COMMENT"`
	}
	err := db.Get(&got, "SELECT COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT "+
		"FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		table, column)
	require.NoError(t, err)
	assert.Equal(t, want.typ, normalizeColumnType(got.ColumnType))
	assert.Equal(t, want.nullable, got.IsNullable)
	assert.Equal(t, want.comment, got.ColumnComment)
	if want.defNull {
		assert.False(t, got.ColumnDefault.Valid, "column default = %#v", got.ColumnDefault)
		return
	}
	require.True(t, got.ColumnDefault.Valid)
	assert.Equal(t, strings.ToLower(want.def), normalizeDefault(got.ColumnDefault.String))
}

func normalizeColumnType(typ string) string {
	typ = strings.ToLower(typ)
	return regexp.MustCompile(`\b(tinyint|smallint|mediumint|int|bigint)\(\d+\)`).ReplaceAllString(typ, "$1")
}

func normalizeDefault(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	return strings.TrimSuffix(v, "()")
}

func runTwiceNoExtraExec(t *testing.T, counter *execCounter, fn func() error) {
	t.Helper()
	counter.execs = 0
	require.NoError(t, fn())
	first := counter.execs
	require.Equal(t, 1, first)
	require.NoError(t, fn())
	assert.Equal(t, first, counter.execs)
}

func validateTestDatabase(name string) error {
	lower := strings.ToLower(name)
	switch lower {
	case "hcm", "mysql", "information_schema", "performance_schema", "sys":
		return fmt.Errorf("refusing application or system database %q", name)
	}
	if !testDatabaseRe.MatchString(name) {
		return fmt.Errorf("refusing database %q, only hcm_migrate_util_test_<pid>_<n> is allowed", name)
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
		return "", fmt.Errorf("migrate util mysql tests refuse non-loopback host %s", host)
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
	name := fmt.Sprintf("hcm_migrate_util_test_%d_%d", os.Getpid(), atomic.AddUint64(&testDBSeq, 1))
	if err := validateTestDatabase(name); err != nil {
		return "", err
	}
	return name, nil
}

func openIsolatedMySQL(t *testing.T) (*sqlx.DB, orm.Interface, *execCounter) {
	t.Helper()
	addr, err := localMySQLAddr()
	require.NoError(t, err)
	dbName, err := nextTestDatabase()
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
	db, err := sqlx.Connect("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	var current sql.NullString
	require.NoError(t, db.Get(&current, "SELECT DATABASE()"))
	require.True(t, current.Valid)
	require.Equal(t, dbName, current.String)
	require.NotEqual(t, "hcm", strings.ToLower(current.String))

	counter := &execCounter{}
	raw := orm.InitOrm(db, orm.MetricsRegisterer(prometheus.NewRegistry()))
	return db, &trackingOrm{inner: raw, counter: counter}, counter
}

func quoteTestDatabase(name string) (string, error) {
	if err := validateTestDatabase(name); err != nil {
		return "", err
	}
	return "`" + name + "`", nil
}

type execCounter struct {
	execs int
}

type trackingOrm struct {
	inner   orm.Interface
	counter *execCounter
}

func (t *trackingOrm) Do() orm.DoOrm {
	return &trackingDo{inner: t.inner.Do(), counter: t.counter}
}

func (t *trackingOrm) Txn(tx *sqlx.Tx) orm.DoOrmWithTransaction {
	panic("trackingOrm: Txn not used")
}

func (t *trackingOrm) AutoTxn(kt *kit.Kit, run orm.TxnFunc) (interface{}, error) {
	panic("trackingOrm: AutoTxn not used")
}

func (t *trackingOrm) TableSharding(opts ...orm.TableShardingOpt) orm.Interface {
	panic("trackingOrm: TableSharding not used")
}

func (t *trackingOrm) ModifySQLOpts(opts ...orm.ModifySQLOpt) orm.Interface {
	panic("trackingOrm: ModifySQLOpts not used")
}

type trackingDo struct {
	inner   orm.DoOrm
	counter *execCounter
}

func (d *trackingDo) Select(ctx context.Context, dest interface{}, expr string, arg map[string]interface{}) error {
	return d.inner.Select(ctx, dest, expr, arg)
}

func (d *trackingDo) Count(ctx context.Context, expr string, arg map[string]interface{}) (uint64, error) {
	return d.inner.Count(ctx, expr, arg)
}

func (d *trackingDo) Delete(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	return d.inner.Delete(ctx, expr, arg)
}

func (d *trackingDo) Update(ctx context.Context, expr string, arg map[string]interface{}) (int64, error) {
	return d.inner.Update(ctx, expr, arg)
}

func (d *trackingDo) Exec(ctx context.Context, expr string) (int64, error) {
	d.counter.execs++
	return d.inner.Exec(ctx, expr)
}

func (d *trackingDo) Insert(ctx context.Context, expr string, data interface{}) error {
	return d.inner.Insert(ctx, expr, data)
}

func (d *trackingDo) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	return d.inner.BulkInsert(ctx, expr, args)
}
