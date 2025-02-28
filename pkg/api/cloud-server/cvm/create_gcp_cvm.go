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

// GcpCvmCreateReq ...
type GcpCvmCreateReq struct {
	BkBizID       int64  `json:"bk_biz_id" validate:"omitempty"`
	AccountID     string `json:"account_id" validate:"required"`
	BkCloudID     *int64 `json:"bk_cloud_id" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Zone          string `json:"zone" validate:"required"`
	Name          string `json:"name" validate:"required,min=1,max=60"`
	InstanceType  string `json:"instance_type" validate:"required"`
	CloudImageID  string `json:"cloud_image_id" validate:"required"`
	CloudVpcID    string `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID string `json:"cloud_subnet_id" validate:"required"`

	SystemDisk struct {
		DiskType   typecvm.GcpDiskType `json:"disk_type" validate:"required"`
		DiskSizeGB int64               `json:"disk_size_gb" validate:"required,min=10"`
	} `json:"system_disk" validate:"required"`

	DataDisk []struct {
		DiskType   typecvm.GcpDiskType `json:"disk_type" validate:"required"`
		DiskSizeGB int64               `json:"disk_size_gb" validate:"required,min=10"`
		DiskCount  int64               `json:"disk_count" validate:"required,min=1"`
		Mode       typecvm.GcpDiskMode `json:"mode" validate:"required"`
		AutoDelete *bool               `json:"auto_delete" validate:"required"`
	} `json:"data_disk" validate:"omitempty"`

	// 访问主机的ssh公钥
	Password string `json:"password" validate:"required"`

	RequiredCount int64 `json:"required_count" validate:"required,min=1,max=500"`

	Memo *string `json:"memo" validate:"omitempty"`

	PublicIPAssigned bool `json:"public_ip_assigned" validate:"omitempty"`
}

// Validate ...
func (req *GcpCvmCreateReq) Validate(isFromBiz bool) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if isFromBiz && req.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	if isFromBiz && req.BkCloudID == nil {
		return errors.New("bk_cloud_id is required")
	}

	if err := validator.ValidateCvmName(enumor.Gcp, req.Name); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.RequiredCount > constant.BatchOperationMaxLimit {
		return fmt.Errorf("required count should <= %d", constant.BatchOperationMaxLimit)
	}

	// 校验系统硬盘
	if !req.isMultipleOfTwo(req.SystemDisk.DiskSizeGB) {
		return fmt.Errorf("disk size[%d] should be not multiple of 2GB", req.SystemDisk.DiskSizeGB)
	}
	// 校验数据盘
	for _, d := range req.DataDisk {
		if !req.isMultipleOfTwo(d.DiskSizeGB) {
			return fmt.Errorf("disk size[%d] should be not multiple of 2GB", d.DiskSizeGB)
		}
	}

	return nil
}

func (req *GcpCvmCreateReq) isMultipleOfTwo(size int64) bool {
	return size%2 == 0
}
