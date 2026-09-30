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

package register

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/orm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample IDs in the readable uppercase form, ordered so that idA < idB < idC as
// strings, which is how the last sort key compares them.
const (
	idA = "20260101-1200-A-0001"
	idB = "20260101-1200-B-0002"
	idC = "20260101-1200-C-0003"
)

const (
	tsDefault  = "20260905160000"
	tagDefault = "add_x"
)

// noopUp is a valid up function for registration tests.
func noopUp(_ context.Context, _ orm.Interface) error {
	return nil
}

// releasedPkg builds a released migration import path under database.
func releasedPkg(database, version, timestamp, tag string) string {
	return fmt.Sprintf("%s%s/%s/%s_%s_%s", constant.MigrationPkgPrefix, database, version, version, timestamp, tag)
}

// pendingPkg builds a pending migration import path under database.
func pendingPkg(database, timestamp, tag string) string {
	return fmt.Sprintf("%s%s/pending/%s_%s", constant.MigrationPkgPrefix, database, timestamp, tag)
}

func mustRegist(t *testing.T, r *Registry, id, version, timestamp, tag string) {
	t.Helper()
	var pkgPath string
	if IsPending(version) {
		pkgPath = pendingPkg(r.database, timestamp, tag)
	} else {
		pkgPath = releasedPkg(r.database, version, timestamp, tag)
	}
	require.NotPanics(t, func() {
		r.regist(pkgPath, id, version, timestamp, noopUp)
	})
}

func TestNewMigrationSurfacesValidatorErrors(t *testing.T) {
	validPkg := releasedPkg("main", "v1.9.3", tsDefault, tagDefault)

	t.Run("invalid id", func(t *testing.T) {
		m, err := NewMigration("main", validPkg, "bad-id", "v1.9.3", tsDefault, noopUp)
		require.Error(t, err)
		assert.Equal(t, Migration{}, m)
		assert.Contains(t, err.Error(), "invalid migration id")
		assert.Contains(t, err.Error(), "id: \"bad-id\"")
		assert.Contains(t, err.Error(), "version: \"v1.9.3\"")
		assert.Contains(t, err.Error(), "timestamp: \""+tsDefault+"\"")
		assert.Contains(t, err.Error(), "package: \""+validPkg+"\"")
	})

	t.Run("invalid timestamp", func(t *testing.T) {
		m, err := NewMigration("main", validPkg, idA, "v1.9.3", "2026090516", noopUp)
		require.Error(t, err)
		assert.Equal(t, Migration{}, m)
		assert.Contains(t, err.Error(), "invalid timestamp")
		assert.Contains(t, err.Error(), "id: \""+idA+"\"")
		assert.Contains(t, err.Error(), "timestamp: \"2026090516\"")
		assert.Contains(t, err.Error(), "package: \""+validPkg+"\"")
	})

	t.Run("invalid package", func(t *testing.T) {
		badPkg := "main/v1.9.3/v1.9.3_" + tsDefault + "_add_x"
		m, err := NewMigration("main", badPkg, idA, "v1.9.3", tsDefault, noopUp)
		require.Error(t, err)
		assert.Equal(t, Migration{}, m)
		assert.Contains(t, err.Error(), "is not under")
		assert.Contains(t, err.Error(), "id: \""+idA+"\"")
		assert.Contains(t, err.Error(), "package: \""+badPkg+"\"")
	})

	t.Run("valid builds migration", func(t *testing.T) {
		m, err := NewMigration("main", validPkg, idA, "v1.9.3", tsDefault, noopUp)
		require.NoError(t, err)
		assert.Equal(t, idA, m.ID)
		assert.Equal(t, "v1.9.3", m.Version)
		assert.Equal(t, tsDefault, m.Timestamp)
		assert.Equal(t, "main/v1.9.3/v1.9.3_"+tsDefault+"_"+tagDefault, m.Pkg)
		assert.NotNil(t, m.Up)
	})
}

