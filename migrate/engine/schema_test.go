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
	"errors"
	"fmt"
	"strings"
	"testing"

	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Readable migration IDs. These tests must not touch register.Main or
// register.Obs.
const (
	migA = "20260101-1200-A-0001"
	migB = "20260101-1200-B-0002"
	migC = "20260101-1200-C-0003"
	migD = "20260101-1200-D-0004"
	migE = "20260101-1200-E-0005"
	migF = "20260101-1200-F-0006"
)

func noopUp(_ context.Context, _ orm.Interface) error {
	return nil
}

// pkgTag keeps only lowercase letters and digits for a migration directory tag.
func pkgTag(raw string) string {
	raw = strings.ToLower(raw)
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == ' ' || r == '-':
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "x"
	}
	return out
}

// mustMigration builds one migration via NewMigration with ParsedVersion
// populated. A composite literal leaves the unexported parsed field zero, so
// baseline selection would treat it as v0.0.0. Engine tests cannot Regist
// because types here live under hcm/migrate/engine, not migrations/.
func mustMigration(t *testing.T, id, version, timestamp, tag string) register.Migration {
	t.Helper()
	safe := pkgTag(tag)
	var pkgPath string
	if register.IsPending(version) {
		pkgPath = fmt.Sprintf("%smain/pending/%s_%s", constant.MigrationPkgPrefix, timestamp, safe)
	} else {
		pkgPath = fmt.Sprintf("%smain/%s/%s_%s_%s", constant.MigrationPkgPrefix, version, version, timestamp, safe)
	}
	m, err := register.NewMigration("main", pkgPath, id, version, timestamp, noopUp)
	require.NoError(t, err)
	return m
}

func mustVersion(t *testing.T, raw string) register.Version {
	t.Helper()
	v, err := register.Parse(raw)
	require.NoError(t, err)
	return v
}

// assertNoSentinel reports that err is a plain failure: it wraps none of the
// exit-code sentinels.
func assertNoSentinel(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrUsage)
	assert.NotErrorIs(t, err, ErrPrecondition)
	assert.NotErrorIs(t, err, ErrMissed)
}

func TestRecordTableDDL(t *testing.T) {
	assert.Equal(t, "hcm_migration_record", constant.MigrationRecordTable)
	assert.True(t, strings.HasPrefix(recordTableDDL, "CREATE TABLE `"+constant.MigrationRecordTable+"`"))
	assert.NotContains(t, strings.ToUpper(recordTableDDL), "IF NOT EXISTS")
	assert.Contains(t, recordTableDDL, fmt.Sprintf("`migration_id` VARCHAR(%d) NOT NULL", constant.MigrationIDMaxLen))
	assert.Contains(t, recordTableDDL, "`migration_id` VARCHAR(64) NOT NULL")
	assert.Contains(t, recordTableDDL, fmt.Sprintf("`version` VARCHAR(%d) NOT NULL", constant.MigrationVersionMaxLen))
	assert.Contains(t, recordTableDDL, fmt.Sprintf("`applied_pkg` VARCHAR(%d) NOT NULL DEFAULT ''", constant.MigrationPkgMaxLen))
	assert.Contains(t, recordTableDDL, "`message` VARCHAR(1024) NOT NULL DEFAULT ''")
	assert.Contains(t, recordTableDDL, "PRIMARY KEY (`id`)")
	assert.Contains(t, recordTableDDL, "UNIQUE KEY `uidx_migration_id` (`migration_id`)")
	assert.Contains(t, recordTableDDL, "COLLATE=utf8mb4_bin")
	assert.Contains(t, recordTableDDL, "COMMENT '最近一次成功时注册的版本'")
	assert.Contains(t, recordTableDDL, "COMMENT '执行该迁移 ID 的 Go 包'")
	assert.Contains(t, recordTableDDL, "COMMENT '执行状态，取值 running/success/failed'")
}

func TestAuditTableDDL(t *testing.T) {
	assert.Equal(t, "hcm_migration_audit", constant.MigrationAuditTable)
	assert.True(t, strings.HasPrefix(auditTableDDL, "CREATE TABLE `"+constant.MigrationAuditTable+"`"))
	assert.NotContains(t, strings.ToUpper(auditTableDDL), "IF NOT EXISTS")
	assert.Contains(t, auditTableDDL, "UNIQUE KEY `uidx_run_id` (`run_id`)")
	assert.Contains(t, auditTableDDL, "COLLATE=utf8mb4_bin")
}

