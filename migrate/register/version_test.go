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
	"math"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	"hcm/pkg/criteria/constant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name       string
		raw        string
		wantErr    bool
		wantP      [3]int
		wantSuffix *Suffix
	}{
		{
			name:  "three segments",
			raw:   "v1.9.3",
			wantP: [3]int{1, 9, 3},
		},
		{
			name:       "numeric fourth segment",
			raw:        "v1.9.3.1",
			wantP:      [3]int{1, 9, 3},
			wantSuffix: &Suffix{Seq: 1},
		},
		{
			name:       "labeled fourth segment",
			raw:        "v1.9.3-tenant.1",
			wantP:      [3]int{1, 9, 3},
			wantSuffix: &Suffix{Label: "tenant", Seq: 1},
		},
		{
			// A plain ".0" carries no ordering information and is folded away.
			name:  "numeric zero folds to no suffix",
			raw:   "v1.9.3.0",
			wantP: [3]int{1, 9, 3},
		},
		{
			// A label is never folded, it still marks a feature branch.
			name:       "labeled zero keeps suffix",
			raw:        "v1.9.3-tenant.0",
			wantP:      [3]int{1, 9, 3},
			wantSuffix: &Suffix{Label: "tenant", Seq: 0},
		},
		{
			name:  "zero segments are valid",
			raw:   "v0.0.0",
			wantP: [3]int{0, 0, 0},
		},
		{
			// A segment that is exactly "0" is not a leading zero. Rejecting
			// every zero would make v1.0.0 illegal.
			name:  "lone zero is not a leading zero",
			raw:   "v1.0.0",
			wantP: [3]int{1, 0, 0},
		},
		{
			name:  "lone zero in the third segment",
			raw:   "v1.9.0",
			wantP: [3]int{1, 9, 0},
		},
		{
			name:  "numeric zero on the zero version still folds",
			raw:   "v0.0.0.0",
			wantP: [3]int{0, 0, 0},
		},
		{
			name:       "numeric fourth on the zero version",
			raw:        "v0.0.0.1",
			wantP:      [3]int{0, 0, 0},
			wantSuffix: &Suffix{Seq: 1},
		},
		{
			// Shortest label the grammar allows, and seq 0 must stay.
			name:       "single letter label keeps a zero seq",
			raw:        "v1.9.3-a.0",
			wantP:      [3]int{1, 9, 3},
			wantSuffix: &Suffix{Label: "a", Seq: 0},
		},
		{
			name:       "label of a letter followed by a digit",
			raw:        "v1.9.3-a0.1",
			wantP:      [3]int{1, 9, 3},
			wantSuffix: &Suffix{Label: "a0", Seq: 1},
		},
		{
			name:       "multi digit numbers",
			raw:        "v10.20.30.40",
			wantP:      [3]int{10, 20, 30},
			wantSuffix: &Suffix{Seq: 40},
		},
		{
			name:       "label with digits",
			raw:        "v1.9.3-tenant2.1",
			wantP:      [3]int{1, 9, 3},
			wantSuffix: &Suffix{Label: "tenant2", Seq: 1},
		},
		{name: "missing v prefix", raw: "1.9.3", wantErr: true},
		{name: "two segments", raw: "v1.9", wantErr: true},
		{name: "five segments", raw: "v1.9.3.1.2", wantErr: true},
		{name: "label after dot", raw: "v1.9.3.tenant.1", wantErr: true},
		{name: "uppercase label", raw: "v1.9.3-Tenant.1", wantErr: true},
		{name: "label without seq", raw: "v1.9.3-tenant", wantErr: true},
		{name: "label starting with digit", raw: "v1.9.3-2tenant.1", wantErr: true},
		{name: "leading zero in first segment", raw: "v01.9.3", wantErr: true},
		{name: "leading zero in second segment", raw: "v1.09.3", wantErr: true},
		{name: "leading zero in third segment", raw: "v1.9.03", wantErr: true},
		{name: "leading zero in fourth segment", raw: "v1.9.3.01", wantErr: true},
		{name: "leading zero in labeled seq", raw: "v1.9.3-tenant.01", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "placeholder is not a version", raw: constant.MigrationPendingVersion, wantErr: true},
		{name: "lowercase pending is not a version", raw: "pending", wantErr: true},
		{name: "trailing space", raw: "v1.9.3 ", wantErr: true},
		{name: "leading space", raw: " v1.9.3", wantErr: true},
		{name: "uppercase prefix", raw: "V1.9.3", wantErr: true},
		{name: "negative segment", raw: "v-1.9.3", wantErr: true},
		{name: "plus sign", raw: "v1.9.+3", wantErr: true},
		{name: "bare prefix", raw: "v", wantErr: true},
		{name: "one segment", raw: "v1", wantErr: true},
		{name: "empty middle segment", raw: "v1..9.3", wantErr: true},
		{name: "trailing dot", raw: "v1.9.3.", wantErr: true},
		{name: "empty label", raw: "v1.9.3-.1", wantErr: true},
		{name: "label with a missing seq", raw: "v1.9.3-tenant.", wantErr: true},
		{name: "extra segment after a label", raw: "v1.9.3-tenant.1.2", wantErr: true},
		{name: "numeric and labeled fourth together", raw: "v1.9.3.1-tenant.1", wantErr: true},
		{name: "underscore in label", raw: "v1.9.3-tenant_a.1", wantErr: true},
		{name: "hyphen in label", raw: "v1.9.3-ten-ant.1", wantErr: true},
		{name: "uppercase inside label", raw: "v1.9.3-tenAnt.1", wantErr: true},
		{name: "non ascii label", raw: "v1.9.3-租户.1", wantErr: true},
		// "00" is a leading zero, not the foldable ".0". Folding it would
		// make v1.9.3.00 a third spelling of v1.9.3.
		{name: "double zero is a leading zero", raw: "v1.9.3.00", wantErr: true},
		{name: "double zero in a labeled seq", raw: "v1.9.3-tenant.00", wantErr: true},
		{name: "placeholder with a suffix", raw: constant.MigrationPendingVersion + ".1", wantErr: true},
		{name: "placeholder glued to a version", raw: "v1.9.3" + constant.MigrationPendingVersion, wantErr: true},
		{name: "placeholder with surrounding space", raw: " " + constant.MigrationPendingVersion, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.raw)
			if tc.wantErr {
				assert.Error(t, err)
				// A failed parse must not hand back a half-filled version
				// that a caller could sort by accident.
				assert.Equal(t, Version{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantP[0], got.P1)
			assert.Equal(t, tc.wantP[1], got.P2)
			assert.Equal(t, tc.wantP[2], got.P3)
			assert.Equal(t, tc.wantSuffix, got.Suffix)
			assert.Equal(t, tc.raw, got.Raw)
		})
	}
}

