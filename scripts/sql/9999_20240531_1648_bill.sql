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

/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 支持云账单
*/

START TRANSACTION;

-- 账单同步器表
create table if not exists `account_bill_puller` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `vendor` varchar(16) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `pull_mode` varchar(64) not null,
  `sync_period` varchar(64) not null,
  `bill_delay` varchar(64) not null,
  `final_bill_calendar_date` bigint(1) not null,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`),
  unique key `idx_account_id` (`first_account_id`, `second_account_id`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_puller', '0');

-- 月账单汇总表
create table if not exists `account_bill_summary` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `vendor` varchar(16) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `bill_year` bigint(1) not null,
  `bill_month` tinyint(1) not null,
  `current_version` varchar(64) not null,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`),
  unique key `idx_bill_date` (`first_account_id`, `second_account_id`, `bill_year`, `bill_month`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_summary', '0');

-- 月账单汇总版本表
create table if not exists `account_bill_summary_version` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `vendor` varchar(16) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `bill_year` bigint(1) not null,
  `bill_month` tinyint(1) not null,
  `version_id` varchar(64) not null,
  `currency` varchar(64) not null,
  `cost` decimal(28,8) not null,
  `rmb_cost` decimal(28,8) not null,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`),
  unique key `idx_bill_date_version` (`first_account_id`, `second_account_id`, `bill_year`, `bill_month`, `version_id`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_summary_version', '0');

-- 每日账单汇总表
create table if not exists `account_bill_summary_daily` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `vendor` varchar(16) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `bill_year` bigint(1) not null,
  `bill_month` tinyint(1) not null,
  `bill_day` tinyint(1) not null,
  `version_id` varchar(64) not null,
  `currency` varchar(64) not null,
  `cost` decimal(28,8) not null,
  `rmb_cost` decimal(28,8) not null,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`),
  unique key `idx_bill_date_version` (`first_account_id`, `second_account_id`, `bill_year`, `bill_month`, `bill_day`, `version_id`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_summary_daily', '0');

-- 分账后的明细数据
create table if not exists `account_bill_item` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `vendor` varchar(16) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `bill_year` bigint(1) not null,
  `bill_month` tinyint(1) not null,
  `bill_day` tinyint(1) not null,
  `version_id` varchar(64) not null,
  `currency` varchar(64) not null,
  `cost` decimal(28,8) not null,
  `rmb_cost` decimal(28,8) not null,
  `hc_product_code` varchar(128),
  `hc_product_name` varchar(128),
  `res_amount` decimal(28,8),
  `res_amount_unit` varchar(64),
  `extension` json,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_item', '0');

-- 调账明细数据
create table if not exists `account_bill_adjustment_item` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `vendor` varchar(16) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `bill_year` bigint(1) not null,
  `bill_month` tinyint(1) not null,
  `bill_day` tinyint(1) not null,
  `type` varchar(64) not null,
  `memo` varchar(255) default '',
  `operator` varchar(64) not null,
  `currency` varchar(64) not null,
  `cost` decimal(28,8) not null,
  `rmb_cost` decimal(28,8) not null,
  `state` varchar(64) not null,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_adjustment_item', '0');

-- 每日拉取任务状态表
create table if not exists `account_bill_daily_pull_task` (
  `id` varchar(64) not null,
  `first_account_id` varchar(64) not null,
  `second_account_id` varchar(64) not null,
  `product_id` bigint(1),
  `bk_biz_id` bigint(1),
  `bill_year` bigint(1) not null,
  `bill_month` tinyint(1) not null,
  `bill_day` tinyint(1) not null,
  `version_id` varchar(64) not null,
  `state` varchar(64) not null,
  `message` varchar(255) default '',
  `count` bigint(1) not null,
  `currency` varchar(64) not null,
  `cost` decimal(28,8) not null,
  `created_at` timestamp not null default current_timestamp,
  `updated_at` timestamp not null default current_timestamp on update current_timestamp,
  primary key (`id`),
  unique key `idx_bill_date_version` (`first_account_id`, `second_account_id`, `bill_year`, `bill_month`, `bill_day`, `version_id`)
) engine = innodb default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account_bill_daily_pull_task', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT