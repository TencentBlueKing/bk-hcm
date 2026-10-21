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

package cli

import (
	"fmt"
	"io"

	"hcm/migrate/engine"
	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/spf13/pflag"
)

// migrateOptions are the options of the up command.
type migrateOptions struct {
	globalOptions
	to       string
	catchUp  bool
	planOnly bool
}

// migrateRun is one database in an up command: its connection, audit, plan and result.
type migrateRun struct {
	reg    *register.Registry
	o      orm.Interface
	audit  *schema.Audit
	plan   *engine.Plan
	result *engine.ExecuteResult
}

// runMigrate runs the up command: prepare every database, then execute only
// when every plan passed.
func (r *runner) runMigrate(kt *kit.Kit, args []string) error {
	var f migrateOptions
	fs := pflag.NewFlagSet("up", pflag.ContinueOnError)
	f.bind(fs)
	fs.StringVar(&f.to, "to", "", "version limit, only migrations at or below it run; empty means no limit")
	fs.BoolVar(&f.catchUp, "catch-up", false, "run missed migrations below the database version")
	fs.BoolVar(&f.planOnly, "plan", false, "print the plan only, write nothing and run nothing")
	if err := r.parse(fs, args); err != nil {
		return err
	}

	regs, err := r.selectRegistries(f.databases)
	if err != nil {
		return err
	}
	opts := engine.Options{CatchUp: f.catchUp, AllowPending: f.allowPending}
	if f.to != "" {
		ceiling, err := parseVersionFlag("--to", f.to)
		if err != nil {
			return err
		}
		opts.Ceiling = &ceiling
	}
	r.printMigrateMode(f)

	source, err := r.openSource(f.configFile)
	if err != nil {
		return err
	}

	// 先全部连库并出计划，任一库失败即停。
	runs, err := r.prepareMigrate(kt, source, regs, opts, f.planOnly)
	if err != nil {
		r.endMigrateAudits(kt, runs, err)
		return err
	}

	plans := make([]*engine.Plan, 0, len(runs))
	for _, run := range runs {
		printPlan(r.stdout, run.plan)
		plans = append(plans, run.plan)
	}
	//plan有问题则一个库都不执行，退出码取 6 > 5 > 4。
	if err := engine.CollectPlanErrors(kt, plans); err != nil {
		for _, run := range runs {
			run.result = engine.NewBlockedResult(run.plan, err)
		}
		r.endMigrateAudits(kt, runs, err)
		return err
	}
	// 只打印计划，不写审计、不跑迁移。
	if f.planOnly {
		return nil
	}

	// 按库顺序执行，失败即停；审计结束行写进程退出码。
	err = r.executeMigrate(kt, runs)
	printMigrateSummary(r.stdout, runs)
	r.endMigrateAudits(kt, runs, err)
	return err
}

// printMigrateMode prints the mode the up command received, so a wrongly enabled
// switch shows in the job log.
func (r *runner) printMigrateMode(f migrateOptions) {
	mode := "default"
	if f.catchUp {
		mode = "catch-up"
	}
	fmt.Fprintf(r.stdout, "hcm-migrate up: mode: %s, allow pending: %t, to: %s, plan only: %t\n",
		mode, f.allowPending, orNone(f.to), f.planOnly)
}

// prepareMigrate connects to each database, begins its audit row, and builds its
// plan. It stops at the first database that fails; the runs before it are
// returned with a result that says why they did not run.
func (r *runner) prepareMigrate(kt *kit.Kit, source dataSource, regs []*register.Registry, opts engine.Options,
	planOnly bool) ([]*migrateRun, error) {

	runs := make([]*migrateRun, 0, len(regs))
	for _, reg := range regs {
		o, ok, err := source.Open(kt, reg)
		if err != nil {
			blockRuns(runs, err)
			return runs, err
		}
		if !ok {
			continue
		}

		run := &migrateRun{reg: reg, o: o}
		if !planOnly {
			run.audit = schema.NewAudit(kt, o).Begin(kt, schema.AuditMeta{Command: "up", Args: r.args})
		}
		runs = append(runs, run)

		run.plan, err = engine.Prepare(kt, o, reg, opts)
		if err != nil {
			run.result = &engine.ExecuteResult{Database: reg.Database(), Messages: []string{err.Error()}, Err: err}
			blockRuns(runs[:len(runs)-1], err)
			return runs, err
		}
	}
	return runs, nil
}

// executeMigrate executes the plans in order and stops at the first database that
// fails. The databases after it are not executed.
func (r *runner) executeMigrate(kt *kit.Kit, runs []*migrateRun) error {
	for i, run := range runs {
		run.result = engine.Execute(kt, run.o, run.plan)
		if run.result.Err != nil {
			err := fmt.Errorf("migrate database %s failed, err: %w", run.reg.Database(), run.result.Err)
			blockRuns(runs[i+1:], err)
			return err
		}
	}
	return nil
}

// blockRuns sets the result of every run left unexecuted because of cause.
func blockRuns(runs []*migrateRun, cause error) {
	for _, run := range runs {
		run.result = engine.NewBlockedResult(run.plan, cause)
	}
}

// endMigrateAudits ends every audit row begun with the process exit code of err.
func (r *runner) endMigrateAudits(kt *kit.Kit, runs []*migrateRun, err error) {
	code := migrate.ExitCode(err)
	for _, run := range runs {
		if run.audit != nil && run.result != nil {
			run.audit.End(kt, run.result.AuditResult(code))
		}
	}
}

// printPlan prints the decision on every migration of one database, then its
// issues and warnings.
func printPlan(w io.Writer, p *engine.Plan) {
	fmt.Fprintf(w, "\ndatabase %s, current version: %s\n", p.Database, orNone(p.CurrentRaw()))
	if len(p.Items) == 0 {
		fmt.Fprintln(w, "  no registered migrations")
		return
	}
	tw := newTable(w)
	fmt.Fprintln(tw, "  ACTION\tID\tVERSION\tPKG\tAPPLIED_PKG")
	for _, item := range p.Items {
		m := item.Migration
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\n", item.Action, m.ID, m.Version, m.Pkg, item.AppliedPkg)
	}
	tw.Flush()
	printLines(w, "issues", issueMessages(p.Issues))
	printLines(w, "warnings", p.Warnings)
}

// printMigrateSummary prints the outcome of every database of the up command.
func printMigrateSummary(w io.Writer, runs []*migrateRun) {
	fmt.Fprintln(w, "\nsummary:")
	for _, run := range runs {
		res := run.result
		if res == nil {
			continue
		}
		if res.Err != nil {
			fmt.Fprintf(w, "  database %s: failed, executed: %d, version: %s -> %s, err: %v\n", res.Database,
				len(res.Executed), orNone(res.VersionBefore), orNone(res.VersionAfter), res.Err)
			continue
		}
		fmt.Fprintf(w, "  database %s: success, executed: %d, skipped: %d, version: %s -> %s\n", res.Database,
			len(res.Executed), len(res.Skipped), orNone(res.VersionBefore), orNone(res.VersionAfter))
	}
}

// printLines prints a titled list, nothing when it is empty.
func printLines(w io.Writer, title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	fmt.Fprintf(w, "  %s:\n", title)
	for _, line := range lines {
		fmt.Fprintf(w, "    - %s\n", line)
	}
}

func issueMessages(issues []engine.Issue) []string {
	out := make([]string, 0, len(issues))
	for _, issue := range issues {
		out = append(out, issue.Message)
	}
	return out
}
