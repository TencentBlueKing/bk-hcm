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

// Package migration is the migration task_management.
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
const ID = "20241128-0000-TASK-MANAGEMENT-023A"

func init() {
	register.Main.Regist(ID, "v1.7.0", "20241128000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0028_20241128_task_management.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "task_management", "create table if not exists `task_management`\n"+
		"(\n"+
		"    `id`          varchar(64) not null,\n"+
		"    `bk_biz_id`   bigint      not null,\n"+
		"    `source`      varchar(16) not null,\n"+
		"    `vendors`     json        not null,\n"+
		"    `state`       varchar(16) not null,\n"+
		"    `account_ids` json        not null,\n"+
		"    `resource`    varchar(16) not null,\n"+
		"    `operations`  json        not null,\n"+
		"    `flow_ids`    json                 default NULL,\n"+
		"    `extension`   json                 default NULL,\n"+
		"    `creator`     varchar(64) not null,\n"+
		"    `reviser`     varchar(64) not null,\n"+
		"    `created_at`  timestamp   not null default current_timestamp,\n"+
		"    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    index `idx_state` (`state`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='任务管理表'"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "task_detail", "create table if not exists `task_detail`\n"+
		"(\n"+
		"    `id`                 varchar(64) not null,\n"+
		"    `bk_biz_id`          bigint      not null,\n"+
		"    `task_management_id` varchar(64) not null,\n"+
		"    `flow_id`            varchar(64)          default '',\n"+
		"    `task_action_ids`    json                 default NULL,\n"+
		"    `operation`          varchar(64) not null,\n"+
		"    `param`              json        not null,\n"+
		"    `state`              varchar(16) not null,\n"+
		"    `reason`             varchar(1024)        default '',\n"+
		"    `extension`          json                 default NULL,\n"+
		"    `creator`            varchar(64) not null,\n"+
		"    `reviser`            varchar(64) not null,\n"+
		"    `result`             json                 default NULL,\n"+
		"    `created_at`         timestamp   not null default current_timestamp,\n"+
		"    `updated_at`         timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    index `idx_task_management_id` (`task_management_id`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='任务详情表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "task_management", "0"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "task_detail", "0"); err != nil {
		return err
	}
	return nil
}
