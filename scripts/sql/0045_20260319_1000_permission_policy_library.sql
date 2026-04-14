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
    SQLVER=0045,HCMVER=v1.8.11

    Notes:
    1. 添加权限策略库表 permission_policy_library
*/

START TRANSACTION;

-- 1. 权限策略库表
CREATE TABLE IF NOT EXISTS `permission_policy_library` (
    `id`              varchar(64)   NOT NULL                              COMMENT '策略库ID',
    `name`            varchar(128)  NOT NULL                              COMMENT '策略库名称',
    `policy_document` longtext      NOT NULL                              COMMENT '当前版本的权限策略JSON内容',
    `policy_hash`     varchar(64)   NOT NULL                              COMMENT '策略内容SHA256哈希值',
    `version`         int           NOT NULL DEFAULT 1                    COMMENT '当前版本号，从1开始递增',
    `bk_biz_ids`      json          NOT NULL                              COMMENT '允许使用的业务ID列表',
    `memo`            varchar(255)  DEFAULT NULL                          COMMENT '策略库描述',
    `vendor`          varchar(16)   NOT NULL                              COMMENT '云厂商',
    `tenant_id`       varchar(64)   NOT NULL DEFAULT 'default'            COMMENT '租户ID',
    `creator`         varchar(64)   NOT NULL                              COMMENT '创建者',
    `reviser`         varchar(64)   NOT NULL                              COMMENT '更新者',
    `created_at`      timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP    COMMENT '创建时间',
    `updated_at`      timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_bin COMMENT='权限策略库表';


INSERT INTO id_generator(`resource`, `max_id`) VALUES ('permission_policy_library', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.11' as `hcm_ver`, '0045' as `sql_ver`;

COMMIT;
