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
    1. 新增插件 Schema 权威存储表 iac_plugin_schema
*/

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `iac_plugin_schema`
(
    `id`             varchar(64) NOT NULL COMMENT 'HCM主键ID',
    `plugin_name`    varchar(64) NOT NULL COMMENT '插件名，对应iac_plugin.plugin_name',
    `plugin_version` varchar(64) NOT NULL COMMENT '插件版本，如1.81.5',
    `schema_hash`    varchar(80) NOT NULL COMMENT 'sha256:+64位hex，随完整Schema返回',
    `schema_detail`  json        NOT NULL COMMENT '完整Schema JSON（对应内联SchemaDetail：含protocol_version、capabilities、provider、resource_schemas、data_source_schemas等）',
    `creator`        varchar(64) NOT NULL COMMENT '创建者',
    `reviser`        varchar(64) NOT NULL COMMENT '更新者',
    `created_at`     timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uk_plugin_name_plugin_version` (`plugin_name`, `plugin_version`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_bin COMMENT ='插件Schema权威存储表';

INSERT INTO id_generator(`resource`, `max_id`)
VALUES ('iac_plugin_schema', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
