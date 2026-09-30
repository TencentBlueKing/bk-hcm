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

package engine

import (
	"fmt"

	"hcm/migrate/register"
	"hcm/migrate/util"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// recordTableDDL is the CREATE TABLE statement for the migration record table.
// It has no IF NOT EXISTS; CreateTableIfNotExists decides whether to run it.
var recordTableDDL = fmt.Sprintf("CREATE TABLE `%s` ("+
	"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键', "+
	"`migration_id` VARCHAR(%d) NOT NULL COMMENT '迁移 ID', "+
	"`version` VARCHAR(%d) NOT NULL COMMENT '最近一次成功时注册的版本', "+
	"`applied_pkg` VARCHAR(%d) NOT NULL DEFAULT '' COMMENT '执行该迁移 ID 的 Go 包', "+
	"`status` VARCHAR(16) NOT NULL COMMENT '执行状态，取值 running/success/failed', "+
	"`message` VARCHAR(%d) NOT NULL DEFAULT '' COMMENT '失败摘要', "+
	"`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', "+
	"`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', "+
	"PRIMARY KEY (`id`), "+
	"UNIQUE KEY `uidx_migration_id` (`migration_id`)"+
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='migration 执行记录表'",
	constant.MigrationRecordTable, constant.MigrationIDMaxLen, constant.MigrationVersionMaxLen,
	constant.MigrationPkgMaxLen, constant.MigrationRecordMessageMaxBytes)

// auditTableDDL is the CREATE TABLE statement for the migration audit table.
// It has no IF NOT EXISTS; CreateTableIfNotExists decides whether to run it.
var auditTableDDL = fmt.Sprintf("CREATE TABLE `%s` ("+
	"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键', "+
	"`run_id` VARCHAR(64) NOT NULL COMMENT '一次进程运行的 ID，各库共用，也是日志 rid', "+
	"`command` VARCHAR(16) NOT NULL COMMENT '子命令，取值 up/init', "+
	"`args` VARCHAR(%d) NOT NULL DEFAULT '' COMMENT '命令行参数', "+
	"`binary_version` VARCHAR(%d) NOT NULL COMMENT '二进制版本，未注入时为 debug', "+
	"`git_hash` VARCHAR(%d) NOT NULL COMMENT '二进制 commit，未注入时为 unknown', "+
	"`status` VARCHAR(16) NOT NULL COMMENT '运行状态，取值 running/success/failed', "+
	"`exit_code` INT NULL COMMENT '进程退出码，running 时为空', "+
	"`version_before` VARCHAR(%d) NOT NULL DEFAULT '' COMMENT '开始时的库当前版本，空表示无已定版记录', "+
	"`version_after` VARCHAR(%d) NOT NULL DEFAULT '' COMMENT '结束时的库当前版本', "+
	"`skipped` JSON NULL COMMENT '被跳过的迁移', "+
	"`warnings` JSON NULL COMMENT '运行中的警告，只追加', "+
	"`message` JSON NULL COMMENT '校验问题和错误，只追加', "+
	"`start_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间', "+
	"`end_at` DATETIME NULL COMMENT '结束时间，未结束为空', "+
	"PRIMARY KEY (`id`), "+
	"UNIQUE KEY `uidx_run_id` (`run_id`), "+
	"KEY `idx_start_at` (`start_at`)"+
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='migration 运行审计表'",
	constant.MigrationAuditTable, constant.MigrationAuditArgsMaxBytes, constant.MigrationAuditBuildMaxBytes,
	constant.MigrationAuditBuildMaxBytes, constant.MigrationVersionMaxLen, constant.MigrationVersionMaxLen)

// InitResult is the result of InitTables on one database.
type InitResult struct {
	// RecordCreated reports whether this call created the record table.
	RecordCreated bool
	// AuditCreated reports whether this call created the audit table.
	AuditCreated bool
	// Adopted lists the migrations recorded as baseline, empty unless this
	// call created the record table with a baseline.
	Adopted []register.Migration
}

// InitTables creates the audit table and the record table of one database,
// each only when it does not exist yet. The database is initialized when
// both tables exist; a call on an initialized database changes nothing.
//
// A nil baseline means init --mode=empty. With a baseline (--mode=adopt) and
// only when this call created the record table, every migration at or below
// baseline is recorded as success without running it.
//
// migrations must be in execution order, as returned by Registry.All.
func InitTables(kt *kit.Kit, o orm.Interface, baseline *register.Version,
	migrations []register.Migration) (InitResult, error) {

	auditCreated, err := util.CreateTableIfNotExists(kt.Ctx, o, constant.MigrationAuditTable, auditTableDDL)
	if err != nil {
		logs.Errorf("create migration audit table failed, err: %v, rid: %s", err, kt.Rid)
		return InitResult{}, err
	}

	recordCreated, adopted, err := initRecordTable(kt, o, baseline, migrations)
	if err != nil {
		return InitResult{AuditCreated: auditCreated}, err
	}

	if !auditCreated && !recordCreated {
		logs.Infof("migration tables already exist, skip init, rid: %s", kt.Rid)
	}
	return InitResult{RecordCreated: recordCreated, AuditCreated: auditCreated, Adopted: adopted}, nil
}

// initRecordTable creates the record table and, when it is created with a
// baseline, records the baseline migrations.
func initRecordTable(kt *kit.Kit, o orm.Interface, baseline *register.Version,
	migrations []register.Migration) (created bool, adopted []register.Migration, err error) {

	created, err = util.CreateTableIfNotExists(kt.Ctx, o, constant.MigrationRecordTable, recordTableDDL)
	if err != nil {
		logs.Errorf("create migration record table failed, err: %v, rid: %s", err, kt.Rid)
		return false, nil, err
	}
	if !created {
		logs.Infof("migration record table already exists, skip baseline, rid: %s", kt.Rid)
		return false, nil, nil
	}
	if baseline == nil {
		return true, nil, nil
	}

	adopted = baselineMigrations(*baseline, migrations)
	if err = insertBaseline(kt, o, adopted); err != nil {
		// Drop the record table created by this call so init can be retried.
		if dropErr := util.DropTable(kt.Ctx, o, constant.MigrationRecordTable); dropErr != nil {
			logs.Errorf("drop half initialized migration record table failed, err: %v, rid: %s", dropErr, kt.Rid)
			return false, nil, fmt.Errorf("insert baseline records failed and table %s must be dropped manually "+
				"before retrying init, insert err: %v, drop err: %v", constant.MigrationRecordTable, err, dropErr)
		}
		return false, nil, err
	}

	logs.Infof("init migration record table success, baseline: %s, adopted: %d, rid: %s",
		baseline.Raw, len(adopted), kt.Rid)
	return true, adopted, nil
}

// baselineMigrations returns the migrations at or below baseline. PENDING
// migrations are never included. A duplicate migration ID is kept only at
// its first position.
func baselineMigrations(baseline register.Version, migrations []register.Migration) []register.Migration {
	seen := make(map[string]struct{}, len(migrations))
	out := make([]register.Migration, 0, len(migrations))
	for _, m := range migrations {
		v, ok := m.ParsedVersion()
		if !ok || register.Compare(v, baseline) > 0 {
			continue
		}
		if _, dup := seen[m.ID]; dup {
			continue
		}
		seen[m.ID] = struct{}{}
		out = append(out, m)
	}
	return out
}

// insertBaseline records migrations as success in a single statement.
func insertBaseline(kt *kit.Kit, o orm.Interface, migrations []register.Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	rows := make([]Record, 0, len(migrations))
	for _, m := range migrations {
		rows = append(rows, Record{MigrationID: m.ID, Version: m.Version, AppliedPkg: m.Pkg,
			Status: enumor.MigrationStatusSuccess})
	}

	expr := fmt.Sprintf("INSERT INTO `%s` (`migration_id`, `version`, `applied_pkg`, `status`, `message`) "+
		"VALUES (:migration_id, :version, :applied_pkg, :status, :message)", constant.MigrationRecordTable)
	if err := o.Do().BulkInsert(kt.Ctx, expr, rows); err != nil {
		logs.Errorf("insert baseline migration records failed, err: %v, count: %d, rid: %s", err, len(rows), kt.Rid)
		return fmt.Errorf("insert baseline migration records failed, err: %v", err)
	}
	return nil
}
