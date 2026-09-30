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
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Canonical lowercase UUIDs. register.Regist rejects any other spelling, and
// these tests must not touch register.Main or register.Obs.
const (
	migA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"
	migB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2"
	migC = "cccccccc-cccc-4ccc-8ccc-ccccccccccc3"
	migD = "dddddddd-dddd-4ddd-8ddd-ddddddddddd4"
	migE = "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee5"
	migF = "ffffffff-ffff-4fff-8fff-fffffffffff6"
)

func noopUp(_ context.Context, _ orm.Interface) error {
	return nil
}

// mustMigration registers one migration into a fresh registry and returns it
// with ParsedVersion populated. A composite literal leaves the unexported
// parsed field zero, so baseline selection would treat it as v0.0.0.
func mustMigration(t *testing.T, id, version, timestamp, description string) register.Migration {
	t.Helper()
	reg := &register.Registry{}
	reg.Regist(id, version, timestamp, description, noopUp)
	all := reg.All()
	require.Len(t, all, 1)
	return all[0]
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
	assert.Equal(t, "hcm_migration_record", RecordTable)
	assert.True(t, strings.HasPrefix(recordTableDDL, "CREATE TABLE `"+RecordTable+"`"))
	assert.NotContains(t, strings.ToUpper(recordTableDDL), "IF NOT EXISTS")
	assert.Contains(t, recordTableDDL, "`migration_id` VARCHAR(64) NOT NULL")
	assert.Contains(t, recordTableDDL, fmt.Sprintf("`version` VARCHAR(%d) NOT NULL", register.MaxVersionLen))
	assert.Contains(t, recordTableDDL, "`message` VARCHAR(1024) NOT NULL DEFAULT ''")
	assert.Contains(t, recordTableDDL, "PRIMARY KEY (`id`)")
	assert.Contains(t, recordTableDDL, "UNIQUE KEY `uidx_migration_id` (`migration_id`)")
	assert.Contains(t, recordTableDDL, "COLLATE=utf8mb4_bin")
	assert.Contains(t, recordTableDDL, "COMMENT '最近一次成功时注册的版本'")
	assert.Contains(t, recordTableDDL, "COMMENT '执行状态，取值 running/success/failed'")
}

