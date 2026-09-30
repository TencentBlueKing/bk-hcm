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
	"math/rand"
	"strings"
	"testing"

	"hcm/pkg/dal/dao/orm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample IDs in canonical lowercase form, ordered so that idA < idB < idC as
// strings, which is how the last sort key compares them.
const (
	idA = "0a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d"
	idB = "1a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d"
	idC = "2a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d"
)

// noopUp is a valid up function for registration tests.
func noopUp(_ context.Context, _ orm.Interface) error {
	return nil
}

func TestRegistValidation(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		version     string
		timestamp   string
		description string
		up          UpFunc
		wantPanic   bool
	}{
		{
			name:        "three segment version",
			id:          idA,
			version:     "v1.9.3",
			timestamp:   "20260905160000",
			description: "add bk_asset_id",
			up:          noopUp,
		},
		{
			name:        "numeric fourth segment",
			id:          idA,
			version:     "v1.9.3.1",
			timestamp:   "20260905160000",
			description: "ziyan backfill",
			up:          noopUp,
		},
		{
			name:        "labeled fourth segment",
			id:          idA,
			version:     "v1.9.3-tenant.1",
			timestamp:   "20260905160000",
			description: "tenant only change",
			up:          noopUp,
		},
		{
			// Accepted at registration, rejected later by the pre-execution scan.
			name:        "pending placeholder",
			id:          idA,
			version:     PendingVersion,
			timestamp:   "20260905160000",
			description: "not released yet",
			up:          noopUp,
		},
		{
			name:        "empty description is allowed",
			id:          idA,
			version:     "v1.9.3",
			timestamp:   "20260905160000",
			description: "",
			up:          noopUp,
		},
		{
			name:        "leap day",
			id:          idA,
			version:     "v1.9.3",
			timestamp:   "20240229000000",
			description: "d",
			up:          noopUp,
		},
		{
			name:        "last second of a day",
			id:          idA,
			version:     "v1.9.3",
			timestamp:   "20261231235959",
			description: "d",
			up:          noopUp,
		},
		{
			name: "invalid version", id: idA, version: "v1.9.3-Tenant.1",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			// v1.0.0- + 55 chars + .1 is 64 characters, the column width.
			name: "version of exactly 64 characters", id: idA,
			version: "v1.0.0-" + strings.Repeat("a", 55) + ".1", timestamp: "20260905160000",
			description: "d", up: noopUp,
		},
		{
			name: "version of 65 characters", id: idA,
			version: "v1.0.0-" + strings.Repeat("a", 56) + ".1", timestamp: "20260905160000",
			description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "empty version", id: idA, version: "",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			// Only the uppercase constant is the placeholder.
			name: "lowercase pending", id: idA, version: "pending",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "timestamp too short", id: idA, version: "v1.9.3",
			timestamp: "2026090516", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "timestamp too long", id: idA, version: "v1.9.3",
			timestamp: "202609051600000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "empty timestamp", id: idA, version: "v1.9.3",
			timestamp: "", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "timestamp with a non digit", id: idA, version: "v1.9.3",
			timestamp: "2026090516000a", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "timestamp with a space", id: idA, version: "v1.9.3",
			timestamp: "2026090516000 ", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "all zero timestamp", id: idA, version: "v1.9.3",
			timestamp: "00000000000000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "month 13", id: idA, version: "v1.9.3",
			timestamp: "20261301000000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "day 00", id: idA, version: "v1.9.3",
			timestamp: "20260900000000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "september 31", id: idA, version: "v1.9.3",
			timestamp: "20260931000000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "february 29 outside a leap year", id: idA, version: "v1.9.3",
			timestamp: "20260229000000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "hour 24", id: idA, version: "v1.9.3",
			timestamp: "20260905240000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "minute 60", id: idA, version: "v1.9.3",
			timestamp: "20260905166000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "second 60", id: idA, version: "v1.9.3",
			timestamp: "20260905160060", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "id is not a uuid", id: "add-bk-asset-id", version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "empty id", id: "", version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			// MySQL compares the recorded ID case-insensitively, Go does not.
			// One migration must have exactly one spelling.
			name: "uppercase uuid", id: strings.ToUpper(idA), version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "uuid without hyphens", id: strings.ReplaceAll(idA, "-", ""), version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "uuid in braces", id: "{" + idA + "}", version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "uuid with a urn prefix", id: "urn:uuid:" + idA, version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "uuid with trailing space", id: idA + " ", version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "uuid with a non hex character", id: "g0000000-0000-4000-8000-000000000000", version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: noopUp, wantPanic: true,
		},
		{
			name: "nil up", id: idA, version: "v1.9.3",
			timestamp: "20260905160000", description: "d", up: nil, wantPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Registry{database: "test"}
			call := func() { r.Regist(tc.id, tc.version, tc.timestamp, tc.description, tc.up) }

			if tc.wantPanic {
				assert.Panics(t, call)
				// A rejected migration must not end up half registered.
				assert.Empty(t, r.All())
				return
			}

			require.NotPanics(t, call)
			all := r.All()
			require.Len(t, all, 1)
			assert.Equal(t, tc.id, all[0].ID)
			assert.Equal(t, tc.version, all[0].Version)
			assert.Equal(t, tc.timestamp, all[0].Timestamp)
			assert.Equal(t, tc.description, all[0].Description)
			assert.NotNil(t, all[0].Up)
		})
	}
}

