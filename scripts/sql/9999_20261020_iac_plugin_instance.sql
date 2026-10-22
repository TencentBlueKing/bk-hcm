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
    1. 新增 IaC 插件实例表 iac_plugin_instance
*/

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `iac_plugin_instance`
(
    `id`           varchar(64)  NOT NULL COMMENT '主键',
    `plugin_name`  varchar(64)  NOT NULL COMMENT '所属插件名，对应iac_plugin.plugin_name',
    `name`         varchar(128) NOT NULL COMMENT '实例名称',
    `attrs`        json         NOT NULL COMMENT '非敏感属性（JSON对象），对应provider.block中非敏感部分',
    `secret_attrs` json         NOT NULL COMMENT '敏感属性（JSON对象，叶子值密文），对应provider.block中敏感部分',
    `extension`    json         NOT NULL COMMENT '业务扩展字段（JSON对象），如关联HCM云账号{"account_id":"xxx"}',
    `memo`         varchar(255)          DEFAULT '' COMMENT '备注',
    `tenant_id`    varchar(64)  NOT NULL DEFAULT 'default' COMMENT '租户ID',
    `creator`      varchar(64)  NOT NULL COMMENT '创建者',
    `reviser`      varchar(64)  NOT NULL COMMENT '更新者',
    `created_at`   timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_plugin_name` (`plugin_name`),
    KEY `idx_tenant_plugin` (`tenant_id`, `plugin_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_bin COMMENT ='IaC 插件实例表';

INSERT INTO id_generator(`resource`, `max_id`)
VALUES ('iac_plugin_instance', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
