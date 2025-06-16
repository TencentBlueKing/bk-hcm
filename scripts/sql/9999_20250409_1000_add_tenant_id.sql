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
    2. 【开启多租户开关】之后，再需要执行该脚本（不开启多租户开关的话，执行之后可能会影响索引查询效率）
*/

START TRANSACTION;

# 目前该表有唯一索引
alter table `account`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_name`,
    add unique index `idx_name_tenant_id`(`name`, `tenant_id`);

# 目前该表有唯一索引
alter table `account_bill_exchange_rate`
    add column tenant_id varchar(64) default 'default' not null after `exchange_rate`,
    drop index `idx_uk_year_month_from_currency_to_currency`,
    add unique index `idx_year_month_from_currency_to_currency_tenant_id`(`year`,`month`,`from_currency`,`to_currency`,`tenant_id`);

# 目前该表没有任何索引
alter table `account_bill_sync_record`
    add column tenant_id varchar(64) default 'default' not null after `adjustment_flow_id`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `application`
    add column tenant_id varchar(64) default 'default' not null after `delivery_detail`,
    drop index `idx_uk_source_sn`,
    add unique index `idx_source_sn_tenant_id`(`source`,`sn`,`tenant_id`);

# 目前该表有唯一索引
alter table `approval_process`
    add column tenant_id varchar(64) default 'default' not null after `service_id`,
    drop index `idx_uk_type`,
    add unique index `idx_type_tenant_id`(`application_type`,`tenant_id`);

# 目前该表有普通索引，可以设置租户ID的单独索引
alter table `audit`
    add column tenant_id varchar(64) default 'default' not null,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `cloud_selection_biz_type`
    add column tenant_id varchar(64) default 'default' not null after `deployment_architecture`,
    drop index `idx_uk_biz_type`,
    add unique index `idx_biz_type_tenant_id`(`biz_type`,`tenant_id`);

# 目前该表有唯一索引
alter table `cloud_selection_idc`
    add column tenant_id varchar(64) default 'default' not null after `region`,
    drop index `idx_uk_bk_biz_id_name`,
    add unique index `idx_bk_biz_id_name_tenant_id`(`bk_biz_id`,`name`,`tenant_id`);

# 目前该表有唯一索引
alter table `cloud_selection_scheme`
    add column tenant_id varchar(64) default 'default' not null after `result_idc_ids`,
    drop index `idx_uk_bk_biz_id_name`,
    add unique index `idx_bk_biz_id_name_tenant_id`(`bk_biz_id`,`name`,`tenant_id`);

# 目前该表有唯一索引
alter table `main_account`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `root_account`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `root_account_bill_config`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_vendor_account_id`,
    add unique index `idx_vendor_root_account_id_tenant_id`(`vendor`,`root_account_id`,`tenant_id`);

# 目前该表有唯一索引
alter table `user_collection`
    add column tenant_id varchar(64) default 'default' not null after `res_id`,
    drop index `idx_uk_user_res_type_res_id`,
    add unique index `idx_user_res_type_res_id_tenant_id`(`user`,`res_type`,`res_id`,`tenant_id`);

# 资源相关的表
# 目前该表有唯一索引
alter table `argument_template`
    add column tenant_id varchar(64) default 'default' not null after `memo`,
    drop index `idx_uk_bk_biz_id_cloud_id`,
    add unique index `idx_bk_biz_id_cloud_id_tenant_id`(`bk_biz_id`,`cloud_id`,`tenant_id`);

# 目前该表有唯一索引
alter table `aws_region`
    add column tenant_id varchar(64) default 'default' not null after `endpoint`,
    drop index `idx_uk_account_id_region_id`,
    add unique index `idx_account_id_region_id_tenant_id`(`account_id`,`region_id`,`tenant_id`);

# 目前该表有唯一索引
alter table `aws_route`
    add column tenant_id varchar(64) default 'default' not null after `propagated`,
    drop index `idx_uk_route_table_id_destination_cidr_block`,
    drop index `idx_uk_route_table_id_destination_ipv6_cidr_block`,
    drop index `idx_uk_route_table_id_cloud_dest_prefix_list_id`,
    add unique index `idx_route_table_id_destination_cidr_block_tenant_id`(`route_table_id`,`destination_cidr_block`,`tenant_id`),
    add unique index `idx_route_table_id_destination_ipv6_cidr_block_tenant_id`(`route_table_id`,`destination_ipv6_cidr_block`,`tenant_id`),
    add unique index `idx_route_table_id_cloud_dest_prefix_list_id_tenant_id`(`route_table_id`,`cloud_destination_prefix_list_id`,`tenant_id`);