func TestIsPending(t *testing.T) {
	testCases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "placeholder constant", raw: constant.MigrationPendingVersion, want: true},
		{name: "uppercase literal", raw: "PENDING", want: true},
		{name: "lowercase is not the placeholder", raw: "pending", want: false},
		{name: "mixed case is not the placeholder", raw: "Pending", want: false},
		{name: "a real version", raw: "v1.9.3", want: false},
		{name: "empty", raw: "", want: false},
		{name: "trailing space", raw: constant.MigrationPendingVersion + " ", want: false},
		{name: "leading space", raw: " " + constant.MigrationPendingVersion, want: false},
		{name: "trailing newline", raw: constant.MigrationPendingVersion + "\n", want: false},
		{name: "placeholder as a prefix of a longer string", raw: constant.MigrationPendingVersion + "x", want: false},
		{name: "placeholder as a suffix of a longer string", raw: "x" + constant.MigrationPendingVersion, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsPending(tc.raw))
		})
	}
}

func TestCompare(t *testing.T) {
	testCases := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal three segments", a: "v1.9.3", b: "v1.9.3", want: 0},
		{name: "numeric zero equals three segments", a: "v1.9.3", b: "v1.9.3.0", want: 0},
		{name: "numeric zero equals three segments reversed", a: "v1.9.3.0", b: "v1.9.3", want: 0},
		{name: "labeled zero is not three segments", a: "v1.9.3", b: "v1.9.3-tenant.0", want: -1},
		{name: "first segment", a: "v1.9.3", b: "v2.0.0", want: -1},
		{name: "second segment", a: "v1.9.3", b: "v1.10.0", want: -1},
		{name: "third segment numeric not lexical", a: "v1.9.9", b: "v1.9.10", want: -1},
		{name: "three segments before numeric fourth", a: "v1.9.3", b: "v1.9.3.1", want: -1},
		{name: "numeric fourth before labeled fourth", a: "v1.9.3.2", b: "v1.9.3-tenant.1", want: -1},
		{name: "same label by seq", a: "v1.9.3-tenant.2", b: "v1.9.3-tenant.10", want: -1},
		{name: "same numeric line by seq", a: "v1.9.3.2", b: "v1.9.3.10", want: -1},
		{name: "across labels lexical", a: "v1.9.3-tenant.1", b: "v1.9.3-zone.1", want: -1},
		{name: "across labels ignores seq", a: "v1.9.3-tenant.9", b: "v1.9.3-zone.1", want: -1},
		{name: "fourth segment loses to next release", a: "v1.9.3-tenant.9", b: "v1.9.4", want: -1},
		{name: "equal labeled versions", a: "v1.9.3-tenant.1", b: "v1.9.3-tenant.1", want: 0},
		{name: "equal numeric fourth", a: "v1.9.3.2", b: "v1.9.3.2", want: 0},
		{name: "folded zero equals itself", a: "v1.9.3.0", b: "v1.9.3.0", want: 0},
		// "v9..." > "v10..." as strings. The comparator must not do that.
		{name: "first segment numeric not lexical", a: "v9.0.0", b: "v10.0.0", want: -1},
		// A bigger third segment must not beat a bigger second segment.
		{name: "second segment beats a larger third", a: "v1.9.99", b: "v1.10.0", want: -1},
		// Any fourth segment on an older prefix still runs first.
		{name: "first segment beats any fourth segment", a: "v1.99.99-tenant.99", b: "v2.0.0", want: -1},
		{name: "zero version before the next patch", a: "v0.0.0", b: "v0.0.1", want: -1},
		// ".0" folds onto the three-segment version, so it stays before ".1".
		{name: "folded zero before numeric fourth", a: "v1.9.3.0", b: "v1.9.3.1", want: -1},
		{name: "folded zero before a labeled zero", a: "v1.9.3.0", b: "v1.9.3-tenant.0", want: -1},
		// Empty label is the numeric line. Its seq, however large, stays
		// before the first labeled line.
		{name: "large numeric seq before the earliest label", a: "v1.9.3.99", b: "v1.9.3-a.0", want: -1},
		{name: "labeled zero before the next seq", a: "v1.9.3-tenant.0", b: "v1.9.3-tenant.1", want: -1},
		// Label decides before seq, including when the earlier label has
		// the larger seq.
		{name: "label beats seq", a: "v1.9.3-a.9", b: "v1.9.3-b.0", want: -1},
		// "tenant" is a prefix of "tenant2" and must still sort first.
		{name: "shorter label before a longer sharing prefix", a: "v1.9.3-tenant.9", b: "v1.9.3-tenant2.0", want: -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := Parse(tc.a)
			require.NoError(t, err)
			b, err := Parse(tc.b)
			require.NoError(t, err)

			assert.Equal(t, tc.want, Compare(a, b))
			// The order must be antisymmetric, otherwise sorting is undefined.
			assert.Equal(t, -tc.want, Compare(b, a))
		})
	}
}

