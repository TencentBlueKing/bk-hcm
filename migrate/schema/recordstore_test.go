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

package schema

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordDBTags(t *testing.T) {
	assert.Equal(t, enumor.MigrationStatus("running"), enumor.MigrationStatusRunning)
	assert.Equal(t, enumor.MigrationStatus("success"), enumor.MigrationStatusSuccess)
	assert.Equal(t, enumor.MigrationStatus("failed"), enumor.MigrationStatusFailed)
	assert.Equal(t, 1024, constant.MigrationRecordMessageMaxBytes)

	want := []struct {
		field string
		tag   string
	}{
		{field: "ID", tag: "id"},
		{field: "MigrationID", tag: "migration_id"},
		{field: "Version", tag: "version"},
		{field: "AppliedPkg", tag: "applied_pkg"},
		{field: "Status", tag: "status"},
		{field: "Message", tag: "message"},
		{field: "CreatedAt", tag: "created_at"},
		{field: "UpdatedAt", tag: "updated_at"},
	}
	typ := reflect.TypeOf(Record{})
	for _, tc := range want {
		t.Run(tc.field, func(t *testing.T) {
			field, ok := typ.FieldByName(tc.field)
			require.True(t, ok)
			assert.Equal(t, tc.tag, field.Tag.Get("db"))
		})
	}
}

func TestRecordStore_Load(t *testing.T) {
	kt := kit.New()

	t.Run("select error", func(t *testing.T) {
		do := newFakeDo()
		do.selectErr = errors.New("conn reset")
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assertNoSentinel(t, err)
		assert.Contains(t, err.Error(), "conn reset")
		assert.Contains(t, err.Error(), "list migration records failed")
	})

	t.Run("empty table", func(t *testing.T) {
		do := newFakeDo()
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.NotNil(t, records)
		assert.Empty(t, records)
		selects := do.callsOf("select")
		require.Len(t, selects, 1)
		assert.Contains(t, selects[0].expr, "FROM `hcm_migration_record`")
		assert.Contains(t, selects[0].expr, "`applied_pkg`")
	})

	t.Run("indexes known statuses by migration id", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{ID: 1, MigrationID: migA, Version: "v1.9.3", Status: enumor.MigrationStatusRunning},
			{ID: 2, MigrationID: migB, Version: "v1.9.4", AppliedPkg: "main/pending/20260101120000_ok",
				Status: enumor.MigrationStatusSuccess, Message: "ok"},
			{ID: 3, MigrationID: migC, Version: "v1.9.5", Status: enumor.MigrationStatusFailed, Message: "boom"},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.Len(t, records, 3)
		assert.Equal(t, do.selectRows[0], records[migA])
		assert.Equal(t, do.selectRows[1], records[migB])
		assert.Equal(t, enumor.MigrationStatusFailed, records[migC].Status)
		assert.Equal(t, "boom", records[migC].Message)
	})

	t.Run("success with empty applied_pkg is precondition", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{MigrationID: migA, Version: "v1.9.3", AppliedPkg: "", Status: enumor.MigrationStatusSuccess},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), migA)
		assert.Contains(t, err.Error(), "empty applied_pkg")
	})

	t.Run("running and failed with empty applied_pkg are ok", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{MigrationID: migA, Version: "v1.9.3", AppliedPkg: "", Status: enumor.MigrationStatusRunning},
			{MigrationID: migB, Version: "v1.9.4", AppliedPkg: "", Status: enumor.MigrationStatusFailed},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.Len(t, records, 2)
	})

	testCases := []struct {
		name   string
		status enumor.MigrationStatus
	}{
		{name: "empty status", status: ""},
		{name: "uppercase success", status: "SUCCESS"},
		{name: "done", status: "done"},
		{name: "mixed case running", status: "Running"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := newFakeDo()
			do.selectRows = []Record{
				{MigrationID: migA, Version: "v1.9.3", AppliedPkg: "main/pending/20260101120000_a",
					Status: enumor.MigrationStatusSuccess},
				{MigrationID: migB, Version: "v1.9.4", Status: tc.status},
				{MigrationID: migC, Version: "v1.9.5", Status: enumor.MigrationStatusFailed},
			}
			records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
			require.Error(t, err)
			assert.Nil(t, records)
			assert.ErrorIs(t, err, migrate.ErrPrecondition)
			assert.Contains(t, err.Error(), migB)
			assert.Contains(t, err.Error(), string(tc.status))
		})
	}

	// Exact "PENDING" is accepted; only case/space variants stay unparsable.
	badVersions := []string{"1.9.3", "", "v1.9", "v1.9.3-Tenant.1", "v01.9.3",
		"pending", "Pending", " PENDING", "PENDING "}
	statuses := []enumor.MigrationStatus{
		enumor.MigrationStatusRunning, enumor.MigrationStatusFailed, enumor.MigrationStatusSuccess,
	}
	for _, status := range statuses {
		for _, bad := range badVersions {
			t.Run(fmt.Sprintf("unparsable %s version %q", status, bad), func(t *testing.T) {
				do := newFakeDo()
				do.selectRows = []Record{
					{MigrationID: migA, Version: "v1.9.3", AppliedPkg: "main/pending/20260101120000_a",
						Status: enumor.MigrationStatusSuccess},
					{MigrationID: migB, Version: bad, AppliedPkg: "main/pending/20260101120000_b", Status: status},
				}
				records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
				require.Error(t, err)
				assert.Nil(t, records)
				assert.ErrorIs(t, err, migrate.ErrPrecondition)
				assert.Contains(t, err.Error(), migB)
				assert.Contains(t, err.Error(), "unparsable version")
				if bad == "" {
					assert.Contains(t, err.Error(), `""`)
				} else {
					assert.Contains(t, err.Error(), bad)
				}
			})
		}
	}

}