# 目前该表没有任何索引
alter table `azure_region`
    add column tenant_id varchar(64) default 'default' not null after `paired_region_id`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `azure_route`
    add column tenant_id varchar(64) default 'default' not null after `provisioning_state`,
    drop index `idx_cloud_id`,
    drop index `idx_uk_route_table_id_name`,
    drop index `idx_uk_route_table_id_address_prefix`,
    add unique index `idx_cloud_id_tenant_id`(`cloud_id`,`tenant_id`),
    add unique index `idx_route_table_id_name_tenant_id`(`route_table_id`,`name`,`tenant_id`),
    add unique index `idx_route_table_id_address_prefix_tenant_id`(`route_table_id`,`address_prefix`,`tenant_id`);

# 目前该表没有任何索引
alter table `azure_resource_group`
    add column tenant_id varchar(64) default 'default' not null after `account_id`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `cvm`
    add column tenant_id varchar(64) default 'default' not null,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表没有任何索引
alter table `disk`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `eip`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `gcp_firewall_rule`
    add column tenant_id varchar(64) default 'default' not null after `bk_biz_id`,
    drop index `idx_uk_cloud_id`,
    drop index `idx_uk_account_id_name`,
    add unique index `idx_cloud_id_tenant_id`(`cloud_id`,`tenant_id`),
    add unique index `idx_account_id_name_tenant_id`(`account_id`,`name`,`tenant_id`);

# 目前该表有唯一索引
alter table `gcp_region`
    add column tenant_id varchar(64) default 'default' not null after `self_link`,
    drop index `idx_uk_region_id_status`,
    add unique index `idx_region_id_status_tenant_id`(`region_id`,`status`,`tenant_id`);

# 目前该表有唯一索引
alter table `gcp_route`
    add column tenant_id varchar(64) default 'default' not null after `memo`,
    drop index `idx_uk_cloud_id`,
    add unique index `idx_cloud_id_tenant_id`(`cloud_id`,`tenant_id`);

# 目前该表没有任何索引
alter table `huawei_region`
    add column tenant_id varchar(64) default 'default' not null after `locales_es_es`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `huawei_route`
    add column tenant_id varchar(64) default 'default' not null after `memo`,
    drop index `idx_uk_route_table_id_destination`,
    add unique index `idx_route_table_id_destination_tenant_id`(`route_table_id`,`destination`,`tenant_id`);

# 目前该表没有任何索引
alter table `image`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `load_balancer`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor_region`,
    add unique index `idx_cloud_id_vendor_region_tenant_id`(`cloud_id`,`vendor`,`region`,`tenant_id`);

# 目前该表有唯一索引
alter table `network_interface`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表没有任何索引
alter table `recycle_record`
    add column tenant_id varchar(64) default 'default' not null after `status`,
    add index `idx_tenant_id`(`tenant_id`);

# 目前该表有唯一索引
alter table `route_table`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `security_group`
    add column tenant_id varchar(64) default 'default' not null after `tags`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `ssl_cert`
    add column tenant_id varchar(64) default 'default' not null after `memo`,
    drop index `idx_uk_bk_biz_id_cloud_id`,
    add unique index `idx_bk_biz_id_cloud_id_tenant_id`(`bk_biz_id`,`cloud_id`,`tenant_id`);

# 目前该表有唯一索引
alter table `subnet`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `tcloud_region`
    add column tenant_id varchar(64) default 'default' not null after `status`,
    drop index `idx_uk_region_id_status`,
    add unique index `idx_region_id_status_tenant_id`(`region_id`,`status`,`tenant_id`);

# 目前该表有唯一索引
alter table `tcloud_route`
    add column tenant_id varchar(64) default 'default' not null after `memo`,
    drop index `idx_uk_cloud_route_table_id_cloud_id`,
    add unique index `idx_cloud_route_table_id_cloud_id_tenant_id`(`cloud_route_table_id`,`cloud_id`,`tenant_id`);

# 目前该表有唯一索引
alter table `vpc`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

# 目前该表有唯一索引
alter table `zone`
    add column tenant_id varchar(64) default 'default' not null after `extension`,
    drop index `idx_uk_cloud_id_vendor`,
    add unique index `idx_cloud_id_vendor_tenant_id`(`cloud_id`,`vendor`,`tenant_id`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9990' as `sql_ver`;

COMMIT;
