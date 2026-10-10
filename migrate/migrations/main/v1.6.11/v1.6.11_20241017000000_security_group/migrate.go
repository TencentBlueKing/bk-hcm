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

// Package migration is the migration security_group.
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
const ID = "20241017-0000-SECURITY-GROUP-621A"

func init() {
	register.Main.Regist(ID, "v1.6.11", "20241017000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0027_20241017_security_group.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "cloud_update_time",
		Type:    "VARCHAR(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "云上修改时间",
		After:   "association_template_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "cloud_created_time",
		Type:    "VARCHAR(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		Comment: "云上创建时间",
		After:   "association_template_id",
	}); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "tags",
		Type:    "json",
		NotNull: true,
		Comment: "标签",
		After:   "extension",
	}); err != nil {
		return err
	}
	return nil
}
