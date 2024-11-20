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

func TestCreateLayer4ListenerExecutor_convertDataToPreview_validateFailed(t *testing.T) {
	type args struct {
		i [][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "WRR", "0", "disable（）", "用户的备注"},
			}},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateLayer4ListenerPreviewExecutor{}
			err := executor.convertDataToPreview(tt.args.i)
			tt.wantErr(t, err)
		})
	}
}

func TestCreateLayer4ListenerExecutor_convertDataToPreview(t *testing.T) {
	type args struct {
		i [][]string
	}
	tests := []struct {
		name string
		args args
		want CreateLayer4ListenerDetail
	}{
		{
			name: "test",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "WRR", "0", "disable", "用户的备注"},
			}},
			want: CreateLayer4ListenerDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.TcpProtocol,
				ListenerPorts:  []int{8888},
				Scheduler:      "WRR",
				Session:        0,
				HealthCheck:    false,
				UserRemark:     "用户的备注",
				Status:         "",
				ValidateResult: []string{},
			},
		},
		{
			name: "end_port",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "tcp", "[8888, 8889]", "WRR", "10", "enable"},
			}},
			want: CreateLayer4ListenerDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.TcpProtocol,
				ListenerPorts:  []int{8888, 8889},
				Scheduler:      "WRR",
				Session:        10,
				HealthCheck:    true,
				UserRemark:     "",
				Status:         "",
				ValidateResult: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateLayer4ListenerPreviewExecutor{}
			_ = executor.convertDataToPreview(tt.args.i)
			assert.Equal(t, tt.want, *executor.details[0])
		})
	}
}

func TestCreateLayer4ListenerDetail_validate(t *testing.T) {
	tests := []struct {
		name       string
		args       *CreateLayer4ListenerDetail
		wantStatus ImportStatus
	}{
		{
			name: "validate protocol executable",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889},
				Scheduler:     "WRR",
				Session:       30,
			},
			wantStatus: Executable,
		},
		{
			name: "validate protocol",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.HttpsProtocol,
				ListenerPorts: []int{8888, 8889},
				Scheduler:     "WRR",
				Session:       30,
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate Scheduler normal",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889},
				Scheduler:     enumor.WRR,
				Session:       30,
			},
			wantStatus: Executable,
		},
		{
			name: "validate Scheduler wrong scheduler ",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889},
				Scheduler:     "WRR2",
				Session:       30,
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate session less than 30",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889},
				Scheduler:     "WRR",
				Session:       10,
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate session normal",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889},
				Scheduler:     "WRR",
				Session:       0,
			},
			wantStatus: Executable,
		},
		{
			name: "validate port 3 ports",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 8889, 9000},
				Scheduler:     "WRR",
				Session:       0,
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate port out of range",
			args: &CreateLayer4ListenerDetail{
				Protocol:      enumor.TcpProtocol,
				ListenerPorts: []int{8888, 65536},
				Scheduler:     "WRR",
				Session:       0,
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate port 0",
			args: &CreateLayer4ListenerDetail{
				Protocol:  enumor.TcpProtocol,
				Scheduler: "WRR",
				Session:   0,
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
