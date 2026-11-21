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
	"slices"
	"strings"

	"hcm/migrate/register"
	"hcm/migrate/schema"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/spf13/pflag"
)

// initOptions are the options of the init command.
type initOptions struct {
	globalOptions
	mode      string
	baselines []string
	planOnly  bool
}

// runInit creates the migration tables of each selected database. With
// --mode=adopt, a database given a --baseline and whose record table this
// call creates gets the migrations at or below the baseline recorded as
// success; a selected database without a baseline only gets empty tables.
func (r *runner) runInit(kt *kit.Kit, args []string) error {
	var f initOptions
	fs := pflag.NewFlagSet("init", pflag.ContinueOnError)
	f.bind(fs)
	fs.StringVar(&f.mode, "mode", "", "required, empty: only create the tables; adopt: also record a baseline")
	fs.StringArrayVar(&f.baselines, "baseline", nil, "database=version, repeatable, required by --mode=adopt")
	fs.BoolVar(&f.planOnly, "plan", false, "print the tables to create and the migrations to record only")
	if err := r.parse(fs, args); err != nil {
		return err
	}

	// 连库前校验，避免建了一半才发现参数不对。
	baselines, err := r.parseInitOptions(f)
	if err != nil {
		return err
	}
	// 按 --database 选出要处理的库，不传则是已启用的全部库。
	regs, err := r.selectRegistries(f.databases)
	if err != nil {
		return err
	}
	// 读配置文件里的库连接。
	source, err := r.openSource(f.configFile)
	if err != nil {
		return err
	}

	// 逐库建表，某个库失败则后面的库不再处理。
	for _, reg := range regs {
		o, ok, err := source.Open(kt, reg)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		baseline := baselines[reg.Database()]
		if f.planOnly {
			if err := r.planInit(kt, o, reg, baseline); err != nil {
				return err
			}
			continue
		}
		// 建表，并写本库的审计。
		if err := r.initDatabase(kt, o, reg, baseline); err != nil {
			return err
		}
	}
	return nil
}

// initDatabase runs InitTables on one database and writes its audit row. The
// audit table may only exist after InitTables, so the row is begun then.
func (r *runner) initDatabase(kt *kit.Kit, o orm.Interface, reg *register.Registry,
	baseline *register.Version) error {

	res, err := schema.InitTables(kt, o, baseline, reg.All())
	// 审计表可能刚建出来，所以 Begin 放在 InitTables 之后，并立刻按本库结果 End。
	audit := schema.NewAudit(kt, o).Begin(kt, schema.AuditMeta{Command: "init", Args: r.args})

	result := schema.AuditResult{ExitCode: migrate.ExitCode(err)}
	for _, m := range res.Adopted {
		result.Skipped = append(result.Skipped, schema.SkippedMigration{ID: m.ID, Version: m.Version, Pkg: m.Pkg,
			Reason: enumor.MigrationSkipBaseline})
	}
	if err != nil {
		err = fmt.Errorf("init database %s failed, err: %w", reg.Database(), err)
		result.Messages = []string{err.Error()}
	}
	audit.End(kt, result)
	if err != nil {
		return err
	}

	if !res.AuditCreated && !res.RecordCreated {
		fmt.Fprintf(r.stdout, "database %s: already initialized, nothing changed\n", reg.Database())
		return nil
	}
	fmt.Fprintf(r.stdout, "database %s: created audit table: %t, created record table: %t, "+
		"recorded %d baseline migrations\n", reg.Database(), res.AuditCreated, res.RecordCreated, len(res.Adopted))
	return nil
}

// parseInitOptions checks --mode and --baseline and returns the baseline of
// each named database.
func (r *runner) parseInitOptions(f initOptions) (map[string]*register.Version, error) {
	mode := enumor.MigrationInitMode(f.mode)
	if f.mode == "" {
		return nil, fmt.Errorf("%w: --mode is required, want empty or adopt", migrate.ErrUsage)
	}
	if err := mode.Validate(); err != nil {
		return nil, fmt.Errorf("%w: --mode: %v", migrate.ErrUsage, err)
	}
	if mode == enumor.MigrationInitModeEmpty && len(f.baselines) > 0 {
		return nil, fmt.Errorf("%w: --baseline only applies to --mode=adopt", migrate.ErrUsage)
	}
	if mode == enumor.MigrationInitModeAdopt && len(f.baselines) == 0 {
		return nil, fmt.Errorf("%w: --mode=adopt needs at least one --baseline database=version",
			migrate.ErrUsage)
	}

	known, err := r.selectRegistries(nil)
	if err != nil {
		return nil, err
	}
	names := make(map[string]struct{}, len(known))
	for _, reg := range known {
		names[reg.Database()] = struct{}{}
	}

	baselines := make(map[string]*register.Version, len(f.baselines))
	for _, value := range f.baselines {
		name, raw, found := strings.Cut(value, "=")
		name, raw = strings.TrimSpace(name), strings.TrimSpace(raw)
		if !found || name == "" || raw == "" {
			return nil, fmt.Errorf("%w: --baseline %q, want database=version", migrate.ErrUsage, value)
		}
		if _, ok := names[name]; !ok {
			return nil, fmt.Errorf("%w: --baseline %q names unknown database %q", migrate.ErrUsage, value, name)
		}
		v, err := parseVersionFlag("--baseline "+name, raw)
		if err != nil {
			return nil, err
		}
		baselines[name] = &v
	}
	return baselines, nil
}

// planInit prints what init would do on one database. It writes nothing.
func (r *runner) planInit(kt *kit.Kit, o orm.Interface, reg *register.Registry, baseline *register.Version) error {
	missing, err := schema.MissingTables(kt, o)
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		fmt.Fprintf(r.stdout, "\ndatabase %s: already initialized, init would change nothing\n", reg.Database())
		return nil
	}
	fmt.Fprintf(r.stdout, "\ndatabase %s: would create %s\n", reg.Database(), strings.Join(missing, ", "))

	// Baseline records are written only when init creates the record table.
	if baseline == nil || !slices.Contains(missing, constant.MigrationRecordTable) {
		fmt.Fprintln(r.stdout, "  would record no baseline migrations")
		return nil
	}
	adopted := schema.BaselineMigrations(*baseline, reg.All())
	fmt.Fprintf(r.stdout, "  would record %d migrations at or below %s as success:\n", len(adopted), baseline.Raw)
	tw := newTable(r.stdout)
	fmt.Fprintln(tw, "    ID\tVERSION\tPKG")
	for _, m := range adopted {
		fmt.Fprintf(tw, "    %s\t%s\t%s\n", m.ID, m.Version, m.Pkg)
	}
	tw.Flush()
	return nil
}
