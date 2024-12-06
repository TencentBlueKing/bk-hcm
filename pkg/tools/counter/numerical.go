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

package counter

import "strconv"

// NewNumStringCounter counter in string format
func NewNumStringCounter(init, base int) func() string {
	return func() (r string) {
		r = strconv.FormatInt(int64(init), base)
		init++
		return r
	}
}

// NewNumberCounter decimal counter in int format
func NewNumberCounter(init int) func() int {
	return func() (r int) {
		r = init
		init++
		return r
	}
}

// NewNumberCounterWithPrev decimal counter in int format with previous value
func NewNumberCounterWithPrev(init, base int) func() (cur string, prev string) {
	current := init
	return func() (cur string, prev string) {
		cur = strconv.FormatInt(int64(current), base)
		if current-1 >= init {
			prev = strconv.FormatInt(int64(current-1), base)
		}
		current++
		return cur, prev
	}
}
