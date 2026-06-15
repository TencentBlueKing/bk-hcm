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
    SQLVER=0047,HCMVER=v1.9.0

    Notes:
    1. 添加账号密钥表 account_secret
*/

START TRANSACTION;

-- 1. 账号密钥表
CREATE TABLE IF NOT EXISTS `account_secret` (
    `id` varchar(64) NOT NULL COMMENT '密钥ID',
    `account_id` varchar(64) NOT NULL COMMENT '账号ID',
    `vendor` varchar(16) NOT NULL COMMENT '云厂商',
    `type` varchar(16) NOT NULL COMMENT '密钥类型',
    `status` varchar(16) NOT NULL COMMENT '密钥状态',
    `extension` json NOT NULL COMMENT '云厂商差异扩展字段',
    `tenant_id` varchar(64) NOT NULL COMMENT '租户ID' default 'default',
    `creator` varchar(64) NOT NULL COMMENT '创建者',
    `reviser` varchar(64) NOT NULL COMMENT '更新者',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_bin COMMENT='账号密钥表';

CREATE INDEX idx_account_id ON account_secret(account_id);

INSERT INTO id_generator(`resource`, `max_id`)
VALUES ('account_secret', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.9.0' as `hcm_ver`, '0047' as `sql_ver`;

COMMIT;
