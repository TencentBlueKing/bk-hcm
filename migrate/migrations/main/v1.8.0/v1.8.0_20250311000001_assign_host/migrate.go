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

// Package migration is the migration assign_host.
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
const ID = "20250311-0000-ASSIGN-HOST-A80B"

func init() {
	register.Main.Regist(ID, "v1.8.0", "20250311000001", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0032_20250311_assign_host.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.DropColumn(ctx, o, "vpc", "bk_cloud_id"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "cvm",
		Column:  "bk_host_id",
		Type:    "bigint",
		Default: util.ExprDefault("-1"),
		Comment: "主机ID",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "cvm", "idx_bk_host_id", []string{"bk_host_id", "id"}, false); err != nil {
		return err
	}
	return nil
}
