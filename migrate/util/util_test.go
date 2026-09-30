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

package util

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// countFnFor returns a Count implementation that reports exist for the given
// probe (as used by MetaOrm's Has* methods).
func countFnFor(exist bool) func(string, map[string]interface{}) (uint64, error) {
	return func(_ string, _ map[string]interface{}) (uint64, error) {
		if exist {
			return 1, nil
		}
		return 0, nil
	}
}

func TestCreateTableIfNotExists(t *testing.T) {
	testCases := []struct {
		name        string
		table       string
		ddl         string
		tableExist  bool
		countErr    error
		execErr     error
		wantExec    bool
		wantCreated bool
		wantErr     bool
	}{
		{
			name:        "create when absent",
			table:       "cvm",
			ddl:         "CREATE TABLE `cvm` (id BIGINT)",
			wantExec:    true,
			wantCreated: true,
		},
		{
			name:       "skip when present",
			table:      "cvm",
			ddl:        "CREATE TABLE `cvm` (id BIGINT)",
			tableExist: true,
			wantExec:   false,
		},
		{
			// scripts/sql/0001_20230227_2045_init_db.sql
			name:  "0001_id_generator create when absent",
			table: "id_generator",
			ddl: "create table if not exists `id_generator` (`resource` varchar(64) not null, " +
				"`max_id` varchar(64) not null, primary key (`resource`))",
			wantExec:    true,
			wantCreated: true,
		},
		{
			name:  "0001_id_generator skip when present",
			table: "id_generator",
			ddl: "create table if not exists `id_generator` (`resource` varchar(64) not null, " +
				"`max_id` varchar(64) not null, primary key (`resource`))",
			tableExist: true,
		},
		{
			// scripts/sql/0047_20260130_1800_account_secret.sql
			name:  "0047_account_secret create when absent",
			table: "account_secret",
			ddl: "CREATE TABLE IF NOT EXISTS `account_secret` (`id` varchar(64) NOT NULL COMMENT '密钥ID', " +
				"`account_id` varchar(64) NOT NULL COMMENT '账号ID')",
			wantExec:    true,
			wantCreated: true,
		},
		{
			name:    "invalid table name",
			table:   "cvm; DROP TABLE x",
			ddl:     "CREATE TABLE x (id BIGINT)",
			wantErr: true,
		},
		{
			name:    "empty ddl",
			table:   "cvm",
			ddl:     "  ",
			wantErr: true,
		},
		{
			name:    "empty string ddl",
			table:   "cvm",
			ddl:     "",
			wantErr: true,
		},
		{
			name:    "exec error",
			table:   "cvm",
			ddl:     "CREATE TABLE `cvm` (id BIGINT)",
			execErr: errors.New("disk full"),
			wantErr: true,
			// The statement is attempted; created stays false.
			wantExec: true,
		},
		{
			name:     "count error",
			table:    "cvm",
			ddl:      "CREATE TABLE `cvm` (id BIGINT)",
			countErr: errors.New("information_schema down"),
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.tableExist), execErr: tc.execErr}
			if tc.countErr != nil {
				do.countFn = func(string, map[string]interface{}) (uint64, error) {
					return 0, tc.countErr
				}
			}
			created, err := CreateTableIfNotExists(context.Background(), newFakeOrm(do), tc.table, tc.ddl)
			assert.Equal(t, tc.wantCreated, created)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.wantExec {
				assert.Equal(t, []string{tc.ddl}, do.execCalls)
			} else {
				assert.Empty(t, do.execCalls)
			}
		})
	}
}

func TestDropTable(t *testing.T) {
	testCases := []struct {
		name       string
		table      string
		tableExist bool
		wantDDL    string
	}{
		{name: "drop when present", table: "cvm", tableExist: true, wantDDL: "DROP TABLE `cvm`"},
		{name: "skip when absent", table: "cvm"},
		{
			// scripts/sql/0036_20241217_1630_resource_plan.sql
			name: "0036_res_plan_demand drop when present", table: "res_plan_demand", tableExist: true,
			wantDDL: "DROP TABLE `res_plan_demand`",
		},
		{
			// scripts/sql/0080_20260618_1004_resource_dissolve.sql
			name: "0080_recycle_module_info skip when absent", table: "recycle_module_info",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.tableExist)}
			err := DropTable(context.Background(), newFakeOrm(do), tc.table)
			assert.NoError(t, err)
			if tc.wantDDL == "" {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.wantDDL}, do.execCalls)
		})
	}
}

