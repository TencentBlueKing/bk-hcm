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
    1. 新增 IaC 插件注册表 iac_plugin（每个插件的每个版本一条记录，平台级，不带租户）
*/

START TRANSACTION;

-- 1. IaC 插件注册表
CREATE TABLE IF NOT EXISTS `iac_plugin`
(
    `id`             varchar(64)  NOT NULL COMMENT 'HCM主键ID',
    `plugin_name`    varchar(64)  NOT NULL COMMENT '插件名',
    `plugin_version` varchar(64)  NOT NULL COMMENT '插件版本',
    `protocol`       varchar(32)  NOT NULL COMMENT '插件协议：tf_provider / http',
    `state`          varchar(32)  NOT NULL COMMENT '状态：init / enabled / disabled / failed',
    `message`        varchar(255) NOT NULL DEFAULT '' COMMENT '系统写入的状态说明，failed 时写明原因',
    `extension`      json         NOT NULL COMMENT '协议相关信息，结构随 protocol 不同',
    `memo`           varchar(255)          DEFAULT '' COMMENT '备注',
    `creator`        varchar(64)  NOT NULL COMMENT '创建者',
    `reviser`        varchar(64)  NOT NULL COMMENT '更新者',
    `created_at`     timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uk_plugin_name_version` (`plugin_name`, `plugin_version`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_bin COMMENT ='IaC插件注册表';

INSERT INTO id_generator(`resource`, `max_id`)
VALUES ('iac_plugin', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
