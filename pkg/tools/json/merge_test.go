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

package json

import "testing"

func TestUpdateMerge(t *testing.T) {
	source := &struct {
		Str    string `json:"str"`
		Uint64 uint64 `json:"uint64"`
		Struct struct {
			Float float32 `json:"float"`
		} `json:"struct"`
		Map   map[string]interface{} `json:"map"`
		Bool  bool                   `json:"bool"`
		Test1 string                 `json:"test1,omitempty"`
		Test2 float64                `json:"test2,omitempty"`
	}{
		Str:    "string",
		Uint64: 100,
		Struct: struct {
			Float float32 `json:"float"`
		}{Float: 1.23},
		Map: map[string]interface{}{
			"test": "aaa",
		},
		Bool: true,
	}

	destination := `
{
	"str": "stringtest",
	"uint64": 111,
	"struct": {
		"float": 3.21
	},
	"map": {
		"key": "value"
	},
	"bool": false,
	"test1": "test1",
	"test2": 1.11
}`

	merged, err := UpdateMerge(source, destination)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	expected := `{"str":"string","uint64":100,"struct":{"float":1.23},"map":{"test":"aaa"},"bool":true,"test1":"test1","test2":1.11}`
	if merged != expected {
		t.Errorf("update merge result(%s) not matched expected", merged)
		return
	}
}
