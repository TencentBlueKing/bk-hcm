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
    1. 添加组织拓扑表 org_topo
*/

START TRANSACTION;

--  1. 组织拓扑表
create table if not exists `org_topo`
(
    `id`           varchar(16)  not null,
    `dept_id`      varchar(64)  not null comment '组织部门ID',
    `dept_name`    varchar(64)  not null comment '组织部门名称',
    `full_name`    varchar(256) not null comment '部门的完整名称',
    `level`        int          not null comment '部门所属等级',
    `parent`       varchar(64)  not null comment '部门的上级部门ID',
    `has_children` tinyint(1)   not null default '0' comment '是否有下级部门(0:否1:有)',
    `memo`         varchar(256)          default '' comment '备注',
    `creator`      varchar(64)  not null comment '创建人',
    `reviser`      varchar(64)  not null comment '更新人',
    `created_at`   timestamp    not null default current_timestamp comment '该记录创建的时间',
    `updated_at`   timestamp    not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uk_dept_id` (`dept_id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='组织拓扑表';

insert into id_generator(`resource`, `max_id`)
values ('org_topo', '0');

CREATE OR REPLACE DEFINER='ADMIN'@'localhost' VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
