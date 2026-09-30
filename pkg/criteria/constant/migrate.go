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

package constant

const (
	// MigrationPkgPrefix is the import path prefix of every migration package.
	MigrationPkgPrefix = "hcm/migrate/migrations/"

	// MigrationDatabaseMain is the migration database name of the main database.
	MigrationDatabaseMain = "main"
	// MigrationDatabaseObs is the migration database name of the OBS database.
	MigrationDatabaseObs = "obs"

	// MigrationPendingVersion is the version of a migration whose release version is not decided yet.
	MigrationPendingVersion = "PENDING"

	// MigrationIDMaxLen is the maximum length of a migration ID.
	MigrationIDMaxLen = 64
	// MigrationVersionMaxLen is the maximum length of a migration version.
	MigrationVersionMaxLen = 64
	// MigrationPkgMaxLen is the maximum length of a migration package path without MigrationPkgPrefix.
	MigrationPkgMaxLen = 255
)

const (
	// MigrationRecordTable is the migration record table of every migrated database.
	MigrationRecordTable = "hcm_migration_record"
	// MigrationRecordMessageMaxBytes is the maximum bytes of the record message column.
	MigrationRecordMessageMaxBytes = 1024

	// MigrationAuditTable is the migration audit table of every migrated database.
	MigrationAuditTable = "hcm_migration_audit"
	// MigrationAuditArgsMaxBytes is the maximum bytes of the audit args column.
	MigrationAuditArgsMaxBytes = 1024
	// MigrationAuditBuildMaxBytes is the maximum bytes of the audit binary_version and git_hash columns.
	MigrationAuditBuildMaxBytes = 64
	// MigrationAuditItemMaxBytes is the maximum bytes of one item in the audit warnings and message columns.
	MigrationAuditItemMaxBytes = 1024
	// MigrationAuditMaxItems is the maximum items of the audit warnings and message columns.
	MigrationAuditMaxItems = 200
)