func TestRecordStore_LoadVersions(t *testing.T) {
	kt := kit.New()

	t.Run("unknown status is reported before an unparsable version", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{MigrationID: migA, Version: "v1.9.3", AppliedPkg: "main/pending/20260101120000_a",
				Status: enumor.MigrationStatusSuccess},
			{MigrationID: migB, Version: "PENDING", Status: "done"},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), migB)
		assert.Contains(t, err.Error(), "unknown status")
		assert.Contains(t, err.Error(), "done")
		assert.NotContains(t, err.Error(), "unparsable version")
	})

	t.Run("accepts PENDING on every known status", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{ID: 1, MigrationID: migA, Version: constant.MigrationPendingVersion,
				Status: enumor.MigrationStatusRunning},
			{ID: 2, MigrationID: migB, Version: constant.MigrationPendingVersion,
				AppliedPkg: "main/pending/20260101120000_b", Status: enumor.MigrationStatusSuccess},
			{ID: 3, MigrationID: migC, Version: constant.MigrationPendingVersion,
				Status: enumor.MigrationStatusFailed, Message: "boom"},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.Len(t, records, 3)
		assert.Equal(t, do.selectRows[0], records[migA])
		assert.Equal(t, do.selectRows[1], records[migB])
		assert.Equal(t, do.selectRows[2], records[migC])
	})

	t.Run("PENDING success with empty applied_pkg is precondition", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{MigrationID: migA, Version: constant.MigrationPendingVersion, AppliedPkg: "",
				Status: enumor.MigrationStatusSuccess},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), migA)
		assert.Contains(t, err.Error(), "empty applied_pkg")
	})

	t.Run("mix of PENDING and versioned rows loads fine", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{ID: 1, MigrationID: migA, Version: "v1.9.2", AppliedPkg: "main/pending/20260101120000_a",
				Status: enumor.MigrationStatusSuccess},
			{ID: 2, MigrationID: migB, Version: constant.MigrationPendingVersion,
				AppliedPkg: "main/pending/20260101120000_b", Status: enumor.MigrationStatusSuccess},
			{ID: 3, MigrationID: migC, Version: constant.MigrationPendingVersion,
				Status: enumor.MigrationStatusRunning},
			{ID: 4, MigrationID: migD, Version: "v1.9.3", Status: enumor.MigrationStatusFailed},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.Len(t, records, 4)
		assert.Equal(t, do.selectRows[0], records[migA])
		assert.Equal(t, do.selectRows[1], records[migB])
		assert.Equal(t, do.selectRows[2], records[migC])
		assert.Equal(t, do.selectRows[3], records[migD])
	})

	t.Run("indexes labeled and folded versions of every status", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{ID: 1, MigrationID: migA, Version: "v1.9.3-tenant.1", Status: enumor.MigrationStatusRunning},
			{ID: 2, MigrationID: migB, Version: "v1.9.3.0", AppliedPkg: "main/pending/20260101120000_b",
				Status: enumor.MigrationStatusSuccess},
			{ID: 3, MigrationID: migC, Version: "v1.9.4", Status: enumor.MigrationStatusFailed},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.Len(t, records, 3)
		assert.Equal(t, do.selectRows[0], records[migA])
		assert.Equal(t, do.selectRows[1], records[migB])
		assert.Equal(t, do.selectRows[2], records[migC])
	})
}

