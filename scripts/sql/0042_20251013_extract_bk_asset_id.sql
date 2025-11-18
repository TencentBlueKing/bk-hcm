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
 * We undertake not to change the open source license (MIT License) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

/*
    SQLVER=0042,HCMVER=v1.8.8
    Notes:
    1. 添加bk_asset_id字段到cvm表
*/

START TRANSACTION;

-- 首先添加bk_asset_id字段到cvm表
ALTER TABLE cvm ADD COLUMN bk_asset_id VARCHAR(64) DEFAULT '' COMMENT '固资号';

-- 更新版本信息
CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.8' as `hcm_ver`, '0042' as `sql_ver`;

COMMIT;