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
    1. 添加全局配置表 global_config
*/

START TRANSACTION;

--  1. 全局配置表
create table if not exists `global_config`
(
    `id`           varchar(64) not null comment '主键',
    `config_key`   varchar(64) not null comment 'key',
    `config_value` json        not null comment 'value',
    `config_type`  varchar(64) not null comment '类型',
    `memo`         varchar(255)         default '' comment '备注',
    `creator`      varchar(64) not null comment '创建者',
    `reviser`      varchar(64) not null comment '更新者',
    `created_at`   timestamp   not null default current_timestamp comment '创建时间',
    `updated_at`   timestamp   not null default current_timestamp on update current_timestamp comment '更新时间',
    primary key (`id`),
    unique key `idx_global_config_id` (`config_type`, `config_key`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='全局配置表';

insert into id_generator(`resource`, `max_id`)
values ('global_config', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
