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
    1. clb优化索引
*/

START TRANSACTION;

-- tcloud_lb_url_rule
alter table tcloud_lb_url_rule
    add index idx_cloud_lbl_id_rule_type (cloud_lbl_id, rule_type);
alter table tcloud_lb_url_rule
    add index idx_lb_id_rule_type (lb_id, rule_type);

-- load_balancer_target
alter table load_balancer_target
    add index idx_target_group_id (target_group_id);

-- target_group_listener_rule_rel
alter table target_group_listener_rule_rel
    add index idx_listener_rule_id (listener_rule_id);
alter table target_group_listener_rule_rel
    add index idx_lb_id_cloud_rule_id (lb_id, cloud_listener_rule_id);
alter table target_group_listener_rule_rel
    add index idx_lb_id_cloud_lbl_id (lb_id, cloud_lbl_id);
alter table target_group_listener_rule_rel
    add index idx_lbl_id (lbl_id);

-- load_balancer_listener
alter table load_balancer_listener
    add index idx_vendor_account_id_bk_biz_id_cloud_lb_id (vendor, account_id, bk_biz_id, cloud_lb_id);

-- load_balancer_target_group
alter table load_balancer_target_group
    add index idx_bk_biz_id (bk_biz_id);

-- audit
alter table audit
    add index idx_bk_biz_id (bk_biz_id);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;