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
    1. 支持云账号管理
*/

START TRANSACTION;

-- 1. 二级账号表
create table if not exists `main_account`
(
    `id`                    varchar(64)     not null,
    `vendor`                varchar(16)     not null,
    `cloud_id`              varchar(64)     not null,
    `email`                 varchar(255)    not null,
    `managers`              json            not null,
    `bak_managers`          json            not null,
    `site`                  varchar(32)     not null,
    `business_type`         varchar(64)     not null,
    `status`                varchar(32)     not null,
    `parent_account_name`   varchar(255)    not null,
    `parent_account_id`     varchar(64)     not null,
    `dept_id`               bigint(1)       not null,
    `bk_biz_id`             bigint(1)       not null,
    `op_product_id`         bigint(1)       not null,
    `memo`                  varchar(512)            default '',
    `extension`             json            not null,
    `creator`        varchar(64) not null,
    `reviser`        varchar(64) not null,
    `created_at`     timestamp   not null default current_timestamp,
    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,
    primary key(`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
value ('main_account','0');

-- 2. 一级账号表
create table if not exists `root_account`
(
    `id`                    varchar(64)     not null,
    `name`                  varchar(64)     not null,
    `vendor`                varchar(16)     not null,
    `cloud_id`              varchar(64)     not null,
    `email`                 varchar(255)    not null,
    `managers`              json            not null,
    `bak_managers`          json            not null,
    `site`                  varchar(32)     not null,
    `dept_id`               bigint(1)       not null,
    `memo`                  varchar(512)            default '',
    `extension`             json            not null,
    `creator`        varchar(64) not null,
    `reviser`        varchar(64) not null,
    `created_at`     timestamp   not null default current_timestamp,
    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,
    primary key(`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
value ('root_account','0');


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT