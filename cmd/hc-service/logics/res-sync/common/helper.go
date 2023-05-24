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

package common

import "strings"

// CloudIDClassByResGroupName 将多个资源组的云ID进行分类
func CloudIDClassByResGroupName(cloudIDs []string) map[string][]string {
	resGroupNameIDMap := make(map[string][]string)
	for _, id := range cloudIDs {
		tmp := id[strings.Index(id, "resourcegroups/")+15:]
		resGroupName := tmp[:strings.Index(tmp, "/")]
		if _, exist := resGroupNameIDMap[resGroupName]; !exist {
			resGroupNameIDMap[resGroupName] = make([]string, 0)
		}

		resGroupNameIDMap[resGroupName] = append(resGroupNameIDMap[resGroupName], id)
	}

	return resGroupNameIDMap
}
