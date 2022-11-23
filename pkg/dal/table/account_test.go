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

package table

import (
	"fmt"
	"testing"
)

func TestInsertSql(t *testing.T) {
	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, AccountTable, AccountColumns.ColumnExpr(),
		AccountColumns.ColonNameExpr())

	insertSql := "INSERT INTO account (name, memo, creator, reviser, created_at, updated_at)	" +
		"VALUES(:spec.name, :spec.memo, :revision.creator, :revision.reviser, now(), now())"

	if sql != insertSql {
		t.Errorf("insert sql not right, sql: %s", sql)
		return
	}
}

func TestSelectSql(t *testing.T) {
	sql := fmt.Sprintf(`SELECT %s FROM %s`, AccountColumns.NamedExpr(), AccountTable)

	selectSql := "SELECT id, name as 'spec.name', memo as 'spec.memo', creator as 'revision.creator', reviser " +
		"as 'revision.reviser', created_at as 'revision.created_at', updated_at as 'revision.updated_at' FROM account"

	if sql != selectSql {
		t.Errorf("select sql not right, sql: %s", sql)
		return
	}
}

func TestUpdateSql(t *testing.T) {
	account := &Account{
		ID: 1,
		Spec: &AccountSpec{
			Name: "updated-account-test",
		},
		Revision: &Revision{
			Reviser: "Tom",
		},
	}

	opts := NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields("id")
	expr, _, err := RearrangeSQLDataWithOption(account, opts)
	if err != nil {
		t.Error(err)
		return
	}

	if expr != "name = :name, updated_at = now(), reviser = :reviser" {
		t.Errorf("update set expr not right, expr: %s", expr)
		return
	}
}
