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
    SQLVER=0027,HCMVER=v1.6.11

    Notes:
    1. 修改`security_group`表，增加`type`字段
*/

START TRANSACTION;

ALTER TABLE security_group
    ADD COLUMN `cloud_update_time`  VARCHAR(64) NOT NULL DEFAULT '' COMMENT '云上修改时间' AFTER `association_template_id`,
    ADD COLUMN `cloud_created_time` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '云上创建时间' AFTER `association_template_id`,
    ADD COLUMN `tags`               json        NOT NULL COMMENT '标签' AFTER `extension`;

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.6.11' as `hcm_ver`, '0027' as `sql_ver`;

COMMIT;

