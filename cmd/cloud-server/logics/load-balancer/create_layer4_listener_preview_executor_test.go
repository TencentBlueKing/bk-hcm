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
		rawData [][]string
		headers []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "invalid health check",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "WRR", "0", "disable（）", "用户的备注"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "均衡方式",
					"会话保持(0为不开启)", "健康检查", "监听器名称(可选)", "用户备注(可选)", "导出备注(可选)"},
			},
			wantErr: assert.Error,
		},
		{
			name: "headers不足,被删除了监听器名称",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "WRR", "0", "disable", "用户的备注"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "均衡方式",
					"会话保持(0为不开启)", "健康检查", "用户备注(可选)", "导出备注(可选)"},
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateLayer4ListenerPreviewExecutor{}
			err := executor.convertDataToPreview(tt.args.rawData, tt.args.headers)
			tt.wantErr(t, err)
		})
	}
}

func TestCreateLayer4ListenerExecutor_convertDataToPreview(t *testing.T) {
	type args struct {
		rawData [][]string
		headers []string
	}
	tests := []struct {
		name string
		args args
		want CreateLayer4ListenerDetail
	}{
		{
			name: "test",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "WRR", "0", "disable", "自定义监听器名称", "用户的备注"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "均衡方式",
					"会话保持(0为不开启)", "健康检查", "监听器名称(可选)", "用户备注(可选)", "导出备注(可选)",
				},
			},
			want: CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					ClbVipDomain: "127.0.0.1",
					CloudClbID:   "lb-xxxxx1",
					Protocol:     enumor.TcpProtocol,
					Scheduler:    "WRR",
					Session:      0,
					Name:         "自定义监听器名称",
					UserRemark:   "用户的备注",
				},
				ListenerPorts:  []int{8888},
				HealthCheck:    false,
				Status:         "",
				ValidateResult: []string{},
			},
		},
		{
			name: "end_port",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "[8888, 8889]", "WRR", "10", "enable"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "均衡方式",
					"会话保持(0为不开启)", "健康检查", "监听器名称(可选)", "用户备注(可选)", "导出备注(可选)",
				}},
			want: CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					ClbVipDomain: "127.0.0.1",
					CloudClbID:   "lb-xxxxx1",
					Protocol:     enumor.TcpProtocol,
					Scheduler:    "WRR",
					Session:      10,
					UserRemark:   "",
				},
				ListenerPorts:  []int{8888, 8889},
				HealthCheck:    true,
				Status:         "",
				ValidateResult: []string{},
			},
		},
		{
			name: "填写了监听器名称,没有填写用户备注",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "[8888, 8889]", "WRR", "10", "enable", "自定义监听器名称"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "均衡方式",
					"会话保持(0为不开启)", "健康检查", "监听器名称(可选)", "用户备注(可选)", "导出备注(可选)",
				}},
			want: CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					ClbVipDomain: "127.0.0.1",
					CloudClbID:   "lb-xxxxx1",
					Protocol:     enumor.TcpProtocol,
					Scheduler:    "WRR",
					Session:      10,
					UserRemark:   "",
					Name:         "自定义监听器名称",
				},
				ListenerPorts:  []int{8888, 8889},
				HealthCheck:    true,
				Status:         "",
				ValidateResult: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateLayer4ListenerPreviewExecutor{}
			_ = executor.convertDataToPreview(tt.args.rawData, tt.args.headers)
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
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR",
					Session:   30,
				},
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: Executable,
		},
		{
			name: "validate protocol",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.HttpsProtocol,
					Scheduler: "WRR",
					Session:   30,
				},
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate Scheduler normal",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: enumor.WRR,
					Session:   30,
				},
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: Executable,
		},
		{
			name: "validate Scheduler wrong scheduler ",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR2",
					Session:   30,
				},
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate session less than 30",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR",
					Session:   10,
				},
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate session normal",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR",
					Session:   0,
				},
				ListenerPorts: []int{8888, 8889},
			},
			wantStatus: Executable,
		},
		{
			name: "validate port 3 ports",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR",
					Session:   0,
				},
				ListenerPorts: []int{8888, 8889, 9000},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate port out of range",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR",
					Session:   0,
				},
				ListenerPorts: []int{8888, 65536},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate port 0",
			args: &CreateLayer4ListenerDetail{
				Layer4ListenerDetail: Layer4ListenerDetail{
					Protocol:  enumor.TcpProtocol,
					Scheduler: "WRR",
					Session:   0,
				},
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
