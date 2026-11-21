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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"
	"hcm/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditBegin(t *testing.T) {
	t.Run("empty rid returns nil and makes no calls", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = ""
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up", Args: []string{"--db=main"}})
		assert.Nil(t, a)
		assert.Empty(t, do.calls)
	})

	t.Run("insert error returns nil", func(t *testing.T) {
		do := newFakeDo()
		do.insertErr = errors.New("duplicate run_id")
		kt := kit.New()
		kt.Rid = "run-1"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		assert.Nil(t, a)
		assert.NotEmpty(t, do.callsOf("insert"))
		a.End(kt, AuditResult{ExitCode: 0})
		assert.Empty(t, do.callsOf("update"))
	})

	t.Run("running-row select error still inserts", func(t *testing.T) {
		do := newFakeDo()
		do.selectErr = errors.New("select failed")
		kt := kit.New()
		kt.Rid = "run-2"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "init"})
		require.NotNil(t, a)
		assert.Empty(t, a.warnings)
		require.Len(t, do.callsOf("insert"), 1)
	})

	t.Run("previous running row prepends warning on insert", func(t *testing.T) {
		do := newFakeDo()
		do.selectNonRecord = []struct {
			RunID string `db:"run_id"`
		}{{RunID: "old-run"}}
		kt := kit.New()
		kt.Rid = "run-3"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up", Args: []string{"a", "b"}})
		require.NotNil(t, a)
		require.Len(t, a.warnings, 1)
		assert.Contains(t, a.warnings[0], "old-run")

		inserts := do.callsOf("insert")
		require.Len(t, inserts, 1)
		arg, ok := inserts[0].arg.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "a b", arg["args"])
		assert.Equal(t, enumor.MigrationStatusRunning, arg["status"])
	})

	t.Run("previous running warning appears first in End warnings", func(t *testing.T) {
		do := newFakeDo()
		do.selectNonRecord = []struct {
			RunID string `db:"run_id"`
		}{{RunID: "old-run"}}
		kt := kit.New()
		kt.Rid = "run-4"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		a.End(kt, AuditResult{ExitCode: 0, Warnings: []string{"later"}})
		updates := do.callsOf("update")
		require.Len(t, updates, 1)
		arg, ok := updates[0].arg.(map[string]interface{})
		require.True(t, ok)
		var warnings []string
		require.NoError(t, json.Unmarshal([]byte(arg["warnings"].(string)), &warnings))
		require.Len(t, warnings, 2)
		assert.Contains(t, warnings[0], "old-run")
		assert.Equal(t, "later", warnings[1])
	})

	t.Run("args truncated at 1024 bytes on rune boundary", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-args"
		raw := strings.Repeat("a", 1023) + "中"
		require.Greater(t, len(raw), constant.MigrationAuditArgsMaxBytes)
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up", Args: []string{raw}})
		require.NotNil(t, a)
		arg := do.callsOf("insert")[0].arg.(map[string]interface{})
		got := arg["args"].(string)
		assert.Equal(t, migrate.TruncateUTF8(raw, constant.MigrationAuditArgsMaxBytes), got)
		assert.LessOrEqual(t, len(got), constant.MigrationAuditArgsMaxBytes)
		assert.True(t, utf8.ValidString(got))
	})

	t.Run("binary_version and git_hash from version package", func(t *testing.T) {
		prevV, prevG := version.VERSION, version.GITHASH
		version.VERSION = "test-version-xyz"
		version.GITHASH = "abcdef0123456789"
		t.Cleanup(func() {
			version.VERSION = prevV
			version.GITHASH = prevG
		})

		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-ver"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		arg := do.callsOf("insert")[0].arg.(map[string]interface{})
		assert.Equal(t, "test-version-xyz", arg["binary_version"])
		assert.Equal(t, "abcdef0123456789", arg["git_hash"])
	})
}

