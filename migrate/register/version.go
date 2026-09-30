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

// Package register holds the in-process migration registries for hcm-migrate,
// plus the version grammar every migration file is keyed by. Migration files
// import only this package and hcm/migrate/util, never hcm/migrate/engine.
package register

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"
)

// PendingVersion is the placeholder registered by a migration whose release
// version is not decided yet. Migration files must use this constant instead
// of writing the literal. It is not a parsable version on purpose: init and
// up scan for it and refuse to run, so an undecided migration can never be
// executed under a guessed version. The directory holding those files stays
// lowercase, "pending/".
const PendingVersion = "PENDING"

// versionRe matches the three accepted spellings in one pass:
//
//	v1.9.3              three segments, no fourth
//	v1.9.3.1            numeric fourth segment
//	v1.9.3-tenant.1     labeled fourth segment
//
// A label must be lowercase and must be followed by a sequence number, so
// "v1.9.3-tenant" and "v1.9.3-Tenant.1" are rejected.
var versionRe = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:\.(\d+)|-([a-z][a-z0-9]*)\.(\d+))?$`)

// Suffix is the optional fourth segment of a version. An empty Label means a
// plain numeric sequence used by the internal line (".1"); a non-empty Label
// marks a feature branch ("-tenant.1").
type Suffix struct {
	// Label is the feature branch name, empty for the internal numeric line.
	Label string
	// Seq orders the specific changes within one line, compared numerically.
	Seq int
}

// Version is the sort key of a migration. A nil Suffix means the version has
// no fourth segment.
type Version struct {
	// P1, P2 and P3 are the three leading numbers, aligned with the product version.
	P1, P2, P3 int
	// Suffix is the fourth segment, nil when absent.
	Suffix *Suffix
	// Raw is the string this version was parsed from.
	Raw string
}

// IsPending reports whether raw is the undecided-version placeholder.
func IsPending(raw string) bool {
	return raw == PendingVersion
}

// Parse turns a version string into a Version. It rejects anything the
// grammar does not allow, including the PendingVersion placeholder: callers
// that accept undecided migrations must check IsPending separately.
//
// A numeric fourth segment of 0 is folded away, so "v1.9.3.0" parses to the
// same version as "v1.9.3" and the two compare equal. A labeled segment is
// never folded: "v1.9.3-tenant.0" keeps its label and still sorts after
// "v1.9.3".
func Parse(raw string) (Version, error) {
	m := versionRe.FindStringSubmatch(raw)
	if m == nil {
		return Version{}, fmt.Errorf("invalid version %q, want vX.Y.Z, vX.Y.Z.N or vX.Y.Z-<label>.N", raw)
	}

	nums := make([]int, 0, 3)
	for i, seg := range []string{m[1], m[2], m[3]} {
		n, err := parseSegment(seg)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q, segment %d: %v", raw, i+1, err)
		}
		nums = append(nums, n)
	}

	v := Version{P1: nums[0], P2: nums[1], P3: nums[2], Raw: raw}

	switch {
	case m[4] != "":
		seq, err := parseSegment(m[4])
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q, fourth segment: %v", raw, err)
		}
		// A plain ".0" carries no ordering information, fold it away so that
		// v1.9.3.0 and v1.9.3 are one version rather than two adjacent ones.
		if seq != 0 {
			v.Suffix = &Suffix{Seq: seq}
		}
	case m[5] != "":
		seq, err := parseSegment(m[6])
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q, fourth segment: %v", raw, err)
		}
		v.Suffix = &Suffix{Label: m[5], Seq: seq}
	}

	return v, nil
}

// parseSegment converts one numeric segment, rejecting leading zeros so that
// a version string and its parsed form stay one to one ("v1.09.3" would
// otherwise be a second spelling of "v1.9.3").
func parseSegment(seg string) (int, error) {
	if len(seg) > 1 && seg[0] == '0' {
		return 0, fmt.Errorf("leading zero in %q", seg)
	}
	n, err := strconv.Atoi(seg)
	if err != nil {
		return 0, fmt.Errorf("not a number: %q", seg)
	}
	return n, nil
}

// Compare returns -1, 0 or +1 ordering a before b, equal, or a after b. It
// compares P1, P2 and P3 numerically, then the fourth segment: absent sorts
// before present, so within one prefix the three-segment file runs first;
// two present segments compare by Label and then by Seq numerically.
//
// Label is compared before Seq because the empty label is the internal
// numeric line, and the whole numeric line must run before a feature branch
// line on the same prefix: v1.9.3.9 before v1.9.3-tenant.1. Comparing Seq
// alone would call those two equal and interleave the lines.
//
// Between two different non-empty labels the lexicographic result carries no
// meaning. A registry holding two labels is rejected before anything is
// sorted. It is kept so that 0 always means "the same version", which the
// --to ceiling and the database current version both rely on.
//
// Migrations sharing a version are further ordered by timestamp and then
// migration ID. Those two fields live on the migration, not on the version,
// so the registry applies them on top of this result.
func Compare(a, b Version) int {
	if c := cmp.Or(cmp.Compare(a.P1, b.P1), cmp.Compare(a.P2, b.P2), cmp.Compare(a.P3, b.P3)); c != 0 {
		return c
	}

	switch {
	case a.Suffix == nil && b.Suffix == nil:
		return 0
	case a.Suffix == nil:
		return -1
	case b.Suffix == nil:
		return 1
	}

	return cmp.Or(cmp.Compare(a.Suffix.Label, b.Suffix.Label), cmp.Compare(a.Suffix.Seq, b.Suffix.Seq))
}
