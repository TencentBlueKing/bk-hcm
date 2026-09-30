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
	"reflect"
	"testing"
	"unsafe"

	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckInitialized(t *testing.T) {
	kt := kit.New()

	testCases := []struct {
		name      string
		setup     func(*fakeDo)
		wantErr   bool
		wantSent  bool
		errHas    []string
		wantCount int
	}{
		{
			name: "both tables exist",
			setup: func(do *fakeDo) {
				do.countDefault = 1
			},
			wantCount: 2,
		},
		{
			name: "audit missing",
			setup: func(do *fakeDo) {
				do.countQueue = []countReply{{n: 0}, {n: 1}}
			},
			wantErr:   true,
			wantSent:  true,
			errHas:    []string{constant.MigrationAuditTable, "do not exist", "run init first"},
			wantCount: 2,
		},
		{
			name: "record missing",
			setup: func(do *fakeDo) {
				do.countQueue = []countReply{{n: 1}, {n: 0}}
			},
			wantErr:   true,
			wantSent:  true,
			errHas:    []string{constant.MigrationRecordTable, "do not exist"},
			wantCount: 2,
		},
		{
			name: "both missing lists both",
			setup: func(do *fakeDo) {
				do.countDefault = 0
			},
			wantErr:  true,
			wantSent: true,
			errHas: []string{constant.MigrationAuditTable, constant.MigrationRecordTable,
				"do not exist"},
			wantCount: 2,
		},
		{
			name: "has table query error is plain",
			setup: func(do *fakeDo) {
				do.countErr = errors.New("schema down")
			},
			wantErr:   true,
			wantSent:  false,
			errHas:    []string{"check migration table", constant.MigrationAuditTable, "schema down"},
			wantCount: 1,
		},
		{
			name: "second table query error is plain",
			setup: func(do *fakeDo) {
				do.countQueue = []countReply{{n: 1}, {err: errors.New("timeout")}}
			},
			wantErr:   true,
			wantSent:  false,
			errHas:    []string{"check migration table", constant.MigrationRecordTable, "timeout"},
			wantCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := newFakeDo()
			tc.setup(do)
			err := schema.CheckInitialized(kt, newFakeOrm(do))
			assert.Len(t, do.callsOf("count"), tc.wantCount)
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tc.wantSent {
				assert.ErrorIs(t, err, migrate.ErrPrecondition)
			} else {
				assertNoSentinel(t, err)
			}
			for _, s := range tc.errHas {
				assert.Contains(t, err.Error(), s)
			}
			for _, c := range do.callsOf("count") {
				arg, ok := c.arg.(map[string]interface{})
				require.True(t, ok)
				_, hasTable := arg["table"]
				assert.True(t, hasTable)
			}
		})
	}
}

