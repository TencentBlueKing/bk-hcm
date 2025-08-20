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

package loadbalancer

import (
	"fmt"
	"testing"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/assert"
)

func Test_parseVendor(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name    string
		args    args
		want    enumor.Vendor
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "tcloud",
			args: args{
				[]string{"云厂商", constant.CLBExcelHeaderTCloud},
			},
			want:    enumor.TCloud,
			wantErr: assert.NoError,
		},
		{
			name: "no data",
			args: args{
				[]string{},
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "unsupport vendor",
			args: args{
				[]string{"云厂商", constant.CLBExcelHeaderTCloud + "123"},
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVendor(tt.args.columns)
			if !tt.wantErr(t, err, fmt.Sprintf("parseVendor(%v)", tt.args.columns)) {
				return
			}
			assert.Equalf(t, tt.want, got, "parseVendor(%v)", tt.args.columns)
		})
	}
}