// orderedVersions is the execution order from the design document, from
// first to last. v1.9.3.0 is left out because it is the same version as
// v1.9.3 and would make the expectation ambiguous.
var orderedVersions = []string{
	"v1.9.3",
	"v1.9.3.1",
	"v1.9.3.2",
	"v1.9.3-tenant.1",
	"v1.9.3-tenant.2",
	"v1.9.4",
	"v1.9.10",
}

func TestSortOrder(t *testing.T) {
	parse := func(raws []string) []Version {
		out := make([]Version, 0, len(raws))
		for _, raw := range raws {
			v, err := Parse(raw)
			require.NoError(t, err)
			out = append(out, v)
		}
		return out
	}

	raws := func(vs []Version) []string {
		out := make([]string, 0, len(vs))
		for _, v := range vs {
			out = append(out, v.Raw)
		}
		return out
	}

	t.Run("shuffled input sorts to the documented order", func(t *testing.T) {
		shuffled := append([]string(nil), orderedVersions...)
		rnd := rand.New(rand.NewSource(1))
		rnd.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		got := parse(shuffled)
		sort.Slice(got, func(i, j int) bool { return Compare(got[i], got[j]) < 0 })

		assert.Equal(t, orderedVersions, raws(got))
	})

	t.Run("order is stable across repeated sorts of different shuffles", func(t *testing.T) {
		for seed := int64(0); seed < 20; seed++ {
			shuffled := append([]string(nil), orderedVersions...)
			rnd := rand.New(rand.NewSource(seed))
			rnd.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})

			got := parse(shuffled)
			sort.Slice(got, func(i, j int) bool { return Compare(got[i], got[j]) < 0 })

			assert.Equal(t, orderedVersions, raws(got), "seed %d", seed)
		}
	})
}

