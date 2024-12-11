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
    SQLVER=9999MVER=v9.9.9

    Notes:
    1. 修改cvm表唯一键索引为cloud_id, vendor, region
    2. 修改security_group表唯一键索引为cloud_id, vendor, region
    3. 修改vpc表唯一键索引为cloud_id, vendor, region
    4. 修改subnet表唯一键索引为cloud_id, vendor, region
    5. 修改tcloud_security_group_rule表唯一键索引为cloud_security_group_id, cloud_policy_index, type, region
    6. disk表添加唯一键索引: cloud_id, vendor, region
*/

START TRANSACTION;

-- 1. 修改cvm表唯一键索引为cloud_id, vendor, region
alter table cvm
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table cvm
    drop key idx_uk_cloud_id_vendor;

-- 2. 修改security_group表唯一键索引为cloud_id, vendor, region
alter table security_group
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table security_group
drop key idx_uk_cloud_id_vendor;

-- 3. 修改vpc表唯一键索引为cloud_id, vendor, region
alter table vpc
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table vpc
drop key idx_uk_cloud_id_vendor;

-- 4. 修改subnet表唯一键索引为cloud_id, vendor, region
alter table subnet
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table subnet
drop key idx_uk_cloud_id_vendor;


-- 5. 修改tcloud_security_group_rule表唯一键索引为cloud_security_group_id, cloud_policy_index, type, region
alter table tcloud_security_group_rule
    add constraint idx_uk_cloud_security_group_id_cloud_policy_index_type_region unique (cloud_security_group_id, cloud_policy_index, type, region);
alter table tcloud_security_group_rule
drop key idx_uk_cloud_security_group_id_cloud_policy_index_type;

-- 6. disk表添加唯一键索引: cloud_id, vendor, region
alter table disk
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;


COMMIT
