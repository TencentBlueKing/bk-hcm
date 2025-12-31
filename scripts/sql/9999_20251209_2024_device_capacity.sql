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
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 新增 device_capacity 表
*/

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `device_capacity`
(
    `id`           VARCHAR(64) NOT NULL COMMENT '唯一标识',
    `require_type` INT         NOT NULL COMMENT '需求类型',
    `region`       VARCHAR(64) NOT NULL COMMENT '地域',
    `zone`         VARCHAR(64) NOT NULL COMMENT '可用区',
    `device_type`  VARCHAR(64) NOT NULL COMMENT '机型',
    `capacity`     BIGINT      NOT NULL COMMENT '库存',
    `extension`    JSON        NOT NULL COMMENT '扩展字段',
    `creator`      VARCHAR(64) NOT NULL COMMENT '创建者',
    `reviser`      VARCHAR(64) NOT NULL COMMENT '更新者',
    `created_at`   DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_uk_require_type_region_zone_device_type` (`require_type`, `region`, `zone`, `device_type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='机型库存表';

insert into id_generator(`resource`, `max_id`)
values ('device_capacity', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
