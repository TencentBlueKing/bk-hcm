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

// UpdateMerge 该函数将source 更新到 destination上，并返回新的json数据
// 其中source和destination都是map[string]interface{}或struct json后的字符串
func UpdateMerge(source, destination string) (mergeData string, err error) {
	// 先转为map[string]interface{}
	sourceMap := map[string]interface{}{}
	err = UnmarshalFromString(source, &sourceMap)
	if err != nil {
		return
	}

	destinationMap := map[string]interface{}{}
	err = UnmarshalFromString(destination, &destinationMap)
	if err != nil {
		return
	}

	// 遍历覆盖更新
	for k, v := range sourceMap {
		// Note: 这里不判断是否相等，因为interface如果涉及指针，比较起来麻烦
		destinationMap[k] = v
	}

	// 重新转为json string
	mergeData, err = MarshalToString(destinationMap)
	return
}
