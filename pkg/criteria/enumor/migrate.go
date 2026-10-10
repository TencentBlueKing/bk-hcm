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

package enumor

import "fmt"

// MigrationStatus is the status of a migration record or a migration audit row.
type MigrationStatus string

const (
	// MigrationStatusRunning is the in-progress status.
	MigrationStatusRunning MigrationStatus = "running"
	// MigrationStatusSuccess is the successful terminal status.
	MigrationStatusSuccess MigrationStatus = "success"
	// MigrationStatusFailed is the failed terminal status.
	MigrationStatusFailed MigrationStatus = "failed"
)

// Validate checks whether the migration status is one of the declared values.
func (s MigrationStatus) Validate() error {
	switch s {
	case MigrationStatusRunning, MigrationStatusSuccess, MigrationStatusFailed:
		return nil
	default:
		return fmt.Errorf("unsupported migration status: %s", s)
	}
}

// MigrationSkipReason is the reason a migration is skipped in one run.
type MigrationSkipReason string

const (
	// MigrationSkipApplied means the migration ID has a success record, or
	// an earlier migration of the same ID runs in the same run.
	MigrationSkipApplied MigrationSkipReason = "applied"
	// MigrationSkipAboveMaxVersion means the migration is above the version limit of the run.
	MigrationSkipAboveMaxVersion MigrationSkipReason = "above_max_version"
	// MigrationSkipBaseline means init adopt records the migration as success without running it.
	MigrationSkipBaseline MigrationSkipReason = "baseline"
)

// MigrationInitMode is the --mode of hcm-migrate init.
type MigrationInitMode string

const (
	// MigrationInitModeEmpty only creates the migration tables.
	MigrationInitModeEmpty MigrationInitMode = "empty"
	// MigrationInitModeAdopt creates the migration tables and records the
	// migrations at or below each database baseline as success.
	MigrationInitModeAdopt MigrationInitMode = "adopt"
)

// Validate checks whether the init mode is one of the declared values.
func (m MigrationInitMode) Validate() error {
	switch m {
	case MigrationInitModeEmpty, MigrationInitModeAdopt:
		return nil
	default:
		return fmt.Errorf("unsupported migration init mode: %q, want empty or adopt", m)
	}
}

// MigrationAction is the decision on one migration in a plan.
type MigrationAction string

const (
	// MigrationActionExecute runs the migration.
	MigrationActionExecute MigrationAction = "EXECUTE"
	// MigrationActionSkipSuccess skips the migration because its ID is applied.
	MigrationActionSkipSuccess MigrationAction = "SKIP-SUCCESS"
	// MigrationActionSkipBackfill skips the migration and replaces the recorded
	// PENDING version with the registered version.
	MigrationActionSkipBackfill MigrationAction = "SKIP-BACKFILL"
	// MigrationActionAboveMaxVersion leaves the migration out because it is above the version limit.
	MigrationActionAboveMaxVersion MigrationAction = "ABOVE-MAX-VERSION"
	// MigrationActionMissing marks a missed migration; the plan fails.
	MigrationActionMissing MigrationAction = "MISSING"
	// MigrationActionIDReuse marks a suspected ID reuse; the plan fails.
	MigrationActionIDReuse MigrationAction = "ID-REUSE"
	// MigrationActionPendingDenied marks a PENDING migration without --allow-pending;
	// the plan fails.
	MigrationActionPendingDenied MigrationAction = "PENDING-DENIED"
)

// MigrationIssueKind is the kind of a problem found before execution. It
// decides the exit code of the run.
type MigrationIssueKind string

const (
	// MigrationIssueMissed is a missed migration in the default mode, exit code 4.
	MigrationIssueMissed MigrationIssueKind = "missed"
	// MigrationIssuePending is a PENDING migration without --allow-pending, exit code 5.
	MigrationIssuePending MigrationIssueKind = "pending"
	// MigrationIssueLabels is more than one version line in a registry, exit code 5.
	// A numeric fourth segment is its own line.
	MigrationIssueLabels MigrationIssueKind = "labels"
	// MigrationIssueIDReuse is a migration skipped by an ID that a package with
	// another migration suffix applied, exit code 6.
	MigrationIssueIDReuse MigrationIssueKind = "id_reuse"
)
