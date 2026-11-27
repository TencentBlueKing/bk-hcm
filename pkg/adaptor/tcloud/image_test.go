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

package tcloud

import (
	"testing"

	"hcm/pkg/criteria/constant"

	"github.com/stretchr/testify/require"
)

func TestChangeArchitecture(t *testing.T) {
	arm := "arm"
	x86 := "x86_64"
	other := "other"

	tests := []struct {
		name         string
		architecture *string
		want         string
	}{
		{name: "nil defaults to x86_64", architecture: nil, want: constant.X86},
		{name: "arm maps to arm64", architecture: &arm, want: constant.Arm64},
		{name: "x86_64 stays unchanged", architecture: &x86, want: constant.X86},
		{name: "other value stays unchanged", architecture: &other, want: other},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, changeArchitecture(tt.architecture))
		})
	}
}
