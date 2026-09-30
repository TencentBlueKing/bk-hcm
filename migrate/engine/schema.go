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
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// RecordTable is the migration record table kept in every migrated database.
const RecordTable = "hcm_migration_record"

// recordTableDDL has no IF NOT EXISTS on purpose: util.CreateTableIfNotExists
// decides whether to run it, and reports whether this call created the table.
// The version column is register.MaxVersionLen wide, the same limit Regist
// enforces, so a version that registered can always be stored whole.
// utf8mb4_bin makes the migration_id unique key compare bytes, the same way
// Go compares the ID.
var recordTableDDL = fmt.Sprintf("CREATE TABLE `%s` ("+
	"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键', "+
	"`migration_id` VARCHAR(64) NOT NULL COMMENT '迁移 ID', "+
	"`version` VARCHAR(%d) NOT NULL COMMENT '最近一次成功时注册的版本', "+
	"`status` VARCHAR(16) NOT NULL COMMENT '执行状态，取值 running/success/failed', "+
	"`message` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '失败摘要', "+
	"`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', "+
	"`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', "+
	"PRIMARY KEY (`id`), "+
	"UNIQUE KEY `uidx_migration_id` (`migration_id`)"+
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='migration 执行记录表'",
	RecordTable, register.MaxVersionLen)

// InitRecordTable creates the record table of one database.
//
// A nil baseline means init --mode=empty. With a baseline (--mode=adopt) and
// only when this call created the table, every migration at or below baseline
// is recorded as success without running it. When the table already exists
// the call changes nothing and returns created false.
//
// migrations must be in execution order, as returned by Registry.All. The
// returned adopted slice lists the migrations recorded as baseline.
func InitRecordTable(kt *kit.Kit, o orm.Interface, baseline *register.Version,
	migrations []register.Migration) (created bool, adopted []register.Migration, err error) {

	created, err = util.CreateTableIfNotExists(kt.Ctx, o, RecordTable, recordTableDDL)
	if err != nil {
		logs.Errorf("create migration record table failed, err: %v, rid: %s", err, kt.Rid)
		return false, nil, err
	}
	if !created {
		logs.Infof("migration record table already exists, skip init, rid: %s", kt.Rid)
		return false, nil, nil
	}
	if baseline == nil {
		return true, nil, nil
	}

	adopted = baselineMigrations(*baseline, migrations)
	if err = insertBaseline(kt, o, adopted); err != nil {
		// A partial baseline leaves an empty table. The next init treats that table
		// as already initialized and does nothing, then up reruns every migration
		// at or below the baseline. Drop the table this call created so init can
		// be retried as a whole.
		if dropErr := util.DropTable(kt.Ctx, o, RecordTable); dropErr != nil {
			logs.Errorf("drop half initialized migration record table failed, err: %v, rid: %s", dropErr, kt.Rid)
			return false, nil, fmt.Errorf("insert baseline records failed and table %s must be dropped manually "+
				"before retrying init, insert err: %v, drop err: %v", RecordTable, err, dropErr)
		}
		return false, nil, err
	}

	logs.Infof("init migration record table success, baseline: %s, adopted: %d, rid: %s",
		baseline.Raw, len(adopted), kt.Rid)
	return true, adopted, nil
}

// baselineMigrations returns the migrations at or below baseline. PENDING
// migrations are never included. A migration ID registered more than once is
// kept only at its first position, so each ID yields a single record.
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

// insertBaseline records migrations as success in a single statement, so the
// baseline is either written entirely or not at all.
func insertBaseline(kt *kit.Kit, o orm.Interface, migrations []register.Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	rows := make([]Record, 0, len(migrations))
	for _, m := range migrations {
		rows = append(rows, Record{MigrationID: m.ID, Version: m.Version, Status: StatusSuccess})
	}

	expr := fmt.Sprintf("INSERT INTO `%s` (`migration_id`, `version`, `status`, `message`) "+
		"VALUES (:migration_id, :version, :status, :message)", RecordTable)
	if err := o.Do().BulkInsert(kt.Ctx, expr, rows); err != nil {
		logs.Errorf("insert baseline migration records failed, err: %v, count: %d, rid: %s", err, len(rows), kt.Rid)
		return fmt.Errorf("insert baseline migration records failed, err: %v", err)
	}
	return nil
}
