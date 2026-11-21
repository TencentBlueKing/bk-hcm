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

// Package migration is the migration permission_template.
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
const ID = "20260324-1001-PERMISSION-TEMPLATE-4304"

func init() {
	register.Main.Regist(ID, "v1.9.0", "20260324100100", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0051_20260324_1001_permission_template.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "permission_template", "CREATE TABLE IF NOT EXISTS `permission_template` (\n"+
		"    `id`                       varchar(64)   NOT NULL                              COMMENT '本地模板ID',\n"+
		"    `cloud_id`                 varchar(64)   NOT NULL                              COMMENT '云上策略ID',\n"+
		"    `name`                     varchar(128)  NOT NULL                              COMMENT '模板名称',\n"+
		"    `account_id`               varchar(64)   NOT NULL                              COMMENT '所属二级账号ID',\n"+
		"    `policy_library_id`        varchar(64)   DEFAULT NULL                          COMMENT '来源权限策略库ID',\n"+
		"    `policy_library_version`   int           DEFAULT NULL                          COMMENT '权限策略库版本',\n"+
		"    `policy_library_sync_time` timestamp     NULL DEFAULT NULL                     COMMENT '权限策略库同步时间',\n"+
		"    `policy_document`          longtext      NOT NULL                              COMMENT '策略JSON内容',\n"+
		"    `policy_hash`              varchar(64)   NOT NULL                              COMMENT '策略内容哈希值',\n"+
		"    `memo`                     varchar(255)  DEFAULT NULL                          COMMENT '描述',\n"+
		"    `extension`                json          DEFAULT NULL                          COMMENT '云厂商差异扩展字段',\n"+
		"    `vendor`                   varchar(16)   NOT NULL                              COMMENT '云厂商',\n"+
		"    `tenant_id`                varchar(64)   NOT NULL DEFAULT 'default'            COMMENT '租户ID',\n"+
		"    `creator`                  varchar(64)   NOT NULL                              COMMENT '创建者',\n"+
		"    `reviser`                  varchar(64)   NOT NULL                              COMMENT '更新者',\n"+
		"    `created_at`               timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP    COMMENT '创建时间',\n"+
		"    `updated_at`               timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB\n"+
		"  DEFAULT CHARSET=utf8mb4\n"+
		"  COLLATE=utf8mb4_bin COMMENT='权限模板表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "permission_template", "0"); err != nil {
		return err
	}
	return nil
}
