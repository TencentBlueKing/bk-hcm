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
    SQLVER=0049,HCMVER=v1.9.0

    Notes:
    1. 账号表新增字段：email（邮箱）、security_managers（安全负责人）、cloud_created_at（云上创建时间）
*/

START TRANSACTION;

ALTER TABLE `account`
    ADD COLUMN `email` varchar(64) DEFAULT NULL COMMENT '邮箱',
    ADD COLUMN `security_managers` json DEFAULT NULL COMMENT '安全负责人',
    ADD COLUMN `cloud_created_at` varchar(64) NULL COMMENT '云上创建时间';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.9.0' as `hcm_ver`, '0049' as `sql_ver`;

COMMIT;
