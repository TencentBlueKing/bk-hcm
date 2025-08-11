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
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
)

func TestLayer4ListenerBindRSExecutor_convertDataToPreview(t *testing.T) {
	type args struct {
		rawData [][]string
		headers []string
	}
	tests := []struct {
		name    string
		args    args
		want    Layer4ListenerBindRSDetail
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "CVM", "127.0.0.1", "8000", "50", "用户的备注"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "后端类型", "rs_ip",
					"rs_port", "权重(0-100)", "用户备注(可选)", "导出备注(可选)"},
			},
			want: Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					ClbVipDomain: "127.0.0.1",
					CloudClbID:   "lb-xxxxx1",
					Protocol:     enumor.TcpProtocol,
					InstType:     enumor.CvmInstType,
					RsIp:         "127.0.0.1",
					Weight:       cvt.ValToPtr(int64(50)),
					UserRemark:   "用户的备注",
				},
				ListenerPort:   []int{8888},
				RsPort:         []int{8000},
				Status:         "",
				ValidateResult: []string{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "end_port",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "[8888, 8889]", "CVM", "127.0.0.1", "[8888, 8889]", "50"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "后端类型", "rs_ip",
					"rs_port", "权重(0-100)", "用户备注(可选)", "导出备注(可选)"}},
			want: Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					ClbVipDomain: "127.0.0.1",
					CloudClbID:   "lb-xxxxx1",
					Protocol:     enumor.TcpProtocol,
					InstType:     enumor.CvmInstType,
					RsIp:         "127.0.0.1",
					Weight:       cvt.ValToPtr(int64(50)),
					UserRemark:   "",
				},
				ListenerPort:   []int{8888, 8889},
				RsPort:         []int{8888, 8889},
				Status:         "",
				ValidateResult: []string{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "表头缺失",
			args: args{
				rawData: [][]string{
					{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "CVM", "127.0.0.1", "8000", "50", "用户的备注"},
				},
				headers: []string{"负载均衡vip/域名", "负载均衡云ID", "监听器协议", "监听器端口", "后端类型", "rs_ip",
					"rs_port", "权重(0-100)", "导出备注(可选)"},
			},
			want: Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					ClbVipDomain: "127.0.0.1",
					CloudClbID:   "lb-xxxxx1",
					Protocol:     enumor.TcpProtocol,
					InstType:     enumor.CvmInstType,
					RsIp:         "127.0.0.1",
					Weight:       cvt.ValToPtr(int64(50)),
					UserRemark:   "用户的备注",
				},
				ListenerPort:   []int{8888},
				RsPort:         []int{8000},
				Status:         "",
				ValidateResult: []string{},
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Layer4ListenerBindRSPreviewExecutor{}
			err := executor.convertDataToPreview(tt.args.rawData, tt.args.headers)
			tt.wantErr(t, err)
			if len(executor.details) > 0 {
				assert.Equal(t, tt.want, *executor.details[0])
			}
		})
	}
}

func TestLayer4ListenerBindRSDetail_validate(t *testing.T) {
	tests := []struct {
		name       string
		args       *Layer4ListenerBindRSDetail
		wantStatus ImportStatus
	}{
		{
			name: "validate protocol executable",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: enumor.CvmInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(50)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888, 8889},
			},
			wantStatus: Executable,
		},
		{
			name: "validate protocol not executable",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.HttpsProtocol,
					InstType: enumor.CvmInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(50)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate port not executable",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: enumor.CvmInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(50)),
				},
				ListenerPort: []int{8888, 70000},
				RsPort:       []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate rs port not executable",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: enumor.CvmInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(50)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888, 70000},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate instType not executable",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: "213",
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(50)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate weight out of range 101",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: enumor.EniInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(101)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888, 8889},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "端口段设置错误",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: enumor.EniInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(100)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888},
			},
			wantStatus: NotExecutable,
		},
		{
			name: "端口段设置错误",
			args: &Layer4ListenerBindRSDetail{
				Layer4RsDetail: Layer4RsDetail{
					Protocol: enumor.TcpProtocol,
					InstType: enumor.EniInstType,
					RsIp:     "127.0.0.1",
					Weight:   cvt.ValToPtr(int64(100)),
				},
				ListenerPort: []int{8888, 8889},
				RsPort:       []int{8888, 9000},
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
