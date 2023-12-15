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
	"os"
	"testing"

	"hcm/pkg/adaptor/types"
	typeclb "hcm/pkg/adaptor/types/clb"
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func genClientSet() (clientSet *clientSet) {
	s := &types.BaseSecret{
		CloudSecretID:  os.Getenv("TCLOUD_AK"),
		CloudSecretKey: os.Getenv("TCLOUD_SK"),
		CloudAccountID: "",
	}
	return newClientSet(s, profile.NewClientProfile())
}

var clbId string

func TestTCloudImpl_ListLoadBalancer(t1 *testing.T) {
	cli := genClientSet()
	kt := core.NewBackendKit()
	type fields struct {
		clientSet *clientSet
	}
	type args struct {
		kt  *kit.Kit
		opt *typeclb.TCloudCLBListOpt
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantClbList []typeclb.TCloudCLB
		wantErr     bool
	}{
		{
			name: "list clb",
			fields: fields{
				clientSet: cli,
			},
			args: args{
				kt: kt,
				opt: &typeclb.TCloudCLBListOpt{
					Region: "ap-guangzhou",
				},
			},
			wantClbList: nil,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TCloudImpl{
				clientSet: tt.fields.clientSet,
			}
			gotClbList, err := t.ListLoadBalancer(tt.args.kt, tt.args.opt)
			for i, clb := range gotClbList {
				t1.Logf("%d: %+v, %+v, %+v,%+v", i,
					cvt.PtrToVal(clb.LoadBalancerName),
					cvt.PtrToVal(clb.LoadBalancerId),
					slice.Map(clb.LoadBalancerVips, cvt.PtrToVal[string]),
					cvt.PtrToVal(clb.Domain))
				clbId = cvt.PtrToVal(clb.LoadBalancerId)
			}
			if (err != nil) != tt.wantErr {
				t1.Errorf("ListLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestTCloudImpl_ListListeners(t1 *testing.T) {
	cli := genClientSet()
	kt := core.NewBackendKit()
	type fields struct {
		clientSet *clientSet
	}
	type args struct {
		kt  *kit.Kit
		opt *typeclb.TCloudListenerListOpt
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantListeners []typeclb.TCloudListener
		wantErr       bool
	}{
		{
			name: "test list",
			fields: fields{
				clientSet: cli,
			},
			args: args{
				kt: kt,
				opt: &typeclb.TCloudListenerListOpt{
					Region: "ap-guangzhou",
					ClbID:  clbId,
				},
			},
			wantListeners: nil,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TCloudImpl{
				clientSet: tt.fields.clientSet,
			}
			gotListeners, err := t.ListListeners(tt.args.kt, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t1.Errorf("ListListeners() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, listener := range gotListeners {
				t1.Logf("%d: %s, %s, %s, %d", i, cvt.PtrToVal(listener.ListenerId),
					cvt.PtrToVal(listener.ListenerName),
					cvt.PtrToVal(listener.Protocol), cvt.ValToPtr(listener.Port),
				)
			}

		})
	}
}

func TestTCloudImpl_ListTargets(t1 *testing.T) {
	cli := genClientSet()
	kt := core.NewBackendKit()
	type fields struct {
		clientSet *clientSet
	}
	type args struct {
		kt  *kit.Kit
		opt *typeclb.TCloudTargetListOpt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "list rs",
			fields: fields{
				clientSet: cli,
			},
			args: args{
				kt: kt,
				opt: &typeclb.TCloudTargetListOpt{
					Region: "ap-guangzhou",
					ClbID:  clbId,
				},
			},

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TCloudImpl{
				clientSet: tt.fields.clientSet,
			}
			gotTargets, err := t.ListTargets(tt.args.kt, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t1.Errorf("ListTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, target := range gotTargets {
				data, _ := json.Marshal(target)
				t1.Logf("%d, %s", i, string(data))
			}
		})
	}
}
