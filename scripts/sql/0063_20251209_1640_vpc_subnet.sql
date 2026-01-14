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
    SQLVER=0063,HCMVER=v1.8.8.9

    Notes:
    1. account 表新增索引 idx_vendor_type
    2. vpc 表新增 extension.enable_cvm 字段、新增索引 idx_vendor_account_id
    3. subnet 表新增 extension.enable_cvm 字段
*/

START TRANSACTION;

ALTER TABLE account ADD KEY idx_vendor_type (vendor, type);

ALTER TABLE vpc ADD KEY idx_vendor_account_id (vendor, account_id);

-- 为自研云 VPC 添加 extension.enable_cvm 字段（排除默认VPC）
UPDATE vpc SET extension = JSON_SET(extension, '$.enable_cvm', true)
WHERE vendor = 'tcloud-ziyan' AND JSON_EXTRACT(extension, '$.enable_cvm') IS NULL AND `name` != "Default-VPC";

-- 为自研云 Subnet 添加 extension.enable_cvm 字段
UPDATE subnet SET extension = JSON_SET(extension, '$.enable_cvm', true)
WHERE vendor = 'tcloud-ziyan' AND JSON_EXTRACT(extension, '$.enable_cvm') IS NULL;

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.8.9' as `hcm_ver`, '0063' as `sql_ver`;

COMMIT;
