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

// Package migration is the migration init_db.
package migration

import (
	"context"

	"hcm/migrate/register"
	"hcm/migrate/util"
	"hcm/pkg/dal/dao/orm"
)

// ID is the unique identifier of this migration file, in the form
// <date>-<time>-<desc>-<random>. It must never change once executed.
// Migrations that share an ID run only the earlier one in execution order.
const ID = "20230227-2045-INIT-DB-D1BA"

func init() {
	register.Main.Regist(ID, "v1.0.0", "20230227204500", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0001_20230227_2045_init_db.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "id_generator", "create table if not exists `id_generator`\n"+
		"(\n"+
		"    `resource` varchar(64) not null,\n"+
		"    `max_id`   varchar(64) not null,\n"+
		"    primary key (`resource`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "security_group", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "tcloud_security_group_rule", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "aws_security_group_rule", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "azure_security_group_rule", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "huawei_security_group_rule", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "gcp_firewall_rule", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "vpc", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "subnet", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "disk", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "tcloud_region", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "aws_region", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "eip", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "huawei_region", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "azure_region", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "zone", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "image", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "cvm", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "azure_resource_group", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "gcp_region", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "route_table", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "tcloud_route", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "aws_route", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "azure_route", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "huawei_route", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "gcp_route", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "application", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "approval_process", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "network_interface", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "recycle_record", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "audit", "create table if not exists `audit`\n"+
		"(\n"+
		"    `id`           bigint(1) unsigned not null auto_increment,\n"+
		"    `res_id`       varchar(64)                 default '',\n"+
		"    `cloud_res_id` varchar(255)                default '',\n"+
		"    `res_name`     varchar(255)                default '',\n"+
		"    `res_type`     varchar(50)        not null,\n"+
		"    `action`       varchar(20)        not null,\n"+
		"    `bk_biz_id`    bigint(1)          not null default -1,\n"+
		"    `vendor`       varchar(16)                 default '',\n"+
		"    `account_id`   varchar(64)                 default '',\n"+
		"    `operator`     varchar(64)        not null,\n"+
		"    `source`       varchar(20)        not null,\n"+
		"    `rid`          varchar(64)        not null,\n"+
		"    `app_code`     varchar(64)                 default '',\n"+
		"    `detail`       json                        default null,\n"+
		"    `created_at`   timestamp          not null default current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account", "create table if not exists `account`\n"+
		"(\n"+
		"    `id`             varchar(64) not null,\n"+
		"    `vendor`         varchar(16) not null,\n"+
		"    `name`           varchar(64) not null,\n"+
		"    `managers`       json        not null,\n"+
		"    `type`           varchar(32) not null,\n"+
		"    `site`           varchar(32) not null,\n"+
		"    `sync_status`    varchar(32) not null,\n"+
		"    `price`          varchar(16)          default '',\n"+
		"    `price_unit`     varchar(8)           default '',\n"+
		"    `memo`           varchar(255)         default '',\n"+
		"    `extension`      json        not null,\n"+
		"    `creator`        varchar(64) not null,\n"+
		"    `reviser`        varchar(64) not null,\n"+
		"    `created_at`     timestamp   not null default current_timestamp,\n"+
		"    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_name` (`name`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_biz_rel", "create table if not exists `account_biz_rel`\n"+
		"(\n"+
		"    `id`         bigint(1) unsigned not null auto_increment,\n"+
		"    `bk_biz_id`  bigint(1)          not null,\n"+
		"    `account_id` varchar(64)        not null,\n"+
		"    `creator`    varchar(64)        not null,\n"+
		"    `created_at` timestamp          not null default current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_bk_biz_id_account_id` (`bk_biz_id`, `account_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "security_group", "create table if not exists `security_group`\n"+
		"(\n"+
		"    `id`                      varchar(64)  not null,\n"+
		"    `vendor`                  varchar(16)  not null,\n"+
		"    `cloud_id`                varchar(255) not null,\n"+
		"    `bk_biz_id`               bigint(1)    not null default -1,\n"+
		"    `region`                  varchar(20)  not null,\n"+
		"    `name`                    varchar(60)  not null,\n"+
		"    `account_id`              varchar(64)  not null,\n"+
		"    `memo`                    varchar(255)          default '',\n"+
		"    `association_template_id` varchar(64)           default 0,\n"+
		"    `extension`               json         not null,\n"+
		"    `creator`                 varchar(64)  not null,\n"+
		"    `reviser`                 varchar(64)  not null,\n"+
		"    `created_at`              timestamp    not null default current_timestamp,\n"+
		"    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "tcloud_security_group_rule", "create table if not exists `tcloud_security_group_rule`\n"+
		"(\n"+
		"    `id`                             varchar(64)  not null,\n"+
		"    `cloud_policy_index`             bigint(1)    not null,\n"+
		"    `type`                           varchar(20)  not null,\n"+
		"    `cloud_security_group_id`        varchar(255) not null,\n"+
		"    `security_group_id`              varchar(64)  not null,\n"+
		"    `account_id`                     varchar(64)  not null,\n"+
		"    `region`                         varchar(20)  not null,\n"+
		"    `version`                        varchar(255) not null,\n"+
		"    `action`                         varchar(10)  not null,\n"+
		"    `protocol`                       varchar(10)           default null,\n"+
		"    `port`                           varchar(255)          default null,\n"+
		"    `cloud_service_id`               varchar(255)          default null,\n"+
		"    `cloud_service_group_id`         varchar(255)          default null,\n"+
		"    `ipv4_cidr`                      varchar(255)          default null,\n"+
		"    `ipv6_cidr`                      varchar(255)          default null,\n"+
		"    `cloud_target_security_group_id` varchar(255)          default null,\n"+
		"    `cloud_address_id`               varchar(255)          default null,\n"+
		"    `cloud_address_group_id`         varchar(255)          default null,\n"+
		"    `memo`                           varchar(60)           default null,\n"+
		"    `creator`                        varchar(64)  not null,\n"+
		"    `reviser`                        varchar(64)  not null,\n"+
		"    `created_at`                     timestamp    not null default current_timestamp,\n"+
		"    `updated_at`                     timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_security_group_id_cloud_policy_index_type` (`cloud_security_group_id`, `cloud_policy_index`, `type`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "aws_security_group_rule", "create table if not exists `aws_security_group_rule`\n"+
		"(\n"+
		"    `id`                             varchar(64)  not null,\n"+
		"    `cloud_id`                       varchar(255) not null,\n"+
		"    `cloud_security_group_id`        varchar(255) not null,\n"+
		"    `cloud_group_owner_id`           varchar(255) not null,\n"+
		"    `account_id`                     varchar(64)  not null,\n"+
		"    `region`                         varchar(20)  not null,\n"+
		"    `security_group_id`              varchar(64)  not null,\n"+
		"    `type`                           varchar(20)  not null,\n"+
		"    `ipv4_cidr`                      varchar(255)          default null,\n"+
		"    `ipv6_cidr`                      varchar(255)          default null,\n"+
		"    `memo`                           varchar(60)           default null,\n"+
		"    `from_port`                      bigint(1)             default 0,\n"+
		"    `to_port`                        bigint(1)             default 0,\n"+
		"    `protocol`                       varchar(10)           default null,\n"+
		"    `cloud_prefix_list_id`           varchar(255)          default null,\n"+
		"    `cloud_target_security_group_id` varchar(255)          default null,\n"+
		"    `creator`                        varchar(64)  not null,\n"+
		"    `reviser`                        varchar(64)  not null,\n"+
		"    `created_at`                     timestamp    not null default current_timestamp,\n"+
		"    `updated_at`                     timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id` (`cloud_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "huawei_security_group_rule", "create table if not exists `huawei_security_group_rule`\n"+
		"(\n"+
		"    `id`                            varchar(64)  not null,\n"+
		"    `cloud_id`                      varchar(255) not null,\n"+
		"    `type`                          varchar(20)  not null,\n"+
		"    `cloud_security_group_id`       varchar(255) not null,\n"+
		"    `security_group_id`             varchar(64)  not null,\n"+
		"    `account_id`                    varchar(64)  not null,\n"+
		"    `region`                        varchar(20)  not null,\n"+
		"    `action`                        varchar(10)  not null,\n"+
		"    `cloud_project_id`              varchar(255)          default '',\n"+
		"    `memo`                          varchar(255)          default '',\n"+
		"    `protocol`                      varchar(10)           default '',\n"+
		"    `ethertype`                     varchar(10)           default '',\n"+
		"    `cloud_remote_group_id`         varchar(255)          default '',\n"+
		"    `remote_ip_prefix`              varchar(255)          default '',\n"+
		"    `cloud_remote_address_group_id` varchar(255)          default '',\n"+
		"    `port`                          varchar(255)          default '',\n"+
		"    `priority`                      int(1) unsigned       default 0,\n"+
		"    `creator`                       varchar(64)  not null,\n"+
		"    `reviser`                       varchar(64)  not null,\n"+
		"    `created_at`                    timestamp    not null default current_timestamp,\n"+
		"    `updated_at`                    timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id` (`cloud_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "azure_security_group_rule", "create table if not exists `azure_security_group_rule`\n"+
		"(\n"+
		"    `id`                                       varchar(64)  not null,\n"+
		"    `cloud_id`                                 varchar(255) not null,\n"+
		"    `cloud_security_group_id`                  varchar(255) not null,\n"+
		"    `account_id`                               varchar(64)  not null,\n"+
		"    `security_group_id`                        varchar(64)  not null,\n"+
		"    `type`                                     varchar(20)  not null,\n"+
		"    `region`                                   varchar(20)  not null,\n"+
		"    `provisioning_state`                       varchar(20)  not null,\n"+
		"    `etag`                                     varchar(255)          default '',\n"+
		"    `name`                                     varchar(255)          default '',\n"+
		"    `memo`                                     varchar(140)          default '',\n"+
		"    `destination_address_prefix`               varchar(255)          default '',\n"+
		"    `destination_address_prefixes`             json                  default null,\n"+
		"    `cloud_destination_app_security_group_ids` json                  default null,\n"+
		"    `destination_port_range`                   varchar(255)          default '',\n"+
		"    `destination_port_ranges`                  json                  default null,\n"+
		"    `protocol`                                 varchar(10)           default '',\n"+
		"    `source_address_prefix`                    varchar(255)          default '',\n"+
		"    `source_address_prefixes`                  json                  default null,\n"+
		"    `cloud_source_app_security_group_ids`      json                  default null,\n"+
		"    `source_port_range`                        varchar(255)          default '',\n"+
		"    `source_port_ranges`                       json                  default null,\n"+
		"    `priority`                                 bigint(1)             default 0,\n"+
		"    `access`                                   varchar(20)           default '',\n"+
		"    `creator`                                  varchar(64)  not null,\n"+
		"    `reviser`                                  varchar(64)  not null,\n"+
		"    `created_at`                               timestamp    not null default current_timestamp,\n"+
		"    `updated_at`                               timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id` (`cloud_id`),\n"+
		"    unique key `idx_uk_name` (`name`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "gcp_firewall_rule", "create table if not exists `gcp_firewall_rule`\n"+
		"(\n"+
		"    `id`                      varchar(64)  not null,\n"+
		"    `cloud_id`                varchar(255) not null,\n"+
		"    `name`                    varchar(62)           default '',\n"+
		"    `priority`                bigint(1)             default 0,\n"+
		"    `memo`                    varchar(2048)         default '',\n"+
		"    `cloud_vpc_id`            varchar(255)          default '',\n"+
		"    `vpc_id`                  varchar(64)           default '',\n"+
		"    `vpc_self_link`           varchar(255)          default '',\n"+
		"    `account_id`              varchar(64)           default '',\n"+
		"    `source_ranges`           json                  default null,\n"+
		"    `destination_ranges`      json                  default null,\n"+
		"    `source_tags`             json                  default null,\n"+
		"    `target_tags`             json                  default null,\n"+
		"    `source_service_accounts` json                  default null,\n"+
		"    `target_service_accounts` json                  default null,\n"+
		"    `denied`                  json                  default null,\n"+
		"    `allowed`                 json                  default null,\n"+
		"    `type`                    varchar(20)           default '',\n"+
		"    `log_enable`              boolean               default false,\n"+
		"    `disabled`                boolean               default false,\n"+
		"    `self_link`               varchar(255)          default '',\n"+
		"    `bk_biz_id`               bigint(1)    not null default -1,\n"+
		"    `creator`                 varchar(64)  not null,\n"+
		"    `reviser`                 varchar(64)  not null,\n"+
		"    `created_at`              timestamp    not null default current_timestamp,\n"+
		"    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id` (`cloud_id`),\n"+
		"    unique key `idx_uk_name` (`name`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "vpc", "create table if not exists `vpc`\n"+
		"(\n"+
		"    `id`          varchar(64)  not null,\n"+
		"    `vendor`      varchar(16)  not null,\n"+
		"    `account_id`  varchar(64)  not null,\n"+
		"    `cloud_id`    varchar(255) not null,\n"+
		"    `name`        varchar(128) not null,\n"+
		"    `region`      varchar(255) not null,\n"+
		"    `category`    varchar(32)  not null,\n"+
		"    `memo`        varchar(255)          default '',\n"+
		"    `bk_cloud_id` bigint(1)             default -1,\n"+
		"    `bk_biz_id`   bigint(1)    not null default -1,\n"+
		"    \n"+
		"    `extension`   json         not null,\n"+
		"    \n"+
		"    `creator`     varchar(64)  not null,\n"+
		"    `reviser`     varchar(64)  not null,\n"+
		"    `created_at`  timestamp    not null default current_timestamp,\n"+
		"    `updated_at`  timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "subnet", "create table if not exists `subnet`\n"+
		"(\n"+
		"    `id`                   varchar(64)  not null,\n"+
		"    `vendor`               varchar(16)  not null,\n"+
		"    `account_id`           varchar(64)  not null,\n"+
		"    `cloud_vpc_id`         varchar(255) not null,\n"+
		"    `cloud_route_table_id` varchar(255)          default '',\n"+
		"    `cloud_id`             varchar(255) not null,\n"+
		"    `name`                 varchar(128) not null,\n"+
		"    `region`               varchar(255) not null,\n"+
		"    `zone`                 varchar(255) not null,\n"+
		"    `ipv4_cidr`            json         not null,\n"+
		"    `ipv6_cidr`            json         not null,\n"+
		"    `memo`                 varchar(255)          default '',\n"+
		"    `vpc_id`               varchar(64)  not null,\n"+
		"    `route_table_id`       varchar(64)           default '',\n"+
		"    `bk_biz_id`            bigint(1)    not null default -1,\n"+
		"    \n"+
		"    `extension`            json         not null,\n"+
		"    \n"+
		"    `creator`              varchar(64)  not null,\n"+
		"    `reviser`              varchar(64)  not null,\n"+
		"    `created_at`           timestamp    not null default current_timestamp,\n"+
		"    `updated_at`           timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "huawei_region", "create table if not exists `huawei_region`\n"+
		"(\n"+
		"    `id`            varchar(64) not null,\n"+
		"    `region_id`     varchar(64) not null,\n"+
		"    `type`          varchar(20) not null,\n"+
		"    `service`       varchar(20) not null,\n"+
		"    `locales_pt_br` varchar(20)          default '',\n"+
		"    `locales_zh_cn` varchar(20)          default '',\n"+
		"    `locales_en_us` varchar(20)          default '',\n"+
		"    `locales_es_us` varchar(20)          default '',\n"+
		"    `locales_es_es` varchar(20)          default '',\n"+
		"    `creator`       varchar(64) not null,\n"+
		"    `reviser`       varchar(64) not null,\n"+
		"    `created_at`    timestamp   not null default current_timestamp,\n"+
		"    `updated_at`    timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "azure_resource_group", "create table if not exists `azure_resource_group`\n"+
		"(\n"+
		"    `id`         varchar(64) not null,\n"+
		"    `name`       varchar(64) not null,\n"+
		"    `type`       varchar(64) not null,\n"+
		"    `location`   varchar(64) not null,\n"+
		"    `account_id` varchar(64) not null,\n"+
		"    `creator`    varchar(64) not null,\n"+
		"    `reviser`    varchar(64) not null,\n"+
		"    `created_at` timestamp   not null default current_timestamp,\n"+
		"    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "disk", "create table if not exists `disk`\n"+
		"(\n"+
		"    `id`             varchar(64)        not null,\n"+
		"    `vendor`         varchar(16)        not null,\n"+
		"    `name`           varchar(128)       not null,\n"+
		"    `account_id`     varchar(64)        not null,\n"+
		"    `cloud_id`       varchar(255)       not null,\n"+
		"    `bk_biz_id`      bigint(1)          not null default -1,\n"+
		"    `region`         varchar(128)       not null,\n"+
		"    `zone`           varchar(128)       not null,\n"+
		"    `disk_size`      bigint(1) unsigned not null,\n"+
		"    `disk_type`      varchar(128)       not null,\n"+
		"    `status`         varchar(128)       not null,\n"+
		"    `recycle_status` varchar(32)                 default '',\n"+
		"    `is_system_disk` boolean                     default false,\n"+
		"    `memo`           varchar(255)                default '',\n"+
		"    `extension`      json               not null,\n"+
		"    `creator`        varchar(64)        not null,\n"+
		"    `reviser`        varchar(64)        not null,\n"+
		"    `created_at`     timestamp          not null default current_timestamp,\n"+
		"    `updated_at`     timestamp          not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "azure_region", "create table if not exists `azure_region`\n"+
		"(\n"+
		"    `id`                  varchar(64)  not null,\n"+
		"    `cloud_id`            varchar(255) not null,\n"+
		"    `name`                varchar(64)  not null,\n"+
		"    `type`                varchar(64)  not null,\n"+
		"    `display_name`        varchar(64)  not null,\n"+
		"    `region_display_name` varchar(64)  not null,\n"+
		"    `geography_group`     varchar(64)  not null,\n"+
		"    `latitude`            varchar(64)           default '',\n"+
		"    `longitude`           varchar(64)  not null,\n"+
		"    `physical_location`   varchar(64)           default '',\n"+
		"    `region_type`         varchar(64)  not null,\n"+
		"    `paired_region_name`  varchar(64)           default '',\n"+
		"    `paired_region_id`    varchar(255)          default '',\n"+
		"    `creator`             varchar(64)  not null,\n"+
		"    `reviser`             varchar(64)  not null,\n"+
		"    `created_at`          timestamp    not null default current_timestamp,\n"+
		"    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "tcloud_region", "create table if not exists `tcloud_region`\n"+
		"(\n"+
		"    `id`          varchar(64) not null,\n"+
		"    `vendor`      varchar(16) not null,\n"+
		"    `region_id`   varchar(32) not null,\n"+
		"    `region_name` varchar(64) not null,\n"+
		"    `status`      varchar(32)          default '',\n"+
		"    `creator`     varchar(64)          default '',\n"+
		"    `reviser`     varchar(64)          default '',\n"+
		"    `created_at`  timestamp   not null default current_timestamp,\n"+
		"    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_region_id_status` (`region_id`, `status`),\n"+
		"    key `idx_uk_vendor` (`vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4 comment ='云厂商支持的地区列表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "aws_region", "create table if not exists `aws_region`\n"+
		"(\n"+
		"    `id`          varchar(64) not null,\n"+
		"    `vendor`      varchar(16) not null,\n"+
		"    `region_id`   varchar(32) not null,\n"+
		"    `region_name` varchar(64) not null,\n"+
		"    `status`      varchar(32)          default '',\n"+
		"    `endpoint`    varchar(64)          default '',\n"+
		"    `creator`     varchar(64)          default '',\n"+
		"    `reviser`     varchar(64)          default '',\n"+
		"    `created_at`  timestamp   not null default current_timestamp,\n"+
		"    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_region_id_status` (`region_id`, `status`),\n"+
		"    key `idx_uk_vendor` (`vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4 comment ='云厂商支持的地区列表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "gcp_region", "create table if not exists `gcp_region`\n"+
		"(\n"+
		"    `id`          varchar(64) not null,\n"+
		"    `vendor`      varchar(16) not null,\n"+
		"    `region_id`   varchar(32) not null,\n"+
		"    `region_name` varchar(64) not null,\n"+
		"    `status`      varchar(32)          default '',\n"+
		"    `self_link`   varchar(255)         default '',\n"+
		"    `creator`     varchar(64)          default '',\n"+
		"    `reviser`     varchar(64)          default '',\n"+
		"    `created_at`  timestamp   not null default current_timestamp,\n"+
		"    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_region_id_status` (`region_id`, `status`),\n"+
		"    key `idx_uk_vendor` (`vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4 comment ='云厂商支持的地区列表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "image", "create table if not exists `image`\n"+
		"(\n"+
		"    `id`           varchar(64)  not null,\n"+
		"    `vendor`       varchar(16)  not null,\n"+
		"    `name`         varchar(128) not null,\n"+
		"    `cloud_id`     varchar(512) not null,\n"+
		"    `architecture` varchar(64)  not null,\n"+
		"    `platform`     varchar(128) not null,\n"+
		"    `state`        varchar(64)  not null,\n"+
		"    `type`         varchar(128) not null,\n"+
		"    `extension`    json         not null,\n"+
		"    `creator`      varchar(64)  not null,\n"+
		"    `reviser`      varchar(64)  not null,\n"+
		"    `created_at`   timestamp    not null default current_timestamp,\n"+
		"    `updated_at`   timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "eip", "create table if not exists `eip`\n"+
		"(\n"+
		"    `id`            varchar(64)  not null,\n"+
		"    `vendor`        varchar(16)  not null,\n"+
		"    `name`          varchar(128)          DEFAULT '',\n"+
		"    `account_id`    varchar(64)  not null,\n"+
		"    `cloud_id`      varchar(255) not null,\n"+
		"    `bk_biz_id`     bigint(1)    not null default -1,\n"+
		"    `region`        varchar(128) not null,\n"+
		"    `public_ip`     varchar(128) not null,\n"+
		"    `private_ip`    varchar(128) not null,\n"+
		"    `instance_id`   varchar(128) not null,\n"+
		"    `instance_type` varchar(64)  not null,\n"+
		"    `status`        varchar(128) not null,\n"+
		"    `extension`     json         not null,\n"+
		"    `creator`       varchar(64)  not null,\n"+
		"    `reviser`       varchar(64)  not null,\n"+
		"    `created_at`    timestamp    not null default current_timestamp,\n"+
		"    `updated_at`    timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "zone", "create table if not exists `zone`\n"+
		"(\n"+
		"    `id`         varchar(64)  not null,\n"+
		"    `vendor`     varchar(16)  not null,\n"+
		"    `cloud_id`   varchar(255) not null,\n"+
		"    `name`       varchar(64)  not null,\n"+
		"    `name_cn`    varchar(64)  not null,\n"+
		"    `state`      varchar(64)  not null,\n"+
		"    `region`     varchar(64)  not null,\n"+
		"    `extension`  json         not null,\n"+
		"    `creator`    varchar(64)  not null,\n"+
		"    `reviser`    varchar(64)  not null,\n"+
		"    `created_at` timestamp    not null default current_timestamp,\n"+
		"    `updated_at` timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "route_table", "create table if not exists `route_table`\n"+
		"(\n"+
		"    `id`           varchar(64)  not null,\n"+
		"    `vendor`       varchar(16)  not null,\n"+
		"    `account_id`   varchar(64)  not null,\n"+
		"    `cloud_id`     varchar(255) not null,\n"+
		"    `cloud_vpc_id` varchar(255) not null,\n"+
		"    `name`         varchar(128) not null,\n"+
		"    `region`       varchar(255) not null,\n"+
		"    `memo`         varchar(255)          default '',\n"+
		"    `vpc_id`       varchar(64)  not null,\n"+
		"    `bk_biz_id`    bigint(1)             default -1,\n"+
		"    \n"+
		"    `extension`    json         not null,\n"+
		"    \n"+
		"    `creator`      varchar(64)  not null,\n"+
		"    `reviser`      varchar(64)  not null,\n"+
		"    `created_at`   timestamp    not null default current_timestamp,\n"+
		"    `updated_at`   timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "tcloud_route", "create table if not exists `tcloud_route`\n"+
		"(\n"+
		"    `id`                          varchar(64)  not null,\n"+
		"    `cloud_id`                    varchar(64)  not null,\n"+
		"    `route_table_id`              varchar(64)  not null,\n"+
		"    `cloud_route_table_id`        varchar(64)  not null,\n"+
		"    `destination_cidr_block`      varchar(32)  not null,\n"+
		"    `destination_ipv6_cidr_block` varchar(64)           default '',\n"+
		"    `gateway_type`                varchar(32)  not null,\n"+
		"    `cloud_gateway_id`            varchar(255) not null,\n"+
		"    `enabled`                     boolean               default false,\n"+
		"    `route_type`                  varchar(32)  not null,\n"+
		"    `published_to_vbc`            boolean               default false,\n"+
		"    `memo`                        varchar(255)          default '',\n"+
		"    \n"+
		"    `creator`                     varchar(64)  not null,\n"+
		"    `reviser`                     varchar(64)  not null,\n"+
		"    `created_at`                  timestamp    not null default current_timestamp,\n"+
		"    `updated_at`                  timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id` (`cloud_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "aws_route", "create table if not exists `aws_route`\n"+
		"(\n"+
		"    `id`                                    varchar(64) not null,\n"+
		"    `route_table_id`                        varchar(64) not null,\n"+
		"    `cloud_route_table_id`                  varchar(64) not null,\n"+
		"    `destination_cidr_block`                varchar(32)          default null,\n"+
		"    `destination_ipv6_cidr_block`           varchar(64)          default null,\n"+
		"    `cloud_destination_prefix_list_id`      varchar(255)         default '',\n"+
		"    `cloud_carrier_gateway_id`              varchar(255)         default '',\n"+
		"    `core_network_arn`                      varchar(255)         default '',\n"+
		"    `cloud_egress_only_internet_gateway_id` varchar(255)         default '',\n"+
		"    `cloud_gateway_id`                      varchar(255)         default '',\n"+
		"    `cloud_instance_id`                     varchar(255)         default '',\n"+
		"    `cloud_instance_owner_id`               varchar(255)         default '',\n"+
		"    `cloud_local_gateway_id`                varchar(255)         default '',\n"+
		"    `cloud_nat_gateway_id`                  varchar(255)         default '',\n"+
		"    `cloud_network_interface_id`            varchar(255)         default '',\n"+
		"    `cloud_transit_gateway_id`              varchar(255)         default '',\n"+
		"    `cloud_vpc_peering_connection_id`       varchar(255)         default '',\n"+
		"    `state`                                 varchar(32) not null,\n"+
		"    `propagated`                            boolean              default false,\n"+
		"    \n"+
		"    `creator`                               varchar(64) not null,\n"+
		"    `reviser`                               varchar(64) not null,\n"+
		"    `created_at`                            timestamp   not null default current_timestamp,\n"+
		"    `updated_at`                            timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_route_table_id_destination_cidr_block` (`route_table_id`, `destination_cidr_block`),\n"+
		"    unique key `idx_uk_route_table_id_destination_ipv6_cidr_block` (`route_table_id`, `destination_ipv6_cidr_block`),\n"+
		"    unique key `idx_uk_route_table_id_cloud_dest_prefix_list_id` (`route_table_id`, `cloud_destination_prefix_list_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "azure_route", "create table if not exists `azure_route`\n"+
		"(\n"+
		"    `id`                   varchar(64)  not null,\n"+
		"    `cloud_id`             varchar(255) not null,\n"+
		"    `route_table_id`       varchar(64)  not null,\n"+
		"    `cloud_route_table_id` varchar(255) not null,\n"+
		"    `name`                 varchar(80)  not null,\n"+
		"    `address_prefix`       varchar(64)  not null,\n"+
		"    `next_hop_type`        varchar(32)  not null,\n"+
		"    `next_hop_ip_address`  varchar(255)          default '',\n"+
		"    `provisioning_state`   varchar(32)  not null,\n"+
		"    \n"+
		"    `creator`              varchar(64)  not null,\n"+
		"    `reviser`              varchar(64)  not null,\n"+
		"    `created_at`           timestamp    not null default current_timestamp,\n"+
		"    `updated_at`           timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_cloud_id` (`cloud_id`),\n"+
		"    unique key `idx_uk_route_table_id_name` (`route_table_id`, `name`),\n"+
		"    unique key `idx_uk_route_table_id_address_prefix` (`route_table_id`, `address_prefix`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "huawei_route", "create table if not exists `huawei_route`\n"+
		"(\n"+
		"    `id`                   varchar(64)  not null,\n"+
		"    `route_table_id`       varchar(64)  not null,\n"+
		"    `cloud_route_table_id` varchar(64)  not null,\n"+
		"    `type`                 varchar(32)  not null,\n"+
		"    `destination`          varchar(64)  not null,\n"+
		"    `nexthop`              varchar(255) not null,\n"+
		"    `memo`                 varchar(255)          default '',\n"+
		"    \n"+
		"    `creator`              varchar(64)  not null,\n"+
		"    `reviser`              varchar(64)  not null,\n"+
		"    `created_at`           timestamp    not null default current_timestamp,\n"+
		"    `updated_at`           timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_route_table_id_destination` (`route_table_id`, `destination`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "gcp_route", "create table if not exists `gcp_route`\n"+
		"(\n"+
		"    `id`                  varchar(64)  not null,\n"+
		"    `cloud_id`            varchar(64)  not null,\n"+
		"    `route_table_id`      varchar(64)  not null,\n"+
		"    `vpc_id`              varchar(64)  not null,\n"+
		"    `cloud_vpc_id`        varchar(255) not null,\n"+
		"    `self_link`           varchar(255) not null,\n"+
		"    `name`                varchar(128) not null,\n"+
		"    `dest_range`          varchar(64)  not null,\n"+
		"    `next_hop_gateway`    varchar(255)          default '',\n"+
		"    `next_hop_ilb`        varchar(255)          default '',\n"+
		"    `next_hop_instance`   varchar(255)          default '',\n"+
		"    `next_hop_ip`         varchar(255)          default '',\n"+
		"    `next_hop_network`    varchar(255)          default '',\n"+
		"    `next_hop_peering`    varchar(255)          default '',\n"+
		"    `next_hop_vpn_tunnel` varchar(255)          default '',\n"+
		"    `priority`            int(1) unsigned       default 0,\n"+
		"    `route_status`        varchar(32)  not null,\n"+
		"    `route_type`          varchar(32)  not null,\n"+
		"    `tags`                json         not null,\n"+
		"    `memo`                varchar(255)          default '',\n"+
		"    \n"+
		"    `creator`             varchar(64)  not null,\n"+
		"    `reviser`             varchar(64)  not null,\n"+
		"    `created_at`          timestamp    not null default current_timestamp,\n"+
		"    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id` (`cloud_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "cvm", "create table if not exists `cvm`\n"+
		"(\n"+
		"    `id`                     varchar(64)  not null,\n"+
		"    `cloud_id`               varchar(255) not null,\n"+
		"    `name`                   varchar(255) not null,\n"+
		"    `vendor`                 varchar(16)  not null,\n"+
		"    `bk_biz_id`              bigint(1)    not null default -1,\n"+
		"    `bk_cloud_id`            bigint(1)             default -1,\n"+
		"    `account_id`             varchar(64)  not null,\n"+
		"    `region`                 varchar(20)  not null,\n"+
		"    `zone`                   varchar(20)           default '',\n"+
		"    `cloud_vpc_ids`          json         not null,\n"+
		"    `vpc_ids`                json         not null,\n"+
		"    `cloud_subnet_ids`       json         not null,\n"+
		"    `subnet_ids`             json         not null,\n"+
		"    `cloud_image_id`         varchar(255) not null,\n"+
		"    `image_id`               varchar(64)  not null,\n"+
		"    `os_name`                varchar(255) not null,\n"+
		"    `memo`                   varchar(255)          default '',\n"+
		"    `status`                 varchar(50)  not null,\n"+
		"    `recycle_status`         varchar(32)           default '',\n"+
		"    `private_ipv4_addresses` json                  default null,\n"+
		"    `private_ipv6_addresses` json                  default null,\n"+
		"    `public_ipv4_addresses`  json                  default null,\n"+
		"    `public_ipv6_addresses`  json                  default null,\n"+
		"    `machine_type`           varchar(50)  not null,\n"+
		"    `extension`              json         not null,\n"+
		"    `cloud_created_time`     varchar(50)           default '',\n"+
		"    `cloud_launched_time`    varchar(50)           default '',\n"+
		"    `cloud_expired_time`     varchar(50)           default '',\n"+
		"    `creator`                varchar(64)  not null,\n"+
		"    `reviser`                varchar(64)  not null,\n"+
		"    `created_at`             timestamp    not null default current_timestamp,\n"+
		"    `updated_at`             timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "application", "create table if not exists `application`\n"+
		"(\n"+
		"    `id`              varchar(64) not null,\n"+
		"    `sn`              varchar(64) not null,\n"+
		"    `type`            varchar(64) not null,\n"+
		"    `status`          varchar(32) not null,\n"+
		"    `applicant`       varchar(64) not null,\n"+
		"    `memo`            varchar(255)         default '',\n"+
		"    `content`         json        not null,\n"+
		"    `delivery_detail` json        not null,\n"+
		"    `creator`         varchar(64) not null,\n"+
		"    `reviser`         varchar(64) not null,\n"+
		"    `created_at`      timestamp   not null default current_timestamp,\n"+
		"    `updated_at`      timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_sn` (`sn`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "approval_process", "create table if not exists `approval_process`\n"+
		"(\n"+
		"    `id`               varchar(64) not null,\n"+
		"    `application_type` varchar(64) not null,\n"+
		"    `service_id`       bigint(1)   not null,\n"+
		"    `creator`          varchar(64) not null,\n"+
		"    `reviser`          varchar(64) not null,\n"+
		"    `created_at`       timestamp   not null default current_timestamp,\n"+
		"    `updated_at`       timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_type` (`application_type`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "disk_cvm_rel", "create table if not exists `disk_cvm_rel`\n"+
		"(\n"+
		"    `id`         bigint(1) unsigned not null auto_increment,\n"+
		"    `disk_id`    varchar(64)        not null,\n"+
		"    `cvm_id`     varchar(64)        not null,\n"+
		"    `creator`    varchar(64)        not null,\n"+
		"    `created_at` timestamp          not null default current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_disk_id_cvm_id` (`disk_id`, `cvm_id`),\n"+
		"    constraint disk_cvm_rel_cvm_id foreign key (disk_id) REFERENCES disk (id) ON DELETE CASCADE,\n"+
		"    constraint disk_cvm_rel_disk_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "security_group_cvm_rel", "create table if not exists `security_group_cvm_rel`\n"+
		"(\n"+
		"    `id`                bigint(1) unsigned not null auto_increment,\n"+
		"    `security_group_id` varchar(64)        not null,\n"+
		"    `cvm_id`            varchar(64)        not null,\n"+
		"    `creator`           varchar(64)        not null,\n"+
		"    `created_at`        timestamp          not null default current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_security_group_id_cvm_id` (`security_group_id`, `cvm_id`),\n"+
		"    constraint security_group_cvm_rel_security_group_id foreign key (security_group_id) REFERENCES security_group (id) ON DELETE CASCADE,\n"+
		"    constraint security_group_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "eip_cvm_rel", "create table if not exists `eip_cvm_rel`\n"+
		"(\n"+
		"    `id`         bigint(1) unsigned not null auto_increment,\n"+
		"    `eip_id`     varchar(64)        not null,\n"+
		"    `cvm_id`     varchar(64)        not null,\n"+
		"    `creator`    varchar(64)        not null,\n"+
		"    `created_at` timestamp          not null default current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_eip_id_cvm_id` (`eip_id`, `cvm_id`),\n"+
		"    constraint eip_cvm_rel_eip_id foreign key (eip_id) REFERENCES eip (id) ON DELETE CASCADE,\n"+
		"    constraint eip_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "network_interface", "create table if not exists `network_interface`\n"+
		"(\n"+
		"    `id`              varchar(64)  not null,\n"+
		"    `account_id`      varchar(64)  not null,\n"+
		"    `vendor`          varchar(16)  not null default '',\n"+
		"    `name`            varchar(64)  not null,\n"+
		"    `region`          varchar(255) not null default '',\n"+
		"    `zone`            varchar(255) not null default '',\n"+
		"    `cloud_id`        varchar(255)          default '',\n"+
		"    `vpc_id`          varchar(255) not null default '',\n"+
		"    `cloud_vpc_id`    varchar(255) not null default '',\n"+
		"    `subnet_id`       varchar(64)  not null default '',\n"+
		"    `cloud_subnet_id` varchar(255)          default '',\n"+
		"    `private_ipv4`    json                  default null,\n"+
		"    `private_ipv6`    json                  default null,\n"+
		"    `public_ipv4`     json                  default null,\n"+
		"    `public_ipv6`     json                  default null,\n"+
		"    `bk_biz_id`       bigint                default '-1',\n"+
		"    `instance_id`     varchar(255)          default '',\n"+
		"    `extension`       json                  default null,\n"+
		"    `creator`         varchar(64)           default '',\n"+
		"    `reviser`         varchar(64)           default '',\n"+
		"    `created_at`      timestamp    not null default current_timestamp,\n"+
		"    `updated_at`      timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "network_interface_cvm_rel", "create table if not exists `network_interface_cvm_rel`\n"+
		"(\n"+
		"    `id`                   bigint unsigned not null auto_increment,\n"+
		"    `cvm_id`               varchar(64)     not null,\n"+
		"    `network_interface_id` varchar(64)     not null,\n"+
		"    `creator`              varchar(64)          default '',\n"+
		"    `created_at`           timestamp       null default current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cvm_id_network_interface_id` (`cvm_id`, `network_interface_id`),\n"+
		"    constraint network_interface_cvm_rel_network_id foreign key (network_interface_id) REFERENCES network_interface (id) ON DELETE CASCADE,\n"+
		"    constraint network_interface_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "recycle_record", "create table if not exists `recycle_record`\n"+
		"(\n"+
		"    `id`           bigint(1) unsigned not null auto_increment,\n"+
		"    `task_id`      varchar(64)        not null,\n"+
		"    `vendor`       varchar(32)        not null,\n"+
		"    `res_type`     varchar(64)        not null,\n"+
		"    `res_id`       varchar(64)        not null,\n"+
		"    `cloud_res_id` varchar(255)       not null,\n"+
		"    `res_name`     varchar(255)                default '',\n"+
		"    `bk_biz_id`    bigint(1)          not null,\n"+
		"    `account_id`   varchar(64)        not null,\n"+
		"    `region`       varchar(255)       not null,\n"+
		"    `detail`       json               not null,\n"+
		"    `status`       varchar(32)        not null,\n"+
		"    `creator`      varchar(64)        not null,\n"+
		"    `reviser`      varchar(64)        not null,\n"+
		"    `created_at`   timestamp          not null default current_timestamp,\n"+
		"    `updated_at`   timestamp          not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_res_type_res_id` (`res_type`, `res_id`),\n"+
		"    unique key `idx_res_type_vendor_cloud_res_id` (`res_type`, `vendor`, `cloud_res_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	return nil
}