func TestAddColumn(t *testing.T) {
	testCases := []struct {
		name        string
		opt         AddColumnOpt
		columnExist bool
		wantDDL     string
		wantErr     bool
	}{
		{
			name:    "minimal",
			opt:     AddColumnOpt{Table: "cvm", Column: "bk_asset_id", Type: "varchar(64)"},
			wantDDL: "ALTER TABLE `cvm` ADD COLUMN `bk_asset_id` varchar(64)",
		},
		{
			name: "full option",
			opt: AddColumnOpt{
				Table:   "cvm",
				Column:  "bk_asset_id",
				Type:    "varchar(64)",
				NotNull: true,
				Default: StringDefault(""),
				Comment: "it's the asset id",
				After:   "id",
			},
			wantDDL: "ALTER TABLE `cvm` ADD COLUMN `bk_asset_id` varchar(64) NOT NULL DEFAULT '' " +
				"COMMENT 'it\\'s the asset id' AFTER `id`",
		},
		{
			name:        "skip when column exists",
			opt:         AddColumnOpt{Table: "cvm", Column: "bk_asset_id", Type: "varchar(64)"},
			columnExist: true,
		},
		{
			name:    "missing type",
			opt:     AddColumnOpt{Table: "cvm", Column: "bk_asset_id"},
			wantErr: true,
		},
		{
			name:    "invalid column name",
			opt:     AddColumnOpt{Table: "cvm", Column: "bk asset id", Type: "varchar(64)"},
			wantErr: true,
		},
		{
			name:    "empty column name",
			opt:     AddColumnOpt{Table: "cvm", Column: "", Type: "bigint"},
			wantErr: true,
		},
		{
			name:    "leading digit column name",
			opt:     AddColumnOpt{Table: "cvm", Column: "1core", Type: "bigint"},
			wantErr: true,
		},
		{
			name:    "backtick in column name",
			opt:     AddColumnOpt{Table: "cvm", Column: "id`", Type: "bigint"},
			wantErr: true,
		},
		{
			name:    "dotted db.table",
			opt:     AddColumnOpt{Table: "hcm.cvm", Column: "id", Type: "bigint"},
			wantErr: true,
		},
		{
			name: "empty default expression",
			opt: AddColumnOpt{
				Table: "cvm", Column: "bk_biz_id", Type: "bigint", Default: ExprDefault("   "),
			},
			wantErr: true,
		},
		{
			name: "quoted NULL is not keyword NULL",
			opt: AddColumnOpt{
				Table: "cvm", Column: "memo", Type: "varchar(255)", Default: StringDefault("NULL"),
			},
			wantDDL: "ALTER TABLE `cvm` ADD COLUMN `memo` varchar(255) DEFAULT 'NULL'",
		},
		{
			name: "chinese comment and default escape quote and backslash",
			opt: AddColumnOpt{
				Table:   "cvm",
				Column:  "memo",
				Type:    "varchar(255)",
				Default: StringDefault("it's\\x"),
				Comment: "减免退还核心数，it's a\\core",
			},
			wantDDL: "ALTER TABLE `cvm` ADD COLUMN `memo` varchar(255) DEFAULT 'it\\'s\\\\x' " +
				"COMMENT '减免退还核心数，it\\'s a\\\\core'",
		},
		{
			// scripts/sql/0005_20230530_2100.sql
			name: "0005_approval_process_managers",
			opt: AddColumnOpt{
				Table: "approval_process", Column: "managers", Type: "varchar(255)", NotNull: true,
			},
			wantDDL: "ALTER TABLE `approval_process` ADD COLUMN `managers` varchar(255) NOT NULL",
		},
		{
			// scripts/sql/0016_20240516_1630_application_source.sql
			name: "0016_application_source",
			opt: AddColumnOpt{
				Table: "application", Column: "source", Type: "varchar(64)",
				Default: StringDefault("itsm"), After: "id",
			},
			wantDDL: "ALTER TABLE `application` ADD COLUMN `source` varchar(64) DEFAULT 'itsm' AFTER `id`",
		},
		{
			// scripts/sql/0032_20250311_assign_host.sql
			name: "0032_bk_host_id",
			opt: AddColumnOpt{
				Table: "cvm", Column: "bk_host_id", Type: "bigint",
				Default: ExprDefault("-1"), Comment: "主机ID",
			},
			wantDDL: "ALTER TABLE `cvm` ADD COLUMN `bk_host_id` bigint DEFAULT -1 COMMENT '主机ID'",
		},
		{
			// scripts/sql/0035_account_biz_id.sql
			name: "0035_account_bk_biz_id",
			opt: AddColumnOpt{
				Table: "account", Column: "bk_biz_id", Type: "bigint", NotNull: true,
				Default: ExprDefault("0"), Comment: "管理业务ID",
			},
			wantDDL: "ALTER TABLE `account` ADD COLUMN `bk_biz_id` bigint NOT NULL DEFAULT 0 COMMENT '管理业务ID'",
		},
		{
			// scripts/sql/0042_20251013_extract_bk_asset_id.sql
			name: "0042_bk_asset_id",
			opt: AddColumnOpt{
				Table: "cvm", Column: "bk_asset_id", Type: "VARCHAR(64)",
				Default: StringDefault(""), Comment: "固资号",
			},
			wantDDL: "ALTER TABLE `cvm` ADD COLUMN `bk_asset_id` VARCHAR(64) DEFAULT '' COMMENT '固资号'",
		},
		{
			// scripts/sql/0044_20250701_1530_rolling_server_notice.sql
			name: "0044_not_notice",
			opt: AddColumnOpt{
				Table: "rolling_applied_record", Column: "not_notice", Type: "boolean",
				Default: ExprDefault("false"), Comment: "滚服到期是否不提醒",
			},
			wantDDL: "ALTER TABLE `rolling_applied_record` ADD COLUMN `not_notice` boolean DEFAULT false " +
				"COMMENT '滚服到期是否不提醒'",
		},
		{
			// scripts/sql/0047_20260130_1800_account_secret.sql created_at shape
			name: "0047_created_at_current_timestamp",
			opt: AddColumnOpt{
				Table: "account_secret", Column: "created_at", Type: "timestamp", NotNull: true,
				Default: ExprDefault("CURRENT_TIMESTAMP"), Comment: "创建时间",
			},
			wantDDL: "ALTER TABLE `account_secret` ADD COLUMN `created_at` timestamp NOT NULL " +
				"DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'",
		},
		{
			// scripts/sql/0049_20260202_0958_account_fields.sql
			name: "0049_security_managers",
			opt: AddColumnOpt{
				Table: "account", Column: "security_managers", Type: "json",
				Default: ExprDefault("NULL"), Comment: "安全负责人",
			},
			wantDDL: "ALTER TABLE `account` ADD COLUMN `security_managers` json DEFAULT NULL COMMENT '安全负责人'",
		},
		{
			// scripts/sql/0062_20251205_1445_tcloud_ziyan_region_zone.sql
			name: "0062_zone_source",
			opt: AddColumnOpt{
				Table: "zone", Column: "source", Type: "varchar(64)", NotNull: true,
				Default: StringDefault("sync"), Comment: "来源：sync-同步，manually-手动添加", After: "state",
			},
			wantDDL: "ALTER TABLE `zone` ADD COLUMN `source` varchar(64) NOT NULL DEFAULT 'sync' " +
				"COMMENT '来源：sync-同步，manually-手动添加' AFTER `state`",
		},
		{
			// scripts/sql/0065_20251219_add_image_region_field.sql
			name: "0065_image_region",
			opt: AddColumnOpt{
				Table: "image", Column: "region", Type: "VARCHAR(64)",
				Default: StringDefault(""), After: "cloud_id",
			},
			wantDDL: "ALTER TABLE `image` ADD COLUMN `region` VARCHAR(64) DEFAULT '' AFTER `cloud_id`",
		},
		{
			// scripts/sql/0068_20260227_1123_res_plan_sub_ticket.sql
			name: "0068_operate_info",
			opt: AddColumnOpt{
				Table: "res_plan_sub_ticket", Column: "operate_info", Type: "varchar(100)",
				Default: ExprDefault("NULL"), Comment: "管理员审批意见，最多 100 字",
			},
			wantDDL: "ALTER TABLE `res_plan_sub_ticket` ADD COLUMN `operate_info` varchar(100) DEFAULT NULL " +
				"COMMENT '管理员审批意见，最多 100 字'",
		},
		{
			// scripts/sql/0070_20260121_1600_add_exempted_returned_core.sql
			// same shape as scripts/obssql/0003_20260121_1600_add_exempted_returned_core.sql
			name: "0070_exempted_returned_core",
			opt: AddColumnOpt{
				Table: "rolling_fine_detail", Column: "exempted_returned_core", Type: "bigint unsigned",
				NotNull: true, Default: ExprDefault("0"), Comment: "减免退还核心数",
			},
			wantDDL: "ALTER TABLE `rolling_fine_detail` ADD COLUMN `exempted_returned_core` bigint unsigned " +
				"NOT NULL DEFAULT 0 COMMENT '减免退还核心数'",
		},
		{
			// scripts/sql/0075_20260429_1100_device_type_gpu_amount.sql
			name: "0075_gpu_amount",
			opt: AddColumnOpt{
				Table: "device_type", Column: "gpu_amount", Type: "DOUBLE", NotNull: true,
				Default: ExprDefault("0"), Comment: "GPU卡数", After: "memory",
			},
			wantDDL: "ALTER TABLE `device_type` ADD COLUMN `gpu_amount` DOUBLE NOT NULL DEFAULT 0 " +
				"COMMENT 'GPU卡数' AFTER `memory`",
		},
		{
			// scripts/sql/0080_20260618_1004_resource_dissolve.sql
			name: "0080_operators",
			opt: AddColumnOpt{
				Table: "recycle_host_info", Column: "operators", Type: "json", Comment: "负责人列表",
			},
			wantDDL: "ALTER TABLE `recycle_host_info` ADD COLUMN `operators` json COMMENT '负责人列表'",
		},
		{
			// scripts/sql/0082_20260602_device_type_tech_class_res_amt.sql
			name: "0082_tech_class_res_amt",
			opt: AddColumnOpt{
				Table: "device_type", Column: "tech_class_res_amt", Type: "DECIMAL(10,2)", NotNull: true,
				Default: ExprDefault("0"), Comment: "技术分类资源量", After: "technical_class",
			},
			wantDDL: "ALTER TABLE `device_type` ADD COLUMN `tech_class_res_amt` DECIMAL(10,2) NOT NULL DEFAULT 0 " +
				"COMMENT '技术分类资源量' AFTER `technical_class`",
		},
		{
			// scripts/sql/0084_20260804_account_bill_adjustment_item_add_res_sub_class.sql
			name: "0084_res_sub_class",
			opt: AddColumnOpt{
				Table: "account_bill_adjustment_item", Column: "res_sub_class", Type: "varchar(64)",
				Default: ExprDefault("NULL"),
				Comment: "调账资源子类，gpu_card 下为卡型、gpu_api 下为模型厂商",
				After:   "res_class",
			},
			wantDDL: "ALTER TABLE `account_bill_adjustment_item` ADD COLUMN `res_sub_class` varchar(64) " +
				"DEFAULT NULL COMMENT '调账资源子类，gpu_card 下为卡型、gpu_api 下为模型厂商' " +
				"AFTER `res_class`",
		},
		{
			// scripts/sql/0087_20260603_1558_aiagent_session_bk_biz_id.sql
			name: "0087_aiagent_session_bk_biz_id",
			opt: AddColumnOpt{
				Table: "aiagent_session", Column: "bk_biz_id", Type: "BIGINT", NotNull: true,
				Default: ExprDefault("-1"), Comment: "会话所属业务；-1=未分配", After: "user",
			},
			wantDDL: "ALTER TABLE `aiagent_session` ADD COLUMN `bk_biz_id` BIGINT NOT NULL DEFAULT -1 " +
				"COMMENT '会话所属业务；-1=未分配' AFTER `user`",
		},
		{
			// scripts/obssql/0004_add_city_res_class_fields.sql
			name: "obssql_0004_CityId",
			opt: AddColumnOpt{
				Table: "obs_aws_bills", Column: "CityId", Type: "int(11)", NotNull: true,
				Default: ExprDefault("0"),
			},
			wantDDL: "ALTER TABLE `obs_aws_bills` ADD COLUMN `CityId` int(11) NOT NULL DEFAULT 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.columnExist)}
			err := AddColumn(context.Background(), newFakeOrm(do), tc.opt)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, do.execCalls)
				return
			}
			assert.NoError(t, err)
			if tc.wantDDL == "" {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.wantDDL}, do.execCalls)
		})
	}
}

