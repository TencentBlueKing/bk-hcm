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
