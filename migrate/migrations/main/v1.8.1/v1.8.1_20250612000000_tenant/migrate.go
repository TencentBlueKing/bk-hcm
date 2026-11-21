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

// Package migration is the migration tenant.
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
const ID = "20250612-0000-TENANT-C547"

func init() {
	register.Main.Regist(ID, "v1.8.1", "20250612000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0033_20250612_tenant.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "tenant", "create table if not exists `tenant` (\n"+
		"    `id` varchar(64) NOT NULL COMMENT '唯一ID',\n"+
		"    `tenant_id` varchar(64) NOT NULL COMMENT '租户ID',\n"+
		"    `status` varchar(64) NOT NULL DEFAULT 'enable' COMMENT '租户状态(enable:启用 disable:禁用)',\n"+
		"    `creator` varchar(64) NOT NULL COMMENT '创建人',\n"+
		"    `reviser` varchar(64) NOT NULL COMMENT '修改人',\n"+
		"    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '该记录创建的时间',\n"+
		"    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n"+
		"    PRIMARY KEY (`id`),\n"+
		"    UNIQUE KEY `idx_uk_tenant_id` (`tenant_id`),\n"+
		"    KEY `idx_status` (`status`)\n"+
		") ENGINE=InnoDB\n"+
		"  DEFAULT CHARSET=utf8mb4\n"+
		"  COLLATE=utf8mb4_bin COMMENT='租户表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "tenant", "0"); err != nil {
		return err
	}
	return nil
}
