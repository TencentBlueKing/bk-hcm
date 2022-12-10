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

package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"hcm/pkg/criteria/enumor"
)

func TestMergeNamedColumns(t *testing.T) {
	namedA := ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}}
	namedB := MergeColumnDescriptors("nested",
		ColumnDescriptors{
			{Column: "name", NamedC: "name", Type: enumor.String},
			{Column: "memo", NamedC: "memo", Type: enumor.String},
		},
	)

	merged := MergeColumns(nil, namedA, namedB)
	compared := MergeColumns(nil, ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "name", NamedC: "nested.name", Type: enumor.String},
		{Column: "memo", NamedC: "nested.memo", Type: enumor.String},
	})

	fmt.Println("columns: ", merged.Columns())
	if !reflect.DeepEqual(merged.Columns(), compared.Columns()) {
		t.Errorf("test merged columns failed, not equal")
		return
	}

	fmt.Println("column expr: ", merged.ColumnExpr())
	if merged.ColumnExpr() != compared.ColumnExpr() {
		t.Errorf("test merged columns expr failed, not equal")
		return
	}

	fmt.Println("named columns expr: ", merged.NamedExpr())
	if merged.NamedExpr() != compared.NamedExpr() {
		t.Errorf("test merged columns named expr failed,  not equal")
		return
	}

	fmt.Println("colon named columns expr: ", merged.ColonNameExpr())
	if merged.ColonNameExpr() != compared.ColonNameExpr() {
		t.Errorf("test merged columns named expr failed,  not equal")
		return
	}

	fmt.Println("without column: ", merged.WithoutColumn("id"))
	if !reflect.DeepEqual(merged.WithoutColumn("id"), map[string]enumor.ColumnType{
		"name": enumor.String,
		"memo": enumor.String,
	}) {
		t.Errorf("test merged without columns failed,  not equal")
		return
	}
}

type deepEmbedded struct {
	Age      int       `db:"age"`
	Birthday time.Time `db:"birthday"`
}

type embedded struct {
	Name string `db:"name"`
	// test not pointer struct case.
	Deep deepEmbedded `db:"deep"`
}

type cases struct {
	// test flat field
	ID int `db:"id"`
	// test embedded struct
	Embedded *embedded `db:"embedded"`
	// test interface
	Iter interface{} `db:"iter"`
	// test time.
	Time time.Time `db:"time"`
}

func TestRearrangeSQLDataWithOptionFully(t *testing.T) {
	c := cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	opts := NewFieldOptions().
		AddBlankedFields("age").
		AddIgnoredFields("id")
	expr, toUpdate, err := RearrangeSQLDataWithOption(c, opts)
	if err != nil {
		t.Errorf("parse data field, err: %v", err)
		return
	}

	// validate result
	if strings.Contains(expr, ":id") {
		t.Errorf("test ignored field failed")
		return
	}

	if !strings.Contains(expr, ":name") {
		t.Errorf("test embedded *struct field failed")
		return
	}

	if !strings.Contains(expr, ":age") {
		t.Errorf("test deep embedded *struct and blanked field failed")
		return
	}

	if strings.Contains(expr, ":birthday") {
		t.Errorf("test deep embedded *struct and *NOT* blanked field failed")
		return
	}

	if !strings.Contains(expr, ":iter") {
		t.Errorf("test interface field failed")
		return
	}

	if !strings.Contains(expr, ":time") {
		t.Errorf("test not blanked time field failed")
		return
	}

	fmt.Println("expr: ", expr)
	js, _ := json.MarshalIndent(toUpdate, "", "    ")
	fmt.Printf("to update: %s\n", js)

	// test result should be like this:
	// expr:  iter = :iter, time = :time, name = :name, age = :age
	// to update: {
	//    "age": 0,
	//    "birthday": "0001-01-01T00:00:00Z",
	//    "id": 20,
	//    "iter": "123456789",
	//    "name": "demo",
	//    "time": "2022-01-02T11:00:53.692535+08:00"
	// }
}

func TestRecursiveGetTaggedFieldValues(t *testing.T) {
	c := cases{
		// test ignored field
		ID: 20,
		Embedded: &embedded{
			Name: "demo",
			Deep: deepEmbedded{
				// test blanked field
				Age: 0,
				// test not blanked field.
				Birthday: time.Time{},
			},
		},
		Iter: "123456789",
		Time: time.Now(),
	}

	kv, err := RecursiveGetTaggedFieldValues(c)
	if err != nil {
		t.Errorf("get kv failed, err: %v", err)
		return
	}

	js, _ := json.MarshalIndent(kv, "", "    ")
	fmt.Printf("kv json: %s\n", js)
}
