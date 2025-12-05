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
    SQLVER=0059,HCMVER=v1.8.7.15

    Notes:
    1. 扩容 res_plan_sub_ticket 与 res_plan_ticket_status 的 message 字段，支持记录完整失败信息
*/

START TRANSACTION;

ALTER TABLE `res_plan_sub_ticket`
    MODIFY COLUMN `message` TEXT NOT NULL COMMENT '子单据信息';

ALTER TABLE `res_plan_ticket_status`
    MODIFY COLUMN `message` TEXT NOT NULL COMMENT '单据失败信息';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.7.15' as `hcm_ver`, '0059' as `sql_ver`;

COMMIT;

