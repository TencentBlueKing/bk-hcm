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
    1. 修改`account_bill_month_task`表，增加`type`字段
    2.
    3. 修改`account_bill_month_task`表，修改`summary_detail`字段为json类型
    3. 修改`account_bill_sync_record`表，修改`detail`字段为json类型
    4. 修改`load_balancer_target`表，`增加target_group_id` 索引
*/

START TRANSACTION;

ALTER TABLE account_bill_month_task
    ADD `type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '任务类型' AFTER `vendor`;

ALTER TABLE account_bill_month_task DROP KEY `idx_root_account_id_year_month`;
ALTER TABLE account_bill_month_task
    ADD UNIQUE KEY `idx_root_account_id_year_month_type` (`root_account_id`,`bill_year`,`bill_month`,`type`);


UPDATE account_bill_month_task SET summary_detail = '[]' WHERE summary_detail='' ;
ALTER TABLE account_bill_month_task MODIFY `summary_detail` JSON NOT NULL  ;

UPDATE account_bill_sync_record SET detail = '[]' WHERE detail='' ;
ALTER TABLE account_bill_sync_record MODIFY `detail` JSON NOT NULL  ;

ALTER TABLE  load_balancer_target ADD INDEX index_target_group_id(`target_group_id`);



CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;

