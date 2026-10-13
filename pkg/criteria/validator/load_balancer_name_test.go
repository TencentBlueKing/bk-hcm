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

package validator

import "testing"

func TestValidateLoadBalancerName(t *testing.T) {
	tests := []struct {
		name    string
		lbName  string
		wantErr bool
	}{
		{
			name:    "valid name with dot in middle",
			lbName:  "tiyan-relaysvr1v1.osgame.qq.com-CAP1",
			wantErr: false,
		},
		{
			name:    "valid simple name",
			lbName:  "hcm-test-clb",
			wantErr: false,
		},
		{
			name:    "valid name with multiple dots",
			lbName:  "a.b.c-test1",
			wantErr: false,
		},
		{
			name:    "invalid name starts with dot",
			lbName:  ".invalid-name",
			wantErr: true,
		},
		{
			name:    "invalid name ends with dot",
			lbName:  "invalid-name.",
			wantErr: true,
		},
		{
			name:    "invalid empty name",
			lbName:  "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateLoadBalancerName(tc.lbName)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
