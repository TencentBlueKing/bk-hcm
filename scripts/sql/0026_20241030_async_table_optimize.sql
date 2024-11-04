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
    SQLVER=0026,HCMVER=v1.6.10

    Notes:
    1. 修改`async_flow`表，增加`worker`,`state`索引
    2. 修改`async_flow`表，增加`state`索引
    3. 修改`account_bill_item`表，增加(`root_account_id`, `main_account_id`, `bill_day`)索引
    4. 修改`account_bill_sync_record`表，增加`adjustment_flow_id`字段
*/

START TRANSACTION;

alter table async_flow
    add index idx_worker_state_id (worker, state);
alter table async_flow
    add index idx_state_id (state, id);

alter table account_bill_item
    add index `idx_root_main_bill_day` (`root_account_id`, `main_account_id`, `bill_day`);


alter table account_bill_sync_record
    add adjustment_flow_id varchar(64) not null default '' after detail;

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.6.10' as `hcm_ver`, '0026' as `sql_ver`;

COMMIT;

