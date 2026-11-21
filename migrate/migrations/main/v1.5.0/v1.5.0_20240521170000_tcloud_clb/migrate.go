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

// Package migration is the migration tcloud_clb.
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
const ID = "20240521-1700-TCLOUD-CLB-387A"

func init() {
	register.Main.Regist(ID, "v1.5.0", "20240521170000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0018_20240521_1700_tcloud_clb.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "load_balancer", "create table `load_balancer`\n"+
		"(\n"+
		"    `id`                     varchar(64)  not null,\n"+
		"    `cloud_id`               varchar(255) not null,\n"+
		"    `name`                   varchar(255) not null,\n"+
		"    `vendor`                 varchar(16)  not null,\n"+
		"\n"+
		"    `bk_biz_id`              bigint       not null default -1,\n"+
		"    `account_id`             varchar(64)  not null,\n"+
		"\n"+
		"    `region`                 varchar(20)  not null,\n"+
		"    `zones`                  json         not null,\n"+
		"    `backup_zones`           json                  default null,\n"+
		"    `lb_type`                varchar(64)  not null,\n"+
		"    `ip_version`             varchar(64)  not null default '',\n"+
		"\n"+
		"    `vpc_id`                 varchar(255) not null,\n"+
		"    `cloud_vpc_id`           varchar(255) not null,\n"+
		"    `cloud_subnet_id`        varchar(255) not null,\n"+
		"    `subnet_id`              varchar(255) not null,\n"+
		"\n"+
		"    `private_ipv4_addresses` json                  default null,\n"+
		"    `private_ipv6_addresses` json                  default null,\n"+
		"    `public_ipv4_addresses`  json                  default null,\n"+
		"    `public_ipv6_addresses`  json                  default null,\n"+
		"\n"+
		"    `domain`                 varchar(255) not null,\n"+
		"    `status`                 varchar(64)  not null,\n"+
		"    `memo`                   varchar(255)          default '',\n"+
		"    `cloud_created_time`     varchar(64)           default '',\n"+
		"    `cloud_status_time`      varchar(64)           default '',\n"+
		"    `cloud_expired_time`     varchar(64)           default '',\n"+
		"    `extension`              json         not null,\n"+
		"\n"+
		"    `creator`                varchar(64)  not null,\n"+
		"    `reviser`                varchar(64)  not null,\n"+
		"    `created_at`             timestamp    not null default current_timestamp,\n"+
		"    `updated_at`             timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='负载均衡表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "security_group_common_rel", "create table `security_group_common_rel`\n"+
		"(\n"+
		"    `id`                bigint unsigned not null auto_increment,\n"+
		"    `vendor`            varchar(16)     not null,\n"+
		"    `res_id`            varchar(64)     not null,\n"+
		"    `res_type`          varchar(64)     not null,\n"+
		"    `security_group_id` varchar(64)     not null,\n"+
		"    `priority`          int             not null,\n"+
		"\n"+
		"    `creator`           varchar(64)     not null,\n"+
		"    `reviser`           varchar(64)     not null,\n"+
		"    `created_at`        timestamp       not null default current_timestamp,\n"+
		"    `updated_at`        timestamp       not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_vendor_res_type_res_id_sg_id`\n"+
		"        (`vendor`, `res_type`, `res_id`, `security_group_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='通用安全组资源关联表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "load_balancer_listener", "create table `load_balancer_listener`\n"+
		"(\n"+
		"    `id`             varchar(64)  not null,\n"+
		"    `cloud_id`       varchar(255) not null,\n"+
		"    `name`           varchar(255) not null,\n"+
		"    `vendor`         varchar(16)  not null,\n"+
		"\n"+
		"    `account_id`     varchar(64)  not null,\n"+
		"    `bk_biz_id`      bigint(1)    not null default -1,\n"+
		"\n"+
		"    `lb_id`          varchar(255) not null,\n"+
		"    `cloud_lb_id`    varchar(255) not null,\n"+
		"    `protocol`       varchar(64)  not null,\n"+
		"    `port`           bigint       not null,\n"+
		"    `default_domain` varchar(255)          default null,\n"+
		"    `zones`          json,\n"+
		"    `sni_switch`     bigint                default 0,\n"+
		"    `extension`      json                  default null,\n"+
		"    `memo`           varchar(255)          default '',\n"+
		"\n"+
		"    `creator`        varchar(64)  not null,\n"+
		"    `reviser`        varchar(64)  not null,\n"+
		"    `created_at`     timestamp    not null default current_timestamp,\n"+
		"    `updated_at`     timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`),\n"+
		"    key `idx_lb_id`(`lb_id`)\n"+
		"\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='负载均衡监听器'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "tcloud_lb_url_rule", "create table `tcloud_lb_url_rule`\n"+
		"(\n"+
		"    `id`                    varchar(64)  not null,\n"+
		"    `cloud_id`              varchar(255) not null,\n"+
		"    `name`                  varchar(255) not null,\n"+
		"    `rule_type`             varchar(64)  not null,\n"+
		"\n"+
		"    `lb_id`                 varchar(255) not null,\n"+
		"    `cloud_lb_id`           varchar(255) not null,\n"+
		"    `lbl_id`                varchar(255) not null,\n"+
		"    `cloud_lbl_id`          varchar(255) not null,\n"+
		"    `target_group_id`       varchar(255)          default '',\n"+
		"    `cloud_target_group_id` varchar(255)          default '',\n"+
		"\n"+
		"    `domain`                varchar(255)          default '',\n"+
		"    `url`                   varchar(255)          default '',\n"+
		"    `scheduler`             varchar(64)  not null,\n"+
		"    `session_type`          varchar(64)           default '',\n"+
		"    `session_expire`        bigint                default 0,\n"+
		"    `health_check`          json                  default null,\n"+
		"    `certificate`           json                  default null,\n"+
		"    `memo`                  varchar(255)          default '',\n"+
		"\n"+
		"\n"+
		"    `creator`               varchar(64)  not null,\n"+
		"    `reviser`               varchar(64)  not null,\n"+
		"    `created_at`            timestamp    not null default current_timestamp,\n"+
		"    `updated_at`            timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_lbl_id` (`cloud_id`, `lbl_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='负载均衡七层规则'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "load_balancer_target", "create table `load_balancer_target`\n"+
		"(\n"+
		"    `id`                    varchar(64)  not null,\n"+
		"    `account_id`            varchar(64)  not null,\n"+
		"\n"+
		"    `inst_type`             varchar(255) not null,\n"+
		"    `inst_id`               varchar(255) not null,\n"+
		"    `cloud_inst_id`         varchar(255) not null,\n"+
		"    `inst_name`             varchar(255) not null,\n"+
		"\n"+
		"    `target_group_id`       varchar(255)          default '',\n"+
		"    `cloud_target_group_id` varchar(255)          default '',\n"+
		"\n"+
		"    `port`                  bigint       not null,\n"+
		"    `weight`                bigint       not null,\n"+
		"    `private_ip_address`    json         not null,\n"+
		"    `public_ip_address`     json         not null,\n"+
		"    `cloud_vpc_ids`         json                  default null,\n"+
		"    `zone`                  varchar(255) not null default '',\n"+
		"    `memo`                  varchar(255)          default '',\n"+
		"\n"+
		"    `creator`               varchar(64)  not null,\n"+
		"    `reviser`               varchar(64)  not null,\n"+
		"    `created_at`            timestamp    not null default current_timestamp,\n"+
		"    `updated_at`            timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_target_group_id_cloud_inst_id_port` (`cloud_target_group_id`, `cloud_inst_id`, `port`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='负载均衡目标'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "load_balancer_target_group", "create table `load_balancer_target_group`\n"+
		"(\n"+
		"    `id`                varchar(64)  not null,\n"+
		"    `cloud_id`          varchar(255) not null,\n"+
		"    `name`              varchar(255) not null,\n"+
		"    `vendor`            varchar(16)  not null,\n"+
		"    `target_group_type` varchar(16)  not null,\n"+
		"\n"+
		"    `account_id`        varchar(64)  not null,\n"+
		"    `bk_biz_id`         bigint(1)    not null default -1,\n"+
		"    `vpc_id`            varchar(255) not null,\n"+
		"    `cloud_vpc_id`      varchar(255) not null,\n"+
		"\n"+
		"    `region`            varchar(20)  not null,\n"+
		"    `protocol`          varchar(64)  not null,\n"+
		"    `port`              bigint       not null,\n"+
		"    `weight`            bigint       not null default -1,\n"+
		"    `health_check`      json                  default null,\n"+
		"\n"+
		"    `memo`              varchar(255)          default '',\n"+
		"    `extension`         json         not null default ('{}'),\n"+
		"\n"+
		"    `creator`           varchar(64)  not null,\n"+
		"    `reviser`           varchar(64)  not null,\n"+
		"    `created_at`        timestamp    not null default current_timestamp,\n"+
		"    `updated_at`        timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='负载均衡目标组'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "target_group_listener_rule_rel", "create table `target_group_listener_rule_rel`\n"+
		"(\n"+
		"    `id`                     varchar(64) not null,\n"+
		"    `listener_rule_id`       varchar(64) not null,\n"+
		"    `listener_rule_type`     varchar(64) not null,\n"+
		"    `cloud_listener_rule_id` varchar(64) not null,\n"+
		"    `target_group_id`        varchar(64) not null,\n"+
		"    `cloud_target_group_id`  varchar(64) not null,\n"+
		"    `lb_id`                  varchar(64) not null,\n"+
		"    `cloud_lb_id`            varchar(64) not null,\n"+
		"    `lbl_id`                 varchar(64) not null,\n"+
		"    `cloud_lbl_id`           varchar(64) not null,\n"+
		"    `binding_status`         varchar(64) not null,\n"+
		"    `detail`                 json                 default null,\n"+
		"\n"+
		"    `creator`                varchar(64) not null,\n"+
		"    `reviser`                varchar(64) not null,\n"+
		"    `created_at`             timestamp   not null default current_timestamp,\n"+
		"    `updated_at`             timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_target_group_id_listener_rule_id_listener_rule_type` (`target_group_id`, `listener_rule_id`, `listener_rule_type`),\n"+
		"    key `idx_lbid_binding_status_rule_type`(`lb_id`, `binding_status`, `listener_rule_type`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='目标组监听器关系表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "resource_flow_rel", "create table `resource_flow_rel`\n"+
		"(\n"+
		"    `id`         varchar(64) not null,\n"+
		"    `res_id`     varchar(64) not null,\n"+
		"    `res_type`   varchar(64) not null,\n"+
		"    `flow_id`    varchar(64) not null,\n"+
		"    `task_type`  varchar(64) not null,\n"+
		"    `status`     varchar(64) not null,\n"+
		"\n"+
		"    `creator`    varchar(64) not null,\n"+
		"    `reviser`    varchar(64) not null,\n"+
		"    `created_at` timestamp   not null default current_timestamp,\n"+
		"    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_res_id_flow_id` (`res_id`, `res_type`, `flow_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='资源与异步任务的关系表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "resource_flow_lock", "create table `resource_flow_lock`\n"+
		"(\n"+
		"    `res_type`   varchar(64) not null,\n"+
		"    `res_id`     varchar(64) not null,\n"+
		"    `owner`      varchar(64) not null,\n"+
		"\n"+
		"    `creator`    varchar(64) not null,\n"+
		"    `reviser`    varchar(64) not null,\n"+
		"    `created_at` timestamp   not null default current_timestamp,\n"+
		"    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`res_type`, `res_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate = utf8mb4_bin comment ='资源与异步任务锁定的表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "load_balancer", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "security_group_common_rel", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "load_balancer_listener", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "tcloud_lb_url_rule", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "load_balancer_target", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "load_balancer_target_group", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "target_group_listener_rule_rel", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "resource_flow_rel", "0"); err != nil {
		return err
	}
	return nil
}
