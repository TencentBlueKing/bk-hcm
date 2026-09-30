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
	"testing"

	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPlan(t *testing.T) {
	current := mustVersion(t, "v1.9.3")
	ceiling := mustVersion(t, "v1.9.3")
	lowCeiling := mustVersion(t, "v1.0.0")

	successRec := func(m register.Migration) schema.Record {
		return schema.Record{MigrationID: m.ID, Version: m.Version, AppliedPkg: m.Pkg,
			Status: enumor.MigrationStatusSuccess}
	}

	testCases := []struct {
		name               string
		migrations         []register.Migration
		records            schema.Records
		current            register.Version
		hasReleasedVersion bool
		opts               Options
		wantAction         []enumor.MigrationAction
		wantPkg            []string // AppliedPkg per item; empty string means unset
		wantIssues         []enumor.MigrationIssueKind
		wantWarn           int
		check              func(t *testing.T, p *Plan)
	}{
		{
			name: "empty registry empty records",
		},
		{
			name: "empty registry with success records ignored",
			records: schema.Records{migA: {
				MigrationID: migA, Version: "v1.9.3", AppliedPkg: "main/x", Status: enumor.MigrationStatusSuccess,
			}},
			current:            current,
			hasReleasedVersion: true,
		},
		{
			name: "pending denied regardless of ceiling",
			migrations: []register.Migration{
				mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "pending"),
			},
			opts:       Options{Ceiling: &ceiling},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionPendingDenied},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssuePending},
			check: func(t *testing.T, p *Plan) {
				assert.Contains(t, p.Issues[0].Message, "--allow-pending")
				assert.Contains(t, p.Issues[0].Message, migA)
			},
		},
		{
			name: "pending allowed with ceiling is over ceiling",
			migrations: []register.Migration{
				mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "pending"),
			},
			opts:       Options{AllowPending: true, Ceiling: &ceiling},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionAboveMaxVersion},
		},
		{
			name: "pending allowed without ceiling executes",
			migrations: []register.Migration{
				mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "pending"),
			},
			opts:       Options{AllowPending: true},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "version equal to ceiling executes",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3", "20260101120000", "eq"),
			},
			opts:       Options{Ceiling: &ceiling},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "version above ceiling is over ceiling",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "above"),
			},
			opts:       Options{Ceiling: &ceiling},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionAboveMaxVersion},
		},
		{
			name: "ceiling below everything",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.2", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.3", "20260101120000", "b"),
			},
			opts: Options{Ceiling: &lowCeiling},
			wantAction: []enumor.MigrationAction{
				enumor.MigrationActionAboveMaxVersion, enumor.MigrationActionAboveMaxVersion,
			},
		},
		{
			name: "over ceiling not compared with owner",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "other_tag"),
			},
			records: schema.Records{
				migA: {MigrationID: migA, Version: "v1.9.3",
					AppliedPkg: "main/v1.9.3/v1.9.3_20260101120000_owner_tag",
					Status:     enumor.MigrationStatusSuccess},
			},
			current:            current,
			hasReleasedVersion: true,
			opts:               Options{Ceiling: &ceiling},
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionAboveMaxVersion},
			wantPkg:            []string{""},
		},
		{
			name: "running record does not seed owner so resumes",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "resume"),
			},
			records: schema.Records{
				migA: {MigrationID: migA, Version: "v1.9.4", Status: enumor.MigrationStatusRunning},
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "failed record does not seed owner so resumes",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "resume"),
			},
			records: schema.Records{
				migA: {MigrationID: migA, Version: "v1.9.4", Status: enumor.MigrationStatusFailed, Message: "boom"},
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "record owned different suffix is id reuse",
			migrations: func() []register.Migration {
				m := mustMigration(t, migA, "v1.9.3", "20260101120000", "registered")
				return []register.Migration{m}
			}(),
			records: func() schema.Records {
				owner := mustMigration(t, migA, "v1.9.3", "20260101120000", "recorded")
				return schema.Records{migA: successRec(owner)}
			}(),
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionIDReuse},
			wantIssues:         []enumor.MigrationIssueKind{enumor.MigrationIssueIDReuse},
			check: func(t *testing.T, p *Plan) {
				assert.Contains(t, p.Issues[0].Message, "recorded version")
				assert.Contains(t, p.Issues[0].Message, "applied_pkg")
				assert.NotContains(t, p.Issues[0].Message, "registered by pkg")
				require.NotEmpty(t, p.Items[0].AppliedPkg)
			},
		},
		{
			name: "in-run owned different suffix is id reuse",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "first"),
				mustMigration(t, migA, "v1.9.4", "20260102120000", "second"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionIDReuse},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssueIDReuse},
			check: func(t *testing.T, p *Plan) {
				assert.Contains(t, p.Issues[0].Message, "registered by pkg")
				assert.NotContains(t, p.Issues[0].Message, "recorded version")
				assert.Equal(t, p.Items[0].Migration.Pkg, p.Items[1].AppliedPkg)
			},
		},
		{
			name: "owner suffix unreadable never matches",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3", "20260101120000", "ok"),
			},
			records: schema.Records{
				migA: {MigrationID: migA, Version: "v1.9.3", AppliedPkg: "main/v1.9.3/not-a-migration-dir",
					Status: enumor.MigrationStatusSuccess},
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionIDReuse},
			wantIssues:         []enumor.MigrationIssueKind{enumor.MigrationIssueIDReuse},
			wantPkg:            []string{"main/v1.9.3/not-a-migration-dir"},
		},
		{
			name: "migration suffix unreadable never matches owner",
			migrations: func() []register.Migration {
				m := mustMigration(t, migA, "v1.9.3", "20260101120000", "ok")
				m.Pkg = "main/v1.9.3/broken"
				return []register.Migration{m}
			}(),
			records: func() schema.Records {
				owner := mustMigration(t, migA, "v1.9.3", "20260101120000", "ok")
				return schema.Records{migA: successRec(owner)}
			}(),
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionIDReuse},
			wantIssues:         []enumor.MigrationIssueKind{enumor.MigrationIssueIDReuse},
		},
		{
			name: "same suffix pending owner released m is skip backfill",
			migrations: func() []register.Migration {
				return []register.Migration{
					mustMigration(t, migA, "v1.9.3", "20260101120000", "same"),
				}
			}(),
			records: func() schema.Records {
				owner := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "same")
				return schema.Records{migA: successRec(owner)}
			}(),
			wantAction: []enumor.MigrationAction{enumor.MigrationActionSkipBackfill},
			check: func(t *testing.T, p *Plan) {
				owner := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "same")
				assert.Equal(t, owner.Pkg, p.Items[0].AppliedPkg)
			},
		},
		{
			name: "backfill updates owner version for later drift warning",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3", "20260101120000", "same"),
				mustMigration(t, migA, "v1.9.4", "20260101120000", "same"),
			},
			records: func() schema.Records {
				owner := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "same")
				return schema.Records{migA: successRec(owner)}
			}(),
			wantAction: []enumor.MigrationAction{enumor.MigrationActionSkipBackfill, enumor.MigrationActionSkipSuccess},
			wantWarn:   1,
			check: func(t *testing.T, p *Plan) {
				assert.Contains(t, p.Warnings[0], "migration version drift")
				assert.Contains(t, p.Warnings[0], "recorded: v1.9.3")
				assert.Contains(t, p.Warnings[0], "registered: v1.9.4")
			},
		},
		{
			name: "same suffix success skip with version drift warning",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "same"),
			},
			records: func() schema.Records {
				owner := mustMigration(t, migA, "v1.9.3", "20260101120000", "same")
				return schema.Records{migA: successRec(owner)}
			}(),
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionSkipSuccess},
			wantWarn:           1,
			check: func(t *testing.T, p *Plan) {
				assert.Contains(t, p.Warnings[0], "migration version drift")
			},
		},
		{
			name: "same suffix same version skip without warning",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3", "20260101120000", "same"),
			},
			records: func() schema.Records {
				owner := mustMigration(t, migA, "v1.9.3", "20260101120000", "same")
				return schema.Records{migA: successRec(owner)}
			}(),
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionSkipSuccess},
			wantWarn:           0,
		},
		{
			name: "pending owner and pending m skip without warning",
			migrations: []register.Migration{
				mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "same"),
			},
			records: func() schema.Records {
				owner := mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "same")
				return schema.Records{migA: successRec(owner)}
			}(),
			opts:       Options{AllowPending: true},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionSkipSuccess},
			wantWarn:   0,
		},
		{
			name: "missed migration when version at or below current",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.2", "20260101120000", "missed"),
				mustMigration(t, migB, "v1.9.3", "20260101120000", "equal"),
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionMissing, enumor.MigrationActionMissing},
			wantIssues:         []enumor.MigrationIssueKind{enumor.MigrationIssueMissed, enumor.MigrationIssueMissed},
			check: func(t *testing.T, p *Plan) {
				assert.Contains(t, p.Issues[0].Message, "missed migration")
				assert.Contains(t, p.Issues[0].Message, "current: v1.9.3")
			},
		},
		{
			name: "catch-up executes missed",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.2", "20260101120000", "missed"),
			},
			current:            current,
			hasReleasedVersion: true,
			opts:               Options{CatchUp: true},
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "above current executes",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "next"),
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "no current executes released",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.2", "20260101120000", "first"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute},
		},
		{
			name: "missing claims id so later same suffix skips",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.2", "20260101120000", "same"),
				mustMigration(t, migA, "v1.9.2", "20260101120000", "same"),
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction: []enumor.MigrationAction{
				enumor.MigrationActionMissing, enumor.MigrationActionSkipSuccess,
			},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssueMissed},
		},
		{
			name: "missing claims id so later different suffix is id reuse",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.2", "20260101120000", "first"),
				mustMigration(t, migA, "v1.9.2", "20260102120000", "second"),
			},
			current:            current,
			hasReleasedVersion: true,
			wantAction:         []enumor.MigrationAction{enumor.MigrationActionMissing, enumor.MigrationActionIDReuse},
			wantIssues:         []enumor.MigrationIssueKind{enumor.MigrationIssueMissed, enumor.MigrationIssueIDReuse},
		},
		{
			name: "execute claims id second same suffix skip success",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "same"),
				mustMigration(t, migA, "v1.9.4", "20260101120000", "same"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionSkipSuccess},
		},
		{
			name: "three duplicate ids execute skip id reuse",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "first"),
				mustMigration(t, migA, "v1.9.4", "20260101120000", "first"),
				mustMigration(t, migA, "v1.9.4", "20260103120000", "third"),
			},
			wantAction: []enumor.MigrationAction{
				enumor.MigrationActionExecute, enumor.MigrationActionSkipSuccess, enumor.MigrationActionIDReuse,
			},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssueIDReuse},
		},
		{
			name: "three copies same suffix",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4", "20260101120000", "same"),
				mustMigration(t, migA, "v1.9.4", "20260101120000", "same"),
				mustMigration(t, migA, "v1.9.5", "20260101120000", "same"),
			},
			wantAction: []enumor.MigrationAction{
				enumor.MigrationActionExecute, enumor.MigrationActionSkipSuccess, enumor.MigrationActionSkipSuccess,
			},
			wantWarn: 1,
		},
		{
			name: "two version labels one issue listing sorted labels",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3-alpha.1", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.3-beta.1", "20260101120000", "b"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionExecute},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssueLabels},
			check: func(t *testing.T, p *Plan) {
				require.Len(t, p.Issues, 1)
				assert.Contains(t, p.Issues[0].Message, "more than one version label")
				assert.Contains(t, p.Issues[0].Message, "alpha: ["+migA+"]")
				assert.Contains(t, p.Issues[0].Message, "beta: ["+migB+"]")
				// Sorted: alpha before beta.
				assert.Less(t, indexOf(p.Issues[0].Message, "alpha"), indexOf(p.Issues[0].Message, "beta"))
			},
		},
		{
			name: "single label no issue",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3-tenant.1", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.3-tenant.2", "20260101120000", "b"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionExecute},
		},
		{
			name: "three segment and one label is one line",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.4-tenant.1", "20260101120000", "b"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionExecute},
		},
		{
			name: "numeric fourth and one label is two lines",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4.1", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.4-tenant.1", "20260101120000", "b"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionExecute},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssueLabels},
			check: func(t *testing.T, p *Plan) {
				require.Len(t, p.Issues, 1)
				assert.Contains(t, p.Issues[0].Message, ": ["+migA+"]")
				assert.Contains(t, p.Issues[0].Message, "tenant: ["+migB+"]")
				assert.Less(t, indexOf(p.Issues[0].Message, ": ["), indexOf(p.Issues[0].Message, "tenant"))
			},
		},
		{
			name: "no labels no issue",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.3", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.3.1", "20260101120000", "b"),
			},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionExecute},
		},
		{
			name: "labels above ceiling still counted",
			migrations: []register.Migration{
				mustMigration(t, migA, "v1.9.4-alpha.1", "20260101120000", "a"),
				mustMigration(t, migB, "v1.9.4-beta.1", "20260101120000", "b"),
			},
			opts: Options{Ceiling: &ceiling},
			wantAction: []enumor.MigrationAction{
				enumor.MigrationActionAboveMaxVersion, enumor.MigrationActionAboveMaxVersion,
			},
			wantIssues: []enumor.MigrationIssueKind{enumor.MigrationIssueLabels},
		},
		{
			name: "pending ignored by label check",
			migrations: []register.Migration{
				mustMigration(t, migA, constant.MigrationPendingVersion, "20260101120000", "p"),
				mustMigration(t, migB, "v1.9.3-tenant.1", "20260101120000", "t"),
			},
			opts:       Options{AllowPending: true},
			wantAction: []enumor.MigrationAction{enumor.MigrationActionExecute, enumor.MigrationActionExecute},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := BuildPlan("main", tc.migrations, tc.records, tc.current, tc.hasReleasedVersion, tc.opts)
			require.Equal(t, "main", p.Database)
			assert.Equal(t, tc.hasReleasedVersion, p.HasReleasedVersion)
			if tc.hasReleasedVersion {
				assert.Equal(t, tc.current, p.Current)
			}
			require.Len(t, p.Items, len(tc.migrations))
			if tc.wantAction != nil {
				require.Len(t, p.Items, len(tc.wantAction))
				for i, want := range tc.wantAction {
					assert.Equal(t, want, p.Items[i].Action, "item %d id=%s", i, p.Items[i].Migration.ID)
				}
			}
			if tc.wantPkg != nil {
				require.Len(t, p.Items, len(tc.wantPkg))
				for i, want := range tc.wantPkg {
					assert.Equal(t, want, p.Items[i].AppliedPkg, "item %d AppliedPkg", i)
				}
			}
			gotKinds := make([]enumor.MigrationIssueKind, 0, len(p.Issues))
			for _, issue := range p.Issues {
				gotKinds = append(gotKinds, issue.Kind)
			}
			if tc.wantIssues == nil {
				assert.Empty(t, p.Issues)
				assert.True(t, p.Passed())
			} else {
				assert.Equal(t, tc.wantIssues, gotKinds)
				assert.False(t, p.Passed())
			}
			assert.Len(t, p.Warnings, tc.wantWarn)
			if tc.check != nil {
				tc.check(t, p)
			}
		})
	}
}

