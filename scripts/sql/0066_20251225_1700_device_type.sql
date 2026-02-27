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
    SQLVER=0066,HCMVER=v1.8.9.1

    Notes:
    1. 创建 device_type 表
*/

START TRANSACTION;

-- 创建新的 device_type 表
CREATE TABLE IF NOT EXISTS `device_type`
(
    `id`                VARCHAR(64) NOT NULL COMMENT '唯一ID',
    `vendor`            VARCHAR(64)          DEFAULT NULL COMMENT '云厂商',
    `device_type`       VARCHAR(64) NOT NULL COMMENT '机型',
    `device_type_class` VARCHAR(64) NOT NULL COMMENT '通/专用机型，SpecialType专用，CommonType通用',
    `device_class`      VARCHAR(64) NOT NULL COMMENT '机型分类',
    `device_family`     VARCHAR(64) NOT NULL COMMENT '机型族',
    `core_type`         VARCHAR(64) NOT NULL COMMENT '核心类型',
    `cpu_core`          BIGINT(1)   NOT NULL COMMENT 'CPU核心数',
    `memory`            BIGINT(1)   NOT NULL COMMENT '内存大小，单位GB',
    `technical_class`   VARCHAR(64) NOT NULL COMMENT '技术分类',
    `region`            VARCHAR(64) NOT NULL COMMENT '地域',
    `zone`              VARCHAR(64) NOT NULL COMMENT '可用区',
    `disable`           TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '是否不使用',
    `source`            VARCHAR(64) NOT NULL DEFAULT 'sync' COMMENT '机型来源：sync-同步，manually-手动添加',
    `creator`           VARCHAR(64) NOT NULL COMMENT '创建者',
    `reviser`           VARCHAR(64) NOT NULL COMMENT '更新者',
    `created_at`        TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`        TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_vendor_region_zone_device_type` (`vendor`, `region`, `zone`, `device_type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='机型表';


insert into id_generator(`resource`, `max_id`)
values ('device_type', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.9.1' as `hcm_ver`, '0066' as `sql_ver`;

COMMIT;
