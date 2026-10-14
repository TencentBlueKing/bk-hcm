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

// Package migration is the migration sub_account_async_flow.
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
const ID = "20231130-1604-SUB-ACCOUNT-ASYNC-FLOW-F056"

func init() {
	register.Main.Regist(ID, "v1.2.1", "20231130160400", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0012_20231130_1604.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "sub_account", "create table if not exists `sub_account`\n"+
		"(\n"+
		"    `id`         varchar(64)  not null,\n"+
		"    `cloud_id`   varchar(255) not null,\n"+
		"    `name`       varchar(255) not null,\n"+
		"    `vendor`     varchar(16)  not null,\n"+
		"    `site`       varchar(32)  not null,\n"+
		"    `account_id` varchar(64)  not null,\n"+
		"    `extension`  json         not null,\n"+
		"    `managers`   json,\n"+
		"    `bk_biz_ids` json,\n"+
		"    `memo`       varchar(255)          default '',\n"+
		"    `creator`    varchar(64)  not null,\n"+
		"    `reviser`    varchar(64)  not null,\n"+
		"    `created_at` timestamp    not null default current_timestamp,\n"+
		"    `updated_at` timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_vendor_cloud_id` (`vendor`, `cloud_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "sub_account", "0"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE aws_security_group_rule modify column memo varchar(255) default ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "delete\n"+
		"from aws_region"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "aws_region",
		Column:  "account_id",
		Type:    "varchar(64)",
		NotNull: true,
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "aws_region", "idx_uk_region_id_status"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "aws_region", "idx_uk_account_id_region_id", []string{"account_id", "region_id"}, true); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "user_collection", "create table if not exists `user_collection`\n"+
		"(\n"+
		"    `id`         varchar(64) not null,\n"+
		"    `user`       varchar(64) not null,\n"+
		"    `res_type`   varchar(50) not null,\n"+
		"    `res_id`     varchar(64) not null,\n"+
		"    `creator`    varchar(64) not null,\n"+
		"    `created_at` timestamp   not null default current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_user_res_type_res_id` (`user`, `res_type`, `res_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "user_collection", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_sync_detail", "create table if not exists `account_sync_detail`\n"+
		"(\n"+
		"    `id`                varchar(64) not null,\n"+
		"    `vendor`            varchar(16) not null,\n"+
		"    `account_id`        varchar(64) not null,\n"+
		"    `res_name`          varchar(64) not null,\n"+
		"    `res_status`        varchar(64) not null,\n"+
		"    `res_end_time`      varchar(64)          default '',\n"+
		"    `res_failed_reason` json                 default null,\n"+
		"    `creator`           varchar(64) not null,\n"+
		"    `reviser`           varchar(64) not null,\n"+
		"    `created_at`        timestamp   not null default current_timestamp,\n"+
		"    `updated_at`        timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_vendor_account_id_res_name` (`vendor`, `account_id`, `res_name`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_sync_detail", "0"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account",
		Column:  "recycle_reserve_time",
		Type:    "bigint",
		Default: util.ExprDefault("-1"),
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "recycle_record",
		Column:  "recycled_at",
		Type:    "timestamp",
		NotNull: true,
		Default: util.ExprDefault("current_timestamp"),
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "sub_account",
		Column:  "account_type",
		Type:    "varchar(64)",
		Default: util.StringDefault(""),
	}); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "async_flow", "create table if not exists `async_flow`\n"+
		"(\n"+
		"    `id`         varchar(64) not null,\n"+
		"    `name`       varchar(64) not null,\n"+
		"    `state`      varchar(16) not null,\n"+
		"    `reason`     json                 default null,\n"+
		"    `share_data` json                 default null,\n"+
		"    `memo`       varchar(64) not null,\n"+
		"    `worker`     varchar(64) not null,\n"+
		"    `creator`    varchar(64) not null,\n"+
		"    `reviser`    varchar(64) not null,\n"+
		"    `created_at` timestamp   not null default current_timestamp,\n"+
		"    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "async_flow", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "async_flow_task", "create table if not exists `async_flow_task`\n"+
		"(\n"+
		"    `id`          varchar(64) not null,\n"+
		"    `flow_id`     varchar(64) not null,\n"+
		"    `flow_name`   varchar(64) not null,\n"+
		"    `action_id`   varchar(64) not null,\n"+
		"    `action_name` varchar(64) not null,\n"+
		"    `params`      json                 default null,\n"+
		"    `retry`       json        not null,\n"+
		"    `depend_on`   varchar(64)          default '',\n"+
		"    `state`       varchar(16) not null,\n"+
		"    `reason`      json                 default null,\n"+
		"    `result`      json                 default null,\n"+
		"    `creator`     varchar(64) not null,\n"+
		"    `reviser`     varchar(64) not null,\n"+
		"    `created_at`  timestamp   not null default current_timestamp,\n"+
		"    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "async_flow_task", "0"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "eip",
		Column:  "recycle_status",
		Type:    "varchar(32)",
		Default: util.StringDefault(""),
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "network_interface",
		Column:  "recycle_status",
		Type:    "varchar(32)",
		Default: util.StringDefault(""),
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "image",
		Column:  "os_type",
		Type:    "varchar(32)",
		Default: util.StringDefault(""),
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "recycle_record",
		Column:  "recycle_type",
		Type:    "varchar(64)",
		Default: util.StringDefault(""),
	}); err != nil {
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