func TestPlanCurrentRaw(t *testing.T) {
	assert.Equal(t, "", (&Plan{}).CurrentRaw())
	assert.Equal(t, "v1.9.3", (&Plan{HasReleasedVersion: true, Current: mustVersion(t, "v1.9.3")}).CurrentRaw())
}

func TestCollectPlanErrors(t *testing.T) {
	kt := kit.New()

	t.Run("nil plans", func(t *testing.T) {
		require.NoError(t, CollectPlanErrors(kt, nil))
	})

	t.Run("all passed", func(t *testing.T) {
		require.NoError(t, CollectPlanErrors(kt, []*Plan{
			{Database: "main"},
			{Database: "aux"},
		}))
	})

	t.Run("id reuse beats registry and missed", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{{
			Database: "main",
			Issues: []Issue{
				{Kind: enumor.MigrationIssueMissed, Message: "missed one"},
				{Kind: enumor.MigrationIssuePending, Message: "pending one"},
				{Kind: enumor.MigrationIssueIDReuse, Message: "reuse one"},
			},
		}})
		require.Error(t, err)
		assert.ErrorIs(t, err, migrate.ErrIDReuse)
		assert.NotErrorIs(t, err, migrate.ErrRegistry)
		assert.NotErrorIs(t, err, migrate.ErrMissed)
		assert.Contains(t, err.Error(), "3 problems found")
		assert.Contains(t, err.Error(), "[main] missed one")
		assert.Contains(t, err.Error(), "[main] pending one")
		assert.Contains(t, err.Error(), "[main] reuse one")
	})

	t.Run("pending beats missed", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{{
			Database: "main",
			Issues: []Issue{
				{Kind: enumor.MigrationIssueMissed, Message: "missed"},
				{Kind: enumor.MigrationIssuePending, Message: "pending"},
			},
		}})
		require.Error(t, err)
		assert.ErrorIs(t, err, migrate.ErrRegistry)
		assert.NotErrorIs(t, err, migrate.ErrMissed)
		assert.NotErrorIs(t, err, migrate.ErrIDReuse)
	})

	t.Run("labels beats missed", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{{
			Database: "aux",
			Issues: []Issue{
				{Kind: enumor.MigrationIssueMissed, Message: "missed"},
				{Kind: enumor.MigrationIssueLabels, Message: "labels"},
			},
		}})
		require.Error(t, err)
		assert.ErrorIs(t, err, migrate.ErrRegistry)
		assert.NotErrorIs(t, err, migrate.ErrMissed)
		assert.Contains(t, err.Error(), "[aux] labels")
	})

	t.Run("unknown kind is a plain failure", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{{
			Database: "main",
			Issues:   []Issue{{Kind: "other", Message: "x"}},
		}})
		require.Error(t, err)
		assert.Equal(t, constant.MigrationExitFailure, migrate.ExitCode(err))
		assert.NotErrorIs(t, err, migrate.ErrMissed)
		assert.NotErrorIs(t, err, migrate.ErrRegistry)
		assert.NotErrorIs(t, err, migrate.ErrIDReuse)
		assert.Contains(t, err.Error(), "unknown issue kind")
		assert.Contains(t, err.Error(), "[main] x")
	})

	t.Run("only missed", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{{
			Database: "main",
			Issues:   []Issue{{Kind: enumor.MigrationIssueMissed, Message: "missed"}},
		}})
		require.Error(t, err)
		assert.ErrorIs(t, err, migrate.ErrMissed)
		assert.NotErrorIs(t, err, migrate.ErrRegistry)
		assert.NotErrorIs(t, err, migrate.ErrIDReuse)
		assert.Contains(t, err.Error(), "1 problems found")
	})

	t.Run("issues across databases", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{
			{Database: "main", Issues: []Issue{{Kind: enumor.MigrationIssueMissed, Message: "a"}}},
			{Database: "aux", Issues: []Issue{{Kind: enumor.MigrationIssueMissed, Message: "b"}}},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, migrate.ErrMissed)
		assert.Contains(t, err.Error(), "2 problems found")
		assert.Contains(t, err.Error(), "[main] a")
		assert.Contains(t, err.Error(), "[aux] b")
	})

	t.Run("wraps exactly one sentinel", func(t *testing.T) {
		err := CollectPlanErrors(kt, []*Plan{{
			Database: "main",
			Issues:   []Issue{{Kind: enumor.MigrationIssueIDReuse, Message: "x"}},
		}})
		require.Error(t, err)
		assert.True(t, errors.Is(err, migrate.ErrIDReuse))
		unwrapped := errors.Unwrap(err)
		assert.Equal(t, migrate.ErrIDReuse, unwrapped)
	})
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
