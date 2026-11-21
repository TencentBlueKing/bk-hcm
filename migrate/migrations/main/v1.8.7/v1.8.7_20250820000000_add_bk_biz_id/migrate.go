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

// Package migration is the migration add_bk_biz_id.
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
const ID = "20250820-0000-ADD-BK-BIZ-ID-0055"

func init() {
	register.Main.Regist(ID, "v1.8.7", "20250820000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0041_20250820_add_bk_biz_id.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_lb_url_rule",
		Column:  "bk_biz_id",
		Type:    "bigint",
		NotNull: true,
		Default: util.ExprDefault("0"),
		Comment: "业务ID",
		After:   "cloud_lb_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_lb_url_rule",
		Column:  "account_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "账号ID",
		After:   "bk_biz_id",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_lb_url_rule", "idx_bk_biz_id", []string{"bk_biz_id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_lb_url_rule", "idx_account_id", []string{"account_id"}, false); err != nil {
		return err
	}
	if err := execSQL(ctx, o, "UPDATE `tcloud_lb_url_rule` t\n"+
		"INNER JOIN `load_balancer` lb ON t.lb_id = lb.id\n"+
		"SET \n"+
		"    t.bk_biz_id = lb.bk_biz_id,\n"+
		"    t.account_id = lb.account_id\n"+
		"WHERE \n"+
		"    t.bk_biz_id = 0 OR t.account_id = ''"); err != nil {
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