func TestDropColumn(t *testing.T) {
	testCases := []struct {
		name   string
		table  string
		column string
		exist  bool
		want   string
	}{
		{name: "drop when present", table: "cvm", column: "bk_asset_id", exist: true,
			want: "ALTER TABLE `cvm` DROP COLUMN `bk_asset_id`"},
		{name: "skip when absent", table: "cvm", column: "bk_asset_id"},
		{
			// scripts/sql/0002_20230329_1510.sql
			name: "0002_eip_instance_id", table: "eip", column: "instance_id", exist: true,
			want: "ALTER TABLE `eip` DROP COLUMN `instance_id`",
		},
		{
			// scripts/sql/0025_20240814_1145_bill_account_cloud_id.sql
			name: "0025_root_account_name", table: "account_bill_summary_main",
			column: "root_account_name", exist: true,
			want: "ALTER TABLE `account_bill_summary_main` DROP COLUMN `root_account_name`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.exist)}
			err := DropColumn(context.Background(), newFakeOrm(do), tc.table, tc.column)
			assert.NoError(t, err)
			if tc.want == "" {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.want}, do.execCalls)
		})
	}
}

func TestRenameColumn(t *testing.T) {
	testCases := []struct {
		name     string
		oldExist bool
		newExist bool
		wantDDL  string
		wantErr  bool
	}{
		{
			name:     "rename when old exists and new does not",
			oldExist: true,
			wantDDL:  "ALTER TABLE `cvm` RENAME COLUMN `old` TO `new`",
		},
		{
			name:     "skip when new already exists",
			newExist: true,
		},
		{
			name:    "error when neither exists",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{
				countFn: func(expr string, arg map[string]interface{}) (uint64, error) {
					column, _ := arg["column"].(string)
					switch column {
					case "new":
						if tc.newExist {
							return 1, nil
						}
					case "old":
						if tc.oldExist {
							return 1, nil
						}
					}
					return 0, nil
				},
			}

			err := RenameColumn(context.Background(), newFakeOrm(do), "cvm", "old", "new")
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, do.execCalls)
				return
			}
			assert.NoError(t, err)
			if tc.wantDDL == "" {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.wantDDL}, do.execCalls)
		})
	}
}

