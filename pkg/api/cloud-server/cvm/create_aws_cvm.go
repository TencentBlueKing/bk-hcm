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
 * specific language governing permissions and limitations under the License.w
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package cscvm

import (
	"errors"
	"fmt"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// AwsCvmCreateReq ...
type AwsCvmCreateReq struct {
	BkBizID               int64    `json:"bk_biz_id" validate:"omitempty"`
	AccountID             string   `json:"account_id" validate:"required"`
	BkCloudID             *int64   `json:"bk_cloud_id" validate:"required"`
	Region                string   `json:"region" validate:"required"`
	Zone                  string   `json:"zone" validate:"required"`
	Name                  string   `json:"name" validate:"required,min=1,max=60"`
	InstanceType          string   `json:"instance_type" validate:"required"`
	CloudImageID          string   `json:"cloud_image_id" validate:"required"`
	CloudVpcID            string   `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID         string   `json:"cloud_subnet_id" validate:"required"`
	PublicIPAssigned      bool     `json:"public_ip_assigned" validate:"omitempty"`
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids" validate:"required,min=1"`

	SystemDisk struct {
		DiskType   typecvm.AwsVolumeType `json:"disk_type" validate:"required"`
		DiskSizeGB int64                 `json:"disk_size_gb" validate:"required,min=1,max=16384"`
	} `json:"system_disk" validate:"required"`

	DataDisk []struct {
		DiskType   typecvm.AwsVolumeType `json:"disk_type" validate:"required"`
		DiskSizeGB int64                 `json:"disk_size_gb" validate:"required,min=1,max=16384"`
		DiskCount  int64                 `json:"disk_count" validate:"required,min=1"`
	} `json:"data_disk" validate:"omitempty"`

	// Note: aws是通过执行用户脚本添加密码的，可能有特殊字符会导致脚本执行失败，而且这是无法通过DryRun测试出来的
	Password          string `json:"password" validate:"required"`
	ConfirmedPassword string `json:"confirmed_password" validate:"eqfield=Password"`

	RequiredCount int64 `json:"required_count" validate:"required,min=1,max=500"`

	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *AwsCvmCreateReq) Validate(isFromBiz bool) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if isFromBiz && req.BkBizID == 0 {
		return errors.New("biz is required")
	}

	if isFromBiz && req.BkCloudID == nil {
		return errors.New("bk_cloud_id is required")
	}

	if err := validator.ValidateCvmName(enumor.Aws, req.Name); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.RequiredCount > constant.BatchOperationMaxLimit {
		return fmt.Errorf("required count should <= %d", constant.BatchOperationMaxLimit)
	}

	// 校验系统盘
	if err := req.validateDiskSize(req.SystemDisk.DiskType, req.SystemDisk.DiskSizeGB); err != nil {
		return err
	}

	dataDiskTotal := 0
	// 校验数据盘
	for _, d := range req.DataDisk {
		dataDiskTotal += int(d.DiskCount)
		if err := req.validateDiskSize(d.DiskType, d.DiskSizeGB); err != nil {
			return err
		}
	}

	if dataDiskTotal > 23 {
		return errors.New("data disk count should <= 23")
	}

	return nil
}

func (req *AwsCvmCreateReq) validateDiskSize(diskType typecvm.AwsVolumeType, diskSizeGB int64) error {
	switch diskType {
	case typecvm.GP2, typecvm.GP3:
		if diskSizeGB < 1 || diskSizeGB > 16384 {
			return fmt.Errorf("disk size should be 1 - 16384 when disk type is %s", diskType)
		}
	case typecvm.IO1, typecvm.IO2:
		if diskSizeGB < 4 || diskSizeGB > 16384 {
			return fmt.Errorf("disk size should be 4 - 16384 when disk type is %s", diskType)
		}
	case typecvm.ST1, typecvm.SC1:
		if diskSizeGB < 125 || diskSizeGB > 16384 {
			return fmt.Errorf("disk size should be 125 - 16384 when disk type is %s", diskType)
		}
	case typecvm.Standard:
		if diskSizeGB < 1 || diskSizeGB > 1024 {
			return fmt.Errorf("disk size should be 1 - 1024 when disk type is %s", diskType)
		}
	}
	return nil
}