func TestBaselineMigrations(t *testing.T) {
	at := func(id, version string) register.Migration {
		return mustMigration(t, id, version, "20260101120000", version)
	}
	pending := mustMigration(t, migF, register.PendingVersion, "20260101120000", "undecided")
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
			// Compare orders v1.9.3 < v1.9.3.1 < v1.9.3-tenant.1 < v1.9.4.
			// A v1.9.4 baseline therefore adopts the labeled line and the numeric fourth.
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
				mustMigration(t, migA, register.PendingVersion, "20260101120000", "pending first"),
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
			// The above-baseline copy is not inserted, so it must not consume the ID.
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

func TestInitRecordTable(t *testing.T) {
	kt := kit.New()
	sample := []register.Migration{
		mustMigration(t, migA, "v1.9.3", "20260101120000", "three"),
		mustMigration(t, migA, "v1.9.2", "20260102120000", "dup"),
		mustMigration(t, migB, "v1.9.3.0", "20260101120000", "folded"),
		mustMigration(t, migC, "v1.9.3.1", "20260101120000", "fourth"),
		mustMigration(t, migD, "v1.9.4", "20260101120000", "later"),
		mustMigration(t, migE, register.PendingVersion, "20260101120000", "pending"),
	}
	baseline := mustVersion(t, "v1.9.3")

	t.Run("table exists", func(t *testing.T) {
		do := newFakeDo()
		do.countDefault = 1
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Empty(t, do.callsOf("exec"))
		assert.Empty(t, do.callsOf("bulk-insert"))
		assert.Empty(t, do.callsOf("insert"))
		counts := do.callsOf("count")
		require.NotEmpty(t, counts)
		arg, ok := counts[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, RecordTable, arg["table"])
		assert.Contains(t, counts[0].expr, "information_schema.tables")
	})

	t.Run("created with nil baseline inserts nothing", func(t *testing.T) {
		do := newFakeDo()
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), nil, sample)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Nil(t, adopted)
		assert.Equal(t, []string{recordTableDDL}, execExprs(do))
		assert.Empty(t, do.callsOf("bulk-insert"))
	})

	t.Run("created with baseline bulk inserts the selection once", func(t *testing.T) {
		do := newFakeDo()
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.NoError(t, err)
		assert.True(t, created)
		require.Equal(t, []string{migA, migB}, migrationIDs(adopted))
		assert.Equal(t, []string{"v1.9.3", "v1.9.3.0"}, migrationVersions(adopted))

		bulks := do.callsOf("bulk-insert")
		require.Len(t, bulks, 1)
		assert.Contains(t, bulks[0].expr, "INSERT INTO `hcm_migration_record`")
		assert.Contains(t, bulks[0].expr, ":migration_id")
		assert.NotContains(t, bulks[0].expr, "ON DUPLICATE KEY UPDATE")
		rows, ok := bulks[0].arg.([]Record)
		require.True(t, ok)
		require.Len(t, rows, len(adopted))
		for i, m := range adopted {
			assert.Equal(t, m.ID, rows[i].MigrationID)
			assert.Equal(t, m.Version, rows[i].Version)
			assert.Equal(t, StatusSuccess, rows[i].Status)
			assert.Equal(t, "", rows[i].Message)
		}
		assert.Empty(t, do.callsOf("insert"))
		assert.Equal(t, []string{recordTableDDL}, execExprs(do))
	})

	t.Run("empty selection does not bulk insert", func(t *testing.T) {
		do := newFakeDo()
		low := mustVersion(t, "v1.0.0")
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &low, sample)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Empty(t, adopted)
		assert.Empty(t, do.callsOf("bulk-insert"))
		assert.Equal(t, []string{recordTableDDL}, execExprs(do))
	})

	t.Run("has table error", func(t *testing.T) {
		do := newFakeDo()
		do.countErr = errors.New("information_schema down")
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Empty(t, do.callsOf("exec"))
		assert.Empty(t, do.callsOf("bulk-insert"))
		assertNoSentinel(t, err)
		assert.Contains(t, err.Error(), "information_schema down")
	})

	t.Run("create exec error", func(t *testing.T) {
		do := newFakeDo()
		do.execErr = errors.New("disk full")
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Equal(t, []string{recordTableDDL}, execExprs(do))
		assert.Empty(t, do.callsOf("bulk-insert"))
		assert.Contains(t, err.Error(), "disk full")
		assert.NotContains(t, err.Error(), "must be dropped manually")
	})

	t.Run("bulk insert failure drops the new table", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {n: 1}}
		do.bulkErr = errors.New("duplicate key")
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assertNoSentinel(t, err)
		assert.Contains(t, err.Error(), "duplicate key")
		assert.NotContains(t, err.Error(), "must be dropped manually")
		assert.Equal(t, []string{recordTableDDL, "DROP TABLE `hcm_migration_record`"}, execExprs(do))
	})

	t.Run("bulk insert failure and drop exec failure", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {n: 1}}
		do.bulkErr = errors.New("duplicate key")
		do.execErr = errors.New("drop denied")
		do.execErrMatch = "DROP TABLE"
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Contains(t, err.Error(), "must be dropped manually")
		assert.Contains(t, err.Error(), RecordTable)
		assert.Contains(t, err.Error(), "duplicate key")
		assert.Contains(t, err.Error(), "drop denied")
		assert.Equal(t, []string{recordTableDDL, "DROP TABLE `hcm_migration_record`"}, execExprs(do))
	})

	t.Run("bulk insert failure and drop probe failure", func(t *testing.T) {
		do := newFakeDo()
		do.countQueue = []countReply{{n: 0}, {err: errors.New("schema down")}}
		do.bulkErr = errors.New("duplicate key")
		created, adopted, err := InitRecordTable(kt, newFakeOrm(do), &baseline, sample)
		require.Error(t, err)
		assert.False(t, created)
		assert.Nil(t, adopted)
		assert.Contains(t, err.Error(), "must be dropped manually")
		assert.Contains(t, err.Error(), "schema down")
		assert.Equal(t, []string{recordTableDDL}, execExprs(do))
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
