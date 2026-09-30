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
	"errors"

	"hcm/pkg/criteria/constant"
)

// Errors wrapping these sentinels with %w are told apart by the CLI through
// errors.Is and mapped to exit codes; any other error is a plain failure,
// exit code 1. An error wraps at most one of them.
var (
	// ErrUsage marks argument or config errors, exit code 2.
	ErrUsage = errors.New("usage error")
	// ErrPrecondition marks a database whose migration tables are missing or
	// hold invalid records, exit code 3. CheckInitialized returns it for a
	// missing table; RecordStore.Load returns it for an unknown status, an
	// unparsable version, or a success record without applied_pkg.
	ErrPrecondition = errors.New("precondition failed")
	// ErrMissed marks a migration below the database version missed in the
	// default mode, exit code 4. Rerunning with --catch-up runs it.
	ErrMissed = errors.New("missed migration")
	// ErrRegistry marks invalid registry content, exit code 5: a PENDING
	// migration without --allow-pending, or more than one version label.
	ErrRegistry = errors.New("invalid registry")
	// ErrIDReuse marks a migration skipped by an ID applied by a package with
	// another migration suffix, exit code 6.
	ErrIDReuse = errors.New("suspected migration id reuse")
)

// ExitCode returns the process exit code of err: constant.MigrationExitSuccess
// for nil, the code of the sentinel err wraps, or constant.MigrationExitFailure.
func ExitCode(err error) int {
	switch {
	case err == nil:
		return constant.MigrationExitSuccess
	case errors.Is(err, ErrUsage):
		return constant.MigrationExitUsage
	case errors.Is(err, ErrPrecondition):
		return constant.MigrationExitPrecondition
	case errors.Is(err, ErrMissed):
		return constant.MigrationExitMissed
	case errors.Is(err, ErrRegistry):
		return constant.MigrationExitRegistry
	case errors.Is(err, ErrIDReuse):
		return constant.MigrationExitIDReuse
	default:
		return constant.MigrationExitFailure
	}
}
