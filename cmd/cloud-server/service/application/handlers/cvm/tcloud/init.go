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
	"hcm/cmd/cloud-server/service/application/handlers"
	typecvm "hcm/pkg/adaptor/types/cvm"
	proto "hcm/pkg/api/cloud-server/application"
	hcproto "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
)

// ApplicationOfCreateTCloudCvm ...
type ApplicationOfCreateTCloudCvm struct {
	handlers.BaseApplicationHandler

	req *proto.TCloudCvmCreateReq
}

// NewApplicationOfCreateTCloudCvm ...
func NewApplicationOfCreateTCloudCvm(
	opt *handlers.HandlerOption, req *proto.TCloudCvmCreateReq,
) *ApplicationOfCreateTCloudCvm {
	return &ApplicationOfCreateTCloudCvm{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(opt, enumor.CreateCvm, enumor.TCloud),
		req:                    req,
	}
}

func (a *ApplicationOfCreateTCloudCvm) toHcProtoTCloudBatchCreateReq(dryRun bool) *hcproto.TCloudBatchCreateReq {
	req := a.req

	// 自动续费&续费周期
	instanceChargePrepaid := &typecvm.TCloudInstanceChargePrepaid{
		Period: &req.InstanceChargePaidPeriod,
		// 默认通知但不自动续费
		RenewFlag: typecvm.NotifyAndManualRenew,
	}
	if req.AutoRenew {
		instanceChargePrepaid.RenewFlag = typecvm.NotifyAndAutoRenew
	}

	// 数据盘
	dataDisk := make([]typecvm.TCloudDataDisk, 0)
	for _, d := range req.DataDisk {
		for i := int64(0); i < d.DiskCount; i++ {
			dataDisk = append(dataDisk, typecvm.TCloudDataDisk{
				DiskSizeGB: &d.DiskSizeGB,
				DiskType:   d.DiskType,
			})
		}
	}

	return &hcproto.TCloudBatchCreateReq{
		DryRun:                dryRun,
		AccountID:             req.AccountID,
		Region:                req.Region,
		Name:                  req.Name,
		Zone:                  req.Zone,
		InstanceType:          req.InstanceType,
		CloudImageID:          req.CloudImageID,
		Password:              req.Password,
		RequiredCount:         req.RequiredCount,
		CloudSecurityGroupIDs: req.CloudSecurityGroupIDs,
		// TODO: 暂不支持
		// ClientToken:           ,
		CloudVpcID:            req.CloudVpcID,
		CloudSubnetID:         req.CloudSubnetID,
		InstanceChargeType:    req.InstanceChargeType,
		InstanceChargePrepaid: instanceChargePrepaid,
		SystemDisk: &typecvm.TCloudSystemDisk{
			DiskType:   req.SystemDisk.DiskType,
			DiskSizeGB: &req.SystemDisk.DiskSizeGB,
		},
		DataDisk:         dataDisk,
		PublicIPAssigned: req.PublicIPAssigned,
	}
}
