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

// Package engine plans and runs registered migrations.
package engine

import (
	"fmt"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/migrate"
)

// Prepare reads the records of one initialized database and builds its plan.
// It writes nothing. A database missing either migration table, or holding
// an invalid record, returns migrate.ErrPrecondition.
func Prepare(kt *kit.Kit, o orm.Interface, reg *register.Registry, opts Options) (*Plan, error) {
	if err := schema.CheckInitialized(kt, o); err != nil {
		logs.Errorf("check migration tables failed, err: %v, database: %s, rid: %s", err, reg.Database(), kt.Rid)
		return nil, err
	}

	records, err := schema.NewRecordStore(o).Load(kt)
	if err != nil {
		logs.Errorf("load migration records failed, err: %v, database: %s, rid: %s", err, reg.Database(), kt.Rid)
		return nil, err
	}

	current, hasReleasedVersion, err := schema.CurrentVersion(records)
	if err != nil {
		logs.Errorf("get current migration version failed, err: %v, database: %s, rid: %s", err, reg.Database(),
			kt.Rid)
		return nil, err
	}

	p := BuildPlan(reg.Database(), reg.All(), records, current, hasReleasedVersion, opts)
	logs.Infof("build migration plan success, database: %s, current: %s, items: %d, issues: %d, warnings: %d, "+
		"rid: %s", p.Database, p.CurrentRaw(), len(p.Items), len(p.Issues), len(p.Warnings), kt.Rid)
	return p, nil
}

// CollectPlanErrors returns nil when every plan passed. Otherwise it logs every
// issue of every plan and returns one error listing them all, wrapping only the
// sentinel of the highest priority: migrate.ErrIDReuse, then migrate.ErrRegistry, then
// migrate.ErrMissed.
func CollectPlanErrors(kt *kit.Kit, plans []*Plan) error {
	kinds := make(map[enumor.MigrationIssueKind]bool)
	var lines []string
	for _, p := range plans {
		for _, issue := range p.Issues {
			logs.Errorf("check migration plan failed, kind: %s, issue: %s, database: %s, rid: %s", issue.Kind,
				issue.Message, p.Database, kt.Rid)
			kinds[issue.Kind] = true
			lines = append(lines, fmt.Sprintf("[%s] %s", p.Database, issue.Message))
		}
	}
	if len(lines) == 0 {
		return nil
	}

	var sentinel error
	switch {
	case kinds[enumor.MigrationIssueIDReuse]:
		sentinel = migrate.ErrIDReuse
	case kinds[enumor.MigrationIssuePending], kinds[enumor.MigrationIssueLabels]:
		sentinel = migrate.ErrRegistry
	default:
		sentinel = migrate.ErrMissed
	}
	return fmt.Errorf("%w: %d problems found before execution, %s", sentinel, len(lines), strings.Join(lines, "; "))
}

// ExecuteResult is the outcome of one database in a run.
type ExecuteResult struct {
	Database      string
	VersionBefore string
	VersionAfter  string
	// Executed lists the IDs of the migrations run successfully, in order.
	Executed []string
	Skipped  []schema.SkippedMigration
	Warnings []string
	// Messages are the issues of the plan and the error that stopped it.
	Messages []string
	// Err is the error that stopped this database, nil on success.
	Err error
}

// AuditResult returns r as the audit result of a run ending with exitCode.
func (r *ExecuteResult) AuditResult(exitCode int) schema.AuditResult {
	return schema.AuditResult{
		ExitCode:      exitCode,
		VersionBefore: r.VersionBefore,
		VersionAfter:  r.VersionAfter,
		Skipped:       r.Skipped,
		Warnings:      r.Warnings,
		Messages:      r.Messages,
	}
}

// NewBlockedResult returns the result of a plan left unexecuted because the
// checks of this run failed. cause is the error returned by CollectPlanErrors.
// A plan without issues of its own was blocked by another database.
func NewBlockedResult(p *Plan, cause error) *ExecuteResult {
	r := newResult(p)
	if p.Passed() {
		r.Messages = append(r.Messages, fmt.Sprintf("not executed, err: %v", cause))
	}
	r.Err = cause
	return r
}

// newResult returns a result holding the version, warnings and issues of p.
func newResult(p *Plan) *ExecuteResult {
	r := &ExecuteResult{
		Database:      p.Database,
		VersionBefore: p.CurrentRaw(),
		VersionAfter:  p.CurrentRaw(),
		Warnings:      slices.Clone(p.Warnings),
	}
	for _, issue := range p.Issues {
		r.Messages = append(r.Messages, issue.Message)
	}
	return r
}

