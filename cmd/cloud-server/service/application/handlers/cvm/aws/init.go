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

package aws

import (
	"hcm/cmd/cloud-server/service/application/handlers"
	typecvm "hcm/pkg/adaptor/types/cvm"
	proto "hcm/pkg/api/cloud-server/application"
	hcproto "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
)

// ApplicationOfCreateAwsCvm ...
type ApplicationOfCreateAwsCvm struct {
	handlers.BaseApplicationHandler

	req *proto.AwsCvmCreateReq
}

// NewApplicationOfCreateAwsCvm ...
func NewApplicationOfCreateAwsCvm(
	opt *handlers.HandlerOption, req *proto.AwsCvmCreateReq,
) *ApplicationOfCreateAwsCvm {
	return &ApplicationOfCreateAwsCvm{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(opt, enumor.CreateCvm, enumor.Aws),
		req:                    req,
	}
}

func (a *ApplicationOfCreateAwsCvm) toHcProtoAwsBatchCreateReq(dryRun bool) *hcproto.AwsBatchCreateReq {
	req := a.req

	blockDeviceMapping := make([]typecvm.AwsBlockDeviceMapping, 0)
	// 系统盘
	blockDeviceMapping = append(blockDeviceMapping, typecvm.AwsBlockDeviceMapping{
		Ebs: &typecvm.AwsEbs{
			VolumeSizeGB: req.SystemDisk.DiskSizeGB,
			VolumeType:   req.SystemDisk.DiskType,
		},
	})
	// 数据盘
	for _, d := range req.DataDisk {
		for i := int64(0); i < d.DiskCount; i++ {
			blockDeviceMapping = append(blockDeviceMapping, typecvm.AwsBlockDeviceMapping{
				Ebs: &typecvm.AwsEbs{
					VolumeSizeGB: d.DiskSizeGB,
					VolumeType:   d.DiskType,
				},
			})
		}
	}

	return &hcproto.AwsBatchCreateReq{
		DryRun:                dryRun,
		AccountID:             req.AccountID,
		Region:                req.Region,
		Zone:                  req.Zone,
		Name:                  req.Name,
		InstanceType:          req.InstanceType,
		CloudImageID:          req.CloudImageID,
		CloudSubnetID:         req.CloudSubnetID,
		PublicIPAssigned:      req.PublicIPAssigned,
		CloudSecurityGroupIDs: req.CloudSecurityGroupIDs,
		BlockDeviceMapping:    blockDeviceMapping,
		Password:              req.Password,
		RequiredCount:         req.RequiredCount,
		// TODO: 暂不支持
		// ClientToken: nil,
	}
}
