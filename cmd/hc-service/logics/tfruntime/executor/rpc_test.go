/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package executor

import (
	"context"
	"sync"

	"hcm/pkg/plugin/tfplugin5"

	"google.golang.org/grpc"
)

var _ tfplugin5.ProviderClient = (*fakePlugin)(nil)

type fakePlugin struct {
	mu    sync.Mutex
	calls map[string]int

	getSchema        func(*tfplugin5.GetProviderSchema_Request) (*tfplugin5.GetProviderSchema_Response, error)
	prepare          func(*tfplugin5.PrepareProviderConfig_Request) (*tfplugin5.PrepareProviderConfig_Response, error)
	configure        func(*tfplugin5.Configure_Request) (*tfplugin5.Configure_Response, error)
	validateResource func(*tfplugin5.ValidateResourceTypeConfig_Request) (*tfplugin5.ValidateResourceTypeConfig_Response, error)
	validateData     func(*tfplugin5.ValidateDataSourceConfig_Request) (*tfplugin5.ValidateDataSourceConfig_Response, error)
	upgrade          func(*tfplugin5.UpgradeResourceState_Request) (*tfplugin5.UpgradeResourceState_Response, error)
	read             func(*tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error)
	plan             func(*tfplugin5.PlanResourceChange_Request) (*tfplugin5.PlanResourceChange_Response, error)
	apply            func(*tfplugin5.ApplyResourceChange_Request) (*tfplugin5.ApplyResourceChange_Response, error)
	importState      func(*tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error)
	readData         func(*tfplugin5.ReadDataSource_Request) (*tfplugin5.ReadDataSource_Response, error)
	stop             func(*tfplugin5.Stop_Request) (*tfplugin5.Stop_Response, error)
}

func (f *fakePlugin) hit(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[name]++
}

func (f *fakePlugin) count(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[name]
}

func (f *fakePlugin) GetSchema(_ context.Context, in *tfplugin5.GetProviderSchema_Request, _ ...grpc.CallOption) (*tfplugin5.GetProviderSchema_Response, error) {
	f.hit("GetSchema")
	if f.getSchema != nil {
		return f.getSchema(in)
	}
	return &tfplugin5.GetProviderSchema_Response{}, nil
}

func (f *fakePlugin) PrepareProviderConfig(_ context.Context, in *tfplugin5.PrepareProviderConfig_Request, _ ...grpc.CallOption) (*tfplugin5.PrepareProviderConfig_Response, error) {
	f.hit("PrepareProviderConfig")
	if f.prepare != nil {
		return f.prepare(in)
	}
	return &tfplugin5.PrepareProviderConfig_Response{}, nil
}

func (f *fakePlugin) Configure(_ context.Context, in *tfplugin5.Configure_Request, _ ...grpc.CallOption) (*tfplugin5.Configure_Response, error) {
	f.hit("Configure")
	if f.configure != nil {
		return f.configure(in)
	}
	return &tfplugin5.Configure_Response{}, nil
}

func (f *fakePlugin) ValidateResourceTypeConfig(_ context.Context, in *tfplugin5.ValidateResourceTypeConfig_Request, _ ...grpc.CallOption) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
	f.hit("ValidateResourceTypeConfig")
	if f.validateResource != nil {
		return f.validateResource(in)
	}
	return &tfplugin5.ValidateResourceTypeConfig_Response{}, nil
}

func (f *fakePlugin) ValidateDataSourceConfig(_ context.Context, in *tfplugin5.ValidateDataSourceConfig_Request, _ ...grpc.CallOption) (*tfplugin5.ValidateDataSourceConfig_Response, error) {
	f.hit("ValidateDataSourceConfig")
	if f.validateData != nil {
		return f.validateData(in)
	}
	return &tfplugin5.ValidateDataSourceConfig_Response{}, nil
}

func (f *fakePlugin) UpgradeResourceState(_ context.Context, in *tfplugin5.UpgradeResourceState_Request, _ ...grpc.CallOption) (*tfplugin5.UpgradeResourceState_Response, error) {
	f.hit("UpgradeResourceState")
	if f.upgrade != nil {
		return f.upgrade(in)
	}
	return &tfplugin5.UpgradeResourceState_Response{}, nil
}

func (f *fakePlugin) ReadResource(_ context.Context, in *tfplugin5.ReadResource_Request, _ ...grpc.CallOption) (*tfplugin5.ReadResource_Response, error) {
	f.hit("ReadResource")
	if f.read != nil {
		return f.read(in)
	}
	return &tfplugin5.ReadResource_Response{}, nil
}