func TestRecordStore_MarkRunning(t *testing.T) {
	kt := kit.New()
	m := register.Migration{ID: migA, Version: "v1.9.3", Pkg: "main/pending/20260101120000_add_x"}

	t.Run("insert upsert", func(t *testing.T) {
		do := newFakeDo()
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkRunning(kt, m))
		inserts := do.callsOf("insert")
		require.Len(t, inserts, 1)
		assert.Contains(t, inserts[0].expr, "ON DUPLICATE KEY UPDATE")
		assert.Contains(t, inserts[0].expr, "`status` = :status")
		assert.Contains(t, inserts[0].expr, "`message` = ''")
		assert.Contains(t, inserts[0].expr, "`applied_pkg`")
		assert.Contains(t, inserts[0].expr, ":applied_pkg")
		assert.Contains(t, inserts[0].expr, "INSERT INTO `hcm_migration_record`")
		assert.Contains(t, inserts[0].expr, "ON DUPLICATE KEY UPDATE `applied_pkg` = :applied_pkg")
		arg, ok := inserts[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, m.ID, arg["migration_id"])
		assert.Equal(t, m.Version, arg["version"])
		assert.Equal(t, m.Pkg, arg["applied_pkg"])
		assert.Equal(t, enumor.MigrationStatusRunning, arg["status"])
		assert.NotContains(t, arg, "message")
	})

	t.Run("insert error", func(t *testing.T) {
		do := newFakeDo()
		do.insertErr = errors.New("duplicate")
		err := NewRecordStore(newFakeOrm(do)).MarkRunning(kt, m)
		require.Error(t, err)
		assertNoSentinel(t, err)
		assert.Contains(t, err.Error(), m.ID)
		assert.Contains(t, err.Error(), "duplicate")
		assert.Contains(t, err.Error(), "running")
	})
}

