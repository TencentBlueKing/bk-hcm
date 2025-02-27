/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package tag

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// TCloudBatchTagResRequest ...
type TCloudBatchTagResRequest struct {
	AccountID string               `json:"account_id" validate:"required"`
	Resources []TCloudResourceInfo `json:"resources,omitempty" validate:"required,min=1,max=10,dive,required"`
	Tags      []core.TagPair       `json:"tags,omitempty" validate:"required,min=1,max=10"`
}

// Validate ...
func (r *TCloudBatchTagResRequest) Validate() error {
	return validator.Validate.Struct(r)
}

// TCloudResourceInfo ...
type TCloudResourceInfo struct {
	Region     string                   `json:"region,omitempty" validate:"required"`
	ResType    enumor.CloudResourceType `json:"res_type,omitempty" validate:"required"`
	ResCloudID string                   `json:"res_cloud_id,omitempty" validate:"required"`
}

// Validate ...
func (i *TCloudResourceInfo) Validate() error {
	return validator.Validate.Struct(i)
}

// Convert to tcloud resource string
func (i *TCloudResourceInfo) Convert(account string) string {
	var service, resType string
	switch i.ResType {
	case enumor.SecurityGroupCloudResType:
		// sg resource is under cvm service
		service = "cvm"
		resType = "sg"
	case enumor.LoadBalancerCloudResType:
		service = "clb"
		resType = "clb"
	case enumor.CvmCloudResType:
		service = "cvm"
		resType = "instance"
	case enumor.DiskCloudResType:
		service = "cvm"
		resType = "volume"
	default:
		return ""
	}
	return fmt.Sprintf("qcs::%s:%s:uin/%s:%s/%s", service, i.Region, account, resType, i.ResCloudID)
}
