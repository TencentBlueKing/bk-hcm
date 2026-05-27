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
    SQLVER=0051,HCMVER=v1.9.0

    Notes:
    1. 添加权限模板表 permission_template
*/

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `permission_template` (
    `id`                       varchar(64)   NOT NULL                              COMMENT '本地模板ID',
    `cloud_id`                 varchar(64)   NOT NULL                              COMMENT '云上策略ID',
    `name`                     varchar(128)  NOT NULL                              COMMENT '模板名称',
    `account_id`               varchar(64)   NOT NULL                              COMMENT '所属二级账号ID',
    `policy_library_id`        varchar(64)   DEFAULT NULL                          COMMENT '来源权限策略库ID',
    `policy_library_version`   int           DEFAULT NULL                          COMMENT '权限策略库版本',
    `policy_library_sync_time` timestamp     NULL DEFAULT NULL                     COMMENT '权限策略库同步时间',
    `policy_document`          longtext      NOT NULL                              COMMENT '策略JSON内容',
    `policy_hash`              varchar(64)   NOT NULL                              COMMENT '策略内容哈希值',
    `memo`                     varchar(255)  DEFAULT NULL                          COMMENT '描述',
    `extension`                json          DEFAULT NULL                          COMMENT '云厂商差异扩展字段',
    `vendor`                   varchar(16)   NOT NULL                              COMMENT '云厂商',
    `tenant_id`                varchar(64)   NOT NULL DEFAULT 'default'            COMMENT '租户ID',
    `creator`                  varchar(64)   NOT NULL                              COMMENT '创建者',
    `reviser`                  varchar(64)   NOT NULL                              COMMENT '更新者',
    `created_at`               timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP    COMMENT '创建时间',
    `updated_at`               timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_bin COMMENT='权限模板表';

INSERT INTO id_generator(`resource`, `max_id`) VALUES ('permission_template', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.9.0' as `hcm_ver`, '0051' as `sql_ver`;

COMMIT;
