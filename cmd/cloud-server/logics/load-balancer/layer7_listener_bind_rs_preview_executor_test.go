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

func TestLayer7ListenerBindRSExecutor_convertDataToPreview(t *testing.T) {
	type args struct {
		i [][]string
	}
	tests := []struct {
		name string
		args args
		want Layer7ListenerBindRSDetail
	}{
		{
			name: "test",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "tcp", "8888", "tencent.com", "/",
					"CVM", "127.0.0.1", "8000", "50", "用户的备注"},
			}},
			want: Layer7ListenerBindRSDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.TcpProtocol,
				ListenerPort:   []int{8888},
				Domain:         "tencent.com",
				URLPath:        "/",
				InstType:       enumor.CvmInstType,
				RsIp:           "127.0.0.1",
				RsPort:         []int{8000},
				Weight:         cvt.ValToPtr(50),
				UserRemark:     "用户的备注",
				Status:         "",
				ValidateResult: []string{},
			},
		},
		{
			name: "end_port",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "tcp", "[8888, 8889]", "tencent.com", "/",
					"CVM", "127.0.0.1   ", "[8888, 8889]", "50"},
			}},
			want: Layer7ListenerBindRSDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.TcpProtocol,
				ListenerPort:   []int{8888, 8889},
				Domain:         "tencent.com",
				URLPath:        "/",
				InstType:       enumor.CvmInstType,
				RsIp:           "127.0.0.1",
				RsPort:         []int{8888, 8889},
				Weight:         cvt.ValToPtr(50),
				UserRemark:     "",
				Status:         "",
				ValidateResult: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Layer7ListenerBindRSPreviewExecutor{}
			_ = executor.convertDataToPreview(tt.args.i)
			assert.Equal(t, tt.want, *executor.details[0])
		})
	}
}

func TestLayer7ListenerBindRSDetail_validate(t *testing.T) {
	tests := []struct {
		name       string
		args       *Layer7ListenerBindRSDetail
		wantStatus ImportStatus
	}{
		{
			name: "validate protocol executable",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     enumor.CvmInstType,
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(50),
			},
			wantStatus: Executable,
		},
		{
			name: "validate protocol not executable",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.TcpProtocol,
				ListenerPort: []int{8888, 8889},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     enumor.CvmInstType,
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(50),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate port not executable",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpsProtocol,
				ListenerPort: []int{8888, 70000},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     enumor.CvmInstType,
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(50),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate rs port not executable",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     enumor.CvmInstType,
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 70000},
				Weight:       cvt.ValToPtr(50),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate instType not executable",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     "213",
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(50),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate weight out of range 101",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     enumor.EniInstType,
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(101),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "端口段设置错误",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				Domain:       "tencent.com",
				URLPath:      "/",
				InstType:     enumor.EniInstType,
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888},
				Weight:       cvt.ValToPtr(100),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "端口段设置错误,长度不一致",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				InstType:     enumor.EniInstType,
				Domain:       "tencent.com",
				URLPath:      "/",
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 9000},
				Weight:       cvt.ValToPtr(100),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "设置端口段, 权重为0",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				InstType:     enumor.EniInstType,
				Domain:       "tencent.com",
				URLPath:      "/",
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(0),
			},
			wantStatus: NotExecutable,
		},
		{
			name: "domain为空",
			args: &Layer7ListenerBindRSDetail{
				Protocol:     enumor.HttpProtocol,
				ListenerPort: []int{8888, 8889},
				InstType:     enumor.EniInstType,
				Domain:       "",
				URLPath:      "/",
				RsIp:         "127.0.0.1",
				RsPort:       []int{8888, 8889},
				Weight:       cvt.ValToPtr(0),
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
