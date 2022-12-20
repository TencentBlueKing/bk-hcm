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

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// UpdateMerge 该函数将 source 覆盖更新到 destination 上，并返回新的json数据。
// 其中source是 map[string]interface{} 或 struct，destination是 map[string]interface{} 或 struct json 后的字符串。
func UpdateMerge(source interface{}, destination string) (string, error) {
	sourceJson, err := iteratorJson.Marshal(source)
	if err != nil {
		return "", err
	}

	return gjson.Get(fmt.Sprintf("[%s,%s]", destination, sourceJson), "@join").String(), nil
}
