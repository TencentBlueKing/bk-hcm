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

package cvm

import (
	corecvm "hcm/pkg/api/core/cloud/cvm"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"google.golang.org/api/compute/v1"
)

// TCloudCvmSync ...
type TCloudCvmSync struct {
	IsUpdate bool
	Cvm      *cvm.Instance
}

// TCloudDSCvmSync ...
type TCloudDSCvmSync struct {
	Cvm corecvm.Cvm[corecvm.TCloudCvmExtension]
}

// AwsCvmSync ...
type AwsCvmSync struct {
	IsUpdate bool
	Cvm      *ec2.Instance
}

// AwsDSCvmSync ...
type AwsDSCvmSync struct {
	Cvm corecvm.Cvm[corecvm.AwsCvmExtension]
}

// HuaWeiCvmSync ...
type HuaWeiCvmSync struct {
	IsUpdate bool
	Cvm      model.ServerDetail
}

// HuaWeiCvmSync ...
type HuaWeiDSCvmSync struct {
	Cvm corecvm.Cvm[corecvm.HuaWeiCvmExtension]
}

// GcpCvmSync
type GcpCvmSync struct {
	IsUpdate bool
	Cvm      *compute.Instance
}

// GcpDSCvmSync ...
type GcpDSCvmSync struct {
	Cvm corecvm.Cvm[corecvm.GcpCvmExtension]
}

// AzureCvmSync ...
type AzureCvmSync struct {
	IsUpdate bool
	Cvm      *armcompute.VirtualMachine
}

// AzureDSCvmSync
type AzureDSCvmSync struct {
	Cvm corecvm.Cvm[corecvm.AzureCvmExtension]
}

// OperateSync
type CVMOperateSync struct {
	HCRelID      string
	RelID        string
	HCInstanceID string
	InstanceID   string
}