func TestBaselineMigrations(t *testing.T) {
	at := func(id, version string) register.Migration {
		return mustMigration(t, id, version, "20260101120000", version)
	}
	pending := mustMigration(t, migF, constant.MigrationPendingVersion, "20260101120000", "undecided")
	// Same ID, later version listed first, so dedupe must keep input order
	// rather than the smaller version.
	dupHigh := mustMigration(t, migA, "v1.9.3", "20260201120000", "dup high")
	dupLow := mustMigration(t, migA, "v1.9.2", "20260101120000", "dup low")
	above := mustMigration(t, migB, "v1.9.10", "20260101120000", "above")
	belowSameID := mustMigration(t, migB, "v1.9.3", "20260301120000", "below same id")

	testCases := []struct {
		name     string
		baseline string
		in       []register.Migration
		wantID   []string
		wantVer  []string
	}{
		{
			name:     "empty input",
			baseline: "v1.9.3",
		},
		{
			name:     "nil input",
			baseline: "v1.9.3",
			in:       nil,
		},
		{
			name:     "equal three segments included",
			baseline: "v1.9.3",
			in:       []register.Migration{at(migA, "v1.9.3")},
			wantID:   []string{migA},
			wantVer:  []string{"v1.9.3"},
		},
		{
			name:     "folded zero equals baseline and keeps raw spelling",
			baseline: "v1.9.3",
			in:       []register.Migration{at(migA, "v1.9.3.0")},
			wantID:   []string{migA},
			wantVer:  []string{"v1.9.3.0"},
		},
		{
			name:     "numeric fourth above three segment baseline",
			baseline: "v1.9.3",
			in:       []register.Migration{at(migA, "v1.9.3"), at(migB, "v1.9.3.1")},
			wantID:   []string{migA},
			wantVer:  []string{"v1.9.3"},
		},
		{
			name:     "labeled fourth above three segment baseline",
			baseline: "v1.9.3",
			in:       []register.Migration{at(migA, "v1.9.3"), at(migB, "v1.9.3-tenant.1")},
			wantID:   []string{migA},
			wantVer:  []string{"v1.9.3"},
		},
		{
			name:     "feature label and numeric fourth below next release",
			baseline: "v1.9.4",
			in: []register.Migration{
				at(migA, "v1.9.3"),
				at(migB, "v1.9.3.1"),
				at(migC, "v1.9.3-tenant.1"),
				at(migD, "v1.9.4"),
				at(migE, "v1.9.10"),
			},
			wantID:  []string{migA, migB, migC, migD},
			wantVer: []string{"v1.9.3", "v1.9.3.1", "v1.9.3-tenant.1", "v1.9.4"},
		},
		{
			name:     "pending skipped",
			baseline: "v1.9.10",
			in:       []register.Migration{pending, at(migA, "v1.9.3")},
			wantID:   []string{migA},
			wantVer:  []string{"v1.9.3"},
		},
		{
			name:     "pending then same id real version keeps the real one",
			baseline: "v1.9.4",
			in: []register.Migration{
				mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "pending first"),
				mustMigration(t, migA, "v1.9.3", "20260201120000", "real second"),
			},
			wantID:  []string{migA},
			wantVer: []string{"v1.9.3"},
		},
		{
			name:     "duplicate id keeps the first in input order",
			baseline: "v1.9.4",
			in:       []register.Migration{dupHigh, dupLow, at(migC, "v1.9.2")},
			wantID:   []string{migA, migC},
			wantVer:  []string{"v1.9.3", "v1.9.2"},
		},
		{
			name:     "above-baseline duplicate does not hide a later eligible copy",
			baseline: "v1.9.4",
			in:       []register.Migration{above, belowSameID},
			wantID:   []string{migB},
			wantVer:  []string{"v1.9.3"},
		},
		{
			name:     "everything above baseline",
			baseline: "v1.0.0",
			in:       []register.Migration{at(migA, "v1.9.3"), pending},
		},
		{
			name:     "numeric not lexical",
			baseline: "v1.9.9",
			in:       []register.Migration{at(migA, "v1.9.9"), at(migB, "v1.9.10")},
			wantID:   []string{migA},
			wantVer:  []string{"v1.9.9"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := baselineMigrations(mustVersion(t, tc.baseline), tc.in)
			require.Len(t, got, len(tc.wantID))
			for i := range tc.wantID {
				assert.Equal(t, tc.wantID[i], got[i].ID)
				assert.Equal(t, tc.wantVer[i], got[i].Version)
			}
		})
	}
}

