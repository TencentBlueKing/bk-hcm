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
	"strings"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/migrate"
	"hcm/pkg/tools/json"
	"hcm/pkg/version"
)

// AuditMeta describes the run being audited.
type AuditMeta struct {
	// Command is the subcommand, enumeration values such as: up/init.
	Command string
	// Args is the command line after the program name, os.Args[1:].
	Args []string
}

// SkippedMigration is one item of the skipped column.
type SkippedMigration struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Pkg     string `json:"pkg"`
	// AppliedPkg is the recorded applied_pkg for enumor.MigrationSkipApplied, empty otherwise.
	AppliedPkg string                     `json:"applied_pkg"`
	Reason     enumor.MigrationSkipReason `json:"reason"`
}

// AuditResult is the result of a run written by End.
type AuditResult struct {
	// ExitCode is the process exit code. Zero ends the row as success; any
	// other value ends it as failed.
	ExitCode      int
	VersionBefore string
	VersionAfter  string
	Skipped       []SkippedMigration
	// Warnings are the warnings of the run, in the order they happened.
	Warnings []string
	// Messages are the validation problems and errors, in the order they
	// happened. They go to the message column.
	Messages []string
}

// Audit is the audit row of one run on one database. A nil *Audit is valid
// and does nothing.
type Audit struct {
	o     orm.Interface
	runID string
	// warnings are collected by Begin and prepended by End to result.Warnings.
	warnings []string
}

// NewAudit returns the audit of this run, keyed by kt.Rid. It returns nil
// when kt.Rid is empty.
func NewAudit(kt *kit.Kit, o orm.Interface) *Audit {
	if kt.Rid == "" {
		logs.Warnf("skip migration audit, rid is empty")
		return nil
	}
	return &Audit{o: o, runID: kt.Rid}
}

// Begin inserts the running row of this audit. It returns nil when the row
// cannot be written; failures are only logged. A row of an earlier run still
// in running is reported as a warning of this run. Calling Begin on a nil
// *Audit returns nil.
func (a *Audit) Begin(kt *kit.Kit, meta AuditMeta) *Audit {
	if a == nil {
		return nil
	}

	if prev, found := lastRunningAudit(kt, a.o); found {
		msg := fmt.Sprintf("previous migration run did not end, run_id: %s", prev)
		logs.Warnf("%s, rid: %s", msg, kt.Rid)
		a.warnings = append(a.warnings, msg)
	}

	expr := fmt.Sprintf("INSERT INTO `%s` (`run_id`, `command`, `args`, `binary_version`, `git_hash`, `status`) "+
		"VALUES (:run_id, :command, :args, :binary_version, :git_hash, :status)", constant.MigrationAuditTable)
	arg := map[string]interface{}{
		"run_id":         a.runID,
		"command":        meta.Command,
		"args":           migrate.TruncateUTF8(strings.Join(meta.Args, " "), constant.MigrationAuditArgsMaxBytes),
		"binary_version": migrate.TruncateUTF8(version.VERSION, constant.MigrationAuditBuildMaxBytes),
		"git_hash":       migrate.TruncateUTF8(version.GITHASH, constant.MigrationAuditBuildMaxBytes),
		"status":         enumor.MigrationStatusRunning,
	}
	if err := a.o.Do().Insert(kt.Ctx, expr, arg); err != nil {
		logs.Warnf("insert migration audit failed, skip audit, err: %v, rid: %s", err, kt.Rid)
		return nil
	}

	return a
}

// lastRunningAudit returns the run_id of the latest row still in running.
func lastRunningAudit(kt *kit.Kit, o orm.Interface) (string, bool) {
	expr := fmt.Sprintf("SELECT `run_id` FROM `%s` WHERE `status` = :status ORDER BY `id` DESC LIMIT 1",
		constant.MigrationAuditTable)
	rows := make([]struct {
		RunID string `db:"run_id"`
	}, 0)
	arg := map[string]interface{}{"status": enumor.MigrationStatusRunning}
	if err := o.Do().Select(kt.Ctx, &rows, expr, arg); err != nil {
		logs.Warnf("query running migration audit failed, err: %v, rid: %s", err, kt.Rid)
		return "", false
	}
	if len(rows) == 0 {
		return "", false
	}
	return rows[0].RunID, true
}

// End writes the result into the row in one update. Failures are only
// logged. Calling End on a nil *Audit does nothing.
func (a *Audit) End(kt *kit.Kit, result AuditResult) {
	if a == nil {
		return
	}

	status := enumor.MigrationStatusSuccess
	if result.ExitCode != 0 {
		status = enumor.MigrationStatusFailed
	}

	skipped := result.Skipped
	if skipped == nil {
		skipped = make([]SkippedMigration, 0)
	}
	warnings := append(append(make([]string, 0, len(a.warnings)+len(result.Warnings)), a.warnings...),
		result.Warnings...)

	skippedJSON, err := json.MarshalToString(skipped)
	if err != nil {
		logs.Warnf("marshal migration audit skipped failed, err: %v, rid: %s", err, kt.Rid)
		return
	}
	warningsJSON, err := json.MarshalToString(capAuditItems(warnings))
	if err != nil {
		logs.Warnf("marshal migration audit warnings failed, err: %v, rid: %s", err, kt.Rid)
		return
	}
	messageJSON, err := json.MarshalToString(capAuditItems(result.Messages))
	if err != nil {
		logs.Warnf("marshal migration audit message failed, err: %v, rid: %s", err, kt.Rid)
		return
	}

	expr := fmt.Sprintf("UPDATE `%s` SET `status` = :status, `exit_code` = :exit_code, "+
		"`version_before` = :version_before, `version_after` = :version_after, `skipped` = :skipped, "+
		"`warnings` = :warnings, `message` = :message, `end_at` = NOW() WHERE `run_id` = :run_id",
		constant.MigrationAuditTable)
	arg := map[string]interface{}{
		"status":         status,
		"exit_code":      result.ExitCode,
		"version_before": result.VersionBefore,
		"version_after":  result.VersionAfter,
		"skipped":        skippedJSON,
		"warnings":       warningsJSON,
		"message":        messageJSON,
		"run_id":         a.runID,
	}
	affected, err := a.o.Do().Update(kt.Ctx, expr, arg)
	if err != nil {
		logs.Warnf("update migration audit failed, err: %v, rid: %s", err, kt.Rid)
		return
	}
	if affected == 0 {
		logs.Warnf("update migration audit affected no row, rid: %s", kt.Rid)
	}
}

// capAuditItems truncates every item to constant.MigrationAuditItemMaxBytes
// and keeps at most constant.MigrationAuditMaxItems items, the last one
// replaced by a count of those left out. It never returns nil.
func capAuditItems(items []string) []string {
	n := len(items)
	if n > constant.MigrationAuditMaxItems {
		n = constant.MigrationAuditMaxItems - 1
	}

	out := make([]string, 0, min(len(items), constant.MigrationAuditMaxItems))
	for _, item := range items[:n] {
		out = append(out, migrate.TruncateUTF8(item, constant.MigrationAuditItemMaxBytes))
	}
	if n < len(items) {
		out = append(out, fmt.Sprintf("... %d more omitted", len(items)-n))
	}
	return out
}
