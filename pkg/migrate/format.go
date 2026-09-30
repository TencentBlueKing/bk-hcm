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

// Package migrate provides the format checks and string helpers of hcm-migrate.
package migrate

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"hcm/pkg/criteria/constant"
)

// migrationIDRe matches yyyyMMdd-HHmm-TAG-XXXX, e.g. 20260923-1945-ADD-BK-ASSET-ID-A3F9.
var migrationIDRe = regexp.MustCompile(`^\d{8}-\d{4}-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[0-9A-F]{4}$`)

// migrationIDTimeLen is the length of the yyyyMMdd-HHmm head of a migration ID.
const migrationIDTimeLen = 13

// ValidateMigrationID checks the format, the length and the creation time of a migration ID.
func ValidateMigrationID(id string) error {
	if len(id) > constant.MigrationIDMaxLen {
		return fmt.Errorf("invalid migration id, longer than %d characters", constant.MigrationIDMaxLen)
	}
	if !migrationIDRe.MatchString(id) {
		return fmt.Errorf("invalid migration id, want yyyyMMdd-HHmm-TAG-XXXX in uppercase")
	}
	if _, err := time.Parse("20060102-1504", id[:migrationIDTimeLen]); err != nil {
		return fmt.Errorf("invalid migration id, want a real yyyyMMdd-HHmm")
	}
	return nil
}

// timestampRe matches the 14-digit creation timestamp yyyyMMddHHmmss.
var timestampRe = regexp.MustCompile(`^\d{14}$`)

// ValidateTimestamp checks that ts is 14 digits and a real yyyyMMddHHmmss time.
func ValidateTimestamp(ts string) error {
	if !timestampRe.MatchString(ts) {
		return fmt.Errorf("invalid timestamp, want 14 digits")
	}
	if _, err := time.Parse("20060102150405", ts); err != nil {
		return fmt.Errorf("invalid timestamp, want yyyyMMddHHmmss")
	}
	return nil
}

// migrationDirRe matches the last element of a migration package path, an
// optional version prefix followed by the migration suffix:
//
//	v1.9.3_20260905160000_add_bk_asset_id
//	20260920153012_add_legacy_asset
var migrationDirRe = regexp.MustCompile(`^(?:(v[^_/]+)_)?((\d{14})_[a-z0-9]+(?:_[a-z0-9]+)*)$`)

// ParsePkgPath strips constant.MigrationPkgPrefix from pkgPath and returns the
// rest. The rest must be database/group/dir, with the given database and a
// dir whose timestamp equals the given timestamp.
func ParsePkgPath(database, pkgPath, timestamp string) (string, error) {
	pkg, found := strings.CutPrefix(pkgPath, constant.MigrationPkgPrefix)
	if !found {
		return "", fmt.Errorf("package %q is not under %s", pkgPath, constant.MigrationPkgPrefix)
	}

	parts := strings.Split(pkg, "/")
	if len(parts) != 3 || parts[1] == "" {
		return "", fmt.Errorf("package %q is not database/group/dir under %s", pkgPath, constant.MigrationPkgPrefix)
	}
	if parts[0] != database {
		return "", fmt.Errorf("package %q is under database %q, not %q", pkgPath, parts[0], database)
	}

	match := migrationDirRe.FindStringSubmatch(parts[2])
	if match == nil {
		return "", fmt.Errorf("package %q has an invalid directory name, want [<version>_]<timestamp>_<tag>",
			pkgPath)
	}
	if match[3] != timestamp {
		return "", fmt.Errorf("package %q has timestamp %s, registered %s", pkgPath, match[3], timestamp)
	}
	if len(pkg) > constant.MigrationPkgMaxLen {
		return "", fmt.Errorf("package %q is longer than %d characters after %s", pkgPath,
			constant.MigrationPkgMaxLen, constant.MigrationPkgPrefix)
	}

	return pkg, nil
}

// PkgSuffix returns the migration suffix of pkg, the last path element
// without its version prefix, e.g. 20260905160000_add_bk_asset_id. ok is
// false when the last element is not a migration directory name.
func PkgSuffix(pkg string) (suffix string, ok bool) {
	dir := pkg[strings.LastIndex(pkg, "/")+1:]
	match := migrationDirRe.FindStringSubmatch(dir)
	if match == nil {
		return "", false
	}
	return match[2], true
}

// TruncateUTF8 cuts s to at most limit bytes without splitting a multi-byte
// character. Invalid UTF-8 is replaced with "?" first.
func TruncateUTF8(s string, limit int) string {
	s = strings.ToValidUTF8(s, "?")
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
