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

import (
	"testing"

	corecvm "hcm/pkg/api/core/cloud/cvm"
)

func TestOnShelfDate(t *testing.T) {
	cases := []struct {
		name string
		host corecvm.BaseCvm
		want string
	}{
		{
			name: "date only",
			host: corecvm.BaseCvm{
				CloudLaunchedTime: "2026-02-02",
			},
			want: "2026-02-02",
		},
		{
			name: "rfc3339 datetime",
			host: corecvm.BaseCvm{
				CloudLaunchedTime: "2026-03-03T01:02:03Z",
			},
			want: "2026-03-03",
		},
		{
			name: "datetime layout",
			host: corecvm.BaseCvm{
				CloudLaunchedTime: "2026-04-04 12:13:14",
			},
			want: "2026-04-04",
		},
		{
			name: "invalid datetime fallback",
			host: corecvm.BaseCvm{
				CloudLaunchedTime: "2026-05-05-invalid",
			},
			want: "2026-05-05",
		},
		{name: "both empty", host: corecvm.BaseCvm{}, want: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := getOnShelfDate(c.host); got != c.want || err != nil {
				t.Errorf("onShelfDate(%+v) = %q, want %q", c.host, got, c.want)
			}
		})
	}
}
