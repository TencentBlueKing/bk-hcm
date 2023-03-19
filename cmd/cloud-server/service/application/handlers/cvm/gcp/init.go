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

package gcp

import (
	"hcm/cmd/cloud-server/service/application/handlers"
	typecvm "hcm/pkg/adaptor/types/cvm"
	proto "hcm/pkg/api/cloud-server/application"
	hcproto "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
)

// ApplicationOfCreateGcpCvm ...
type ApplicationOfCreateGcpCvm struct {
	handlers.BaseApplicationHandler

	vendor           enumor.Vendor
	req              *proto.GcpCvmCreateReq
	platformManagers []string
}

// NewApplicationOfCreateGcpCvm ...
func NewApplicationOfCreateGcpCvm(
	opt *handlers.HandlerOption, req *proto.GcpCvmCreateReq, platformManagers []string,
) *ApplicationOfCreateGcpCvm {
	return &ApplicationOfCreateGcpCvm{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(opt, enumor.CreateCvm),
		vendor:                 enumor.Gcp,
		req:                    req,
		platformManagers:       platformManagers,
	}
}

func (a *ApplicationOfCreateGcpCvm) toHcProtoGcpBatchCreateReq() *hcproto.GcpBatchCreateReq {
	req := a.req

	dataDisk := make([]typecvm.GcpDataDisk, 0)
	// 数据盘
	for _, d := range req.DataDisk {
		for i := int64(0); i < d.DiskCount; i++ {
			dataDisk = append(dataDisk, typecvm.GcpDataDisk{
				DiskName:   d.DiskName,
				DiskType:   d.DiskType,
				SizeGb:     d.DiskSizeGB,
				Mode:       d.Mode,
				AutoDelete: d.AutoDelete,
			})
		}
	}
	description := ""
	if req.Memo != nil {
		description = *req.Memo
	}

	return &hcproto.GcpBatchCreateReq{
		AccountID:     req.AccountID,
		NamePrefix:    req.Name,
		Region:        req.Region,
		Zone:          req.Zone,
		InstanceType:  req.InstanceType,
		CloudImageID:  req.CloudImageID,
		Password:      req.Password,
		RequiredCount: req.RequiredCount,
		RequestID:     a.Cts.Kit.Rid,
		CloudVpcID:    req.CloudVpcID,
		CloudSubnetID: req.CloudSubnetID,
		Description:   description,
		SystemDisk: &typecvm.GcpOsDisk{
			DiskType: req.SystemDisk.DiskType,
			SizeGb:   req.SystemDisk.DiskSizeGB,
		},
		DataDisk: dataDisk,
	}
}
