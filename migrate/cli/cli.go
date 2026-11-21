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

// Package cli is the command line of hcm-migrate: init, up, status and list.
package cli

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"hcm/migrate/engine"
	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/migrate"

	"github.com/spf13/pflag"
)

const usage = `Usage: hcm-migrate <command> [flags]

Commands:
  init     create the migration tables, and record a baseline on an adopted database
  up       run the pending migrations
  status   show the migration progress of each database
  list     list the registered migrations in execution order

Global flags:
  -c, --config-file string   data-service config file
  -d, --database strings     databases to use, repeatable or comma separated, default all
      --allow-pending        allow PENDING migrations to run

Run "hcm-migrate <command> --help" for the flags of a command.
`

// errHelp marks a --help request. It ends the command with exit code 0.
var errHelp = errors.New("help requested")

// dataSource opens the database of a registry. ok is false when the database
// is not configured.
type dataSource interface {
	Open(kt *kit.Kit, reg *register.Registry) (o orm.Interface, ok bool, err error)
}

// runner holds the dependencies of one command run.
type runner struct {
	stdout io.Writer
	stderr io.Writer
	// args is the command line after the program name, written to the audit.
	args []string
	// loadSource loads the database configs from the config file.
	loadSource func(configFile string) (dataSource, error)
	// selectRegistries returns the registries named by --database.
	selectRegistries func(values []string) ([]*register.Registry, error)
}

// Run runs the command in args, the command line after the program name, and
// returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	r := &runner{
		stdout: stdout,
		stderr: stderr,
		args:   args,
		loadSource: func(configFile string) (dataSource, error) {
			return engine.LoadDataSource(configFile)
		},
		selectRegistries: engine.SelectRegistries,
	}
	return r.run(args)
}

func (r *runner) run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(r.stderr, usage)
		return constant.MigrationExitUsage
	}

	// 一次命令一个 rid，各库审计共用。
	kt := kit.New()
	cmd, rest := args[0], args[1:]
	var err error
	switch cmd {
	case "init":
		err = r.runInit(kt, rest)
	case "up":
		err = r.runMigrate(kt, rest)
	case "status":
		err = r.runStatus(kt, rest)
	case "list":
		err = r.runList(rest)
	case "help", "-h", "--help":
		fmt.Fprint(r.stdout, usage)
		return constant.MigrationExitSuccess
	default:
		fmt.Fprintf(r.stderr, "unknown command %q\n\n%s", cmd, usage)
		return constant.MigrationExitUsage
	}

	// --help 退出码 0，不走错误码映射。
	if errors.Is(err, errHelp) {
		return constant.MigrationExitSuccess
	}
	// 其余错误按哨兵映射成退出码 0–6。
	code := migrate.ExitCode(err)
	if err != nil {
		logs.Errorf("hcm-migrate %s failed, err: %v, exit code: %d, rid: %s", cmd, err, code, kt.Rid)
	}
	return code
}

// globalOptions are the options every command accepts.
type globalOptions struct {
	configFile   string
	databases    []string
	allowPending bool
}

func (g *globalOptions) bind(fs *pflag.FlagSet) {
	fs.StringVarP(&g.configFile, "config-file", "c", "", "data-service config file")
	fs.StringArrayVarP(&g.databases, "database", "d", nil,
		"databases to use, repeatable or comma separated, default all")
	fs.BoolVar(&g.allowPending, "allow-pending", false, "allow PENDING migrations to run")
}

// parse parses the flags of one command. Extra positional arguments are a
// usage error.
func (r *runner) parse(fs *pflag.FlagSet, args []string) error {
	fs.SetOutput(r.stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return errHelp
		}
		return fmt.Errorf("%w: %v", migrate.ErrUsage, err)
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("%w: unexpected arguments %v", migrate.ErrUsage, fs.Args())
	}
	return nil
}

// openSource loads the database configs. The config file is required.
func (r *runner) openSource(configFile string) (dataSource, error) {
	if configFile == "" {
		return nil, fmt.Errorf("%w: --config-file is required", migrate.ErrUsage)
	}
	return r.loadSource(configFile)
}

// parseVersionFlag parses a released version given by flag. PENDING and an
// unparsable version are usage errors.
func parseVersionFlag(flag, raw string) (register.Version, error) {
	if register.IsPending(raw) {
		return register.Version{}, fmt.Errorf("%w: %s %s is not a released version", migrate.ErrUsage, flag, raw)
	}
	v, err := register.Parse(raw)
	if err != nil {
		return register.Version{}, fmt.Errorf("%w: %s: %v", migrate.ErrUsage, flag, err)
	}
	return v, nil
}

// orNone returns s, or "none" when s is empty.
func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// newTable returns a writer aligning tab separated columns. Flush it after
// the last row.
func newTable(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}
