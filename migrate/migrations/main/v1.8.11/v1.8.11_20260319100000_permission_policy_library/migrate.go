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

// Package migration is the migration permission_policy_library.
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
const ID = "20260319-1000-PERMISSION-POLICY-LIBRARY-EE3D"

func init() {
	register.Main.Regist(ID, "v1.8.11", "20260319100000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0045_20260319_1000_permission_policy_library.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "permission_policy_library", "CREATE TABLE IF NOT EXISTS `permission_policy_library` (\n"+
		"    `id`              varchar(64)   NOT NULL                              COMMENT '策略库ID',\n"+
		"    `name`            varchar(128)  NOT NULL                              COMMENT '策略库名称',\n"+
		"    `policy_document` longtext      NOT NULL                              COMMENT '当前版本的权限策略JSON内容',\n"+
		"    `policy_hash`     varchar(64)   NOT NULL                              COMMENT '策略内容SHA256哈希值',\n"+
		"    `version`         int           NOT NULL DEFAULT 1                    COMMENT '当前版本号，从1开始递增',\n"+
		"    `bk_biz_ids`      json          NOT NULL                              COMMENT '允许使用的业务ID列表',\n"+
		"    `memo`            varchar(255)  DEFAULT NULL                          COMMENT '策略库描述',\n"+
		"    `vendor`          varchar(16)   NOT NULL                              COMMENT '云厂商',\n"+
		"    `tenant_id`       varchar(64)   NOT NULL DEFAULT 'default'            COMMENT '租户ID',\n"+
		"    `creator`         varchar(64)   NOT NULL                              COMMENT '创建者',\n"+
		"    `reviser`         varchar(64)   NOT NULL                              COMMENT '更新者',\n"+
		"    `created_at`      timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP    COMMENT '创建时间',\n"+
		"    `updated_at`      timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB\n"+
		"  DEFAULT CHARSET=utf8mb4\n"+
		"  COLLATE=utf8mb4_bin COMMENT='权限策略库表'"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "permission_policy_library", "0"); err != nil {
		return err
	}
	return nil
}