func TestRegistRejectsVersionLongerThanColumn(t *testing.T) {
	r := &Registry{database: "main"}
	version := "v1.0.0-" + strings.Repeat("a", 56) + ".1"
	require.Greater(t, len(version), MaxVersionLen)

	defer func() {
		rec := recover()
		require.NotNil(t, rec)
		msg, ok := rec.(string)
		require.True(t, ok)
		assert.Contains(t, msg, "longer than 64 characters")
		assert.Contains(t, msg, version)
		assert.Empty(t, r.All())
	}()

	r.Regist(idA, version, "20260905160000", "d", noopUp)
}

func TestRegistPanicMessageLocatesTheFile(t *testing.T) {
	r := &Registry{database: "main"}

	defer func() {
		rec := recover()
		require.NotNil(t, rec, "invalid registration must panic")

		msg, ok := rec.(string)
		require.True(t, ok, "panic value should be a string, got %T", rec)

		// The panic text is all the reader gets to find the offending file,
		// so it must carry every field that was registered.
		assert.Contains(t, msg, "main")
		assert.Contains(t, msg, idA)
		assert.Contains(t, msg, "v1.9.3-Tenant.1")
		assert.Contains(t, msg, "20260905160000")
		assert.Contains(t, msg, "add bk_asset_id")
	}()

	r.Regist(idA, "v1.9.3-Tenant.1", "20260905160000", "add bk_asset_id", noopUp)
}