func TestInsertBaselineEmpty(t *testing.T) {
	do := newFakeDo()
	kt := kit.New()
	require.NoError(t, insertBaseline(kt, newFakeOrm(do), nil))
	require.NoError(t, insertBaseline(kt, newFakeOrm(do), []register.Migration{}))
	assert.Empty(t, do.callsOf("bulk-insert"))
}

func TestInitTables(t *testing.T) {
	kt := kit.New()
	sample := []register.Migration{
		mustMigration(t, migA, "v1.9.3", "20260101120000", "three"),
		mustMigration(t, migA, "v1.9.2", "20260102120000", "dup"),
		mustMigration(t, migB, "v1.9.3.0", "20260101120000", "folded"),
		mustMigration(t, migC, "v1.9.3.1", "20260101120000", "fourth"),
		mustMigration(t, migD, "v1.9.4", "20260101120000", "later"),
		mustMigration(t, migE, constant.MigrationPendingVersion, "20260101120000", "pending"),
	}
	baseline := mustVersion(t, "v1.9.3")

	countTables := func(do *fakeDo) []string {
		out := make([]string, 0)
		for _, c := range do.callsOf("count") {
			arg, ok := c.arg.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := arg["table"].(string); ok {
				out = append(out, name)
			}
		}
		return out
	}

	t.Run("neither exists empty mode creates both without baseline", func(t *testing.T) {
		do := newFakeDo()
		result, err := InitTables(kt, newFakeOrm(do), nil, sample)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Equal(t, []string{auditTableDDL, recordTableDDL}, execExprs(do))
		assert.Empty(t, do.callsOf("bulk-insert"))
		assert.Equal(t, []string{constant.MigrationAuditTable, constant.MigrationRecordTable}, countTables(do))
	})

	t.Run("neither exists adopt mode inserts baseline with applied_pkg", func(t *testing.T) {
		do := newFakeDo()
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		require.Equal(t, []string{migA, migB}, migrationIDs(result.Adopted))
		assert.Equal(t, []string{"v1.9.3", "v1.9.3.0"}, migrationVersions(result.Adopted))

		bulks := do.callsOf("bulk-insert")
		require.Len(t, bulks, 1)
		assert.Contains(t, bulks[0].expr, "INSERT INTO `hcm_migration_record`")
		assert.NotContains(t, bulks[0].expr, "ON DUPLICATE KEY UPDATE")
		rows, ok := bulks[0].arg.([]Record)
		require.True(t, ok)
		require.Len(t, rows, len(result.Adopted))
		for i, m := range result.Adopted {
			assert.Equal(t, m.ID, rows[i].MigrationID)
			assert.Equal(t, m.Version, rows[i].Version)
			assert.Equal(t, m.Pkg, rows[i].AppliedPkg)
			assert.Equal(t, enumor.MigrationStatusSuccess, rows[i].Status)
			assert.Equal(t, "", rows[i].Message)
		}
		assert.Equal(t, []string{auditTableDDL, recordTableDDL}, execExprs(do))
	})

	t.Run("both exist no ddl no baseline even in adopt", func(t *testing.T) {
		do := newFakeDo()
		do.countDefault = 1
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.NoError(t, err)
		assert.False(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Empty(t, do.callsOf("exec"))
		assert.Empty(t, do.callsOf("bulk-insert"))
		assert.Equal(t, []string{constant.MigrationAuditTable, constant.MigrationRecordTable}, countTables(do))
	})

	t.Run("record exists audit missing creates only audit no baseline", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {n: 1}}
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Equal(t, []string{auditTableDDL}, execExprs(do))
		assert.Empty(t, do.callsOf("bulk-insert"))
	})

	t.Run("audit exists record missing creates record and applies baseline", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 1}, {n: 0}}
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.NoError(t, err)
		assert.False(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		require.Equal(t, []string{migA, migB}, migrationIDs(result.Adopted))
		assert.Equal(t, []string{recordTableDDL}, execExprs(do))
		require.Len(t, do.callsOf("bulk-insert"), 1)
	})

	t.Run("audit create error never checks record table", func(t *testing.T) {
		do := newFakeDo()
		do.countErr = errors.New("audit schema down")
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.Equal(t, InitResult{}, result)
		assert.Contains(t, err.Error(), "audit schema down")
		assert.Empty(t, do.callsOf("exec"))
		assert.Equal(t, []string{constant.MigrationAuditTable}, countTables(do))
		assertNoSentinel(t, err)
	})

	t.Run("audit create exec error never checks record table", func(t *testing.T) {
		do := newFakeDo()
		do.execErr = errors.New("audit disk full")
		do.execErrMatch = constant.MigrationAuditTable
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.Equal(t, InitResult{}, result)
		assert.Contains(t, err.Error(), "audit disk full")
		assert.Equal(t, []string{auditTableDDL}, execExprs(do))
		assert.Equal(t, []string{constant.MigrationAuditTable}, countTables(do))
	})

	t.Run("record create error still reports AuditCreated", func(t *testing.T) {
		do := newFakeDo()
		do.execErr = errors.New("record disk full")
		do.execErrMatch = constant.MigrationRecordTable
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.True(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Contains(t, err.Error(), "record disk full")
		assert.Equal(t, []string{auditTableDDL, recordTableDDL}, execExprs(do))
		assert.Empty(t, do.callsOf("bulk-insert"))
	})

	t.Run("baseline insert failure drops record keeps audit", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {n: 0}, {n: 1}}
		do.bulkErr = errors.New("duplicate key")
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.True(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Contains(t, err.Error(), "duplicate key")
		assert.NotContains(t, err.Error(), "must be dropped manually")
		assert.Equal(t, []string{auditTableDDL, recordTableDDL, "DROP TABLE `hcm_migration_record`"}, execExprs(do))
		for _, expr := range execExprs(do) {
			assert.NotContains(t, expr, "DROP TABLE `hcm_migration_audit`")
		}
	})

	t.Run("baseline insert failure and drop exec failure", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {n: 0}, {n: 1}}
		do.bulkErr = errors.New("duplicate key")
		do.execErr = errors.New("drop denied")
		do.execErrMatch = "DROP TABLE"
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.True(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Contains(t, err.Error(), "must be dropped manually")
		assert.Contains(t, err.Error(), constant.MigrationRecordTable)
		assert.Contains(t, err.Error(), "duplicate key")
		assert.Contains(t, err.Error(), "drop denied")
		assert.Equal(t, []string{auditTableDDL, recordTableDDL, "DROP TABLE `hcm_migration_record`"}, execExprs(do))
	})

	t.Run("baseline insert failure and drop probe failure", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {n: 0}, {err: errors.New("schema down")}}
		do.bulkErr = errors.New("duplicate key")
		result, err := InitTables(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.True(t, result.AuditCreated)
		assert.False(t, result.RecordCreated)
		assert.Nil(t, result.Adopted)
		assert.Contains(t, err.Error(), "must be dropped manually")
		assert.Contains(t, err.Error(), "schema down")
		assert.Equal(t, []string{auditTableDDL, recordTableDDL}, execExprs(do))
	})

	t.Run("empty selection does not bulk insert", func(t *testing.T) {
		do := newFakeDo()
		low := mustVersion(t, "v1.0.0")
		result, err := InitTables(kt, newFakeOrm(do), &low, sample)
		require.NoError(t, err)
		assert.True(t, result.AuditCreated)
		assert.True(t, result.RecordCreated)
		assert.Empty(t, result.Adopted)
		assert.Empty(t, do.callsOf("bulk-insert"))
		assert.Equal(t, []string{auditTableDDL, recordTableDDL}, execExprs(do))
	})
}

func execExprs(do *fakeDo) []string {
	calls := do.callsOf("exec")
	out := make([]string, 0, len(calls))
	for _, c := range calls {
		out = append(out, c.expr)
	}
	return out
}

func migrationIDs(ms []register.Migration) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.ID)
	}
	return out
}

func migrationVersions(ms []register.Migration) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.Version)
	}
	return out
}
