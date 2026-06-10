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

package cmdb

import "testing"

func TestOnShelfDate(t *testing.T) {
	cases := []struct {
		name             string
		cloudCreatedTime string
		cloudLaunched    string
		want             string
	}{
		{name: "created time present", cloudCreatedTime: "2026-01-01", cloudLaunched: "2026-02-02", want: "2026-01-01"},
		{name: "fallback to launched time", cloudCreatedTime: "", cloudLaunched: "2026-02-02", want: "2026-02-02"},
		{name: "both empty", cloudCreatedTime: "", cloudLaunched: "", want: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := onShelfDate(c.cloudCreatedTime, c.cloudLaunched); got != c.want {
				t.Errorf("onShelfDate(%q, %q) = %q, want %q", c.cloudCreatedTime, c.cloudLaunched, got, c.want)
			}
		})
	}
}
