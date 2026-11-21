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

// Package migration is the migration bill_account_cloud_id.
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
const ID = "20240814-1145-BILL-ACCOUNT-CLOUD-ID-DC78"

func init() {
	register.Main.Regist(ID, "v1.6.1", "20240814114500", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0024_20240814_1145_bill_account_cloud_id.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_daily_pull_task",
		Column:  "root_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "root account cloud id",
		After:   "root_account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_daily_pull_task",
		Column:  "main_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "main account cloud id",
		After:   "main_account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_summary_daily",
		Column:  "root_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "root account cloud id",
		After:   "root_account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_summary_daily",
		Column:  "main_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "main account cloud id",
		After:   "main_account_id",
	}); err != nil {
		return err
	}
	if err := util.DropColumn(ctx, o, "account_bill_summary_main", "root_account_name"); err != nil {
		return err
	}
	if err := util.DropColumn(ctx, o, "account_bill_summary_main", "main_account_name"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_summary_main",
		Column:  "root_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "root account cloud id",
		After:   "root_account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_summary_main",
		Column:  "main_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "main account cloud id",
		After:   "main_account_id",
	}); err != nil {
		return err
	}
	if err := util.DropColumn(ctx, o, "account_bill_summary_root", "root_account_name"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_summary_root",
		Column:  "root_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "root account cloud id",
		After:   "root_account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_month_task",
		Column:  "root_account_cloud_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "root account cloud id",
		After:   "root_account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_sync_record",
		Column:  "count",
		Type:    "bigint",
		NotNull: true,
		Default: util.ExprDefault("0"),
		After:   "currency",
	}); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_daily_pull_task AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id\n"+
		"SET ab.root_account_cloud_id = ra.cloud_id\n"+
		"WHERE ab.root_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_daily_pull_task AS ab JOIN main_account AS ma ON ab.main_account_id = ma.id\n"+
		"SET ab.main_account_cloud_id = ma.cloud_id\n"+
		"WHERE ab.main_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_summary_daily AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id\n"+
		"SET ab.root_account_cloud_id = ra.cloud_id\n"+
		"WHERE ab.root_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_summary_daily AS ab JOIN main_account AS ma ON ab.main_account_id = ma.id\n"+
		"SET ab.main_account_cloud_id = ma.cloud_id\n"+
		"WHERE ab.main_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_summary_main AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id\n"+
		"SET ab.root_account_cloud_id = ra.cloud_id\n"+
		"WHERE ab.root_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_summary_main AS ab JOIN main_account AS ma ON ab.main_account_id = ma.id\n"+
		"SET ab.main_account_cloud_id = ma.cloud_id\n"+
		"WHERE ab.main_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_summary_root AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id\n"+
		"SET ab.root_account_cloud_id = ra.cloud_id\n"+
		"WHERE ab.root_account_cloud_id = ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE account_bill_month_task AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id\n"+
		"SET ab.root_account_cloud_id = ra.cloud_id\n"+
		"WHERE ab.root_account_cloud_id = ''"); err != nil {
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
