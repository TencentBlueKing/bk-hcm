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
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Readable migration IDs. These tests must not mutate register.Main.
const (
	migA = "20260101-1200-A-0001"
	migB = "20260101-1200-B-0002"
	migC = "20260101-1200-C-0003"
	migD = "20260101-1200-D-0004"
)

func noopUp(_ context.Context, _ orm.Interface) error {
	return nil
}

func pkgTag(raw string) string {
	raw = strings.ToLower(raw)
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == ' ' || r == '-':
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "x"
	}
	return out
}

func mustMigration(t *testing.T, database, id, version, timestamp, tag string) register.Migration {
	t.Helper()
	safe := pkgTag(tag)
	var pkgPath string
	if register.IsPending(version) {
		pkgPath = fmt.Sprintf("%s%s/pending/%s_%s", constant.MigrationPkgPrefix, database, timestamp, safe)
	} else {
		pkgPath = fmt.Sprintf("%s%s/%s/%s_%s_%s", constant.MigrationPkgPrefix, database, version, version,
			timestamp, safe)
	}
	m, err := register.NewMigration(database, pkgPath, id, version, timestamp, noopUp)
	require.NoError(t, err)
	return m
}

func mustVersion(t *testing.T, raw string) register.Version {
	t.Helper()
	v, err := register.Parse(raw)
	require.NoError(t, err)
	return v
}

// mustRegistry builds a registry with an unexported list. CLI tests cannot
// Regist because types here are not under migrations/.
func mustRegistry(t *testing.T, database string, ms []register.Migration) *register.Registry {
	t.Helper()
	reg := &register.Registry{}
	v := reflect.ValueOf(reg).Elem()
	dbField := v.FieldByName("database")
	*(*string)(unsafe.Pointer(dbField.UnsafeAddr())) = database
	listField := v.FieldByName("list")
	*(*[]register.Migration)(unsafe.Pointer(listField.UnsafeAddr())) = append([]register.Migration(nil), ms...)
	return reg
}

// mapSource is a fake dataSource keyed by registry database name.
type mapSource struct {
	orms          map[string]orm.Interface
	notConfigured map[string]bool
	openErr       error
	openErrDB     string
}

func (s *mapSource) Open(_ *kit.Kit, reg *register.Registry) (orm.Interface, bool, error) {
	if s.openErr != nil && (s.openErrDB == "" || s.openErrDB == reg.Database()) {
		return nil, false, s.openErr
	}
	if s.notConfigured != nil && s.notConfigured[reg.Database()] {
		return nil, false, nil
	}
	o, ok := s.orms[reg.Database()]
	if !ok {
		return nil, false, nil
	}
	return o, true, nil
}

// selectFrom returns a selectRegistries func over known, with the same
// filtering rules as engine.SelectRegistries.
func selectFrom(known []*register.Registry) func([]string) ([]*register.Registry, error) {
	return func(values []string) ([]*register.Registry, error) {
		if len(values) == 0 {
			return slices.Clone(known), nil
		}
		names := make([]string, 0, len(known))
		for _, r := range known {
			names = append(names, r.Database())
		}
		wanted := make(map[string]struct{})
		for _, value := range values {
			for _, name := range strings.Split(value, ",") {
				name = strings.TrimSpace(name)
				if name == "" {
					return nil, fmt.Errorf("%w: --database %q contains an empty database name",
						migrate.ErrUsage, value)
				}
				if !slices.Contains(names, name) {
					return nil, fmt.Errorf("%w: unknown database %q, supported: %s",
						migrate.ErrUsage, name, strings.Join(names, ", "))
				}
				wanted[name] = struct{}{}
			}
		}
		out := make([]*register.Registry, 0, len(wanted))
		for _, r := range known {
			if _, ok := wanted[r.Database()]; ok {
				out = append(out, r)
			}
		}
		return out, nil
	}
}

func newTestRunner(t *testing.T, args []string, source dataSource, regs []*register.Registry) (
	*runner, *bytes.Buffer, *bytes.Buffer) {

	t.Helper()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	r := &runner{
		stdout: stdout,
		stderr: stderr,
		args:   args,
		loadSource: func(string) (dataSource, error) {
			return source, nil
		},
		selectRegistries: selectFrom(regs),
	}
	return r, stdout, stderr
}

func assertExit(t *testing.T, got, want int, stdout, stderr *bytes.Buffer) {
	t.Helper()
	if !assert.Equal(t, want, got) {
		t.Logf("stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
}
