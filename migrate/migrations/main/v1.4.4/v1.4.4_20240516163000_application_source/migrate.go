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

// Package migration is the migration application_source.
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
const ID = "20240516-1630-APPLICATION-SOURCE-EC7F"

func init() {
	register.Main.Regist(ID, "v1.4.4", "20240516163000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0016_20240516_1630_application_source.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "application",
		Column:  "source",
		Type:    "varchar(64)",
		Default: util.StringDefault("itsm"),
		After:   "id",
	}); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "update application set source ='itsm' where source=''"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "application", "idx_uk_sn"); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "application", "idx_uk_source_sn", "ALTER TABLE application add constraint idx_uk_source_sn unique (source, sn)"); err != nil {
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
