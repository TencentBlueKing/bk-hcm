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

package lblogic

import (
	"testing"

	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/assert"
)

func TestCreateLayer7ListenerExecutor_convertDataToPreview(t *testing.T) {
	type args struct {
		i [][]string
	}
	tests := []struct {
		name string
		args args
		want CreateLayer7ListenerDetail
	}{
		{
			name: "HTTPs test",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1",
					"https", "8888", "MUTUAL", "[9GXQ9dV2,DQq54hR3]", "Bw0pFuKG", "用户的备注"},
			}},
			want: CreateLayer7ListenerDetail{
				ClbVipDomain:  "127.0.0.1",
				CloudClbID:    "lb-xxxxx1",
				Protocol:      enumor.HttpsProtocol,
				ListenerPorts: []int{8888},
				SSLMode:       "MUTUAL",
				CertCloudIDs:  []string{"9GXQ9dV2", "DQq54hR3"},
				CACloudID:     "Bw0pFuKG",

				UserRemark:     "用户的备注",
				Status:         "",
				ValidateResult: []string{},
			},
		},
		{
			name: "HTTP test",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "http", "8888"},
			}},
			want: CreateLayer7ListenerDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.HttpProtocol,
				ListenerPorts:  []int{8888},
				SSLMode:        "",
				CertCloudIDs:   nil,
				CACloudID:      "",
				UserRemark:     "",
				Status:         "",
				ValidateResult: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateLayer7ListenerPreviewExecutor{}
			_ = executor.convertDataToPreview(tt.args.i)
			assert.Equal(t, tt.want, *executor.details[0])
		})
	}
}

func TestCreateLayer7ListenerDetail_validate(t *testing.T) {
	tests := []struct {
		name       string
		args       *CreateLayer7ListenerDetail
		wantStatus ImportStatus
	}{
		{
			name: "validate protocol executable",
			args: &CreateLayer7ListenerDetail{
				Protocol:      enumor.HttpsProtocol,
				ListenerPorts: []int{8888, 8889},
				SSLMode:       "MUTUAL",
				CertCloudIDs:  []string{"9GXQ9dV2", "DQq54hR3"},
				CACloudID:     "Bw0pFuKG",
			},
			wantStatus: Executable,
		},
		{
			name: "validate protocol not executable",
			args: &CreateLayer7ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889},
				SSLMode:       "MUTUAL",
				CertCloudIDs:  []string{"9GXQ9dV2", "DQq54hR3"},
				CACloudID:     "Bw0pFuKG",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate ListenerPorts has 3 port",
			args: &CreateLayer7ListenerDetail{
				Protocol:      enumor.HttpsProtocol,
				ListenerPorts: []int{8888, 8889, 9000},
				SSLMode:       "MUTUAL",
				CertCloudIDs:  []string{"9GXQ9dV2", "DQq54hR3"},
				CACloudID:     "Bw0pFuKG",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate http with cert info",
			args: &CreateLayer7ListenerDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPorts: []int{8888, 8889},
				SSLMode:       "MUTUAL",
				CertCloudIDs:  []string{"9GXQ9dV2", "DQq54hR3"},
				CACloudID:     "Bw0pFuKG",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate http with empty cert info",
			args: &CreateLayer7ListenerDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: Executable,
		},
		{
			name: "validate https with empty cert info",
			args: &CreateLayer7ListenerDetail{
				Protocol:      enumor.HttpsProtocol,
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.validate()
			assert.Equal(t, tt.wantStatus, tt.args.Status)
		})
	}
}
