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
    SQLVER=0067,HCMVER=v1.8.9.1

    Notes:
    1. 新增 tcloud_ziyan_pm_device_type 表
*/

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `tcloud_ziyan_pm_device_type`
(
    `id`          VARCHAR(64) NOT NULL COMMENT '唯一标识',
    `device_type` VARCHAR(64) NOT NULL COMMENT '机型',
    `raid`        VARCHAR(64) NOT NULL COMMENT 'RAID 类型',
    `cpu_core`    INT         NOT NULL COMMENT '核心数',
    `memory`      INT         NOT NULL COMMENT '内存',
    `disable`     TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '是否不使用',
    `creator`     VARCHAR(64) NOT NULL COMMENT '创建者',
    `reviser`     VARCHAR(64) NOT NULL COMMENT '更新者',
    `created_at`  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_uk_device_type` (`device_type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='腾讯自研云物理机机型表';

insert into id_generator(`resource`, `max_id`)
values ('tcloud_ziyan_pm_device_type', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.9.1' as `hcm_ver`, '0067' as `sql_ver`;

COMMIT;
