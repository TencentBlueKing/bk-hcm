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
    1. 在需要支持多租户的数据表中，新增租户ID
*/

START TRANSACTION;

alter table `account`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_name`,
    add unique index `idx_uk_tenant_id_name`(`tenant_id`,`name`);

alter table `account_bill_exchange_rate`
    add column tenant_id varchar(64) default '' not null after `exchange_rate`,
    drop index `idx_uk_year_month_from_currency_to_currency`,
    add unique index `idx_uk_tenant_id_year_month_from_currency_to_currency`(`tenant_id`,`year`,`month`,`from_currency`,`to_currency`);

alter table `account_bill_sync_record`
    add column tenant_id varchar(64) default '' not null after `adjustment_flow_id`,
    add index `idx_tenant_id_vendor`(`tenant_id`,`vendor`);

alter table `application`
    add column tenant_id varchar(64) default '' not null after `delivery_detail`,
    drop index `idx_uk_source_sn`,
    add unique index `idx_uk_tenant_id_source_sn`(`tenant_id`,`source`,`sn`);

alter table `approval_process`
    add column tenant_id varchar(64) default '' not null after `service_id`,
    drop index `idx_uk_type`,
    add unique index `idx_uk_tenant_id_type`(`tenant_id`,`application_type`);

alter table `audit`
    add column tenant_id varchar(64) default '' not null,
    drop index `idx_bk_biz_id`,
    add index `idx_tenant_id_bk_biz_id`(`tenant_id`,`bk_biz_id`,`id`);

alter table `cloud_selection_biz_type`
    add column tenant_id varchar(64) default '' not null after `deployment_architecture`,
    drop index `idx_uk_biz_type`,
    add unique index `idx_tenant_id_biz_type`(`tenant_id`,`biz_type`);

alter table `cloud_selection_idc`
    add column tenant_id varchar(64) default '' not null after `region`,
    drop index `idx_uk_bk_biz_id_name`,
    add unique index `idx_tenant_id_bk_biz_id_name`(`tenant_id`,`bk_biz_id`,`name`);

alter table `cloud_selection_scheme`
    add column tenant_id varchar(64) default '' not null after `result_idc_ids`,
    drop index `idx_uk_bk_biz_id_name`,
    add unique index `idx_tenant_id_bk_biz_id_name`(`tenant_id`,`bk_biz_id`,`name`);

alter table `main_account`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `root_account`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `root_account_bill_config`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_vendor_account_id`,
    add unique index `idx_tenant_id_vendor_root_account_id`(`tenant_id`,`vendor`,`root_account_id`);

alter table `user_collection`
    add column tenant_id varchar(64) default '' not null after `res_id`,
    drop index `idx_uk_user_res_type_res_id`,
    add unique index `idx_tenant_id_user_res_type_res_id`(`tenant_id`,`user`,`res_type`,`res_id`);

# 资源相关的表
alter table `argument_template`
    add column tenant_id varchar(64) default '' not null after `memo`,
    drop index `idx_uk_bk_biz_id_cloud_id`,
    add unique index `idx_tenant_id_bk_biz_id_cloud_id`(`tenant_id`,`bk_biz_id`,`cloud_id`);

alter table `aws_region`
    add column tenant_id varchar(64) default '' not null after `endpoint`,
    drop index `idx_uk_account_id_region_id`,
    drop index `idx_uk_vendor`,
    add unique index `idx_tenant_id_account_id_region_id`(`tenant_id`,`account_id`,`region_id`),
    add index `idx_tenant_id_vendor`(`tenant_id`,`vendor`);

alter table `aws_route`
    add column tenant_id varchar(64) default '' not null after `propagated`,
    drop index `idx_uk_route_table_id_destination_cidr_block`,
    drop index `idx_uk_route_table_id_destination_ipv6_cidr_block`,
    drop index `idx_uk_route_table_id_cloud_dest_prefix_list_id`,
    add unique index `idx_tenant_id_route_table_id_destination_cidr_block`(`tenant_id`,`route_table_id`,`destination_cidr_block`),
    add unique index `idx_tenant_id_route_table_id_destination_ipv6_cidr_block`(`tenant_id`,`route_table_id`,`destination_ipv6_cidr_block`),
    add unique index `idx_tenant_id_route_table_id_cloud_dest_prefix_list_id`(`tenant_id`,`route_table_id`,`cloud_destination_prefix_list_id`);

alter table `azure_region`
    add column tenant_id varchar(64) default '' not null after `paired_region_id`,
    add index `idx_tenant_id`(`tenant_id`);