func TestNewMigrationVersion(t *testing.T) {
	pkgOf := func(timestamp string) string {
		return releasedPkg("main", "v1.9.3", timestamp, tagDefault)
	}

	testCases := []struct {
		name      string
		version   string
		timestamp string
		pkgPath   string
		wantErr   bool
	}{
		{name: "three segment version", version: "v1.9.3", timestamp: tsDefault},
		{name: "numeric fourth segment", version: "v1.9.3.1", timestamp: tsDefault,
			pkgPath: releasedPkg("main", "v1.9.3.1", tsDefault, tagDefault)},
		{name: "labeled fourth segment", version: "v1.9.3-tenant.1", timestamp: tsDefault,
			pkgPath: releasedPkg("main", "v1.9.3-tenant.1", tsDefault, tagDefault)},
		{name: "pending placeholder", version: constant.MigrationPendingVersion, timestamp: tsDefault,
			pkgPath: pendingPkg("main", tsDefault, tagDefault)},
		{name: "version of exactly 64 characters",
			version: "v1.0.0-" + strings.Repeat("a", 55) + ".1", timestamp: tsDefault,
			pkgPath: releasedPkg("main", "v1.0.0-"+strings.Repeat("a", 55)+".1", tsDefault, tagDefault)},
		{name: "invalid version", version: "v1.9.3-Tenant.1", timestamp: tsDefault, wantErr: true},
		{name: "version of 65 characters",
			version: "v1.0.0-" + strings.Repeat("a", 56) + ".1", timestamp: tsDefault, wantErr: true},
		{name: "empty version", version: "", timestamp: tsDefault, wantErr: true},
		{name: "lowercase pending", version: "pending", timestamp: tsDefault, wantErr: true},
		{name: "nil up", version: "v1.9.3", timestamp: tsDefault, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkgPath := tc.pkgPath
			if pkgPath == "" {
				pkgPath = pkgOf(tc.timestamp)
			}
			var up UpFunc = noopUp
			if tc.name == "nil up" {
				up = nil
			}
			m, err := NewMigration("main", pkgPath, idA, tc.version, tc.timestamp, up)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, Migration{}, m)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.version, m.Version)
			assert.Equal(t, tc.timestamp, m.Timestamp)
			assert.NotNil(t, m.Up)
		})
	}
}

type valueMigrator struct{}

func (valueMigrator) Up(_ context.Context, _ orm.Interface) error { return nil }

type ptrMigrator struct{}

func (*ptrMigrator) Up(_ context.Context, _ orm.Interface) error { return nil }

type flagMigrator struct {
	called *bool
}

func (f *flagMigrator) Up(_ context.Context, _ orm.Interface) error {
	*f.called = true
	return nil
}

func TestRegistMigrator(t *testing.T) {
	t.Run("nil migrator panics with database and id", func(t *testing.T) {
		r := &Registry{database: "main"}
		defer func() {
			rec := recover()
			require.NotNil(t, rec)
			msg, ok := rec.(string)
			require.True(t, ok)
			assert.Contains(t, msg, "main")
			assert.Contains(t, msg, idA)
			assert.Contains(t, msg, "migrator is nil")
			assert.Empty(t, r.All())
		}()
		r.Regist(idA, "v1.9.3", tsDefault, nil)
	})

	t.Run("typed nil pointer panics", func(t *testing.T) {
		r := &Registry{database: "main"}
		var m *ptrMigrator
		defer func() {
			rec := recover()
			require.NotNil(t, rec)
			msg, ok := rec.(string)
			require.True(t, ok)
			assert.Contains(t, msg, "nil")
			assert.Contains(t, msg, idA)
			assert.Empty(t, r.All())
		}()
		r.Regist(idA, "v1.9.3", tsDefault, m)
	})

	t.Run("value and pointer receivers share package path", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			m    Migrator
		}{
			{name: "value", m: valueMigrator{}},
			{name: "pointer", m: &ptrMigrator{}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				r := &Registry{database: "main"}
				defer func() {
					rec := recover()
					require.NotNil(t, rec)
					msg, ok := rec.(string)
					require.True(t, ok)
					assert.Contains(t, msg, "hcm/migrate/register")
					assert.Contains(t, msg, "main")
					assert.Empty(t, r.All())
				}()
				r.Regist(idA, "v1.9.3", tsDefault, tc.m)
			})
		}
	})

	t.Run("success via regist with injected path", func(t *testing.T) {
		r := &Registry{database: "main"}
		pkgPath := releasedPkg("main", "v1.9.3", tsDefault, "add_bk_asset_id")
		require.NotPanics(t, func() {
			r.regist(pkgPath, idA, "v1.9.3", tsDefault, noopUp)
		})
		all := r.All()
		require.Len(t, all, 1)
		assert.Equal(t, idA, all[0].ID)
		assert.Equal(t, "main/v1.9.3/v1.9.3_"+tsDefault+"_add_bk_asset_id", all[0].Pkg)
	})

	t.Run("Migration.Up invokes Migrator.Up", func(t *testing.T) {
		r := &Registry{database: "main"}
		called := false
		_, up, err := migratorOf(&flagMigrator{called: &called})
		require.NoError(t, err)
		r.regist(releasedPkg("main", "v1.9.3", tsDefault, tagDefault), idA, "v1.9.3", tsDefault, up)
		require.NoError(t, r.All()[0].Up(context.Background(), nil))
		assert.True(t, called)
	})
}

