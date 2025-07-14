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

package bill

import (
	"testing"
)

func TestBase64String_checkSize(t *testing.T) {
	type args struct {
		expectedSize int
	}
	tests := []struct {
		name    string
		b       Base64String
		args    args
		wantErr bool
	}{
		{
			name: "Hello, World!",
			b:    Base64String("SGVsbG8sIHdvcmxkIQ=="),
			args: args{
				expectedSize: 1,
			},
			wantErr: true,
		},
		{
			name: "Hello, World! 2",
			b:    Base64String("SGVsbG8sIHdvcmxkIQ=="),
			args: args{
				expectedSize: 1024, // 1KB
			},
			wantErr: false,
		},
		{
			name: "Hello, World! 3",
			b:    Base64String("SGVsbG8sIHdvcmxkIQ=="),
			args: args{
				expectedSize: 20,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.checkSize(tt.args.expectedSize); (err != nil) != tt.wantErr {
				t.Errorf("checkSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
