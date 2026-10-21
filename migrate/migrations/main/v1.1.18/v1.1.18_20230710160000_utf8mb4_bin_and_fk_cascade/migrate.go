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

// Package migration is the migration utf8mb4_bin_and_fk_cascade.
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
const ID = "20230710-1600-UTF8MB4-BIN-AND-FK-CASCADE-040A"

func init() {
	register.Main.Regist(ID, "v1.1.18", "20230710160000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0006_20230710_1600.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.DropConstraint(ctx, o, "disk_cvm_rel", "disk_cvm_rel_cvm_id", "ALTER TABLE disk_cvm_rel drop foreign key disk_cvm_rel_cvm_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "disk_cvm_rel", "disk_cvm_rel_disk_id", "ALTER TABLE disk_cvm_rel drop foreign key disk_cvm_rel_disk_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "security_group_cvm_rel", "security_group_cvm_rel_security_group_id", "ALTER TABLE security_group_cvm_rel drop foreign key security_group_cvm_rel_security_group_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "security_group_cvm_rel", "security_group_cvm_rel_cvm_id", "ALTER TABLE security_group_cvm_rel drop foreign key security_group_cvm_rel_cvm_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "eip_cvm_rel", "eip_cvm_rel_eip_id", "ALTER TABLE eip_cvm_rel drop foreign key eip_cvm_rel_eip_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "eip_cvm_rel", "eip_cvm_rel_cvm_id", "ALTER TABLE eip_cvm_rel drop foreign key eip_cvm_rel_cvm_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "network_interface_cvm_rel", "network_interface_cvm_rel_network_id", "ALTER TABLE network_interface_cvm_rel drop foreign key network_interface_cvm_rel_network_id"); err != nil {
		return err
	}
	if err := util.DropConstraint(ctx, o, "network_interface_cvm_rel", "network_interface_cvm_rel_cvm_id", "ALTER TABLE network_interface_cvm_rel drop foreign key network_interface_cvm_rel_cvm_id"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE eip_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE security_group_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE account_biz_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE disk_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE network_interface_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE account CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE account_bill_config CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE application CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE approval_process CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE audit CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE aws_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE aws_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE aws_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE azure_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE azure_resource_group CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE azure_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE azure_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE cvm CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE disk CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE eip CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE gcp_firewall_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE gcp_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE gcp_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE huawei_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE huawei_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE huawei_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE id_generator CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE image CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE network_interface CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE recycle_record CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE route_table CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE security_group CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE subnet CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE tcloud_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE tcloud_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE tcloud_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE vpc CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE zone CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "disk_cvm_rel", "disk_cvm_rel_cvm_id", "ALTER TABLE disk_cvm_rel add constraint disk_cvm_rel_cvm_id foreign key (disk_id) REFERENCES disk (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "disk_cvm_rel", "disk_cvm_rel_disk_id", "ALTER TABLE disk_cvm_rel add constraint disk_cvm_rel_disk_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "security_group_cvm_rel", "security_group_cvm_rel_security_group_id", "ALTER TABLE security_group_cvm_rel add constraint security_group_cvm_rel_security_group_id foreign key (security_group_id) REFERENCES security_group (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "security_group_cvm_rel", "security_group_cvm_rel_cvm_id", "ALTER TABLE security_group_cvm_rel add constraint security_group_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "eip_cvm_rel", "eip_cvm_rel_eip_id", "ALTER TABLE eip_cvm_rel add constraint eip_cvm_rel_eip_id foreign key (eip_id) REFERENCES eip (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "eip_cvm_rel", "eip_cvm_rel_cvm_id", "ALTER TABLE eip_cvm_rel add constraint eip_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "network_interface_cvm_rel", "network_interface_cvm_rel_network_id", "ALTER TABLE network_interface_cvm_rel add constraint network_interface_cvm_rel_network_id foreign key (network_interface_id) REFERENCES network_interface (id) ON DELETE CASCADE"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "network_interface_cvm_rel", "network_interface_cvm_rel_cvm_id", "ALTER TABLE network_interface_cvm_rel add constraint network_interface_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE"); err != nil {
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
