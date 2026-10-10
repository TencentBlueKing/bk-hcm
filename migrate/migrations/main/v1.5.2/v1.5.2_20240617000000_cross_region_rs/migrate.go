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

// Package migration is the migration cross_region_rs.
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
const ID = "20240617-0000-CROSS-REGION-RS-F9F0"

func init() {
	register.Main.Regist(ID, "v1.5.2", "20240617000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0019_20240617_cross_region_rs.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "load_balancer_target",
		Column:  "ip",
		Type:    "varchar(255)",
		NotNull: true,
		Default: util.ExprDefault("(private_ip_address ->> '$[0]')"),
		After:   "inst_type",
	}); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE load_balancer_target modify ip varchar(255) not null default ''"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "load_balancer_target", "idx_uk_cloud_target_group_id_cloud_inst_id_port"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "load_balancer_target", "idx_uk_cloud_target_group_id_ip_port_cloud_inst_id", "ALTER TABLE load_balancer_target add constraint idx_uk_cloud_target_group_id_ip_port_cloud_inst_id unique (cloud_target_group_id, ip, port, cloud_inst_id)"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "target_group_listener_rule_rel",
		Column:  "vendor",
		Type:    "varchar(16)",
		NotNull: true,
		Default: util.StringDefault("tcloud"),
		After:   "id",
	}); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE target_group_listener_rule_rel modify vendor varchar(16) not null default ''"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE ssl_cert modify cloud_created_time varchar(32) default '' not null"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE ssl_cert modify cloud_expired_time varchar(32) default '' not null"); err != nil {
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
