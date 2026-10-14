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

// Package migration is the migration async_table_optimize.
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
const ID = "20241030-0000-ASYNC-TABLE-OPTIMIZE-589A"

func init() {
	register.Main.Regist(ID, "v1.6.10", "20241030000000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0026_20241030_async_table_optimize.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddIndex(ctx, o, "async_flow", "idx_worker_state_id", []string{"worker", "state"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "async_flow", "idx_state_id", []string{"state", "id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "account_bill_item", "idx_root_main_bill_day", []string{"root_account_id", "main_account_id", "bill_day"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_sync_record",
		Column:  "adjustment_flow_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault(""),
		After:   "detail",
	}); err != nil {
		return err
	}
	return nil
}
