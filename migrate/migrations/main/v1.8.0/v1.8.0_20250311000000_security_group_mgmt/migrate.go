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

// Package migration is the migration security_group_mgmt.
package migration

import (
	"context"
	"fmt"

	"hcm/migrate/register"
	"hcm/migrate/util"
	"hcm/pkg/dal/dao/orm"
)

// ID is the unique identifier of this migration file, in the form
// <date>-<time>-<desc>-<random>. It must never change once executed.
// Migrations that share an ID run only the earlier one in execution order.
const ID = "20250311-0000-SECURITY-GROUP-MGMT-4BDF"

func init() {
	register.Main.Regist(ID, "v1.8.0", "20250311000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0031_20250311_security_group_mgmt.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if _, err := util.CreateTableIfNotExists(ctx, o, "res_usage_biz_rel", "CREATE TABLE `res_usage_biz_rel`\n"+
		"(\n"+
		"    `rel_id`         bigint unsigned NOT NULL AUTO_INCREMENT,\n"+
		"    `res_type`       varchar(64)     NOT NULL COMMENT '资源类型',\n"+
		"    `res_id`         varchar(64)     NOT NULL COMMENT '资源ID',\n"+
		"    `usage_biz_id`   bigint          NOT NULL COMMENT '使用业务ID',\n"+
		"    `res_vendor`     varchar(64)     NOT NULL DEFAULT '' COMMENT '云资源厂商',\n"+
		"    `res_cloud_id`   varchar(255)    NOT NULL DEFAULT '' COMMENT '云资源ID',\n"+
		"    `rel_creator`    varchar(64)     not null comment '创建者',\n"+
		"    `rel_created_at` timestamp       not null default current_timestamp comment '创建时间',\n"+
		"    PRIMARY KEY (`rel_id`),\n"+
		"    UNIQUE KEY `idx_uk_res_type_usage_biz_id_res_id` (`res_type`, `usage_biz_id`, `res_id`),\n"+
		"    KEY idx_res_type_res_id_usage_biz_id (res_type, res_id, usage_biz_id)\n"+
		")"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "mgmt_type",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "管理类型",
		After:   "account_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "mgmt_biz_id",
		Type:    "bigint",
		NotNull: true,
		Default: util.ExprDefault("-1"),
		Comment: "管理业务ID",
		After:   "mgmt_type",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "manager",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "负责人",
		After:   "mgmt_biz_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "bak_manager",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "备份负责人",
		After:   "manager",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "security_group_common_rel", "idx_security_group_id", []string{"security_group_id"}, false); err != nil {
		return err
	}
	if err := util.RenameColumn(ctx, o, "security_group_common_rel", "vendor", "res_vendor"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE security_group_common_rel MODIFY COLUMN res_vendor varchar(16)"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_security_group_rule", "idx_security_group_id_cloud_target_security_group_id_region", []string{"security_group_id", "cloud_target_security_group_id", "region"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "huawei_security_group_rule", "idx_security_group_id_cloud_remote_group_id_region", []string{"security_group_id", "cloud_remote_group_id", "region"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "aws_security_group_rule", "idx_security_group_id_cloud_target_security_group_id_region", []string{"security_group_id", "cloud_target_security_group_id", "region"}, false); err != nil {
		return err
	}
	return nil
}

// execSQL runs a statement that util has no idempotent helper for. Re-running it is safe, see the plan.
func execSQL(ctx context.Context, o orm.Interface, sql string) error {
	if _, err := o.Do().Exec(ctx, sql); err != nil {
		return fmt.Errorf("exec %q failed, err: %v", sql, err)
	}
	return nil
}