alter table `azure_route`
    add column tenant_id varchar(64) default '' not null after `provisioning_state`,
    drop index `idx_cloud_id`,
    drop index `idx_uk_route_table_id_name`,
    drop index `idx_uk_route_table_id_address_prefix`,
    add unique index `idx_tenant_id_cloud_id`(`tenant_id`,`cloud_id`),
    add unique index `idx_tenant_id_route_table_id_name`(`tenant_id`,`route_table_id`,`name`),
    add unique index `idx_tenant_id_route_table_id_address_prefix`(`tenant_id`,`route_table_id`,`address_prefix`);

alter table `azure_resource_group`
    add column tenant_id varchar(64) default '' not null after `account_id`,
    add index `idx_tenant_id`(`tenant_id`);

alter table `cvm`
    add column tenant_id varchar(64) default '' not null,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`),
    add index `idx_tenant_id_vendor_id`(`tenant_id`,`vendor`,`id`),
    add index `idx_tenant_id_bk_biz_id_created_at`(`tenant_id`,`bk_biz_id`,`created_at`);

alter table `disk`
    add column tenant_id varchar(64) default '' not null after `extension`,
    add index `idx_tenant_id`(`tenant_id`);

alter table `eip`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `gcp_firewall_rule`
    add column tenant_id varchar(64) default '' not null after `bk_biz_id`,
    drop index `idx_uk_cloud_id`,
    drop index `idx_uk_account_id_name`,
    add unique index `idx_tenant_id_cloud_id`(`tenant_id`,`cloud_id`),
    add unique index `idx_tenant_id_account_id_name`(`tenant_id`,`account_id`,`name`);

alter table `gcp_region`
    add column tenant_id varchar(64) default '' not null after `self_link`,
    drop index `idx_uk_region_id_status`,
    drop index `idx_uk_vendor`,
    add unique index `idx_tenant_id_region_id_status`(`tenant_id`,`region_id`,`status`),
    add index `idx_tenant_id_vendor`(`tenant_id`,`vendor`);

alter table `gcp_route`
    add column tenant_id varchar(64) default '' not null after `memo`,
    drop index `idx_uk_cloud_id`,
    add unique index `idx_tenant_id_cloud_id`(`tenant_id`,`cloud_id`);

alter table `huawei_region`
    add column tenant_id varchar(64) default '' not null after `locales_es_es`,
    add index `idx_tenant_id`(`tenant_id`);

alter table `huawei_route`
    add column tenant_id varchar(64) default '' not null after `memo`,
    drop index `idx_uk_route_table_id_destination`,
    add unique index `idx_tenant_id_route_table_id_destination`(`tenant_id`,`route_table_id`,`destination`);

alter table `huawei_security_group_rule`
    add column tenant_id varchar(64) default '' not null after `priority`,
    drop index `idx_uk_cloud_id`,
    add unique index `idx_tenant_id_cloud_id`(`tenant_id`,`cloud_id`);

alter table `image`
    add column tenant_id varchar(64) default '' not null after `extension`,
    add index `idx_tenant_id`(`tenant_id`);

alter table `load_balancer`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor_region`,
    add unique index `idx_tenant_id_cloud_id_vendor_region`(`tenant_id`,`cloud_id`,`vendor`,`region`);

alter table `network_interface`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `recycle_record`
    add column tenant_id varchar(64) default '' not null after `status`,
    add index `idx_tenant_id`(`tenant_id`);

alter table `route_table`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `security_group`
    add column tenant_id varchar(64) default '' not null after `tags`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `ssl_cert`
    add column tenant_id varchar(64) default '' not null after `memo`,
    drop index `idx_uk_bk_biz_id_cloud_id`,
    add unique index `idx_tenant_id_bk_biz_id_cloud_id`(`tenant_id`,`bk_biz_id`,`cloud_id`);

alter table `subnet`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `tcloud_region`
    add column tenant_id varchar(64) default '' not null after `status`,
    drop index `idx_uk_region_id_status`,
    drop index `idx_uk_vendor`,
    add unique index `idx_uk_tenant_id_region_id_status`(`tenant_id`,`region_id`,`status`),
    add index `idx_tenant_id_vendor`(`tenant_id`,`vendor`);

alter table `tcloud_route`
    add column tenant_id varchar(64) default '' not null after `memo`,
    drop index `idx_uk_cloud_route_table_id_cloud_id`,
    add unique index `idx_tenant_id_cloud_route_table_id_cloud_id`(`tenant_id`,`cloud_route_table_id`,`cloud_id`);

alter table `vpc`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

alter table `zone`
    add column tenant_id varchar(64) default '' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_tenant_id_cloud_id_vendor`(`tenant_id`,`cloud_id`,`vendor`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9990' as `sql_ver`;

COMMIT;