func TestPrepare(t *testing.T) {
	kt := kit.New()

	t.Run("not initialized", func(t *testing.T) {
		do := newFakeDo()
		do.countDefault = 0
		p, err := Prepare(kt, newFakeOrm(do), mustRegistry(t, "main", nil), Options{})
		require.Error(t, err)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Empty(t, do.callsOf("select"))
	})

	t.Run("has table error", func(t *testing.T) {
		do := newFakeDo()
		do.countErr = errors.New("down")
		p, err := Prepare(kt, newFakeOrm(do), mustRegistry(t, "main", nil), Options{})
		require.Error(t, err)
		assert.Nil(t, p)
		assertNoSentinel(t, err)
		assert.Empty(t, do.callsOf("select"))
	})

	t.Run("load invalid row", func(t *testing.T) {
		do := newFakeDo()
		do.countDefault = 1
		do.selectRows = []schema.Record{
			{MigrationID: migA, Version: "v1.9.3", AppliedPkg: "", Status: enumor.MigrationStatusSuccess},
		}
		p, err := Prepare(kt, newFakeOrm(do), mustRegistry(t, "main", nil), Options{})
		require.Error(t, err)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, migrate.ErrPrecondition)
		assert.Contains(t, err.Error(), "empty applied_pkg")
	})

	t.Run("empty registry builds empty plan", func(t *testing.T) {
		do := newFakeDo()
		do.countDefault = 1
		p, err := Prepare(kt, newFakeOrm(do), mustRegistry(t, "main", nil), Options{})
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, "main", p.Database)
		assert.False(t, p.HasReleasedVersion)
		assert.Empty(t, p.Items)
		assert.True(t, p.Passed())
	})

	t.Run("builds plan from records and registry", func(t *testing.T) {
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "next")
		owned := mustMigration(t, migB, "v1.9.3", "20260101120000", "owned")
		do := newFakeDo()
		do.countDefault = 1
		do.selectRows = []schema.Record{successRec(owned)}
		p, err := Prepare(kt, newFakeOrm(do), mustRegistry(t, "main", []register.Migration{m, owned}), Options{})
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.True(t, p.HasReleasedVersion)
		assert.Equal(t, "v1.9.3", p.Current.Raw)
		require.Len(t, p.Items, 2)
		// All() sorts by version: owned v1.9.3 before m v1.9.4.
		assert.Equal(t, enumor.MigrationActionSkipSuccess, p.Items[0].Action)
		assert.Equal(t, migB, p.Items[0].Migration.ID)
		assert.Equal(t, enumor.MigrationActionExecute, p.Items[1].Action)
		assert.Equal(t, migA, p.Items[1].Migration.ID)
	})
}

