//go:build mock

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

// Package adaptor 对云API的封装。该文件用于集成测试使用，通过go build -tags mock 编译
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
}

// New an Adaptor pointer
func New() *Adaptor {
	logs.Infof("Using mock server")

	return &Adaptor{}
}

// TCloud returns tencent cloud operations.
func (a *Adaptor) TCloud(s *types.BaseSecret) (tcloud.TCloud, error) {
	mockTcloud := mocktcloud.GetMockCloud()
	return mockTcloud, nil
}

// Aws returns Aws operations.
func (a *Adaptor) Aws(s *types.BaseSecret, cloudAccountID string) (*aws.Aws, error) {
	return nil, errors.New("mock of aws not implemented")
}

// Gcp returns Gcp operations.
func (a *Adaptor) Gcp(credential *types.GcpCredential) (*gcp.Gcp, error) {
	return nil, errors.New("mock of gcp not implemented")
}

// Azure returns Azure operations.
func (a *Adaptor) Azure(credential *types.AzureCredential) (*azure.Azure, error) {
	return nil, errors.New("mock of azure not implemented")
}

// HuaWei returns HuaWei operations.
func (a *Adaptor) HuaWei(s *types.BaseSecret) (*huawei.HuaWei, error) {
	return nil, errors.New("mock of huawei not implemented")
}
