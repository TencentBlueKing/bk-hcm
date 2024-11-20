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

package filter

import (
	"fmt"

	"hcm/pkg/tools/rand"
)

// fieldPlaceholderName 如果用户的查询条件中同时有两个相同的字段名的话，查询语句会出现两个相同占位符的语句，没办法将值赋进去。
// 所以，需要随机生成一个后缀，避免这个问题的出现。
// 问题语句: select * from test where time < :time and time < :time
// 最终语句：select * from test where time < :time_abcd and time < :time_bdcs
func fieldPlaceholderName(field string) string {
	return fmt.Sprintf("%s_%s", field, rand.String(4))
}
