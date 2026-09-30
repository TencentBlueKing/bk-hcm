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

// Package register holds the in-process migration registries and the version
// grammar. Migration files must not import hcm/migrate/engine.
package register

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"

	"hcm/pkg/criteria/constant"
)

// versionRe matches vX.Y.Z, vX.Y.Z.N, or vX.Y.Z-<label>.N. A label must be
// lowercase and followed by a sequence number.
var versionRe = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:\.(\d+)|-([a-z][a-z0-9]*)\.(\d+))?$`)

// Suffix is the optional fourth segment of a version. An empty Label is a
// numeric sequence (".1"); a non-empty Label is a labeled sequence ("-tenant.1").
type Suffix struct {
	// Label is the feature label, empty for the numeric line.
	Label string
	// Seq orders changes within one line, compared numerically.
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
	return raw == constant.MigrationPendingVersion
}

// Parse turns a version string into a Version. It rejects anything outside
// the grammar, including constant.MigrationPendingVersion; callers that
// accept undecided migrations must check IsPending separately. A numeric
// fourth segment of 0 is folded away so "v1.9.3.0" equals "v1.9.3". A
// labeled segment is never folded.
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

// parseSegment converts one numeric segment. Leading zeros are rejected.
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

// Compare returns -1, 0 or +1 ordering a before b. It compares P1, P2 and P3
// numerically, then the fourth segment: absent before present; then Label,
// then Seq. Empty Label sorts before a non-empty Label.
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
