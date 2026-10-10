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

// Package migration is the migration bill.
package migration

import (
	"context"

	"hcm/migrate/register"
	"hcm/migrate/util"
	"hcm/pkg/dal/dao/orm"
)

// ID is the unique identifier of this migration file, in the form
// <date>-<time>-<desc>-<random>. It must never change once executed.
// Migrations that share an ID run only the earlier one in execution order.
const ID = "20240531-1648-BILL-F93C"

func init() {
	register.Main.Regist(ID, "v1.6.0", "20240531164800", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0023_20240531_1648_bill.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_summary_version", "create table if not exists `account_bill_summary_version`\n"+
		"(\n"+
		"    `id`                varchar(64)     not null,\n"+
		"    `first_account_id`  varchar(64)     not null,\n"+
		"    `second_account_id` varchar(64)     not null,\n"+
		"    `vendor`            varchar(16)     not null,\n"+
		"    `product_id`        bigint(1),\n"+
		"    `bk_biz_id`         bigint(1),\n"+
		"    `bill_year`         bigint(1)       not null,\n"+
		"    `bill_month`        tinyint(1)      not null,\n"+
		"    `version_id`        varchar(64)     not null,\n"+
		"    `currency`          varchar(64)     not null,\n"+
		"    `cost`              decimal(38, 10) not null,\n"+
		"    `rmb_cost`          decimal(38, 10) not null,\n"+
		"    `created_at`        timestamp       not null default current_timestamp,\n"+
		"    `updated_at`        timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_bill_date_version` (`first_account_id`, `second_account_id`, `bill_year`, `bill_month`,\n"+
		"                                        `version_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_summary_version", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_summary_daily", "create table if not exists `account_bill_summary_daily`\n"+
		"(\n"+
		"    `id`              varchar(64)     not null,\n"+
		"    `root_account_id` varchar(64)     not null,\n"+
		"    `main_account_id` varchar(64)     not null,\n"+
		"    `vendor`          varchar(16)     not null,\n"+
		"    `product_id`      bigint(1),\n"+
		"    `bk_biz_id`       bigint(1),\n"+
		"    `bill_year`       bigint(1)       not null,\n"+
		"    `bill_month`      tinyint(1)      not null,\n"+
		"    `bill_day`        tinyint(1)      not null,\n"+
		"    `version_id`      bigint(1)       not null,\n"+
		"    `currency`        varchar(64)     not null,\n"+
		"    `cost`            decimal(38, 10) not null,\n"+
		"    `count`           bigint(1)       not null,\n"+
		"    `creator`         varchar(64)     not null,\n"+
		"    `reviser`         varchar(64)     not null,\n"+
		"    `created_at`      timestamp       not null default current_timestamp,\n"+
		"    `updated_at`      timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_bill_date_version`\n"+
		"        (`root_account_id`, `main_account_id`, `bill_year`, `bill_month`, `bill_day`, `version_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_summary_daily", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_item", "create table if not exists `account_bill_item`\n"+
		"(\n"+
		"    `id`              varchar(64)     not null,\n"+
		"    `root_account_id` varchar(64)     not null,\n"+
		"    `main_account_id` varchar(64)     not null,\n"+
		"    `vendor`          varchar(16)     not null,\n"+
		"    `product_id`      bigint(1),\n"+
		"    `bk_biz_id`       bigint(1),\n"+
		"    `bill_year`       bigint(1)       not null,\n"+
		"    `bill_month`      tinyint(1)      not null,\n"+
		"    `bill_day`        tinyint(1)      not null,\n"+
		"    `version_id`      bigint(1)       not null,\n"+
		"    `currency`        varchar(64)     not null,\n"+
		"    `cost`            decimal(38, 10) not null,\n"+
		"    `hc_product_code` varchar(128),\n"+
		"    `hc_product_name` varchar(128),\n"+
		"    `res_amount`      decimal(38, 10),\n"+
		"    `res_amount_unit` varchar(64),\n"+
		"    `extension`       json,\n"+
		"    `creator`         varchar(64)     not null,\n"+
		"    `reviser`         varchar(64)     not null,\n"+
		"    `created_at`      timestamp       not null default current_timestamp,\n"+
		"    `updated_at`      timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    index `idx_bill_item`\n"+
		"        (`vendor`, `bill_year`, `bill_month`, `root_account_id`, `product_id`, `main_account_id`, `bill_day`,\n"+
		"         `bk_biz_id`, `version_id`),\n"+
		"    index `idx_bill_item_root_account_id`\n"+
		"        (`vendor`, `bill_year`, `bill_month`, `root_account_id`),\n"+
		"    index `idx_bill_item_created_at`\n"+
		"        (`vendor`, `bill_year`, `bill_month`, `created_at`),\n"+
		"    index `idx_bill_item__main_account_id`\n"+
		"        (`vendor`, `bill_year`, `bill_month`, `main_account_id`),\n"+
		"    index `idx_bill_item_product_id`\n"+
		"        (`vendor`, `bill_year`, `bill_month`, `product_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_item", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_adjustment_item", "create table if not exists `account_bill_adjustment_item`\n"+
		"(\n"+
		"    `id`              varchar(64)     not null,\n"+
		"    `root_account_id` varchar(64)     not null,\n"+
		"    `main_account_id` varchar(64)     not null,\n"+
		"    `vendor`          varchar(16)     not null,\n"+
		"    `product_id`      bigint(1),\n"+
		"    `bk_biz_id`       bigint(1),\n"+
		"    `bill_year`       bigint(1)       not null,\n"+
		"    `bill_month`      tinyint(1)      not null,\n"+
		"    `bill_day`        tinyint(1)      not null,\n"+
		"    `type`            varchar(64)     not null,\n"+
		"    `memo`            varchar(255)             default '',\n"+
		"    `operator`        varchar(64)     not null,\n"+
		"    `currency`        varchar(64)     not null,\n"+
		"    `cost`            decimal(38, 10) not null,\n"+
		"    `rmb_cost`        decimal(38, 10) not null,\n"+
		"    `state`           varchar(64)     not null,\n"+
		"    `creator`         varchar(64)     not null,\n"+
		"    `created_at`      timestamp       not null default current_timestamp,\n"+
		"    `updated_at`      timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_adjustment_item", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_daily_pull_task", "create table if not exists `account_bill_daily_pull_task`\n"+
		"(\n"+
		"    `id`                    varchar(64)     not null,\n"+
		"    `root_account_id`       varchar(64)     not null,\n"+
		"    `main_account_id`       varchar(64)     not null,\n"+
		"    `vendor`                varchar(16)     not null,\n"+
		"    `product_id`            bigint(1),\n"+
		"    `bk_biz_id`             bigint(1),\n"+
		"    `bill_year`             bigint(1)       not null,\n"+
		"    `bill_month`            tinyint(1)      not null,\n"+
		"    `bill_day`              tinyint(1)      not null,\n"+
		"    `version_id`            bigint(1)       not null,\n"+
		"    `state`                 varchar(64)     not null,\n"+
		"    `count`                 bigint(1)       not null,\n"+
		"    `currency`              varchar(64)     not null,\n"+
		"    `cost`                  decimal(38, 10) not null,\n"+
		"    `flow_id`               varchar(64)     not null,\n"+
		"    `split_flow_id`         varchar(64)     not null,\n"+
		"    `daily_summary_flow_id` varchar(64)     not null,\n"+
		"    `created_at`            timestamp       not null default current_timestamp,\n"+
		"    `updated_at`            timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_bill_date_version`\n"+
		"        (`root_account_id`, `main_account_id`, `bill_year`, `bill_month`, `bill_day`, `version_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_daily_pull_task", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_summary_root", "create table if not exists `account_bill_summary_root`\n"+
		"(\n"+
		"    `id`                            varchar(64)     not null,\n"+
		"    `root_account_id`               varchar(64)     not null,\n"+
		"    `root_account_name`             varchar(64)     not null,\n"+
		"    `vendor`                        varchar(16)     not null,\n"+
		"    `bill_year`                     bigint(1)       not null,\n"+
		"    `bill_month`                    tinyint(1)      not null,\n"+
		"    `last_synced_version`           bigint(1)       not null,\n"+
		"    `current_version`               bigint(1)       not null,\n"+
		"    `currency`                      varchar(64)     not null,\n"+
		"    `last_month_cost_synced`        decimal(38, 10) not null,\n"+
		"    `last_month_rmb_cost_synced`    decimal(38, 10) not null,\n"+
		"    `current_month_cost_synced`     decimal(38, 10) not null,\n"+
		"    `current_month_rmb_cost_synced` decimal(38, 10) not null,\n"+
		"    `month_on_month_value`          float,\n"+
		"    `current_month_cost`            decimal(38, 10) not null,\n"+
		"    `current_month_rmb_cost`        decimal(38, 10) not null,\n"+
		"    `rate`                          float,\n"+
		"    `adjustment_cost`               decimal(38, 10) not null,\n"+
		"    `adjustment_rmb_cost`           decimal(38, 10) not null,\n"+
		"    `state`                         varchar(64)     not null,\n"+
		"    `bk_biz_num`                    bigint(1)       not null,\n"+
		"    `product_num`                   bigint(1)       not null,\n"+
		"    `created_at`                    timestamp       not null default current_timestamp,\n"+
		"    `updated_at`                    timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_bill_date`\n"+
		"        (`root_account_id`, `bill_year`, `bill_month`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_summary_root", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_summary_main", "create table if not exists `account_bill_summary_main`\n"+
		"(\n"+
		"    `id`                            varchar(64)     not null,\n"+
		"    `root_account_id`               varchar(64)     not null,\n"+
		"    `root_account_name`             varchar(64),\n"+
		"    `main_account_id`               varchar(64)     not null,\n"+
		"    `main_account_name`             varchar(64)     not null,\n"+
		"    `vendor`                        varchar(16)     not null,\n"+
		"    `product_id`                    bigint(1),\n"+
		"    `product_name`                  varchar(64),\n"+
		"    `bk_biz_id`                     bigint(1),\n"+
		"    `bk_biz_name`                   varchar(64),\n"+
		"    `bill_year`                     bigint(1)       not null,\n"+
		"    `bill_month`                    tinyint(1)      not null,\n"+
		"    `last_synced_version`           bigint(1)       not null,\n"+
		"    `current_version`               bigint(1)       not null,\n"+
		"    `currency`                      varchar(64)     not null,\n"+
		"    `last_month_cost_synced`        decimal(38, 10) not null,\n"+
		"    `last_month_rmb_cost_synced`    decimal(38, 10) not null,\n"+
		"    `current_month_cost_synced`     decimal(38, 10) not null,\n"+
		"    `current_month_rmb_cost_synced` decimal(38, 10) not null,\n"+
		"    `month_on_month_value`          float,\n"+
		"    `current_month_cost`            decimal(38, 10) not null,\n"+
		"    `current_month_rmb_cost`        decimal(38, 10) not null,\n"+
		"    `rate`                          float,\n"+
		"    `adjustment_cost`               decimal(38, 10) not null,\n"+
		"    `adjustment_rmb_cost`           decimal(38, 10) not null,\n"+
		"    `state`                         varchar(64)     not null,\n"+
		"    `created_at`                    timestamp       not null default current_timestamp,\n"+
		"    `updated_at`                    timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_bill_date`\n"+
		"        (`root_account_id`, `main_account_id`, `bill_year`, `bill_month`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_summary_main", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "root_account_bill_config", "CREATE TABLE `root_account_bill_config`\n"+
		"(\n"+
		"    `id`                  varchar(64) not null,\n"+
		"    `vendor`              varchar(16) not null default '',\n"+
		"    `root_account_id`     varchar(64) not null,\n"+
		"    `cloud_database_name` varchar(64)          default '',\n"+
		"    `cloud_table_name`    varchar(64)          default '',\n"+
		"    `err_msg`             json                 default NULL,\n"+
		"    `extension`           json                 default NULL,\n"+
		"    `creator`             varchar(64)          default '',\n"+
		"    `reviser`             varchar(64)          default '',\n"+
		"    `created_at`          timestamp   not null default current_timestamp,\n"+
		"    `updated_at`          timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_vendor_account_id` (`vendor`, `root_account_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "root_account_bill_config", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_exchange_rate", "create table account_bill_exchange_rate\n"+
		"(\n"+
		"    `id`            varchar(64)     not null,\n"+
		"    `year`          int             not null,\n"+
		"    `month`         int             not null,\n"+
		"    `from_currency` varchar(32)     not null comment '转换前货币',\n"+
		"    `to_currency`   varchar(32)     not null comment '转换后货币',\n"+
		"    `exchange_rate` decimal(38, 10) not null comment '转换汇率，单位转换前货币对应的转换后货币数量',\n"+
		"    `creator`       varchar(64)     not null,\n"+
		"    `reviser`       varchar(64)     not null,\n"+
		"    `created_at`    timestamp       not null default current_timestamp,\n"+
		"    `updated_at`    timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_year_month_from_currency_to_currency` (`year`, `month`, `from_currency`, `to_currency`)\n"+
		") comment '转换汇率'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_exchange_rate", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_sync_record", "create table `account_bill_sync_record`\n"+
		"(\n"+
		"    `id`         varchar(64)     not null,\n"+
		"    `vendor`     varchar(16)     not null default '',\n"+
		"    `bill_year`  int             not null,\n"+
		"    `bill_month` int             not null,\n"+
		"    `state`      varchar(64)     not null,\n"+
		"    `currency`   varchar(32)     not null comment '货币',\n"+
		"    `cost`       decimal(38, 10) not null,\n"+
		"    `rmb_cost`   decimal(38, 10) not null,\n"+
		"    `detail`     text            not null,\n"+
		"    `operator`   varchar(64)     not null,\n"+
		"    `creator`    varchar(64)     not null,\n"+
		"    `reviser`    varchar(64)     not null,\n"+
		"    `created_at` timestamp       not null default current_timestamp,\n"+
		"    `updated_at` timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") comment '同步记录'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_sync_record", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_month_task", "create table `account_bill_month_task`\n"+
		"(\n"+
		"    `id`              varchar(64)     not null,\n"+
		"    `root_account_id` varchar(64)     not null,\n"+
		"    `vendor`          varchar(16)     not null default '',\n"+
		"    `bill_year`       int             not null,\n"+
		"    `bill_month`      int             not null,\n"+
		"    `version_id`      bigint(1)       not null,\n"+
		"    `state`           varchar(64)     not null,\n"+
		"    `currency`        varchar(32)     not null comment '货币',\n"+
		"    `count`           int             not null default 0,\n"+
		"    `cost`            decimal(38, 10) not null,\n"+
		"    `pull_index`      bigint(1)       not null,\n"+
		"    `pull_flow_id`    varchar(64)     not null,\n"+
		"    `split_index`     bigint(1)       not null,\n"+
		"    `split_flow_id`   varchar(64)     not null,\n"+
		"    `summary_flow_id` varchar(64)     not null,\n"+
		"    `summary_detail`  text            not null,\n"+
		"    `creator`         varchar(64)     not null,\n"+
		"    `reviser`         varchar(64)     not null,\n"+
		"    `created_at`      timestamp       not null default current_timestamp,\n"+
		"    `updated_at`      timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_root_account_id_year_month` (`root_account_id`, `bill_year`, `bill_month`)\n"+
		") comment '月度任务'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_bill_month_task", "0"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "async_flow_task", "idx_state_updated_at", []string{"state", "updated_at"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "async_flow_task", "idx_flow_id", []string{"flow_id"}, false); err != nil {
		return err
	}
	return nil
}
