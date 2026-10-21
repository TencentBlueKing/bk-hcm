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

// Package migration is the migration multiple_month_task.
package migration

import (
	"context"
	"fmt"

	"hcm/migrate/register"
	"hcm/migrate/util"
	"hcm/pkg/dal/dao/orm"
)

// ID is the unique identifier of this migration file, in the form
// <date>-<time>-<desc>-<random>. It must never change once executed.
// Migrations that share an ID run only the earlier one in execution order.
const ID = "20241009-0000-MULTIPLE-MONTH-TASK-0EF3"

func init() {
	register.Main.Regist(ID, "v1.6.8", "20241009000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0025_20241009_multiple_month_task.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_month_task",
		Column:  "type",
		Type:    "VARCHAR(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "任务类型",
		After:   "vendor",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "account_bill_month_task", "idx_root_account_id_year_month"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "account_bill_month_task", "idx_root_account_id_year_month_type", []string{"root_account_id", "bill_year", "bill_month", "type"}, true); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_month_task SET summary_detail = '[]' WHERE summary_detail=''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE account_bill_month_task MODIFY `summary_detail` JSON NOT NULL"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_sync_record SET detail = '[]' WHERE detail=''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE account_bill_sync_record MODIFY `detail` JSON NOT NULL"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "load_balancer_target", "index_target_group_id", []string{"target_group_id"}, false); err != nil {
		return err
	}
	return nil
}

// execSQL runs a statement that util has no idempotent helper for. Re-running it is safe, see the plan.
func execSQL(ctx context.Context, o orm.Interface, sql string) error {
	if _, err := o.Do().Exec(ctx, sql); err != nil {
		return fmt.Errorf("exec %q failed, err: %v", sql, err)
	}
	return nil
}
