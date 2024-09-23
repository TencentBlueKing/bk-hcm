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
    1. 添加任务管理表task_management
    2. 添加任务详情表task_detail
*/

START TRANSACTION;

--  1. 任务管理表
create table if not exists `task_management`
(
    `id`                  varchar(64)  not null,
    `bk_biz_id`           bigint       not null,
    `source`              varchar(16)  not null,
    `vendor`              varchar(16)  not null,
    `state`               varchar(16)  not null,
    `account_id`          varchar(64)  not null,
    `resource`            varchar(16)  not null,
    `operations`          json         not null,
    `flow_ids`            json         default NULL,
    `extension`           json         default NULL,
    `creator`             varchar(64)  not null,
    `reviser`             varchar(64)  not null,
    `created_at`          timestamp    not null default current_timestamp,
    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    index `idx_state` (`state`)
    ) engine = innodb
    default charset = utf8mb4
    collate utf8mb4_bin comment ='任务管理表';

--  2. 任务详情表
create table if not exists `task_detail`
(
    `id`                  varchar(64)  not null,
    `bk_biz_id`           bigint       not null,
    `task_management_id`  varchar(64)  not null,
    `flow_id`             varchar(64)  default '',
    `task_action_ids`     json         default NULL,
    `operation`           varchar(64)  not null,
    `param`               json         not null,
    `state`               varchar(16)  not null,
    `reason`              varchar(255) default '',
    `extension`           json         default NULL,
    `creator`             varchar(64)  not null,
    `reviser`             varchar(64)  not null,
    `result`              json         default NULL,
    `created_at`          timestamp    not null default current_timestamp,
    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    index `idx_task_management_id` (`task_management_id`)
    ) engine = innodb
    default charset = utf8mb4
    collate utf8mb4_bin comment ='任务详情表';

insert into id_generator(`resource`, `max_id`)
values ('task_management', '0'),
       ('task_detail', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT
