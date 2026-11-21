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

// Package migration is the migration account_secret.
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
const ID = "20260130-1800-ACCOUNT-SECRET-AFA4"

func init() {
	register.Main.Regist(ID, "v1.9.0", "20260130180000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0047_20260130_1800_account_secret.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "account_secret", "CREATE TABLE IF NOT EXISTS `account_secret` (\n"+
		"    `id` varchar(64) NOT NULL COMMENT '密钥ID',\n"+
		"    `account_id` varchar(64) NOT NULL COMMENT '账号ID',\n"+
		"    `vendor` varchar(16) NOT NULL COMMENT '云厂商',\n"+
		"    `type` varchar(16) NOT NULL COMMENT '密钥类型',\n"+
		"    `status` varchar(16) NOT NULL COMMENT '密钥状态',\n"+
		"    `extension` json NOT NULL COMMENT '云厂商差异扩展字段',\n"+
		"    `tenant_id` varchar(64) NOT NULL COMMENT '租户ID' default 'default',\n"+
		"    `creator` varchar(64) NOT NULL COMMENT '创建者',\n"+
		"    `reviser` varchar(64) NOT NULL COMMENT '更新者',\n"+
		"    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n"+
		"    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB\n"+
		"  DEFAULT CHARSET=utf8mb4\n"+
		"  COLLATE=utf8mb4_bin COMMENT='账号密钥表'"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "account_secret", "idx_account_id", []string{"account_id"}, false); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "account_secret", "0"); err != nil {
		return err
	}
	return nil
}
