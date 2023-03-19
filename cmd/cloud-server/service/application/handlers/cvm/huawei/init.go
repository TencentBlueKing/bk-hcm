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

package huawei

import (
	"hcm/cmd/cloud-server/service/application/handlers"
	typecvm "hcm/pkg/adaptor/types/cvm"
	proto "hcm/pkg/api/cloud-server/application"
	hcproto "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
)

// ApplicationOfCreateHuaWeiCvm ...
type ApplicationOfCreateHuaWeiCvm struct {
	handlers.BaseApplicationHandler

	vendor           enumor.Vendor
	req              *proto.HuaWeiCvmCreateReq
	platformManagers []string
}

// NewApplicationOfCreateHuaWeiCvm ...
func NewApplicationOfCreateHuaWeiCvm(
	opt *handlers.HandlerOption, req *proto.HuaWeiCvmCreateReq, platformManagers []string,
) *ApplicationOfCreateHuaWeiCvm {
	return &ApplicationOfCreateHuaWeiCvm{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(opt, enumor.CreateCvm),
		vendor:                 enumor.HuaWei,
		req:                    req,
		platformManagers:       platformManagers,
	}
}

func (a *ApplicationOfCreateHuaWeiCvm) toHcProtoHuaWeiBatchCreateReq(dryRun bool) *hcproto.HuaWeiBatchCreateReq {
	req := a.req
	// 数据盘
	dataVolumes := make([]typecvm.HuaWeiVolume, 0)
	for _, d := range req.DataDisk {
		for i := int64(0); i < d.DiskCount; i++ {
			dataVolumes = append(dataVolumes, typecvm.HuaWeiVolume{
				VolumeType: d.DiskType,
				SizeGB:     int32(d.DiskSizeGB),
			})
		}
	}

	// 计费
	periodType := typecvm.Month
	periodNum := int32(req.InstanceChargePaidPeriod)
	if periodNum > 9 {
		periodType = typecvm.Year
		periodNum = int32(req.InstanceChargePaidPeriod / 12)
	}

	return &hcproto.HuaWeiBatchCreateReq{
		DryRun:                dryRun,
		AccountID:             req.AccountID,
		Region:                req.Region,
		Name:                  req.Name,
		Zone:                  req.Zone,
		InstanceType:          req.InstanceType,
		CloudImageID:          req.CloudImageID,
		Password:              req.Password,
		RequiredCount:         int32(req.RequiredCount),
		CloudSecurityGroupIDs: req.CloudSecurityGroupIDs,
		// TODO: 暂不支持
		// ClientToken: nil,
		CloudVpcID:    req.CloudVpcID,
		CloudSubnetID: req.CloudSubnetID,
		Description:   req.Memo,
		RootVolume: &typecvm.HuaWeiVolume{
			VolumeType: req.SystemDisk.DiskType,
			SizeGB:     int32(req.SystemDisk.DiskSizeGB),
		},
		DataVolume: dataVolumes,
		InstanceCharge: &typecvm.HuaWeiInstanceCharge{
			ChargingMode: req.InstanceChargeType,
			PeriodType:   &periodType,
			PeriodNum:    &periodNum,
			IsAutoRenew:  &req.AutoRenew,
		},
	}
}
