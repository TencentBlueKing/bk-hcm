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

package idgenerator

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"hcm/pkg/dal/table"
	"hcm/pkg/kit"

	_ "github.com/go-sql-driver/mysql" // import mysql drive, used to create conn.
	"github.com/jmoiron/sqlx"
)

func TestIDGeneratorOne(t *testing.T) {
	source := "root:admin@tcp(127.0.0.1:3306)/hcm?parseTime=true&charset=utf8mb4"
	db, err := sqlx.Connect("mysql", source)
	checkErr(err)

	idGenerator := New(db, DefaultMaxRetryCount)

	id, err := idGenerator.One(kit.New(), table.AccountTable)
	checkErr(err)

	if len(id) != 8 {
		checkErr(errors.New("id length not right"))
	}
}

func TestIDGeneratorBatch(t *testing.T) {
	source := "root:admin@tcp(127.0.0.1:3306)/hcm?parseTime=true&charset=utf8mb4"
	db, err := sqlx.Connect("mysql", source)
	checkErr(err)

	idGenerator := New(db, DefaultMaxRetryCount)

	ids, err := idGenerator.Batch(kit.New(), table.AccountTable, 98)
	checkErr(err)

	if len(ids) != 98 {
		checkErr(errors.New("ids count is not right"))
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
