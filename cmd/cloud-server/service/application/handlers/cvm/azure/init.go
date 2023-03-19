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

package azure

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	typecvm "hcm/pkg/adaptor/types/cvm"
	proto "hcm/pkg/api/cloud-server/application"
	hcproto "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
)

// ApplicationOfCreateAzureCvm ...
type ApplicationOfCreateAzureCvm struct {
	handlers.BaseApplicationHandler

	vendor           enumor.Vendor
	req              *proto.AzureCvmCreateReq
	platformManagers []string
}

// NewApplicationOfCreateAzureCvm ...
func NewApplicationOfCreateAzureCvm(
	opt *handlers.HandlerOption, req *proto.AzureCvmCreateReq, platformManagers []string,
) *ApplicationOfCreateAzureCvm {
	return &ApplicationOfCreateAzureCvm{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(opt, enumor.CreateCvm),
		vendor:                 enumor.Azure,
		req:                    req,
		platformManagers:       platformManagers,
	}
}

func (a *ApplicationOfCreateAzureCvm) toHcProtoAzureBatchCreateReq() *hcproto.AzureBatchCreateReq {
	req := a.req

	dataDisk := make([]typecvm.AzureDataDisk, 0)
	index := 1
	for _, d := range req.DataDisk {
		for i := int64(0); i < d.DiskCount; i++ {
			dataDisk = append(dataDisk, typecvm.AzureDataDisk{
				Name:   fmt.Sprintf("data%d", index),
				SizeGB: int32(d.DiskSizeGB),
			})
			index += 1
		}
	}

	return &hcproto.AzureBatchCreateReq{
		AccountID:            req.AccountID,
		ResourceGroupName:    req.ResourceGroupName,
		Region:               req.Region,
		Name:                 req.Name,
		Zones:                []string{req.Zone},
		InstanceType:         req.InstanceType,
		CloudImageID:         req.CloudImageID,
		Username:             req.Username,
		Password:             req.Password,
		CloudSubnetID:        req.CloudSubnetID,
		CloudSecurityGroupID: req.CloudSecurityGroupIDs[0],
		OSDisk: &typecvm.AzureOSDisk{
			Name:   "os1",
			SizeGB: int32(req.SystemDisk.DiskSizeGB),
		},
		DataDisk:      dataDisk,
		RequiredCount: req.RequiredCount,
	}
}
