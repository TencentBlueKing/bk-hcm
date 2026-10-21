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

// Package migration is the migration account.
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
const ID = "20240603-1700-ACCOUNT-6FE2"

func init() {
	register.Main.Regist(ID, "v1.6.0", "20240603170000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0022_20240603_1700_account.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "main_account", "create table if not exists `main_account`\n"+
		"(\n"+
		"    `id`                    varchar(64)     not null,\n"+
		"    `name`                  varchar(255)    not null,\n"+
		"    `vendor`                varchar(16)     not null,\n"+
		"    `cloud_id`              varchar(64)     not null,\n"+
		"    `email`                 varchar(255)    not null,\n"+
		"    `managers`              json            not null,\n"+
		"    `bak_managers`          json            not null,\n"+
		"    `site`                  varchar(32)     not null,\n"+
		"    `business_type`         varchar(64)     not null,\n"+
		"    `status`                varchar(32)     not null,\n"+
		"    `parent_account_name`   varchar(255)    not null,\n"+
		"    `parent_account_id`     varchar(64)     not null,\n"+
		"    `dept_id`               bigint(1)       not null,\n"+
		"    `bk_biz_id`             bigint(1)       not null,\n"+
		"    `op_product_id`         bigint(1)       not null,\n"+
		"    `memo`                  varchar(512)            default '',\n"+
		"    `extension`             json            not null,\n"+
		"    `creator`        varchar(64) not null,\n"+
		"    `reviser`        varchar(64) not null,\n"+
		"    `created_at`     timestamp   not null default current_timestamp,\n"+
		"    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key(`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "main_account", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "root_account", "create table if not exists `root_account`\n"+
		"(\n"+
		"    `id`                    varchar(64)     not null,\n"+
		"    `name`                  varchar(64)     not null,\n"+
		"    `vendor`                varchar(16)     not null,\n"+
		"    `cloud_id`              varchar(64)     not null,\n"+
		"    `email`                 varchar(255)    not null,\n"+
		"    `managers`              json            not null,\n"+
		"    `bak_managers`          json            not null,\n"+
		"    `site`                  varchar(32)     not null,\n"+
		"    `dept_id`               bigint(1)       not null,\n"+
		"    `memo`                  varchar(512)            default '',\n"+
		"    `extension`             json            not null,\n"+
		"    `creator`        varchar(64) not null,\n"+
		"    `reviser`        varchar(64) not null,\n"+
		"    `created_at`     timestamp   not null default current_timestamp,\n"+
		"    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key(`id`),\n"+
		"    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "root_account", "0"); err != nil {
		return err
	}
	return nil
}