func TestAddIndex(t *testing.T) {
	testCases := []struct {
		name       string
		table      string
		index      string
		columns    []string
		unique     bool
		indexExist bool
		wantDDL    string
		wantErr    bool
	}{
		{
			name:    "add plain index",
			table:   "cvm",
			index:   "idx_uk_asset",
			columns: []string{"vendor", "bk_asset_id"},
			wantDDL: "ALTER TABLE `cvm` ADD INDEX `idx_uk_asset` (`vendor`, `bk_asset_id`)",
		},
		{
			name:    "add unique index",
			table:   "cvm",
			index:   "idx_uk_asset",
			columns: []string{"vendor", "bk_asset_id"},
			unique:  true,
			wantDDL: "ALTER TABLE `cvm` ADD UNIQUE INDEX `idx_uk_asset` (`vendor`, `bk_asset_id`)",
		},
		{
			name:       "skip when exists",
			table:      "cvm",
			index:      "idx_uk_asset",
			columns:    []string{"vendor"},
			indexExist: true,
		},
		{
			name:    "no columns",
			table:   "cvm",
			index:   "idx_uk_asset",
			columns: nil,
			wantErr: true,
		},
		{
			name:    "invalid column",
			table:   "cvm",
			index:   "idx_uk_asset",
			columns: []string{"vendor; DROP TABLE x"},
			wantErr: true,
		},
		{
			// scripts/sql/0023_20240531_1648_bill.sql
			name:    "0023_idx_state_updated_at",
			table:   "async_flow_task",
			index:   "idx_state_updated_at",
			columns: []string{"state", "updated_at"},
			wantDDL: "ALTER TABLE `async_flow_task` ADD INDEX `idx_state_updated_at` (`state`, `updated_at`)",
		},
		{
			// scripts/sql/0023_20240531_1648_bill.sql
			name:    "0023_idx_flow_id",
			table:   "async_flow_task",
			index:   "idx_flow_id",
			columns: []string{"flow_id"},
			wantDDL: "ALTER TABLE `async_flow_task` ADD INDEX `idx_flow_id` (`flow_id`)",
		},
		{
			// scripts/sql/0047_20260130_1800_account_secret.sql
			name:    "0047_idx_account_id",
			table:   "account_secret",
			index:   "idx_account_id",
			columns: []string{"account_id"},
			wantDDL: "ALTER TABLE `account_secret` ADD INDEX `idx_account_id` (`account_id`)",
		},
		{
			// scripts/sql/0047_20260130_1800_account_secret.sql
			name:       "0047_idx_account_id skip when exists",
			table:      "account_secret",
			index:      "idx_account_id",
			columns:    []string{"account_id"},
			indexExist: true,
		},
		{
			// scripts/sql/0003_20230516_1600.sql
			name:    "0003_idx_uk_name_cloud_security_group_id",
			table:   "azure_security_group_rule",
			index:   "idx_uk_name_cloud_security_group_id",
			columns: []string{"name", "cloud_security_group_id"},
			unique:  true,
			wantDDL: "ALTER TABLE `azure_security_group_rule` ADD UNIQUE INDEX " +
				"`idx_uk_name_cloud_security_group_id` (`name`, `cloud_security_group_id`)",
		},
		{
			// scripts/sql/0032_20250311_assign_host.sql
			name:    "0032_idx_bk_host_id",
			table:   "cvm",
			index:   "idx_bk_host_id",
			columns: []string{"bk_host_id", "id"},
			wantDDL: "ALTER TABLE `cvm` ADD INDEX `idx_bk_host_id` (`bk_host_id`, `id`)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.indexExist)}
			err := AddIndex(context.Background(), newFakeOrm(do), tc.table, tc.index, tc.columns, tc.unique)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, do.execCalls)
				return
			}
			assert.NoError(t, err)
			if tc.wantDDL == "" {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.wantDDL}, do.execCalls)
		})
	}
}

