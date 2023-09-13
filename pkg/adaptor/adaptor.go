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

// Package adaptor ...
package adaptor

import (
	"errors"

	"hcm/pkg/adaptor/aws"
	"hcm/pkg/adaptor/azure"
	"hcm/pkg/adaptor/gcp"
	"hcm/pkg/adaptor/huawei"
	mocktcloud "hcm/pkg/adaptor/mock/tcloud"
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/logs"
)

// Adaptor holds all the supported operations by the adaptor.
type Adaptor struct {
	EnableCloudMock bool
}

// Option Adaptor options
type Option struct {
	EnableCloudMock bool
}

// New an Adaptor pointer
func New(opt Option) *Adaptor {
	if opt.EnableCloudMock {
		logs.Infof("Using mock server")
	}
	return &Adaptor{EnableCloudMock: opt.EnableCloudMock}
}

// TCloud returns tencent cloud operations.
func (a *Adaptor) TCloud(s *types.BaseSecret) (tcloud.TCloud, error) {
	if a.EnableCloudMock {
		mockTcloud := mocktcloud.GetMockCloud()
		return mockTcloud, nil
	}
	return tcloud.NewTCloud(s)
}

// Aws returns Aws operations.
func (a *Adaptor) Aws(s *types.BaseSecret, cloudAccountID string) (*aws.Aws, error) {
	if a.EnableCloudMock {
		return nil, errors.New("mock of aws not implemented")
	}
	return aws.NewAws(s, cloudAccountID)
}

// Gcp returns Gcp operations.
func (a *Adaptor) Gcp(credential *types.GcpCredential) (*gcp.Gcp, error) {
	if a.EnableCloudMock {
		return nil, errors.New("mock of gcp not implemented")
	}
	return gcp.NewGcp(credential)
}

// Azure returns Azure operations.
func (a *Adaptor) Azure(credential *types.AzureCredential) (*azure.Azure, error) {
	if a.EnableCloudMock {
		return nil, errors.New("mock of azure not implemented")
	}
	return azure.NewAzure(credential)
}

// HuaWei returns HuaWei operations.
func (a *Adaptor) HuaWei(s *types.BaseSecret) (*huawei.HuaWei, error) {
	if a.EnableCloudMock {
		return nil, errors.New("mock of huawei not implemented")
	}
	return huawei.NewHuaWei(s)
}
