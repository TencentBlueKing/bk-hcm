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
	"fmt"
	"slices"
	"strings"

	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/migrate"
)

// Options are the up options that change the plan.
type Options struct {
	// Ceiling is the --to version limit, nil when there is no limit.
	Ceiling *register.Version
	// CatchUp turns off the missed migration check (--catch-up).
	CatchUp bool
	// AllowPending lets PENDING migrations run (--allow-pending).
	AllowPending bool
}

// Issue is one problem found before execution. Every issue of every database
// is reported before anything runs.
type Issue struct {
	Kind enumor.MigrationIssueKind
	// Message is one line naming the full migration ID and what is wrong.
	Message string
}

// PlanItem is the decision on one registered migration.
type PlanItem struct {
	Migration register.Migration
	Action    enumor.MigrationAction
	// AppliedPkg is the package owning the migration ID, set for skip and
	// ID-REUSE items: the recorded applied_pkg, or an earlier copy of the ID
	// planned in this run.
	AppliedPkg string
}

// Plan is the checked plan of one database.
type Plan struct {
	Database string
	// Current is the database version at the start of the run. It is valid
	// only when HasReleasedVersion is true.
	Current register.Version
	// HasReleasedVersion is true when at least one success record has a released
	// version. False means there is no database version yet: the record table
	// is empty, or it holds only running, failed, and PENDING rows. The
	// missed-migration check applies only when HasReleasedVersion is true.
	HasReleasedVersion bool
	// Records are the records the plan was built from.
	Records schema.Records
	// Items follow the execution order, one per registered migration.
	Items    []PlanItem
	Issues   []Issue
	Warnings []string
}

// Passed reports whether the plan has no issue and may be executed.
func (p *Plan) Passed() bool {
	return len(p.Issues) == 0
}

// CurrentRaw returns the raw current version, empty when there is none.
func (p *Plan) CurrentRaw() string {
	if !p.HasReleasedVersion {
		return ""
	}
	return p.Current.Raw
}

// idOwner is the package that ran, or is planned to run, a migration ID.
// Every later migration with the ID is skipped against it.
type idOwner struct {
	pkg string
	// version is the version recorded for the ID after this plan, updated when
	// a planned backfill replaces PENDING.
	version string
	// recorded is true when this owner is a success record. False means an
	// earlier copy in this walk claimed the ID.
	recorded bool
}

// BuildPlan checks one database and decides every migration. migrations must
// be in execution order, as returned by Registry.All. current is the version
// computed by schema.CurrentVersion once for this run.
//
// A migration ID is owned by the package in its success record, or else by
// its first copy planned to run. Every later copy is skipped against that
// owner, and a different migration suffix there is a suspected ID reuse, so
// a copy in the same registry and a copy applied by an earlier build are
// checked by the same rule.
func BuildPlan(database string, migrations []register.Migration, records schema.Records, current register.Version,
	hasReleasedVersion bool, opts Options) *Plan {

	p := &Plan{Database: database, Current: current, HasReleasedVersion: hasReleasedVersion, Records: records}
	p.Issues = append(p.Issues, labelIssues(migrations)...)

	owners := make(map[string]*idOwner, len(records))
	for id, r := range records {
		if r.Status == enumor.MigrationStatusSuccess {
			owners[id] = &idOwner{pkg: r.AppliedPkg, version: r.Version, recorded: true}
		}
	}

	for _, m := range migrations {
		p.Items = append(p.Items, p.checkOne(m, owners, opts))
	}
	return p
}

// checkOne checks one migration. It returns the plan item and records any
// issue or warning on p. Label problems are not checked here; BuildPlan
// collects those once for the whole registry.
func (p *Plan) checkOne(m register.Migration, owners map[string]*idOwner, opts Options) PlanItem {
	item := PlanItem{Migration: m}

	if m.IsPending() && !opts.AllowPending {
		item.Action = enumor.MigrationActionPendingDenied
		p.addIssue(enumor.MigrationIssuePending, fmt.Sprintf(
			"pending migration is not allowed without --allow-pending, id: %s, pkg: %s", m.ID, m.Pkg))
		return item
	}

	if aboveMaxVersion(m, opts.Ceiling) {
		item.Action = enumor.MigrationActionAboveMaxVersion
		return item
	}

	if owner, found := owners[m.ID]; found {
		return p.skipByID(m, owner)
	}

	switch {
	case m.IsPending():
		// Allowed PENDING belongs to no release, so it is not compared with the
		// current version. It reaches here only with --allow-pending, and only
		// when no ceiling excluded it.
		item.Action = enumor.MigrationActionExecute
	case opts.CatchUp:
		// Catch-up runs a released migration that the default mode would call missed.
		item.Action = enumor.MigrationActionExecute
	case p.missed(m):
		item.Action = enumor.MigrationActionMissing
		p.addIssue(enumor.MigrationIssueMissed, fmt.Sprintf(
			"missed migration, id: %s, version: %s, current: %s, pkg: %s",
			m.ID, m.Version, p.Current.Raw, m.Pkg))
	default:
		item.Action = enumor.MigrationActionExecute
	}
	// A missed migration also owns its ID, so later copies are still checked
	// and every issue is reported at once.
	owners[m.ID] = &idOwner{pkg: m.Pkg, version: m.Version}
	return item
}