func TestDropIndex(t *testing.T) {
	testCases := []struct {
		name  string
		table string
		index string
		exist bool
		want  string
	}{
		{name: "drop when present", table: "cvm", index: "idx_uk_asset", exist: true,
			want: "ALTER TABLE `cvm` DROP INDEX `idx_uk_asset`"},
		{name: "skip when absent", table: "cvm", index: "idx_uk_asset"},
		{
			// scripts/sql/0002_20230329_1510.sql
			name: "0002_idx_uk_name", table: "azure_security_group_rule", index: "idx_uk_name", exist: true,
			want: "ALTER TABLE `azure_security_group_rule` DROP INDEX `idx_uk_name`",
		},
		{
			// scripts/sql/0087_20260603_1558_aiagent_session_bk_biz_id.sql
			name: "0087_idx_app_user", table: "aiagent_session", index: "idx_app_user", exist: true,
			want: "ALTER TABLE `aiagent_session` DROP INDEX `idx_app_user`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.exist)}
			err := DropIndex(context.Background(), newFakeOrm(do), tc.table, tc.index)
			assert.NoError(t, err)
			if tc.want == "" {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.want}, do.execCalls)
		})
	}
}

func TestAddConstraint(t *testing.T) {
	testCases := []struct {
		name       string
		table      string
		constraint string
		ddl        string
		exist      bool
		wantErr    bool
	}{
		{
			name:       "add when absent",
			table:      "cvm",
			constraint: "fk_account",
			ddl: "ALTER TABLE `cvm` ADD CONSTRAINT `fk_account` FOREIGN KEY (`account_id`) " +
				"REFERENCES `account` (`id`)",
		},
		{
			name:       "skip when present",
			table:      "cvm",
			constraint: "fk_account",
			ddl: "ALTER TABLE `cvm` ADD CONSTRAINT `fk_account` FOREIGN KEY (`account_id`) " +
				"REFERENCES `account` (`id`)",
			exist: true,
		},
		{
			name:       "empty ddl",
			table:      "cvm",
			constraint: "fk_account",
			ddl:        "  ",
			wantErr:    true,
		},
		{
			// scripts/sql/0006_20230710_1600.sql
			name:       "0006_disk_cvm_rel_cvm_id",
			table:      "disk_cvm_rel",
			constraint: "disk_cvm_rel_cvm_id",
			ddl: "alter table disk_cvm_rel add constraint disk_cvm_rel_cvm_id foreign key (disk_id) " +
				"REFERENCES disk (id) ON DELETE CASCADE",
		},
		{
			// scripts/sql/0033_20241104_1700_load_balancer_region.sql
			name:       "0033_idx_uk_cloud_id_vendor_region",
			table:      "load_balancer",
			constraint: "idx_uk_cloud_id_vendor_region",
			ddl: "alter table load_balancer add constraint idx_uk_cloud_id_vendor_region " +
				"unique (cloud_id, vendor, region)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.exist)}
			err := AddConstraint(context.Background(), newFakeOrm(do), tc.table, tc.constraint, tc.ddl)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, do.execCalls)
				return
			}
			assert.NoError(t, err)
			if tc.exist {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.ddl}, do.execCalls)
		})
	}
}

