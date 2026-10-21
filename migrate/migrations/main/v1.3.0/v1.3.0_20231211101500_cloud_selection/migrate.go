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

// Package migration is the migration cloud_selection.
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
const ID = "20231211-1015-CLOUD-SELECTION-4B0B"

func init() {
	register.Main.Regist(ID, "v1.3.0", "20231211101500", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0013_20231211_1015.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "cloud_selection_scheme", "create table if not exists `cloud_selection_scheme`\n"+
		"(\n"+
		"    `id`                      varchar(64)  not null,\n"+
		"    `bk_biz_id`               bigint       not null,\n"+
		"    `name`                    varchar(255) not null,\n"+
		"    `biz_type`                varchar(64)  not null,\n"+
		"    `vendors`                 json         not null,\n"+
		"    `deployment_architecture` json         not null,\n"+
		"    `cover_ping`              double       not null,\n"+
		"    `composite_score`         double       not null,\n"+
		"    `net_score`               double       not null,\n"+
		"    `cost_score`              double       not null,\n"+
		"    `cover_rate`              double       not null,\n"+
		"    `user_distribution`       json         not null,\n"+
		"    `result_idc_ids`          json         not null,\n"+
		"    `creator`                 varchar(64)  not null,\n"+
		"    `reviser`                 varchar(64)  not null,\n"+
		"    `created_at`              timestamp    not null default current_timestamp,\n"+
		"    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_bk_biz_id_name` (`bk_biz_id`, `name`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='云选型方案表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "cloud_selection_scheme", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "cloud_selection_biz_type", "create table if not exists `cloud_selection_biz_type`\n"+
		"(\n"+
		"    `id`                      varchar(64) not null,\n"+
		"    `biz_type`                varchar(64) not null,\n"+
		"    `cover_ping`              double      not null,\n"+
		"    `deployment_architecture` json        not null,\n"+
		"    `creator`                 varchar(64) not null,\n"+
		"    `reviser`                 varchar(64) not null,\n"+
		"    `created_at`              timestamp   not null default current_timestamp,\n"+
		"    `updated_at`              timestamp   not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (`id`),\n"+
		"    unique key `idx_uk_biz_type` (`biz_type`)\n"+
		") engine = innodb\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='云选型业务类型表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "cloud_selection_biz_type", "0"); err != nil {
		return err
	}
	if _, err := util.CreateTableIfNotExists(ctx, o, "cloud_selection_idc", "create table if not exists `cloud_selection_idc`\n"+
		"(\n"+
		"    `id`         varchar(64)  not null,\n"+
		"    `bk_biz_id`  bigint       not null,\n"+
		"    `name`       varchar(255) not null,\n"+
		"    `vendor`     varchar(16)  not null,\n"+
		"    `country`    varchar(255) not null,\n"+
		"    `region`     varchar(255) not null,\n"+
		"    `creator`    varchar(64)  not null,\n"+
		"    `reviser`    varchar(64)  not null,\n"+
		"    `created_at` timestamp    not null default current_timestamp,\n"+
		"    `updated_at` timestamp    not null default current_timestamp on update current_timestamp,\n"+
		"    primary key (id),\n"+
		"    unique key `idx_uk_bk_biz_id_name` (`bk_biz_id`, `name`)\n"+
		") engine = InnoDB\n"+
		"  default charset = utf8mb4\n"+
		"  collate utf8mb4_bin comment ='云选型IDC信息表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "cloud_selection_idc", "0"); err != nil {
		return err
	}
	return nil
}