func TestRegistRejectsVersionLongerThanColumn(t *testing.T) {
	r := &Registry{database: "main"}
	version := "v1.0.0-" + strings.Repeat("a", 56) + ".1"
	require.Greater(t, len(version), constant.MigrationVersionMaxLen)

	defer func() {
		rec := recover()
		require.NotNil(t, rec)
		msg, ok := rec.(string)
		require.True(t, ok)
		assert.Contains(t, msg, "longer than 64 characters")
		assert.Contains(t, msg, version)
		assert.Empty(t, r.All())
	}()

	r.regist(releasedPkg("main", version, tsDefault, tagDefault), idA, version, tsDefault, noopUp)
}

func TestRegistPanicMessageLocatesTheFile(t *testing.T) {
	r := &Registry{database: "main"}
	pkgPath := releasedPkg("main", "v1.9.3", tsDefault, "add_bk_asset_id")

	defer func() {
		rec := recover()
		require.NotNil(t, rec, "invalid registration must panic")

		msg, ok := rec.(string)
		require.True(t, ok, "panic value should be a string, got %T", rec)

		assert.Contains(t, msg, "main")
		assert.Contains(t, msg, idA)
		assert.Contains(t, msg, "v1.9.3-Tenant.1")
		assert.Contains(t, msg, tsDefault)
		assert.Contains(t, msg, pkgPath)
	}()

	r.regist(pkgPath, idA, "v1.9.3-Tenant.1", tsDefault, noopUp)
}

func TestMigrationParsedVersion(t *testing.T) {
	r := &Registry{database: "main"}
	mustRegist(t, r, idA, "v1.9.3.0", tsDefault, "folded")
	mustRegist(t, r, idB, constant.MigrationPendingVersion, "20260905160001", "undecided")

	all := r.All()
	require.Len(t, all, 2)

	versioned, pending := all[0], all[1]

	assert.False(t, versioned.IsPending())
	got, ok := versioned.ParsedVersion()
	require.True(t, ok)
	assert.Equal(t, 0, Compare(got, mustParseVersion(t, "v1.9.3")))
	assert.Equal(t, "v1.9.3.0", versioned.Version, "Raw version string is kept as registered")

	assert.True(t, pending.IsPending())
	got, ok = pending.ParsedVersion()
	assert.False(t, ok)
	assert.Equal(t, Version{}, got, "a pending migration must not expose a zero version as a real one")
}

func TestRegistDuplicateID(t *testing.T) {
	t.Run("same registry same id both register", func(t *testing.T) {
		r := &Registry{database: "main"}
		require.NotPanics(t, func() {
			mustRegist(t, r, idA, "v1.9.3", tsDefault, "external_copy")
			mustRegist(t, r, idA, "v1.9.3.1", tsDefault, "internal_copy")
		})

		all := r.All()
		require.Len(t, all, 2)
		assert.Equal(t, idA, all[0].ID)
		assert.Equal(t, idA, all[1].ID)
		assert.Equal(t, []string{"v1.9.3", "v1.9.3.1"}, []string{all[0].Version, all[1].Version})
	})

	t.Run("same id same version both register", func(t *testing.T) {
		r := &Registry{database: "main"}
		require.NotPanics(t, func() {
			mustRegist(t, r, idA, "v1.9.3", tsDefault, "first")
			mustRegist(t, r, idA, "v1.9.3", tsDefault, "second")
		})
		assert.Len(t, r.All(), 2)
	})

	t.Run("different registries same id", func(t *testing.T) {
		main := &Registry{database: "main"}
		obs := &Registry{database: "obs"}
		require.NotPanics(t, func() {
			mustRegist(t, main, idA, "v1.9.3", tsDefault, "main_copy")
			mustRegist(t, obs, idA, "v1.9.3", tsDefault, "obs_copy")
		})

		assert.Len(t, main.All(), 1)
		assert.Len(t, obs.All(), 1)
	})
}

func TestRegistryInstances(t *testing.T) {
	assert.Equal(t, "main", Main.Database())
	assert.Equal(t, "obs", Obs.Database())
	assert.NotSame(t, Main, Obs, "the two databases must not share one registry")
}

func TestAllReturnsACopy(t *testing.T) {
	r := &Registry{database: "main"}
	mustRegist(t, r, idA, "v1.9.3", tsDefault, "first")
	mustRegist(t, r, idB, "v1.9.4", tsDefault, "second")

	got := r.All()
	require.Len(t, got, 2)

	got[0], got[1] = got[1], got[0]
	got[0].Version = "v9.9.9"
	got = append(got, Migration{ID: idC})

	again := r.All()
	require.Len(t, again, 2)
	assert.Equal(t, []string{idA, idB}, []string{again[0].ID, again[1].ID})
	assert.Equal(t, []string{"v1.9.3", "v1.9.4"}, []string{again[0].Version, again[1].Version})
}

