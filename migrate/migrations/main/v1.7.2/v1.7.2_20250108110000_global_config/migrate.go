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

// Package migration is the migration global_config.
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
const ID = "20250108-1100-GLOBAL-CONFIG-8E44"

func init() {
	register.Main.Regist(ID, "v1.7.2", "20250108110000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0030_20250108_1100_global_config.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "global_config", "create table if not exists `global_config`\n"+
		"(\n"+
		"    `id`           varchar(64) not null comment '主键',\n"+
		"    `config_key`   varchar(64) not null comment 'key',\n"+
		"    `config_value` json        not null comment 'value',\n"+
		"    `config_type`  varchar(64) not null comment '类型',\n"+
		"    `memo`         varchar(255)         default '' comment '备注',\n"+
		"    `creator`      varchar(64) not null comment '创建者',\n"+
		"    `reviser`      varchar(64) not null comment '更新者',\n"+
		"    `created_at`   timestamp   not null default current_timestamp comment '创建时间',\n"+
		"    `updated_at`   timestamp   not null default current_timestamp on update current_timestamp comment '更新时间',\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_global_config_id` (`config_type`, `config_key`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='全局配置表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "global_config", "0"); err != nil {
		return err
	}
	return nil
}