// skipByID decides a migration whose ID is already owned.
func (p *Plan) skipByID(m register.Migration, owner *idOwner) PlanItem {
	item := PlanItem{Migration: m, AppliedPkg: owner.pkg}

	if !sameSuffix(owner.pkg, m.Pkg) {
		item.Action = enumor.MigrationActionIDReuse
		if owner.recorded {
			p.addIssue(enumor.MigrationIssueIDReuse, fmt.Sprintf(
				"suspected migration id reuse, id: %s, recorded version: %s, applied_pkg: %s, pkg: %s",
				m.ID, owner.version, owner.pkg, m.Pkg))
		} else {
			p.addIssue(enumor.MigrationIssueIDReuse, fmt.Sprintf(
				"suspected migration id reuse, id: %s, registered by pkg: %s and pkg: %s",
				m.ID, owner.pkg, m.Pkg))
		}
		return item
	}

	if register.IsPending(owner.version) && !m.IsPending() {
		item.Action = enumor.MigrationActionSkipBackfill
		// Later copies of this ID in this walk compare against the released
		// version, so a different version is drift instead of another backfill.
		owner.version = m.Version
		return item
	}

	item.Action = enumor.MigrationActionSkipSuccess
	if owner.version != m.Version {
		p.Warnings = append(p.Warnings, fmt.Sprintf("migration version drift, id: %s, recorded: %s, "+
			"registered: %s, applied_pkg: %s, pkg: %s", m.ID, owner.version, m.Version, owner.pkg, m.Pkg))
	}
	return item
}

func (p *Plan) addIssue(kind enumor.MigrationIssueKind, msg string) {
	p.Issues = append(p.Issues, Issue{Kind: kind, Message: msg})
}

// labelIssues reports every version line when the registry holds more than
// one. A fourth segment with an empty label is a line too. A version with no
// fourth segment is not. The order between two lines exists only to make
// Compare total, so such a registry must not run.
func labelIssues(migrations []register.Migration) []Issue {
	labels := make(map[string][]string)
	for _, m := range migrations {
		v, ok := m.ParsedVersion()
		if !ok || v.Suffix == nil {
			continue
		}
		labels[v.Suffix.Label] = append(labels[v.Suffix.Label], m.ID)
	}
	if len(labels) < 2 {
		return nil
	}

	names := make([]string, 0, len(labels))
	for name := range labels {
		names = append(names, name)
	}
	slices.Sort(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s: [%s]", name, strings.Join(labels[name], ", ")))
	}
	return []Issue{{Kind: enumor.MigrationIssueLabels, Message: fmt.Sprintf(
		"more than one version label in registry, %s", strings.Join(parts, "; "))}}
}

// aboveMaxVersion reports whether m is out of --to. A PENDING migration belongs
// to no release and is above any version limit.
func aboveMaxVersion(m register.Migration, ceiling *register.Version) bool {
	if ceiling == nil {
		return false
	}
	v, ok := m.ParsedVersion()
	return !ok || register.Compare(v, *ceiling) > 0
}

// missed reports whether a released migration is at or below the database
// version. PENDING never reaches here. A migration whose version does not
// parse is not treated as missed.
func (p *Plan) missed(m register.Migration) bool {
	v, ok := m.ParsedVersion()
	return ok && p.HasReleasedVersion && register.Compare(v, p.Current) <= 0
}

// sameSuffix reports whether two migration packages share a migration suffix.
// A package whose suffix cannot be read never matches.
func sameSuffix(a, b string) bool {
	sa, okA := migrate.PkgSuffix(a)
	sb, okB := migrate.PkgSuffix(b)
	return okA && okB && sa == sb
}