func TestMigrationParsedVersion(t *testing.T) {
	r := &Registry{database: "test"}
	r.Regist(idA, "v1.9.3.0", "20260905160000", "folded", noopUp)
	r.Regist(idB, PendingVersion, "20260905160001", "undecided", noopUp)

	all := r.All()
	require.Len(t, all, 2)

	versioned, pending := all[0], all[1]

	assert.False(t, versioned.IsPending())
	got, ok := versioned.ParsedVersion()
	require.True(t, ok)
	// ".0" folds away, so this compares equal to the three segment version.
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
		// The normal result of merging one change from another line: same
		// ID, one copy three-segment and one with a fourth segment.
		require.NotPanics(t, func() {
			r.Regist(idA, "v1.9.3", "20260905160000", "external copy", noopUp)
			r.Regist(idA, "v1.9.3.1", "20260905160000", "internal copy", noopUp)
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
			r.Regist(idA, "v1.9.3", "20260905160000", "first", noopUp)
			r.Regist(idA, "v1.9.3", "20260905160000", "second", noopUp)
		})
		assert.Len(t, r.All(), 2)
	})

	t.Run("different registries same id", func(t *testing.T) {
		main := &Registry{database: "main"}
		obs := &Registry{database: "obs"}
		require.NotPanics(t, func() {
			main.Regist(idA, "v1.9.3", "20260905160000", "main copy", noopUp)
			obs.Regist(idA, "v1.9.3", "20260905160000", "obs copy", noopUp)
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
	r := &Registry{database: "test"}
	r.Regist(idA, "v1.9.3", "20260905160000", "first", noopUp)
	r.Regist(idB, "v1.9.4", "20260905160000", "second", noopUp)

	got := r.All()
	require.Len(t, got, 2)

	// Reorder, overwrite and extend the returned slice.
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

	// Expected execution order: version first, then timestamp, then ID.
	// Pending has no version to compare and goes last.
	want := []entry{
		{id: idA, version: "v1.9.3", timestamp: "20260905160000"},
		// Same version, earlier timestamp wins over a smaller ID.
		{id: idC, version: "v1.9.3.0", timestamp: "20260905160001"},
		{id: idA, version: "v1.9.3.0", timestamp: "20260905160002"},
		// Same version and timestamp, the ID decides.
		{id: idA, version: "v1.9.3.1", timestamp: "20260905160000"},
		{id: idB, version: "v1.9.3.1", timestamp: "20260905160000"},
		{id: idC, version: "v1.9.3.1", timestamp: "20260905160000"},
		{id: idA, version: "v1.9.3-tenant.1", timestamp: "20260905160000"},
		{id: idA, version: "v1.9.4", timestamp: "20260101000000"},
		// Numeric, not lexical: v1.9.10 is after v1.9.4.
		{id: idA, version: "v1.9.10", timestamp: "20260101000000"},
		// Pending last, and among themselves by timestamp then ID.
		{id: idA, version: PendingVersion, timestamp: "20260905160000"},
		{id: idB, version: PendingVersion, timestamp: "20260905160000"},
		{id: idA, version: PendingVersion, timestamp: "20260905160001"},
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

			r := &Registry{database: "test"}
			for _, e := range shuffled {
				r.Regist(e.id, e.version, e.timestamp, "d", noopUp)
			}

			assert.Equal(t, want, keys(r.All()), "seed %d", seed)
		}
	})

	t.Run("repeated calls return the same order", func(t *testing.T) {
		r := &Registry{database: "test"}
		for _, e := range want {
			r.Regist(e.id, e.version, e.timestamp, "d", noopUp)
		}

		first := keys(r.All())
		for i := 0; i < 5; i++ {
			assert.Equal(t, first, keys(r.All()), "call %d", i)
		}
	})

	t.Run("description does not affect the order", func(t *testing.T) {
		r := &Registry{database: "test"}
		r.Regist(idA, "v1.9.3", "20260905160000", "zzz runs first", noopUp)
		r.Regist(idB, "v1.9.4", "20260905160000", "aaa runs second", noopUp)

		all := r.All()
		require.Len(t, all, 2)
		assert.Equal(t, []string{"v1.9.3", "v1.9.4"}, []string{all[0].Version, all[1].Version})
	})

	t.Run("identical sort keys keep registration order", func(t *testing.T) {
		r := &Registry{database: "test"}
		r.Regist(idA, "v1.9.3", "20260905160000", "registered first", noopUp)
		r.Regist(idA, "v1.9.3", "20260905160000", "registered second", noopUp)

		all := r.All()
		require.Len(t, all, 2)
		assert.Equal(t, "registered first", all[0].Description)
		assert.Equal(t, "registered second", all[1].Description)
	})

	t.Run("empty registry", func(t *testing.T) {
		r := &Registry{database: "test"}
		assert.Empty(t, r.All())
	})
}

// TestAllOrderMatchesComparator checks that All agrees with the version
// comparator: whenever Compare says one version is smaller, the migration
// carrying it comes first, regardless of timestamp and ID.
func TestAllOrderMatchesComparator(t *testing.T) {
	r := &Registry{database: "test"}
	// The later version deliberately gets the earlier timestamp and the
	// smaller ID, so only the version can produce the expected order.
	r.Regist(idC, "v1.9.10", "20260101000000", "later version", noopUp)
	r.Regist(idA, "v1.9.9", "20261231235959", "earlier version", noopUp)

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
