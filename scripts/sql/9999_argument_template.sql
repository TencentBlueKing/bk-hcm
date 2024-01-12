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
    SQLVER=9999,HCMVER=9999

    Notes:
    1. 新增参数模版的数据表
*/

START TRANSACTION;

create table if not exists `argument_template`
(
    `id`                  varchar(64)  not null,
    `cloud_id`            varchar(255) not null,
    `name`                varchar(255) not null,
    `vendor`              varchar(16)  not null,
    `bk_biz_id`           bigint       not null default '-1',
    `account_id`          varchar(64)  not null,
    `type`                varchar(16)  not null,
    `templates`           json         not null,
    `group_templates`     json         not null,
    `memo`                varchar(255)          default '',
    `creator`             varchar(64)  not null,
    `reviser`             varchar(64)  not null,
    `created_at`          timestamp    not null default current_timestamp,
    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_cloud_id` (`bk_biz_id`, `cloud_id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='参数模版表';

insert into id_generator(`resource`, `max_id`)
values ('argument_template', '0');

alter table tcloud_security_group_rule
    add column `service_id` varchar(64) default null after `port`,
    add column `service_group_id` varchar(64) default null after `cloud_service_id`,
    add column `address_id` varchar(64) default null after `cloud_target_security_group_id`,
    add column `address_group_id` varchar(64) default null after `cloud_address_id`;

COMMIT