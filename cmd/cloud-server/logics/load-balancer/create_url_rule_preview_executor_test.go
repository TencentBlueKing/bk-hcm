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
	"fmt"
	"testing"

	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/assert"
)

func TestCreateURLRuleExecutor_convertDataToPreview_validateFailed(t *testing.T) {
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
				{"127.0.0.1", "lb-xxxxx1", "http", "8888",
					"www.tencent.com", "是", "/", "WRR", "0", "enable", "用户的备注"},
			}},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateUrlRulePreviewExecutor{}
			err := executor.convertDataToPreview(tt.args.i)
			tt.wantErr(t, err)
		})
	}
}

func TestCreateUrlRuleExecutor_convertDataToPreview(t *testing.T) {
	type args struct {
		i [][]string
	}
	tests := []struct {
		name string
		args args
		want CreateUrlRuleDetail
	}{
		{
			name: "test",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "http", "8888",
					"www.tencent.com", "TRUE", "/", "WRR", "0", "enable", "用户的备注"},
			}},
			want: CreateUrlRuleDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.HttpProtocol,
				ListenerPort:   []int{8888},
				Domain:         "www.tencent.com",
				DefaultDomain:  true,
				UrlPath:        "/",
				Scheduler:      "WRR",
				Session:        0,
				HealthCheck:    true,
				UserRemark:     "用户的备注",
				Status:         "",
				ValidateResult: []string{},
			},
		},
		{
			name: "end_port",
			args: args{i: [][]string{
				{"127.0.0.1", "lb-xxxxx1", "tcp", "[8888, 8889]", "www.tencent.com", "TRUE", "/", "WRR", "0", "disable"},
			}},
			want: CreateUrlRuleDetail{
				ClbVipDomain:   "127.0.0.1",
				CloudClbID:     "lb-xxxxx1",
				Protocol:       enumor.TcpProtocol,
				ListenerPort:   []int{8888, 8889},
				Domain:         "www.tencent.com",
				DefaultDomain:  true,
				UrlPath:        "/",
				Scheduler:      "WRR",
				Session:        0,
				HealthCheck:    false,
				UserRemark:     "",
				Status:         "",
				ValidateResult: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &CreateUrlRulePreviewExecutor{}
			_ = executor.convertDataToPreview(tt.args.i)
			assert.Equal(t, tt.want, *executor.details[0])
		})
	}
}

func TestCreateUrlRuleDetail_validate(t *testing.T) {
	tests := []struct {
		name       string
		args       *CreateUrlRuleDetail
		wantStatus ImportStatus
	}{
		{
			name: "validate protocol executable",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888, 8889},
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
				Scheduler:     "WRR",
				Session:       30,
			},
			wantStatus: Executable,
		},
		{
			name: "validate protocol not executable",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.QuicProtocol,
				ListenerPort:  []int{8888, 8889},
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
				Scheduler:     "WRR",
				Session:       30,
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate Scheduler normal",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpsProtocol,
				ListenerPort:  []int{8888, 8889},
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
				Scheduler:     enumor.WRR,
				Session:       30,
			},
			wantStatus: Executable,
		},
		{
			name: "validate Scheduler wrong scheduler ",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888, 8889},
				Scheduler:     "WRR2",
				Session:       30,
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate session less than 30",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888, 8889},
				Scheduler:     "WRR",
				Session:       10,
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate session normal",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888, 8889},
				Scheduler:     "WRR",
				Session:       0,
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
			},
			wantStatus: Executable,
		},
		{
			name: "validate port 3 ports",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888, 8889, 9000},
				Scheduler:     "WRR",
				Session:       0,
				Domain:        "www.tencent.com",
				DefaultDomain: true,
				UrlPath:       "/",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate domain is empty",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888},
				Scheduler:     "WRR",
				Session:       0,
				Domain:        "",
				DefaultDomain: true,
				UrlPath:       "/",
			},
			wantStatus: NotExecutable,
		},
		{
			name: "validate url path is empty",
			args: &CreateUrlRuleDetail{
				Protocol:      enumor.HttpProtocol,
				ListenerPort:  []int{8888},
				Scheduler:     "WRR",
				Session:       0,
				Domain:        "",
				DefaultDomain: true,
				UrlPath:       "/",
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

func Test_decodeClassifyKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   enumor.ProtocolType
		want2   int
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{
			name:    "normal case",
			args:    args{key: "clb-xxxxx1/http/8888"},
			want:    "clb-xxxxx1",
			want1:   "http",
			want2:   8888,
			wantErr: assert.NoError,
		},
		{
			name:    "bad case",
			args:    args{key: "clb-xxxxx1/http/"},
			want:    "",
			want1:   "",
			want2:   0,
			wantErr: assert.Error,
		},
		{
			name:    "bad case",
			args:    args{key: "clb-xxxxx1/http"},
			want:    "",
			want1:   "",
			want2:   0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := decodeClassifyKey(tt.args.key)
			if !tt.wantErr(t, err, fmt.Sprintf("decodeClassifyKey(%v)", tt.args.key)) {
				return
			}
			assert.Equalf(t, tt.want, got, "decodeClassifyKey(%v)", tt.args.key)
			assert.Equalf(t, tt.want1, got1, "decodeClassifyKey(%v)", tt.args.key)
			assert.Equalf(t, tt.want2, got2, "decodeClassifyKey(%v)", tt.args.key)
		})
	}
}
