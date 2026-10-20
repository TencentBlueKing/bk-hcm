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

package application

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRemoveSenseField 含 password 的顶层字段被移除，其余字段原样保留。
func TestRemoveSenseField(t *testing.T) {
	content := `{"name":"vm-1","password":"secret","root_password":"x","nested":{"password":"keep"},"count":2}`

	got := RemoveSenseField(content)

	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(got), &m))
	require.NotContains(t, m, "password")
	require.NotContains(t, m, "root_password")
	require.JSONEq(t, `"vm-1"`, string(m["name"]))
	require.JSONEq(t, `2`, string(m["count"]))
	// 只处理顶层字段，嵌套对象内容不动
	require.JSONEq(t, `{"password":"keep"}`, string(m["nested"]))
}

// TestRemoveSenseField_Empty 内容为空时返回空对象。
func TestRemoveSenseField_Empty(t *testing.T) {
	require.Equal(t, "{}", RemoveSenseField(""))
	require.Equal(t, "{}", RemoveSenseField(`{"password":"secret"}`))
}

// TestRebuildContent_ReplaceAndAppend 丢弃的字段被移除，追加的字段覆盖同名字段，其余字段保留。
func TestRebuildContent_ReplaceAndAppend(t *testing.T) {
	content := `{"a":1,"clusters":["old"],"b":"x"}`

	got := rebuildContent(content, func(key string) bool { return key == "clusters" },
		map[string]string{"clusters": `[{"cloud_cluster_id":"tgw-1"}]`})

	require.JSONEq(t, `{"a":1,"b":"x","clusters":[{"cloud_cluster_id":"tgw-1"}]}`, got)
}

// TestRebuildContent_NoSkipNoExtras 没有丢弃也没有追加时内容等价不变。
func TestRebuildContent_NoSkipNoExtras(t *testing.T) {
	content := `{"a":1,"b":{"c":[1,2]}}`

	got := rebuildContent(content, func(string) bool { return false }, nil)

	require.JSONEq(t, content, got)
}

// TestRebuildContent_ExtrasOnlyIntoEmpty 内容为空时只输出追加的字段。
func TestRebuildContent_ExtrasOnlyIntoEmpty(t *testing.T) {
	got := rebuildContent("", func(string) bool { return false }, map[string]string{"clusters": "[]"})

	require.JSONEq(t, `{"clusters":[]}`, got)
}