func TestAllOrder(t *testing.T) {
	type entry struct {
		id        string
		version   string
		timestamp string
	}

	want := []entry{
		{id: idA, version: "v1.9.3", timestamp: tsDefault},
		{id: idC, version: "v1.9.3.0", timestamp: "20260905160001"},
		{id: idA, version: "v1.9.3.0", timestamp: "20260905160002"},
		{id: idA, version: "v1.9.3.1", timestamp: tsDefault},
		{id: idB, version: "v1.9.3.1", timestamp: tsDefault},
		{id: idC, version: "v1.9.3.1", timestamp: tsDefault},
		{id: idA, version: "v1.9.3-tenant.1", timestamp: tsDefault},
		{id: idA, version: "v1.9.4", timestamp: "20260101000000"},
		{id: idA, version: "v1.9.10", timestamp: "20260101000000"},
		{id: idA, version: constant.MigrationPendingVersion, timestamp: tsDefault},
		{id: idB, version: constant.MigrationPendingVersion, timestamp: tsDefault},
		{id: idA, version: constant.MigrationPendingVersion, timestamp: "20260905160001"},
	}

	key := func(m Migration) entry {
		return entry{id: m.ID, version: m.Version, timestamp: m.Timestamp}
	}
	keys := func(ms []Migration) []entry {
		out := make([]entry, 0, len(ms))
		for _, m := range ms {
			out = append(out, key(m))
		}
		return out
	}

	t.Run("registration order does not change the result", func(t *testing.T) {
		for seed := int64(0); seed < 20; seed++ {
			shuffled := append([]entry(nil), want...)
			rnd := rand.New(rand.NewSource(seed))
			rnd.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})

			r := &Registry{database: "main"}
			for i, e := range shuffled {
				mustRegist(t, r, e.id, e.version, e.timestamp, fmt.Sprintf("t%d", i))
			}

			assert.Equal(t, want, keys(r.All()), "seed %d", seed)
		}
	})

	t.Run("repeated calls return the same order", func(t *testing.T) {
		r := &Registry{database: "main"}
		for i, e := range want {
			mustRegist(t, r, e.id, e.version, e.timestamp, fmt.Sprintf("r%d", i))
		}

		first := keys(r.All())
		for i := 0; i < 5; i++ {
			assert.Equal(t, first, keys(r.All()), "call %d", i)
		}
	})

	t.Run("pkg does not affect the order", func(t *testing.T) {
		r := &Registry{database: "main"}
		mustRegist(t, r, idA, "v1.9.3", tsDefault, "zzz_runs_first")
		mustRegist(t, r, idB, "v1.9.4", tsDefault, "aaa_runs_second")

		all := r.All()
		require.Len(t, all, 2)
		assert.Equal(t, []string{"v1.9.3", "v1.9.4"}, []string{all[0].Version, all[1].Version})
	})

	t.Run("identical sort keys keep registration order", func(t *testing.T) {
		r := &Registry{database: "main"}
		mustRegist(t, r, idA, "v1.9.3", tsDefault, "registered_first")
		mustRegist(t, r, idA, "v1.9.3", tsDefault, "registered_second")

		all := r.All()
		require.Len(t, all, 2)
		assert.Contains(t, all[0].Pkg, "registered_first")
		assert.Contains(t, all[1].Pkg, "registered_second")
	})

	t.Run("empty registry", func(t *testing.T) {
		r := &Registry{database: "main"}
		assert.Empty(t, r.All())
	})
}

// TestAllOrderMatchesComparator checks that All agrees with the version
// comparator: whenever Compare says one version is smaller, the migration
// carrying it comes first, regardless of timestamp and ID.
func TestAllOrderMatchesComparator(t *testing.T) {
	r := &Registry{database: "main"}
	mustRegist(t, r, idC, "v1.9.10", "20260101000000", "later_version")
	mustRegist(t, r, idA, "v1.9.9", "20261231235959", "earlier_version")

	all := r.All()
	require.Len(t, all, 2)
	assert.Equal(t, "v1.9.9", all[0].Version)
	assert.Equal(t, "v1.9.10", all[1].Version)

	for i := 0; i+1 < len(all); i++ {
		a, aOK := all[i].ParsedVersion()
		b, bOK := all[i+1].ParsedVersion()
		require.True(t, aOK)
		require.True(t, bOK)
		assert.LessOrEqual(t, Compare(a, b), 0)
	}
}

func mustParseVersion(t *testing.T, raw string) Version {
	t.Helper()
	v, err := Parse(raw)
	require.NoError(t, err)
	return v
}