func TestDropConstraint(t *testing.T) {
	testCases := []struct {
		name       string
		table      string
		constraint string
		ddl        string
		exist      bool
		wantErr    bool
	}{
		{
			name:       "drop when present",
			table:      "cvm",
			constraint: "fk_account",
			ddl:        "ALTER TABLE `cvm` DROP FOREIGN KEY `fk_account`",
			exist:      true,
		},
		{
			name:       "skip when absent",
			table:      "cvm",
			constraint: "fk_account",
			ddl:        "ALTER TABLE `cvm` DROP FOREIGN KEY `fk_account`",
		},
		{
			name:       "empty ddl",
			table:      "cvm",
			constraint: "fk_account",
			ddl:        "  ",
			exist:      true,
			wantErr:    true,
		},
		{
			// scripts/sql/0006_20230710_1600.sql
			name:       "0006_disk_cvm_rel_cvm_id",
			table:      "disk_cvm_rel",
			constraint: "disk_cvm_rel_cvm_id",
			ddl:        "alter table disk_cvm_rel drop foreign key disk_cvm_rel_cvm_id",
			exist:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{countFn: countFnFor(tc.exist)}
			err := DropConstraint(context.Background(), newFakeOrm(do), tc.table, tc.constraint, tc.ddl)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, do.execCalls)
				return
			}
			assert.NoError(t, err)
			if !tc.exist {
				assert.Empty(t, do.execCalls)
				return
			}
			assert.Equal(t, []string{tc.ddl}, do.execCalls)
		})
	}
}

func TestInsertIDGenerator(t *testing.T) {
	testCases := []struct {
		name      string
		resource  string
		maxID     string
		insertErr error
		wantErr   bool
	}{
		{
			// scripts/sql/0001_20230227_2045_init_db.sql
			name: "0001_account", resource: "account", maxID: "0",
		},
		{
			// scripts/sql/0045_20260319_1000_permission_policy_library.sql
			name: "0045_permission_policy_library", resource: "permission_policy_library", maxID: "0",
		},
		{
			// scripts/sql/0047_20260130_1800_account_secret.sql
			name: "0047_account_secret", resource: "account_secret", maxID: "0",
		},
		{name: "empty resource", maxID: "0", wantErr: true},
		{name: "empty max_id", resource: "cvm", wantErr: true},
		{name: "insert error", resource: "cvm", maxID: "0", insertErr: errors.New("duplicate key"), wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{insertErr: tc.insertErr}
			err := InsertIDGenerator(context.Background(), newFakeOrm(do), tc.resource, tc.maxID)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, []fakeInsertCall{{
				expr: idGeneratorInsertExpr,
				data: map[string]interface{}{"resource": tc.resource, "max_id": tc.maxID},
			}}, do.insertCalls)
		})
	}
}
