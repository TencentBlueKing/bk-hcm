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

// Package migration is the migration account_bill_config.
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
const ID = "20230516-1600-ACCOUNT-BILL-CONFIG-545B"

func init() {
	register.Main.Regist(ID, "v1.1.3", "20230516160000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0003_20230516_1600.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.InsertIDGenerator(ctx, o, "account_bill_config", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_bill_config", "CREATE TABLE `account_bill_config`\n"+
		"(\n"+
		"    `id`                  varchar(64) not null,\n"+
		"    `vendor`              varchar(16) not null default '',\n"+
		"    `account_id`          varchar(64) not null,\n"+
		"    `cloud_database_name` varchar(64)          default '',\n"+
		"    `cloud_table_name`    varchar(64)          default '',\n"+
		"    `status`              tinyint unsigned default '0',\n"+
		"    `err_msg`             json                 default NULL,\n"+
		"    `extension`           json                 default NULL,\n"+
		"    `creator`             varchar(64)          default '',\n"+
		"    `reviser`             varchar(64)          default '',\n"+
		"    `created_at`          timestamp   not null default current_timestamp,\n"+
		"    `updated_at`          timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_vendor_account_id` (`vendor`, `account_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "azure_security_group_rule", "idx_uk_name"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_security_group_rule", "idx_uk_name_cloud_security_group_id", []string{"name", "cloud_security_group_id"}, true); err != nil {
		return err
	}
	return nil
}
