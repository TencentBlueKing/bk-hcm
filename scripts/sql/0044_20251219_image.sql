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
    SQLVER=0044,HCMVER=v1.8.11
    Notes:
    1. 为image表添加region字段，用于存储镜像的地域信息
    3. 创建(vendor, region)联合索引，优化查询效率
*/

START TRANSACTION;

-- 添加region字段
ALTER TABLE `image`
    ADD COLUMN `region` VARCHAR(64) DEFAULT '' AFTER `cloud_id`;

-- 从extension字段中提取region值并填充到region字段
-- TCloud: extension.region
UPDATE `image`
SET `region` = JSON_UNQUOTE(JSON_EXTRACT(`extension`, '$.region'))
WHERE JSON_EXTRACT(`extension`, '$.region') IS NOT NULL
  AND JSON_EXTRACT(`extension`, '$.region') != '';

-- 创建联合索引
ALTER TABLE `image`
    ADD INDEX `idx_vendor_region` (`vendor`, `region`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.11' as `hcm_ver`, '0044' as `sql_ver`;

COMMIT;