func TestParseSegmentBounds(t *testing.T) {
	maxSeg := strconv.FormatInt(int64(math.MaxInt), 10)
	// One past the largest int. On a 32-bit arch this is smaller than on
	// 64-bit, and it still does not fit in int either way.
	overflow := strconv.FormatUint(uint64(math.MaxInt)+1, 10)

	t.Run("largest int segment is accepted", func(t *testing.T) {
		got, err := Parse("v" + maxSeg + ".0.0")
		require.NoError(t, err)
		assert.Equal(t, math.MaxInt, got.P1)
		assert.Nil(t, got.Suffix)
	})

	t.Run("largest int as a numeric fourth segment", func(t *testing.T) {
		got, err := Parse("v1.9.3." + maxSeg)
		require.NoError(t, err)
		require.NotNil(t, got.Suffix)
		assert.Empty(t, got.Suffix.Label)
		assert.Equal(t, math.MaxInt, got.Suffix.Seq)
	})

	t.Run("one past the largest int is rejected", func(t *testing.T) {
		got, err := Parse("v" + overflow + ".0.0")
		assert.Error(t, err)
		assert.Equal(t, Version{}, got)
	})

	t.Run("one past the largest int in the fourth segment", func(t *testing.T) {
		got, err := Parse("v1.9.3." + overflow)
		assert.Error(t, err)
		assert.Equal(t, Version{}, got)
	})

	t.Run("one past the largest int in a labeled seq", func(t *testing.T) {
		got, err := Parse("v1.9.3-tenant." + overflow)
		assert.Error(t, err)
		assert.Equal(t, Version{}, got)
	})
}

// boundaryVersions mixes the cases a string sort gets wrong: a folded ".0",
// a numeric seq larger than any label seq, a label that is a prefix of
// another label, and v9 before v10.
var boundaryVersions = []string{
	"v0.0.0",
	"v1.9.3",
	"v1.9.3.0",
	"v1.9.3.1",
	"v1.9.3.99",
	"v1.9.3-a.0",
	"v1.9.3-tenant.0",
	"v1.9.3-tenant.1",
	"v1.9.3-tenant2.0",
	"v1.9.3-zone.0",
	"v1.9.4",
	"v1.9.10",
	"v1.10.0",
	"v2.0.0",
	"v9.0.0",
	"v10.0.0",
}

func TestSortTotalOrder(t *testing.T) {
	parse := func(raws []string) []Version {
		out := make([]Version, 0, len(raws))
		for _, raw := range raws {
			v, err := Parse(raw)
			require.NoError(t, err)
			out = append(out, v)
		}
		return out
	}

	// v1.9.3 and v1.9.3.0 compare equal, so a raw-string expectation would
	// depend on sort stability. Check the order relation instead.
	assertTotalOrder := func(t *testing.T, sorted []Version) {
		t.Helper()
		for i := range sorted {
			assert.Equal(t, 0, Compare(sorted[i], sorted[i]), "not reflexive at %s", sorted[i].Raw)
			for j := i + 1; j < len(sorted); j++ {
				got := Compare(sorted[i], sorted[j])
				if got > 0 {
					t.Errorf("%s sorts after %s", sorted[i].Raw, sorted[j].Raw)
				}
				assert.Equal(t, -got, Compare(sorted[j], sorted[i]))
			}
		}
	}

	t.Run("reversed input", func(t *testing.T) {
		raws := append([]string(nil), boundaryVersions...)
		for i, j := 0, len(raws)-1; i < j; i, j = i+1, j-1 {
			raws[i], raws[j] = raws[j], raws[i]
		}
		got := parse(raws)
		sort.Slice(got, func(i, j int) bool { return Compare(got[i], got[j]) < 0 })
		assertTotalOrder(t, got)
	})

	t.Run("shuffled input stays a total order", func(t *testing.T) {
		for seed := int64(0); seed < 20; seed++ {
			raws := append([]string(nil), boundaryVersions...)
			rnd := rand.New(rand.NewSource(seed))
			rnd.Shuffle(len(raws), func(i, j int) { raws[i], raws[j] = raws[j], raws[i] })

			got := parse(raws)
			sort.Slice(got, func(i, j int) bool { return Compare(got[i], got[j]) < 0 })
			assertTotalOrder(t, got)
		}
	})

	t.Run("strict positions around the ties", func(t *testing.T) {
		got := parse(boundaryVersions)
		sort.Slice(got, func(i, j int) bool { return Compare(got[i], got[j]) < 0 })

		index := map[string]int{}
		for i, v := range got {
			index[v.Raw] = i
		}
		before := func(earlier, later string) {
			t.Helper()
			assert.Less(t, index[earlier], index[later], "%s should sort before %s", earlier, later)
		}
		before("v0.0.0", "v1.9.3")
		before("v1.9.3", "v1.9.3.1")
		before("v1.9.3.0", "v1.9.3.1")
		before("v1.9.3.99", "v1.9.3-a.0")
		before("v1.9.3-a.0", "v1.9.3-tenant.0")
		before("v1.9.3-tenant.1", "v1.9.3-tenant2.0")
		before("v1.9.3-tenant2.0", "v1.9.3-zone.0")
		before("v1.9.10", "v1.10.0")
		before("v9.0.0", "v10.0.0")
		assert.Equal(t, 0, Compare(got[index["v1.9.3"]], got[index["v1.9.3.0"]]))
	})
}
