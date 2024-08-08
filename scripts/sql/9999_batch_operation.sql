/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
*/

START TRANSACTION;

create table batch_operation (
    `id` varchar(64) not null,
    `bk_biz_id` bigint not null default -1,
    `audit_id` bigint(1) unsigned not null,
    `detail` json not null,
    `creator` varchar(64) not null,
    `created_at` timestamp not null default current_timestamp,
    `updated_at` timestamp not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb default charset = utf8mb4 collate = utf8mb4_bin comment = '批量操作表';

create table `batch_operation_async_flow_rel` (
     `id` bigint unsigned not null auto_increment,
     `batch_operation_id` varchar(64) not null,
     `audit_id` bigint(1) unsigned not null,
     `flow_id` varchar(64) not null,
     `creator` varchar(64) not null,
     `created_at` timestamp not null default current_timestamp,
     primary key (`id`),
     unique key `idx_uk_batch_operation_id_audit_id_flow_id` (`batch_operation_id`, `audit_id`, `flow_id`)
) engine = innodb default charset = utf8mb4 collate = utf8mb4_bin comment = '批量操作与任务流关系表';

insert into id_generator(`resource`, `max_id`)
values ('batch_operation', '0'),
       ('batch_operation_async_flow_rel', '0');


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT