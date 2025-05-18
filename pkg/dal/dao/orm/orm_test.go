/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package orm

import (
	"testing"

	"hcm/pkg/criteria/constant"

	"github.com/stretchr/testify/assert"
)

func Test_doTableSharding_replaceFromJoinTableName(t *testing.T) {

	tests := []struct {
		name              string
		tableShardingOpts []TableShardingOpt
		origin            string
		wantReplaced      string
	}{
		{
			name:              "simple select",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "select * from table_1",
			wantReplaced:      "select * from table_1_replaced",
		},
		{
			name:              "simple select with db name and backquote",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "select * from db1.`table_1`",
			wantReplaced:      "select * from db1.`table_1_replaced`",
		},
		{
			name:              "simple select with db name back quoted",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "select * from `db1`.`table_1`",
			wantReplaced:      "select * from `db1`.`table_1_replaced`",
		},
		{
			name:              "select with join",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "select * from table_1 join table_2 on table_1.id = table_2.id",
			wantReplaced:      "select * from table_1_replaced join table_2_replaced on table_1_replaced.id = table_2_replaced.id",
		},
		{
			name:              "select,update,delete,join,table alias,sub query",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin: `SELECT a.id, b.name FROM users AS a JOIN accounts AS b ON a.id = b.user_id
			WHERE a.age > 30;
			DELETE FROM logs WHERE created_at < NOW() - INTERVAL '7 days';
			SELECT * FROM (SELECT name FROM sub_table WHERE id > 10) AS subquery
			JOIN another_table ON subquery.id = another_table.id;`,
			wantReplaced: `SELECT a.id, b.name FROM users_replaced AS a JOIN accounts_replaced AS b ON a.id = b.user_id
			WHERE a.age > 30;
			DELETE FROM logs_replaced WHERE created_at < NOW() - INTERVAL '7 days';
			SELECT * FROM (SELECT name FROM sub_table_replaced WHERE id > 10) AS subquery
			JOIN another_table_replaced ON subquery.id = another_table_replaced.id;`,
		},
		{
			name:              "database,backquote,select,update,delete,join,table alias,sub query",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin: "SELECT a.id, b.name FROM `db1`.`users` AS a JOIN `accounts` AS b ON a.id = b.user_id\n\t\t\t" +
				"WHERE a.age > 30;\n\t\t\t" +
				"DELETE FROM logs WHERE created_at < NOW() - INTERVAL '7 days';\n\t\t\t" +
				"SELECT * FROM (SELECT name FROM `db2`.`sub_table` WHERE id > 10) AS subquery\n\t\t\t" +
				"JOIN another_table ON subquery.id = another_table.id;",
			wantReplaced: "SELECT a.id, b.name FROM `db1`.`users_replaced` AS a JOIN `accounts_replaced` AS b ON a.id = b.user_id\n\t\t\t" +
				"WHERE a.age > 30;\n\t\t\t" +
				"DELETE FROM logs_replaced WHERE created_at < NOW() - INTERVAL '7 days';\n\t\t\t" +
				"SELECT * FROM (SELECT name FROM `db2`.`sub_table_replaced` WHERE id > 10) AS subquery\n\t\t\t" +
				"JOIN another_table_replaced ON subquery.id = another_table_replaced.id;",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReplaced := replaceFromJoinTableName(tt.tableShardingOpts, tt.origin)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
		})
	}
}

func Test_doTableSharding_replaceInsertTableName(t *testing.T) {

	tests := []struct {
		name              string
		tableShardingOpts []TableShardingOpt
		origin            string
		wantReplaced      string
	}{
		{
			name:              "simple insert",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            " insert   into  table_1 value (1,'aa','bc');",
			wantReplaced:      " insert   into  table_1_replaced value (1,'aa','bc');",
		},
		{
			name:              "insert with database name and backquote",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "insert   into  `hcm`.`account` value (1,'aa','bc');",
			wantReplaced:      "insert   into  `hcm`.`account_replaced` value (1,'aa','bc');",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReplaced := replaceInsertTableName(tt.tableShardingOpts, tt.origin)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
		})
	}
}

func Test_doTableSharding_replaceUpdateTableName(t *testing.T) {

	tests := []struct {
		name              string
		tableShardingOpts []TableShardingOpt
		origin            string
		wantReplaced      string
	}{
		{
			name:              "simple update",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "UPDATE t1 SET salary = salary  WHERE department = 'HR';\n\t\t\t",
			wantReplaced:      "UPDATE t1_replaced SET salary = salary  WHERE department = 'HR';\n\t\t\t",
		},
		{
			name:              "simple update with db name and backquote  ",
			tableShardingOpts: []TableShardingOpt{&AppendSuffixOpt{"_replaced"}},
			origin:            "  UPDate    `db1`.`table_1` SET f1 = f1 * 1.1 WHERE table_1.department = 'd';",
			wantReplaced:      "  UPDate    `db1`.`table_1_replaced` SET f1 = f1 * 1.1 WHERE table_1_replaced.department = 'd';",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotReplaced := replaceUpdateTableName(tt.tableShardingOpts, tt.origin)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
		})
	}
}

func NewTestInjectTenantIDOpt(tenantID string, enabledTenant bool) *InjectTenantIDOpt {
	return &InjectTenantIDOpt{tenantID: tenantID, enabledTenant: enabledTenant}
}

// Test_TenantID_injectInsertTenantID 测试 TenantID - INSERT.
func Test_TenantID_injectInsertTenantID(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		tenantID     string
		enableTenant bool
		origin       string
		wantReplaced string
		wantArgs     map[string]interface{}
	}{
		{
			name:         "简单INSERT语句",
			args:         map[string]interface{}{"name": "test"},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "INSERT INTO cvm (name) VALUES(:name)",
			wantReplaced: "INSERT INTO cvm (name, tenant_id) VALUES(:name, :tenant_id)",
			wantArgs:     map[string]interface{}{"name": "test", "tenant_id": "tenant-1"},
		},
		{
			name:         "带多列的INSERT语句",
			args:         map[string]interface{}{"name": "test", "age": 20},
			tenantID:     "tenant-2",
			enableTenant: true,
			origin:       "INSERT INTO db1.cvm (name, age) VALUES (:name, :age),(:name, :age)",
			wantReplaced: "INSERT INTO db1.cvm (name, age, tenant_id) VALUES(:name, :age, :tenant_id),(:name, :age, :tenant_id)",
			wantArgs:     map[string]interface{}{"name": "test", "age": 20, "tenant_id": "tenant-2"},
		},
		{
			name:         "带数据库名+引号的INSERT语句",
			args:         map[string]interface{}{"name": "test"},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "INSERT INTO db1.`cvm` (name) VALUES (:name)",
			wantReplaced: "INSERT INTO db1.`cvm` (name, tenant_id) VALUES(:name, :tenant_id)",
			wantArgs:     map[string]interface{}{"name": "test", "tenant_id": "tenant-3"},
		},
		{
			name:         "HCM用到的INSERT-场景",
			args:         map[string]interface{}{"name": "test"},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "INSERT INTO main_account (id,name) VALUES(\"00000001\",\"name1\"),(\"00000002\",\"name2\")",
			wantReplaced: "INSERT INTO main_account (id,name, tenant_id) VALUES(\"00000001\",\"name1\", :tenant_id),(\"00000002\",\"name2\", :tenant_id)",
			wantArgs:     map[string]interface{}{"name": "test", "tenant_id": "tenant-3"},
		},
		{
			name:         "带数据库名的INSERT语句-SELECT场景-支持开启多租户才可以注入tenant_id",
			args:         map[string]interface{}{"id": "00000001", "name": "test"},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "INSERT INTO db1.`cvm` (id,name) SELECT id,name FROM disk WHERE id=1",
			wantReplaced: "INSERT INTO db1.`cvm` (id,name, tenant_id) SELECT id,name,tenant_id FROM disk WHERE tenant_id = :tenant_id AND id=1",
			wantArgs:     map[string]interface{}{"id": "00000001", "name": "test", "tenant_id": "tenant-3"},
		},
		{
			name:         "带数据库名的INSERT语句-SELECT场景-不支持开启多租户无法注入tenant_id",
			args:         map[string]interface{}{"id": "00000001", "name": "test"},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "INSERT INTO db1.`cvm` (id,name) SELECT id,name FROM no_tenant_table WHERE id=1",
			wantReplaced: "INSERT INTO db1.`cvm` (id,name) SELECT id,name FROM no_tenant_table WHERE id=1",
			wantArgs:     map[string]interface{}{"id": "00000001", "name": "test"},
		},
		{
			name:         "带数据库名的INSERT语句-SELECT场景-不支持开启多租户无法注入tenant_id",
			args:         map[string]interface{}{"id": "00000001", "name": "test"},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "INSERT INTO db1.`cvm` (id,name) SELECT id,name FROM no_tenant_table WHERE id=1",
			wantReplaced: "INSERT INTO db1.`cvm` (id,name) SELECT id,name FROM no_tenant_table WHERE id=1",
			wantArgs:     map[string]interface{}{"id": "00000001", "name": "test"},
		},
		{
			name:         "Test-CVM",
			args:         map[string]interface{}{"id": "00000001", "name": "test"},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "INSERT INTO cvm (id, cloud_id, name, creator, reviser, created_at, updated_at) VALUES(:id, :cloud_id, :name, :creator, :reviser, now(), now())",
			wantReplaced: "INSERT INTO cvm (id, cloud_id, name, creator, reviser, created_at, updated_at, tenant_id) VALUES(:id, :cloud_id, :name, :creator, :reviser, now(), now(), :tenant_id)",
			wantArgs:     map[string]interface{}{"id": "00000001", "name": "test", "tenant_id": "tenant-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReplaced, gotArgs := NewTestInjectTenantIDOpt(tt.tenantID, tt.enableTenant).InjectInsertSQL(tt.origin, tt.args)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

// Test_TenantID_InjectUpdateTenantID 测试 TenantID - UPDATE.
func Test_TenantID_InjectUpdateTenantID(t *testing.T) {
	tests := []struct {
		name         string
		origin       string
		args         map[string]interface{}
		tenantID     string
		enableTenant bool
		wantReplaced string
		wantArgs     map[string]interface{}
	}{
		{
			name:         "带WHERE子句的SQL",
			args:         map[string]interface{}{"name": "test", "id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "UPDATE cvm SET name = :name WHERE id = :id",
			wantReplaced: "UPDATE cvm SET name = :name WHERE tenant_id = :tenant_id AND id = :id",
			wantArgs:     map[string]interface{}{"name": "test", "id": 1, "tenant_id": "tenant-1"},
		},
		{
			name:         "不带WHERE子句的SQL",
			args:         map[string]interface{}{"name": "test"},
			tenantID:     "tenant-2",
			enableTenant: true,
			origin:       "UPDATE cvm SET name = :name",
			wantReplaced: "UPDATE cvm SET name = :name WHERE tenant_id = :tenant_id",
			wantArgs:     map[string]interface{}{"name": "test", "tenant_id": "tenant-2"},
		},
		{
			name:         "带数据库名称的SQL",
			args:         map[string]interface{}{"name": "test", "id": 1},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "UPDATE db1.cvm SET name = :name WHERE id = :id",
			wantReplaced: "UPDATE db1.cvm SET name = :name WHERE tenant_id = :tenant_id AND id = :id",
			wantArgs:     map[string]interface{}{"name": "test", "id": 1, "tenant_id": "tenant-3"},
		},
		{
			name:         "不支持多租户的表",
			args:         map[string]interface{}{"name": "test", "id": 1},
			tenantID:     "tenant-4",
			enableTenant: true,
			origin:       "UPDATE no_tenant_table SET name = :name WHERE id = :id",
			wantReplaced: "UPDATE no_tenant_table SET name = :name WHERE id = :id",
			wantArgs:     map[string]interface{}{"name": "test", "id": 1},
		},
		{
			name:         "空参数",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "UPDATE cvm SET name = :name",
			wantReplaced: "UPDATE cvm SET name = :name WHERE tenant_id = :tenant_id",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
		{
			name:         "空参数",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "UPDATE cvm SET locked=1, locked_cpu_core=100 where id = :id",
			wantReplaced: "UPDATE cvm SET locked=1, locked_cpu_core=100 where tenant_id = :tenant_id AND id = :id",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
		{
			name:         "HCM用到的UPDATE-场景1",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "UPDATE main_account SET name=\"test01\" WHERE name=\"test\"",
			wantReplaced: "UPDATE main_account SET name=\"test01\" WHERE tenant_id = :tenant_id AND name=\"test\"",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
		{
			name:         "HCM用到的UPDATE-场景2",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "UPDATE cvm SET state=\"success\" WHERE id=\"00000001\"",
			wantReplaced: "UPDATE cvm SET state=\"success\" WHERE tenant_id = :tenant_id AND id=\"00000001\"",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
		{
			name:         "HCM用到的UPDATE-不支持多租户的表-场景3",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "UPDATE no_tenant_table SET state=\"success\" where id = \"00000001\" and state = \"running\"",
			wantReplaced: "UPDATE no_tenant_table SET state=\"success\" where id = \"00000001\" and state = \"running\"",
			wantArgs:     map[string]interface{}{},
		},
		{
			name:         "HCM用到的UPDATE-场景4",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "UPDATE cvm SET owner=\"00000003\" WHERE res_id = \"00000001\" AND res_type = \"cvm\" AND owner = \"00000002\"",
			wantReplaced: "UPDATE cvm SET owner=\"00000003\" WHERE tenant_id = :tenant_id AND res_id = \"00000001\" AND res_type = \"cvm\" AND owner = \"00000002\"",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
		{
			name:         "HCM用到的UPDATE-场景5",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "update cvm set recycle_status = \"recovered\" where id in (\"00000001\")",
			wantReplaced: "update cvm set recycle_status = \"recovered\" where tenant_id = :tenant_id AND id in (\"00000001\")",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReplaced, gotArgs := NewTestInjectTenantIDOpt(tt.tenantID, tt.enableTenant).InjectUpdateSQL(tt.origin, tt.args)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

// Test_TenantID_InjectDeleteTenantID 测试 TenantID - DELETE.
func Test_TenantID_InjectDeleteTenantID(t *testing.T) {
	tests := []struct {
		name         string
		origin       string
		args         map[string]interface{}
		tenantID     string
		enableTenant bool
		wantReplaced string
		wantArgs     map[string]interface{}
	}{
		{
			name:         "空参数",
			args:         nil,
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "DELETE FROM cvm",
			wantReplaced: "DELETE FROM cvm WHERE tenant_id = :tenant_id",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-5"},
		},
		{
			name:         "带WHERE子句的SQL",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "DELETE FROM `cvm` WHERE id = :id",
			wantReplaced: "DELETE FROM `cvm` WHERE tenant_id = :tenant_id AND id = :id",
			wantArgs:     map[string]interface{}{"id": 1, "tenant_id": "tenant-1"},
		},
		{
			name:         "不带WHERE子句的SQL",
			args:         nil,
			tenantID:     "tenant-2",
			enableTenant: true,
			origin:       "DELETE FROM cvm",
			wantReplaced: "DELETE FROM cvm WHERE tenant_id = :tenant_id",
			wantArgs:     map[string]interface{}{"tenant_id": "tenant-2"},
		},
		{
			name:         "带数据库名称的SQL",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "DELETE FROM db1.cvm WHERE id = :id",
			wantReplaced: "DELETE FROM db1.cvm WHERE tenant_id = :tenant_id AND id = :id",
			wantArgs:     map[string]interface{}{"id": 1, "tenant_id": "tenant-3"},
		},
		{
			name:         "带数据库名称的SQL-带引号",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "DELETE FROM db1.`cvm` WHERE id = :id",
			wantReplaced: "DELETE FROM db1.`cvm` WHERE tenant_id = :tenant_id AND id = :id",
			wantArgs:     map[string]interface{}{"id": 1, "tenant_id": "tenant-3"},
		},
		{
			name:         "不支持多租户的表",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-4",
			enableTenant: true,
			origin:       "DELETE FROM no_tenant_table WHERE id = :id",
			wantReplaced: "DELETE FROM no_tenant_table WHERE id = :id",
			wantArgs:     map[string]interface{}{"id": 1},
		},
		{
			name:         "HCM使用到的场景",
			args:         map[string]interface{}{"vendor": "tcloud", "account_id": "00000001"},
			tenantID:     "tenant-4",
			enableTenant: true,
			origin:       "DELETE FROM cvm WHERE vendor = :vendor AND account_id = :account_id",
			wantReplaced: "DELETE FROM cvm WHERE tenant_id = :tenant_id AND vendor = :vendor AND account_id = :account_id",
			wantArgs:     map[string]interface{}{"vendor": "tcloud", "account_id": "00000001", "tenant_id": "tenant-4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReplaced, gotArgs := NewTestInjectTenantIDOpt(tt.tenantID, tt.enableTenant).InjectDeleteSQL(tt.origin, tt.args)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

// TestInjectJoinTenantID_BasicQueries 测试 TenantID - SELECT - 简单场景的测试用例
func TestInjectJoinTenantID_BasicQueries(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		tenantID     string
		enableTenant bool
		origin       string
		wantReplaced string
	}{
		{
			name:         "简单SELECT查询",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "select * from cvm",
			wantReplaced: "select * from cvm WHERE cvm.tenant_id = :tenant_id",
		},
		{
			name:         "简单SELECT查询-表名带引号",
			args:         map[string]interface{}{"id": 2},
			tenantID:     "tenant-2",
			enableTenant: true,
			origin:       "select * from `cvm` cv",
			wantReplaced: "select * from `cvm` cv WHERE cv.tenant_id = :tenant_id",
		},
		{
			name:         "带WHERE条件的SELECT查询",
			args:         map[string]interface{}{"id": 3},
			tenantID:     "tenant-3",
			enableTenant: true,
			origin:       "select * from `cvm` as cv where id = :id",
			wantReplaced: "select * from `cvm` as cv where cv.tenant_id = :tenant_id AND id = :id",
		},
		{
			name:         "带数据库名的SELECT",
			args:         map[string]interface{}{"id": 9},
			tenantID:     "tenant-9",
			enableTenant: true,
			origin:       "select * from db1.cvm cv",
			wantReplaced: "select * from db1.cvm cv WHERE cv.tenant_id = :tenant_id",
		},
		{
			name:         "带数据库名的SELECT-表名带引号",
			args:         map[string]interface{}{"id": 10},
			tenantID:     "tenant-10",
			enableTenant: true,
			origin:       "select * from db1.`cvm` as cv order by cv.id",
			wantReplaced: "select * from db1.`cvm` as cv  WHERE cv.tenant_id = :tenant_id order by cv.id",
		},
		{
			name:         "带数据库名的SELECT-数据库名+表名都带引号",
			args:         map[string]interface{}{"id": 10},
			tenantID:     "tenant-10",
			enableTenant: true,
			origin:       "select * from `db1`.`cvm` as cv",
			wantReplaced: "select * from `db1`.`cvm` as cv WHERE cv.tenant_id = :tenant_id",
		},
		{
			name:         "带JOIN的SELECT查询-带AS",
			args:         map[string]interface{}{"id": 4},
			tenantID:     "tenant-4",
			enableTenant: true,
			origin:       "select * from network_interface AS ni join network_interface_cvm_rel AS rel on ni.id = rel.network_interface_id inner join table AS tb on ni.id = tb.network_interface_id where rel.id=1",
			wantReplaced: "select * from network_interface AS ni join network_interface_cvm_rel AS rel on ni.id = rel.network_interface_id inner join table AS tb on ni.id = tb.network_interface_id where ni.tenant_id = :tenant_id AND rel.id=1",
		},
		{
			name:         "带JOIN的SELECT查询-不带AS",
			args:         map[string]interface{}{"id": 5},
			tenantID:     "tenant-5",
			enableTenant: true,
			origin:       "select * from network_interface ni join network_interface_cvm_rel rel on ni.id = rel.network_interface_id where rel.id=1",
			wantReplaced: "select * from network_interface ni join network_interface_cvm_rel rel on ni.id = rel.network_interface_id where ni.tenant_id = :tenant_id AND rel.id=1",
		},
		{
			name:         "带子查询的SELECT-不带WHERE",
			args:         map[string]interface{}{"id": 6},
			tenantID:     "tenant-6",
			enableTenant: true,
			origin:       "select * from (select * from `cvm` where id > :id) as sub",
			wantReplaced: "select * from (select * from `cvm` where cvm.tenant_id = :tenant_id AND id > :id) as sub",
		},
		{
			name:         "带子查询的SELECT-带WHERE",
			args:         map[string]interface{}{"id": 7},
			tenantID:     "tenant-7",
			enableTenant: true,
			origin:       "select * from (select * from cvm where id > :id) as sub where id=1 and name=\"test\"",
			wantReplaced: "select * from (select * from cvm where cvm.tenant_id = :tenant_id AND id > :id) as sub where id=1 and name=\"test\"",
		},
		{
			name:         "带子查询的SELECT+JOIN不支持多租户的表",
			args:         map[string]interface{}{"id": 8},
			tenantID:     "tenant-8",
			enableTenant: true,
			origin:       "select * from (select * from cvm where id > :id) as sub join table AS ni on sub.id = ni.id where ni.id=1",
			wantReplaced: "select * from (select * from cvm where cvm.tenant_id = :tenant_id AND id > :id) as sub join table AS ni on sub.id = ni.id where ni.id=1",
		},
		{
			name:         "带子查询的SELECT+JOIN支持多租户的表",
			args:         map[string]interface{}{"id": 8},
			tenantID:     "tenant-8",
			enableTenant: true,
			origin:       "select * from (select * from cvm where id > :id) as sub left join disk AS dk on sub.id = dk.id inner join table AS tb on sub.id = tb.id where dk.id=1",
			wantReplaced: "select * from (select * from cvm where cvm.tenant_id = :tenant_id AND id > :id) as sub left join disk AS dk on sub.id = dk.id inner join table AS tb on sub.id = tb.id where dk.tenant_id = :tenant_id AND dk.id=1",
		},
	}

	runInjectTenantIDTest(t, tests)
}

// TestInjectJoinTenantID_HCMQueries 测试 TenantID - SELECT - HCM场景的测试用例
func TestInjectJoinTenantID_HCMQueries(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		tenantID     string
		enableTenant bool
		origin       string
		wantReplaced string
	}{
		{
			name:         "HCM用到的SELECT-场景1",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT * FROM main_account WHERE id=\"00000001\" ORDER BY id ASC LIMIT 0, 10",
			wantReplaced: "SELECT * FROM main_account WHERE main_account.tenant_id = :tenant_id AND id=\"00000001\" ORDER BY id ASC LIMIT 0, 10",
		},
		{
			name:         "HCM用到的SELECT-场景2",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT COUNT(DISTINCT vendor) FROM root_account WHERE vendor=\"tcloud\"",
			wantReplaced: "SELECT COUNT(DISTINCT vendor) FROM root_account WHERE root_account.tenant_id = :tenant_id AND vendor=\"tcloud\"",
		},
		{
			name:         "HCM用到的SELECT-场景3",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT DISTINCT vendor FROM root_account WHERE vendor=\"tcloud\" ORDER BY id ASC LIMIT 0, 10",
			wantReplaced: "SELECT DISTINCT vendor FROM root_account WHERE root_account.tenant_id = :tenant_id AND vendor=\"tcloud\" ORDER BY id ASC LIMIT 0, 10",
		},
		{
			name:         "HCM用到的SELECT-场景4",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT id FROM cvm WHERE bill_year=2025 AND bill_month=3 ORDER BY id ASC LIMIT 0, 10",
			wantReplaced: "SELECT id FROM cvm WHERE cvm.tenant_id = :tenant_id AND bill_year=2025 AND bill_month=3 ORDER BY id ASC LIMIT 0, 10",
		},
		{
			name:         "HCM用到的SELECT-场景5",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT * FROM cvm WHERE id IN (\"00000001\",\"00000002\")",
			wantReplaced: "SELECT * FROM cvm WHERE cvm.tenant_id = :tenant_id AND id IN (\"00000001\",\"00000002\")",
		},
		{
			name:         "HCM用到的SELECT-场景6",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT ANY_VALUE(currency) currency,COUNT(*) AS count, SUM(COST) AS cost  FROM cvm WHERE bill_year=2025 AND bill_month=3",
			wantReplaced: "SELECT ANY_VALUE(currency) currency,COUNT(*) AS count, SUM(COST) AS cost  FROM cvm WHERE cvm.tenant_id = :tenant_id AND bill_year=2025 AND bill_month=3",
		},
		{
			name:         "HCM用到的SELECT-场景7",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT bk_biz_id, SUM(last_month_cost_synced) as last_month_cost_synced, SUM(last_month_rmb_cost_synced) as last_month_rmb_cost_synced, SUM(current_month_cost_synced) as current_month_cost_synced, SUM(current_month_rmb_cost_synced) as current_month_rmb_cost_synced, SUM(current_month_cost) as current_month_cost, SUM(current_month_rmb_cost) as current_month_rmb_cost, SUM(adjustment_cost) as adjustment_cost, SUM(adjustment_rmb_cost) as adjustment_rmb_cost FROM cvm WHERE vendor=\"tcloud\" group by bk_biz_id ORDER BY bk_biz_id ASC LIMIT 0, 10",
			wantReplaced: "SELECT bk_biz_id, SUM(last_month_cost_synced) as last_month_cost_synced, SUM(last_month_rmb_cost_synced) as last_month_rmb_cost_synced, SUM(current_month_cost_synced) as current_month_cost_synced, SUM(current_month_rmb_cost_synced) as current_month_rmb_cost_synced, SUM(current_month_cost) as current_month_cost, SUM(current_month_rmb_cost) as current_month_rmb_cost, SUM(adjustment_cost) as adjustment_cost, SUM(adjustment_rmb_cost) as adjustment_rmb_cost FROM cvm WHERE cvm.tenant_id = :tenant_id AND vendor=\"tcloud\" group by bk_biz_id ORDER BY bk_biz_id ASC LIMIT 0, 10",
		},
		{
			name:         "HCM用到的SELECT-场景8",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT * FROM cvm where id in (\"00000001\")",
			wantReplaced: "SELECT * FROM cvm where cvm.tenant_id = :tenant_id AND id in (\"00000001\")",
		},
		{
			name:         "HCM用到的SELECT-场景9",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT count(distinct(cvm.id)) FROM cvm as cvm left join disk_cvm_rel as rel on cvm.id = rel.cvm_id WHERE cvm.vendor=\"tcloud\" and disk_id != \"00000001\" and rel.cvm_id is NULL",
			wantReplaced: "SELECT count(distinct(cvm.id)) FROM cvm as cvm left join disk_cvm_rel as rel on cvm.id = rel.cvm_id WHERE cvm.tenant_id = :tenant_id AND cvm.vendor=\"tcloud\" and disk_id != \"00000001\" and rel.cvm_id is NULL",
		},
		{
			name:         "HCM用到的SELECT-场景10",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT cvm.id as id, name,vendor FROM cvm as cvm left join disk_cvm_rel as rel on cvm.id = rel.cvm_id WHERE cvm.vendor=\"tcloud\" group by cvm.id ORDER BY id ASC LIMIT 0, 10",
			wantReplaced: "SELECT cvm.id as id, name,vendor FROM cvm as cvm left join disk_cvm_rel as rel on cvm.id = rel.cvm_id WHERE cvm.tenant_id = :tenant_id AND cvm.vendor=\"tcloud\" group by cvm.id ORDER BY id ASC LIMIT 0, 10",
		},
		{
			name:         "HCM用到的SELECT-场景11",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT vendor,account_id,disk.id as id,rel.cvm_id as cvm_id FROM disk_cvm_rel as rel left join disk as disk on rel.disk_id = disk.id where cvm_id in (\"00000001\")",
			wantReplaced: "SELECT vendor,account_id,disk.id as id,rel.cvm_id as cvm_id FROM disk_cvm_rel as rel left join disk as disk on rel.disk_id = disk.id where disk.tenant_id = :tenant_id AND cvm_id in (\"00000001\")",
		},
		{
			name:         "HCM用到的SELECT-场景12",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT COUNT(*) FROM network_interface AS ni LEFT JOIN network_interface_cvm_rel AS rel ON rel.network_interface_id = ni.id WHERE vendor=\"tcloud\"",
			wantReplaced: "SELECT COUNT(*) FROM network_interface AS ni LEFT JOIN network_interface_cvm_rel AS rel ON rel.network_interface_id = ni.id WHERE ni.tenant_id = :tenant_id AND vendor=\"tcloud\"",
		},
		{
			name:         "HCM用到的SELECT-场景13",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT account_id,vendor,name,ni.id as id,rel.cvm_id as cvm_id FROM network_interface AS ni LEFT JOIN network_interface_cvm_rel AS rel ON rel.network_interface_id = ni.id WHERE vendor=\"tcloud\" ORDER BY id ASC LIMIT 0, 10",
			wantReplaced: "SELECT account_id,vendor,name,ni.id as id,rel.cvm_id as cvm_id FROM network_interface AS ni LEFT JOIN network_interface_cvm_rel AS rel ON rel.network_interface_id = ni.id WHERE ni.tenant_id = :tenant_id AND vendor=\"tcloud\" ORDER BY id ASC LIMIT 0, 10",
		},
		{
			name:         "HCM用到的SELECT-场景14",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT cvm_id,network_interface_id,ni.id as id,rel.cvm_id as cvm_id FROM network_interface_cvm_rel AS rel LEFT JOIN network_interface AS ni ON rel.network_interface_id = ni.id WHERE rel.cvm_id IN (\"00000001\") AND ni.vendor = \"tcloud\"",
			wantReplaced: "SELECT cvm_id,network_interface_id,ni.id as id,rel.cvm_id as cvm_id FROM network_interface_cvm_rel AS rel LEFT JOIN network_interface AS ni ON rel.network_interface_id = ni.id WHERE ni.tenant_id = :tenant_id AND rel.cvm_id IN (\"00000001\") AND ni.vendor = \"tcloud\"",
		},
		{
			name:         "HCM用到的SELECT-场景15",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT vendor,cloud_id, sg.id as id,rel.res_id as res_id, sg.vendor AS vendor,sg.reviser AS reviser,sg.updated_at AS updated_at, rel.res_type,rel.priority FROM security_group_common_rel AS rel LEFT JOIN security_group AS sg ON rel.security_group_id = sg.id WHERE res_id IN (\"00000001\") AND res_type = \"cvm\"",
			wantReplaced: "SELECT vendor,cloud_id, sg.id as id,rel.res_id as res_id, sg.vendor AS vendor,sg.reviser AS reviser,sg.updated_at AS updated_at, rel.res_type,rel.priority FROM security_group_common_rel AS rel LEFT JOIN security_group AS sg ON rel.security_group_id = sg.id WHERE sg.tenant_id = :tenant_id AND res_id IN (\"00000001\") AND res_type = \"cvm\"",
		},
		{
			name:         "HCM用到的SELECT-场景16",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "select table_name as name from information_schema.columns where column_name = \"account_id\"",
			wantReplaced: "select table_name as name from information_schema.columns where column_name = \"account_id\"",
		},
		{
			name:         "HCM用到的SELECT-场景17",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT vendor,name, account.id as id,rel.bk_biz_id as bk_biz_id FROM account_biz_rel AS rel LEFT JOIN account AS account ON rel.account_id = account.id WHERE rel.bk_biz_id in (213)",
			wantReplaced: "SELECT vendor,name, account.id as id,rel.bk_biz_id as bk_biz_id FROM account_biz_rel AS rel LEFT JOIN account AS account ON rel.account_id = account.id WHERE account.tenant_id = :tenant_id AND rel.bk_biz_id in (213)",
		},
		{
			name:         "HCM用到的SELECT-场景18",
			args:         map[string]interface{}{"id": 1},
			tenantID:     "tenant-1",
			enableTenant: true,
			origin:       "SELECT account_id as group_field, COUNT(*) as count FROM subnet WHERE vendor=\"tcloud\" GROUP BY account_id",
			wantReplaced: "SELECT account_id as group_field, COUNT(*) as count FROM subnet WHERE subnet.tenant_id = :tenant_id AND vendor=\"tcloud\" GROUP BY account_id",
		},
	}

	runInjectTenantIDTest(t, tests)
}

func runInjectTenantIDTest(t *testing.T, testCases []struct {
	name         string
	args         map[string]interface{}
	tenantID     string
	enableTenant bool
	origin       string
	wantReplaced string
}) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotReplaced, gotArgs := NewTestInjectTenantIDOpt(tt.tenantID, tt.enableTenant).
				InjectJoinSQL(tt.origin, tt.args)
			assert.Equal(t, tt.wantReplaced, gotReplaced)
			assert.Equal(t, tt.tenantID, gotArgs[constant.TenantIDField])
		})
	}
}

type AppendSuffixOpt struct {
	Suffix string
}

func (r *AppendSuffixOpt) Match(name string) bool {
	return true
}

func (r *AppendSuffixOpt) ReplaceTableName(old string) string {
	if len(r.Suffix) > 0 {
		return old + r.Suffix
	}
	return old
}