func TestAuditEnd(t *testing.T) {
	t.Run("nil receiver does not panic", func(t *testing.T) {
		var a *Audit
		assert.NotPanics(t, func() {
			a.End(kit.New(), AuditResult{ExitCode: 1})
		})
	})

	for _, tc := range []struct {
		name     string
		exitCode int
		want     enumor.MigrationStatus
	}{
		{name: "exit 0 success", exitCode: 0, want: enumor.MigrationStatusSuccess},
		{name: "exit 1 failed", exitCode: 1, want: enumor.MigrationStatusFailed},
		{name: "exit 3 failed", exitCode: 3, want: enumor.MigrationStatusFailed},
		{name: "exit 4 failed", exitCode: 4, want: enumor.MigrationStatusFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			do := newFakeDo()
			kt := kit.New()
			kt.Rid = "run-status"
			a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
			require.NotNil(t, a)
			a.End(kt, AuditResult{ExitCode: tc.exitCode})
			arg := do.callsOf("update")[0].arg.(map[string]interface{})
			assert.Equal(t, tc.want, arg["status"])
			assert.Equal(t, tc.exitCode, arg["exit_code"])
		})
	}

	t.Run("nil slices become empty JSON arrays", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-nil"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		a.End(kt, AuditResult{ExitCode: 0, Skipped: nil, Warnings: nil, Messages: nil})
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		assert.Equal(t, "[]", arg["skipped"])
		assert.Equal(t, "[]", arg["warnings"])
		assert.Equal(t, "[]", arg["message"])
	})

	t.Run("exactly 200 warnings kept whole", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-200"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		warnings := make([]string, 200)
		for i := range warnings {
			warnings[i] = fmt.Sprintf("w%d", i)
		}
		a.End(kt, AuditResult{ExitCode: 0, Warnings: warnings})
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		var got []string
		require.NoError(t, json.Unmarshal([]byte(arg["warnings"].(string)), &got))
		require.Len(t, got, 200)
		assert.Equal(t, "w0", got[0])
		assert.Equal(t, "w199", got[199])
	})

	t.Run("201 warnings keep 199 plus omitted", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-201"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		warnings := make([]string, 201)
		for i := range warnings {
			warnings[i] = fmt.Sprintf("w%d", i)
		}
		a.End(kt, AuditResult{ExitCode: 0, Warnings: warnings})
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		var got []string
		require.NoError(t, json.Unmarshal([]byte(arg["warnings"].(string)), &got))
		require.Len(t, got, 200)
		assert.Equal(t, "w0", got[0])
		assert.Equal(t, "w198", got[198])
		assert.Equal(t, "... 2 more omitted", got[199])
	})

	t.Run("item of 1025 bytes truncated", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-trunc"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		item := strings.Repeat("b", constant.MigrationAuditItemMaxBytes+1)
		a.End(kt, AuditResult{ExitCode: 1, Messages: []string{item}})
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		var msgs []string
		require.NoError(t, json.Unmarshal([]byte(arg["message"].(string)), &msgs))
		require.Len(t, msgs, 1)
		assert.Equal(t, migrate.TruncateUTF8(item, constant.MigrationAuditItemMaxBytes), msgs[0])
		assert.Equal(t, constant.MigrationAuditItemMaxBytes, len(msgs[0]))
	})

	t.Run("invalid utf-8 replaced", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-utf8"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		a.End(kt, AuditResult{ExitCode: 1, Messages: []string{"bad\xff\xfebytes"}})
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		var msgs []string
		require.NoError(t, json.Unmarshal([]byte(arg["message"].(string)), &msgs))
		require.Len(t, msgs, 1)
		assert.Equal(t, "bad?bytes", msgs[0])
	})

	t.Run("skipped JSON has required keys", func(t *testing.T) {
		do := newFakeDo()
		kt := kit.New()
		kt.Rid = "run-skip"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		a.End(kt, AuditResult{
			ExitCode: 0,
			Skipped: []SkippedMigration{{
				ID: migA, Version: "v1.9.3", Pkg: "main/pending/20260101120000_a",
				AppliedPkg: "main/pending/20260101120000_a", Reason: enumor.MigrationSkipApplied,
			}},
		})
		arg := do.callsOf("update")[0].arg.(map[string]interface{})
		var items []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(arg["skipped"].(string)), &items))
		require.Len(t, items, 1)
		assert.Equal(t, migA, items[0]["id"])
		assert.Equal(t, "v1.9.3", items[0]["version"])
		assert.Equal(t, "main/pending/20260101120000_a", items[0]["pkg"])
		assert.Equal(t, "main/pending/20260101120000_a", items[0]["applied_pkg"])
		assert.Equal(t, string(enumor.MigrationSkipApplied), items[0]["reason"])
	})

}

func TestAuditEndUpdate(t *testing.T) {
	t.Run("update error does not panic", func(t *testing.T) {
		do := newFakeDo()
		do.updateErr = errors.New("deadlock")
		kt := kit.New()
		kt.Rid = "run-upd-err"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		assert.NotPanics(t, func() {
			a.End(kt, AuditResult{ExitCode: 0})
		})
	})

	t.Run("affected 0 does not panic", func(t *testing.T) {
		do := newFakeDo()
		do.updateAffected = 0
		kt := kit.New()
		kt.Rid = "run-upd-0"
		a := NewAudit(kt, newFakeOrm(do)).Begin(kt, AuditMeta{Command: "up"})
		require.NotNil(t, a)
		assert.NotPanics(t, func() {
			a.End(kt, AuditResult{ExitCode: 0})
		})
	})
}

func TestCapAuditItems(t *testing.T) {
	assert.Equal(t, []string{}, capAuditItems(nil))
	assert.Equal(t, []string{}, capAuditItems([]string{}))
}