// fail stops r with err.
func (r *ExecuteResult) fail(err error) *ExecuteResult {
	r.Err = err
	r.Messages = append(r.Messages, err.Error())
	return r
}

// Execute runs a passed plan in order and stops at the first error. Only
// EXECUTE items run migrations, each between a running record and a success
// or failed record; SKIP-BACKFILL items only rewrite the recorded version.
// There is no outer transaction and no retry: a failed or interrupted
// migration runs again on the next run.
func Execute(kt *kit.Kit, o orm.Interface, p *Plan) *ExecuteResult {
	r := newResult(p)
	if !p.Passed() {
		return r.fail(fmt.Errorf("plan of database %s has %d issues, not executed", p.Database, len(p.Issues)))
	}

	for _, w := range p.Warnings {
		logs.Warnf("%s, database: %s, rid: %s", w, p.Database, kt.Rid)
	}

	store := schema.NewRecordStore(o)
	after, hasAfter := p.Current, p.HasReleasedVersion
	raise := func(m register.Migration) {
		if v, ok := m.ParsedVersion(); ok && (!hasAfter || register.Compare(v, after) > 0) {
			after, hasAfter = v, true
			r.VersionAfter = v.Raw
		}
	}

	for _, item := range p.Items {
		m := item.Migration
		switch item.Action {
		case enumor.MigrationActionExecute:
			if err := run(kt, o, store, m); err != nil {
				return r.fail(err)
			}
			r.Executed = append(r.Executed, m.ID)
			raise(m)
		case enumor.MigrationActionSkipBackfill:
			if err := store.BackfillPendingVersion(kt, m); err != nil {
				return r.fail(err)
			}
			r.skip(kt, item, enumor.MigrationSkipApplied)
			raise(m)
		case enumor.MigrationActionSkipSuccess:
			r.skip(kt, item, enumor.MigrationSkipApplied)
		case enumor.MigrationActionAboveMaxVersion:
			r.skip(kt, item, enumor.MigrationSkipAboveMaxVersion)
		default:
			return r.fail(fmt.Errorf("migration %s has action %s, not executable", m.ID, item.Action))
		}
	}

	logs.Infof("migrate database success, database: %s, executed: %d, skipped: %d, version before: %s, "+
		"version after: %s, rid: %s", p.Database, len(r.Executed), len(r.Skipped), r.VersionBefore, r.VersionAfter,
		kt.Rid)
	return r
}

// skip logs a skipped item and adds it to r.
func (r *ExecuteResult) skip(kt *kit.Kit, item PlanItem, reason enumor.MigrationSkipReason) {
	m := item.Migration
	logs.Infof("skip migration, reason: %s, id: %s, version: %s, pkg: %s, applied_pkg: %s, rid: %s", reason, m.ID,
		m.Version, m.Pkg, item.AppliedPkg, kt.Rid)
	r.Skipped = append(r.Skipped, schema.SkippedMigration{ID: m.ID, Version: m.Version, Pkg: m.Pkg,
		AppliedPkg: item.AppliedPkg, Reason: reason})
}

// run runs one migration and records its status.
func run(kt *kit.Kit, o orm.Interface, store *schema.RecordStore, m register.Migration) error {
	if err := store.MarkRunning(kt, m); err != nil {
		return err
	}

	logs.Infof("run migration start, id: %s, version: %s, pkg: %s, rid: %s", m.ID, m.Version, m.Pkg, kt.Rid)
	start := time.Now()
	if err := callUp(kt, o, m); err != nil {
		if markErr := store.MarkFailed(kt, m, err); markErr != nil {
			return fmt.Errorf("run migration %s failed, err: %v, and mark failed failed, err: %v", m.ID, err, markErr)
		}
		return fmt.Errorf("run migration %s failed, err: %v", m.ID, err)
	}

	if err := store.MarkSuccess(kt, m); err != nil {
		return err
	}
	logs.Infof("run migration success, id: %s, version: %s, cost: %s, rid: %s", m.ID, m.Version,
		time.Since(start), kt.Rid)
	return nil
}

// callUp runs m.Up and turns a panic into an error, so the record ends as
// failed instead of staying running and the process exits with code 1.
func callUp(kt *kit.Kit, o orm.Interface, m register.Migration) (err error) {
	defer func() {
		if p := recover(); p != nil {
			logs.Errorf("migration panicked, err: %v, id: %s, stack: %s, rid: %s", p, m.ID, debug.Stack(), kt.Rid)
			err = fmt.Errorf("panic: %v", p)
		}
	}()
	return m.Up(kt.Ctx, o)
}