func (f *fakePlugin) PlanResourceChange(_ context.Context, in *tfplugin5.PlanResourceChange_Request, _ ...grpc.CallOption) (*tfplugin5.PlanResourceChange_Response, error) {
	f.hit("PlanResourceChange")
	if f.plan != nil {
		return f.plan(in)
	}
	return &tfplugin5.PlanResourceChange_Response{}, nil
}

func (f *fakePlugin) ApplyResourceChange(_ context.Context, in *tfplugin5.ApplyResourceChange_Request, _ ...grpc.CallOption) (*tfplugin5.ApplyResourceChange_Response, error) {
	f.hit("ApplyResourceChange")
	if f.apply != nil {
		return f.apply(in)
	}
	return &tfplugin5.ApplyResourceChange_Response{}, nil
}

func (f *fakePlugin) ImportResourceState(_ context.Context, in *tfplugin5.ImportResourceState_Request, _ ...grpc.CallOption) (*tfplugin5.ImportResourceState_Response, error) {
	f.hit("ImportResourceState")
	if f.importState != nil {
		return f.importState(in)
	}
	return &tfplugin5.ImportResourceState_Response{}, nil
}

func (f *fakePlugin) ReadDataSource(_ context.Context, in *tfplugin5.ReadDataSource_Request, _ ...grpc.CallOption) (*tfplugin5.ReadDataSource_Response, error) {
	f.hit("ReadDataSource")
	if f.readData != nil {
		return f.readData(in)
	}
	return &tfplugin5.ReadDataSource_Response{}, nil
}

func (f *fakePlugin) Stop(_ context.Context, in *tfplugin5.Stop_Request, _ ...grpc.CallOption) (*tfplugin5.Stop_Response, error) {
	f.hit("Stop")
	if f.stop != nil {
		return f.stop(in)
	}
	return &tfplugin5.Stop_Response{}, nil
}

func (f *fakePlugin) GetMetadata(context.Context, *tfplugin5.GetMetadata_Request, ...grpc.CallOption) (*tfplugin5.GetMetadata_Response, error) {
	return &tfplugin5.GetMetadata_Response{}, nil
}

func (f *fakePlugin) GetResourceIdentitySchemas(context.Context, *tfplugin5.GetResourceIdentitySchemas_Request, ...grpc.CallOption) (*tfplugin5.GetResourceIdentitySchemas_Response, error) {
	return &tfplugin5.GetResourceIdentitySchemas_Response{}, nil
}

func (f *fakePlugin) UpgradeResourceIdentity(context.Context, *tfplugin5.UpgradeResourceIdentity_Request, ...grpc.CallOption) (*tfplugin5.UpgradeResourceIdentity_Response, error) {
	return &tfplugin5.UpgradeResourceIdentity_Response{}, nil
}

func (f *fakePlugin) MoveResourceState(context.Context, *tfplugin5.MoveResourceState_Request, ...grpc.CallOption) (*tfplugin5.MoveResourceState_Response, error) {
	return &tfplugin5.MoveResourceState_Response{}, nil
}

func (f *fakePlugin) ValidateEphemeralResourceConfig(context.Context, *tfplugin5.ValidateEphemeralResourceConfig_Request, ...grpc.CallOption) (*tfplugin5.ValidateEphemeralResourceConfig_Response, error) {
	return &tfplugin5.ValidateEphemeralResourceConfig_Response{}, nil
}

func (f *fakePlugin) OpenEphemeralResource(context.Context, *tfplugin5.OpenEphemeralResource_Request, ...grpc.CallOption) (*tfplugin5.OpenEphemeralResource_Response, error) {
	return &tfplugin5.OpenEphemeralResource_Response{}, nil
}

func (f *fakePlugin) RenewEphemeralResource(context.Context, *tfplugin5.RenewEphemeralResource_Request, ...grpc.CallOption) (*tfplugin5.RenewEphemeralResource_Response, error) {
	return &tfplugin5.RenewEphemeralResource_Response{}, nil
}

func (f *fakePlugin) CloseEphemeralResource(context.Context, *tfplugin5.CloseEphemeralResource_Request, ...grpc.CallOption) (*tfplugin5.CloseEphemeralResource_Response, error) {
	return &tfplugin5.CloseEphemeralResource_Response{}, nil
}

func (f *fakePlugin) GetFunctions(context.Context, *tfplugin5.GetFunctions_Request, ...grpc.CallOption) (*tfplugin5.GetFunctions_Response, error) {
	return &tfplugin5.GetFunctions_Response{}, nil
}

func (f *fakePlugin) CallFunction(context.Context, *tfplugin5.CallFunction_Request, ...grpc.CallOption) (*tfplugin5.CallFunction_Response, error) {
	return &tfplugin5.CallFunction_Response{}, nil
}
