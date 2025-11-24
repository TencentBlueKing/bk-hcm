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
    SQLVER=9999,HCMVER=v9.9.9.9

    Notes:
    1. 添加CVM申请单统计配置表 cvm_apply_order_statistics_config
*/

START TRANSACTION;

--  1. CVM申请单统计配置表
create table if not exists `cvm_apply_order_statistics_config`
(
    `id`                      varchar(64)  not null comment '主键',
    `stat_month`              varchar(16)  not null comment '年月，格式：YYYY-MM',
    `bk_biz_id`               bigint       not null comment '业务ID',
    `sub_order_ids`            text         not null default '' comment '子订单号，多个用逗号分隔',
    `start_at`                varchar(64)           default '' comment '开始时间',
    `end_at`                  varchar(64)           default '' comment '结束时间',
    `memo`                    varchar(255)          default '' comment '备注',
    `extension`               json         not null comment '扩展字段',
    `creator`                 varchar(64)  not null comment '创建者',
    `reviser`                 varchar(64)  not null comment '更新者',
    `created_at`              timestamp    not null default current_timestamp comment '创建时间',
    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp comment '更新时间',
    primary key (`id`),
    key `idx_stat_month` (`stat_month`)
    ) engine = innodb default charset = utf8mb4 comment='CVM申请单统计配置表';

insert into id_generator(`resource`, `max_id`)
values ('cvm_apply_order_statistics_config', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;