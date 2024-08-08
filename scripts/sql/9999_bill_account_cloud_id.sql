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
    1. 修改`account_bill_daily_pull_task`表，增加`root_account_cloud_id`及`main_account_cloud_id`字段
    2. 修改`account_bill_summary_daily`表，增加`root_account_cloud_id`及`main_account_cloud_id`字段
    3. 修改`account_bill_summary_main`表，删除 `root_account_name` ,增加`root_account_cloud_id`及`main_account_cloud_id`字段
    4. 修改`account_bill_summary_root`表，增加`root_account_cloud_id`字段
    5. 修改`account_bill_month_task`表，增加`root_account_cloud_id`字段
    6. 修改`account_bill_sync_record`表，增加`cost`字段
*/

START TRANSACTION;

ALTER TABLE `account_bill_daily_pull_task`
    ADD COLUMN `root_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'root account cloud id' AFTER `root_account_id`,
    ADD COLUMN `main_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'main account cloud id' AFTER `main_account_id`;

ALTER TABLE `account_bill_summary_daily`
    ADD COLUMN `root_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'root account cloud id' AFTER `root_account_id`,
    ADD COLUMN `main_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'main account cloud id' AFTER `main_account_id`;

ALTER TABLE `account_bill_summary_main`
    DROP COLUMN `root_account_name`,
    DROP COLUMN `main_account_name`,
    ADD COLUMN `root_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'root account cloud id' AFTER `root_account_id`,
    ADD COLUMN `main_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'main account cloud id' AFTER `main_account_id`;


ALTER TABLE `account_bill_summary_root`
    DROP COLUMN `root_account_name`,
    ADD COLUMN `root_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'root account cloud id' AFTER `root_account_id`;


ALTER TABLE `account_bill_month_task`
    ADD COLUMN `root_account_cloud_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'root account cloud id' AFTER `root_account_id`;


ALTER TABLE `account_bill_sync_record`
    ADD COLUMN `count` bigint NOT NULL DEFAULT 0 AFTER `currency`;


-- account_bill_daily_pull_task
UPDATE account_bill_daily_pull_task AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id
SET ab.root_account_cloud_id = ra.cloud_id
WHERE ab.root_account_cloud_id = '';

UPDATE account_bill_daily_pull_task AS ab JOIN main_account AS ma ON ab.main_account_id = ma.id
SET ab.main_account_cloud_id = ma.cloud_id
WHERE ab.main_account_cloud_id = '';

-- account_bill_summary_daily
UPDATE account_bill_summary_daily AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id
SET ab.root_account_cloud_id = ra.cloud_id
WHERE ab.root_account_cloud_id = '';
UPDATE account_bill_summary_daily AS ab JOIN main_account AS ma ON ab.main_account_id = ma.id
SET ab.main_account_cloud_id = ma.cloud_id
WHERE ab.main_account_cloud_id = '';

-- account_bill_summary_main
UPDATE account_bill_summary_main AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id
SET ab.root_account_cloud_id = ra.cloud_id
WHERE ab.root_account_cloud_id = '';
UPDATE account_bill_summary_main AS ab JOIN main_account AS ma ON ab.main_account_id = ma.id
SET ab.main_account_cloud_id = ma.cloud_id
WHERE ab.main_account_cloud_id = '';

-- account_bill_summary_root
UPDATE account_bill_summary_root AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id
SET ab.root_account_cloud_id = ra.cloud_id
WHERE ab.root_account_cloud_id = '';

-- account_bill_month_task
UPDATE account_bill_month_task AS ab JOIN root_account AS ra ON ab.root_account_id = ra.id
SET ab.root_account_cloud_id = ra.cloud_id
WHERE ab.root_account_cloud_id = '';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT