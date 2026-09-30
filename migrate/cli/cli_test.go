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
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"hcm/migrate/engine"
	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunUsageAndExitCodes(t *testing.T) {
	mainReg := mustRegistry(t, "main", []register.Migration{
		mustMigration(t, "main", migA, "v1.9.3", "20260101120000", "a"),
		mustMigration(t, "main", migB, constant.MigrationPendingVersion, "20260102120000", "pending"),
	})
	obsReg := mustRegistry(t, "obs", []register.Migration{
		mustMigration(t, "obs", migC, "v1.9.3", "20260101120000", "c"),
	})
	regs := []*register.Registry{mainReg, obsReg}
	fakeSrc := &mapSource{orms: map[string]orm.Interface{}}

	type tc struct {
		name       string
		args       []string
		wantCode   int
		stdoutHas  []string
		stderrHas  []string
		stdoutMiss []string
		setup      func(*runner)
	}

	testCases := []tc{
		{
			name: "no args", args: nil, wantCode: constant.MigrationExitUsage,
			stderrHas: []string{"Usage: hcm-migrate"},
		},
		{
			name: "unknown command", args: []string{"fly"}, wantCode: constant.MigrationExitUsage,
			stderrHas: []string{`unknown command "fly"`, "Usage: hcm-migrate"},
		},
		{
			name: "help", args: []string{"help"}, wantCode: constant.MigrationExitSuccess,
			stdoutHas: []string{"Usage: hcm-migrate"},
		},
		{
			name: "short help", args: []string{"-h"}, wantCode: constant.MigrationExitSuccess,
			stdoutHas: []string{"Usage: hcm-migrate"},
		},
		{
			name: "long help", args: []string{"--help"}, wantCode: constant.MigrationExitSuccess,
			stdoutHas: []string{"Usage: hcm-migrate"},
		},
		{
			name: "init --help", args: []string{"init", "--help"}, wantCode: constant.MigrationExitSuccess,
		},
		{
			name: "up --help", args: []string{"up", "--help"}, wantCode: constant.MigrationExitSuccess,
		},
		{
			name: "status --help", args: []string{"status", "--help"}, wantCode: constant.MigrationExitSuccess,
		},
		{
			name: "list --help", args: []string{"list", "--help"}, wantCode: constant.MigrationExitSuccess,
		},
		{
			name: "init extra args", args: []string{"init", "--mode=empty", "-c", "x.yaml", "extra"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name: "up extra args", args: []string{"up", "-c", "x.yaml", "extra"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name: "status extra args", args: []string{"status", "-c", "x.yaml", "extra"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name: "list extra args", args: []string{"list", "extra"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name: "init missing config", args: []string{"init", "--mode=empty"},
			wantCode: constant.MigrationExitUsage, stderrHas: []string{},
		},
		{
			name: "up missing config", args: []string{"up"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name: "status missing config", args: []string{"status"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "bad database via engine.SelectRegistries",
			args:     []string{"list", "-d", "unknown"},
			wantCode: constant.MigrationExitUsage,
			setup: func(r *runner) {
				r.selectRegistries = engine.SelectRegistries
			},
		},
		{
			name:     "empty database name",
			args:     []string{"list", "-d", "main,"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "up --to PENDING",
			args:     []string{"up", "-c", "x.yaml", "--to", "PENDING"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "up --to invalid",
			args:     []string{"up", "-c", "x.yaml", "--to", "v1"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:      "up --to valid reaches openSource",
			args:      []string{"up", "-c", "x.yaml", "--to", "v1.9.3"},
			wantCode:  constant.MigrationExitSuccess,
			stdoutHas: []string{"mode: default", "to: v1.9.3", "plan only: false"},
			setup: func(r *runner) {
				// no registries selected → nothing to open
				r.selectRegistries = selectFrom(nil)
			},
		},
		{
			name:     "init mode missing",
			args:     []string{"init", "-c", "x.yaml"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init mode invalid",
			args:     []string{"init", "-c", "x.yaml", "--mode=full"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init empty with baseline",
			args:     []string{"init", "-c", "x.yaml", "--mode=empty", "--baseline", "main=v1.9.3"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init adopt without baseline",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init baseline malformed no equals",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt", "--baseline", "main"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init baseline malformed leading equals",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt", "--baseline", "=v1.9.3"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init baseline malformed trailing equals",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt", "--baseline", "main="},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init baseline unknown database",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt", "--baseline", "foo=v1.9.3"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init baseline PENDING",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt", "--baseline", "main=PENDING"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "init baseline invalid version",
			args:     []string{"init", "-c", "x.yaml", "--mode=adopt", "--baseline", "main=v1"},
			wantCode: constant.MigrationExitUsage,
		},
		{
			name:     "list output order PENDING last exit 0 without config",
			args:     []string{"list", "-d", "main"},
			wantCode: constant.MigrationExitSuccess,
			stdoutHas: []string{
				"database main, 2 migrations",
				migA, "v1.9.3",
				migB, constant.MigrationPendingVersion,
			},
		},
		{
			name:     "loadSource usage-wrapped → 2",
			args:     []string{"status", "-c", "x.yaml"},
			wantCode: constant.MigrationExitUsage,
			setup: func(r *runner) {
				r.loadSource = func(string) (dataSource, error) {
					return nil, fmt.Errorf("%w: load config failed", migrate.ErrUsage)
				}
			},
		},
		{
			name:     "loadSource plain → 1",
			args:     []string{"status", "-c", "x.yaml"},
			wantCode: constant.MigrationExitFailure,
			setup: func(r *runner) {
				r.loadSource = func(string) (dataSource, error) {
					return nil, errors.New("disk gone")
				}
			},
		},
		{
			name:     "dataSource Open error → 1",
			args:     []string{"status", "-c", "x.yaml", "-d", "main"},
			wantCode: constant.MigrationExitFailure,
			setup: func(r *runner) {
				r.loadSource = func(string) (dataSource, error) {
					return &mapSource{openErr: errors.New("connect refused")}, nil
				}
			},
		},
		{
			name:       "not-configured database skipped silently on stdout",
			args:       []string{"status", "-c", "x.yaml", "-d", "main"},
			wantCode:   constant.MigrationExitSuccess,
			stdoutMiss: []string{"not configured", "skipped", "database main"},
			setup: func(r *runner) {
				r.loadSource = func(string) (dataSource, error) {
					return &mapSource{notConfigured: map[string]bool{"main": true}}, nil
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, stdout, stderr := newTestRunner(t, tc.args, fakeSrc, regs)
			if tc.setup != nil {
				tc.setup(r)
			}
			code := r.run(tc.args)
			assertExit(t, code, tc.wantCode, stdout, stderr)
			for _, s := range tc.stdoutHas {
				assert.Contains(t, stdout.String(), s)
			}
			for _, s := range tc.stdoutMiss {
				assert.NotContains(t, stdout.String(), s)
			}
			for _, s := range tc.stderrHas {
				assert.Contains(t, stderr.String(), s)
			}
		})
	}
}

func TestRunListPENDINGLast(t *testing.T) {
	pending := mustMigration(t, "main", migB, constant.MigrationPendingVersion, "20260101120000", "pending")
	early := mustMigration(t, "main", migA, "v1.9.2", "20260102120000", "early")
	later := mustMigration(t, "main", migC, "v1.9.3", "20260101120000", "later")
	// Register PENDING first; All() still sorts it last.
	reg := mustRegistry(t, "main", []register.Migration{pending, later, early})
	r, stdout, stderr := newTestRunner(t, []string{"list"}, &mapSource{}, []*register.Registry{reg})
	code := r.run([]string{"list"})
	assertExit(t, code, constant.MigrationExitSuccess, stdout, stderr)
	out := stdout.String()
	posA := strings.Index(out, migA)
	posC := strings.Index(out, migC)
	posB := strings.Index(out, migB)
	require.GreaterOrEqual(t, posA, 0)
	require.GreaterOrEqual(t, posC, 0)
	require.GreaterOrEqual(t, posB, 0)
	assert.Less(t, posA, posC, "v1.9.2 before v1.9.3")
	assert.Less(t, posC, posB, "released before PENDING")
}

func TestRunPublicAPI(t *testing.T) {
	// Run wires real SelectRegistries; list with no args still exits 0.
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"list", "-d", "main"}, stdout, stderr)
	assert.Equal(t, constant.MigrationExitSuccess, code)
	assert.Contains(t, stdout.String(), "database main")
}

func TestParseVersionFlag(t *testing.T) {
	testCases := []struct {
		name    string
		raw     string
		wantErr bool
		wantRaw string
	}{
		{name: "PENDING", raw: "PENDING", wantErr: true},
		{name: "invalid", raw: "v1", wantErr: true},
		{name: "valid", raw: "v1.9.3", wantRaw: "v1.9.3"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := parseVersionFlag("--to", tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, migrate.ErrUsage)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantRaw, v.Raw)
		})
	}
}

func TestOrNone(t *testing.T) {
	assert.Equal(t, "none", orNone(""))
	assert.Equal(t, "v1.9.3", orNone("v1.9.3"))
}
