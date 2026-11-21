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

// Package migration is the migration azure_sg_rule_uk_and_cleanup.
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
const ID = "20230329-1510-AZURE-SG-RULE-UK-AND-CLEANUP-E68E"

func init() {
	register.Main.Regist(ID, "v1.1.0", "20230329151000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0002_20230329_1510.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.DropIndex(ctx, o, "azure_security_group_rule", "idx_uk_name"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_security_group_rule", "idx_uk_name", []string{"name", "cloud_security_group_id"}, true); err != nil {
		return err
	}
	if err := util.DropColumn(ctx, o, "eip", "instance_id"); err != nil {
		return err
	}
	if err := util.DropColumn(ctx, o, "eip", "instance_type"); err != nil {
		return err
	}
	if err := util.DropColumn(ctx, o, "account", "sync_status"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "recycle_record", "idx_res_type_vendor_cloud_res_id"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "recycle_record", "idx_res_type_res_id"); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "ALTER TABLE recycle_record modify column `id` varchar(64) not null"); err != nil {
		return err
	}
	if err := util.InsertIDGenerator(ctx, o, "recycle_record_task_id", "0"); err != nil {
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
