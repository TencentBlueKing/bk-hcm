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
	"errors"
	"fmt"
	"io"

	"hcm/migrate/engine"
	"hcm/migrate/register"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/spf13/pflag"
)

// runStatus prints the progress of each selected database. It is read only:
// it creates no table and writes no record or audit row. It takes no mode;
// it builds the plan the default mode would, with PENDING allowed, only to
// mark the migrations at or below the current version.
func (r *runner) runStatus(kt *kit.Kit, args []string) error {
	var f globalOptions
	fs := pflag.NewFlagSet("status", pflag.ContinueOnError)
	f.bind(fs)
	if err := r.parse(fs, args); err != nil {
		return err
	}

	regs, err := r.selectRegistries(f.databases)
	if err != nil {
		return err
	}
	source, err := r.openSource(f.configFile)
	if err != nil {
		return err
	}

	// A database that is not initialized is reported and the others are still
	// shown; the command then exits with the first such error.
	var notInitialized error
	for _, reg := range regs {
		o, ok, err := source.Open(kt, reg)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		p, err := engine.Prepare(kt, o, reg, engine.Options{AllowPending: true})
		if err != nil {
			if !errors.Is(err, migrate.ErrPrecondition) {
				return err
			}
			fmt.Fprintf(r.stdout, "\ndatabase %s: %v\n", reg.Database(), err)
			if notInitialized == nil {
				notInitialized = err
			}
			continue
		}
		printStatus(r.stdout, p)
	}
	return notInitialized
}

// runList prints the registered migrations of each selected database in
// execution order. It needs no config file and no database connection;
// PENDING migrations are listed like the others.
func (r *runner) runList(args []string) error {
	var f globalOptions
	fs := pflag.NewFlagSet("list", pflag.ContinueOnError)
	f.bind(fs)
	if err := r.parse(fs, args); err != nil {
		return err
	}

	regs, err := r.selectRegistries(f.databases)
	if err != nil {
		return err
	}
	for _, reg := range regs {
		printList(r.stdout, reg)
	}
	return nil
}

// printStatus prints the progress of one database: the current version, the
// record count of each status, and the migrations without a success record.
// Those at or below the current version are marked; the default mode of up
// fails on them.
func printStatus(w io.Writer, p *engine.Plan) {
	counts := make(map[enumor.MigrationStatus]int)
	for _, rec := range p.Records {
		counts[rec.Status]++
	}
	fmt.Fprintf(w, "\ndatabase %s, current version: %s\n", p.Database, orNone(p.CurrentRaw()))
	fmt.Fprintf(w, "  records: success %d, failed %d, running %d\n", counts[enumor.MigrationStatusSuccess],
		counts[enumor.MigrationStatusFailed], counts[enumor.MigrationStatusRunning])

	fmt.Fprintln(w, "  not executed:")
	tw := newTable(w)
	for _, item := range p.Items {
		m := item.Migration
		switch item.Action {
		case enumor.MigrationActionExecute:
			fmt.Fprintf(tw, "    %s\t%s\t%s\n", m.ID, m.Version, m.Pkg)
		case enumor.MigrationActionMissing:
			fmt.Fprintf(tw, "    %s\t%s\t%s\t<= current version\n", m.ID, m.Version, m.Pkg)
		}
	}
	tw.Flush()
}

// printList prints the registered migrations of one registry in execution order.
func printList(w io.Writer, reg *register.Registry) {
	all := reg.All()
	fmt.Fprintf(w, "\ndatabase %s, %d migrations\n", reg.Database(), len(all))
	if len(all) == 0 {
		return
	}
	tw := newTable(w)
	fmt.Fprintln(tw, "  #\tID\tVERSION\tTIMESTAMP\tPKG")
	for i, m := range all {
		fmt.Fprintf(tw, "  %d\t%s\t%s\t%s\t%s\n", i+1, m.ID, m.Version, m.Timestamp, m.Pkg)
	}
	tw.Flush()
}
