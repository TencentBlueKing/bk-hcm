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

// Package migration is the migration argument_template.
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
const ID = "20240305-1000-ARGUMENT-TEMPLATE-F96D"

func init() {
	register.Main.Regist(ID, "v1.4.0", "20240305100000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0015_20240305_1000_argument_template.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "argument_template", "create table if not exists `argument_template`\n"+
		"(\n"+
		"    `id`                  varchar(64)  not null,\n"+
		"    `cloud_id`            varchar(255) not null,\n"+
		"    `name`                varchar(255) not null,\n"+
		"    `vendor`              varchar(16)  not null,\n"+
		"    `bk_biz_id`           bigint       not null default '-1',\n"+
		"    `account_id`          varchar(64)  not null,\n"+
		"    `type`                varchar(16)  not null,\n"+
		"    `templates`           json         not null,\n"+
		"    `group_templates`     json         not null,\n"+
		"    `memo`                varchar(255)          default '',\n"+
		"    `creator`             varchar(64)  not null,\n"+
		"    `reviser`             varchar(64)  not null,\n"+
		"    `created_at`          timestamp    not null default current_timestamp,\n"+
		"    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_bk_biz_id_cloud_id` (`bk_biz_id`, `cloud_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='参数模版表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "argument_template", "0"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_security_group_rule",
		Column:  "service_id",
		Type:    "varchar(64)",
		Default: util.ExprDefault("null"),
		After:   "port",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_security_group_rule",
		Column:  "service_group_id",
		Type:    "varchar(64)",
		Default: util.ExprDefault("null"),
		After:   "cloud_service_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_security_group_rule",
		Column:  "address_id",
		Type:    "varchar(64)",
		Default: util.ExprDefault("null"),
		After:   "cloud_target_security_group_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_security_group_rule",
		Column:  "address_group_id",
		Type:    "varchar(64)",
		Default: util.ExprDefault("null"),
		After:   "cloud_address_id",
	}); err != nil {
		return err
	}
	return nil
}
