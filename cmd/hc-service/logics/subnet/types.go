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

// Package subnet defines subnet logics.
package subnet

import (
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types"
	hcservice "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/client"
	"hcm/pkg/criteria/validator"
)

// Subnet logics.
type Subnet struct {
	client  *client.ClientSet
	adaptor *cloudclient.CloudAdaptorClient
}

// NewSubnet new subnet logics.
func NewSubnet(client *client.ClientSet, adaptor *cloudclient.CloudAdaptorClient) *Subnet {
	return &Subnet{
		client:  client,
		adaptor: adaptor,
	}
}

// SubnetCreateOptions create subnet options.
type SubnetCreateOptions[T hcservice.SubnetCreateExt] struct {
	BkBizID    int64                          `validate:"required"`
	AccountID  string                         `validate:"required"`
	Region     string                         `validate:"required"`
	CloudVpcID string                         `validate:"required"`
	CreateReqs []hcservice.SubnetCreateReq[T] `validate:"min=1,max=100"`
}

// Validate SubnetCreateReq.
func (c SubnetCreateOptions[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureSubnetSyncOptions sync azure subnet options.
type AzureSubnetSyncOptions struct {
	BkBizID       int64               `validate:"required"`
	AccountID     string              `validate:"required"`
	CloudVpcID    string              `validate:"required"`
	ResourceGroup string              `validate:"required"`
	Subnets       []types.AzureSubnet `validate:"min=1,max=100"`
}

// Validate AzureSubnetSyncOptions.
func (c AzureSubnetSyncOptions) Validate() error {
	return validator.Validate.Struct(c)
}
