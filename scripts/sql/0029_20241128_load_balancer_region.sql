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
    SQLVER=0029,HCMVER=v1.7.0

    Notes:
    1. 修改负载均衡表唯一键索引为cloud_id, vendor, region
    2. 负载均衡监听器表新增region字段
    3. 负载均衡监听器表修改唯一键索引为(cloud_id, vendor, region) 并增加(lb_id,cloud_id)索引
    4. 负载均衡目标组表修改唯一键索引为(cloud_id, vendor, region)
    5. 腾讯云URL规则表新增region字段
    6. 腾讯云URL规则表修改唯一键索引为(cloud_id, vendor, region)
    7. 负载均衡target表新增target_group_region
    8. 负载均衡target表修改唯一键索引为(cloud_target_group_id, ip, port, cloud_inst_id, target_group_region)
*/

START TRANSACTION;

-- 1. 修改负载均衡表唯一键索引为cloud_id, vendor, region
alter table load_balancer
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table load_balancer
    drop key idx_uk_cloud_id_vendor;


-- 2. 负载均衡监听器表新增region字段, 并补充索引 (vendor, account_id, bk_biz_id, cloud_lb_id)
alter table load_balancer_listener
    add column region varchar(20) default '' not null after `default_domain`;


-- 3. 负载均衡监听器表 修改唯一键索引为(cloud_id, vendor, region)，并增加(lb_id,cloud_id)索引
alter table load_balancer_listener
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table load_balancer_listener
    drop key idx_uk_cloud_id_vendor;
alter table load_balancer_listener
    add index idx_lb_id_cloud_id (lb_id, cloud_id, id);

alter table load_balancer_listener
    add index idx_vendor_account_id_bk_biz_id_cloud_lb_id (vendor, account_id, bk_biz_id, cloud_lb_id);

-- 4. 负载均衡目标组表修改唯一键索引为cloud_id, vendor, region
alter table load_balancer_target_group
    add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region);
alter table load_balancer_target_group
    drop key idx_uk_cloud_id_vendor;

alter table load_balancer_target_group
    add index idx_bk_biz_id (bk_biz_id, id);

-- 5. 腾讯云URL规则表新增region字段
alter table tcloud_lb_url_rule
    add column region varchar(20) default '' not null after `cloud_target_group_id`;

-- 6. 腾讯云URL规则表修改唯一键索引为cloud_id, vendor, region, 增加索引(lbl_id, rule_type)、(lb_id, rule_type)
alter table tcloud_lb_url_rule
    add constraint idx_uk_cloud_id_cloud_lbl_id_region unique (cloud_id, cloud_lbl_id, region);
alter table tcloud_lb_url_rule
    drop key idx_uk_cloud_id_lbl_id;

alter table tcloud_lb_url_rule
    add index idx_cloud_lbl_id_rule_type (lbl_id, rule_type);
alter table tcloud_lb_url_rule
    add index idx_lb_id_rule_type (lb_id, rule_type);

-- 7. 负载均衡target表新增target_group_region，并增加索引target_group_id
alter table load_balancer_target
    add column target_group_region varchar(20) default '' not null after `inst_name`;

alter table load_balancer_target
    add index idx_target_group_id (target_group_id, id);

-- 8. 负载均衡target表修改唯一键索引为cloud_target_group_id, ip, port, cloud_inst_id, target_group_region
alter table load_balancer_target
    add constraint idx_uk_cloud_target_group_id_ip_port_target_group_region
        unique (cloud_target_group_id, ip, port, target_group_region);
alter table load_balancer_target
    drop key idx_uk_cloud_target_group_id_ip_port_cloud_inst_id;

-- 9. 负载均衡表新增tags字段
alter table load_balancer
    add tags JSON null after cloud_expired_time;

-- 10. 目标组规则关系表 target_group_listener_rule_rel 增加索引
alter table target_group_listener_rule_rel
    add index idx_listener_rule_id (listener_rule_id);
alter table target_group_listener_rule_rel
    add index idx_lb_id_cloud_rule_id (lb_id, cloud_listener_rule_id);
alter table target_group_listener_rule_rel
    add index idx_lb_id_cloud_lbl_id (lb_id, cloud_lbl_id, id);
alter table target_group_listener_rule_rel
    add index idx_lbl_id (lbl_id, id);

-- 11. audit审计表增加业务id索引
alter table audit
    add index idx_bk_biz_id (bk_biz_id, id);


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.7.0' as `hcm_ver`, '0029' as `sql_ver`;


COMMIT
