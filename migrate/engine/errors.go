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

// Package engine runs registered migrations against the databases and keeps
// the per-database migration records.
package engine

import "errors"

// Errors wrapping these sentinels with %w are told apart by the CLI through
// errors.Is and mapped to exit codes; any other error is a plain failure.
var (
	// ErrUsage marks argument or config errors, exit code 2.
	ErrUsage = errors.New("usage error")
	// ErrPrecondition marks unmet preconditions, exit code 3. Load returns it
	// for an unknown status or an unparsable version. A missing record table
	// is a query error from Load and maps to exit code 1; the executor checks
	// the table before Load and wraps this sentinel.
	ErrPrecondition = errors.New("precondition failed")
	// ErrMissed marks a migration below the database version missed in the
	// default mode, exit code 4.
	ErrMissed = errors.New("missed migration")
)