func TestRecordStore_MarkSuccess(t *testing.T) {
	kt := kit.New()
	m := register.Migration{ID: migA, Version: "v1.9.4", Pkg: "main/pending/20260101120000_add_x"}

	t.Run("updates status and registered version", func(t *testing.T) {
		do := newFakeDo()
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkSuccess(kt, m))
		updates := do.callsOf("update")
		require.Len(t, updates, 1)
		assert.Contains(t, updates[0].expr, "`status` = :status")
		assert.Contains(t, updates[0].expr, "`message` = ''")
		assert.Contains(t, updates[0].expr, "`version` = :version")
		assert.Contains(t, updates[0].expr, "WHERE `migration_id` = :migration_id")
		assert.NotContains(t, updates[0].expr, "applied_pkg")
		arg, ok := updates[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, m.ID, arg["migration_id"])
		assert.Equal(t, m.Version, arg["version"])
		assert.Equal(t, enumor.MigrationStatusSuccess, arg["status"])
		assert.NotContains(t, arg, "applied_pkg")
	})

	t.Run("zero affected rows", func(t *testing.T) {
		do := newFakeDo()
		do.updateAffected = 0
		err := NewRecordStore(newFakeOrm(do)).MarkSuccess(kt, m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
		assert.Contains(t, err.Error(), m.ID)
		assert.Contains(t, err.Error(), "success")
		assertNoSentinel(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		do := newFakeDo()
		do.updateErr = errors.New("deadlock")
		err := NewRecordStore(newFakeOrm(do)).MarkSuccess(kt, m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deadlock")
		assert.Contains(t, err.Error(), m.ID)
		assert.NotContains(t, err.Error(), "record not found")
		assertNoSentinel(t, err)
	})
}

func TestRecordStore_MarkFailed(t *testing.T) {
	kt := kit.New()
	m := register.Migration{ID: migB, Version: "v1.9.3", Pkg: "main/pending/20260101120000_add_x"}

	t.Run("nil error stores empty message", func(t *testing.T) {
		do := newFakeDo()
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkFailed(kt, m, nil))
		arg := updateArg(t, do)
		assert.Equal(t, "", arg["message"])
		assert.Equal(t, enumor.MigrationStatusFailed, arg["status"])
		assert.Equal(t, m.ID, arg["migration_id"])
		_, hasVersion := arg["version"]
		assert.False(t, hasVersion)
		expr := do.callsOf("update")[0].expr
		assert.Contains(t, expr, "`message` = :message")
		assert.NotContains(t, expr, "`version`")
		assert.NotContains(t, expr, "applied_pkg")
	})

	t.Run("truncates message", func(t *testing.T) {
		do := newFakeDo()
		raw := strings.Repeat("迁移失败，磁盘已满。", 80)
		require.Greater(t, len(raw), constant.MigrationRecordMessageMaxBytes)
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkFailed(kt, m, errors.New(raw)))
		arg := updateArg(t, do)
		msg, ok := arg["message"].(string)
		require.True(t, ok)
		assert.LessOrEqual(t, len(msg), constant.MigrationRecordMessageMaxBytes)
		assert.True(t, utf8.ValidString(msg))
		assert.True(t, strings.HasPrefix(raw, msg))
		assert.NotEqual(t, raw, msg)
		assert.Equal(t, migrate.TruncateUTF8(raw, constant.MigrationRecordMessageMaxBytes), msg)
	})

	t.Run("zero affected rows", func(t *testing.T) {
		do := newFakeDo()
		do.updateAffected = 0
		err := NewRecordStore(newFakeOrm(do)).MarkFailed(kt, m, errors.New("boom"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
		assert.Contains(t, err.Error(), m.ID)
		assert.Contains(t, err.Error(), "failed")
	})

	t.Run("update error", func(t *testing.T) {
		do := newFakeDo()
		do.updateErr = errors.New("timeout")
		err := NewRecordStore(newFakeOrm(do)).MarkFailed(kt, m, errors.New("boom"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Contains(t, err.Error(), m.ID)
		assert.NotContains(t, err.Error(), "record not found")
	})
}

func updateArg(t *testing.T, do *fakeDo) map[string]interface{} {
	t.Helper()
	updates := do.callsOf("update")
	require.Len(t, updates, 1)
	arg, ok := updates[0].arg.(map[string]interface{})
	require.True(t, ok)
	return arg
}

func TestCurrentVersion(t *testing.T) {
	success := func(versions ...string) Records {
		out := make(Records, len(versions))
		for i, version := range versions {
			id := fmt.Sprintf("id-%d", i)
			out[id] = Record{MigrationID: id, Version: version, Status: enumor.MigrationStatusSuccess}
		}
		return out
	}

	testCases := []struct {
		name     string
		records  Records
		wantRaw  string
		wantOK   bool
		wantExit int
	}{
		{name: "nil map", records: nil},
		{name: "no records", records: Records{}},
		{
			name: "only running and failed",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusRunning, Version: "v1.9.10"},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusFailed, Version: "v1.9.9"},
			},
		},
		{
			name:    "numeric not lexical",
			records: success("v1.9.9", "v1.9.10"),
			wantRaw: "v1.9.10",
			wantOK:  true,
		},
		{
			name: "backfill does not lower current",
			records: Records{
				"current":  {MigrationID: "current", Status: enumor.MigrationStatusSuccess, Version: "v1.9.4"},
				"backfill": {MigrationID: "backfill", Status: enumor.MigrationStatusSuccess, Version: "v1.9.3"},
				"running":  {MigrationID: "running", Status: enumor.MigrationStatusRunning, Version: "v9.9.9"},
			},
			wantRaw: "v1.9.4",
			wantOK:  true,
		},
		{
			// register.Compare orders v1.9.3 < v1.9.3.1 < v1.9.3-tenant.1.
			// The empty label of a numeric fourth segment sorts before a feature label.
			name:    "feature label above numeric fourth",
			records: success("v1.9.3", "v1.9.3-tenant.1", "v1.9.3.1"),
			wantRaw: "v1.9.3-tenant.1",
			wantOK:  true,
		},
		{
			name:    "next release beats a feature label",
			records: success("v1.9.3-tenant.9", "v1.9.4"),
			wantRaw: "v1.9.4",
			wantOK:  true,
		},
		{
			name:    "folded spelling alone keeps its raw text",
			records: success("v1.9.3.0"),
			wantRaw: "v1.9.3.0",
			wantOK:  true,
		},
		{
			// CurrentVersion still skips non-success rows. Load does not: it
			// rejects an unparsable version on any status.
			name: "CurrentVersion ignores unparsable versions on failed and running",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusFailed, Version: "PENDING"},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusRunning, Version: ""},
				migC: {MigrationID: migC, Status: enumor.MigrationStatusRunning, Version: "1.9.3"},
				migD: {MigrationID: migD, Status: enumor.MigrationStatusSuccess, Version: "v1.9.3"},
			},
			wantRaw: "v1.9.3",
			wantOK:  true,
		},
		{
			name: "CurrentVersion ignores only unparsable non-success records",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusFailed, Version: "nope"},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusRunning, Version: "v1.9"},
			},
		},
		{
			name: "success PENDING skipped beside released success",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusSuccess, Version: "v1.9.2"},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusSuccess,
					Version: constant.MigrationPendingVersion},
			},
			wantRaw: "v1.9.2",
			wantOK:  true,
		},
		{
			name: "only PENDING success yields no current",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusSuccess,
					Version: constant.MigrationPendingVersion},
			},
		},
		{
			name: "PENDING on failed and running is ignored",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusFailed,
					Version: constant.MigrationPendingVersion},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusRunning,
					Version: constant.MigrationPendingVersion},
				migC: {MigrationID: migC, Status: enumor.MigrationStatusSuccess, Version: "v1.9.2"},
			},
			wantRaw: "v1.9.2",
			wantOK:  true,
		},
		{
			name: "PENDING success skipped among higher released versions",
			records: Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusSuccess,
					Version: constant.MigrationPendingVersion},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusSuccess, Version: "v1.9.10"},
				migC: {MigrationID: migC, Status: enumor.MigrationStatusSuccess, Version: "v1.9.9"},
			},
			wantRaw: "v1.9.10",
			wantOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			current, ok, err := CurrentVersion(tc.records)
			require.NoError(t, err)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.wantRaw, current.Raw)
				return
			}
			assert.Equal(t, register.Version{}, current)
		})
	}

	badVersions := []string{"1.9.3", "", "v1.9"}
	for _, bad := range badVersions {
		t.Run("unparsable success "+bad, func(t *testing.T) {
			records := Records{
				migA: {MigrationID: migA, Status: enumor.MigrationStatusSuccess, Version: "v1.9.10"},
				migB: {MigrationID: migB, Status: enumor.MigrationStatusSuccess, Version: bad},
				migC: {MigrationID: migC, Status: enumor.MigrationStatusFailed, Version: "v1.9.4"},
			}
			current, ok, err := CurrentVersion(records)
			require.Error(t, err)
			assert.False(t, ok)
			assert.Equal(t, register.Version{}, current)
			assert.ErrorIs(t, err, migrate.ErrPrecondition)
			assert.Contains(t, err.Error(), migB)
		})
	}
}

