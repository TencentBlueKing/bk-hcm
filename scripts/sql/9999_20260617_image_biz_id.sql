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
   SQLVER=9999,HCMVER=v9.9.9

   Notes:
   1. 为image表添加bk_biz_id字段，用于私有镜像关联业务
   2. 将bk_biz_id加入现有idx_vendor_region联合索引，优化厂商+地域+业务维度查询效率
*/

START TRANSACTION;

-- 添加bk_biz_id字段，用于镜像关联业务（-1表示未分配/公共镜像，>0表示绑定到具体业务的私有镜像）
ALTER TABLE `image`
    ADD COLUMN `bk_biz_id` bigint DEFAULT -1 COMMENT '业务ID，-1表示未分配，>0表示绑定的业务' AFTER `os_type`;

-- 删除原有的idx_vendor_region索引，重建为包含bk_biz_id的联合索引
ALTER TABLE `image`
    DROP INDEX `idx_vendor_region`,
    ADD INDEX `idx_vendor_region_biz` (`vendor`, `region`, `bk_biz_id`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
