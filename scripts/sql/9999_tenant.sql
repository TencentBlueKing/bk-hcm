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
    1. 添加租户表 tenant
*/

START TRANSACTION;

--  1. 租户表
create table if not exists `tenant` (
    `id` varchar(64) NOT NULL COMMENT '唯一ID',
    `tenant_id` varchar(64) NOT NULL COMMENT '租户ID',
    `status` varchar(64) NOT NULL DEFAULT 'enable' COMMENT '租户状态(enable:启用 disable:禁用)',
    `creator` varchar(64) NOT NULL COMMENT '创建人',
    `reviser` varchar(64) NOT NULL COMMENT '修改人',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '该记录创建的时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uk_tenant_id` (`tenant_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_0900_ai_ci COMMENT='租户表';

insert into id_generator(`resource`, `max_id`)
values ('tenant', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
