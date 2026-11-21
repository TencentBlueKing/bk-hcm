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

// Package migration is the migration sub_account_fields.
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
const ID = "20260202-1000-SUB-ACCOUNT-FIELDS-CEFE"

func init() {
	register.Main.Regist(ID, "v1.9.0", "20260202100000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0050_20260202_1000_sub_account_fields.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "sub_account",
		Column:  "email",
		Type:    "varchar(64)",
		Default: util.ExprDefault("NULL"),
		Comment: "邮箱",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "sub_account",
		Column:  "phone_num",
		Type:    "varchar(64)",
		Default: util.ExprDefault("NULL"),
		Comment: "手机号",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "sub_account",
		Column:  "country_code",
		Type:    "varchar(16)",
		Default: util.ExprDefault("NULL"),
		Comment: "手机区域代码",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "sub_account",
		Column:  "cloud_created_at",
		Type:    "varchar(64)",
		Comment: "云上创建时间",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "sub_account",
		Column:  "permission_template_ids",
		Type:    "json",
		Default: util.ExprDefault("NULL"),
		Comment: "权限模板ID列表",
	}); err != nil {
		return err
	}
	return nil
}
