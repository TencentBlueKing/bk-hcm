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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"hcm/migrate/register"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordDBTags(t *testing.T) {
	assert.Equal(t, Status("running"), StatusRunning)
	assert.Equal(t, Status("success"), StatusSuccess)
	assert.Equal(t, Status("failed"), StatusFailed)
	assert.Equal(t, 1024, maxMessageBytes)

	want := []struct {
		field string
		tag   string
	}{
		{field: "ID", tag: "id"},
		{field: "MigrationID", tag: "migration_id"},
		{field: "Version", tag: "version"},
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
	})

	t.Run("indexes known statuses by migration id", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{ID: 1, MigrationID: migA, Version: "v1.9.3", Status: StatusRunning},
			{ID: 2, MigrationID: migB, Version: "v1.9.4", Status: StatusSuccess, Message: "ok"},
			{ID: 3, MigrationID: migC, Version: "v1.9.5", Status: StatusFailed, Message: "boom"},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.NoError(t, err)
		require.Len(t, records, 3)
		assert.Equal(t, do.selectRows[0], records[migA])
		assert.Equal(t, do.selectRows[1], records[migB])
		assert.Equal(t, StatusFailed, records[migC].Status)
		assert.Equal(t, "boom", records[migC].Message)
	})

	testCases := []struct {
		name   string
		status Status
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
				{MigrationID: migA, Version: "v1.9.3", Status: StatusSuccess},
				{MigrationID: migB, Version: "v1.9.4", Status: tc.status},
				{MigrationID: migC, Version: "v1.9.5", Status: StatusFailed},
			}
			records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
			require.Error(t, err)
			assert.Nil(t, records)
			assert.ErrorIs(t, err, ErrPrecondition)
			assert.Contains(t, err.Error(), migB)
			assert.Contains(t, err.Error(), string(tc.status))
		})
	}

	badVersions := []string{"1.9.3", "PENDING", "", "v1.9", "v1.9.3-Tenant.1", "v01.9.3"}
	for _, status := range []Status{StatusRunning, StatusFailed, StatusSuccess} {
		for _, bad := range badVersions {
			t.Run(fmt.Sprintf("unparsable %s version %q", status, bad), func(t *testing.T) {
				do := newFakeDo()
				do.selectRows = []Record{
					{MigrationID: migA, Version: "v1.9.3", Status: StatusSuccess},
					{MigrationID: migB, Version: bad, Status: status},
				}
				records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
				require.Error(t, err)
				assert.Nil(t, records)
				assert.ErrorIs(t, err, ErrPrecondition)
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

	t.Run("unknown status is reported before an unparsable version", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{MigrationID: migA, Version: "v1.9.3", Status: StatusSuccess},
			{MigrationID: migB, Version: "PENDING", Status: "done"},
		}
		records, err := NewRecordStore(newFakeOrm(do)).Load(kt)
		require.Error(t, err)
		assert.Nil(t, records)
		assert.ErrorIs(t, err, ErrPrecondition)
		assert.Contains(t, err.Error(), migB)
		assert.Contains(t, err.Error(), "unknown status")
		assert.Contains(t, err.Error(), "done")
		assert.NotContains(t, err.Error(), "unparsable version")
	})

	t.Run("indexes labeled and folded versions of every status", func(t *testing.T) {
		do := newFakeDo()
		do.selectRows = []Record{
			{ID: 1, MigrationID: migA, Version: "v1.9.3-tenant.1", Status: StatusRunning},
			{ID: 2, MigrationID: migB, Version: "v1.9.3.0", Status: StatusSuccess},
			{ID: 3, MigrationID: migC, Version: "v1.9.4", Status: StatusFailed},
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
	m := register.Migration{ID: migA, Version: "v1.9.3"}

	t.Run("insert upsert", func(t *testing.T) {
		do := newFakeDo()
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkRunning(kt, m))
		inserts := do.callsOf("insert")
		require.Len(t, inserts, 1)
		assert.Contains(t, inserts[0].expr, "ON DUPLICATE KEY UPDATE")
		assert.Contains(t, inserts[0].expr, "`status` = :status")
		assert.Contains(t, inserts[0].expr, "`message` = ''")
		assert.Contains(t, inserts[0].expr, "INSERT INTO `hcm_migration_record`")
		arg, ok := inserts[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, m.ID, arg["migration_id"])
		assert.Equal(t, m.Version, arg["version"])
		assert.Equal(t, StatusRunning, arg["status"])
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
	m := register.Migration{ID: migA, Version: "v1.9.4"}

	t.Run("updates status and registered version", func(t *testing.T) {
		do := newFakeDo()
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkSuccess(kt, m))
		updates := do.callsOf("update")
		require.Len(t, updates, 1)
		assert.Contains(t, updates[0].expr, "`status` = :status")
		assert.Contains(t, updates[0].expr, "`message` = ''")
		assert.Contains(t, updates[0].expr, "`version` = :version")
		assert.Contains(t, updates[0].expr, "WHERE `migration_id` = :migration_id")
		arg, ok := updates[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, m.ID, arg["migration_id"])
		assert.Equal(t, m.Version, arg["version"])
		assert.Equal(t, StatusSuccess, arg["status"])
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
	m := register.Migration{ID: migB, Version: "v1.9.3"}

	t.Run("nil error stores empty message", func(t *testing.T) {
		do := newFakeDo()
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkFailed(kt, m, nil))
		arg := updateArg(t, do)
		assert.Equal(t, "", arg["message"])
		assert.Equal(t, StatusFailed, arg["status"])
		assert.Equal(t, m.ID, arg["migration_id"])
		_, hasVersion := arg["version"]
		assert.False(t, hasVersion)
		expr := do.callsOf("update")[0].expr
		assert.Contains(t, expr, "`message` = :message")
		assert.NotContains(t, expr, "`version`")
	})

	t.Run("truncates message", func(t *testing.T) {
		do := newFakeDo()
		raw := strings.Repeat("迁移失败，磁盘已满。", 80)
		require.Greater(t, len(raw), maxMessageBytes)
		require.NoError(t, NewRecordStore(newFakeOrm(do)).MarkFailed(kt, m, errors.New(raw)))
		arg := updateArg(t, do)
		msg, ok := arg["message"].(string)
		require.True(t, ok)
		assert.LessOrEqual(t, len(msg), maxMessageBytes)
		assert.True(t, utf8.ValidString(msg))
		assert.True(t, strings.HasPrefix(raw, msg))
		assert.NotEqual(t, raw, msg)
		assert.Equal(t, truncateMessage(raw, maxMessageBytes), msg)
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

func TestTruncateMessage(t *testing.T) {
	han := "中"
	require.Equal(t, 3, len(han))
	emoji := "😀"
	require.Equal(t, 4, len(emoji))

	testCases := []struct {
		name  string
		msg   string
		limit int
		want  string
	}{
		{name: "empty", msg: "", limit: maxMessageBytes, want: ""},
		{
			name:  "exactly 1024 ascii",
			msg:   strings.Repeat("a", maxMessageBytes),
			limit: maxMessageBytes,
			want:  strings.Repeat("a", maxMessageBytes),
		},
		{
			name:  "1025 ascii",
			msg:   strings.Repeat("a", maxMessageBytes+1),
			limit: maxMessageBytes,
			want:  strings.Repeat("a", maxMessageBytes),
		},
		{
			// Index 1024 is byte 1 of 中, a rune start, so the rune is excluded.
			name:  "cut on byte 1 of han",
			msg:   strings.Repeat("a", 1024) + han,
			limit: 1024,
			want:  strings.Repeat("a", 1024),
		},
		{
			// Index 1024 is byte 2 of 中. The cut walks back to the rune start.
			name:  "cut on byte 2 of han",
			msg:   strings.Repeat("a", 1023) + han,
			limit: 1024,
			want:  strings.Repeat("a", 1023),
		},
		{
			// Index 1024 is byte 3 of 中.
			name:  "cut on byte 3 of han",
			msg:   strings.Repeat("a", 1022) + han,
			limit: 1024,
			want:  strings.Repeat("a", 1022),
		},
		{
			// 😀 occupies indexes 1021..1024, so the limit falls inside the rune.
			name:  "emoji straddles the limit",
			msg:   strings.Repeat("a", 1021) + emoji + "tail",
			limit: 1024,
			want:  strings.Repeat("a", 1021),
		},
		{
			name:  "invalid utf-8 replaced",
			msg:   "bad\xff\xfebytes",
			limit: maxMessageBytes,
			want:  "bad?bytes",
		},
		{
			name:  "invalid utf-8 past the limit is dropped",
			msg:   strings.Repeat("a", 1024) + "\xff",
			limit: 1024,
			want:  strings.Repeat("a", 1024),
		},
		{name: "limit zero", msg: "中文", limit: 0, want: ""},
		{name: "limit zero on empty", msg: "", limit: 0, want: ""},
		{name: "shorter multibyte unchanged", msg: "库当前版本", limit: maxMessageBytes, want: "库当前版本"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateMessage(tc.msg, tc.limit)
			assert.Equal(t, tc.want, got)
			assert.True(t, utf8.ValidString(got))
			assert.LessOrEqual(t, len(got), tc.limit)
			sanitized := strings.ToValidUTF8(tc.msg, "?")
			assert.True(t, strings.HasPrefix(sanitized, got), "got %q is not a prefix of %q", got, sanitized)
			if utf8.ValidString(tc.msg) {
				assert.True(t, strings.HasPrefix(tc.msg, got))
			}
		})
	}
}

func TestCurrentVersion(t *testing.T) {
	success := func(versions ...string) Records {
		out := make(Records, len(versions))
		for i, version := range versions {
			id := fmt.Sprintf("id-%d", i)
			out[id] = Record{MigrationID: id, Version: version, Status: StatusSuccess}
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
				migA: {MigrationID: migA, Status: StatusRunning, Version: "v1.9.10"},
				migB: {MigrationID: migB, Status: StatusFailed, Version: "v1.9.9"},
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
				"current":  {MigrationID: "current", Status: StatusSuccess, Version: "v1.9.4"},
				"backfill": {MigrationID: "backfill", Status: StatusSuccess, Version: "v1.9.3"},
				"running":  {MigrationID: "running", Status: StatusRunning, Version: "v9.9.9"},
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
				migA: {MigrationID: migA, Status: StatusFailed, Version: "PENDING"},
				migB: {MigrationID: migB, Status: StatusRunning, Version: ""},
				migC: {MigrationID: migC, Status: StatusRunning, Version: "1.9.3"},
				migD: {MigrationID: migD, Status: StatusSuccess, Version: "v1.9.3"},
			},
			wantRaw: "v1.9.3",
			wantOK:  true,
		},
		{
			name: "CurrentVersion ignores only unparsable non-success records",
			records: Records{
				migA: {MigrationID: migA, Status: StatusFailed, Version: "nope"},
				migB: {MigrationID: migB, Status: StatusRunning, Version: "v1.9"},
			},
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

	badVersions := []string{"1.9.3", "PENDING", "", "v1.9"}
	for _, bad := range badVersions {
		t.Run("unparsable success "+bad, func(t *testing.T) {
			records := Records{
				migA: {MigrationID: migA, Status: StatusSuccess, Version: "v1.9.10"},
				migB: {MigrationID: migB, Status: StatusSuccess, Version: bad},
				migC: {MigrationID: migC, Status: StatusFailed, Version: "v1.9.4"},
			}
			current, ok, err := CurrentVersion(records)
			require.Error(t, err)
			assert.False(t, ok)
			assert.Equal(t, register.Version{}, current)
			assert.ErrorIs(t, err, ErrPrecondition)
			assert.Contains(t, err.Error(), migB)
		})
	}
}

func TestCurrentVersionEqualRawIsStable(t *testing.T) {
	for i := 0; i < 50; i++ {
		records := Records{
			fmt.Sprintf("folded-%d", i): {MigrationID: migA, Status: StatusSuccess, Version: "v1.9.3.0"},
			fmt.Sprintf("plain-%d", i):  {MigrationID: migB, Status: StatusSuccess, Version: "v1.9.3"},
			fmt.Sprintf("older-%d", i):  {MigrationID: migC, Status: StatusSuccess, Version: "v1.9.2"},
		}
		current, ok, err := CurrentVersion(records)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "v1.9.3", current.Raw)
	}
}

func TestWarnVersionDrift(t *testing.T) {
	kt := kit.New()
	testCases := []struct {
		name       string
		status     Status
		recorded   string
		registered string
		want       bool
	}{
		{
			name: "success same version", status: StatusSuccess,
			recorded: "v1.9.3", registered: "v1.9.3", want: false,
		},
		{
			name: "success same labeled version", status: StatusSuccess,
			recorded: "v1.9.3-tenant.1", registered: "v1.9.3-tenant.1", want: false,
		},
		{
			name: "success drift from feature label to three segments", status: StatusSuccess,
			recorded: "v1.9.3-tenant.1", registered: "v1.9.3", want: true,
		},
		{
			name: "running different version", status: StatusRunning,
			recorded: "v1.9.3-tenant.1", registered: "v1.9.3", want: false,
		},
		{
			name: "failed different version", status: StatusFailed,
			recorded: "v1.9.2", registered: "v1.9.3", want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := Record{MigrationID: migA, Status: tc.status, Version: tc.recorded}
			m := register.Migration{ID: migA, Version: tc.registered}
			assert.Equal(t, tc.want, WarnVersionDrift(kt, rec, m))
		})
	}
}