func TestCurrentVersionEqualRawIsStable(t *testing.T) {
	for i := 0; i < 50; i++ {
		records := Records{
			fmt.Sprintf("folded-%d", i): {
				MigrationID: migA, Status: enumor.MigrationStatusSuccess, Version: "v1.9.3.0",
			},
			fmt.Sprintf("plain-%d", i): {
				MigrationID: migB, Status: enumor.MigrationStatusSuccess, Version: "v1.9.3",
			},
			fmt.Sprintf("older-%d", i): {MigrationID: migC, Status: enumor.MigrationStatusSuccess, Version: "v1.9.2"},
		}
		current, ok, err := CurrentVersion(records)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "v1.9.3", current.Raw)
	}
}

func TestRecordStore_BackfillPendingVersion(t *testing.T) {
	kt := kit.New()

	t.Run("updates only PENDING success row version", func(t *testing.T) {
		do := newFakeDo()
		m := register.Migration{ID: migA, Version: "v1.9.3", Pkg: "main/pending/20260101120000_a"}
		require.NoError(t, NewRecordStore(newFakeOrm(do)).BackfillPendingVersion(kt, m))
		updates := do.callsOf("update")
		require.Len(t, updates, 1)
		expr := updates[0].expr
		assert.Contains(t, expr, "UPDATE `hcm_migration_record` SET `version` = :version")
		assert.Contains(t, expr, "WHERE `migration_id` = :migration_id AND `status` = :status AND `version` = :pending")
		assert.NotContains(t, expr, "applied_pkg")
		assert.NotContains(t, expr, "SET `status`")
		arg, ok := updates[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, map[string]interface{}{
			"migration_id": m.ID,
			"version":      m.Version,
			"status":       enumor.MigrationStatusSuccess,
			"pending":      constant.MigrationPendingVersion,
		}, arg)
	})

	t.Run("zero affected rows is nil", func(t *testing.T) {
		do := newFakeDo()
		do.updateAffected = 0
		m := register.Migration{ID: migA, Version: "v1.9.3"}
		require.NoError(t, NewRecordStore(newFakeOrm(do)).BackfillPendingVersion(kt, m))
		require.Len(t, do.callsOf("update"), 1)
	})

	t.Run("one affected row is nil", func(t *testing.T) {
		do := newFakeDo()
		do.updateAffected = 1
		m := register.Migration{ID: migA, Version: "v1.9.3"}
		require.NoError(t, NewRecordStore(newFakeOrm(do)).BackfillPendingVersion(kt, m))
		require.Len(t, do.callsOf("update"), 1)
	})

	t.Run("update error wraps id version and cause", func(t *testing.T) {
		do := newFakeDo()
		do.updateErr = errors.New("deadlock")
		m := register.Migration{ID: migA, Version: "v1.9.3"}
		err := NewRecordStore(newFakeOrm(do)).BackfillPendingVersion(kt, m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), m.ID)
		assert.Contains(t, err.Error(), m.Version)
		assert.Contains(t, err.Error(), "deadlock")
		assertNoSentinel(t, err)
	})

	t.Run("PENDING registered version issues no SQL", func(t *testing.T) {
		do := newFakeDo()
		m := register.Migration{ID: migA, Version: constant.MigrationPendingVersion}
		err := NewRecordStore(newFakeOrm(do)).BackfillPendingVersion(kt, m)
		require.Error(t, err)
		assert.Contains(t, err.Error(), m.ID)
		assert.Contains(t, err.Error(), constant.MigrationPendingVersion)
		assertNoSentinel(t, err)
		assert.Empty(t, do.callsOf("update"))
	})

	passThrough := []string{"v1.9.3.2", "v1.9.3-tenant.1"}
	for _, version := range passThrough {
		t.Run("passes "+version+" as :version", func(t *testing.T) {
			do := newFakeDo()
			m := register.Migration{ID: migA, Version: version}
			require.NoError(t, NewRecordStore(newFakeOrm(do)).BackfillPendingVersion(kt, m))
			arg := updateArg(t, do)
			assert.Equal(t, version, arg["version"])
			assert.Equal(t, constant.MigrationPendingVersion, arg["pending"])
		})
	}
}
