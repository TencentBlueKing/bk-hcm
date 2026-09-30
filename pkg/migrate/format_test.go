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

package migrate

import (
	"fmt"
	"strings"
	"testing"

	"hcm/pkg/criteria/constant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMigrationID(t *testing.T) {
	id64 := "20260905-1600-" + "A" + strings.Repeat("B", 44) + "-A3F9"
	require.Equal(t, constant.MigrationIDMaxLen, len(id64))
	id65 := "20260905-1600-" + "A" + strings.Repeat("B", 45) + "-A3F9"
	require.Equal(t, constant.MigrationIDMaxLen+1, len(id65))

	testCases := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "minimal single letter tag", id: "20260905-1600-A-0001"},
		{name: "tag with digits", id: "20260905-1600-ADD2-A3F9"},
		{name: "multi segment tag", id: "20260923-1945-ADD-BK-ASSET-ID-A3F9"},
		{name: "segment of only digits after first", id: "20260905-1600-ADD-1234-A3F9"},
		{name: "length exactly 64", id: id64},
		{name: "first tag segment starts with digit", id: "20260905-1600-1ADD-A3F9", wantErr: true},
		{name: "lowercase in tag", id: "20260905-1600-Add-A3F9", wantErr: true},
		{name: "lowercase hex suffix", id: "20260905-1600-ADD-a3f9", wantErr: true},
		{name: "double hyphen", id: "20260905-1600-ADD--BK-A3F9", wantErr: true},
		{name: "trailing hyphen before hex", id: "20260905-1600-ADD--A3F9", wantErr: true},
		{name: "missing tag", id: "20260905-1600-A3F9", wantErr: true},
		{name: "3 hex digits", id: "20260905-1600-ADD-A3F", wantErr: true},
		{name: "5 hex digits", id: "20260905-1600-ADD-A3F90", wantErr: true},
		{name: "length 65", id: id65, wantErr: true},
		{name: "month 13", id: "20261301-1200-ADD-A3F9", wantErr: true},
		{name: "february 30", id: "20260230-1200-ADD-A3F9", wantErr: true},
		{name: "hour 24", id: "20260905-2400-ADD-A3F9", wantErr: true},
		{name: "minute 60", id: "20260905-1660-ADD-A3F9", wantErr: true},
		{name: "old uuid", id: "0a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d", wantErr: true},
		{name: "empty", id: "", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMigrationID(tc.id)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateTimestamp(t *testing.T) {
	testCases := []struct {
		name    string
		ts      string
		wantErr bool
	}{
		{name: "valid", ts: "20260905160000"},
		{name: "leap day", ts: "20240229000000"},
		{name: "last second of a day", ts: "20261231235959"},
		{name: "too short", ts: "2026090516", wantErr: true},
		{name: "too long", ts: "202609051600000", wantErr: true},
		{name: "empty", ts: "", wantErr: true},
		{name: "non digit", ts: "2026090516000a", wantErr: true},
		{name: "space", ts: "2026090516000 ", wantErr: true},
		{name: "all zero", ts: "00000000000000", wantErr: true},
		{name: "month 13", ts: "20261301000000", wantErr: true},
		{name: "day 00", ts: "20260900000000", wantErr: true},
		{name: "september 31", ts: "20260931000000", wantErr: true},
		{name: "february 29 outside a leap year", ts: "20260229000000", wantErr: true},
		{name: "hour 24", ts: "20260905240000", wantErr: true},
		{name: "minute 60", ts: "20260905166000", wantErr: true},
		{name: "second 60", ts: "20260905160060", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTimestamp(tc.ts)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

const (
	tsDefault  = "20260905160000"
	tagDefault = "add_x"
)

func releasedPkg(database, version, timestamp, tag string) string {
	return fmt.Sprintf("%s%s/%s/%s_%s_%s", constant.MigrationPkgPrefix, database, version, version, timestamp, tag)
}

func pendingPkg(database, timestamp, tag string) string {
	return fmt.Sprintf("%s%s/pending/%s_%s", constant.MigrationPkgPrefix, database, timestamp, tag)
}

func TestParsePkgPath(t *testing.T) {
	pkg255 := constant.MigrationPkgPrefix + "main/g/" + tsDefault + "_" + strings.Repeat("a", 233)
	require.Equal(t, constant.MigrationPkgMaxLen, len(strings.TrimPrefix(pkg255, constant.MigrationPkgPrefix)))
	pkg256 := constant.MigrationPkgPrefix + "main/g/" + tsDefault + "_" + strings.Repeat("a", 234)
	require.Equal(t, constant.MigrationPkgMaxLen+1, len(strings.TrimPrefix(pkg256, constant.MigrationPkgPrefix)))

	testCases := []struct {
		name     string
		database string
		pkgPath  string
		ts       string
		wantPkg  string
		wantErr  bool
	}{
		{
			name:     "valid released path",
			database: "main",
			pkgPath:  releasedPkg("main", "v1.9.3", tsDefault, "add_bk_asset_id"),
			ts:       tsDefault,
			wantPkg:  "main/v1.9.3/v1.9.3_" + tsDefault + "_add_bk_asset_id",
		},
		{
			name:     "valid pending path",
			database: "main",
			pkgPath:  pendingPkg("main", tsDefault, "add_legacy_asset"),
			ts:       tsDefault,
			wantPkg:  "main/pending/" + tsDefault + "_add_legacy_asset",
		},
		{
			name:     "path without prefix",
			database: "main",
			pkgPath:  "main/v1.9.3/v1.9.3_" + tsDefault + "_add_x",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "prefix only",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix,
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "2 segments",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix + "main/" + tsDefault + "_add_x",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "4 segments",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix + "main/g/extra/" + tsDefault + "_add_x",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "empty group",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix + "main//" + tsDefault + "_add_x",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "obs registry with main path",
			database: "obs",
			pkgPath:  releasedPkg("main", "v1.9.3", tsDefault, tagDefault),
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "uppercase in dir tag",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix + "main/pending/" + tsDefault + "_ADD_x",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "dir timestamp 13 digits",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix + "main/pending/2026090516000_add_x",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "dir timestamp mismatch",
			database: "main",
			pkgPath:  pendingPkg("main", "20260905160001", tagDefault),
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "trailing underscore in tag",
			database: "main",
			pkgPath:  constant.MigrationPkgPrefix + "main/pending/" + tsDefault + "_add_",
			ts:       tsDefault,
			wantErr:  true,
		},
		{
			name:     "stripped length exactly 255",
			database: "main",
			pkgPath:  pkg255,
			ts:       tsDefault,
			wantPkg:  strings.TrimPrefix(pkg255, constant.MigrationPkgPrefix),
		},
		{
			name:     "stripped length 256",
			database: "main",
			pkgPath:  pkg256,
			ts:       tsDefault,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePkgPath(tc.database, tc.pkgPath, tc.ts)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, "", got)
				if tc.name == "dir timestamp 13 digits" {
					assert.Contains(t, err.Error(), "invalid directory name")
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantPkg, got)
		})
	}
}

func TestPkgSuffix(t *testing.T) {
	testCases := []struct {
		name       string
		pkg        string
		wantSuffix string
		wantOK     bool
	}{
		{
			name:       "stripped path",
			pkg:        "main/v1.9.3.x/v1.9.3.1_20260905160000_add_bk_asset_id",
			wantSuffix: "20260905160000_add_bk_asset_id",
			wantOK:     true,
		},
		{
			name:       "full path with prefix",
			pkg:        constant.MigrationPkgPrefix + "main/v1.9.3.x/v1.9.3.1_20260905160000_add_bk_asset_id",
			wantSuffix: "20260905160000_add_bk_asset_id",
			wantOK:     true,
		},
		{
			name:       "bare dir without slash",
			pkg:        "v1.9.3_20260905160000_add_bk_asset_id",
			wantSuffix: "20260905160000_add_bk_asset_id",
			wantOK:     true,
		},
		{
			name:       "pending dir",
			pkg:        "main/pending/20260920153012_add_legacy_asset",
			wantSuffix: "20260920153012_add_legacy_asset",
			wantOK:     true,
		},
		{name: "invalid dir", pkg: "main/pending/not_a_migration", wantOK: false},
		{name: "empty", pkg: "", wantOK: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := PkgSuffix(tc.pkg)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantSuffix, got)
		})
	}

	t.Run("same suffix across version prefix and group", func(t *testing.T) {
		a, okA := PkgSuffix("main/v1.9.3.x/v1.9.3.1_20260905160000_add_bk_asset_id")
		b, okB := PkgSuffix("main/pending/20260905160000_add_bk_asset_id")
		require.True(t, okA)
		require.True(t, okB)
		assert.Equal(t, a, b)
	})

	t.Run("different tag different suffix", func(t *testing.T) {
		a, okA := PkgSuffix("main/pending/20260905160000_add_bk_asset_id")
		b, okB := PkgSuffix("main/pending/20260905160000_add_other")
		require.True(t, okA)
		require.True(t, okB)
		assert.NotEqual(t, a, b)
	})
}