func TestExecute(t *testing.T) {
	kt := kit.New()

	t.Run("plan with issues writes nothing", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, "v1.9.2", "20260101120000", "missed")
		p := &Plan{
			Database:           "main",
			Current:            mustVersion(t, "v1.9.3"),
			HasReleasedVersion: true,
			Items:              []PlanItem{{Migration: m, Action: enumor.MigrationActionMissing}},
			Issues:             []Issue{{Kind: enumor.MigrationIssueMissed, Message: "missed migration"}},
			Warnings:           []string{"warn"},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.Contains(t, r.Err.Error(), "has 1 issues")
		assert.Empty(t, do.calls)
		assert.Empty(t, r.Executed)
		assert.Equal(t, []string{"warn"}, r.Warnings)
		assert.Contains(t, r.Messages, "missed migration")
		assert.Contains(t, r.Messages[len(r.Messages)-1], "has 1 issues")
		assert.Equal(t, "v1.9.3", r.VersionBefore)
		assert.Equal(t, "v1.9.3", r.VersionAfter)
	})

	t.Run("execute success marks running then success", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "ok")
		upCalled := 0
		m.Up = func(_ context.Context, _ orm.Interface) error {
			upCalled++
			return nil
		}
		p := &Plan{
			Database: "main",
			Items:    []PlanItem{{Migration: m, Action: enumor.MigrationActionExecute}},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Equal(t, 1, upCalled)
		assert.Equal(t, []string{migA}, r.Executed)
		assert.Equal(t, "v1.9.4", r.VersionAfter)
		assert.Equal(t, "", r.VersionBefore)
		require.Len(t, do.callsOf("insert"), 1)
		require.Len(t, do.callsOf("update"), 1)
		insertArg, ok := do.callsOf("insert")[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, m.Pkg, insertArg["applied_pkg"])
		assert.Equal(t, enumor.MigrationStatusRunning, insertArg["status"])
		updateArg := do.callsOf("update")[0].arg.(map[string]interface{})
		assert.Equal(t, enumor.MigrationStatusSuccess, updateArg["status"])
	})

	t.Run("up error marks failed and stops", func(t *testing.T) {
		do := newFakeDo()
		first := mustMigration(t, migA, "v1.9.4", "20260101120000", "fail")
		first.Up = func(_ context.Context, _ orm.Interface) error {
			return errors.New("up boom")
		}
		second := mustMigration(t, migB, "v1.9.5", "20260101120000", "later")
		secondCalled := false
		second.Up = func(_ context.Context, _ orm.Interface) error {
			secondCalled = true
			return nil
		}
		p := &Plan{
			Database: "main",
			Items: []PlanItem{
				{Migration: first, Action: enumor.MigrationActionExecute},
				{Migration: second, Action: enumor.MigrationActionExecute},
			},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.False(t, secondCalled)
		assert.Empty(t, r.Executed)
		assert.Contains(t, r.Err.Error(), "up boom")
		assert.Contains(t, r.Err.Error(), migA)
		assert.Equal(t, "", r.VersionAfter)
		require.Len(t, do.callsOf("insert"), 1)
		require.Len(t, do.callsOf("update"), 1)
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		assert.Equal(t, enumor.MigrationStatusFailed, arg["status"])
		assert.Contains(t, arg["message"], "up boom")
	})

	t.Run("mark failed error reported with up error", func(t *testing.T) {
		do := newFakeDo()
		do.updateErr = errors.New("mark denied")
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "fail")
		m.Up = func(_ context.Context, _ orm.Interface) error {
			return errors.New("up boom")
		}
		p := &Plan{Database: "main", Items: []PlanItem{{Migration: m, Action: enumor.MigrationActionExecute}}}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.Contains(t, r.Err.Error(), "up boom")
		assert.Contains(t, r.Err.Error(), "mark failed failed")
		assert.Contains(t, r.Err.Error(), "mark denied")
	})

	t.Run("mark running error stops before up", func(t *testing.T) {
		do := newFakeDo()
		do.insertErr = errors.New("insert denied")
		upCalled := false
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "x")
		m.Up = func(_ context.Context, _ orm.Interface) error {
			upCalled = true
			return nil
		}
		p := &Plan{Database: "main", Items: []PlanItem{{Migration: m, Action: enumor.MigrationActionExecute}}}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.False(t, upCalled)
		assert.Contains(t, r.Err.Error(), "insert denied")
		assert.Empty(t, do.callsOf("update"))
	})

	t.Run("mark success zero rows", func(t *testing.T) {
		do := newFakeDo()
		do.updateAffected = 0
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "x")
		p := &Plan{Database: "main", Items: []PlanItem{{Migration: m, Action: enumor.MigrationActionExecute}}}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.Contains(t, r.Err.Error(), "record not found")
		assert.Empty(t, r.Executed)
	})

	t.Run("up panic converted and mark failed", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, "v1.9.4", "20260101120000", "panic")
		m.Up = func(_ context.Context, _ orm.Interface) error {
			panic("kaboom")
		}
		p := &Plan{Database: "main", Items: []PlanItem{{Migration: m, Action: enumor.MigrationActionExecute}}}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.Contains(t, r.Err.Error(), "panic: kaboom")
		require.Len(t, do.callsOf("update"), 1)
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		assert.Equal(t, enumor.MigrationStatusFailed, arg["status"])
		assert.Contains(t, arg["message"], "panic: kaboom")
	})

	t.Run("skip backfill", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, "v1.9.3", "20260101120000", "same")
		ownerPkg := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "same").Pkg
		p := &Plan{
			Database: "main",
			Items:    []PlanItem{{Migration: m, Action: enumor.MigrationActionSkipBackfill, AppliedPkg: ownerPkg}},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Empty(t, r.Executed)
		require.Len(t, r.Skipped, 1)
		assert.Equal(t, enumor.MigrationSkipApplied, r.Skipped[0].Reason)
		assert.Equal(t, ownerPkg, r.Skipped[0].AppliedPkg)
		assert.Equal(t, "v1.9.3", r.VersionAfter)
		require.Len(t, do.callsOf("update"), 1)
		assert.Contains(t, do.callsOf("update")[0].expr, "`version` = :version")
	})

	t.Run("skip backfill error stops", func(t *testing.T) {
		do := newFakeDo()
		do.updateErr = errors.New("backfill denied")
		m := mustMigration(t, migA, "v1.9.3", "20260101120000", "same")
		later := mustMigration(t, migB, "v1.9.4", "20260101120000", "later")
		laterCalled := false
		later.Up = func(_ context.Context, _ orm.Interface) error {
			laterCalled = true
			return nil
		}
		p := &Plan{
			Database: "main",
			Items: []PlanItem{
				{Migration: m, Action: enumor.MigrationActionSkipBackfill},
				{Migration: later, Action: enumor.MigrationActionExecute},
			},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.False(t, laterCalled)
		assert.Contains(t, r.Err.Error(), "backfill denied")
	})

	t.Run("skip success and over ceiling", func(t *testing.T) {
		do := newFakeDo()
		skip := mustMigration(t, migA, "v1.9.3", "20260101120000", "skip")
		over := mustMigration(t, migB, "v1.9.5", "20260101120000", "over")
		p := &Plan{
			Database: "main",
			Items: []PlanItem{
				{Migration: skip, Action: enumor.MigrationActionSkipSuccess, AppliedPkg: "owner/pkg"},
				{Migration: over, Action: enumor.MigrationActionAboveMaxVersion},
			},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Empty(t, do.calls)
		require.Len(t, r.Skipped, 2)
		assert.Equal(t, enumor.MigrationSkipApplied, r.Skipped[0].Reason)
		assert.Equal(t, "owner/pkg", r.Skipped[0].AppliedPkg)
		assert.Equal(t, enumor.MigrationSkipAboveMaxVersion, r.Skipped[1].Reason)
		assert.Equal(t, "", r.VersionAfter)
	})

	t.Run("unknown action errors", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, "v1.9.2", "20260101120000", "missing")
		p := &Plan{
			Database: "main",
			Items:    []PlanItem{{Migration: m, Action: enumor.MigrationActionMissing}},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.Error(t, r.Err)
		assert.Contains(t, r.Err.Error(), string(enumor.MigrationActionMissing))
		assert.Contains(t, r.Err.Error(), "not executable")
		assert.Empty(t, do.calls)
	})

	t.Run("pending execute does not raise version", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "pending")
		p := &Plan{
			Database:           "main",
			Current:            mustVersion(t, "v1.9.3"),
			HasReleasedVersion: true,
			Items:              []PlanItem{{Migration: m, Action: enumor.MigrationActionExecute}},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Equal(t, []string{migA}, r.Executed)
		assert.Equal(t, "v1.9.3", r.VersionBefore)
		assert.Equal(t, "v1.9.3", r.VersionAfter)
	})

	t.Run("version after rises only on higher success", func(t *testing.T) {
		do := newFakeDo()
		lower := mustMigration(t, migA, "v1.9.2", "20260101120000", "lower")
		equal := mustMigration(t, migB, "v1.9.3", "20260101120000", "equal")
		higher := mustMigration(t, migC, "v1.9.4", "20260101120000", "higher")
		p := &Plan{
			Database:           "main",
			Current:            mustVersion(t, "v1.9.3"),
			HasReleasedVersion: true,
			Items: []PlanItem{
				{Migration: lower, Action: enumor.MigrationActionExecute},
				{Migration: equal, Action: enumor.MigrationActionExecute},
				{Migration: higher, Action: enumor.MigrationActionExecute},
			},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Equal(t, []string{migA, migB, migC}, r.Executed)
		assert.Equal(t, "v1.9.3", r.VersionBefore)
		assert.Equal(t, "v1.9.4", r.VersionAfter)
	})

	t.Run("no current first released success sets version", func(t *testing.T) {
		do := newFakeDo()
		pending := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "p")
		released := mustMigration(t, migB, "v1.9.2", "20260101120000", "r")
		p := &Plan{
			Database: "main",
			Items: []PlanItem{
				{Migration: pending, Action: enumor.MigrationActionExecute},
				{Migration: released, Action: enumor.MigrationActionExecute},
			},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Equal(t, "", r.VersionBefore)
		assert.Equal(t, "v1.9.2", r.VersionAfter)
	})

	t.Run("backfill raises version when higher", func(t *testing.T) {
		do := newFakeDo()
		m := mustMigration(t, migA, "v1.9.5", "20260101120000", "bf")
		p := &Plan{
			Database:           "main",
			Current:            mustVersion(t, "v1.9.3"),
			HasReleasedVersion: true,
			Items:              []PlanItem{{Migration: m, Action: enumor.MigrationActionSkipBackfill}},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Equal(t, "v1.9.5", r.VersionAfter)
	})

	t.Run("copies warnings from plan", func(t *testing.T) {
		do := newFakeDo()
		p := &Plan{
			Database: "main",
			Warnings: []string{"drift a", "drift b"},
		}
		r := Execute(kt, newFakeOrm(do), p)
		require.NoError(t, r.Err)
		assert.Equal(t, []string{"drift a", "drift b"}, r.Warnings)
	})
}

func TestNewBlockedResult(t *testing.T) {
	cause := fmt.Errorf("%w: blocked", migrate.ErrMissed)

	t.Run("plan with own issues", func(t *testing.T) {
		p := &Plan{
			Database: "main",
			Issues:   []Issue{{Kind: enumor.MigrationIssueMissed, Message: "own issue"}},
			Warnings: []string{"w"},
			Current:  mustVersion(t, "v1.9.3"), HasReleasedVersion: true,
		}
		r := NewBlockedResult(p, cause)
		assert.Equal(t, cause, r.Err)
		assert.Equal(t, []string{"own issue"}, r.Messages)
		assert.Equal(t, []string{"w"}, r.Warnings)
		assert.Equal(t, "v1.9.3", r.VersionBefore)
		assert.NotContains(t, fmt.Sprint(r.Messages), "not executed")
	})

	t.Run("passed plan blocked by another database", func(t *testing.T) {
		p := &Plan{Database: "obs", HasReleasedVersion: false}
		r := NewBlockedResult(p, cause)
		assert.Equal(t, cause, r.Err)
		require.Len(t, r.Messages, 1)
		assert.Contains(t, r.Messages[0], "not executed, err:")
		assert.Contains(t, r.Messages[0], cause.Error())
	})
}

func TestExecute_AuditResult(t *testing.T) {
	r := &ExecuteResult{
		Database:      "main",
		VersionBefore: "v1.9.2",
		VersionAfter:  "v1.9.3",
		Executed:      []string{migA},
		Skipped: []schema.SkippedMigration{{
			ID: migB, Version: "v1.9.3", Pkg: "pkg", AppliedPkg: "owner",
			Reason: enumor.MigrationSkipApplied,
		}},
		Warnings: []string{"w"},
		Messages: []string{"m"},
		Err:      errors.New("stop"),
	}
	got := r.AuditResult(4)
	assert.Equal(t, schema.AuditResult{
		ExitCode:      4,
		VersionBefore: "v1.9.2",
		VersionAfter:  "v1.9.3",
		Skipped:       r.Skipped,
		Warnings:      r.Warnings,
		Messages:      r.Messages,
	}, got)
}

// mustRegistry builds a registry with an unexported list for Prepare tests.
// Engine tests cannot Regist because types here are not under migrations/.
func mustRegistry(t *testing.T, database string, ms []register.Migration) *register.Registry {
	t.Helper()
	reg := &register.Registry{}
	v := reflect.ValueOf(reg).Elem()
	dbField := v.FieldByName("database")
	*(*string)(unsafe.Pointer(dbField.UnsafeAddr())) = database
	listField := v.FieldByName("list")
	*(*[]register.Migration)(unsafe.Pointer(listField.UnsafeAddr())) = append([]register.Migration(nil), ms...)
	return reg
}

func successRec(m register.Migration) schema.Record {
	return schema.Record{MigrationID: m.ID, Version: m.Version, AppliedPkg: m.Pkg,
		Status: enumor.MigrationStatusSuccess}
}
